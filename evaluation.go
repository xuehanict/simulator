package main

import (
	"bufio"
	"fmt"
	"github.com/lightningnetwork/simulator/flash"
	"github.com/lightningnetwork/simulator/landmark"
	"github.com/lightningnetwork/simulator/mara"
	"github.com/lightningnetwork/simulator/mpdv"
	"github.com/lightningnetwork/simulator/spider"
	"github.com/lightningnetwork/simulator/utils"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"sort"
)

func initLoger(str string) *logrus.Logger {
	file := "logs/" + str +".sum" //文件名
	summaryLogFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		fmt.Printf("open log file failed.\n")
	}

	file1 := "logs/" + str + ".log" //文件名
	logFile, err := os.OpenFile(file1, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		fmt.Printf("open log file failed.\n")
	}

	log := logrus.New()
	//log.SetFormatter()

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: summaryLogFile, // 为不同级别设置不同的输出目的
		logrus.InfoLevel:  summaryLogFile,
		logrus.WarnLevel:  summaryLogFile,
		logrus.ErrorLevel: summaryLogFile,
		logrus.FatalLevel: summaryLogFile,
		logrus.PanicLevel: summaryLogFile,
		logrus.TraceLevel: logFile,
	}, &logrus.JSONFormatter{})
	log.AddHook(lfHook)

	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.TraceLevel)
	return log
}

func MaraEval(m *mara.Mara, trans []utils.Tran, algo int, other string) {

	logName := fmt.Sprintf("MARA_%v_%v_%v_%v",
		m.MaxAddLength, m.AmountRate,m.NextHopBound, len(trans))
	log := initLoger(logName+other)
	backupChannelBase := utils.CopyChannels(m.Channels)

	total := 0.0
	totalVolumn := 0.0
	successVolumn := 0.0

	success := 0.0
	pathNumRecord := make([]int, 0)
	usedNumRecord := make([]int, 0)
	pathNumTotal := 0.0
	usedNumTotal := 0.0
	notFound := 0
	allcFailed := 0

	totalProbe := int64(0)
	totalOperation := int64(0)
	totalMaxLength := 0.0
	totalFees := utils.Amount(0)

	for _, tran := range trans {
		total++
		pathN, usedN, metric, err := m.SendPaymentWithBond(utils.RouterID(tran.Src),
			utils.RouterID(tran.Dest), algo, utils.Amount(tran.Val))
		totalProbe += metric.ProbeMessgeNum
		totalVolumn += tran.Val
		if err == nil {
			success++
			pathNumRecord = append(pathNumRecord, pathN)
			usedNumRecord = append(usedNumRecord, usedN)
			pathNumTotal += float64(pathN)
			usedNumTotal += float64(usedN)

			totalOperation += metric.OperationNum
			totalMaxLength += float64(metric.MaxPathLengh)
			totalFees += metric.Fees
			successVolumn += tran.Val

			log.WithFields(logrus.Fields{
				"result":           true,
				"success":          success,
				"total":            total,
				"from":             tran.Src,
				"to":               tran.Dest,
				"amt":              tran.Val,
				"pathN":            pathN,
				"totalVolumn": 		totalVolumn,
				"successVolumn":	successVolumn,
				"usedN":            usedN,
				"totalProbe":       totalProbe,
				"averageMaxLen":    totalMaxLength / success,
				"averageOperation": float64(totalOperation) / success,
				"averageFees":      float64(totalFees) / successVolumn,
			}).Trace("execute a payment.")
		} else {
			if payError, ok := err.(*mara.PaymentError); ok {
				switch payError.Code {
				case mara.FIND_PATH_FAILED:
					notFound++
				case mara.ALLOCARION_FAILED:
					allcFailed++
				}
			}

			log.WithFields(logrus.Fields{
				"result":           false,
				"error":            err.Error(),
				"success":          success,
				"total":            total,
				"from":             tran.Src,
				"to":               tran.Dest,
				"amt":              tran.Val,
				"pathN":            pathN,
				"usedN":            usedN,
				"totalProbe":       totalProbe,
				"averageMaxLen":    totalMaxLength / success,
				"averageOperation": float64(totalOperation) / success,
				"averageFees":      float64(totalFees) / successVolumn,
				"totalVolumn": 		totalVolumn,
				"successVolumn":	successVolumn,
			}).Trace("execute a payment.")
		}

		if total == 5000 {
			log.Trace("execute a round")
			break
		}
	}

	// 执行完一轮交易后，需要重新将备份的channel信息拷贝给m
	m.Channels = utils.CopyChannels(backupChannelBase)

	log.WithFields(logrus.Fields{
		"pathLengthBound": m.MaxAddLength,
		"amountLBrate":    m.AmountRate,
		"averageAllpath":  pathNumTotal / success,
		"averageUsedPath": usedNumTotal / success,
		"totalProbe":       totalProbe,
		"averageFees":      float64(totalFees) / successVolumn,
		"averageOperation": float64(totalOperation) / success,
		"averageMaxLen":    totalMaxLength / success,
		"totalVolumn": 		totalVolumn,
		"successVolumn":	successVolumn,
		"sussessRate":     success / total,
	}).Infof("a round test result shows")
}

func SpiderEval(s *spider.Spider, trans []utils.Tran, other string) {
	logName := fmt.Sprintf("SPIDER_%v", len(trans)) + other
	log := initLoger(logName)

	totalAmt := utils.Amount(0)
	successAmt := utils.Amount(0)
	successNum := 0

	totalProbe := int64(0)
	totalOperation := int64(0)
	totalMaxLength := 0.0
	totalFees := utils.Amount(0)
	totalNum := 0

	fileObj,err := os.OpenFile("sp-success-trans.txt",
		os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	if err!= nil {
		return
	}
	defer fileObj.Close()
	writeObj := bufio.NewWriter(fileObj)

	for _, tran := range trans {
		totalAmt += utils.Amount(tran.Val)
		totalNum++
		metric, err := s.SendPayment(utils.RouterID(tran.Src), utils.RouterID(tran.Dest),
			utils.Amount(tran.Val))
		totalProbe += metric.ProbeMessgeNum
		if err == nil {
			successNum++
			successAmt += utils.Amount(tran.Val)
			totalMaxLength += float64(metric.MaxPathLengh)
			totalFees += metric.Fees
			totalOperation += metric.OperationNum
			target := fmt.Sprintf("%v %v %v\n", tran.Src, tran.Dest, tran.Val)
			if _, err := writeObj.Write([]byte(target)); err == nil {
				if err := writeObj.Flush(); err != nil {panic(err)}
			}
		}
		log.WithFields(logrus.Fields{
			"success":          successNum,
			"total":            totalNum,
			"from":             tran.Src,
			"to":               tran.Dest,
			"amt":              tran.Val,
			"successVolume":    successAmt,
			"totalVolume":      totalAmt,
			"totalProbe":       totalProbe,
			"averageMaxLen":    totalMaxLength / float64(successNum),
			"averageOperation": float64(totalOperation) / float64(successNum),
			"averageFees":      float64(totalFees) / float64(successAmt),
		}).Trace("execute a payment.")
	}

	fileObj_,err := os.OpenFile("sp-channels.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	if err!= nil {
		return
	}
	defer fileObj_.Close()

	writeObj_ := bufio.NewWriter(fileObj_)
	for _, link := range s.Channels {
		target := fmt.Sprintf("%v %v %v %v\n",link.Part1, link.Part2, link.Val1, link.Val2)
		if _,err := writeObj_.Write([]byte(target));err == nil {
			if  err := writeObj_.Flush(); err != nil {panic(err)}
		}
	}
}

func FlashEval(f *flash.Flash, trans []utils.Tran,other string) {
	logName := fmt.Sprintf("FLASH_%v", len(trans)) + other
	log := initLoger(logName)

	tranAmts := make([]float64, 0)
	for _, tran := range trans {
		tranAmts = append(tranAmts, tran.Val)
	}
	sort.Float64s(tranAmts)
	thredhold := utils.Amount(tranAmts[int(0.9*float64(len(tranAmts)))])

	totalAmt := utils.Amount(0)
	successAmt := utils.Amount(0)
	successNum := 0.0
	totalNum := 0

	totalProbe := int64(0)
	totalOperation := int64(0)
	totalMaxLength := 0.0
	totalFees := utils.Amount(0)

	for _, tran := range trans {
		totalAmt += utils.Amount(tran.Val)
		totalNum++
		metric, err := f.SendPayment(utils.Amount(tran.Val), thredhold,
			utils.RouterID(tran.Src), utils.RouterID(tran.Dest))
		totalProbe += metric.ProbeMessgeNum
		if err == nil {
			successNum++
			successAmt += utils.Amount(tran.Val)
			totalMaxLength += float64(metric.MaxPathLengh)
			totalFees += metric.Fees
			totalOperation += metric.OperationNum
		}
		log.WithFields(logrus.Fields{
			"success":          successNum,
			"total":            totalNum,
			"from":             tran.Src,
			"to":               tran.Dest,
			"amt":              tran.Val,
			"successVolume":    successAmt,
			"totalVolume":      totalAmt,
			"totalProbe":       totalProbe,
			"averageMaxLen":    totalMaxLength / successNum,
			"averageOperation": float64(totalOperation) / successNum,
			"averageFees":      totalFees / successAmt,
		}).Trace("execute a payment.")
	}
}

func SWEval(s *landmark.SW, trans []utils.Tran, other string) {
	logName := fmt.Sprintf("SW_%v", len(trans)) + other
	log := initLoger(logName)

	totalAmt := utils.Amount(0)
	successAmt := utils.Amount(0)
	successNum := 0.0
	totalNum := 0

	totalProbe := int64(0)
	totalOperation := int64(0)
	totalMaxLength := 0.0
	totalFees := utils.Amount(0)

	for _, tran := range trans {
		totalAmt += utils.Amount(tran.Val)
		totalNum++
		metric, err := s.SendPayment(utils.RouterID(tran.Src),
			utils.RouterID(tran.Dest), utils.Amount(tran.Val))
		totalProbe += metric.ProbeMessgeNum
		if err == nil {
			successNum++
			successAmt += utils.Amount(tran.Val)
			totalMaxLength += float64(metric.MaxPathLengh)
			totalFees += metric.Fees
			totalOperation += metric.OperationNum
		} else {

		}
		log.WithFields(logrus.Fields{
			"success":          successNum,
			"total":            totalNum,
			"from":             tran.Src,
			"to":               tran.Dest,
			"amt":              tran.Val,
			"successVolume":    successAmt,
			"totalVolume":      totalAmt,
			"totalProbe":       totalProbe,
			"averageMaxLen":    totalMaxLength / successNum,
			"averageOperation": float64(totalOperation) / successNum,
			"averageFees":      totalFees / successAmt,
		}).Trace("execute a payment.")
	}
}

func SMEval(s *landmark.SM, trans []utils.Tran, other string) {
	logName := fmt.Sprintf("SM_%v", len(trans)) + other
	log := initLoger(logName)

	totalAmt := utils.Amount(0)
	successAmt := utils.Amount(0)
	successNum := 0.0
	totalNum := 0

	totalProbe := int64(0)
	totalOperation := int64(0)
	totalMaxLength := 0.0
	totalFees := utils.Amount(0)

	fileObj,err := os.OpenFile("sm-success-trans.txt",
		os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	if err!= nil {
		return
	}
	defer fileObj.Close()
	writeObj := bufio.NewWriter(fileObj)

	for _, tran := range trans {
		totalAmt += utils.Amount(tran.Val)
		totalNum++
		metric, err := s.SendPayment(utils.RouterID(tran.Src),
			utils.RouterID(tran.Dest), utils.Amount(tran.Val))
		totalProbe += metric.ProbeMessgeNum

		if err == nil {
			successNum++
			successAmt += utils.Amount(tran.Val)
			totalMaxLength += float64(metric.MaxPathLengh)
			totalFees += metric.Fees
			totalOperation += metric.OperationNum

			target := fmt.Sprintf("%v %v %v\n", tran.Src, tran.Dest, tran.Val)
			if _, err := writeObj.Write([]byte(target)); err == nil {
				if err := writeObj.Flush(); err != nil {panic(err)}
			}
		}
		log.WithFields(logrus.Fields{
			"success":          successNum,
			"total":            totalNum,
			"from":             tran.Src,
			"to":               tran.Dest,
			"amt":              tran.Val,
			"successVolume":    successAmt,
			"totalVolume":      totalAmt,
			"totalProbe":       totalProbe,
			"averageMaxLen":    totalMaxLength / successNum,
			"averageOperation": float64(totalOperation) / successNum,
			"averageFees":      totalFees / successAmt,
		}).Trace("execute a payment.")
	}

	fileObj_,err := os.OpenFile("sm-channels.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	if err!= nil {
		return
	}
	defer fileObj_.Close()

	writeObj_ := bufio.NewWriter(fileObj_)
	for _, link := range s.Channels {
		target := fmt.Sprintf("%v %v %v %v\n",link.Part1, link.Part2, link.Val1, link.Val2)
		if _,err := writeObj_.Write([]byte(target));err == nil {
			if  err := writeObj_.Flush(); err != nil {panic(err)}
		}
	}
}

func MpdvEval(m *mpdv.Mpdv, epoch int, trans []utils.Tran, other string)  {
	logName := fmt.Sprintf("MPDV_%v_", len(trans)) + other

	log := initLoger(logName)
	totalAmt := utils.Amount(0)
	successAmt := utils.Amount(0)
	successNum := 0.0
	totalNum := 0

	totalProbe := int64(0)
	totalOperation := int64(0)
	totalMaxLength := 0.0
	totalFees := utils.Amount(0)

	fmt.Printf("begin get dests\n")
	_, dests := utils.GetSdrAndRecr(trans[0:epoch])
	fmt.Printf("get dest done \n")
	m.ResetTable(dests)
	fmt.Printf("reset table done\n")
	for _, tran := range trans {
		totalAmt += utils.Amount(tran.Val)
		totalNum++
		metric, err := m.SendPayment(utils.Amount(tran.Val),
			utils.RouterID(tran.Src), utils.RouterID(tran.Dest))
		totalProbe += metric.ProbeMessgeNum

		if err == nil {
			successNum++
			successAmt += utils.Amount(tran.Val)
			totalMaxLength += float64(metric.MaxPathLengh)
			totalFees += metric.Fees
			totalOperation += metric.OperationNum
		}

		if totalNum % epoch == 0 {
			if len(trans) - totalNum < epoch {
				_, dests := utils.GetSdrAndRecr(trans[totalNum:])
				m.ResetTable(dests)
			} else {
				_, dests := utils.GetSdrAndRecr(trans[totalNum:totalNum+epoch])
				m.ResetTable(dests)
			}
		}
		log.WithFields(logrus.Fields{
			"success":          successNum,
			"total":            totalNum,
			"from":             tran.Src,
			"to":               tran.Dest,
			"amt":              tran.Val,
			"successVolume":    successAmt,
			"totalVolume":      totalAmt,
			"totalProbe":       totalProbe,
			"averageMaxLen":    totalMaxLength / successNum,
			"averageOperation": float64(totalOperation) / successNum,
			"averageFees":      totalFees / successAmt,
		}).Trace("execute a payment.")
	}
}



