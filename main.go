package main

import (
	"github.com/lightningnetwork/simulator/mara"

	//"github.com/lightningnetwork/simulator/mara"
	"github.com/urfave/cli"
	"log"
	"os"
	"sync"
	"time"
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
			Name:  "algo",
			Value: "mara",
			Usage: "algorithm to run or test",
		},
		cli.IntFlag{
			Name:  "trans_num",
			Value: 5000,
			Usage: "number of transactions to execute",
		},
	}

	app.Action = func(c *cli.Context) error {
		algo := c.String("algo")
		tranNum := c.Int("trans_num")
		switch algo {
		case "mara":
			amountLB := []float64{0.01, 0.03, 0.05, 0.07, 0.1}
			pathLenth := []int{6, 8, 10}

			wg := sync.WaitGroup{}
			i := 0
			for {
				time.Sleep(time.Second)
				//fmt.Printf("=============%v", i)
				if i == len(amountLB) {
					break
				}
				go func(idx int) {
					wg.Add(1)
					m, trans := mara.GetRippleMaraAndTrans("./data")
					MaraEval(m, trans[0:tranNum], amountLB[idx:idx+1], pathLenth)
					wg.Done()
				}(i)
				i++
			}
			wg.Wait()

		case "sm":

		case "sw":
		case "dijk":
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
