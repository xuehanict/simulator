package flash

import (
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"github.com/starwander/goraph"
	"math"
)

func (f *Flash) convertGraph () (*goraph.Graph,error) {
	graph := goraph.NewGraph()
	for _, node := range f.Nodes {
		err := graph.AddVertex(node.ID,nil)
		if err != nil {
			return nil, err
		}
	}

	for _, node := range f.Nodes {
		for _, nei := range node.Neighbours {
			err := graph.AddEdge(node.ID, nei, 1, nil)
			if err != nil {
				return nil, err
			}
		}
	}
	return graph,nil
}

func (f *Flash) getKshortestPath(from, to utils.RouterID, k int) (
	[]utils.Path,error){
	_, paths, _ := f.goGraph.Yen(from, to, k)
	res := make([]utils.Path, 0)
	for _, path := range paths {
		p := make([]utils.RouterID, 0)
		if len(path) == 0 {
			continue
		}
		for _, id := range path {
			p = append(p, id.(utils.RouterID))
		}
		res = append(res, p)
	}
	return res,nil
}

func (f *Flash) microRouting(src, dest utils.RouterID, amt utils.Amount, k int) (
	*utils.Metrics, error) {

	metric := &utils.Metrics{0,0,0,0}
	var allPaths []utils.Path
	if f.test == false {
		tmpAllPaths, err := f.getKshortestPath(src,dest, k)
		if err != nil {
			return metric, err
		}
		allPaths = tmpAllPaths
	} else {
		allPaths = f.GetShortestPathsForTest(src,dest)
	}
	if allPaths == nil || len(allPaths) < 1 {
		return metric, fmt.Errorf("no path")
	}
	//spew.Dump(allPaths)

	pathSets := f.getCachePaths(src,dest)

	if pathSets == nil {
		pathSets = append(pathSets, utils.CopyPath(allPaths[0]))
		f.addCachePath(src, dest, allPaths[0])
	}

	sentList := make([]utils.Amount,0)
	sentPath := make([]utils.Path,0)
	currPath := pathSets[0]
	for {
		pathCap := utils.GetPathCap(currPath, f.Channels)
		metric.ProbeMessgeNum += int64(len(currPath)-1)
		remaining := amt - amountSum(sentList)
		sent := utils.Amount(0)
		if pathCap > remaining {
			sent = remaining
		} else {
			sent = pathCap
		}
		sentList = append(sentList, sent)
		sentPath = append(sentPath, currPath)
		err := utils.UpdateWeights([]utils.Path{currPath},[]utils.Amount{sent}, f.Channels)
		metric.OperationNum += int64(len(currPath)-1)
		if err != nil {
			return metric, err
		}
		if !(amountSum(sentList) < amt &&
			len(sentList) < len(allPaths)) {
			break
		}
		pathSets = f.getCachePaths(src, dest)
		// 还有缓存的路径没用
		if len(sentList) < len(pathSets) {
			currPath = pathSets[len(sentList)]
		} else {
			rest := pathSetDelete(allPaths, pathSets)
			currPath = rest[0]
			f.addCachePath(src,dest,currPath)
		}
	}
	if math.Abs(float64(amountSum(sentList) - amt)) > 0.0000001 {
		err := f.UpdateWeightsReverse(sentPath, sentList)
		for _, path := range sentPath {
			metric.OperationNum += int64(len(path)-1)
		}
		if err != nil {
			return metric, err
		}
		return metric, fmt.Errorf("not enough")
	} else {
		for i, amt := range sentList {
			metric.Fees += f.GetFee(sentPath[i], amt)
			if len(sentPath[i]) > metric.MaxPathLengh {
				metric.MaxPathLengh = len(sentPath[i])
			}
		}
		return metric, nil
	}
}

func (f *Flash)getCachePaths(src, dest utils.RouterID) []utils.Path {
	srcTable, ok := f.routingTable[src]
	if !ok {
		return nil
	}
	destPaths, ok := srcTable[dest]
	if !ok {
		return nil
	}
	return destPaths
}

func (f *Flash)addCachePath(src, dest utils.RouterID, path utils.Path) {
	_, ok := f.routingTable[src]
	if !ok {
		f.routingTable[src] = make(map[utils.RouterID][]utils.Path)
	}
	_, ok = f.routingTable[src][dest]
	if !ok {
		f.routingTable[src][dest] = make([]utils.Path,0)
	}

	exsist := false
	for _, p := range f.routingTable[src][dest] {
		if isSamePath(path,p) {
			exsist = true
			break
		}
	}
	if !exsist {
		f.routingTable[src][dest] = append(f.routingTable[src][dest], path)
	}
}

func amountSum(a []utils.Amount) utils.Amount {
	res := utils.Amount(0)
	for _, amt := range a {
		res += amt
	}
	return res
}

func pathSetDelete(set1, set2 []utils.Path) []utils.Path {
	resSet := make([]utils.Path,0)
	for _, path1 := range set1 {
		exsist := false
		for _, path2 := range set2 {
			if isSamePath(path1, path2) {
				exsist = true
				break
			}
		}
		if !exsist {
			resSet = append(resSet, utils.CopyPath(path1))
		}
	}
	return resSet
}

func isSamePath(path1, path2 utils.Path) bool {
	if len(path1) != len(path2) {
		return false
	}

	for i := 0; i < len(path1); i++ {
		if path1[i] != path2[i] {
			return false
		}
	}
	return true
}

func (f *Flash)AddShortestPathsTest(src, dest utils.RouterID)  {
	f.rw.Lock()
	_, ok := f.routingTable[src]

	if !ok {
		f.routingTable[src] = make(map[utils.RouterID][]utils.Path)
	}
	f.rw.Unlock()
	_, ok = f.routingTable[src][dest]
	if !ok {
		paths, _ := f.getKshortestPath(src, dest, 4)
		f.rw.Lock()
		f.routingTable[src][dest] = paths
		f.rw.Unlock()
	}
}

func (f *Flash)GetShortestPathsForTest(src, dest utils.RouterID) []utils.Path {
	return f.routingTable[src][dest]
}
