package goors

type RangeSearch interface {
	Query(bottomLeft, topRight Point) []int
	Build()
}
