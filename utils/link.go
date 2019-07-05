package utils

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"
)

/*
 * Val1 指Part1往Part2方向的通道容量
 * Val2 指Part2往Part1方向的通道容量
 * Part1 的id 小于Part2的id
 */
type Link struct {
	Part1 RouterID
	Part2 RouterID
	Val1  Amount
	Val2  Amount
	FeeRate float64
}

// 生成link的key
func GetLinkKey(r1, r2 RouterID) string {
	if r1 < r2 {
		return strconv.Itoa(int(r1)) + "-" + strconv.Itoa(int(r2))
	} else {
		return strconv.Itoa(int(r2)) + "-" + strconv.Itoa(int(r1))
	}
}

func GetLinkValue(from, to RouterID, linkBase map[string]*Link) Amount {
	key := GetLinkKey(from, to)
	if link, ok := linkBase[key]; ok {
		if from < to {
			return link.Val1
		} else {
			return link.Val2
		}
	}
	return 0
}

func GetLinkFeeRate(from, to RouterID, linkBase map[string]*Link) Amount {
	key := GetLinkKey(from, to)
	if link, ok := linkBase[key]; ok {
		return link.FeeRate
	}
	return 0
}

// 更新link的值
func UpdateLinkValue(from, to RouterID, linkBase map[string]*Link,
	amount Amount, flag bool) error {
	key := GetLinkKey(from, to)
	if link, ok := linkBase[key]; ok {
		if from < to {
			if flag == true {
				link.Val1 += amount
			} else {
				if link.Val1 < amount {
					return fmt.Errorf("the value of this channel " +
						"is insufficent")
				}
				link.Val1 -= amount
			}
		} else {
			if flag == true {
				link.Val2 += amount
			} else {
				if link.Val2 < amount {
					return fmt.Errorf("the value of this channel " +
						"is insufficent")
				}
				link.Val2 -= amount
			}
		}
	} else {
		if flag == false {
			return fmt.Errorf("the value of this channel " +
				"is insufficent")
		} else {
			linkBase[key] = &Link{}
			if from < to {
				linkBase[key].Part1 = from
				linkBase[key].Part2 = to
				linkBase[key].Val1 = amount
			} else {
				linkBase[key].Part1 = to
				linkBase[key].Part2 = from
				linkBase[key].Val2 = amount
			}
		}
	}
	return nil
}

func CopyChannels(src map[string]*Link) map[string]*Link {
	res := make(map[string]*Link)
	for key, val := range src {
		res[key] = &Link{
			Part1: val.Part1,
			Part2: val.Part2,
			Val1:  val.Val1,
			Val2:  val.Val2,
			FeeRate: val.FeeRate,
		}
	}
	return res
}

func GetPathCap(path Path, linkBase map[string]*Link) Amount {
	val := math.MaxFloat64
	for i := 0; i < len(path)-1; i++{
		val = math.Min(val, float64(GetLinkValue(path[i], path[i+1], linkBase)))
	}
	return Amount(val)
}

func UpdateWeights(routes []Path, amts []Amount,
	linkBase map[string]*Link) error {

	if len(routes) != len(amts) {
		return fmt.Errorf("routes number is not equal to amts' ")
	}
	for idx, route := range routes {
		for i := 0; i < len(route)-1; i++ {
			// i 到 i+1 的钱减少
			err := UpdateLinkValue(route[i], route[i+1],
				linkBase, amts[idx], false)
			if err != nil {
				return err
			}
			// i+1 到 i 的钱增加
			err = UpdateLinkValue(route[i+1], route[i],
				linkBase, amts[idx], true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func RanddomFeeRate(linkBase map[string]*Link) {
	largeRateNun := len(linkBase)/10
	cursor := 0
	rand.Seed(time.Now().UnixNano())
	for key := range linkBase {
		cursor++
		if cursor < largeRateNun {
			linkBase[key].FeeRate = rand.Float64() * 0.09 + 0.01
		} else {
			linkBase[key].FeeRate = rand.Float64() * 0.009 + 0.001
		}
	}
}
