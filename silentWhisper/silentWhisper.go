package silentWhisper

import (
	"fmt"
	"time"
)

type RouteID int
type RequestID string

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
	value     float64
	upperHop  RouteID
}

type payRes struct {
	requestID RequestID
	root      RouteID
	sender    RouteID
	success   bool
	val       float64
}

type payment struct {
	requestID RequestID
	amount    float64
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
	case *payment:
		r.onPayment(msg.(*payment))
	case *addrWithRoot:
		r.onAddrWithRoot(msg.(*addrWithRoot))
	}
}

func (r *SWRouter) onPayReq(req *payReq) {

}

func (r *SWRouter) onPayRes(res *payRes) {

}

func (r *SWRouter) onPayment(payment *payment) {

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

func (r *SWRouter) sendMsg(id RouteID, msg interface{}) {
	r.RouterBase[id].MsgPool <- msg
}

func NewSWRouter() *SWRouter {
	return nil
}
