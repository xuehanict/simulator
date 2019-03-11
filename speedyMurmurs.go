package main

import (
	"fmt"
)

type RouteID int
type RequestID string

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

type probeInfo struct {
	 requestID RequestID
	 value float64
	 time int64
	 nextHop RouteID
}

/*************消息的多个类型*************/
/**
交易请求信息
*/
type payReq struct {
	requestID RequestID
	sender RouteID
	dest   string
	value  int
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

/**
地址更新时交换的信息
 */
type addrUpdate struct {

}

/**
进行支付时,传递此消息
 */
type payment struct {
	requestID RequestID
	value int
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
		r.onAddrMap(msg.(*addrMap))
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

// 创建link时，会和邻居交换addr的map，然后根据邻居的地址修改自己的map
func (r *SMRouter) onAddrMap(am *addrMap) {
	for root, addr := range am.addrs {
		if _,ok := r.AddrWithRoots[root]; !ok  {
			r.AddrWithRoots[root] = DeriveAddrr(addr)
		}
	}
}

func (r *SMRouter) onPayReq(req *payReq) {

}

func (r *SMRouter) sendPayment (dest RouteID, amount float64) error{

	return nil
}

