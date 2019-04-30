package mara

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/lightningnetwork/simulator/utils"
//	"github.com/davecgh/go-spew/spew"
	"math"
	"github.com/lukpank/go-glpk/glpk"
	"fmt"
)


type Mara struct {
	Name string
	*Graph
}

const PROBE_AMOUNT_RATE = 0.1


func (m *Mara)MaraMC(startID utils.RouterID) *DAG {
	nodes := m.Nodes
	S := make(map[utils.RouterID]struct{})
	T := make(map[utils.RouterID]struct{})

	for _, node := range nodes {
		T[node.ID] = struct{}{}
	}
	delete(T, startID)

	S[startID] = struct{}{}

	ordering := make([]utils.RouterID, 1)
	ordering[0] = startID

	for {
		if len(ordering) == len(nodes) {
			break
		}
		v := utils.RouterID(-1)
		largestD := 0
		for vtx := range T {
			tmpConn := 0
			for n := range nodes[vtx].Neighbours {
				if _, ok := S[n]; ok {
					tmpConn++
				}
			}
			if tmpConn > largestD {
				largestD = tmpConn
				v = vtx
			}
		}
		ordering = append(ordering, v)
		S[v] = struct{}{}
		delete(T, v)
	}
//	spew.Dump(ordering)
	return getDAG(ordering,nodes)
}

func (m *Mara)MaraSPE(startID utils.RouterID) *DAG {
	nodes := m.Nodes
	if _, ok:= m.SPTs[startID]; !ok {
		fmt.Printf("算最短路\n")
		m.SPTs[startID] = dijkstra(m.Nodes,startID)
		fmt.Printf("算完最短路\n")
	}
	spt := m.SPTs[startID]
//	spew.Dump(spt)
	S := make(map[utils.RouterID]struct{})
	S[startID] = struct{}{}

	ordering := make([]utils.RouterID, 1)
	ordering[0] = startID

	for {
		if len(ordering) == len(nodes) {
			break
		}
		v := utils.RouterID(-1)
		T := computeT(spt,S)
		largestD := 0
		for vtx := range T {
			tmpConn := 0
			for n := range nodes[vtx].Neighbours {
				if _, ok := S[n]; ok {
					tmpConn++
				}
			}
			if tmpConn > largestD {
				largestD = tmpConn
				v = vtx
			}
		}
		fmt.Printf("ordering \n")
		ordering = append(ordering, v)
		S[v] = struct{}{}
	}
//	spew.Dump(ordering)
	return getDAG(ordering,nodes)
}

// 获取供交易的路径，沿父节点的方向向上至dest节点
func (m *Mara) getRoutes (src, dest utils.RouterID,
	amount utils.Amount) [][]utils.RouterID {

	if _,ok := m.DAGs[dest]; !ok {
		m.DAGs[dest] = m.MaraSPE(dest)
		//m.DAGs[dest] = m.MaraMC(dest)
	}
	fmt.Printf("DAG构架能完成\n")
	return m.nextHop(nil, src,dest,amount)
}

func (m *Mara) nextHop (curPath []utils.RouterID, current,
	dest utils.RouterID, amount utils.Amount) [][]utils.RouterID{

	// arrived in the end. we return the final path.
	if 	current == dest {
		finalPath := append(curPath, current)
		return  [][]utils.RouterID{finalPath}
	} else {
		// we continue to pass the request until the destination.
		paths := make([][]utils.RouterID,0)
		newCurPath := append(curPath, current)
		for _, pnode := range m.DAGs[dest].vertexs[current].Parents {
			if utils.GetLinkValue(current, pnode, m.Channels) <
				amount * PROBE_AMOUNT_RATE   {
				continue
			}
			fmt.Printf("path 长度%v\n", len(curPath))
			tmpPaths := m.nextHop(newCurPath,pnode, dest, amount)
			if len(tmpPaths) != 0 {
				paths = append(paths, tmpPaths...)
			}
		}
		return paths
	}
}

func (m *Mara)sendPayment(src, dest utils.RouterID,
	amount utils.Amount) error {
	fmt.Printf("从%v发到%v， %v tocken\n", src,dest, amount)
	routes := m.getRoutes(src, dest, amount)
	fmt.Printf("路径数目为：%v\n",len(routes))

	_, err := m.allocMoney(routes, amount)
	return err
}

func (m *Mara)allocMoney(routes [][]utils.RouterID,
		amount utils.Amount) ([]utils.Amount, error) {

	channelIndexs := make(map[string]int,0)
	routeMins := make([]utils.Amount,len(routes))
	channelVals := make(map[string]utils.Amount)

	// 计算出每条路径的最小值，并且获取每条通道的容量
	for j, path := range routes {
		min := utils.Amount(math.MaxFloat64)
		for i := 0; i< len(path) -1 ; i++ {
			val := utils.GetLinkValue(path[i],path[i+1], m.Channels)
			key := utils.GetLinkKey(path[i],path[i+1])
			channelVals[key] = val
			if val < min {
				min = val
			}
		}
		routeMins[j] = min
	}

	// 然后再算出每个通道的索引，以便在线性规划列约束矩阵时使用
	cursor := 0
	for channelKey := range channelVals {
		channelIndexs[channelKey] = cursor
		cursor++
	}
	return m.linearProgram(routes,channelIndexs, routeMins, channelVals,amount)
}

func (m *Mara)linearProgram (
		routes [][]utils.RouterID,
		channelIndexs map[string]int,
		routeMins []utils.Amount,
		channelVals map[string]utils.Amount,
		amount utils.Amount) ([]utils.Amount, error) {

	lp := glpk.New()
	lp.SetProbName("payment allocation")
	lp.SetObjName("Z")
	lp.SetObjDir(glpk.MIN)

	// 所有使用到这个通道的路径的钱的数目之和要小于当前通道的容量
	lp.AddRows(len(channelIndexs) + 1)
	for chanKey, index := range channelIndexs {
		lp.SetRowName(index + 1, chanKey)
		lp.SetRowBnds(index + 1, glpk.DB, 0.0, float64(channelVals[chanKey]))
	}
	lp.SetRowName(len(channelIndexs)+1, "amount")
	lp.SetRowBnds(len(channelIndexs) +1 ,glpk.FX, float64(amount), float64(amount))

	// 所有路径使用的通道容量都应该小于路径能走的最大流量（所有channel中的min）
	lp.AddCols(len(routes))
	for i, min := range routeMins {
		name := fmt.Sprintf("p%d",i+1)
		lp.SetColName(i+1, name)
		lp.SetColBnds(i+1, glpk.DB, 0.0, float64(min))
	}

	// 这里就任意设置一个目标函数 min： p1
	lp.SetObjCoef(1, 1.0)
	ind := []int32{0}
	for i := range routeMins {
		ind = append(ind, int32(i) + 1)
	}

	matrixRows := make(map[string][]int)
	for i, route := range routes {
		for j:=0; j < len(route) - 1; j++{
			key := utils.GetLinkKey(route[j], route[j+1])
			matrixRows[key] = append(matrixRows[key], i)
		}
	}

	for row, paths := range matrixRows{
		a := make([]float64, len(routes)+1)
		for i := range a {
			a[i] = 0
		}
		for _, path := range paths {
			a[path+1] = 1.0
		}
		spew.Dump(a)
		spew.Dump(channelIndexs[row]+1)
		lp.SetMatRow(channelIndexs[row] + 1, ind, a)
	}
	a := make([]float64, len(routes) + 1)
	for i := range a {
		a[i] = 1
	}
	a[0] = 0
	lp.SetMatRow(len(channelIndexs)+1, ind, a)
	spew.Dump(a)
	spew.Dump(len(channelIndexs)+1)

	err := lp.Simplex(nil)
	fmt.Printf("%s = %g", lp.ObjName(), lp.MipObjVal())
	result := make([]utils.Amount, 0)
	for i := 0; i < len(routes); i++ {
		fmt.Printf("; %s = %g", lp.ColName(i+1), lp.ColPrim(i+1))
		result = append(result,utils.Amount(lp.ColPrim(i+1)))
	}
	fmt.Println()
	lp.Delete()

	return result, err
}

func getDAG(ordering []utils.RouterID, nodes map[utils.RouterID]*Node) *DAG {

	mapOrdering := make(map[utils.RouterID]int, len(ordering))
	for index, id := range ordering {
		mapOrdering[id] = index
	}

	//tmpLinks := make([]*Link,0)
	dag := NewDAG(nodes[ordering[0]])
	//dag.edges = tmpLinks
	tmpNodes := copyNodes(nodes)
	dag.vertexs = tmpNodes

	for i := 0; i < len(ordering); i++ {
		for n := range nodes[ordering[i]].Neighbours {
			if mapOrdering[n] > i {

				tmpNodes[ordering[i]].Children = append(tmpNodes[ordering[i]].Children,
					n)
				tmpNodes[n].Parents = append(tmpNodes[n].Parents,
					ordering[i])
				//tmpLinks = append(tmpLinks, link)
			}
		}
	}
	return dag
}

func computeT (dag *DAG, S map[utils.RouterID]struct{}) map[utils.RouterID]struct{} {

	U := dag.vertexs
	T := make(map[utils.RouterID]struct{})

	for id, node := range U {
		if _, ok := S[id]; ok {
			continue
		}
		for _, parent := range node.Parents {
			if _, ok := S[parent]; ok {
				T[id] = struct{}{}
			}
		}
	}
	return T
}
