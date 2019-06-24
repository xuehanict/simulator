package landmark

import (
	"github.com/lightningnetwork/simulator/utils"
	"testing"
)

func TestSM_SendPayment(t *testing.T) {
	graph, err := utils.ParseTestJson(tenNodesGraph)
	if err != nil {
		t.Fatalf("faced error:%v", err)
	}
	s := NewSM(graph, []utils.RouterID{2,7})
	_, err = s.SendPayment( 9, 6, 30)
	if err != nil {
		t.Fatalf("faced error :%v", err)
	}
}

