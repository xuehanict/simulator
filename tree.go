package main

import (
	sw "github.com/lightningnetwork/sm/silentWhisper"
	"time"
	"container/list"

)



func createTree (routers map[sw.RouteID]*sw.SWRouter, links map[string]*sw.Link, root sw.RouteID)  {

	rootNode := routers[root]
	timeNow := time.Now().Unix()
	queue := list.New()
	queue.PushBack(rootNode)
	rootNode.AddrWithRoots[root] = &sw.AddrType{
		Addr: "",
		Parent: -1,
		Time: timeNow,
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
						Addr: addr.Addr + sw.GetRandomString(sw.ADDR_INTERVAL),
						Time: timeNow,
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

