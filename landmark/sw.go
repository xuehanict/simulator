package landmark

import (
	"bytes"
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
)

type SW struct {
	*LandMarkRouting
}

const (
	UP  = true
	DOWN = false
)

func (s *SW) getPaths (src, dest utils.RouterID, metric *utils.Metrics,
	) []utils.Path {
	paths := make([]utils.Path, 0)
	for _, root := range s.Roots {
		destAddr := s.Coordination[dest][root].coordinate
		path := make(utils.Path, 1)
		curr := src
		path[0] = src
		dir := UP
		if src == root {
			dir = DOWN
		}

		for {
			metric.ProbeMessgeNum++
			next := s.getNextHop(destAddr, root, curr, dir)
			path = append(path, next)
			curr = next
			if next == dest {
				break
			}
			if next == root {
				dir = DOWN
			}
		}
		paths = append(paths, path)
	}
	return paths
}

func minPart(amount utils.Amount, mins []utils.Amount) []utils.Amount {
	remainder := amount
	saturated := make(map[int]struct{})
	res := randomPartition(amount, len(mins))

	if len(res) == 0 {
		return nil
	}
	for remainder > 0 {
		remainder = 0
		for i := 0; i < len(res); i++ {
			if res[i] > mins[i] {
				remainder = remainder + res[i] - mins[i]
				res[i] = mins[i]
				saturated[i] = struct{}{}
			}
		}

		if len(saturated) == len(mins) {
			return nil
		} else {
			if remainder > 0 {
				adds := randomPartition(remainder, len(mins)-len(saturated))
				k := 0
				for i := 0; i < len(adds); i++ {
					for {
						if _, ok := saturated[k]; ok {
							k++
						} else {
							break
						}
					}
					res[k] = res[k] + adds[i]
					k++
				}
			}
		}
	}
	return res
}

func (s *SW) SendPayment (src, dest utils.RouterID, amt utils.Amount) (
	*utils.Metrics, error) {
	metric := &utils.Metrics{0,0,0,0}
	paths := s.getPaths(src, dest, metric)
//	spew.Dump(paths)
	caps := make([]utils.Amount, 0)
	for _, path := range paths {
		cap := utils.GetPathCap(path, s.Channels)
		caps = append(caps, cap)
	}

	allcs := minPart(amt,caps)
	if allcs == nil {
		return metric, fmt.Errorf("capacity not insufficient")
	}

	sentList := make([]utils.Amount, 0)
	sentPaths := make([]utils.Path, 0)
	for i, path := range paths {
		if utils.GetPathCap(path, s.Channels) >= allcs[i] {
			err := s.UpdateWeight(paths[i], allcs[i])
			if err != nil {
				return metric, err
			}
			sentList = append(sentList, allcs[i])
			sentPaths = append(sentPaths, path)

		} else {
			// 回滚
			for j := range sentPaths {
				err := s.UpdateWeightReverse(sentPaths[j], sentList[j])
				if err != nil {
					return metric, err
				}
			}
			return metric, nil
		}
		metric.Fees += utils.Amount(len(path)-1)*utils.FEERATE*allcs[i]
		metric.OperationNum += int64(len(path)-1)
		if len(path) > metric.MaxPathLengh {
			metric.MaxPathLengh = len(path)
		}
	}
	return metric, nil
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

func NewSw(g *utils.Graph, roots []utils.RouterID) *SW {
	sw := &SW{
		LandMarkRouting: &LandMarkRouting{
			Graph: g,
			Coordination:make(map[utils.RouterID]map[utils.RouterID]*Addr),
			Roots: roots,
		},
	}
	for _,n := range sw.Nodes {
		sw.Coordination[n.ID] = make(map[utils.RouterID]*Addr)
	}
	sw.SetCoordinations()
	return sw
}

