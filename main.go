package main

import (
	"fmt"
	"github.com/lightningnetwork/simulator/flash"
	"github.com/lightningnetwork/simulator/landmark"
	"github.com/lightningnetwork/simulator/mara"
	"github.com/lightningnetwork/simulator/spider"
	"github.com/lightningnetwork/simulator/utils"

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
			amountLB := []float64{0.05}
			pathAddLenth := []float64{6}

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
					g := utils.GetGraph("./data")
					m := &mara.Mara{
						Graph:g,
					}
					trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-2.txt")
					MaraEval(m, trans[0:tranNum], mara.MARA_MC,amountLB[idx:idx+1], pathAddLenth)
					wg.Done()
				}(i)
				i++
			}
			wg.Wait()

		case "sm":
			g := utils.GetGraph("./data")
			s := landmark.NewSM(g, []utils.RouterID{5, 38, 13})
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-2.txt")
			SMEval(s, trans[0:tranNum])

		case "sw":
			g := utils.GetGraph("./data")
			s := landmark.NewSw(g, []utils.RouterID{5, 38, 13})
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-1.txt")
			SWEval(s, trans[0:tranNum])
		case "dijk":


		case "sp":
			g := utils.GetGraph("./data")
			s := spider.NewSpider(g, spider.WATERFIILING, 4)
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-1.txt")
			SpiderEval(s, trans[0:tranNum])

		case "flash":
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-1.txt")
			trans = trans[0:tranNum]
			g := utils.GetGraph("./data")
			f := flash.NewFlash(g, 20, true)

			wg := sync.WaitGroup{}
			calcuN := 0
			for i:= 0; i < 10; i++ {
				time.Sleep(time.Millisecond * 100)
				go func( idx int) {
					wg.Add(1)
					tmpTrans := trans[idx*tranNum/10:(idx+1)*tranNum/10]
					for _, tran := range tmpTrans {
						f.AddShortestPathsTest(utils.RouterID(tran.Src), utils.RouterID(tran.Dest))
						fmt.Printf("算完一个交易路径:%v \n", calcuN)
						calcuN ++
					}
					wg.Done()
				}(i)
			}
			wg.Wait()
			fmt.Printf("算完所有路径\n")
			FlashEval(f, trans)
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
