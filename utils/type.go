package utils

type RouterID int
type Amount float64
type Path []RouterID

func CopyPath (path Path) Path {
	res := make([]RouterID, 0)
	for _, id := range path {
		res = append(res, id)
	}
	return res
}
