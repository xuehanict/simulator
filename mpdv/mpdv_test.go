package mpdv

import (
	"github.com/lightningnetwork/simulator/utils"
	"testing"
)

func TestMpdv_SendPayment(t *testing.T) {
	g, err := utils.ParseTestJson("../data/ten_nodes_complex.json")
	if err != nil {
		t.Fatal(err)
	}

	m := NewMpdv(g, 100, 0.1)
	m.initTable([]utils.RouterID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	_, err = m.SendPayment(20, 2, 7)
	if err != nil {
		t.Fatal(err)
	}
}
