package utils

import (
	"strconv"
	"fmt"
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
