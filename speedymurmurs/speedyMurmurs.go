package speedymurmurs

import (
	"fmt"
	"math"
	"time"
	"bytes"
	"math/rand"
	"log"
	"github.com/davecgh/go-spew/spew"
)

const (
	ADD = 1
	SUB = 0

	PAYMENT_REQUESTID_LENGTH = 10
	ADDR_REQUESTID_LENGTH    = 10
	ADDR_LENGTH_INTERVAL     = 8

	CLEAN_PROBE_INTERVAL = 20
	PROBE_STORE_TIME     = 3600

	LINK_DIR_RIGHT = true
	LINK_DIR_LEFT  = false

	MSG_POOL_CAPCITY     = 10000
	ADDR_POOL_CAPCITY    = 10000
	PAYMENT_POOL_CAPCITY = 100
)

var SMLogger *log.Logger

type RouteID int
type RequestID string



type SMRouter struct {
	ID              RouteID
	AddrWithRoots   map[RouteID]*AddrType
	Roots           []RouteID
	Neighbours      map[RouteID]struct{}
	RouterBase      map[RouteID]*SMRouter
	payRequestPool  map[RequestID]chan *payRes
	addrRequestPool map[RequestID]chan *addrRes
	probeBase       map[RequestID]map[RouteID]*probeInfo
	LinkBase        map[string]*Link
	MsgPool         chan interface{}
	timer           *time.Ticker
	quit            chan struct{}
}

/*
 * Val1 指Part1往Part2方向的通道容量
 * Val2 指Part2往Part1方向的通道容量
 * Part1 的id 小于Part2的id
 */
type Link struct {
	Part1 RouteID
	Part2 RouteID
	Val1  float64
	Val2  float64
}

type probeInfo struct {
	requestID RequestID
	value     float64
	time      int64
	nextHop   RouteID
	upperHop  RouteID
	root      RouteID
	destAddr  string
}

type AddrType struct {
	Addr   string
	Parent RouteID
}


func (r *SMRouter) Start() {
	for {
		select {
		case msg := <-r.MsgPool:
			r.onMsg(msg)
		case <-r.timer.C:
			r.cleanProbe()
		case <-r.quit:
			fmt.Printf("Router %v closed\n", r.ID)
			return
		}
	}
}

func (r *SMRouter) Stop() {
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
	case *addrRes:
		r.onResetAddrRes(msg.(*addrRes))
	case *addrReq:
		r.onResetAddrReq(msg.(*addrReq))
	case *addrResetNoti:
		r.onNotifyReset(msg.(*addrResetNoti))
	case *Payment:
		r.onPayment(msg.(*Payment))
	}
}

/*
func (r *SMRouter) onLinkAdd(add *Link) {
	// 如果Part1是自己，那么Part2就是对方
	var neighbour RouteID
	if add.Part1 == r.ID {
		r.Neighbours[add.Part2] = struct{}{}
		neighbour = add.Part2
	} else {
		r.Neighbours[add.Part1] = struct{}{}
		neighbour = add.Part2
	}

	// 发送这个当前节点的各个树的地址到邻居
	am := &addrMap{
		router: r.ID,
		addrs:  r.AddrWithRoots,
	}
	r.sendMsg(neighbour, am)
}
*/

/*
// 创建link时，会和邻居交换addr的map，然后根据邻居的地址修改自己的map
func (r *SMRouter) onAddrMap(am *addrMap) {
	for root, addr := range am.addrs {
		if _, ok := r.AddrWithRoots[root]; !ok {
			r.AddrWithRoots[root] = DeriveAddrr(addr)
		}
	}
}
*/

func (r *SMRouter) onPayReq(req *payReq) {

	SMLogger.Printf("R %v received payreq :%v ",r.ID, spew.Sdump(req))

	val := req.value
	root := req.root
	dest := req.dest

	// 如果自己就是dest节点
	if dest == r.AddrWithRoots[root].Addr {
		res := &payRes{
			success:   true,
			sender:    req.sender,
			requestID: req.requestID,
			val:       req.value,
			root:      req.root,
		}
		newProbe := &probeInfo{
			value:     req.value,
			requestID: req.requestID,
			root:      req.root,
			time:      time.Now().Unix(),
			upperHop:  req.upperHop,
			destAddr:  req.dest,
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
			fmt.Printf("raise error:%v", err)
		}
		r.updateLinkValue(r.ID, nextHop, req.value, SUB)
		newProbe := &probeInfo{
			value:     req.value,
			requestID: req.requestID,
			root:      req.root,
			time:      time.Now().Unix(),
			upperHop:  req.upperHop,
			nextHop:   nextHop,
			destAddr:  req.dest,
		}
		req.upperHop = r.ID
		if _, ok := r.probeBase[req.requestID]; ok {
			r.probeBase[req.requestID][req.root] = newProbe
		} else {
			r.probeBase[req.requestID] = make(map[RouteID]*probeInfo)
			r.probeBase[req.requestID][req.root] = newProbe
		}
		r.sendMsg(nextHop, req)
	}
}

func (r *SMRouter) onPayRes(res *payRes) {
	SMLogger.Printf("R %v recieved payres %v ",r.ID, spew.Sdump(res))
	probe := r.probeBase[res.requestID][res.root]
	if res.sender == r.ID {
		r.payRequestPool[res.requestID] <- res
	} else {
		if !res.success {
			nextHop := probe.nextHop
			r.updateLinkValue(r.ID, nextHop, probe.value, ADD)
		}
		r.sendMsg(probe.upperHop, res)
	}
}

func (r *SMRouter) onPayment(pay *Payment) {
	probe := r.probeBase[pay.requestID][pay.root]
	SMLogger.Printf("R %v recieve payment %v", r.ID, pay)

	if probe.destAddr != r.AddrWithRoots[pay.root].Addr {
		//remoteOldVal, _:= r.getLinkValue(probe.nextHop, LINK_DIR_LEFT)
		r.updateLinkValue(probe.nextHop, r.ID, probe.value, ADD)
		//remoteNewVal, _:= r.getLinkValue(probe.nextHop, LINK_DIR_LEFT)
		// 让远程节点检查是否需要重构
		//r.RouterBase[probe.nextHop].monitorLinkChange(remoteOldVal, remoteNewVal, r.ID)

		//selfNewValue, _ := r.getLinkValue(probe.nextHop, LINK_DIR_RIGHT)
		//selfOldValue := selfNewValue - probe.value
		// 本地节点检查是否需要重构
		//r.monitorLinkChange(selfOldValue, selfNewValue, probe.nextHop)

	}
	if r.AddrWithRoots[probe.root].Addr != probe.destAddr {
		r.sendMsg(probe.nextHop,pay)
	}
	delete(r.probeBase[pay.requestID], pay.root)
	if len(r.probeBase[pay.requestID]) == 0 {
		delete(r.probeBase, pay.requestID)
	}
}

func (r *SMRouter) SendPayment(dest RouteID, amount float64) error {

	splittedAmounts := randomPartition(amount, len(r.Roots))
	neighboursToSend := make([]RouteID, len(r.Roots))
	requestID := RequestID(GetRandomString(PAYMENT_REQUESTID_LENGTH))
	r.payRequestPool[requestID] = make(chan *payRes, PAYMENT_POOL_CAPCITY)
	for i, root := range r.Roots {
		destAddr := r.RouterBase[dest].AddrWithRoots[root]
		nextHop, err := r.getNeighbourToSend(root, destAddr.Addr, splittedAmounts[i])
		if err != nil {
			return fmt.Errorf("send payment failed: %v", err)
		}
		neighboursToSend[i] = nextHop
		// ps: 这里的地址是直接从dest节点中拿出的，实际场景应该从
		payreq := &payReq{
			sender:    r.ID,
			requestID: requestID,
			value:     splittedAmounts[i],
			root:      root,
			dest:      r.RouterBase[dest].AddrWithRoots[root].Addr,
			upperHop:  r.ID,
		}

		newProbe := &probeInfo{
			value:     payreq.value,
			requestID: requestID,
			root:      root,
			time:      time.Now().Unix(),
			nextHop:   nextHop,
		}
		if _, ok := r.probeBase[payreq.requestID]; ok {
			r.probeBase[requestID][root] = newProbe
		} else {
			r.probeBase[requestID] = make(map[RouteID]*probeInfo)
			r.probeBase[requestID][root] = newProbe
		}
		r.updateLinkValue(r.ID, nextHop, splittedAmounts[i], SUB)
		r.sendMsg(nextHop, payreq)
	}

	resArray := make([]*payRes, 0)
out:
	for {
		select {
		case res := <-r.payRequestPool[requestID]:
			//SMLogger.Printf("payres :%v", spew.Sdump(res))
			if res.success == false {
				return fmt.Errorf("probe failed")
			}
			resArray = append(resArray, res)
			if len(resArray) == len(r.Roots) {
				break out
			}
		case <-time.After(2 * time.Second):
			return fmt.Errorf("probe failed, timeout")
		}
	}

	/**
	probe 成功，现在开始进行支付，因为在probe的过程中已经减去了支付的那一笔钱，
	因此在真正支付的过程中只需要将接受方的那一部分金额加上就可以了。
	*/
	for _, probe := range r.probeBase[requestID] {

		//TODO(xuehan): add monitor link change

		r.updateLinkValue(probe.nextHop, r.ID, probe.value, ADD)



		payment := &Payment{
			requestID: requestID,
			root: probe.root,
		}
		r.sendMsg(probe.nextHop,payment)
	}
	return nil
}

/**
基于以root为根的生成树，获取到dest的邻居下一跳.
目前的模拟是直接获取邻居的地址，实际场景下应该需要从邻居临时fetch过来
*/
func (r *SMRouter) getNeighbourToSend(root RouteID, dest string,
	amount float64) (RouteID, error) {

	minDis := math.MaxInt32
	minNeighbour := RouteID(-1)
	for n := range r.Neighbours {
		tmpAddr := r.RouterBase[n].AddrWithRoots[root]
		tmpDist :=  getDis(
			tmpAddr.Addr,
			dest, ADDR_LENGTH_INTERVAL)
		linkValue := 0.0
		if r.ID < n {
			link, ok := r.LinkBase[GetLinkKey(r.ID, n)]
			if ok {
				linkValue = link.Val1
			}
		} else {
			link, ok := r.LinkBase[GetLinkKey(n, r.ID)]
			if ok {
				linkValue = link.Val2
			}
		}
		if tmpDist < minDis && amount <= linkValue {
			minDis = tmpDist
			minNeighbour = n
		}
	}
	if minNeighbour == -1 {
		return -1, fmt.Errorf("cann't find a suitable next hop")
	}
	return minNeighbour, nil
}

// TODO(xuehan): add error return
func (r *SMRouter) updateLinkValue(from, to RouteID, value float64,
	flag int) {

	//var oldValue, newValue float64
	if from > to {
		linkKey := GetLinkKey(to, from)
		link, ok := r.LinkBase[linkKey]
		if ok {
			//oldValue = link.Val2
			if flag == ADD {
				link.Val2 += value
			} else {
				if link.Val2 >= value {
					link.Val2 -= value
				} else {
					//TODO(xuehan). log
					fmt.Printf("The fund: %v in the link: %v --> %v "+
						"is less the num: %v to sub", link.Val2, from, to, value)
					return
				}
			}
			//newValue = link.Val2
		} else { // 如果link本身不存在，那么只能加不能减
			//oldValue = 0
			if flag == ADD {
				r.LinkBase[linkKey] = &Link{
					Part1: to,
					Part2: from,
					Val1:  0,
					Val2:  value,
				}
				// 更新邻居信息
				if from == r.ID {
					r.Neighbours[to] = struct {}{}
					r.RouterBase[to].Neighbours[from] = struct{}{}
				} else {
					r.Neighbours[from] = struct{}{}
					r.RouterBase[from].Neighbours[to] = struct{}{}
				}
			} else {
				//TODO(xuehan). log
				fmt.Printf("The fund: %v in the link: %v --> %v "+
					"is less the num: %v to sub", link.Val2, from, to, value)
				return
			}
			//newValue = value
		}

	} else {
		linkKey := GetLinkKey(from, to)
		link, ok := r.LinkBase[linkKey]
		if ok {
			//oldValue = link.Val1
			if flag == ADD {
				link.Val1 += value
			} else {
				if link.Val1 >= value {
					link.Val1 -= value
				} else {
					//TODO(xuehan). log
					fmt.Printf("The fund: %v in the link: %v --> %v "+
						"is less the num: %v to sub", link.Val2, from, to, value)
					return
				}
			}
			//oldValue = link.Val1
		} else { // 如果link本身不存在，那么只能加不能减
			//oldValue = 0
			if flag == ADD {
				r.LinkBase[linkKey] = &Link{
					Part1: from,
					Part2: to,
					Val1:  value,
					Val2:  0,
				}

				// 更新邻居信息
				if from == r.ID {
					r.Neighbours[to] = struct {}{}
					r.RouterBase[to].Neighbours[from] = struct{}{}
				} else {
					r.Neighbours[from] = struct{}{}
					r.RouterBase[from].Neighbours[to] = struct{}{}
				}
			} else {
				//TODO(xuehan). log
				fmt.Printf("The fund: %v in the link: %v --> %v "+
					"is less the num: %v to sub", link.Val1, from, to, value)
				return
			}
			//newValue = value
		}
	}
	/*
	if r.ID == from {
		r.monitorLinkChange(oldValue, newValue, to)
	} else {
		r.monitorLinkChange(oldValue, newValue, from)
	}
	*/
}

// TODO(xuehan): 应该在真正支付时才被调用，所以update应该加个phase选项来区别对待。
func (r *SMRouter) monitorLinkChange(oldValue, newValue float64, neighbour RouteID) error {
	// 用来标记当前活着的root
	aliveRoots := make([]RouteID, 0)
	for _, id := range r.Roots {
		if _, ok := r.RouterBase[id]; ok {
			aliveRoots = append(aliveRoots, id)
		}
	}
	reset := make(map[RouteID]struct{})
	for _, aliveRoot := range aliveRoots {
		if oldValue == 0 && newValue > 0 {
			if _, ok := r.AddrWithRoots[aliveRoot]; !ok {
				reset[aliveRoot] = struct{}{}
				continue
			}
			Val1, err := r.getLinkValue(neighbour, LINK_DIR_RIGHT)
			if err != nil {
				return err
			}
			Val2, err := r.getLinkValue(neighbour, LINK_DIR_LEFT)
			if err != nil {
				return err
			}
			// 这里的条件实际上放宽了，除了判断其和父母的link value以外还应该判断邻居的
			// 和父母的link value是否是双向大于0的。
			if Val1 > 0 && Val2 > 0 {
				valToParent, err := r.getLinkValue(
					r.AddrWithRoots[aliveRoot].Parent, LINK_DIR_RIGHT)
				if err != nil {
					return err
				}
				valFromParent, err := r.getLinkValue(
					r.AddrWithRoots[aliveRoot].Parent, LINK_DIR_LEFT)
				if err != nil {
					return nil
				}
				if (valFromParent > 0 && valToParent == 0) ||
					(valFromParent == 0 && valToParent > 0) {
					reset[aliveRoot] = struct{}{}
					continue
				}
			}
		}

		if oldValue > 0 && newValue == 0 {
			if addr, ok := r.AddrWithRoots[aliveRoot]; ok && addr.Parent == neighbour {
				reset[aliveRoot] = struct{}{}
			}
		}
	}
	// 判断和邻居的链接是否都还大于0，如果两个都等于0，那么就从邻居集合中删掉
	var linkKey string
	if r.ID < neighbour {
		linkKey = GetLinkKey(r.ID, neighbour)
	} else {
		linkKey = GetLinkKey(neighbour, r.ID)
	}
	if r.LinkBase[linkKey].Val1 == 0 &&
		r.LinkBase[linkKey].Val2 == 0 {
		delete(r.Neighbours, neighbour)
		delete(r.RouterBase[neighbour].Neighbours, r.ID)
	}

	// 针对需要重构的地址，按root分别进行重构
	for root := range reset {
		r.resetAddr(root)
	}
	return nil
}

func (r *SMRouter) getLinkValue(neighbour RouteID, direction bool) (float64, error) {

	if r.ID == neighbour {
		return 0, fmt.Errorf("cann't get link value to self")
	}
	if r.ID < neighbour {
		linkKey := GetLinkKey(r.ID, neighbour)
		link, ok := r.LinkBase[linkKey]
		if !ok {
			return 0, nil
		} else {
			if direction == LINK_DIR_RIGHT {
				return link.Val1, nil
			} else {
				return link.Val2, nil
			}
		}
	} else {
		linkKey := GetLinkKey(neighbour, r.ID)
		link, ok := r.LinkBase[linkKey]
		if !ok {
			return 0, nil
		} else {
			if direction == LINK_DIR_RIGHT {
				return link.Val2, nil
			} else {
				return link.Val1, nil
			}
		}
	}
	return 0, nil
}

func (r *SMRouter) cleanProbe() {
	timeNow := time.Now().Unix()
	for requestID, probes := range r.probeBase {
		for root, probe := range probes {
			if timeNow-probe.time >= PROBE_STORE_TIME {
				r.updateLinkValue(r.ID, probe.nextHop,
					probe.value, ADD)
				delete(r.probeBase[requestID], root)
				if len(r.probeBase[requestID]) == 0 {
					delete(r.probeBase, requestID)
				}
			}
		}
	}
}

func (r *SMRouter) resetAddr(resetAddrRoot RouteID) error {
	reqID := RequestID(GetRandomString(ADDR_REQUESTID_LENGTH))
	for nei := range r.Neighbours {
		r.sendMsg(nei, &addrReq{
			reqSrc:  r.ID,
			reqRoot: resetAddrRoot,
			reqID:   reqID,
		})
	}

	uniLinkResponses := make([]*addrRes,0)
	biLinkResponses	:= make([]*addrRes,0)
	children := make([]RouteID,0)
	resNum := 0
out:
	for {
		select {
		case res := <- r.addrRequestPool[reqID]:
			resNum++
			resSrc := res.resSrc
			neiAddrBytes := []byte(res.addr)
			selfAddrBytes := []byte(r.AddrWithRoots[resetAddrRoot].Addr)
			// 判断是否是孩子节点发来的，如果是，则忽略，并且添加到children中，以通知其
			if bytes.Equal(selfAddrBytes,
				neiAddrBytes[0:len(neiAddrBytes)-ADDR_LENGTH_INTERVAL]) {
				children = append(children, res.resSrc)
				if resNum == len(r.Neighbours) {
					break out
				}
				continue
			}
			Val1, err := r.getLinkValue(resSrc,LINK_DIR_RIGHT)
			if err != nil {
				//TODO(xuehan): log
				fmt.Printf("faced error:%v", err)
			}
			Val2, err := r.getLinkValue(resSrc,LINK_DIR_LEFT)
			if err != nil {
				//TODO(xuehan): log
				fmt.Printf("faced error:%v", err)
			}
			if Val2 > 0 && Val1 > 0 {
				biLinkResponses = append(biLinkResponses, res)
			} else if (Val2 > 0 && Val1 ==0) || (Val1 > 0 && Val2 ==0) {
				uniLinkResponses = append(uniLinkResponses, res)
			}
			if resNum == len(r.Neighbours) {
				break out
			}
		case <- time.After(2 * time.Second):
			break out
		}
	}

	if len(biLinkResponses) != 0 {
		idx := rand.Intn(len(biLinkResponses))
		selectRes := biLinkResponses[idx]
		r.AddrWithRoots[resetAddrRoot] = &AddrType{
			Parent: selectRes.resSrc,
			Addr: selectRes.addr + GetRandomString(ADDR_LENGTH_INTERVAL),
		}
	} else if len(uniLinkResponses) != 0 {
		idx := rand.Intn(len(uniLinkResponses))
		selectRes := biLinkResponses[idx]
		r.AddrWithRoots[resetAddrRoot] = &AddrType{
			Parent: selectRes.resSrc,
			Addr: selectRes.addr + GetRandomString(ADDR_LENGTH_INTERVAL),
		}
	}
	// 通知邻居重置地址
	for _, child := range children {
		r.sendMsg(child, &addrResetNoti{
			src: r.ID,
			root: resetAddrRoot,
		})
	}
	return nil
}

func (r *SMRouter) onResetAddrReq(req *addrReq) {
	res := &addrRes{
		resSrc:  r.ID,
		resRoot: req.reqRoot,
		reqID:   req.reqID,
		addr:    r.AddrWithRoots[req.reqRoot].Addr,
	}
	r.sendMsg(req.reqSrc, res)
}

func (r *SMRouter) onResetAddrRes(res *addrRes) {
	r.addrRequestPool[res.reqID] <- res
}

func (r *SMRouter) onNotifyReset(noti *addrResetNoti)  {
	if r.AddrWithRoots[noti.root].Parent == noti.src {
		r.resetAddr(noti.root)
	}
}

func NewSMRouter(id RouteID, roots []RouteID,
	routerBase map[RouteID]*SMRouter,
	linkBase map[string]*Link) *SMRouter {

	router := &SMRouter{
		ID:              id,
		AddrWithRoots:   make(map[RouteID]*AddrType),
		Roots:           roots,
		Neighbours:      make(map[RouteID]struct{}),
		RouterBase:      routerBase,
		payRequestPool:  make(map[RequestID]chan *payRes),
		addrRequestPool: make(map[RequestID]chan *addrRes, ADDR_POOL_CAPCITY),
		probeBase:       make(map[RequestID]map[RouteID]*probeInfo),
		LinkBase:        linkBase,
		MsgPool:         make(chan interface{}, MSG_POOL_CAPCITY),
		timer:           time.NewTicker(CLEAN_PROBE_INTERVAL * time.Second),
		quit:            make(chan struct{}),
	}
	return router
}

func (r *SMRouter) AddLink (n RouteID, toN, fromN float64) {
	if n < r.ID {
		linkKey := GetLinkKey(n, r.ID)
		r.LinkBase[linkKey] = &Link{
			Part1: n,
			Part2: r.ID,
			Val1: fromN,
			Val2: toN,
		}
	} else {
		linkKey := GetLinkKey(r.ID, n)
		r.LinkBase[linkKey] = &Link{
			Part1: r.ID,
			Part2: n,
			Val1: toN,
			Val2: fromN,
		}
	}
	r.Neighbours[n] = struct{}{}
	r.RouterBase[n].Neighbours[r.ID] = struct{}{}
}

func (r *SMRouter)RemoveLink (n RouteID)  {
	if n > r.ID {
		linkKey := GetLinkKey(r.ID, n)
		delete(r.LinkBase, linkKey)
	} else {
		linkKey := GetLinkKey(n, r.ID)
		delete(r.LinkBase, linkKey)
	}
	delete(r.Neighbours, n)
	delete(r.RouterBase[n].Neighbours, r.ID)
}

func (r *SMRouter)GetLink (n RouteID) *Link {
	var linkKey string
	if n < r.ID {
		linkKey = GetLinkKey(n, r.ID)
		link, ok := r.LinkBase[linkKey]
		if ok {
			return link
		} else {
			return nil
		}
	} else {
		linkKey = GetLinkKey(r.ID, n)
		link, ok := r.LinkBase[linkKey]
		if ok {
			return link
		} else {
			return nil
		}
	}
	return nil
}

