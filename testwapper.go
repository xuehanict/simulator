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

var (
	testNameMap = map[int]string {
		dataproc.REMAINDER_SAMPLE: "remainder",
		dataproc.MAP_SAMPLE: "map",
		dataproc.REBALANCE_CHANEL: "balance",
		dataproc.UNIFORMLY_CHANNEL: "uniform",
		dataproc.ORIGION_CHANNEL: "origin",
	}
)

func rippleDataTest(tranNum int) {
	channelAssignWay := dataproc.REBALANCE_CHANEL
	transSampleWay := dataproc.REMAINDER_SAMPLE
	logName := fmt.Sprintf("rippleData-%s-%s",
		testNameMap[channelAssignWay], testNameMap[transSampleWay])

	fmt.Printf("start test\n")
	g := utils.GetGraph("./data")
	utils.RanddomFeeRate(g.Channels)
	oriTrans, _ := utils.GenerateTransFromPath("data/finalSets/static/")
	fmt.Printf("origin trans length is %v\n", len(oriTrans))
	for {
		if dataproc.CutOneDegree(4, g) == 0 {
			break
		}
	}
	idMap := dataproc.ConvertToSeriesID(channelAssignWay, g)
	trans := dataproc.RandomRippleTrans(oriTrans, idMap, tranNum, transSampleWay, false)
	backChannels := utils.CopyChannels(g.Channels)
	fmt.Printf("transaction length is %v", len(trans))
	//time.Sleep(time.Second * 100)

	// sm测试
	fmt.Printf("sm start teset\n")
	sm := landmark.NewSM(g, []utils.RouterID{5, 38, 13})
	SMEval(sm, trans, logName)

	// spider测试
	fmt.Printf("spider start teset\n")
	g.Channels = utils.CopyChannels(backChannels)
	sp := spider.NewSpider(g, spider.WATERFIILING, 4)
	SpiderEval(sp, trans, logName)

	// mara测试
	fmt.Printf("mara start teset\n")
	g.Channels = utils.CopyChannels(backChannels)
	m := &mara.Mara{
		Graph:        g,
		MaxAddLength: 2,
		AmountRate:   0.1,
		NextHopBound: 10,
	}
	MaraEval(m, trans, mara.MARA_MC, logName)

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
	FlashEval(f, trans, logName)
}

func rippleSnapShotDataTest(tranNum int) {
	channelAssignWay := dataproc.REBALANCE_CHANEL
	transSampleWay := dataproc.REMAINDER_SAMPLE
	logName := fmt.Sprintf("rippleSnapshot-%s-%s",
		testNameMap[channelAssignWay], testNameMap[transSampleWay])

	fmt.Printf("start test\n")
	g := utils.GetGraphSnapshot("./data", true)
	utils.RanddomFeeRate(g.Channels)
	oriTrans, _ := utils.GenerateTransFromPath("data/finalSets/static/")
	fmt.Printf("origin trans length is %v\n", len(oriTrans))

	// 对图进行预处理，先删除度为1的结点， 再删除不连通的小部分, 最后再将序号再从0开始编排
	for {
		if dataproc.CutOneDegree(2, g) == 0 {
			break
		}
	}

	//toRemove := dataproc.GetNotConnectedNodes(g)
	//dataproc.RemoveNotConnectNodes(g, toRemove)

	idMap := dataproc.ConvertToSeriesID(channelAssignWay, g)
	trans := dataproc.RandomRippleTrans(oriTrans, idMap, tranNum, transSampleWay, false)
	backChannels := utils.CopyChannels(g.Channels)
	fmt.Printf("transaction length is %v", len(trans))
	//time.Sleep(time.Second * 100)

	// sm测试
	sm := landmark.NewSM(g, []utils.RouterID{5, 38, 13})
	SMEval(sm, trans, logName)

	// spider测试
	g.Channels = utils.CopyChannels(backChannels)
	sp := spider.NewSpider(g, spider.WATERFIILING, 4)
	SpiderEval(sp, trans, logName)

	// mara测试
	g.Channels = utils.CopyChannels(backChannels)
	m := &mara.Mara{
		Graph:        g,
		MaxAddLength: 2,
		AmountRate:   0.1,
		NextHopBound: 20,
	}
	MaraEval(m, trans, mara.MARA_MC,logName)

	// flash测试
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
	FlashEval(f, trans, logName)
}

func lightningDataTest(tranNum int) {
	fmt.Printf("start test\n")
	g, _ := dataproc.ParseLightningGraph("./data/lightning/mainnetgraph.json")
	for {
		if dataproc.CutOneDegree(2, g) == 0 {
			break
		}
	}
	dataproc.ConvertToSeriesID(dataproc.ORIGION_CHANNEL, g)
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

func debugRippleSnapShotDataTest(tranNum int) {
	channelAssignWay := dataproc.REBALANCE_CHANEL
	transSampleWay := dataproc.REMAINDER_SAMPLE
	logName := fmt.Sprintf("debugRippleSnapshot-%s-%s",
		testNameMap[channelAssignWay], testNameMap[transSampleWay])

	fmt.Printf("start test\n")
	g := utils.GetGraphSnapshot("./data", true)
	utils.RanddomFeeRate(g.Channels)
	// 对图进行预处理，先删除度为1的结点， 再删除不连通的小部分, 最后再将序号再从0开始编排
	for {
		if dataproc.CutOneDegree(2, g) == 0 {
			break
		}
	}

	//toRemove := dataproc.GetNotConnectedNodes(g)
	//dataproc.RemoveNotConnectNodes(g, toRemove)
	idMap := utils.GetMap("data/ripple/node_map.txt")
	dataproc.ConvertToSeriesIDWithMap(channelAssignWay, g, idMap)
	trans := GenerateTrans("data/ripple/trans0.txt")
	backChannels := utils.CopyChannels(g.Channels)
	fmt.Printf("transaction length is %v", len(trans))
	//time.Sleep(time.Second * 100)

	// sm测试
	sm := landmark.NewSM(g, []utils.RouterID{5, 38, 13})
	SMEval(sm, trans, logName)

	// spider测试
	g.Channels = utils.CopyChannels(backChannels)
	sp := spider.NewSpider(g, spider.WATERFIILING, 4)
	SpiderEval(sp, trans, logName)

	// mara测试
	g.Channels = utils.CopyChannels(backChannels)
	m := &mara.Mara{
		Graph:        g,
		MaxAddLength: 2,
		AmountRate:   0.1,
		NextHopBound: 20,
	}
	MaraEval(m, trans, mara.MARA_MC,logName)

	// flash测试
	g.Channels = utils.CopyChannels(backChannels)
	f := flash.NewFlash(g, 20, true)
	wg := sync.WaitGroup{}
	calcuN := 0
	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 100)
		go func(idx int) {
			wg.Add(1)
			fmt.Printf("idx is %v\n", idx)
			tmpTrans := trans[idx*len(trans)/10 : (idx+1)*len(trans)/10]
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
	FlashEval(f, trans, logName)
}

