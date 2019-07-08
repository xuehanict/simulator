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

func generateTransFromPath (path string) ([]Tran, error) {
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
	fmt.Printf("trans length is %v\n", len(res))
	return res, nil
}

func SampleTrans(dirPath string, num int) ([]Tran, error) {
	trans, err := generateTransFromPath(dirPath)
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

func GetSdrAndRecr(trans []Tran) ([]RouterID, []RouterID) {
	s := make(map[RouterID]struct{})
	r := make(map[RouterID]struct{})
	senders := make([]RouterID,0)
	recivers := make([]RouterID, 0)
	for _, tran := range trans {
		if _, ok := s[RouterID(tran.Src)]; !ok {
			senders = append(senders, RouterID(tran.Src))
			s[RouterID(tran.Src)] = struct{}{}
		}
		if _, ok := r[RouterID(tran.Dest)]; !ok {
			recivers = append(recivers, RouterID(tran.Src))
			r[RouterID(tran.Src)] = struct{}{}
		}
	}
	return senders, recivers
}

