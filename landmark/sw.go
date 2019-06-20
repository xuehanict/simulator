package landmark

import (
	"bytes"
	"github.com/lightningnetwork/simulator/utils"
)

type SW struct {
	LandMarkRouting
}

const (
	UP  = true
	DOWN = false
)


func (s *SW) getPaths (src, dest utils.RouterID) []utils.Path {
	paths := make([]utils.Path, 0)
	for _, root := range s.Roots {
		destAddr := s.Coordination[dest][root].coordinate
		path := make(utils.Path, 1)
		curr := src
		path[0] = src

		for {
			if curr == dest {
				break
			}




		}






	}

	return paths
}

func (s *SW) SendPayment (src, dest utils.RouterID, amt utils.Amount) (
	bool, error) {

	return false, nil
}

func (s *SW) getNextHop(dest string, root, current utils.RouterID,
	upOrDown bool) utils.RouterID {
	nextHop := utils.RouterID(-1)
	if upOrDown == UP {
		return s.Coordination[current][root].parent
	} else {
		// TODO(xuehan): 这里应该改成从邻居实时pull地址
		bestCpl := getCPL(dest, s.Coordination[current][root].coordinate, AddrInterval)
		for _, n := range s.Nodes[current].Neighbours {
			cpl := getCPL(s.Coordination[n][root].coordinate,
				dest, AddrInterval)
			// 这个地方和模拟器中代码不一样
			if cpl == bestCpl+1 && s.Coordination[n][root].parent == current {
				return n
			}
		}
	}
	return nextHop
}

func getCPL(addr1, addr2 string, interval int) int {
	cpl := 0
	addr1Bytes := []byte(addr1)
	addr2Bytes := []byte(addr2)
	for ; cpl < len(addr1)/interval && cpl < len(addr2)/interval &&
		bytes.Equal(addr1Bytes[0:interval],
			addr2Bytes[0:interval]); cpl++ {
		addr1Bytes = addr1Bytes[interval:]
		addr2Bytes = addr2Bytes[interval:]
	}
	return cpl
}

