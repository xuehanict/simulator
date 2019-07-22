package spider

import "github.com/lightningnetwork/simulator/utils"

func (s *Spider)SendPayments (trans []utils.Tran) {
	allPaths := []utils.Path{}
	transIndexInPaths := make([]int, len(trans))
	for i, tran := range trans {
		transIndexInPaths[i] = len(allPaths)
		paths := s.getPaths(utils.RouterID(tran.Src), utils.RouterID(tran.Dest), 4)
		allPaths = append(allPaths, paths...)
	}







}


