package landmark

import (
	"github.com/lightningnetwork/simulator/utils"
	"testing"
)

const tenNodesGraph        = "../data/ten_nodes.json"

func TestSW_SendPayment(t *testing.T) {
	graph, err := utils.ParseTestJson(tenNodesGraph)
	if err != nil {
		t.Fatalf("faced error:%v", err)
	}
	s := NewSw(graph, []utils.RouterID{2,7})
	_, err = s.SendPayment( 1, 5, 30)
	if err != nil {
		t.Fatalf("faced error :%v", err)
	}
}

func TestSW_Ripple(t *testing.T)  {
	g := utils.GetGraph("../data")
	s := NewSw(g, []utils.RouterID{5, 38, 13})
	_, err := s.SendPayment(38, 12805, 108484.5)
	if err != nil {
		t.Fatalf("faced error :%v", err)
	}
}
