package main

import (
	"github.com/lightningnetwork/simulator/mara"
	"github.com/urfave/cli"
)

func main() {
	/*
		testCase := 2

		switch testCase {
		case 1:
			testSW()
		case 2:
			testSWBigData()
		case 3:
			testSWTree()
		case 4:
			testSMTree()
		case 5:
			testSMBigData()
		}
	*/

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "algo,a",
			Value: "mara",
			Usage: "algorithm to run or test",
		},
		cli.IntFlag{
			Name:  "trans_num",
			Value: 10000,
			Usage: "number of transactions to execute",
		},
	}

	app.Action = func(c *cli.Context) error {
		algo := c.String("algo,a")
		transNum := c.Int("trans_num")

		switch algo {
		case "mara":
			m, trans := mara.GetRippleMaraAndTrans("./data")
			amountLB := []float64{0, 0.01, 0.03, 0.05, 0.07, 0.1}
			pathLenth := []int{6, 8, 10, 12, 14}
			MaraEval(m, trans, amountLB, pathLenth)
		case "sm":

		case "sw":
		case "dijk":
		}

		return nil
	}

}
