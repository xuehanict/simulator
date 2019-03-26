package silentWhisper

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
	//"github.com/davecgh/go-spew/spew"
)

type RouteID int
type RequestID string

const (
	UP              = true
	DOWN            = false
	LINK_DIR_RIGHT  = true
	LINK_DIR_LEFT   = false
	ADD             = 0
	SUB             = 1
	HTLC_CLEAR_TIME = 10
	MESSAGE_BUFFER  = 10000
	PAY_RES_BUFFER  = 10
)

/*
 * val1 指part1往part2方向的通道容量
 * val2 指part2往part1方向的通道容量
 * part1 的id 小于part2的id
 */
type Link struct {
	Part1 RouteID
	Part2 RouteID
	Val1  float64
	Val2  float64
}

type addrType struct {
	Addr   string
	Parent RouteID
	Time   int64
}

type payReq struct {
	requestID RequestID
	root      RouteID
	sender    RouteID
	dest      string
	path      []RouteID
	upOrDown  bool
	value     float64
}

type payRes struct {
	requestID RequestID
	root      RouteID
	success   bool
	path      []RouteID
	value     float64
}

type htlc struct {
	requestID RequestID
	amount    float64
	root      RouteID
	upper     RouteID
	path      []RouteID
	sender    RouteID
	ffOnece   bool
}

type htlcFullfill struct {
	requestID RequestID
	root      RouteID
	success   bool
	reason    string
	sender    RouteID
}

type addrWithRoot struct {
	root RouteID
	addr string
	time int64
	src  RouteID
}

type SWRouter struct {
	ID             RouteID
	AddrWithRoots  map[RouteID]*addrType
	Roots          []RouteID
	Neighbours     map[RouteID]struct{}
	RouterBase     map[RouteID]*SWRouter
	payRequestPool map[RequestID]chan *payRes
	htlcBase       map[RequestID]map[RouteID]*htlc
	htlcPool       map[RequestID]chan *htlcFullfill
	LinkBase       map[string]*Link
	MsgPool        chan interface{}
	timer          *time.Ticker
	quit           chan struct{}
}

func NewSwRouter(id RouteID, roots []RouteID,
	routerBase map[RouteID]*SWRouter,
	linkBase map[string]*Link) *SWRouter {
	router := &SWRouter{
		ID:             id,
		AddrWithRoots:  make(map[RouteID]*addrType),
		Roots:          roots,
		Neighbours:     make(map[RouteID]struct{}),
		RouterBase:     routerBase,
		payRequestPool: make(map[RequestID]chan *payRes),
		htlcBase:       make(map[RequestID]map[RouteID]*htlc),
		htlcPool:       make(map[RequestID]chan *htlcFullfill),
		LinkBase:       linkBase,
		MsgPool:        make(chan interface{}, 1000),
		timer:          time.NewTicker(HTLC_CLEAR_TIME * time.Second),
		quit:           make(chan struct{}),
	}
	return router
}

func (r *SWRouter) Start() {
	mesNum := 0
	for {
		mesNum++
		select {
		case msg := <-r.MsgPool:
			r.onMsg(msg)
			mesNum--
		case <-r.quit:
			r.Printf("mesNum is %v\n", mesNum)
			return
		}
	}
}

func (r *SWRouter) Stop() {
	close(r.quit)
}

func (r *SWRouter) onMsg(msg interface{}) {
	switch msg.(type) {
	case *payRes:
		r.onPayRes(msg.(*payRes))
	case *payReq:
		//r.Printf("收到req\n")
		r.onPayReq(msg.(*payReq))
	case *htlc:
		r.onHTLC(msg.(*htlc))
	case *htlcFullfill:
		r.onHTLCFullfill(msg.(*htlcFullfill))
	case *addrWithRoot:
		//r.Printf("调用onAddrWithRoot\n")
		r.onAddrWithRoot(msg.(*addrWithRoot))
	}
}

func (r *SWRouter) onPayReq(req *payReq) {
	path := append(req.path, r.ID)
	rootIndex := findIndexInPath(req.root, path, true)
	// 到目的地了,并且已经经过了root节点
	if req.dest == r.AddrWithRoots[req.root].Addr &&
		rootIndex != -1 {
		res := &payRes{
			path:      path,
			requestID: req.requestID,
			root:      req.root,
			success:   true,
			value:     req.value,
		}
		r.sendMsg(req.sender, res)
	} else {
		// 如果到了root节点，则修改传播方向为down
		if req.root == r.ID {
			req.upOrDown = DOWN
		}
		nextHop := r.GetNextHop(req.dest, req.root, req.upOrDown)
		//req.path = append(req.path, r.ID)
		linkValue, err := r.getLinkValue(nextHop, LINK_DIR_RIGHT)
		if nextHop == -1 || err != nil {
			r.sendMsg(req.sender, &payRes{
				success:   false,
				requestID: req.requestID,
				path:      path,
			})
		} else {
			if req.value > linkValue {
				req.value = linkValue
			}
			req.path = path
			r.sendMsg(nextHop, req)
		}
	}
}

func (r *SWRouter) onPayRes(res *payRes) {
	r.payRequestPool[res.requestID] <- res
}

func (r *SWRouter) onHTLC(h *htlc) {
	var dir bool
	if _, ok := r.htlcBase[h.requestID]; !ok {
		r.htlcBase[h.requestID] = make(map[RouteID]*htlc)
		dir = true
	} else {
		dir = false
	}
	r.htlcBase[h.requestID][h.root] = newHTLC(h)
	value := h.amount
	requestID := h.requestID
	root := h.root
	path := h.path
	index := findIndexInPath(r.ID, path, dir)

	// 到达目的地了
	if index == len(path)-1 {
		hff := &htlcFullfill{
			success:   true,
			requestID: requestID,
			root:      root,
			sender:    h.sender,
		}
		r.updateLinkValue(r.ID, path[index-1], value, ADD)
		r.sendMsg(path[index-1], hff)
		return
	}

	// 还在半路上
	forwardHTLC := newHTLC(h)
	forwardHTLC.upper = r.ID
	h.upper = r.ID
	err := r.updateLinkValue(r.ID, path[index+1], value, SUB)
	// 钱不够了，那么开始回滚
	if err != nil {
		hff := &htlcFullfill{
			success:   false,
			requestID: requestID,
			root:      root,
			reason:    err.Error(),
		}
		r.sendMsg(path[index-1], hff)
	} else {
		r.sendMsg(path[index+1], forwardHTLC)
	}
}

func (r *SWRouter) SendPayment(dest RouteID, amount float64) error {

	//splittedAmounts := Partition(amount, len(r.Roots))
	neighboursToSend := make([]RouteID, len(r.Roots))

	requestID := RequestID(GetRandomString(10))
	r.payRequestPool[requestID] = make(chan *payRes, PAY_RES_BUFFER)
	r.htlcPool[requestID] = make(chan *htlcFullfill, PAY_RES_BUFFER)
	for i, root := range r.Roots {
		destAddr := r.RouterBase[dest].AddrWithRoots[root]
		var dir bool
		if r.ID == root {
			dir = DOWN
		} else {
			dir = UP
		}
		nextHop := r.GetNextHop(destAddr.Addr, root, dir)
		if nextHop == -1 {
			return fmt.Errorf("send payment failed, " +
				"cann't find next hop")
		}
		neighboursToSend[i] = nextHop
		value, err := r.getLinkValue(nextHop, LINK_DIR_RIGHT)
		if err != nil {
			return err
		}
		// ps: 这里的地址是直接从dest节点中拿出的，实际场景应该从
		payreq := &payReq{
			sender:    r.ID,
			requestID: requestID,
			root:      root,
			dest:      destAddr.Addr,
			path:      make([]RouteID, 0),
			upOrDown:  UP,
			value:     value,
		}
		payreq.path = append(payreq.path, r.ID)
		r.sendMsg(nextHop, payreq)
	}

	resArray := make([]*payRes, 0)
	mins := make([]float64, 0)
out:
	for {
		select {
		case res := <-r.payRequestPool[requestID]:
			if res.success == false {
				return fmt.Errorf("probe failed\n")
			}
			resArray = append(resArray, res)
			mins = append(mins, res.value)
			if len(resArray) == len(r.Roots) {
				//spew.Dump(resArray)
				break out
			}
		case <-time.After(20 * time.Second):
			return fmt.Errorf("probe failed, timeout")
		}
	}

	splitedAmts := minPart(amount, mins)
	if splitedAmts == nil {
		return fmt.Errorf("钱不够\n")
	}
	r.Printf("分成多少钱：%v\n", splitedAmts)
	r.htlcBase[requestID] = make(map[RouteID]*htlc)
	for i, amt := range splitedAmts {
		htlc := &htlc{
			amount:    amt,
			root:      resArray[i].root,
			path:      resArray[i].path,
			upper:     r.ID,
			sender:    r.ID,
			requestID: resArray[i].requestID,
			ffOnece:   true,
		}
		err := r.updateLinkValue(r.ID, resArray[i].path[1], amt, SUB)
		if err != nil {
			return err
		}
		r.htlcBase[requestID][resArray[i].root] = htlc
		r.sendMsg(resArray[i].path[1], htlc)
	}

	hffLen := 0
	for {
		select {
		case hff := <-r.htlcPool[requestID]:
			if hff.success == false {
				return fmt.Errorf("payment failed")
			}
			hffLen++
			if hffLen == len(r.Roots) {
				return nil
			}
		case <-time.After(1 * time.Second):
			return fmt.Errorf(" timeout for payment")
		}
	}
	return nil
}

func newHTLC(h *htlc) *htlc {
	return &htlc{
		amount:    h.amount,
		root:      h.root,
		path:      h.path,
		upper:     h.upper,
		sender:    h.sender,
		requestID: h.requestID,
	}
}

func findIndexInPath(id RouteID, path []RouteID, dir bool) int {

	if dir == true {
		for index, node := range path {
			if node == id {
				return index
			}
		}
	} else {
		for i := len(path) - 1; i >= 0; i-- {
			if path[i] == id {
				return i
			}
		}
	}

	return -1
}

func (r *SWRouter) onHTLCFullfill(hff *htlcFullfill) {

	htlc := r.htlcBase[hff.requestID][hff.root]
	//	index, err := findIndexInPath(r.ID, htlc.path)
	//	if err != nil {
	//		fmt.Printf("faced err:%v ", err)
	//		return
	//	}
	// 如果不是
	//r.Printf("这是：%v的 onHTLCFUllFill\n", r.ID)

	index := findIndexInPath(r.ID, htlc.path, !htlc.ffOnece)

	if r.ID == htlc.sender && index == 0 {
		r.htlcPool[hff.requestID] <- hff
		return
	} else {
		var upperID RouteID
		if index == 0 {
			upperID = htlc.sender
		} else {
			upperID = htlc.path[index-1]
		}
		if hff.success == true {
			r.updateLinkValue(r.ID, upperID, htlc.amount, ADD)
		} else {
			r.updateLinkValue(upperID, r.ID, htlc.amount, ADD)
		}
		htlc.ffOnece = false
		r.sendMsg(upperID, hff)
	}
}

func (r *SWRouter) onAddrWithRoot(awr *addrWithRoot) {
	addr, ok := r.AddrWithRoots[awr.root]
	changed := false
	if !ok || addr.Time < awr.time {
		changed = true
	} else if addr.Time == awr.time &&
		len(addr.Addr) > len(awr.addr)+4 {
		changed = true
	}
	if changed {
		addr = &addrType{
			Time:   awr.time,
			Addr:   awr.addr + GetRandomString(4),
			Parent: awr.src,
		}
		r.AddrWithRoots[awr.root] = addr
		r.Printf("修改了root%v的地址\n", awr.root)
		for nei := range r.Neighbours {
			r.sendMsg(nei, &addrWithRoot{
				root: awr.root,
				addr: addr.Addr,
				time: awr.time,
				src:  r.ID,
			})
		}
	}
}

func NotifyRooterReset(roots []RouteID, routerBase map[RouteID]*SWRouter) {
	time_ := time.Now().Unix()
	for _, root := range roots {
		rootRouter := routerBase[root]
		rootRouter.AddrWithRoots[root] = &addrType{
			Addr: "",
			Time: time_,
		}
		for n := range rootRouter.Neighbours {
			rootRouter.sendMsg(n, &addrWithRoot{
				root: root,
				addr: "",
				time: time_,
				src:  root,
			})
		}
		time.Sleep(3 * time.Second)
	}
}

func (r *SWRouter) getLinkValue(neighbour RouteID, direction bool) (float64, error) {

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

func (r *SWRouter) GetNextHop(dest string, root RouteID,
	upOrDown bool) RouteID {
	nextHop := RouteID(-1)
	if upOrDown == UP {
		return r.AddrWithRoots[root].Parent
	} else {
		// TODO(xuehan): 这里应该改成从邻居实时pull地址
		bestCpl := getCPL(dest, r.AddrWithRoots[root].Addr, 4)
		for n := range r.Neighbours {
			cpl := getCPL(r.RouterBase[n].AddrWithRoots[root].Addr,
				dest, 4)
			if cpl > bestCpl {
				return n
			}
		}
	}
	return nextHop
}

func getCPL(addr1, addr2 string, interval int) int {
	cpl := 0
	addr1Bytes := []byte(addr1)
	addr2Bytes := []byte(addr2)
	for ; cpl < len(addr1)/interval && cpl < len(addr2)/interval &&
		bytes.Equal(addr1Bytes[0:interval],
			addr2Bytes[0:interval]); cpl++ {
		addr1Bytes = addr1Bytes[interval:]
		addr2Bytes = addr2Bytes[interval:]
	}
	return cpl
}

func (r *SWRouter) sendMsg(id RouteID, msg interface{}) {
	//r.Printf("发送消息:%v 到%v\n", msg, id)
	r.RouterBase[id].MsgPool <- msg
}

func (r *SWRouter) Printf(string string, a ...interface{}) {
	fmt.Printf("R"+strconv.Itoa(int(r.ID))+":"+string, a...)
}

// TODO(xuehan): add error return
func (r *SWRouter) updateLinkValue(from, to RouteID, value float64,
	flag int) error {

	//var oldValue, newValue float64
	if from > to {
		linkKey := GetLinkKey(to, from)
		link, ok := r.LinkBase[linkKey]
		if ok {
			//oldValue = link.val2
			if flag == ADD {
				link.Val2 += value
			} else {
				if link.Val2 >= value {
					link.Val2 -= value
				} else {
					//TODO(xuehan). log
					return fmt.Errorf("the fund: %v in the link: %v --> %v "+
						"is less the num: %v to sub", link.Val2, from, to, value)
				}
			}
			//newValue = link.val2
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
					r.Neighbours[to] = struct{}{}
					r.RouterBase[to].Neighbours[from] = struct{}{}
				} else {
					r.Neighbours[from] = struct{}{}
					r.RouterBase[from].Neighbours[to] = struct{}{}
				}
			} else {
				//TODO(xuehan). log
				return fmt.Errorf("the fund: %v in the link: %v --> %v "+
					"is less the num: %v to sub", link.Val2, from, to, value)

			}
			//newValue = value
		}
	} else {
		linkKey := GetLinkKey(from, to)
		link, ok := r.LinkBase[linkKey]
		if ok {
			if flag == ADD {
				link.Val1 += value
			} else {
				if link.Val1 >= value {
					link.Val1 -= value
				} else {
					//TODO(xuehan). log
					return fmt.Errorf("the fund: %v in the link: %v --> %v "+
						"is less the num: %v to sub", link.Val1, from, to, value)
				}
			}
		} else { // 如果link本身不存在，那么只能加不能减
			if flag == ADD {
				r.LinkBase[linkKey] = &Link{
					Part1: from,
					Part2: to,
					Val1:  value,
					Val2:  0,
				}
				// 更新邻居信息
				if from == r.ID {
					r.Neighbours[to] = struct{}{}
					r.RouterBase[to].Neighbours[from] = struct{}{}
				} else {
					r.Neighbours[from] = struct{}{}
					r.RouterBase[from].Neighbours[to] = struct{}{}
				}
			} else {
				//TODO(xuehan). log
				return fmt.Errorf("the fund: %v in the link: %v --> %v "+
					"is less the num: %v to sub", link.Val1, from, to, value)
			}
		}
	}
	return nil
}

func (r *SWRouter) AddLink(n RouteID, toN, fromN float64) {
	if n < r.ID {
		linkKey := GetLinkKey(n, r.ID)
		r.LinkBase[linkKey] = &Link{
			Part1: n,
			Part2: r.ID,
			Val1:  fromN,
			Val2:  toN,
		}
	} else {
		linkKey := GetLinkKey(r.ID, n)
		r.LinkBase[linkKey] = &Link{
			Part1: r.ID,
			Part2: n,
			Val1:  toN,
			Val2:  fromN,
		}
	}
	r.Neighbours[n] = struct{}{}
	r.RouterBase[n].Neighbours[r.ID] = struct{}{}
}

func (r *SWRouter) RemoveLink(n RouteID) {
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

func (r *SWRouter) GetLink(n RouteID) *Link {
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
