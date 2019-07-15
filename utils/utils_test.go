package utils

import (
	"testing"
)

func TestSampleTrans(t *testing.T) {
	trans, err := SampleTrans("../data/finalSets/static/", 5000)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(trans))
}

func TestGraph_LoadDistances(t *testing.T) {

	s := []int{1}
	a := s[0:1]
	t.Log(a)
}

func TestGraph_StoreDistances(t *testing.T) {
	g, _ := ParseTestJson("../data/ten_nodes.json")
	err := g.StoreDistances("../data/finalSets/static/dis",1)
	if err != nil {
		t.Fatal(err)
	}
	dis := g.Distance
	t.Log(dis)
	err = g.LoadDistances("../data/finalSets/static/dis", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(g.Distance)
}

func TestStoreRipple(t *testing.T) {
	g := GetGraph("../data")
	err := g.StoreDistances("../data/finalSets/static/dis_ripple",1)
	if err != nil {
		t.Fatal(err)
	}
}



