package dataproc

import (
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"math/rand"
	"sort"
	"time"
)

const (
	REMAINDER_SAMPLE = 0
	MAP_SAMPLE       = 1

	ORIGION_CHANNEL   = 2
	REBALANCE_CHANEL   = 3
	UNIFORMLY_CHANNEL = 4
	FIX_VALUE_CHANNEL = 5
)

func CutOneDegree(i int, g *utils.Graph) int {
	nodesToDelete := make(map[utils.RouterID]struct{})
	for _, n := range g.Nodes {
		if len(n.Neighbours) < i {
			nodesToDelete[n.ID] = struct{}{}
		}
	}

	fmt.Printf("node to delete length is %v\n", len(nodesToDelete))
	for id := range g.Nodes {
		if _, ok := nodesToDelete[id]; ok {
			delete(g.Nodes, id)
		}
	}
	fmt.Printf("remove node done\n")
	/*
	for _, node := range g.Nodes {
		for nToD := range nodesToDelete {
			node.RemoveNei(nToD)
		}
	}
	 */
	for _, node := range g.Nodes {
		for n := range node.Neighbours {
			if _, ok := nodesToDelete[n]; ok {
				node.RemoveNei(n)
			}
		}
	}

	for key, link :=range g.Channels {
		p1 := link.Part1
		p2 := link.Part2
		if _, ok := g.Nodes[p1]; !ok {
			delete(g.Channels, key)
		}
		if _, ok := g.Nodes[p2]; !ok {
			delete(g.Channels, key)
		}
	}
	fmt.Print("clear edge done\n")
	return len(nodesToDelete)
}

func RemoveZeroEdge(g *utils.Graph)  {
	for key, link := range g.Channels {
		if _, ok := g.Nodes[link.Part1]; !ok  {
			panic(0)
		}
		if _, ok := g.Nodes[link.Part2]; !ok  {
			panic(0)
		}

		if link.Val1 == 0 && link.Val2 == 0 {
			delete(g.Nodes[link.Part1].Neighbours, link.Part2)
			delete(g.Nodes[link.Part2].Neighbours, link.Part1)
			delete(g.Channels, key)
		}
	}
}

func GetMaxComponent(g *utils.Graph) []utils.RouterID {
	allId := make(map[utils.RouterID]struct{})
	sets := make([][]utils.RouterID, 0)
	for key := range g.Nodes {
		allId[key] = struct{}{}
	}

	count := 0
	for id := range allId {
		find := false
		count ++
		if count % 100 == 0 {
			fmt.Print("100 node done\n")
		}
		for i, set := range sets {
			path :=	utils.BfsPath(g.Nodes, id, set[0], false, nil)
			if path != nil {
				find = true
				sets[i] = append(sets[i], id)
				break
			}
		}
		if !find {
			newSet := []utils.RouterID{id}
			sets = append(sets, newSet)
		}
	}

	maxLen := 0
	index := 0
	for i := range sets {
		if len(sets[i]) > maxLen {
			index = i
		}
	}
	return sets[index]
}

func GetNotConnectedNodes(g *utils.Graph)  map[utils.RouterID]struct{} {
	part1 := make(map[utils.RouterID]struct{})
	partOther := make(map[utils.RouterID]struct{})
	for id := range g.Nodes {
		path  := utils.BfsPath(g.Nodes, 4, id,false, g.Channels)
		if path == nil || len(path) == 0 {
			partOther[id] = struct{}{}
		} else {
			part1[id] = struct{}{}
		}
	}
	if len(partOther) > len(part1) {
		return part1
	} else {
		return partOther
	}
}


func RemoveNotConnectNodes(g *utils.Graph, toRemove map[utils.RouterID]struct{}) {
	for id := range g.Nodes {
		if _, ok := toRemove[id]; ok {
			delete(g.Nodes, id)
		}
	}
	fmt.Printf("remove node done\n")
	for _, node := range g.Nodes {
		for nToD := range toRemove {
			node.RemoveNei(nToD)
		}
	}
}

// 使用提前设置好的map作为id映射的依据
func ConvertToSeriesIDWithMap(balanceDistiWay int, g *utils.Graph,
	IDMap map[utils.RouterID]utils.RouterID,
	chanValue utils.Amount) map[utils.RouterID]utils.RouterID {

	finalNodes := make(map[utils.RouterID]*utils.Node)
	for orgID, node := range g.Nodes {
		finalNodes[IDMap[orgID]] = node
	}

	for id, node := range finalNodes {
		newNeis := map[utils.RouterID]struct{}{}
		for neiID := range node.Neighbours {
			newNeis[IDMap[neiID]] = struct{}{}
		}
		node.Neighbours = newNeis
		node.ID = id
	}

	rand.Seed(time.Now().UnixNano())
	//Convert channels
	channels := make(map[string]*utils.Link)
	for _, link := range g.Channels {
		if _, ok := g.Nodes[link.Part1]; !ok {
			continue
		}
		if _, ok := g.Nodes[link.Part2]; !ok {
			continue
		}

		mapped1 := IDMap[link.Part1]
		mapped2 := IDMap[link.Part2]
		linkKey := utils.GetLinkKey(mapped1, mapped2)
		linkAvgValue := (link.Val2 + link.Val1) / 2
		linkValue := utils.Amount(rand.Float64()) * (link.Val2 + link.Val1)
		var newLink *utils.Link
		if mapped1 < mapped2 {
			newLink = &utils.Link{
				Part1: mapped1,
				Part2: mapped2,
			}
			switch balanceDistiWay {
			case ORIGION_CHANNEL:
				newLink.Val1 = link.Val1
				newLink.Val2 = link.Val2
			case REBALANCE_CHANEL:
				newLink.Val1 = linkAvgValue
				newLink.Val2 = linkAvgValue
			case UNIFORMLY_CHANNEL:
				newLink.Val1 = linkValue
				newLink.Val2 = link.Val2 + link.Val1 - linkValue
			case FIX_VALUE_CHANNEL:
				newLink.Val1 = chanValue/2
				newLink.Val2 = chanValue/2
			}
		} else {

			newLink = &utils.Link{
				Part1: mapped2,
				Part2: mapped1,
			}
			switch balanceDistiWay {
			case ORIGION_CHANNEL:
				newLink.Val1 = link.Val2
				newLink.Val2 = link.Val1
			case REBALANCE_CHANEL:
				newLink.Val1 = linkAvgValue
				newLink.Val2 = linkAvgValue
			case UNIFORMLY_CHANNEL:
				newLink.Val1 = linkValue
				newLink.Val2 = link.Val2 + link.Val1 - linkValue
			case FIX_VALUE_CHANNEL:
				newLink.Val1 = chanValue/2
				newLink.Val2 = chanValue/2
			}
		}
		channels[linkKey] = newLink
	}
	g.Channels = channels
	g.Nodes = finalNodes
	return IDMap
}

func ConvertToSeriesID(balanceDistiWay int, g *utils.Graph,
	chanValue utils.Amount) map[utils.RouterID]utils.RouterID {
	i := utils.RouterID(0)
	IDMap := make(map[utils.RouterID]utils.RouterID)
	finalNodes := make(map[utils.RouterID]*utils.Node)
	for orgID, node := range g.Nodes {
		IDMap[orgID] = i
		finalNodes[i] = node
		i++
	}

	for id, node := range finalNodes {
		newNeis := map[utils.RouterID]struct{}{}
		for neiID := range node.Neighbours {
			newNeis[IDMap[neiID]] = struct{}{}
		}
		node.Neighbours = newNeis
		node.ID = id
	}

	rand.Seed(time.Now().UnixNano())
	//Convert channels
	channels := make(map[string]*utils.Link)
	for _, link := range g.Channels {
		if _, ok := g.Nodes[link.Part1]; !ok {
			continue
		}
		if _, ok := g.Nodes[link.Part2]; !ok {
			continue
		}

		mapped1 := IDMap[link.Part1]
		mapped2 := IDMap[link.Part2]
		linkKey := utils.GetLinkKey(mapped1, mapped2)
		linkAvgValue := (link.Val2 + link.Val1) / 2
		linkValue := utils.Amount(rand.Float64()) * (link.Val2 + link.Val1)
		var newLink *utils.Link
		if mapped1 < mapped2 {
			newLink = &utils.Link{
				Part1: mapped1,
				Part2: mapped2,
			}
			switch balanceDistiWay {
			case ORIGION_CHANNEL:
				newLink.Val1 = link.Val1
				newLink.Val2 = link.Val2
			case REBALANCE_CHANEL:
				newLink.Val1 = linkAvgValue
				newLink.Val2 = linkAvgValue
			case UNIFORMLY_CHANNEL:
				newLink.Val1 = linkValue
				newLink.Val2 = link.Val2 + link.Val1 - linkValue
			case FIX_VALUE_CHANNEL:
				newLink.Val1 = chanValue/2
				newLink.Val2 = chanValue/2
			}
		} else {

			newLink = &utils.Link{
				Part1: mapped2,
				Part2: mapped1,
			}
			switch balanceDistiWay {
			case ORIGION_CHANNEL:
				newLink.Val1 = link.Val2
				newLink.Val2 = link.Val1
			case REBALANCE_CHANEL:
				newLink.Val1 = linkAvgValue
				newLink.Val2 = linkAvgValue
			case UNIFORMLY_CHANNEL:
				newLink.Val1 = linkValue
				newLink.Val2 = link.Val2 + link.Val1 - linkValue
			case FIX_VALUE_CHANNEL:
				newLink.Val1 = chanValue/2
				newLink.Val2 = chanValue/2
			}
		}
		channels[linkKey] = newLink
	}
	g.Channels = channels
	g.Nodes = finalNodes
	return IDMap
}

// 从100W条交易中，根据src和dest采样出transNum条数据
func RandomRippleTrans(trans []utils.Tran, IDMap map[utils.RouterID]utils.RouterID, transNum int,
	geneWay int, cutMaxMin bool) []utils.Tran {
	resTrans := make([]utils.Tran, 0)
	rand.Seed(time.Now().UnixNano())

	min, max := float64(0), float64(0)
	if cutMaxMin == true {
		values := make([]float64,0)
		for _, tran := range trans{
			values = append(values, tran.Val)
		}
		sort.Float64s(values)
		min = values[len(values)/10*1]
		max = values[len(values)/10*9]
	}

	for i := 0; len(resTrans) < transNum; i++ {
		tran := trans[rand.Intn(len(trans))]
		if cutMaxMin == true {
			if tran.Val > max || tran.Val < min {
				continue
			}
		}
		if geneWay == REMAINDER_SAMPLE {
			newTran := utils.Tran{
				Src:  tran.Src % len(IDMap),
				Dest: tran.Dest % len(IDMap),
				Val:  tran.Val,
			}
			resTrans = append(resTrans, newTran)
		} else if geneWay == MAP_SAMPLE {
			if _, ok := IDMap[utils.RouterID(tran.Src)]; !ok {
				continue
			}
			if _, ok := IDMap[utils.RouterID(tran.Dest)]; !ok {
				continue
			}
			newTran := utils.Tran{
				Src:  int(IDMap[utils.RouterID(tran.Src)]),
				Dest: int(IDMap[utils.RouterID(tran.Dest)]),
				Val:  tran.Val,
			}
			resTrans = append(resTrans, newTran)
		}
	}
	return resTrans
}
