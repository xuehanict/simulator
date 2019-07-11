package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	ADD = true
	SUB = false
)


/*
type Link struct {
	From utils.RouterID
	To   utils.RouterID
}
*/

/**
 * Node 结构为Graph和DAG中的节点，在无线图中，只会用到Neighbours数组
 * 在DAG中会用到Parents和Children数组以表示方向
 */
type Node struct {
	ID         RouterID
	Parents    []RouterID
	Children   []RouterID
	Neighbours []RouterID
}

type DAG struct {
	Root    *Node
	Vertexs []*Node
	//edges   []*Link
}

type Graph struct {
	Nodes []*Node
	// 图的channels信息
	Channels map[string]*Link
	// 路由所需要的DAG，路由时是一步步往parents的方向传递路由包和支付请求
	DAGs map[RouterID]*DAG
	// 最短路树，用来构建MC-SPE
	SPTs map[RouterID]*DAG
	// key是dest节点，value为到key的距离，二级索引是src
	Distance map[RouterID]map[RouterID]float64
}

func NewDAG(root *Node, len int) *DAG {
	return &DAG{
		Root:    root,
		Vertexs: make([]*Node, len),
		//edges:   make([]*Link, 0),
	}
}

func (n *Node) CheckLink(id RouterID) bool {
	//fmt.Printf("node id is %v", spew.Sdump(n))
	for _, n := range n.Neighbours {
		if n == id {
			return true
		}
	}
	return false
}

func (n *Node) CheckParent(id RouterID) bool {
	for _, p := range n.Parents {
		if p == id {
			return true
		}
	}
	return false
}

func (n *Node) checkChild(id RouterID) bool {
	for _, c := range n.Children {
		if c == id {
			return true
		}
	}
	return false
}

func CopyNodes(src []*Node) []*Node {
	res := make([]*Node, len(src))
	for id, node := range src {
		n := &Node{
			ID:         RouterID(id),
			Neighbours: node.Neighbours,
			Children:   node.Children,
			Parents:    node.Parents,
		}
		res[id] = n
	}
	return res
}

func (n *Node) RemoveNei(id RouterID) {
	newNeis := make([]RouterID, 0)
	for _, nei := range n.Neighbours {
		if nei != id {
			newNeis = append(newNeis, nei)
		}
	}
	n.Neighbours = newNeis
}

// 支付多条路径
func (g *Graph) UpdateWeights(routes []Path,
	amts []Amount) error {

	if len(routes) != len(amts) {
		return fmt.Errorf("routes number is not equal to amts' ")
	}

	for idx, route := range routes {
		for i := 0; i < len(route)-1; i++ {
			// i 到 i+1 的钱减少
			err := UpdateLinkValue(route[i], route[i+1],
				g.Channels, amts[idx], SUB)
			if err != nil {
				return err
			}
			// i+1 到 i 的钱增加
			err = UpdateLinkValue(route[i+1], route[i],
				g.Channels, amts[idx], ADD)
			if err != nil {
				return err
			}
		}
	}
	return nil
}


// 支付一条路径，i -> i+1 的钱减少， i+1 -> i的钱增加
func (g *Graph)UpdateWeight(route Path, amt Amount) error {
	for i := 0; i < len(route)-1; i++ {
		// i 到 i+1 的钱减少
		err := UpdateLinkValue(route[i], route[i+1],
			g.Channels, amt, SUB)
		if err != nil {
			return err
		}
		// i+1 到 i 的钱增加
		err = UpdateLinkValue(route[i+1], route[i],
			g.Channels, amt, ADD)
		if err != nil {
			return err
		}
	}
	return nil
}

// 支付一条路径，i -> i+1 的钱增加， i+1 -> i的钱减少
func (g *Graph)UpdateWeightReverse(route Path, amt Amount) error {
	for i := 0; i < len(route)-1; i++ {
		// i 到 i+1 的钱减少
		err := UpdateLinkValue(route[i], route[i+1],
			g.Channels, amt, ADD)
		if err != nil {
			return err
		}
		// i+1 到 i 的钱增加
		err = UpdateLinkValue(route[i+1], route[i],
			g.Channels, amt, SUB)
		if err != nil {
			return err
		}
	}
	return nil
}

// 更新一条路，但是只是单方面只增加或只减少，面向的是预支付场景
func (g *Graph)UpdateWeighOneDir(route Path, amt Amount, addOrSub bool) error {
	for i := 0; i < len(route)-1; i++ {
		// i 到 i+1 的钱减少
		err := UpdateLinkValue(route[i], route[i+1],
			g.Channels, amt, addOrSub)
		if err != nil {
			return err
		}
	}
	return nil
}

// 反方向回滚支付
func (g *Graph) UpdateWeightsReverse(routes []Path,
	amts []Amount) error {

	if len(routes) != len(amts) {
		return fmt.Errorf("routes number is not equal to amts' ")
	}

	for idx, route := range routes {
		for i := 0; i < len(route)-1; i++ {
			// 从i到i+1的钱增加
			err := UpdateLinkValue(route[i], route[i+1],
				g.Channels, amts[idx], ADD)
			if err != nil {
				return err
			}
			// i+1 到 i 的钱减少
			err = UpdateLinkValue(route[i+1], route[i],
				g.Channels, amts[idx], SUB)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetGraph(data string) *Graph {
	f, err := os.Open(data + "/finalSets/static/ripple-lcc.graph_CREDIT_LINKS")
	if err != nil {
		fmt.Println("os Open error: ", err)
		return nil
	}
	defer f.Close()

	br := bufio.NewReader(f)
	lineNum := 1
	links := make(map[string]*Link, 0)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("br ReadLine error: ", err)
			return nil
		}

		// 过滤掉前面几行的无用信息
		if lineNum < 5 {
			lineNum++
			continue
		}

		splitted := strings.Split(string(line), " ")
		id1, _ := strconv.Atoi(splitted[0])
		id2, _ := strconv.Atoi(splitted[1])
		v1, _ := strconv.ParseFloat(splitted[2], 64)
		v2, _ := strconv.ParseFloat(splitted[3], 64)
		v3, _ := strconv.ParseFloat(splitted[4], 64)
		link := &Link{
			Part1: RouterID(id1),
			Part2: RouterID(id2),
			Val1:  Amount(v3 - v2),
			Val2:  Amount(v2 - v1),
		}
		links[GetLinkKey(link.Part1, link.Part2)] = link
	}

	nodes := make([]*Node, 67149)
	for i := 0; i < 67149; i++ {
		router := &Node{
			ID:         RouterID(i),
			Parents:    make([]RouterID, 0),
			Children:   make([]RouterID, 0),
			Neighbours: make([]RouterID, 0),
		}
		nodes[RouterID(i)] = router
	}

	keySlice := make([]string, 0)
	for k := range links {
		keySlice = append(keySlice, k)
	}
	sort.Strings(keySlice)
	for _, key := range keySlice {
		edge := links[key]
		nodes[edge.Part1].Neighbours = append(nodes[edge.Part1].Neighbours, edge.Part2)
		nodes[edge.Part2].Neighbours = append(nodes[edge.Part2].Neighbours, edge.Part1)
	}

	graph := &Graph{
		Nodes:    nodes,
		Channels: links,
		DAGs:     make(map[RouterID]*DAG),
		SPTs:     make(map[RouterID]*DAG),
		Distance: make(map[RouterID]map[RouterID]float64),
	}
	RanddomFeeRate(links)
	return graph
}

func (g *Graph)UpdateLinkValue(from, to RouterID, amt Amount,
	addOrSub bool) error {
	err := UpdateLinkValue(from,to, g.Channels, amt, addOrSub)
	return err
}

func (g *Graph)GetFee(path Path, amt Amount) Amount {
	fee := Amount(0)
	for i:=0;i<len(path)-1; i++ {
		fee += GetLinkFeeRate(path[i],path[i+1], g.Channels) * amt
	}
	return fee
}

func (g *Graph) StoreDistances(fileName string, threadNum int) error {
	wg := sync.WaitGroup{}
	tryNodes := g.Nodes[1000:10000]

	splitNum := len(tryNodes) / threadNum
	lock := sync.Mutex{}
	num := 0
	for i := 0; i < threadNum;  {
		wg.Add(1)
		go func(n int) {
			calSet := tryNodes[n*splitNum:(n+1)*splitNum]
			if n == threadNum -1 {
				calSet = append(calSet, tryNodes[(n+1)*splitNum:]...)
			}
			for _, node := range calSet {
				_, distances := Dijkstra(g.Nodes, node.ID)
				num ++
				fmt.Printf("one done %v\n", num)
				lock.Lock()
				g.Distance[node.ID] = distances
				lock.Unlock()
			}
			wg.Done()
		}(i)
		i++
	}
	wg.Wait()

	fileObj,err := os.OpenFile(fileName,os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	if err!= nil {
		return err
	}
	defer fileObj.Close()

	writeObj := bufio.NewWriter(fileObj)
	for _, node := range tryNodes {
		//使用Write方法,需要使用Writer对象的Flush方法将buffer中的数据刷到磁盘
		buf := make([]byte, len(g.Nodes))
		for i:=0; i<len(g.Nodes); i++ {
			buf[i] = byte(g.Distance[node.ID][RouterID(i)])
		}
		if _,err := writeObj.Write(buf);err == nil {
			if  err := writeObj.Flush(); err != nil {panic(err)}
		}
	}
	return nil
}

func (g *Graph)LoadDistances(fileName string, dests map[RouterID]struct{}) error {

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	s := make([]byte, len(g.Nodes))
	id := RouterID(0)
	for {
		switch nr, err := f.Read(s[:]); true {
		case nr < 0:
			panic(err.Error())
		case nr == 0: // EOF
			delete(g.Distance,id)
			return nil
		case nr > 0:
			if _, ok := dests[id]; !ok {
				id++
				continue
			}
			fmt.Printf("load one dest%v\n", id)
			g.Distance[id] = make(map[RouterID]float64)
			for i, n := range s {
				g.Distance[id][RouterID(i)] = float64(n)
			}
		}
		id++
	}
	return nil
}
