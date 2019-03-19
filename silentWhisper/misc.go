package silentWhisper

import (
	"math/rand"
	"time"
	"bytes"
	"strconv"
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
func getLinkKey(r1, r2 RouteID) string {
	return strconv.Itoa(int(r1)) + "-" + strconv.Itoa(int(r2))
}
