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

func CheckInPath (id RouterID, path Path)  bool {
	for _, n := range path {
		if n == id {
			return true
		}
	}
	return false
}
