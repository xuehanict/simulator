package main

import (
	"fmt"
	"github.com/lightningnetwork/simulator/dataproc"
	"github.com/lightningnetwork/simulator/flash"
	"github.com/lightningnetwork/simulator/landmark"
	"github.com/lightningnetwork/simulator/mara"
	"github.com/lightningnetwork/simulator/spider"
	"github.com/lightningnetwork/simulator/utils"
	"sync"
	"time"
)

func rippleDataTest(tranNum int) {
	fmt.Printf("start test\n")
	g := utils.GetGraph("./data")
	utils.RanddomFeeRate(g.Channels)
	oriTrans, _ := utils.GenerateTransFromPath("data/finalSets/static/")
	fmt.Printf("origin trans length is %v", len(oriTrans))
	for {
		if dataproc.CutOneDegree(4, g) == 0 {
			break
		}
	}
	idMap := dataproc.ConvertToSeriesID(false, g)
	trans := dataproc.RandomTrans(oriTrans, idMap, 10000)
	backChannels := utils.CopyChannels(g.Channels)
	fmt.Printf("transaction length is %v", len(trans))
	//time.Sleep(time.Second * 100)

	// sm测试
	fmt.Printf("sm start teset\n")
	sm := landmark.NewSM(g, []utils.RouterID{5, 38, 13})
	SMEval(sm, trans, "random-r")

	// spider测试
	fmt.Printf("spider start teset\n")
	g.Channels = utils.CopyChannels(backChannels)
	sp := spider.NewSpider(g, spider.WATERFIILING, 4)
	SpiderEval(sp, trans, "random-r")

	// mara测试
	fmt.Printf("mara start teset\n")
	g.Channels = utils.CopyChannels(backChannels)
	m := &mara.Mara{
		Graph:        g,
		MaxAddLength: 4,
		AmountRate:   0.1,
		NextHopBound: 20,
	}
	MaraEval(m, trans, mara.MARA_MC, "random-r")

	// flash测试
	fmt.Printf("flash start teset\n")
	g.Channels = utils.CopyChannels(backChannels)
	f := flash.NewFlash(g, 20, true)
	wg := sync.WaitGroup{}
	calcuN := 0
	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 100)
		go func(idx int) {
			wg.Add(1)
			tmpTrans := trans[idx*tranNum/10 : (idx+1)*tranNum/10]
			for _, tran := range tmpTrans {
				f.AddShortestPathsTest(utils.RouterID(tran.Src), utils.RouterID(tran.Dest))
				fmt.Printf("算完一个交易路径:%v \n", calcuN)
				calcuN++
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Printf("算完所有路径\n")
	FlashEval(f, trans, "random-r")

}

func lightningDataTest(tranNum int) {
	fmt.Printf("start test\n")
	g, _ := dataproc.ParseLightningGraph("./data/lightning/testnetgraph.json")
	dataproc.CutOneDegree(2, g)
	dataproc.CutOneDegree(2, g)
	dataproc.CutOneDegree(2, g)
	dataproc.ConvertToSeriesID(true, g)
	trans, _ := dataproc.GetLightningTrans(len(g.Nodes), 10000,
		"data/ripple/ripple_val.csv", "data/lightning/BitcoinVal.txt")
	backChannels := utils.CopyChannels(g.Channels)

	// sm测试
	fmt.Printf("sm start teset\n")
	//g.CutOneDegree(2)

	sm := landmark.NewSM(g, []utils.RouterID{5, 38, 13})
	SMEval(sm, trans, "random-l")

	// spider测试
	fmt.Printf("spider start teset\n")
	g.Channels = utils.CopyChannels(backChannels)
	sp := spider.NewSpider(g, spider.WATERFIILING, 4)
	SpiderEval(sp, trans, "random-l")

	// mara测试
	fmt.Printf("mara start teset\n")
	g.Channels = utils.CopyChannels(backChannels)
	m := &mara.Mara{
		Graph:        g,
		MaxAddLength: 4,
		AmountRate:   0.1,
		NextHopBound: 50,
	}
	MaraEval(m, trans, mara.MARA_MC, "random-l")

	// flash测试
	fmt.Printf("flash start teset\n")
	g.Channels = utils.CopyChannels(backChannels)
	f := flash.NewFlash(g, 20, true)
	wg := sync.WaitGroup{}
	calcuN := 0
	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 100)
		go func(idx int) {
			wg.Add(1)
			tmpTrans := trans[idx*tranNum/10 : (idx+1)*tranNum/10]
			for _, tran := range tmpTrans {
				f.AddShortestPathsTest(utils.RouterID(tran.Src), utils.RouterID(tran.Dest))
				fmt.Printf("算完一个交易路径:%v \n", calcuN)
				calcuN++
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Printf("算完所有路径\n")
	FlashEval(f, trans, "random-l")
}
