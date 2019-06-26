package flash

import (
	"github.com/lightningnetwork/simulator/utils"
	"github.com/starwander/goraph"
	"sort"
	"sync"
)

type Flash struct {
	*utils.Graph
	goGraph *goraph.Graph
	routingTable map[utils.RouterID]map[utils.RouterID][]utils.Path
	pathN int

	testTable map[utils.RouterID]map[utils.RouterID][]utils.Path
	test bool
	rw sync.RWMutex
}

func getThreshold(trans []utils.Tran, percent float64) float64 {
	amts := make([]float64,0)
	for _, tran := range trans {
		amts = append(amts, tran.Val)
	}
	sort.Float64s(amts)
	return amts[int((percent)*float64(len(trans)))]
}

func (f *Flash)SendPayment(amt, thredhold utils.Amount, from, to utils.RouterID) (
	*utils.Metrics, error) {
	var err error
	var metric *utils.Metrics
	if amt > thredhold {
		metric, err = f.elephantRouting(amt, from,to)
	} else {
		metric, err = f.microRouting(from, to, amt, 4)
	}
	return metric, err
}

func NewFlash(graph *utils.Graph,pathN int, test bool) *Flash  {
	flash := &Flash{
		Graph: graph,
		routingTable: make(map[utils.RouterID]map[utils.RouterID][]utils.Path),
		testTable: make(map[utils.RouterID]map[utils.RouterID][]utils.Path),
		pathN: pathN,
		test: test,
		rw : sync.RWMutex{},
	}
	flash.goGraph, _ = flash.convertGraph()
	return flash
}
