package spider

import (
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"math"
)

const (
	LP  = iota
	WATERFIILING
)

type Spider struct {
	name  string
	*utils.Graph
	algo  int
	pathNum int
}


func (s *Spider) SendPayment (src, dest utils.RouterID,
	amt utils.Amount, algo int) error {
	switch algo {
	case WATERFIILING:
		paths := s.getPaths(src,dest, s.pathNum)
		routeMins := make([]utils.Amount, 0)
		// 计算出每条路径的最小值，并且获取每条通道的容量
		for j, path := range paths {
			min := utils.Amount(math.MaxFloat64)
			for i := 0; i < len(path)-1; i++ {
				val := utils.GetLinkValue(path[i], path[i+1], s.Channels)
				if val < min {
					min = val
				}
			}
			routeMins[j] = min
		}

		distri, err := s.waterFilling(amt, routeMins)
		if err != nil {
			return fmt.Errorf("insufficient")
		}
		
		err = s.UpdateWeights(paths, distri)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func NewSpider(g *utils.Graph, algo, pathNum int) *Spider {
	return &Spider{
		Graph: g,
		algo: algo,
		pathNum:pathNum,
	}
}

