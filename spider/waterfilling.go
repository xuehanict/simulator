package spider

import (
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
)

func (s *Spider) waterFilling (amt utils.Amount, caps []utils.Amount,
	) ([]utils.Amount, error) {

	sum := utils.Amount(0)
	for _, cap := range caps {
		sum += cap
	}
	if sum < amt {
		return nil, fmt.Errorf("capcities is insufficient")
	}

	remainder := amt
	res := make([]utils.Amount, len(caps))

	for ; remainder > 0; {
		gap, maxSet, _ := getGap(caps)
		amtAssign := utils.Amount(0)
		if remainder > utils.Amount(len(maxSet))*gap {
			amtAssign = gap
		} else {
			amtAssign = remainder / utils.Amount(len(maxSet))
		}
		if gap == 0 {
			amtAssign = remainder / utils.Amount(len(caps))
		}

		for _, index := range maxSet {
			res[index] += amtAssign
			caps[index] -= amtAssign
		}
		remainder -= amtAssign * utils.Amount(len(maxSet))
	}
	return res, nil
}


func (s *Spider) getPaths(src, dest utils.RouterID, k int) []utils.Path {
	nodes := utils.CopyNodes(s.Nodes)
	paths := make([]utils.Path,0)

	for i := 0; i < k; i++ {
		path := utils.BfsPath(nodes, src, dest, false, nil)
		if path == nil {
			break
		}
		if len(path) < 2 {
			break
		}
		nodes = clearEdge(nodes, path)
		paths = append(paths, path)
	}
	return paths
}

func clearEdge(nodes map[utils.RouterID]*utils.Node, path utils.Path) map[utils.RouterID]*utils.Node {
	nodes[path[0]].RemoveNei(nodes[path[1]].ID)
	for i := 1; i < len(path) - 1; i++ {
		nodes[path[i]].RemoveNei(nodes[path[i+1]].ID)
		nodes[path[i+1]].RemoveNei(nodes[path[i]].ID)
	}
	return nodes
}

func getGap (caps []utils.Amount) (utils.Amount, []int, []int){
	largest := utils.Amount(-1)
	second := utils.Amount(-1)
	maxSet := make([]int, 0)
	secSet := make([]int, 0)

	for _, cap := range caps {
		if cap>largest {
			second = largest
			largest = cap
		} else if cap > second && cap != largest {
			second = cap
		}
	}

	for i, cap := range caps{
		if cap == largest {
			maxSet = append(maxSet, i)
		}
		if cap == second {
			secSet = append(secSet, i)
		}
	}

	if second == -1 {
		return 0, maxSet, secSet
	}
	return largest - second, maxSet, secSet
}


