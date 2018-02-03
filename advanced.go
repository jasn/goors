package goors

import (
	"github.com/jasn/gorasp"
	"sort"
)

type RangeSearchAdvanced struct {
	points               []Point
	pointsRankSpace      []pointRankPerm
	xTree                []int
	xTreeHeight          int // number of nodes on root to leaf path including root and leaf.
	rankSelectStructures []gorasp.RankSelect
}

func getNextPowerOfTwo(n int) int {
	if n <= 2 {
		return n
	}
	n = n - 1
	for i := uint(1); n > 0; i++ {
		n = n >> 1
		if n == 0 {
			return 1 << i
		}
	}
	return 1 << 30
}

func (self *RangeSearchAdvanced) Query(bottomLeft, topRight Point) []int {
	var result = []int{}
	return result
}

func setLeavesOfXTree(xTree []int, pointsRankSpace []pointRankPerm) []int {
	arrayLength := len(xTree)
	leafsStartAt := arrayLength / 2
	for i, p := range pointsRankSpace {
		index := i + leafsStartAt
		xTree[index] = p.x
	}

	leafsEndAt := leafsStartAt + len(pointsRankSpace)
	maxInt := 1<<31 - 1
	for i := leafsEndAt; i < len(xTree); i++ {
		xTree[i] = maxInt
	}
	return xTree
}

func setInternalNodesOfXTree(xTree []int) []int {
	var maxSubTree func(n int) int
	maxSubTree = func(n int) int {
		rightChild := 2*n + 2
		if rightChild > len(xTree) {
			return xTree[n]
		}
		return maxSubTree(rightChild)
	}

	maxLeftSubTree := func(n int) int {
		leftChild := 2*n + 1
		if leftChild > len(xTree) {
			return maxSubTree(n)
		}
		return maxSubTree(leftChild)
	}

	arrayLength := len(xTree)
	for i := arrayLength/2 - 1; i >= 0; i-- {
		key := maxLeftSubTree(i)
		xTree[i] = key
	}

	return xTree
}

func (self *RangeSearchAdvanced) makeTreeOnXAxis() {
	arrayLength := 2*getNextPowerOfTwo(len(self.pointsRankSpace)) - 1
	self.xTree = make([]int, arrayLength)
	sort.Sort(byXRank(self.pointsRankSpace))

	self.xTree = setLeavesOfXTree(self.xTree, self.pointsRankSpace)

	self.xTree = setInternalNodesOfXTree(self.xTree)

	self.setXTreeHeight(arrayLength)
}

func (self *RangeSearchAdvanced) setXTreeHeight(arrayLength int) {
	height := uint(1)
	for 1<<height < arrayLength {
		height++
	}
	self.xTreeHeight = int(height)
}

func (self *RangeSearchAdvanced) makeRankSpace() {
	xCoords := make([]float64, len(self.points))
	yCoords := make([]float64, len(self.points))
	for i, p := range self.points {
		xCoords[i] = p.x
		yCoords[i] = p.y
	}
	sort.Float64s(xCoords)
	sort.Float64s(yCoords)

	self.pointsRankSpace = make([]pointRankPerm, len(self.points))
	for index, p := range self.points {
		xRank := sort.Search(len(self.points), func(i int) bool { return p.x < xCoords[i] })
		yRank := sort.Search(len(self.points), func(i int) bool { return p.y < yCoords[i] })
		self.pointsRankSpace[index].x = xRank
		self.pointsRankSpace[index].y = yRank
		self.pointsRankSpace[index].i = index
	}
}

func (self *RangeSearchAdvanced) Build() {

}

func NewRangeSearchAdvanced(points []Point) *RangeSearchAdvanced {
	result := new(RangeSearchAdvanced)
	result.points = points
	return result
}
