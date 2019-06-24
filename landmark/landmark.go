package landmark

import (
	"container/list"
	"github.com/lightningnetwork/simulator/utils"
	"math/rand"
	"sort"
	"time"
)


const (
	AddrInterval = 8
)
type Addr struct {
	coordinate string
	parent utils.RouterID
}

type LandMarkRouting struct {
	*utils.Graph
	Coordination map[utils.RouterID]map[utils.RouterID]*Addr
	Roots []utils.RouterID
}

func randomPartition(amount utils.Amount, num int) []utils.Amount {
	res := make([]float64, num)
	rate := make([]float64, num)

	for i := 0; i < num-1; i++ {
		rate[i] = rand.Float64()
	}
	rate[num-1] = 1

	sort.Float64s(rate)
	res[0] = float64(amount) * rate[0]
	for i := 1; i < num; i++ {
		res[i] = float64(amount) * (rate[i] - rate[i-1])
	}

	amts := make([]utils.Amount,0)
	for _, a := range res  {
		amts = append(amts, utils.Amount(a))
	}
	return amts
}

func (l *LandMarkRouting)SetCoordinations() {
	roots := l.Roots
	for _, root := range roots {
		rootNode := l.Nodes[root]
		queue := list.New()
		queue.PushBack(rootNode)
		l.Coordination[root][root] = &Addr{
			coordinate: "",
			parent: -1,
		}

		bi := true
		assigned := make(map[utils.RouterID]struct{})
		assigned[root] = struct{}{}
		for {
			if queue.Len() == 0 {
				break
			}
			head := queue.Front()
			queue.Remove(head)
			node := head.Value.(*utils.Node)
			addr := l.Coordination[node.ID][root]
			for _, n := range node.Neighbours {
				if _, ok := l.Coordination[n][root]; !ok {
					if (utils.GetLinkValue(n, node.ID, l.Channels) > 0 &&
						utils.GetLinkValue(node.ID,n, l.Channels) > 0) || !bi {
						l.Coordination[n][root] = &Addr{
							parent: node.ID,
							coordinate: addr.coordinate + GetRandomString(AddrInterval),
						}
						assigned[n] = struct{}{}
						queue.PushBack(l.Nodes[n])
					}
				}
			}
			if queue.Len() == 0 && bi {
				bi = false
				for n := range assigned {
					queue.PushBack(l.Nodes[n])
				}
			}
		}
	}
}

// 生成一定长度的随机字符串
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
