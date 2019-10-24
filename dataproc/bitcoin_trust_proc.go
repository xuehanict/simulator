package dataproc

import (
	"bufio"
	"fmt"
	"github.com/lightningnetwork/simulator/utils"
	"io"
	"os"
)

func ReadBitcoinTrustFile(fileName string) ([]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println("os Open error: ", err)
		return nil, err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	// 首先跳过第一行，因为第一行是标头
	_, _, err = br.ReadLine()
	if err != nil {
		return nil, err
	}

	strs := make([]string, 0)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		strs = append(strs, string(line))
	}

	return strs, nil
}

func GetBitCoinTrustGraph(filename string) (*utils.Graph, error) {
	strs, err := ReadBitcoinTrustFile(filename)
	if err != nil {
		return nil, err
	}
	return getGraph(strs)
}



