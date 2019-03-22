package silentWhisper

import (
	"bytes"
	"fmt"
	"time"
)

type RouteID int
type RequestID string

const (
	UP             = true
	DOWN           = false
	LINK_DIR_RIGHT = true
	LINK_DIR_LEFT  = false
	ADD            = 0
	SUB            = 1
)

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

type addrType struct {
	addr   string
	parent RouteID
	time   int64
}

type payReq struct {
	requestID RequestID
	root      RouteID
	sender    RouteID
	dest      string
	path      []RouteID
	upOrDown  bool
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
}

type htlcFullfill struct {
	requestID RequestID
	root      RouteID
	success   bool
	reason    string
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
	htlcPool 	   map[RequestID]chan *htlcFullfill
	LinkBase       map[string]*Link
	MsgPool        chan interface{}
	timer          *time.Ticker
	quit           chan struct{}
}

func (r *SWRouter) start() {
	for {
		select {
		case msg := <-r.MsgPool:
			r.onMsg(msg)
		case <-r.quit:
			fmt.Printf("")
			return
		}
	}
}

func (r *SWRouter) stop() {
	close(r.quit)
}

func (r *SWRouter) onMsg(msg interface{}) {
	switch msg.(type) {
	case *payRes:
		r.onPayRes(msg.(*payRes))
	case *payReq:
		r.onPayReq(msg.(*payReq))
	case *htlc:
		r.onHTLC(msg.(*htlc))
	case *htlcFullfill:
		r.onHTLCFullfill(msg.(*htlcFullfill))
	case *addrWithRoot:
		r.onAddrWithRoot(msg.(*addrWithRoot))
	}
}

func (r *SWRouter) onPayReq(req *payReq) {
	// 到目的地了
	if req.dest == r.AddrWithRoots[req.root].addr {
		res := &payRes{
			path:      append(req.path, r.ID),
			requestID: req.requestID,
			root:      req.root,
			success:   true,
		}
		r.sendMsg(req.sender, res)
	} else {
		// 如果到了root节点，则修改传播方向为down
		if req.root == r.ID {
			req.upOrDown = DOWN
		}
		nextHop := r.getNextHop(req.dest, req.root, req.upOrDown)
		req.path = append(req.path, r.ID)

		if nextHop != -1 {
			r.sendMsg(nextHop, req)
		} else {
			r.sendMsg(req.sender, &payRes{
				success:   false,
				requestID: req.requestID,
				path:      req.path,
			})
		}
	}
}

func (r *SWRouter) onPayRes(res *payRes) {
	r.payRequestPool[res.requestID] <- res
}

func (r *SWRouter) onHTLC(htlc *htlc) {
	r.htlcBase[htlc.requestID][htlc.root] = htlc

	value := htlc.amount
	requestID := htlc.requestID
	root := htlc.root
	path := htlc.path
	index, err := findIndexInPath(r.ID, path)
	if err != nil {
		fmt.Printf("faced error :%v\n", err)
	}
	// 到达目的地了
	if index == len(path) - 1 {
		hff := &htlcFullfill{
			success: true,
			requestID: requestID,
			root: root,
		}
		r.sendMsg(path[index-1],hff)
		return
	}

	// 还在半路上
	htlc.upper = r.ID
	err = r.updateLinkValue(r.ID, path[index+1], value, SUB)
	// 钱不够了，那么开始回滚
	if err != nil {
		hff := &htlcFullfill{
			success: false,
			requestID: requestID,
			root: root,
			reason: err.Error(),
		}
		r.sendMsg(path[index-1], hff)
	} else {
		r.sendMsg(path[index + 1], htlc)
	}
}

func findIndexInPath(id RouteID, path []RouteID) (int, error) {
	for index, node := range path{
		if node == id {
			return index, nil
		}
	}
	return 0, fmt.Errorf("not in the path")
}


func (r *SWRouter) onHTLCFullfill(hff *htlcFullfill) {
	htlc := r.htlcBase[hff.requestID][hff.root]
	index, err := findIndexInPath(r.ID, htlc.path)
	if err != nil {
		fmt.Printf("faced err:%v ", err)
		return
	}
	// 如果不是
	if index != 0 {
		if hff.success == true {
			r.updateLinkValue(r.ID, htlc.upper, htlc.amount, ADD)
		} else {
			r.updateLinkValue(htlc.upper, r.ID, htlc.amount, SUB)
		}
		r.sendMsg(htlc.upper, hff)
	} else {
		r.htlcPool[hff.requestID] <- hff
		return
	}
}

func (r *SWRouter) onAddrWithRoot(awr *addrWithRoot) {
	addr := r.AddrWithRoots[awr.root]
	changed := false
	if addr == nil || addr.time < awr.time {
		changed = true

	} else if addr.time == awr.time &&
		len(addr.addr) > len(awr.addr) + 4 {
		changed = true
	}
	if changed {
		addr = &addrType{
			time:   awr.time,
			addr:   awr.addr + GetRandomString(4),
			parent: awr.src,
		}
		for nei := range r.Neighbours {
			r.sendMsg(nei, &addrWithRoot{
				root: awr.root,
				addr: addr.addr,
				time: addr.time,
				src:  r.ID,
			})
		}
	}
}

func NotifyRooterReset(roots []RouteID, routerBase []SWRouter) {
	for _, root := range roots{
		rootRouter := routerBase[root]
		rootRouter.AddrWithRoots[root] = &addrType{
			addr: "",
			time: time.Now().Unix(),
		}
		for n := range rootRouter.Neighbours {
			rootRouter.sendMsg(n, addrWithRoot{
				root: root,
				addr: "",
				time: time.Now().Unix(),
				src: root,
			})
		}
	}
}

func (r *SWRouter) getLinkValue(neighbour RouteID, direction bool) (float64, error) {

	if r.ID == neighbour {
		return 0, fmt.Errorf("cann't get link value to self")
	}
	if r.ID < neighbour {
		linkKey := getLinkKey(r.ID, neighbour)
		link, ok := r.LinkBase[linkKey]
		if !ok {
			return 0, nil
		} else {
			if direction == LINK_DIR_RIGHT {
				return link.val1, nil
			} else {
				return link.val2, nil
			}
		}
	} else {
		linkKey := getLinkKey(neighbour, r.ID)
		link, ok := r.LinkBase[linkKey]
		if !ok {
			return 0, nil
		} else {
			if direction == LINK_DIR_RIGHT {
				return link.val2, nil
			} else {
				return link.val1, nil
			}
		}
	}
	return 0, nil
}

func (r *SWRouter) getNextHop(dest string, root RouteID,
	upOrDown bool) RouteID {
	nextHop := RouteID(-1)
	if upOrDown == UP {
		return r.AddrWithRoots[root].parent
	} else {

		// TODO(xuehan): 这里应该改成从邻居实时pull地址
		bestCpl := getCPL(dest, r.AddrWithRoots[root].addr, 4)
		for n := range r.Neighbours {
			cpl := getCPL(r.RouterBase[n].AddrWithRoots[root].addr,
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
	r.RouterBase[id].MsgPool <- msg
}

func NewSWRouter() *SWRouter {
	return nil
}

// TODO(xuehan): add error return
func (r *SWRouter) updateLinkValue(from, to RouteID, value float64,
	flag int) error {

	//var oldValue, newValue float64
	if from > to {
		linkKey := getLinkKey(to, from)
		link, ok := r.LinkBase[linkKey]
		if ok {
			//oldValue = link.val2
			if flag == ADD {
				link.val2 += value
			} else {
				if link.val2 >= value {
					link.val2 -= value
				} else {
					//TODO(xuehan). log
					return fmt.Errorf("the fund: %v in the link: %v --> %v "+
						"is less the num: %v to sub", link.val2, from, to, value)
				}
			}
			//newValue = link.val2
		} else { // 如果link本身不存在，那么只能加不能减
			//oldValue = 0
			if flag == ADD {
				r.LinkBase[linkKey] = &Link{
					part1: to,
					part2: from,
					val1:  0,
					val2:  value,
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
				return fmt.Errorf("the fund: %v in the link: %v --> %v "+
					"is less the num: %v to sub", link.val2, from, to, value)

			}
			//newValue = value
		}
	} else {
		linkKey := getLinkKey(from, to)
		link, ok := r.LinkBase[linkKey]
		if ok {
			if flag == ADD {
				link.val1 += value
			} else {
				if link.val1 >= value {
					link.val1 -= value
				} else {
					//TODO(xuehan). log
					return fmt.Errorf("the fund: %v in the link: %v --> %v "+
						"is less the num: %v to sub", link.val1, from, to, value)
				}
			}
		} else { // 如果link本身不存在，那么只能加不能减
			if flag == ADD {
				r.LinkBase[linkKey] = &Link{
					part1: from,
					part2: to,
					val1:  value,
					val2:  0,
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
				return fmt.Errorf("the fund: %v in the link: %v --> %v "+
					"is less the num: %v to sub", link.val1, from, to, value)
			}
		}
	}
	return nil
}

func (r *SWRouter) AddLink (n RouteID, toN, fromN float64) {
	if n < r.ID {
		linkKey := getLinkKey(n, r.ID)
		r.LinkBase[linkKey] = &Link{
			part1: n,
			part2: r.ID,
			val1: fromN,
			val2: toN,
		}
	} else {
		linkKey := getLinkKey(r.ID, n)
		r.LinkBase[linkKey] = &Link{
			part1: r.ID,
			part2: n,
			val1: toN,
			val2: fromN,
		}
	}
	r.Neighbours[n] = struct{}{}
	r.RouterBase[n].Neighbours[r.ID] = struct{}{}
}

func (r *SWRouter)RemoveLink (n RouteID)  {
	if n > r.ID {
		linkKey := getLinkKey(r.ID, n)
		delete(r.LinkBase, linkKey)
	} else {
		linkKey := getLinkKey(n, r.ID)
		delete(r.LinkBase, linkKey)
	}
	delete(r.Neighbours, n)
	delete(r.RouterBase[n].Neighbours, r.ID)
}

func (r *SWRouter)GetLink (n RouteID) *Link {
	var linkKey string
	if n < r.ID {
		linkKey = getLinkKey(n, r.ID)
		link, ok := r.LinkBase[linkKey]
		if ok {
			return link
		} else {
			return nil
		}
	} else {
		linkKey = getLinkKey(r.ID, n)
		link, ok := r.LinkBase[linkKey]
		if ok {
			return link
		} else {
			return nil
		}
	}
	return nil
}
