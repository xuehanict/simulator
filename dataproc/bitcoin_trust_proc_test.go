package dataproc

import (
	"strings"
	"testing"
)

func TestGetBitCoinTrustGraph(t *testing.T) {
	fileName := "../data/bitcoin_trust/soc-sign-bitcoinalpha.csv"
	graph, err := GetBitCoinTrustGraph(fileName)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(graph.Nodes))
	t.Log(len(graph.Channels))
}

func TestNodeNumber(t *testing.T)  {
	fileName := "../data/bitcoin_trust/soc-sign-bitcoinalpha.csv"
	strs, _ := ReadBitcoinTrustFile(fileName)

	nodeSet := make(map[string]struct{})
	for _, str := range strs {
		node1, node2 := strings.Split(str,",")[0], strings.Split(str,",")[1]
		nodeSet[node1] = struct{}{}
		nodeSet[node2] = struct{}{}
	}

	t.Log(len(nodeSet))
}



