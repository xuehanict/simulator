package main

import (
	"fmt"
	"github.com/lightningnetwork/simulator/flash"
	"github.com/lightningnetwork/simulator/landmark"
	"github.com/lightningnetwork/simulator/mara"
	"github.com/lightningnetwork/simulator/mpdv"
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
			Value: "mpdv",
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
			channels := utils.CopyChannels(m.Channels)
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-2.txt")
			//MaraEval(m, trans[0:tranNum], mara.MARA_MC, "tr-2")

			bounds := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
			amtRates := []float64{0.05, 0.1, 0.15,0.20, 0.25}
			for _, bound := range bounds {
				for _, amtRate := range amtRates {
					m.NextHopBound = bound
					m.Channels = utils.CopyChannels(channels)
					m.AmountRate = amtRate
					MaraEval(m, trans[0:tranNum], mara.MARA_MC, "tr-2")
				}
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

		case "mpdv":
			trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-2.txt")
			trans = trans[0:tranNum]
			g := utils.GetGraph("./data")
		 	err := g.LoadDistances("./data/finalSets/static/ripple_dis")
			if err != nil {
				fmt.Printf("load distance faced error")
			}
		 	m := mpdv.NewMpdv(g, 100, 0.1)
			MpdvEval(m, 1000, trans, "tr2")

		case "try":
			/*
			trans, err := utils.SampleTrans("./data/finalSets/static/", 1000)
			if err != nil {
				return err
			}
			*/


			//trans := utils.GenerateTrans("./data/finalSets/static/sampleTr-5.txt")
			g := utils.GetGraph("./data")
			g.StoreDistances("ripple_dis", 6)
			/*
			m := &mara.Mara{
				Graph:g,
				MaxAddLength: 2,
				AmountRate: 0.1,
				NextHopBound: 100,
			}
			testMany(m, trans[0:1000], []int{100,200, 300, 400, 500, 600, 700, 800, 900, 1000})
			*/
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func testMany(m *mara.Mara, trans []utils.Tran, testNum[]int)  {

	type record struct {
		paths []utils.Path
		amts []utils.Amount
		res bool
	}
	records := make(map[*utils.Tran]record)
	probed := 0
	/*
	tranNum := len(trans)

	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	calcuN := 0
	for i:= 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 100)
		go func( idx int) {
			wg.Add(1)
			tmpTrans := trans[idx*tranNum/10:(idx+1)*tranNum/10]
			for _, tran := range tmpTrans {
				paths, amts, err := m.TryPay(utils.RouterID(tran.Src),
					utils.RouterID(tran.Dest), mara.MARA_MC, utils.Amount(tran.Val))
				probed ++
				fmt.Printf("probe %v\n", probed)
				if err != nil {
					lock.Lock()
					records[&trans[i]] = record{
						paths: paths,
						amts: amts,
						res: false,
					}
					lock.Unlock()
				} else {
					lock.Lock()
					records[&trans[i]] = record{
						paths: paths,
						amts: amts,
						res: true,
					}
					lock.Unlock()
				}
				calcuN ++
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
*/

	for i, tran := range trans {
		paths, amts, err := m.TryPay(utils.RouterID(tran.Src),
			utils.RouterID(tran.Dest), mara.MARA_MC, utils.Amount(tran.Val))
		probed ++
		fmt.Printf("probe %v\n", probed)
		if err != nil {
			records[&trans[i]] = record{
				paths: paths,
				amts: amts,
				res: false,
			}
		} else {
			records[&trans[i]] = record{
				paths: paths,
				amts: amts,
				res: true,
			}
		}
	}
	channelsBackup := utils.CopyChannels(m.Channels)
	// 以上探路完
	channels := utils.CopyChannels(m.Channels)
	for _, num := range testNum {
		conflict := 0
		for i := range trans[0:num] {
			if re := records[&trans[i]]; re.res == true {
				channels = utils.CopyChannels(m.Channels)
				err := m.UpdateWeights(re.paths, re.amts)
				if err != nil {
					conflict++
					//回滚
					m.Channels = channels
				}
			}
		}
		channels = utils.CopyChannels(channelsBackup)
		m.Channels = utils.CopyChannels(channels)
		fmt.Printf("total is %v, and conflict is %v\n", num, conflict)
	}
}

