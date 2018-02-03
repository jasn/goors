package goors

type RangeSearchAdvanced struct {
	points []Point
}

func (self *RangeSearchAdvanced) Query(bottomLeft, topRight Point) {
	var result = []int{}
	return result
}

func (self *RangeSearchAdvanced) Build() {

}

func NewRangeSearchAdvanced(points []Points) *RangeSearchAdvanced {
	result := new(RangeSearchAdvanced)
	result.points = points
	return result
}
