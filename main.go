package main

import (
	"io/ioutil"
	"fmt"
	"os"
	"encoding/json"
	"time"
	sw "github.com/lightningnetwork/sm/silentWhisper"
	"sync"
)

func main()  {

	testSW()

}

func testSW()  {
	var (
		nodes= make(map[sw.RouteID]*sw.SWRouter, 0)
		edges= make(map[string]*sw.Link, 0)
		wg = sync.WaitGroup{}
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
		wg.Add(1)
		go r.Start()
		//fmt.Printf("router %v start\n", r.ID)
		fmt.Printf("router %v nei is %v\n", r.ID, r.Neighbours)
		time.Sleep(1 * time.Second)
	}

	sw.NotifyRooterReset(roots, nodes)

	for i := 0 ;i< 2; i++ {
		time.Sleep(1 * time.Second)
		fmt.Printf("wait 1s\n")
	}

	for _, node := range nodes {
		for key, value := range  node.AddrWithRoots{
			node.Printf("root %v address is %v\n", key,value)
		}
	}


	fmt.Printf("send result%v\n",nodes[1].SendPayment(8, 100))


	for {
		fmt.Printf("1111")
		time.Sleep(1000 * time.Second)
	}

	for _,node := range nodes{
		node.Stop()
		wg.Done()
	}
	wg.Wait()
}





