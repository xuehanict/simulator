package flash

import (
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
		for _, id := range path {
			p = append(p, id.(utils.RouterID))
		}
		res = append(res, p)
	}
	return res,nil
}

func (f *Flash) microRouting(src, dest utils.RouterID, amt utils.Amount, k int) (
	bool ,error) {
	allPaths, err := f.getKshortestPath(src,dest, k)
	if err != nil {
		return false, err
	}
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
		remaining := amt - amountSum(sentList)
		sent := utils.Amount(0)
		if pathCap > remaining {
			sent = remaining
		} else {
			sent = pathCap
		}
		sentList = append(sentList, sent)
		sentPath = append(sentPath, currPath)

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
	if math.Abs(float64(amountSum(sentList) - amt)) < 0.000000001 {
		err := f.UpdateWeights(sentPath, sentList)
		if err != nil {
			return false, err
		}
	}
	return true, nil

}

func (f *Flash)getCachePaths(src, dest utils.RouterID) []utils.Path {
	return nil
}

func (f *Flash)addCachePath(src, dest utils.RouterID, path utils.Path)  {

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
