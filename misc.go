package main

import (
	"bufio"
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	tenNodesGraph     = "data/ten_nodes.json"
	tenNodesGraphHalf = "data/ten_nodes_half.json"
)

/**
用来解析图的json文件的辅助结构和函数
*/
type testGraph struct {
	Info  []string   `json:"info"`
	Nodes []testNode `json:"nodes"`
	Edges []testEdge `json:"edges"`
}

type testNode struct {
	Id utils.RouterID `json:"id"`
}

type testEdge struct {
	Node1     utils.RouterID `json:"node_1"`
	Node2     utils.RouterID `json:"node_2"`
	Capacity1 float64        `json:"capacity1"`
	Capacity2 float64        `json:"capacity2"`
}

type tran struct {
	src  int
	dest int
	val  float64
}

func GenerateTrans(filePath string) []utils.Tran {

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("os Open error: ", err)
		return nil
	}
	defer f.Close()

	br := bufio.NewReader(f)
	trans := make([]utils.Tran, 0)
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

		trans = append(trans, utils.Tran{
			Src:  src,
			Dest: dest,
			Val:  val,
		})
	}
	return trans
}

