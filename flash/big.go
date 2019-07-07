package flash

import (
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"github.com/lukpank/go-glpk/glpk"
	"math"
)

func (f *Flash)elephantRouting(amt utils.Amount, from, to utils.RouterID)(
	*utils.Metrics, error) {
	metiric := &utils.Metrics{0,0,0,0}
	paths, _, err := f.findPaths(from, to, metiric)
	if err != nil {
		return metiric, fmt.Errorf("routing failed :%s", err.Error())
	}
	//spew.Dump(paths)
	amts, err := f.allocMoney(amt, paths)
	if err != nil {
		return metiric, err
	}
	//spew.Dump(amts)

	if math.Abs(float64(amountSum(amts)	- amt)) > 0.0000001 {
		return metiric, fmt.Errorf("allocation failed")
	}

	err = f.UpdateWeights(paths, amts)
	for i, path := range paths {
		if amts[i] != 0 {
	metiric.OperationNum += int64(len(path)-1)
			if len(path) > metiric.MaxPathLengh {
				metiric.MaxPathLengh = len(path)
			}
			metiric.Fees += f.GetFee(path, amts[i])
		}
	}
	if err != nil {
		return metiric, err
	}
	return metiric, nil
}

func (f *Flash)findPaths(src, dest utils.RouterID, metric *utils.Metrics)(
	[]utils.Path, []utils.Amount, error) {
		localChannel := utils.CopyChannels(f.Channels)
	pathSet := make([]utils.Path, 0)
	capSet := make([]utils.Amount,0)
	path := utils.BfsPath(f.Nodes, src, dest, true, localChannel)
	metric.ProbeMessgeNum += int64(len(path)-1)

	for i := 0; i < f.pathN && path != nil; i++ {
		pathSet = append(pathSet, path)
		pathCap := utils.GetPathCap(path, localChannel)
		capSet = append(capSet,pathCap)
		err := utils.UpdateWeights([]utils.Path{path}, []utils.Amount{pathCap},
		localChannel)
		if err != nil {
			return nil, nil, err
		}
		path = utils.BfsPath(f.Nodes, src, dest, true, localChannel)
		metric.ProbeMessgeNum += int64(len(path)-1)
	}

	return pathSet, capSet, nil
}

func (f *Flash) allocMoney (amt utils.Amount, paths []utils.Path) ([]utils.Amount, error) {
	channels := make(map[string]int)
	for _, path := range paths {
		for i:=0; i<len(path)-1; i++{
			if _, ok := channels[utils.GetLinkKey(path[i],path[i+1])]; !ok {
				channels[utils.GetLinkKey(path[i],path[i+1])]= len(channels) + 1
			}
		}
	}
	return f.linearProgram(amt, paths, channels)
}

func (f *Flash) linearProgram (amt utils.Amount, paths []utils.Path,
	channelIndex map[string]int) ([]utils.Amount, error){
	lp := glpk.New()
	lp.SetProbName("sample")
	lp.SetObjName("Z")
	lp.SetObjDir(glpk.MIN)
	//TODO(xuehan): modify it

	//spew.Dump(f.Channels)
	lp.AddRows(len(channelIndex) + 1)
	for chanKey, index := range channelIndex {
		//fmt.Printf("添加row %s, index是%v, up 是%v, lo 是%v \n", chanKey, index,
		//	float64(f.Channels[chanKey].Part1), -float64(f.Channels[chanKey].Part2))
		lp.SetRowName(index, chanKey)
		lp.SetRowBnds(index, glpk.DB, - float64(f.Channels[chanKey].Val2),
			float64(f.Channels[chanKey].Val1))
	}
	lp.SetRowName(len(channelIndex)+1, "amount")
	lp.SetRowBnds(len(channelIndex)+1, glpk.FX, float64(amt), float64(amt))

	lp.AddCols(len(paths))
	for i := range paths {
		name := fmt.Sprintf("p%d", i+1)
		lp.SetColName(i+1, name)
		lp.SetColBnds(i+1, glpk.LO, 0.0, 0.0)
	}

	// 费用最低
	for j, path := range paths {
		lp.SetObjCoef(j+1, float64(len(path)))
	}
	// 为了测试，构建一个矩阵，其实可以直接插入
	matrix := make([][]float64, len(channelIndex))
	for key, index := range channelIndex {
		row := make([]float64, 0)
		for _, path := range paths {
			row = append(row, checkInPath(path, key))
		}
		row = append([]float64{0}, row...)
		matrix[index-1] = row
	}

	//spew.Dump(matrix)
	ind := []int32{0}

	for i:= range paths {
		ind = append(ind, int32(i) + 1)
	}

	for i, row := range matrix {
		lp.SetMatRow(i+1, ind, row)
	}
	amtRow := make([]float64,len(paths) + 1)
	for i := range amtRow {
		amtRow[i] = 1
	}
	amtRow[0] = 0
	lp.SetMatRow(len(channelIndex)+1, ind, amtRow)

	err := lp.Simplex(nil)
	//	fmt.Printf("%s = %g", lp.ObjName(), lp.MipObjVal())
	result := make([]utils.Amount, 0)
	for i := 0; i < len(paths); i++ {
		result = append(result, utils.Amount(lp.ColPrim(i+1)))
		//		fmt.Printf("; %s = %g", lp.ColName(i+1), lp.ColPrim(i+1))
	}
	fmt.Println()
	//	lp.Delete()
	return result, err
}

func checkInPath(path utils.Path, key string) float64 {
	for i := 0; i< len(path)-1; i++ {
		if utils.GetLinkKey(path[i], path[i+1]) == key {
			if path[i] > path[i]+1 {
				return -1
			}
			if path[i] < path[i]+1 {
				return 1
			}
		}
	}
	return 0
}

