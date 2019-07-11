package mpdv

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/lightningnetwork/simulator/utils"
	"testing"
)

func TestMpdv_SendPayment(t *testing.T) {
	g, err := utils.ParseTestJson("../data/ten_nodes_complex.json")
	if err != nil {
		t.Fatal(err)
	}

	m := NewMpdv(g, 100, 0.1)
	m.InitTable(map[utils.RouterID]struct{}{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
		6: {},
		7: {},
		8: {},
	})
	metric, err := m.SendPayment(20, 1, 4)
	spew.Dump(metric)
	if err != nil {
		t.Fatal(err)
	}
}
