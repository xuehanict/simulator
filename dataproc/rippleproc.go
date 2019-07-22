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
	MAPP_SAMPLE      = 1

	ORIGION_CHANNEL   = 1
	REBALANCE_CHANEL   = 2
	UNIFORMLY_CHANNEL = 3
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

	for _, node := range g.Nodes {
		for nToD := range nodesToDelete {
			node.RemoveNei(nToD)
		}
	}
	return len(nodesToDelete)
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

func ConvertToSeriesID(balanceDistiWay int, g *utils.Graph) map[utils.RouterID]utils.RouterID {
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
			}
		} else {

			newLink = &utils.Link{
				Part1: mapped1,
				Part2: mapped2,
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

		if _, ok := IDMap[utils.RouterID(tran.Src)]; !ok {
			continue
		}
		if _, ok := IDMap[utils.RouterID(tran.Dest)]; !ok {
			continue
		}

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
		} else if geneWay == MAPP_SAMPLE {
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
