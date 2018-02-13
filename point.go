package goors

type Point struct {
	x, y float64
}

func MakePoint(x, y float64) Point {
	return Point{x, y}
}
