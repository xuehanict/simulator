package mpdv

import (
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"github.com/lukpank/go-glpk/glpk"
	"math"
	"sort"
)

type Mpdv struct {
	*utils.Graph
	// 第一个是key是src，第二个key是dest, 第三个key是邻居，最后是距离
	table map[utils.RouterID]map[utils.RouterID]map[utils.RouterID]int

	NextHopBound int
	amtRate utils.Amount
}

func (m *Mpdv)ResetTable(dests map[utils.RouterID]struct{}) {
	table := make(map[utils.RouterID]map[utils.RouterID]map[utils.RouterID]int)

	/*
	for _, dest := range dests {
		m.SPTs[dest], m.Distance[dest] = utils.Dijkstra(m.Nodes, dest)
	}
	 */

	// 对每个结点构建路由表
	for _, node := range m.Nodes {
		table[node.ID] = make(map[utils.RouterID]map[utils.RouterID]int)
		for dest := range dests {
			table[node.ID][dest] = make(map[utils.RouterID]int)
			for _, nei := range node.Neighbours {
				// 同样距离的结点或者小距离的结点才可能作为下一跳
				if m.Distance[dest][nei] <= m.Distance[dest][node.ID] {
					table[node.ID][dest][nei] = int(m.Distance[dest][nei]) + 1
				}
			}
		}
	}
	m.table = table
}

func (m *Mpdv)InitTable(dests  map[utils.RouterID]struct{}) {
	table := make(map[utils.RouterID]map[utils.RouterID]map[utils.RouterID]int)

	for dest := range dests {
		m.SPTs[dest], m.Distance[dest] = utils.Dijkstra(m.Nodes, dest)
	}

	// 对每个结点构建路由表
	for _, node := range m.Nodes {
		table[node.ID] = make(map[utils.RouterID]map[utils.RouterID]int)
		for dest := range dests {
			table[node.ID][dest] = make(map[utils.RouterID]int)
			for _, nei := range node.Neighbours {
				// 同样距离的结点或者小距离的结点才可能作为下一跳
				if m.Distance[dest][nei] <= m.Distance[dest][node.ID] {
					table[node.ID][dest][nei] = int(m.Distance[dest][nei]) + 1
				}
			}
		}
	}
	m.table = table
}

func (m *Mpdv)findPaths(src, dest utils.RouterID, amt utils.Amount,
	metric *utils.Metrics) ([]utils.Path, error) {
	return m.nextHop(src,dest,amt,nil,metric, m.amtRate), nil
}

func (m *Mpdv)nextHop(current, dest utils.RouterID, amt utils.Amount,
	curPath utils.Path, metric *utils.Metrics, amtRate utils.Amount,
	) []utils.Path {
		// arrived in the end. we return the final path.
	if current == dest {
		newCurPath := make([]utils.RouterID, len(curPath)+1)
		copy(newCurPath, curPath)
		newCurPath[len(newCurPath)-1] = current
		return []utils.Path{newCurPath}
	} else {
		// we continue to pass the request until the destination.
		paths := make([]utils.Path, 0)
		newCurPath := make([]utils.RouterID, len(curPath)+1)
		copy(newCurPath, curPath)
		newCurPath[len(newCurPath)-1] = current

		nextHops := make([]utils.RouterID, 0)
		for next := range m.table[current][dest] {
			if !checkInPath(next, curPath) {
				nextHops = append(nextHops, next)
			}
		}
		sorter := parentSorter{
			current:  current,
			parents:  nextHops,
			channels: m.Channels,
			dis: m.Distance[current],
		}
		sort.Sort(sorter)

		// TODO(xuehan) 判断长度
		if sorter.Len() > m.NextHopBound {
			sorter.parents = sorter.parents[0:m.NextHopBound]
		}
		for _, pnode := range sorter.parents {
			val := utils.GetLinkValue(current, pnode, m.Channels)
			if val < amt*utils.Amount(amtRate) {
				continue
			}
			metric.ProbeMessgeNum++
			tmpPaths := m.nextHop(pnode,dest,amt,newCurPath, metric, amtRate)
			if len(tmpPaths) != 0 {
				paths = append(paths, tmpPaths...)
			}
		}
		return paths
	}
}

func (m *Mpdv)SendPayment(amt utils.Amount, from, to utils.RouterID) (
	*utils.Metrics, error) {
	metiric := &utils.Metrics{0,0,0,0}
	if amt < 0.000001 {
		return metiric,nil
	}
	paths,err := m.findPaths(from, to, amt, metiric)
	if len(paths) == 0 {
		return metiric, fmt.Errorf("routing failed")
	}
	//spew.Dump(paths)
	amts, err := m.allocMoney(amt, paths)
	if err != nil {
		return metiric, err
	}
	//spew.Dump(amts)

	if math.Abs(float64(amountSum(amts)	- amt)) > 0.0000001 {
		return metiric, fmt.Errorf("allocation failed")
	}

	err = m.UpdateWeights(paths, amts)
	for i, path := range paths {
		if amts[i] != 0 {
			metiric.OperationNum += int64(len(path)-1)
			if len(path) > metiric.MaxPathLengh {
				metiric.MaxPathLengh = len(path)
			}
			metiric.Fees += m.GetFee(path, amts[i])
		}
	}
	if err != nil {
		return metiric, err
	}
	return metiric, nil
}


func (m *Mpdv) linearProgram (amt utils.Amount, paths []utils.Path,
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
		lp.SetRowBnds(index, glpk.DB, - float64(m.Channels[chanKey].Val2),
			float64(m.Channels[chanKey].Val1))
	}
	lp.SetRowName(len(channelIndex)+1, "amount")
	lp.SetRowBnds(len(channelIndex)+1, glpk.FX, float64(amt), float64(amt))

	lp.AddCols(len(paths))
	for i := range paths {
		name := fmt.Sprintf("p%d", i+1)
		lp.SetColName(i+1, name)
		lp.SetColBnds(i+1, glpk.LO, 0.0, 0.0)
	}

	// TODO(xuehan): 费用最低
	for j, path := range paths {
		lp.SetObjCoef(j+1, float64(len(path)))
	}
	// 为了测试，构建一个矩阵，其实可以直接插入
	matrix := make([][]float64, len(channelIndex))
	for key, index := range channelIndex {
		row := make([]float64, 0)
		for _, path := range paths {
			row = append(row, checkKeyInPath(path, key))
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

type parentSorter struct {
	current  utils.RouterID
	parents  []utils.RouterID
	channels map[string]*utils.Link
	dis map[utils.RouterID]float64
}

func (s parentSorter) Len() int {
	return len(s.parents)
}

func (s parentSorter) Less(i, j int) bool {
	vi := utils.GetLinkValue(s.current, utils.RouterID(i), s.channels)
	vj := utils.GetLinkValue(s.current, utils.RouterID(j), s.channels)

	if s.dis[utils.RouterID(i)] < s.dis[utils.RouterID(j)] {
		return true
	} else if s.dis[utils.RouterID(i)] == s.dis[utils.RouterID(j)] {
		return vi > vj
	}
	return false
}

func (s parentSorter) Swap(i, j int) {
	s.parents[i], s.parents[j] = s.parents[j], s.parents[i]
}

func checkInPath(id utils.RouterID, path utils.Path) bool  {
	for _, n := range path {
		if n == id {
			return true
		}
	}
	return false
}

func (m *Mpdv) allocMoney (amt utils.Amount, paths []utils.Path) ([]utils.Amount, error) {
	channels := make(map[string]int)
	for _, path := range paths {
		for i:=0; i<len(path)-1; i++{
			if _, ok := channels[utils.GetLinkKey(path[i],path[i+1])]; !ok {
				channels[utils.GetLinkKey(path[i],path[i+1])]= len(channels) + 1
			}
		}
	}
	return m.linearProgram(amt, paths, channels)
}

func checkKeyInPath(path utils.Path, key string) float64 {
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

func amountSum(a []utils.Amount) utils.Amount {
	res := utils.Amount(0)
	for _, amt := range a {
		res += amt
	}
	return res
}

func NewMpdv(graph *utils.Graph, hopBound int, amtRate utils.Amount) *Mpdv {
	return &Mpdv{
		Graph:graph,
		NextHopBound: hopBound,
		amtRate: amtRate,
	}
}