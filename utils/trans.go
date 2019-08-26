package utils

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Tran struct {
	Src  int
	Dest int
	Val  float64
}

// 从单个文件中读取出全部交易
func GenerateTrans(filePath string) []Tran {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("os Open error: ", err)
		return nil
	}
	defer f.Close()

	br := bufio.NewReader(f)
	trans := make([]Tran, 0)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("br ReadLine error: ", err)
			return nil
		}
		splitStr := strings.Split(string(line), " ")
		val, _ := strconv.ParseFloat(splitStr[0], 64)
		src, _ := strconv.Atoi(splitStr[1])
		dest, _ := strconv.Atoi(splitStr[2])

		trans = append(trans, Tran{
			Src:  src,
			Dest: dest,
			Val:  val,
		})
	}

	return trans
}

// 一次性从目录中读取全部文件的所有交易出来
func GenerateTransFromPath (path string) ([]Tran, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	
	res := make([]Tran,0)
	for _, file := range files {
		if strings.Contains(file.Name(), "sample") {
			trans := GenerateTrans(path+file.Name())
			res = append(res, trans...)	
		}
	}
	//fmt.Printf("trans length is %v\n", len(res))
	return res, nil
}

// 从所有交易中进行采样，num是采样的个数
func SampleTrans(dirPath string, num int) ([]Tran, error) {
	trans, err := GenerateTransFromPath(dirPath)
	if err != nil {
		return nil, err
	}
	if len(trans) < num {
		return nil, fmt.Errorf("number of the total trans less than needed")
	}

	sampled := make([]Tran, 0)
	toolMap := make(map[int]Tran)
	rand.Seed(time.Now().UnixNano())
	for ; len(toolMap) < num; {
		idx := rand.Intn(len(trans))
		toolMap[idx] = trans[idx]
	}

	for _, v := range toolMap {
		sampled = append(sampled, v)
	}
	return sampled, nil
}

func GetSdrAndRecr(trans []Tran) (map[RouterID]struct{},
	map[RouterID]struct{}) {
	s := make(map[RouterID]struct{})
	r := make(map[RouterID]struct{})
	for _, tran := range trans {
		s[RouterID(tran.Src)] = struct{}{}
		r[RouterID(tran.Dest)] = struct{}{}
	}
	return s, r
}

