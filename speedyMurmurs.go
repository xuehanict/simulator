package main

import (
	"fmt"
)

type RouteID int

type SMRouter struct {
	ID            RouteID
	AddrWithRoots map[RouteID]string
	Roots         []RouteID
	Neighbours    map[RouteID]struct{}
	RouteBase     map[RouteID]*SMRouter
	LinkBase      map[string]*Link
	MsgPool       chan interface{}
	quit          chan struct{}
}

type Message interface {
}

/*
 * val1 指part1往part2方向的通道容量
 * val2 指part2往part1方向的通道容量
 * part1 的id 小于part2的id
 */
type Link struct {
	part1 RouteID
	part2 RouteID
	val1  float64
	val2  float64
}

/*************消息的多个类型*************/
/**
交易请求信息
*/
type payReq struct {
	sender RouteID
	dest   RouteID
}

/**
地址和root映射的map，在构建时交换
*/
type addrMap struct {
	// 发送人的ID
	router RouteID
	// 关于每个树的root的id和其对应树中的地址
	addrs map[RouteID]string
}

/*************************************/

func (r *SMRouter) start() {
	for {
		select {
		case msg := <-r.MsgPool:
			r.onMsg(msg)
		case <-r.quit:
			fmt.Printf("Router %v closed\n", r.ID)
			return
		}
	}
}

func (r *SMRouter) stop() {
	close(r.quit)
}

func (r *SMRouter) sendMsg(id RouteID, msg interface{}) {
	r.RouteBase[id].MsgPool <- msg
}

func (r *SMRouter) onMsg(msg interface{}) {
	switch msg.(type) {
	case *payReq:
		fmt.Printf("\n")
	case *addrMap:
		fmt.Printf("\n")
	}
}

func (r *SMRouter) onLinkAdd(add *Link) {
	// 如果part1是自己，那么part2就是对方
	var neighbour RouteID
	if add.part1 == r.ID {
		r.Neighbours[add.part2] = struct{}{}
		neighbour = add.part2
	} else {
		r.Neighbours[add.part1] = struct{}{}
		neighbour = add.part2
	}

	// 发送这个当前节点的各个树的地址到邻居
	am := &addrMap{
		router: r.ID,
		addrs:  r.AddrWithRoots,
	}
	r.sendMsg(neighbour, am)
}

func (r *SMRouter) onAddrMap(am *addrMap) {
	for root, addr := range am.addrs {

	}
}

func (r *SMRouter) onPayReq(req *payReq) {

}
