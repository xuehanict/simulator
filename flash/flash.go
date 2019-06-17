package flash

import (
	"github.com/lightningnetwork/simulator/utils"
	"github.com/starwander/goraph"
	"sort"
)

type Flash struct {
	*utils.Graph
	goGraph *goraph.Graph
	routingTable map[utils.RouterID]map[utils.RouterID][]utils.Path
	pathN int
}

func getThreshold(trans []utils.Tran, percent float64) float64 {
	amts := make([]float64,0)
	for _, tran := range trans {
		amts = append(amts, tran.Val)
	}
	sort.Float64s(amts)
	return amts[int((percent)*float64(len(trans)))]
}

func (f *Flash)SendPayment(amt, thredhold utils.Amount, from, to utils.RouterID) error {
	var err error
	if amt > thredhold {
		_, err = f.elephantRouting(amt, from,to)
	} else {
		_, err = f.microRouting(from, to, amt, 4)
	}
	return err
}

func NewFlash(graph *utils.Graph,pathN int) *Flash  {
	flash := &Flash{
		Graph: graph,
		routingTable: make(map[utils.RouterID]map[utils.RouterID][]utils.Path),
		pathN: pathN,
	}
	flash.goGraph, _ = flash.convertGraph()
	return flash
}
