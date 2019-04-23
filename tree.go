package main

import (
	"container/list"
	sw "github.com/lightningnetwork/simulator/silentWhisper"
	"time"
	sm "github.com/lightningnetwork/simulator/speedymurmurs"
)

func createTree(routers map[sw.RouteID]*sw.SWRouter, links map[string]*sw.Link, roots []sw.RouteID) {
	for _, root := range roots {
		rootNode := routers[root]
		timeNow := time.Now().Unix()
		queue := list.New()
		queue.PushBack(rootNode)
		rootNode.AddrWithRoots[root] = &sw.AddrType{
			Addr:   "",
			Parent: -1,
			Time:   timeNow,
		}
		bi := true
		assigned := make(map[sw.RouteID]struct{})
		assigned[root] = struct{}{}
		for {
			if queue.Len() == 0 {
				break
			}
			head := queue.Front()
			queue.Remove(head)
			node := head.Value.(*sw.SWRouter)
			addr := node.AddrWithRoots[root]
			for n := range node.Neighbours {
				if _, ok := routers[n].AddrWithRoots[root]; !ok {
					var linkKey string
					if node.ID < n {
						linkKey = sw.GetLinkKey(node.ID, n)
					} else {
						linkKey = sw.GetLinkKey(n, node.ID)
					}
					link := links[linkKey]
					if (link.Val1 > 0 && link.Val2 > 0) || !bi {
						routers[n].AddrWithRoots[root] = &sw.AddrType{
							Parent: node.ID,
							Addr:   addr.Addr + sw.GetRandomString(sw.ADDR_INTERVAL),
							Time:   timeNow,
						}
						assigned[n] = struct{}{}
						queue.PushBack(routers[n])
					}
				}
			}
			if queue.Len() == 0 && bi {
				bi = false
				for n := range assigned {
					queue.PushBack(routers[n])
				}
			}
		}
		sw.SWLogger.Printf("总共有地址的节点数目是：%v\n", len(assigned))
	}
}

func createTreeSM(routers map[sm.RouteID]*sm.SMRouter, links map[string]*sm.Link, roots []sm.RouteID) {

	for _, root := range roots {
		rootNode := routers[root]
		queue := list.New()
		queue.PushBack(rootNode)
		rootNode.AddrWithRoots[root] = &sm.AddrType{
			Addr:   "",
			Parent: -1,
		}
		bi := true
		assigned := make(map[sm.RouteID]struct{})
		assigned[root] = struct{}{}
		for {
			if queue.Len() == 0 {
				break
			}
			head := queue.Front()
			queue.Remove(head)
			node := head.Value.(*sm.SMRouter)
			addr := node.AddrWithRoots[root]
			for n := range node.Neighbours {
				if _, ok := routers[n].AddrWithRoots[root]; !ok {
					var linkKey string
					if node.ID < n {
						linkKey = sm.GetLinkKey(node.ID, n)
					} else {
						linkKey = sm.GetLinkKey(n, node.ID)
					}
					link := links[linkKey]
					if (link.Val1 > 0 && link.Val2 > 0) || !bi {
						routers[n].AddrWithRoots[root] = &sm.AddrType{
							Parent: node.ID,
							Addr:   addr.Addr + sm.GetRandomString(sm.ADDR_LENGTH_INTERVAL),
						}
						assigned[n] = struct{}{}
						queue.PushBack(routers[n])
					}
				}
			}
			if queue.Len() == 0 && bi {
				bi = false
				for n := range assigned {
					queue.PushBack(routers[n])
				}
			}
		}
	}
}

func clearTree(nodes map[sw.RouteID]*sw.SWRouter, roots []sw.RouteID) {
	for _, root := range roots {
		for _, node := range nodes {
			delete(node.AddrWithRoots,root)
		}
	}
}

func clearTreeSM(nodes map[sm.RouteID]*sm.SMRouter, roots []sm.RouteID)  {
	for _, root := range roots {
		for _, node := range nodes {
			delete(node.AddrWithRoots,root)
		}
	}
}
