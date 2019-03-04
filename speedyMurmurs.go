package main

import "fmt"

type RouteID int

type SMRouter struct {
	ID            RouteID
	AddrWithRoots map[RouteID]string
	Roots         []RouteID
	Neighbours    []RouteID
	RouteBase     map[RouteID]*SMRouter
	LinkBase      map[string]*Link
	MsgPool       chan *Message
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
	part1  RouteID
	part2 RouteID
	val1  float64
	val2  float64
}

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

func (r *SMRouter) sendMsg(id RouteID, msg *Message) {
	r.RouteBase[id].MsgPool <- msg
}

func (r *SMRouter) onMsg(msg *Message) {

}

func (r *RouteID)onLinkAdd(add *Link)  {

}




