package flash

import (
	"github.com/lightningnetwork/simulator/utils"
	"testing"
)

const tenNodesGraph        = "../data/ten_nodes.json"

func TestFlash_SendPayment(t *testing.T) {
	graph, err := utils.ParseTestJson(tenNodesGraph)
	if err != nil {
		t.Fatalf("faced error:%v", err)
	}
	f := NewFlash(graph, 4)
	err = f.SendPayment(40, 50, 1, 4)
	if err != nil {
		t.Fatalf("faced error :%v", err)
	}
	//t.Fatalf("124")
}
