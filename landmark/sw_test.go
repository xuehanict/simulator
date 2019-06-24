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