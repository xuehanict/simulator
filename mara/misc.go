package mara

import (
	"bufio"
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

func GetRippleMaraAndTrans(data string) (*Mara, []Tran) {
	f, err := os.Open(data + "/finalSets/static/ripple-lcc.graph_CREDIT_LINKS")
	if err != nil {
		fmt.Println("os Open error: ", err)
		return nil, nil
	}
	defer f.Close()

	br := bufio.NewReader(f)
	lineNum := 1
	links := make(map[string]*utils.Link, 0)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("br ReadLine error: ", err)
			return nil, nil
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
		link := &utils.Link{
			Part1: utils.RouterID(id1),
			Part2: utils.RouterID(id2),
			Val1:  utils.Amount(v3 - v2),
			Val2:  utils.Amount(v2 - v1),
		}
		links[utils.GetLinkKey(link.Part1, link.Part2)] = link
	}

	nodes := make([]*Node, 67149)
	for i := 0; i < 67149; i++ {
		router := &Node{
			ID:         utils.RouterID(i),
			Parents:    make([]utils.RouterID, 0),
			Children:   make([]utils.RouterID, 0),
			Neighbours: make([]utils.RouterID, 0),
		}
		nodes[utils.RouterID(i)] = router
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

	m := &Mara{
		Graph: &Graph{
			Nodes:    nodes,
			Channels: links,
			DAGs:     make(map[utils.RouterID]*DAG),
			SPTs:     make(map[utils.RouterID]*DAG),
			Distance: make(map[utils.RouterID]map[utils.RouterID]float64),
		},
	}

	fmt.Printf("节点link数据解析完成\n")
	trans := generateTrans(data + "/finalSets/static/sampleTr-2.txt")
	fmt.Printf("交易数据解析完成\n")

	return m, trans
}

type Tran struct {
	Src  int
	Dest int
	Val  float64
}

func generateTrans(filePath string) []Tran {

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("os Open error: ", err)
		return nil
	}
	defer f.Close()

	br := bufio.NewReader(f)
	trans := make([]Tran, 0)
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
		dest, _ := strconv.Atoi(splitStr[2])

		trans = append(trans, Tran{
			Src:  src,
			Dest: dest,
			Val:  val,
		})
	}

	return trans
}
