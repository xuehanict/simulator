package silentWhisper

import (
	"fmt"
	"time"
	"bytes"
)

type RouteID int
type RequestID string

const (
	UP = true
	DOWN = false
	LINK_DIR_RIGHT = true
	LINK_DIR_LEFT = false
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
	value 	  float64
}

type htlc struct {
	requestID RequestID
	amount    float64
}

type htlcFullfill struct {
	requestID RequestID
	success bool
	reason string
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
		res := & payRes{
			path: append(req.path, r.ID),
			requestID: req.requestID,
			root: req.root,
			success: true,
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
				success: false,
				requestID: req.requestID,
				path: req.path,
			})
		}
	}
}

func (r *SWRouter) onPayRes(res *payRes) {
	r.payRequestPool[res.requestID] <- res
}

func (r *SWRouter) onHTLC(htlc *htlc) {

}

func (r *SWRouter) onHTLCFullfill(hff *htlcFullfill)  {

}


func (r *SWRouter) onAddrWithRoot(awr *addrWithRoot) {
	addr := r.AddrWithRoots[awr.root]
	changed := false
	if addr == nil || addr.time < awr.time {
		changed = true

	} else if addr.time == awr.time &&
		len(addr.addr) > len(awr.addr)-4 {
		changed = true
	}
	if changed {
		addr = &addrType{
			time:   awr.time,
			addr:   awr.addr + GetRandomString(4),
			parent: awr.src,
		}
		for  nei := range r.Neighbours {
			r. sendMsg(nei, &addrWithRoot{
				root: awr.root,
				addr: addr.addr,
				time: addr.time,
				src: r.ID,
			})
		}
	}
}

func (r *SWRouter) notifyRooterReset(roots []RequestID) {

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

func (r *SWRouter) getNextHop (dest string, root RouteID,
	upOrDown bool) RouteID {
	nextHop := RouteID(-1)
	if upOrDown == UP {
		return r.AddrWithRoots[root].parent
	} else {

		// TODO(xuehan): 这里应该改成从邻居实时pull地址
		bestCpl := getCPL(dest, r.AddrWithRoots[root].addr, 4)
		for n := range  r.Neighbours {
			cpl := getCPL(r.RouterBase[n].AddrWithRoots[root].addr,
				dest, 4)
			if cpl > bestCpl {
				return n
			}
		}
	}
	return nextHop
}

func getCPL(addr1, addr2 string, interval int) int{
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
