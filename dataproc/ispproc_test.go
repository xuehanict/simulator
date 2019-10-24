package dataproc

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestGetGraph(t *testing.T) {
	fileName := "../data/isp/uunet.csv"
	g, err := GetISPGraph(fileName)
	if err != nil {
		t.Log(err)
	}
	spew.Dump(len(g.Nodes))
	spew.Dump(len(g.Channels))
}



