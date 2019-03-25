package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"encoding/json"
	sw "github.com/lightningnetwork/sm/silentWhisper"
	"time"
)

func main()  {
	var (
		nodes= make(map[sw.RouteID]*sw.SWRouter, 0)
		edges= make(map[string]*sw.Link, 0)
	)
	graphJson, err := ioutil.ReadFile(tenNodesGraph)
	if err != nil {
		fmt.Printf("can't open the json file: %v", err)
		os.Exit(1)
	}

	var g testGraph
	if err := json.Unmarshal(graphJson, &g); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	roots := []sw.RouteID{3}
	for _, node := range g.Nodes{
		router := sw.NewSwRouter(node.Id,roots, nodes, edges)
		nodes[node.Id] = router
	}

	for _, edge := range g.Edges {
		link := &sw.Link{
			Part1: edge.Node1,
			Part2: edge.Node2,
			Val1: edge.Capacity1,
			Val2: edge.Capacity2,
		}
		linkKey := sw.GetLinkKey(edge.Node1,edge.Node2)
		edges[linkKey] = link

		nodes[edge.Node1].Neighbours[edge.Node2] = struct{}{}
		nodes[edge.Node2].Neighbours[edge.Node1] = struct{}{}
	}

	for _,r := range nodes {
		go r.Start()
		//fmt.Printf("router %v start\n", r.ID)
		fmt.Printf("router %v nei is %v\n", r.ID, r.Neighbours)
		time.Sleep(1 * time.Second)
	}

	sw.NotifyRooterReset(roots, nodes)

	for i := 0 ;i< 5; i++ {
		time.Sleep(1 * time.Second)
		fmt.Printf("wait 1s\n")
	}
	for _, node := range nodes {
		fmt.Printf("node %v addr is %v\n", node.ID, node.AddrWithRoots)
	}

}
