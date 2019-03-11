package main

import (
	"bytes"
	"math/rand"
	"strconv"
	"time"
	"sort"
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
	Id RouteID `json:"id"`
}

type testEdge struct {
	Node1     RouteID `json:"node_1"`
	Node2     RouteID `json:"node_2"`
	Capacity1 int64   `json:"capacity1"`
	Capacity2 int64   `json:"capacity2"`
}

func selectRouters() []RouteID {

	return nil
}

// 生成link的key，要求r1 < r2
func getLinkKey(r1, r2 RouteID) string {
	return strconv.Itoa(int(r1)) + "-" + strconv.Itoa(int(r2))
}

// 通过父节点生成孩子节点的地址
func DeriveAddrr(parentAddr string) string {
	return parentAddr + GetRandomString(4)
}

// 生成一定长度的随机字符串
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// speedyMurmurs中用到的计算距离的方式
func getDisSM(neighbour, dest string, lengthInterval int) float64 {
	cpl := 0
	depth := len(neighbour) / lengthInterval
	destLen := len(dest) / lengthInterval
	neiBytes := []byte(neighbour)
	destBytes := []byte(dest)
	for ; cpl < depth && cpl < destLen &&
		bytes.Equal(neiBytes[0:lengthInterval],
			destBytes[0:lengthInterval]); cpl++ {
		neiBytes = neiBytes[lengthInterval:]
		destBytes = neiBytes[lengthInterval:]
	}

	return -float64(cpl) - 1/float64(depth+1+destLen)
}

// silentWhisper中用到的计算距离的方式
func getDisSW(neighbour, dest string, lenthInterval int) int {
	depthN := len(neighbour) / lenthInterval
	depthD := len(dest) / lenthInterval
	cpl := 0
	neiBytes := []byte(neighbour)
	destBytes := []byte(dest)

	for ; cpl < depthD && cpl < depthN &&
		bytes.Equal(neiBytes[0:lenthInterval],
			destBytes[0:lenthInterval]); cpl++ {
			neiBytes = neiBytes[lenthInterval:]
			destBytes = destBytes[lenthInterval:]
	}
	return depthN + depthD - (2 * cpl)
}

func randomPartition (amount float64, num int) []float64 {
	res := make([]float64,num)
	rate := make([]float64, num)

	for i:= 0; i < num - 1; i++ {
		rate[i] = rand.Float64()
	}
	rate[num-1] = 1

	sort.Float64s(rate)
	res[0] = amount * rate[0]
	for i := 1; i < num; i++ {
		res[i] = amount * (rate[i] - rate[i-1])
	}
	return res
}

