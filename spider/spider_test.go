package spider

import (
	"github.com/lightningnetwork/simulator/utils"
	"testing"
)

const (
	tenNodesGraph        = "../data/ten_nodes.json"
	tenNodesGraphComplex = "../data/ten_nodes_complex.json"
	tenNodesGraphHalf    = "../data/ten_nodes_half.json"
)

func TestSpider_SendPayment(t *testing.T) {
	graph, err := utils.ParseTestJson(tenNodesGraph)
	if err != nil {
		t.Fatalf("faced error:%v", err)
	}
	spider := NewSpider(graph, WATERFIILING, 4)
	err = spider.SendPayment(2, 7, 30)
	if err != nil {
		t.Fatalf("faced error:%v", err)
	}
}


