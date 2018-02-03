package goors

type RangeSearchSimple struct {
	points []Point
}

func (self *RangeSearchSimple) Query(bottomLeft, topRight Point) []int {
	var result = []int{}
	isContained := func(p Point) {
		return point.x >= bottomLeft.x && point.x <= topRight.x &&
			point.y >= bottomLeft.y && point.y <= topRight.y
	}
	for index, point := range self.points {
		if isContained(point) {
			append(result, index)
		}
	}
	return result
}

func (self *RangeSearchSimple) Build() {

}

func NewRangeSearchSimple(points []Point) *RangeSearchSimple {
	result := new(RangeSearchSimple)
	result.points = points
	return result
}
