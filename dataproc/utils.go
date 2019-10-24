package dataproc

import (
	"github.com/lightningnetwork/simulator/utils"
	"strconv"
	"strings"
)

// 生成Graph，其channel的值为0
func getGraph(linkStrs []string) (*utils.Graph, error) {

	g := &utils.Graph{
		Nodes:    make(map[utils.RouterID]*utils.Node),
		Channels: make(map[string]*utils.Link),
		DAGs:     make(map[utils.RouterID]*utils.DAG),
		SPTs:     make(map[utils.RouterID]*utils.DAG),
		Distance: make(map[utils.RouterID]map[utils.RouterID]float64),
	}
	for _, linkStr := range linkStrs {
		atrri := strings.Split(linkStr, ",")
		id1, _ := strconv.Atoi(atrri[0])
		id2, _ := strconv.Atoi(atrri[1])
		g.AddNode(utils.RouterID(id1))
		g.AddNode(utils.RouterID(id2))
		g.AddLink(utils.RouterID(id1), utils.RouterID(id2))
	}
	return g, nil
}


