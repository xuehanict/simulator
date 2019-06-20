package landmark

import "github.com/lightningnetwork/simulator/utils"

type SM struct {
	LandMarkRouting
}


func (s *SM) getPaths (src, dest utils.RouterID) []utils.Path {
	return nil
}

func (s *SM) SendPayment (src, dest utils.RouterID, amt utils.Amount) (
	bool, error) {
	splittedAmounts := randomPartition(amt, len(s.Roots))

	return false, nil
}

