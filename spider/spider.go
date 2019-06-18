package spider

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
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
	amt utils.Amount) error {
	switch s.algo {
	case WATERFIILING:
		paths := s.getPaths(src,dest, s.pathNum)
		//spew.Dump(paths)
		routeMins := make([]utils.Amount, len(paths))
		// 计算出每条路径的最小值，并且获取每条通道的容量
		for j, path := range paths {
			routeMins[j] = utils.GetPathCap(path, s.Channels)
		}

		distri, err := s.waterFilling(amt, routeMins)
		if err != nil {
			return fmt.Errorf("insufficient")
		}
		spew.Dump(distri)
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

