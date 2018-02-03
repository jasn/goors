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
	bitArrays            [][]int
	rankSelectStructures []gorasp.RankSelect
	ballInheritance      [][]int
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

func (self *RangeSearchAdvanced) searchAndAppend(point pointRankPerm) {
	var recursivelySearchAndAppend func(node, height int)
	recursivelySearchAndAppend = func(node, height int) {
		firstLeafIndex := len(self.xTree) / 2
		if node >= firstLeafIndex {
			return
		}
		key := self.xTree[node]
		if point.x <= key {
			self.bitArrays[height] = append(self.bitArrays[height], 0)
			self.ballInheritance[height] = append(self.ballInheritance[height], point.i)
			recursivelySearchAndAppend(2*node+1, height+1)
		} else {
			self.bitArrays[height] = append(self.bitArrays[height], 1)
			self.ballInheritance[height] = append(self.ballInheritance[height], point.i)
			recursivelySearchAndAppend(2*node+2, height+1)
		}
	}
	root := 0
	rootHeight := 0
	recursivelySearchAndAppend(root, rootHeight)
}

func (self *RangeSearchAdvanced) buildRankSelectAndBallInheritance() {
	self.initializeRankSelectBallInheritance()

	sort.Sort(byYRank(self.pointsRankSpace))
	for _, p := range self.pointsRankSpace {
		self.searchAndAppend(p)
	}
	for i := 0; i < self.xTreeHeight; i++ {
		self.rankSelectStructures[i] = gorasp.NewRankSelectFast(self.bitArrays[i])
	}
	sort.Sort(byXRank(self.pointsRankSpace))
}

func (self *RangeSearchAdvanced) initializeRankSelectBallInheritance() {
	heightWithoutLeaves := self.xTreeHeight - 1
	self.bitArrays = make([][]int, heightWithoutLeaves)
	self.rankSelectStructures = make([]gorasp.RankSelect, heightWithoutLeaves)
	self.ballInheritance = make([][]int, heightWithoutLeaves)
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

	self.setXTreeHeight()
}

func (self *RangeSearchAdvanced) setXTreeHeight() {
	arrayLength := len(self.xTree)
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
