package main

import (
	"fmt"
	"github.com/lightningnetwork/simulator/mara"
	"github.com/lightningnetwork/simulator/utils"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)


func initLoger() *logrus.Logger {
	file := time.Now().Format("20060102030505") + ".sum" //文件名
	summaryLogFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		fmt.Printf("open log file failed.\n")
	}

	file1 := time.Now().Format("20060102030505") + ".log" //文件名
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

func MaraEval(m *mara.Mara, trans []mara.Tran,
	amoutLB []float64, pathLength []int) {

	log := initLoger()
	backupChannelBase := utils.CopyChannels(m.Channels)

	for _, lb := range amoutLB {
		for _, maxL := range pathLength {

			total := 0
			success := 0
			pathNumRecord := make([]int, 0)
			usedNumRecord := make([]int, 0)
			pathNumTotal := 0
			usedNumTotal := 0
			for _, tran := range trans {
				total++
				pathN, usedN, err := m.SendPaymentWithBond(utils.RouterID(tran.Src),
					utils.RouterID(tran.Dest), utils.Amount(tran.Val), maxL, lb)
				if err == nil {
					success++
					pathNumRecord = append(pathNumRecord, pathN)
					usedNumRecord = append(usedNumRecord, usedN)
					pathNumTotal += pathN
					usedNumTotal += usedN

					log.WithFields(logrus.Fields{
						"result":  true,
						"success": success,
						"total":   total,
					}).Trace("execute a payment.")
				} else {
					log.WithFields(logrus.Fields{
						"result":  false,
						"error":   err.Error(),
						"success": success,
						"total":   total,
					}).Trace("execute a payment.")
				}

				if total == 5000 {
					log.Trace("execute a round" +
						"=======================================================" +
						"\n" +
						"\n" +
						"\n" +
						"=======================================================\n")
					break
				}
			}

			// 执行完一轮交易后，需要重新将备份的channel信息拷贝给m
			m.Channels = utils.CopyChannels(backupChannelBase)

			log.WithFields(logrus.Fields{
				"pathLengthBound": maxL,
				"amountLBrate":    lb,
				"averageAllpath":  pathNumTotal / success,
				"averageUsedPath": usedNumTotal / success,
			}).Infof("a round test result shows")
		}
	}
}
