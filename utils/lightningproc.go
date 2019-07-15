package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

/// Returns a new instance of the directed channel graph.
type ChannelGraph struct {
	/// The list of `LightningNode`s in this channel graph
	Nodes []*LightningNode `protobuf:"bytes,1,rep,name=nodes,proto3" json:"nodes,omitempty"`
	/// The list of `ChannelEdge`s in this channel graph
	Edges                []*ChannelEdge `protobuf:"bytes,2,rep,name=edges,proto3" json:"edges,omitempty"`
}

type LightningNode struct {
	LastUpdate           uint32         `protobuf:"varint,1,opt,name=last_update,proto3" json:"last_update,omitempty"`
	PubKey               string         `protobuf:"bytes,2,opt,name=pub_key,proto3" json:"pub_key,omitempty"`
	Alias                string         `protobuf:"bytes,3,opt,name=alias,proto3" json:"alias,omitempty"`
	Addresses            []*NodeAddress `protobuf:"bytes,4,rep,name=addresses,proto3" json:"addresses,omitempty"`
	Color                string         `protobuf:"bytes,5,opt,name=color,proto3" json:"color,omitempty"`
}

type NodeAddress struct {
	Network              string   `protobuf:"bytes,1,opt,name=network,proto3" json:"network,omitempty"`
	Addr                 string   `protobuf:"bytes,2,opt,name=addr,proto3" json:"addr,omitempty"`
}

type ChannelEdge struct {
	ChannelId            string         `protobuf:"varint,1,opt,name=channel_id,proto3" json:"channel_id,omitempty"`
	ChanPoint            string         `protobuf:"bytes,2,opt,name=chan_point,proto3" json:"chan_point,omitempty"`
	LastUpdate           uint32         `protobuf:"varint,3,opt,name=last_update,proto3" json:"last_update,omitempty"`
	Node1Pub             string         `protobuf:"bytes,4,opt,name=node1_pub,proto3" json:"node1_pub,omitempty"`
	Node2Pub             string         `protobuf:"bytes,5,opt,name=node2_pub,proto3" json:"node2_pub,omitempty"`
	Capacity             string          `protobuf:"varint,6,opt,name=capacity,proto3" json:"capacity,omitempty"`
	Node1Policy          *RoutingPolicy `protobuf:"bytes,7,opt,name=node1_policy,proto3" json:"node1_policy,omitempty"`
	Node2Policy          *RoutingPolicy `protobuf:"bytes,8,opt,name=node2_policy,proto3" json:"node2_policy,omitempty"`
}

type RoutingPolicy struct {
	TimeLockDelta        uint32   `protobuf:"varint,1,opt,name=time_lock_delta,proto3" json:"time_lock_delta,omitempty"`
	MinHtlc              string    `protobuf:"varint,2,opt,name=min_htlc,proto3" json:"min_htlc,omitempty"`
	FeeBaseMsat          string    `protobuf:"varint,3,opt,name=fee_base_msat,proto3" json:"fee_base_msat,omitempty"`
	FeeRateMilliMsat     string    `protobuf:"varint,4,opt,name=fee_rate_milli_msat,proto3" json:"fee_rate_milli_msat,omitempty"`
	Disabled             bool     `protobuf:"varint,5,opt,name=disabled,proto3" json:"disabled,omitempty"`
	MaxHtlcMsat          string   `protobuf:"varint,6,opt,name=max_htlc_msat,proto3" json:"max_htlc_msat,omitempty"`
}

func ParseLightningGraph(filePath string) (*Graph, error) {
	var g ChannelGraph
	graphJson, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(graphJson, &g); err != nil {
		return nil, err
	}

	nodeIDMap := make(map[string]RouterID)
	nodes := make(map[RouterID]*Node)
	channels := make(map[string]*Link)
	index := RouterID(0)
	for _, lightNode := range g.Nodes {
		nodeIDMap[lightNode.PubKey] = index
		node := &Node{
			ID: index,
			Neighbours: make([]RouterID,0),
		}
		nodes[index]=node
		index++
	}

	for _, lightEdge := range g.Edges {
		mapped1 := nodeIDMap[lightEdge.Node1Pub]
		mapped2 := nodeIDMap[lightEdge.Node2Pub]
		linkKey := GetLinkKey(mapped1, mapped2)
		capacity, _ :=  strconv.ParseInt(lightEdge.Capacity, 10, 64)
		if mapped1 < mapped2 {
			link := &Link{
				Part1: mapped1,
				Part2: mapped2,
				Val1: Amount(capacity)/2,
				Val2: Amount(capacity)/2,
			}
			channels[linkKey] = link
		} else {
			link := &Link{
				Part1: mapped2,
				Part2: mapped1,
				Val1: Amount(capacity)/2,
				Val2: Amount(capacity)/2,
			}
			channels[linkKey] = link
		}

		if nodes[mapped1].CheckLink(mapped2) ||
			nodes[mapped2].CheckLink(mapped1) {
			continue
		}
		nodes[mapped1].Neighbours = append(nodes[mapped1].Neighbours, mapped2)
		nodes[mapped2].Neighbours = append(nodes[mapped2].Neighbours, mapped1)
	}
	return &Graph{
		Nodes: nodes,
		Channels:channels,
		DAGs: map[RouterID]*DAG{},
		SPTs: map[RouterID]*DAG{},
		Distance: map[RouterID]map[RouterID]float64{},
	}, nil
}

func GetLightningTrans(nodeNum, transNum int, tranPath, valuePath string) ([]Tran, error){
	// 先根据ripple的数据生成sender和reeiver对
	trans := make([]Tran, 0)
	if fileObj, err := os.Open(tranPath); err == nil {
		defer fileObj.Close()
		reader := bufio.NewReader(fileObj)
		for {
			if line, _, err := reader.ReadLine(); err == nil {
				spplieted := strings.Split(string(line), ",")
				src, _ := strconv.ParseInt(spplieted[0], 10, 32)
				dest, _ := strconv.ParseInt(spplieted[1], 10, 32)
				tran := Tran{
					Src: int(src)%nodeNum,
					Dest: int(dest)%nodeNum,
				}
				if tran.Src == tran.Dest {
					continue
				}
				trans = append(trans, tran)
			} else {
				if err != io.EOF {
					panic("load ripple sender receiver pair failed")
				} else {
					break
				}
			}
		}
	} else {
		return nil, err
	}

	// 再根据bitcoin的真实值生成交易的数值
	values := make([]float64,0)
	f, err := os.Open(valuePath)
	if err != nil {
		fmt.Println("os Open error: ", err)
	}
	defer f.Close()
	br := bufio.NewReader(f)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("br ReadLine error: ", err)
			return nil, err
		}
		amt, _ := strconv.ParseFloat(string(line), 64)
		values = append(values, amt)
	}

	fmt.Printf("number of trans is:%v", len(trans))

	selectTrans := make([]Tran, transNum)
	rand.Seed(time.Now().Unix())
	for i := range selectTrans {
		selectTrans[i] = trans[rand.Intn(len(trans))]
		selectTrans[i].Val = values[rand.Intn(len(values))]
	}
	return selectTrans, nil
}

