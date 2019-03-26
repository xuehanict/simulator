package main

import (
	"io/ioutil"
	"fmt"
	"os"
	"encoding/json"
	"time"
	sw "github.com/lightningnetwork/sm/silentWhisper"
	"sync"
	"bufio"
	"io"
	"strings"
	"strconv"
)

func main()  {
//	testSW()
	testSWBigData()
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

	roots := []sw.RouteID{3,8,9}
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
		//fmt.Printf("router %v nei is %v\n", r.ID, r.Neighbours)
		//time.Sleep(1 * time.Second)
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

	fmt.Printf("send result%v\n",nodes[2].SendPayment(1, 100))

	for _,node := range nodes{
		node.Stop()
		wg.Done()
	}
	wg.Wait()
}

func testSWBigData() {
	f, err := os.Open("data/finalSets/static/ripple-lcc.graph_CREDIT_LINKS")
	if err != nil {
		fmt.Println("os Open error: ", err)
		return
	}
	defer f.Close()

	br := bufio.NewReader(f)
	lineNum := 1
	links := make(map[string]*sw.Link,0)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("br ReadLine error: ", err)
			return
		}
		//
		if lineNum < 5 {
			lineNum ++
			continue
		}

		splitted := strings.Split(string(line), " ")
		id1, _ := strconv.Atoi(splitted[0])
		id2, _ := strconv.Atoi(splitted[1])
		v1, _ := strconv.ParseFloat(splitted[2],64)
		v2,_ := strconv.ParseFloat(splitted[3], 64)
		v3,_ := strconv.ParseFloat(splitted[4], 64)
		link := &sw.Link{
			Part1: sw.RouteID(id1),
			Part2: sw.RouteID(id2),
			Val1: v2 - v1,
			Val2: v3 - v2,
		}
		links[sw.GetLinkKey(link.Part1,link.Part2)] = link
	}

	roots := []sw.RouteID{43788,59333, 100, 500 }
	nodes := make(map[sw.RouteID]*sw.SWRouter, 0)
	for i:=0; i<67149; i++ {
		router := sw.NewSwRouter(sw.RouteID(i),roots, nodes, links)
		nodes[sw.RouteID(i)] = router
	}

	for _, edge := range links {
		nodes[edge.Part1].Neighbours[edge.Part2] = struct{}{}
		nodes[edge.Part2].Neighbours[edge.Part1] = struct{}{}
	}
	wg := sync.WaitGroup{}
	for _,r := range nodes {
		//wg.Add(1)
		go r.Start()
		fmt.Printf("router %v start\n", r.ID)
		//fmt.Printf("router %v nei is %v\n", r.ID, r.Neighbours)
		//time.Sleep(1 * time.Millisecond)
	}

	time.Sleep(3 * time.Second)
	sw.NotifyRooterReset(roots, nodes)

	for i := 0 ;i< 2; i++ {
		time.Sleep(1 * time.Second)
		fmt.Printf("wait 1s\n")
	}

	trans := generateTrans("data/finalSets/static/sampleTr-0.txt")
	total := 0
	success := 0
	for _, tran := range trans{
		total ++
		err := nodes[sw.RouteID(tran.src)].SendPayment(sw.RouteID(tran.dest), tran.val)
		if err == nil {
			success++
		}
		fmt.Printf("err :%v\n", err)
		fmt.Printf("total:%v\n", total)
	}

	fmt.Printf("total :%v\n", total)
	fmt.Printf("success:%v\n", success)


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


func generateTrans (filePath string) []tran {

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("os Open error: ", err)
		return nil
	}
	defer f.Close()

	br := bufio.NewReader(f)
	trans := make([]tran,0)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("br ReadLine error: ", err)
			return nil
		}
		splitStr := strings.Split(string(line), " ")
		val, _ := strconv.ParseFloat(splitStr[0], 64)
		src, _ := strconv.Atoi(splitStr[1])
		dest, _:= strconv.Atoi(splitStr[2])

		trans = append(trans,tran{
			src: src,
			dest: dest,
			val: val,
		})
	}

	return trans
}


type tran struct {
	src int
	dest int
	val float64
}




