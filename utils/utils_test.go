package utils

import "testing"

func TestSampleTrans(t *testing.T) {
	trans, err := SampleTrans("../data/finalSets/static/", 5000)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(trans))
}

