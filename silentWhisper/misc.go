package silentWhisper

import (
	"bytes"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

// 生成一定长度的随机字符串
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// 用到的计算距离的方式
func getDis(neighbour, dest string, lenthInterval int) int {
	depthN := len(neighbour) / lenthInterval
	depthD := len(dest) / lenthInterval
	cpl := 0
	neiBytes := []byte(neighbour)
	destBytes := []byte(dest)

	for ; cpl < depthD && cpl < depthN &&
		bytes.Equal(neiBytes[0:lenthInterval],
			destBytes[0:lenthInterval]); cpl++ {
		neiBytes = neiBytes[lenthInterval:]
		destBytes = destBytes[lenthInterval:]
	}
	return depthN + depthD - (2 * cpl)
}

// 生成link的key，要求r1 < r2
func GetLinkKey(r1, r2 RouteID) string {
	return strconv.Itoa(int(r1)) + "-" + strconv.Itoa(int(r2))
}

func randomPartition(amount float64, num int) []float64 {
	res := make([]float64, num)
	rate := make([]float64, num)

	for i := 0; i < num-1; i++ {
		rate[i] = rand.Float64()
	}
	rate[num-1] = 1

	sort.Float64s(rate)
	res[0] = amount * rate[0]
	for i := 1; i < num; i++ {
		res[i] = amount * (rate[i] - rate[i-1])
	}
	return res
}

func minPart(amount float64, mins []float64) []float64 {
	remainder := amount
	saturated := make(map[int]struct{})
	res := randomPartition(amount, len(mins))

	if len(res) == 0 {
		return nil
	}
	for remainder > 0 {
		remainder = 0
		for i := 0; i < len(res); i++ {
			if res[i] > mins[i] {
				remainder = remainder + res[i] - mins[i]
				res[i] = mins[i]
				saturated[i] = struct{}{}
			}
		}

		if len(saturated) == len(mins) {
			return nil
		} else {
			if remainder > 0 {
				adds := randomPartition(remainder, len(mins)-len(saturated))
				k := 0
				for i := 0; i < len(adds); i++ {
					for {
						if _, ok := saturated[k]; ok {
							k++
						} else {
							break
						}
					}
					res[k] = res[k] + adds[i]
					k++
				}
			}
		}
	}
	return res
}
