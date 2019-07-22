package utils

import (
	"container/list"
)

func BfsPath (nodes map[RouterID]*Node, src, dest RouterID, checkCap bool,
	linkBase map[string]*Link) Path{
	queue := list.New()
	queue.PushBack(nodes[src])
	visited := make([]bool, len(nodes))
	prev := make([]RouterID, len(nodes))
	distance := make([]int, len(nodes))

	for i :=  range visited {
		visited[i] = false
	}
	visited[src] = true
	distance[src] = 0

	for {
		//fmt.Print("one ring")
		if queue.Len() != 0 {
			currNode := queue.Front().Value.(*Node)
			if currNode.ID == dest {
				break
			}
			for nei := range currNode.Neighbours {
				if visited[nei] == false &&
					(checkCap == false || GetLinkValue(currNode.ID,nei, linkBase)>0){
					queue.PushBack(nodes[nei])
					visited[nei] = true
		//			fmt.Printf("push %v\n", nei)
					prev[nei] = currNode.ID
					distance[nei] = distance[currNode.ID] + 1
				}
			}
			queue.Remove(queue.Front())
		//	fmt.Printf("remove %v\n", currNode.ID)
		}else {
			return nil
		}
	}

	path := make([]RouterID, distance[dest]+1)
	path[0] = src
	path[distance[dest]] = dest
	cursor := dest
	for i := distance[dest]-1; i > 0; i-- {
		path[i] = prev[cursor]
		cursor = prev[cursor]
	}
	return path
}


