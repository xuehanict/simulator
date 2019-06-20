package landmark

import (
	"github.com/lightningnetwork/simulator/utils"
	"math/rand"
	"sort"
)


const (
	AddrInterval = 8
)
type addr struct {
	coordinate string
	parent utils.RouterID
}

type LandMarkRouting struct {
	utils.Graph
	Coordination map[utils.RouterID]map[utils.RouterID]addr
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



