package landmark

import (
	"bytes"
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"math"
)

type SM struct {
	*LandMarkRouting
}


func (s *SM) getPaths (src, dest utils.RouterID, amts []utils.Amount) (
	[]utils.Path, error) {

	paths := make([]utils.Path, 0)
	for i, root := range s.Roots {
		path := utils.Path{src}
		curr := src
		for {
			next := s.getNextHop(curr, root, dest, path, amts[i] )
			// 没找到路需要回滚
			if next == -1  {
				err := s.UpdateWeighOneDir(path, amts[i], utils.ADD)
				if err != nil {
					return nil, fmt.Errorf("cann't find nexthop and faced" +
						"err:%v when rollback", err)
				} else {
					return nil, fmt.Errorf("cann't find nextHop")
				}
			}
			err := s.UpdateLinkValue(curr, next, amts[i], utils.SUB)
			if err != nil {
				return nil, err
			}
			curr = next
			path = append(path, next)
			if next == dest {
				paths = append(paths, path)
				break
			}
		}
	}
	return paths, nil
}

func (s *SM) getNextHop (currID, root, dest utils.RouterID,
	currPath utils.Path, amt utils.Amount) utils.RouterID {
	destAddr := s.Coordination[dest][root].coordinate
	minDis := math.MaxInt32
	minNeighbour := utils.RouterID(-1)
	for _, nei := range s.Nodes[currID].Neighbours {
		dis := getDis(s.Coordination[nei][root].coordinate, destAddr, AddrInterval)
		if dis < minDis && utils.GetLinkValue(currID, nei, s.Channels) >= amt &&
			!utils.CheckInPath(nei, currPath ) {
			minNeighbour = nei
			minDis = dis
		}
	}
	return minNeighbour
}

// 用到的计算距离的方式
func getDis(neighbour, dest string, lenthInterval int) int {
	depthN := len(neighbour) / lenthInterval
	depthD := len(dest) / lenthInterval
	cpl := 0
	neiBytes := []byte(neighbour)
	destBytes := []byte(dest)

	for ; cpl < depthD && cpl < depthN &&
		bytes.Equal(neiBytes[0:lenthInterval],
			destBytes[0:lenthInterval]); cpl++ {
		neiBytes = neiBytes[lenthInterval:]
		destBytes = destBytes[lenthInterval:]
	}
	return depthN + depthD - (2 * cpl)
}

func (s *SM) SendPayment (src, dest utils.RouterID, amt utils.Amount) (
	bool, error) {
	splittedAmounts := randomPartition(amt, len(s.Roots))
	paths,err := s.getPaths(src,dest, splittedAmounts)
	if err != nil {
		return false, err
	}
	if len(paths) == 0 {
		return false, fmt.Errorf("no path found")
	}
//	spew.Dump(s.Channels)
	// 因为在探路过程中已经减去了过去的部分钱，所以先加回来，再支付
	for i, path := range paths {
		err := s.UpdateWeighOneDir(path, splittedAmounts[i], utils.ADD)
		if err != nil {
			return false, fmt.Errorf("探路完后失败")
		}
		// 这里才是真正支付
		err = s.UpdateWeight(path, splittedAmounts[i])
		if err != nil {
			return false, fmt.Errorf("探路完支付失败")
		}
	}
	return true, nil
}

func NewSM(g *utils.Graph, roots []utils.RouterID) *SM  {
	sm := &SM{
		LandMarkRouting: &LandMarkRouting{
			Graph: g,
			Coordination:make(map[utils.RouterID]map[utils.RouterID]*Addr),
			Roots: roots,
		},
	}

	for _,n := range sm.Nodes {
		sm.Coordination[n.ID] = make(map[utils.RouterID]*Addr)
	}

	sm.SetCoordinations()
	return sm
}


