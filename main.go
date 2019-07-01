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
			g := utils.GetGraph("./data")
			m := &mara.Mara{
				Graph:g,
				MaxAddLength: 2,
				AmountRate: 0.1,
				NextHopBound: 100,
			}
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-2.txt")
			//MaraEval(m, trans[0:tranNum], mara.MARA_MC, "tr-2")

			bounds := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
			for _, bound := range bounds {
				m.NextHopBound = bound
				MaraEval(m, trans[0:tranNum], mara.MARA_MC, "tr-2")
			}

		case "sm":
			g := utils.GetGraph("./data")
			s := landmark.NewSM(g, []utils.RouterID{5, 38, 13})
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-2.txt")
			SMEval(s, trans[0:tranNum], "tr-2")

		case "sw":
			g := utils.GetGraph("./data")
			s := landmark.NewSw(g, []utils.RouterID{5, 38, 13})
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-2.txt")
			SWEval(s, trans[0:tranNum], "tr-2")
		case "dijk":
			g := utils.GetGraph("./data")
			m := &mara.Mara{
				Graph:g,
				MaxAddLength: 6,
				AmountRate: 0.05,
				NextHopBound: 100,
			}
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-2.txt")
			MaraEval(m, trans[0:tranNum], mara.MARA_SPT, "tr-2")

		case "sp":
			g := utils.GetGraph("./data")
			s := spider.NewSpider(g, spider.WATERFIILING, 4)
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-2.txt")
			SpiderEval(s, trans[0:tranNum], "tr-2")

		case "flash":
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-2.txt")
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
			FlashEval(f, trans, "tr-2")
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func testMany(m *mara.Mara)  {

}

