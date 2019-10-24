package aomdv

import (
	"container/list"
	"github.com/lightningnetwork/simulator/utils"
	"time"
)

type Aomdv struct {
	*utils.Graph

	// 路径表，第一个是key是src， 第二个key是dest
	RoutingPaths map[utils.RouterID]map[utils.RouterID][]utils.Path

	// 当前支付请求池, 因为每次只处理一个request，因此前面的key就代表节点id
	requestPool map[utils.RouterID]struct{}

	// AOMDV所需的，构建不相交的路径的需要
	firstHopList map[utils.RouterID]map[utils.RouterID]struct{}

	//
	adversed_hop map[utils.RouterID]int
}

type node struct {




}




type RREQ struct {
	// 这些是信息是rreq所必须的信息
	// 因为我们是在单线程情况下，不存在并发的情况，因此不用标识broadcast_id
	source         utils.RouterID
	destination    utils.RouterID
	amt            utils.Amount
	sequenceNumber int64

	// 以下三个则是模拟需要增加的
	sdr      utils.RouterID
	curr     utils.RouterID
	firstHop utils.RouterID
}


type RREP struct {
	path []utils.RouterID
}


func copyRREQ(r *RREQ) *RREQ {
	return &RREQ{
		source:         r.source,
		destination:    r.destination,
		amt:            r.amt,
		sequenceNumber: r.sequenceNumber,
		sdr:            r.sdr,
		curr:           r.curr,
		firstHop:       r.firstHop,
	}
}

func (a *Aomdv) FindPaths(src, dest utils.RouterID, amt utils.Amount) []utils.Path {

	rreqList := list.New()
	rreq := &RREQ{
		source:         src,
		destination:    dest,
		amt:            amt,
		sequenceNumber: time.Now().UnixNano(),
		sdr:            src,
		firstHop:       src,
		curr:src,
	}
	rreqList.PushFront(rreq)
	for rreqList.Len() != 0 {
		tmpRREQ := rreqList.Front().Value.(*RREQ)
		a.handleRREQ(tmpRREQ)
		nextHops := a.getNextRREQNextHop(tmpRREQ)

		for _, nextHop := range nextHops {
			if _, ok := a.requestPool[nextHop]; ok {
				continue
			}
			newRREQ := copyRREQ(tmpRREQ)
			newRREQ.sdr = tmpRREQ.curr
			newRREQ.curr = nextHop
			rreqList.PushBack(newRREQ)
		}

	}

	return nil
}

func (a *Aomdv)handleRREQ(r *RREQ)  {

}

func (a *Aomdv) getNextRREQNextHop(r *RREQ) []utils.RouterID {
	return nil
}

func (a *Aomdv) SendPayment(src, dest utils.RouterID, amt utils.Amount) error {
	time.Now().UnixNano()
	return nil
}

func (a *Aomdv) nextHop(current utils.RouterID, currPath utils.Path) []utils.Path {

	return nil
}
