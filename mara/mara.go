package mara

import (
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"github.com/lukpank/go-glpk/glpk"
	fibHeap "github.com/starwander/GoFibonacciHeap"
	"sort"

	//	"github.com/davecgh/go-spew/spew"
	"math"
)

type Mara struct {
	Name string
	*utils.Graph
	NextHopBound int
	AmountRate   float64
	MaxAddLength float64
	PathsBase 	 map[utils.RouterID]map[utils.RouterID][]utils.Path
}

func NewMara(g *utils.Graph, nextHopBound int, amountRate float64,
	maxAddLength float64) *Mara {
	return &Mara{
		Graph: g,
		NextHopBound: nextHopBound,
		AmountRate: amountRate,
		MaxAddLength: maxAddLength,
		PathsBase: make(map[utils.RouterID]map[utils.RouterID][]utils.Path),
	}
}

const (
	PROBE_AMOUNT_RATE      = 0.01
	DEFAULT_PATH_ADD_LENTH = 1
	MAX_ADJACENT           = 100000
	MARA_MC                = 0
	MARA_SPE               = 1
	MARA_SPT               = 2

	// 线性规划目标
	MIN_FEES     = 1
	RATE_BALANCE = 2
)

type capElement struct {
	id      utils.RouterID
	capcity float64
}

func (c *capElement) Tag() interface{} {
	return c.id
}

func (c *capElement) Key() float64 {
	return c.capcity
}

func (m *Mara) MaraMC(startID utils.RouterID) *utils.DAG {
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
	return getDAG(ordering, nodes)
}

// 优化过的MaraMC算法，时间复杂度降低了很多
func (m *Mara) MaraMcOPT(startID utils.RouterID) *utils.DAG {
	nodes := m.Nodes

	if _, ok := m.Distance[startID]; !ok {
		m.SPTs[startID], m.Distance[startID] = utils.Dijkstra(m.Nodes, startID)
	}
	ordering := make([]utils.RouterID, 1)
	ordering[0] = startID

	capcity := fibHeap.NewFibHeap()
	for id, node := range m.Nodes {
		if node.CheckLink(startID) {
			err := capcity.Insert(utils.RouterID(id), MAX_ADJACENT-1)
			if err != nil {
				fmt.Printf("insert to heap error")
			}
		} else {
			err := capcity.Insert(utils.RouterID(id), MAX_ADJACENT)
			if err != nil {
				fmt.Printf("insert to heap error")
			}
		}
	}
	err := capcity.Delete(startID)
	if err != nil {
		fmt.Printf(" delete heap error:%v\n", err)
	}

	for {
		if len(ordering) == len(nodes) {
			break
		}
		max, _ := capcity.ExtractMin()
		maxID := max.(utils.RouterID)
		ordering = append(ordering, maxID)

		for id := range m.Nodes[maxID].Neighbours {
			if tmp := capcity.GetTag(id); tmp != math.Inf(-1) {
				err := capcity.DecreaseKey(id, tmp-1)
				if err != nil {
					fmt.Printf("decrease heap error:%v", err)
				}
			}
		}
	}

	//	spew.Dump(ordering)
	return getDAG(ordering, nodes)
}

// 优化过的MaraSPE，时间复杂度降低了很多
func (m *Mara) MaraSpeOpt(startID utils.RouterID) *utils.DAG {
	nodes := m.Nodes
	if _, ok := m.SPTs[startID]; !ok {
		fmt.Printf("算最短路\n")
		m.SPTs[startID], m.Distance[startID] = utils.Dijkstra(m.Nodes, startID)
		fmt.Printf("算完最短路\n")
	}
	spt := m.SPTs[startID]
	ordering := make([]utils.RouterID, 1)
	ordering[0] = startID

	T := fibHeap.NewFibHeap()
	S := make(map[utils.RouterID]struct{})
	S[startID] = struct{}{}
	capcity := make(map[utils.RouterID]float64)

	// 对所有节点的capcity初始化为0
	for i := range nodes {
		capcity[utils.RouterID(i)] = MAX_ADJACENT
	}

	// 对start的邻居初始化
	for i := range nodes[startID].Neighbours {
		capcity[i] = MAX_ADJACENT - 1
		if spt.Vertexs[i].CheckParent(startID) {
			err := T.Insert(i, capcity[i])
			if err != nil {
				fmt.Printf("insert value to T faced error :%v", err)
			}
		}
	}

	// 循环，每次展开一个
	for {
		if len(ordering) == len(nodes) {
			break
		}
		tag, _ := T.ExtractMin()
		id := tag.(utils.RouterID)
		S[id] = struct{}{}
		for i := range nodes[id].Neighbours {
			if _, ok := S[i]; ok {
				continue
			}
			capcity[i] = capcity[i] - 1
			if spt.Vertexs[i].CheckParent(id) {
				if tmp := T.GetTag(i); tmp != math.Inf(-1) {
					err := T.DecreaseKey(i, capcity[i])
					if err != nil {
						fmt.Printf("decrease failed%v", err)
					}
				} else {
					err := T.Insert(i, capcity[i])
					if err != nil {
						fmt.Printf("insert value to T faced error :%v", err)
					}
				}
			}
		}
		ordering = append(ordering, id)
	}

	//	spew.Dump(ordering)
	return getDAG(ordering, nodes)
}

// 获取供交易的路径，沿父节点的方向向上至dest节点
func (m *Mara) getRoutes(src, dest utils.RouterID,
	amount utils.Amount, metric *utils.Metrics) []utils.Path {

	if _, ok := m.DAGs[dest]; !ok {
		m.DAGs[dest] = m.MaraSpeOpt(dest)
	}
	fmt.Printf("DAG构架能完成\n")
	return m.nextHop(nil, src, dest, amount,
		DEFAULT_PATH_ADD_LENTH, PROBE_AMOUNT_RATE, metric)
}

// 获取供交易的路径，沿父节点的方向向上至dest节点
func (m *Mara) getRoutesWithBond(src, dest utils.RouterID,
	algo int, amount utils.Amount, metric *utils.Metrics) []utils.Path {

	if _, ok := m.DAGs[dest]; !ok {
		switch algo {
		case MARA_MC:
			m.DAGs[dest] = m.MaraMcOPT(dest)
		case MARA_SPE:
			m.DAGs[dest] = m.MaraSpeOpt(dest)
		case MARA_SPT:
			m.DAGs[dest], m.Distance[dest] = utils.Dijkstra(m.Nodes, dest)
		}
	}

	finalLen := m.MaxAddLength + m.Distance[dest][src]
	if m.Distance[dest][src] > 10 {
		finalLen = 2
	}
	//fmt.Printf("DAG构架能完成\n")
	fmt.Printf("maxLength:%v\n", finalLen)
	return m.nextHop(nil, src, dest, amount,
		finalLen, m.AmountRate, metric)
}

type parentSorter struct {
	current  utils.RouterID
	parents  []utils.RouterID
	channels map[string]*utils.Link
}

func (s parentSorter) Len() int {
	return len(s.parents)
}

func (s parentSorter) Less(i, j int) bool {
	vi := utils.GetLinkValue(s.current, utils.RouterID(i), s.channels)
	vj := utils.GetLinkValue(s.current, utils.RouterID(j), s.channels)
	return vi > vj
}

func (s parentSorter) Swap(i, j int) {
	s.parents[i], s.parents[j] = s.parents[j], s.parents[i]
}

func (m *Mara) nextHop(curPath []utils.RouterID, current,
	dest utils.RouterID, amount utils.Amount, maxLength float64,
	amtRate float64, metric *utils.Metrics) []utils.Path {

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

		sorter := parentSorter{
			current:  current,
			parents:  m.DAGs[dest].Vertexs[current].Parents,
			channels: m.Channels,
		}
		sort.Sort(sorter)

		if sorter.Len() > m.NextHopBound {
			sorter.parents = sorter.parents[0:m.NextHopBound]
		}
		for _, pnode := range sorter.parents {

			val := utils.GetLinkValue(current, pnode, m.Channels)
			if val < amount*utils.Amount(amtRate) ||
				float64(len(curPath)) > (maxLength - m.Distance[dest][current]) {
				continue
			}
			metric.ProbeMessgeNum++
			tmpPaths := m.nextHop(newCurPath, pnode, dest, amount,
				maxLength, amtRate, metric)
			if len(tmpPaths) != 0 {
				paths = append(paths, tmpPaths...)
			}
		}
		return paths
	}
}

func (m *Mara) SendPayment(src, dest utils.RouterID,
	amount utils.Amount) error {
	metric := &utils.Metrics{}
	if amount == 0 {
		return nil
	}
	routes := m.getRoutes(src, dest, amount, metric)
	result, err := m.allocMoney(routes, amount)
	if err != nil {
		return err
	}
	err = m.UpdateWeights(routes, result)
	//	spew.Dump(m.Channels)
	return err
}

func (m *Mara) SendPaymentWithBond(src, dest utils.RouterID, algo int,
	amount utils.Amount) (
	int, int, *utils.Metrics, error) {

	metric := &utils.Metrics{0, 0, 0, 0}
	if amount == 0 {
		return 0, 0, metric, nil
	}
	routes := m.getRoutesWithBond(src, dest, algo, amount, metric)
	if len(routes) == 0 {
		return 0, 0, metric, &PaymentError{
			Code:        FIND_PATH_FAILED,
			Description: "cannot find a path",
		}
	}

	result, err := m.allocMoney(routes, amount)
	if err != nil {
		return 0, 0, metric, &PaymentError{
			Code:        ALLOCARION_FAILED,
			Description: "allocation failed :" + err.Error(),
		}
	}

	if len(result) != len(routes) {
		return 0, 0, metric, &PaymentError{
			Code:        ALLOCATION_NOT_MATCH_ROUTE,
			Description: "allocation donn't match routes",
		}
	}

	selectedRoutes := make([]utils.Path, 0)
	selectedResult := make([]utils.Amount, 0)
	total := 0.0

	for i := 0; i < len(result); i++ {
		if result[i] != 0 {
			total += float64(result[i])
			selectedRoutes = append(selectedRoutes, routes[i])
			selectedResult = append(selectedResult, result[i])
		}
	}

	if math.Abs(total-float64(amount)) > 0.0000000001 {
		return 0, 0, metric, &PaymentError{
			Code:        ALLOCARION_FAILED,
			Description: "allocation failed.",
		}
	}

	err = m.UpdateWeights(selectedRoutes, selectedResult)
	for i, path := range selectedRoutes {
		if metric.MaxPathLengh < len(path) {
			metric.MaxPathLengh = len(path)
		}
		metric.OperationNum += int64(len(path) - 1)
		metric.Fees += m.GetFee(path, selectedResult[i])
	}
	if err != nil {
		return len(routes), len(selectedRoutes), metric, &PaymentError{
			Code:        UPDATE_LINK_FAILED,
			Description: fmt.Sprintf("update link failed :%v", err.Error()),
		}
	}
	m.storePathsToCache(src,dest,routes)
	return len(routes), len(selectedRoutes), metric, nil
}

// 给每条路径分配一定的钱
func (m *Mara) allocMoney(routes []utils.Path,
	amount utils.Amount) ([]utils.Amount, error) {

	if len(routes) == 1 {
		cap := utils.GetPathCap(routes[0], m.Channels)
		if cap < amount {
			return nil, fmt.Errorf("no enough money")
		} else {
			return []utils.Amount{amount}, nil
		}
	}
	channelIndexs := make(map[string]int, 0)
	routeMins := make([]utils.Amount, len(routes))
	channelVals := make(map[string]utils.Amount)

	// 计算出每条路径的最小值，并且获取每条通道的容量
	for j, path := range routes {
		min := utils.Amount(math.MaxFloat64)
		for i := 0; i < len(path)-1; i++ {
			val := utils.GetLinkValue(path[i], path[i+1], m.Channels)
			key := utils.GetLinkKey(path[i], path[i+1])
			channelVals[key] = val
			if val < min {
				min = val
			}
		}
		routeMins[j] = min
	}

	// 然后再算出每个通道的索引，以便在线性规划列约束矩阵时使用
	cursor := 0
	for channelKey, val := range channelVals {
		if val > 0 {
			channelIndexs[channelKey] = cursor
			cursor++
		}
	}
	if len(routes) == 0 {
		return nil, fmt.Errorf("cannot find path")
	}
	return m.linearProgram(routes, channelIndexs, routeMins, channelVals, amount)
}

// 进行线性规划运算
func (m *Mara) linearProgram(
	routes []utils.Path,
	channelIndexs map[string]int,
	routeMins []utils.Amount,
	channelVals map[string]utils.Amount,
	amount utils.Amount) ([]utils.Amount, error) {

	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("faced err :%v", e)
		}
	}()

	lp := glpk.New()
	lp.SetProbName("payment allocation")
	lp.SetObjName("Z")
	lp.SetObjDir(glpk.MIN)

	// 所有使用到这个通道的路径的钱的数目之和要小于当前通道的容量
	lp.AddRows(len(channelIndexs) + 1)
	for chanKey, index := range channelIndexs {
		lp.SetRowName(index+1, chanKey)
		lp.SetRowBnds(index+1, glpk.DB, 0.0, float64(channelVals[chanKey]))
	}
	lp.SetRowName(len(channelIndexs)+1, "amount")
	lp.SetRowBnds(len(channelIndexs)+1, glpk.FX, float64(amount), float64(amount))

	// 所有路径使用的通道容量都应该小于路径能走的最大流量（所有channel中的min）
	lp.AddCols(len(routes))
	for i, min := range routeMins {
		name := fmt.Sprintf("p%d", i+1)
		lp.SetColName(i+1, name)
		lp.SetColBnds(i+1, glpk.DB, 0.0, float64(min))
	}
	/*
		for j, route := range routes {
			lp.SetObjCoef(j+1, float64(len(route)))
		}
	*/

	for j, route := range routes {
		lp.SetObjCoef(j+1, float64(m.GetFee(route, utils.Amount(1))))
	}

	ind := []int32{0}
	for i := range routeMins {
		ind = append(ind, int32(i)+1)
	}

	matrixRows := make(map[string][]int)
	for i, route := range routes {
		for j := 0; j < len(route)-1; j++ {
			key := utils.GetLinkKey(route[j], route[j+1])
			matrixRows[key] = append(matrixRows[key], i)
		}
	}

	for row, paths := range matrixRows {
		a := make([]float64, len(routes)+1)
		for i := range a {
			a[i] = 0
		}
		for _, path := range paths {
			a[path+1] = 1.0
		}

		lp.SetMatRow(channelIndexs[row]+1, ind, a)
	}
	a := make([]float64, len(routes)+1)
	for i := range a {
		a[i] = 1
	}
	a[0] = 0
	lp.SetMatRow(len(channelIndexs)+1, ind, a)

	err := lp.Simplex(nil)
	//	fmt.Printf("%s = %g", lp.ObjName(), lp.MipObjVal())
	result := make([]utils.Amount, 0)
	for i := 0; i < len(routes); i++ {
		result = append(result, utils.Amount(lp.ColPrim(i+1)))
		//		fmt.Printf("; %s = %g", lp.ColName(i+1), lp.ColPrim(i+1))
	}
	fmt.Println()
	//	lp.Delete()

	return result, err
}

// 基于前面得到的序列，构建DAG
func getDAG(ordering []utils.RouterID, nodes map[utils.RouterID]*utils.Node) *utils.DAG {

	mapOrdering := make(map[utils.RouterID]int, len(ordering))
	for index, id := range ordering {
		mapOrdering[id] = index
	}

	//tmpLinks := make([]*Link,0)
	dag := utils.NewDAG(nodes[ordering[0]], len(nodes))
	tmpNodes := utils.CopyNodes(nodes)
	dag.Vertexs = tmpNodes

	for i := 0; i < len(ordering); i++ {
		for n := range nodes[ordering[i]].Neighbours {
			if mapOrdering[n] > i {

				tmpNodes[ordering[i]].Children = append(tmpNodes[ordering[i]].Children,
					n)
				tmpNodes[n].Parents = append(tmpNodes[n].Parents,
					ordering[i])
			}
		}
	}
	return dag
}

// 只是得到支付的路径和分配的金额，但是不进行真正的支付行为
func (m *Mara) TryPay(src, dest utils.RouterID, algo int,
	amount utils.Amount) ([]utils.Path, []utils.Amount, error) {
	metric := &utils.Metrics{}
	routes := m.getRoutesWithBond(src, dest, algo, amount, metric)
	if len(routes) == 0 {
		return nil, nil, &PaymentError{
			Code:        FIND_PATH_FAILED,
			Description: "cannot find a path",
		}
	}

	result, err := m.allocMoney(routes, amount)
	if err != nil {
		return nil, nil, &PaymentError{
			Code:        ALLOCARION_FAILED,
			Description: "allocation failed :" + err.Error(),
		}
	}

	if len(result) != len(routes) {
		return nil, nil, &PaymentError{
			Code:        ALLOCATION_NOT_MATCH_ROUTE,
			Description: "allocation donn't match routes",
		}
	}

	selectedRoutes := make([]utils.Path, 0)
	selectedResult := make([]utils.Amount, 0)
	total := 0.0

	for i := 0; i < len(result); i++ {
		if result[i] != 0 {
			total += float64(result[i])
			selectedRoutes = append(selectedRoutes, routes[i])
			selectedResult = append(selectedResult, result[i])
		}
	}

	if math.Abs(total-float64(amount)) > 0.0000000001 {
		return nil, nil, &PaymentError{
			Code:        ALLOCARION_FAILED,
			Description: "allocation failed.",
		}
	}
	return selectedRoutes, selectedResult, nil
}

// 从路径仓库中取出之前得到的路径，以重复利用
func (m *Mara) getPathsFromCache(src,dest utils.RouterID) []utils.Path {
	if cacheMap, ok := m.PathsBase[src]; ok {
		if paths, ok1 := cacheMap[dest]; ok1 {
			return paths
		}
	}
	return nil
}

// 将路径存入到路径仓库中，以备以后继续使用，以减少probe的overhead
func (m *Mara) storePathsToCache(src, dest utils.RouterID, paths []utils.Path)  {
	if _, ok := m.PathsBase[src]; !ok {
		m.PathsBase[src] = make(map[utils.RouterID][]utils.Path)
	}
	m.PathsBase[src][dest] = paths
}

func (m *Mara)clearPathsCache(src, dest utils.RouterID)  {
	delete(m.PathsBase[src], dest)
}

