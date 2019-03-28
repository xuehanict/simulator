package main

import (
	sw "github.com/lightningnetwork/sm/silentWhisper"
	"os"
	"fmt"
	"bufio"
	"io"
	"strings"
	"strconv"
)

const (
	tenNodesGraph = "data/ten_nodes.json"
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
	Id sw.RouteID `json:"id"`
}

type testEdge struct {
	Node1     sw.RouteID `json:"node_1"`
	Node2     sw.RouteID `json:"node_2"`
	Capacity1 float64   `json:"capacity1"`
	Capacity2 float64   `json:"capacity2"`
}

type tran struct {
	src int
	dest int
	val float64
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

