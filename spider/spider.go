package spider

import (
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
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
	amt utils.Amount) (*utils.Metrics, error) {

	metric := &utils.Metrics{0,0,0,0}
	switch s.algo {
	case WATERFIILING:
		paths := s.getPaths(src,dest, s.pathNum)
		//spew.Dump(paths)
		routeMins := make([]utils.Amount, len(paths))
		maxLength := 0
		// 计算出每条路径的最小值，并且获取每条通道的容量
		for j, path := range paths {
			routeMins[j] = utils.GetPathCap(path, s.Channels)
			metric.ProbeMessgeNum += int64(len(path)-1)
		}
		distri, err := s.waterFilling(amt, routeMins)
		if err != nil {
			return metric,fmt.Errorf("insufficient")
		}
		err = s.UpdateWeights(paths, distri)
		if err != nil {
			return metric,err
		}

		// update the metrics
		for i, amt := range distri {
			if amt != 0 {
				metric.OperationNum += int64(len(paths[i]) - 1)
				metric.Fees += s.GetFee(paths[i],distri[i])
				if len(paths[i]) > maxLength {
					maxLength = len(paths[i])
				}
			}
		}
		metric.MaxPathLengh = maxLength
		return metric, nil
	}
	return metric, nil
}

func NewSpider(g *utils.Graph, algo, pathNum int) *Spider {
	return &Spider{
		Graph: g,
		algo: algo,
		pathNum:pathNum,
	}
}

