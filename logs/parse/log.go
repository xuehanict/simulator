package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type logStruct struct {
	Success          int     `json:"success"`
	Total            int     `json:"total"`
	From             uint64  `json:"from"`
	To               uint64  `json:"to"`
	Amt              float64 `json:"amt"`
	SuccessVolume    float64 `json:"successVolume"`
	TotalProbe       uint64  `json:"totalProbe"`
	TotalVolume      float64 `json:"totalVolume"`
	AverageMaxLen    float64 `json:"averageMaxLen"`
	AverageOperation float64 `json:"averageOperation"`
	AverageFees      float64 `json:"averageFees"`
	Msg              string	 `json:"msg"`
}

type logSummary struct {
	SuccessRatio map[int]float64 `json:"success_ratio"`
	SuccessVolume map[int]float64 `json:"success_volume"`
	TotalVolume	 map[int]float64 `json:"total_volume"`
	Name string
}

func getSummary (fileName string) (*logSummary, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sucRatio := make(map[int]float64)
	sucVolume := make(map[int]float64)
	totalVolume := make(map[int]float64)

	br := bufio.NewReader(f)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("br ReadLine error: ", err)
			break
		}
		var oneLog logStruct
		err = json.Unmarshal(line, &oneLog)
		if err != nil {
			continue
		}

		if oneLog.Total % 1000 == 0 {
			sucRatio[oneLog.Total] = float64(oneLog.Success) / float64(oneLog.Total)
			sucVolume[oneLog.Total] = oneLog.SuccessVolume
			totalVolume[oneLog.Total] = oneLog.TotalVolume
		}
	}

	return &logSummary{
		SuccessRatio:sucRatio,
		SuccessVolume: sucVolume,
		TotalVolume: totalVolume,
	}, nil
}

func main() {
	filePath := "logs/"
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		fmt.Printf("faced error :%v\n", err)
	}
	allLogSum := make(map[string]*logSummary)

	for _, file := range files {
		if strings.Contains(file.Name(), ".log") {
			// 处理每个算法
			if strings.Contains(file.Name(), "MARA") {
				ls, err := getSummary(filePath + file.Name())
				if err != nil {
					fmt.Printf("faced error :%v\n", err)
				}
				allLogSum["mara"] = ls
			} else if strings.Contains(file.Name(), "SPIDER") {
				ls, err := getSummary(filePath + file.Name())
				if err != nil {
					fmt.Printf("faced error :%v\n", err)
				}
				allLogSum["spider"] = ls
			} else if strings.Contains(file.Name(), "SM") {
				ls, err := getSummary(filePath + file.Name())
				if err != nil {
					fmt.Printf("faced error :%v\n", err)
				}
				allLogSum["sm"] = ls
			} else if strings.Contains(file.Name(), "FLASH") {
				ls, err := getSummary(filePath + file.Name())
				if err != nil {
					fmt.Printf("faced error :%v\n", err)
				}
				allLogSum["flash"] = ls
			}
		} else { continue }
	}


	tranNumArray := []int{1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000}
	// 先输出成功率
	fileObj,err := os.OpenFile("res/log-res.txt",os.O_RDWR|os.O_CREATE|os.O_TRUNC,0644)
	if err!= nil {
		fmt.Printf("faced error :%v\n", err)
		return
	}
	defer fileObj.Close()

	writeObj := bufio.NewWriter(fileObj)
	str := fmt.Sprintf("succecc ratio \n")
	if _,err := writeObj.Write([]byte(str));err == nil {
		if  err := writeObj.Flush(); err != nil {panic(err)}
	}

	for _, tranNum := range tranNumArray{
		str := fmt.Sprintf("%10d%15.5v%15.5v%15.5v%15.5v\n",
			tranNum,
			allLogSum["mara"].SuccessRatio[tranNum],
			allLogSum["spider"].SuccessRatio[tranNum],
			allLogSum["sm"].SuccessRatio[tranNum],
			allLogSum["flash"].SuccessRatio[tranNum],
		)

		if _,err := writeObj.Write([]byte(str));err == nil {
			if  err := writeObj.Flush(); err != nil {panic(err)}
		}
	}

	str = fmt.Sprintf("succecc volume %15s%15s%15s%15s%15s \n",
		"mara", "spider", "sm", "flash", "total")
	if _,err := writeObj.Write([]byte(str));err == nil {
		if  err := writeObj.Flush(); err != nil {panic(err)}
	}

	for _, tranNum := range tranNumArray{
		str := fmt.Sprintf("%10d%15.7v%15.7v%15.7v%15.7v%15.7v\n",
			tranNum,
			allLogSum["mara"].SuccessVolume[tranNum],
			allLogSum["spider"].SuccessVolume[tranNum],
			allLogSum["sm"].SuccessVolume[tranNum],
			allLogSum["flash"].SuccessVolume[tranNum],
			allLogSum["sm"].TotalVolume[tranNum],
		)

		if _,err := writeObj.Write([]byte(str));err == nil {
			if  err := writeObj.Flush(); err != nil {panic(err)}
		}
	}
	spew.Dump(allLogSum)
}


