package main

import (
	"fmt"
	"math"
	"time"
)

const (
	ADD = 1
	SUB = 0
	PAYMENT_REQUESTID_LENGTH = 10
)



type RouteID int
type RequestID string

type SMRouter struct {
	ID            RouteID
	AddrWithRoots map[RouteID]string
	Roots         []RouteID
	Neighbours    map[RouteID]struct{}
	RouterBase     map[RouteID]*SMRouter
	reqestPool 	  map[RequestID]chan *payRes
	probeBase 	  map[RequestID]map[RouteID]*probeInfo
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
	 upperHop RouteID
	 root RouteID
	 destAddr string
}

/*************消息的多个类型*************/
/**
交易请求信息
*/
type payReq struct {
	requestID RequestID
	root 	RouteID
	sender RouteID
	dest   string
	value  float64
	upperHop  RouteID
}

type payRes struct {
	requestID RequestID
	root RouteID
	sender RouteID
	success bool
	val float64
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
	root RouteID
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
	r.RouterBase[id].MsgPool <- msg
}

func (r *SMRouter) onMsg(msg interface{}) {
	switch msg.(type) {
	case *payReq:
		r.onPayReq(msg.(*payReq))
	case *payRes:
		r.onPayRes(msg.(*payRes))
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
	val := req.value
	root := req.root
	dest := req.dest

	// 如果自己就是dest节点
	if dest == r.AddrWithRoots[root] {
		res := &payRes{
			 success: true,
			 sender: req.sender,
			 requestID: req.requestID,
			 val: req.value,
			 root: req.root,
		}
		newProbe := &probeInfo{
			value: req.value,
			requestID: req.requestID,
			root: req.root,
			time: time.Now().Unix(),
			upperHop: req.upperHop,
			destAddr: req.dest,
		}
		if _, ok := r.probeBase[req.requestID]; ok {
			r.probeBase[req.requestID][req.root] = newProbe
		} else {
			r.probeBase[req.requestID] = make(map[RouteID]*probeInfo)
			r.probeBase[req.requestID][req.root] = newProbe
		}
		r.sendMsg(req.upperHop, res)

	} else {
		nextHop, err := r.getNeighbourToSend(root, dest, val)
		if err != nil {
			// TODO(xuehan): add log
			fmt.Printf("raise error:%v",err)
		}
		r.updateLinkValue(r.ID, nextHop,req.value, SUB)
		req.upperHop = r.ID
		newProbe := &probeInfo{
			value: req.value,
			requestID: req.requestID,
			root: req.root,
			time: time.Now().Unix(),
			upperHop: req.upperHop,
			nextHop: nextHop,
			destAddr: req.dest,
		}
		if _, ok := r.probeBase[req.requestID]; ok {
			r.probeBase[req.requestID][req.root] = newProbe
		} else {
			r.probeBase[req.requestID] = make(map[RouteID]*probeInfo)
			r.probeBase[req.requestID][req.root] = newProbe
		}
		r.sendMsg(nextHop, req)
	}
}

func (r *SMRouter) onPayRes(res *payRes)  {
	probe := r.probeBase[res.requestID][res.root]
	if res.sender == r.ID{
		r.reqestPool[res.requestID] <- res
	} else {
		if !res.success {
			nextHop := probe.nextHop
			r.updateLinkValue(r.ID, nextHop,probe.value, ADD)
		}
		r.sendMsg(probe.upperHop, res)
	}
}

func (r *SMRouter) onPayment (pay *payment) {
	probe := r.probeBase[pay.requestID][pay.root]
	if probe.destAddr != r.AddrWithRoots[pay.root] {
		r.updateLinkValue(probe.nextHop, r.ID, probe.value, ADD)
	}
	delete(r.probeBase[pay.requestID], pay.root)
}

func (r *SMRouter) sendPayment (dest RouteID, amount float64) error{

	splittedAmounts := randomPartition(amount, len(r.Roots))
	neighboursToSend := make([]RouteID, len(r.Roots))
	destAddr := r.RouterBase[dest].AddrWithRoots[dest]
	requestID := RequestID(GetRandomString(PAYMENT_REQUESTID_LENGTH))
	for i, root := range r.Roots {
		nextHop, err  := r.getNeighbourToSend(root, destAddr, splittedAmounts[i])
		if err != nil {
			return fmt.Errorf("send payment failed: %v", err)
		}
		neighboursToSend[i] = nextHop
		// ps: 这里的地址是直接从dest节点中拿出的，实际场景应该从
		payreq := &payReq{
			sender: r.ID,
			requestID: requestID,
			value: splittedAmounts[i],
			root: root,
			dest: r.RouterBase[dest].AddrWithRoots[root],
		}

		newProbe := &probeInfo{
			value: payreq.value,
			requestID: requestID,
			root: root,
			time: time.Now().Unix(),
			nextHop: nextHop,
		}
		if _, ok := r.probeBase[payreq.requestID]; ok {
			r.probeBase[requestID][root] = newProbe
		} else {
			r.probeBase[requestID] = make(map[RouteID]*probeInfo)
			r.probeBase[requestID][root] = newProbe
		}
		r.sendMsg(nextHop,payreq)
	}

	resArray := make([]*payRes,0)
out:
	for {
		select {
		case res := <- r.reqestPool[requestID]:
			if res.success == false {
				return fmt.Errorf("probe failed")
			}
			resArray = append(resArray, res)
			if len(resArray) == len(r.Roots) { break out}
		case <- time.After(2 * time.Second):
			return fmt.Errorf("probe failed, timeout")
		}
	}

	/**
	probe 成功，现在开始进行支付，因为在probe的过程中已经减去了支付的那一笔钱，
	因此在真正支付的过程中只需要将接受方的那一部分金额加上就可以了。
	*/
	for _, probe := range r.probeBase[requestID] {
		r.updateLinkValue(probe.nextHop, r.ID, probe.value, ADD)
	}
	return nil
}

/**
基于以root为根的生成树，获取到dest的邻居下一跳.
目前的模拟是直接获取邻居的地址，实际场景下应该需要从邻居临时fetch过来
 */
func (r *SMRouter) getNeighbourToSend (root RouteID, dest string,
	amount float64) (RouteID, error) {

	minDis := math.MaxInt32
	minNeighbour := RouteID(-1)
	for n := range r.Neighbours {
		tmpAddr := r.RouterBase[n].AddrWithRoots[root]
		tmpDist := getDis(
			tmpAddr,
			dest, 4)
		linkValue := 0.0
		if r.ID < n {
			link, ok := r.LinkBase[getLinkKey(r.ID, n)]
			if ok {
				linkValue = link.val1
			}
		} else {
			link, ok := r.LinkBase[getLinkKey(n, r.ID)]
			if ok {
				linkValue = link.val2
			}
		}
		if tmpDist < minDis && amount < linkValue{
			minDis = tmpDist
			minNeighbour = n
		}
	}
	if minNeighbour == -1 {
		return -1, fmt.Errorf("cann't find a suitable next hop")
	}
	return minNeighbour, nil
}

func (r *SMRouter) updateLinkValue (from, to RouteID, value float64,
	flag int)  {
	if from > to {
		linkKey := getLinkKey(to, from)
		link := r.LinkBase[linkKey]
		if flag == ADD {
			link.val2 += value
		} else {
			link.val2 -= value
		}
	} else {
		linkKey := getLinkKey(from, to)
		link := r.LinkBase[linkKey]
		if flag == ADD {
			link.val1 += value
		} else {
			link.val1 -= value
		}
	}
}

