package goors

import (
	"github.com/jasn/gorasp"
	"math/bits"
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
	xCoords              []float64
	yCoords              []float64
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

func (self *RangeSearchAdvanced) getRankSpacePoints(bottomLeft, topRight Point) (pointRankPerm, pointRankPerm) {
	bottomLeftRes := pointRankPerm{0, 0, -1}
	topRightRes := pointRankPerm{0, 0, -1}

	bottomLeftRes.x = sort.SearchFloat64s(self.xCoords, bottomLeft.x)
	bottomLeftRes.y = sort.SearchFloat64s(self.yCoords, bottomLeft.y)

	topRightRes.x = sort.Search(len(self.xCoords), func(i int) bool {
		return self.xCoords[i] > topRight.x
	})
	topRightRes.y = sort.Search(len(self.yCoords), func(i int) bool {
		return self.yCoords[i] > topRight.y
	})
	return bottomLeftRes, topRightRes
}

func lowestCommonAncestor(left, right int) int {
	xor := uint64((left + 1) ^ (right + 1))
	zeros := bits.LeadingZeros64(xor)
	shift := uint(64 - zeros)
	return ((left + 1) >> shift) - 1
}

func descendLeft(currentIndex int, rankSelectStruct gorasp.RankSelect) int {
	onesLeft := rankSelectStruct.RankOfIndex(currentIndex)
	zeros := currentIndex - int(onesLeft)
	return zeros
}

func descendRight(currentIndex int, rankSelectStruct gorasp.RankSelect) int {
	onesLeft := rankSelectStruct.RankOfIndex(currentIndex)
	return int(onesLeft)
}

func (self *RangeSearchAdvanced) reportLeftHanging(lca, yLeft, yRight int) {

}

func (self *RangeSearchAdvanced) reportRightHanging(lca, yLeft, yRight int) {

}

func (self *RangeSearchAdvanced) descendToLca(lca, yLeft, yRight int) (int, int) {
	searchKey := self.xTree[lca]
	yLeftNew := yLeft
	yRightNew := yRight
	node := 0
	for node != lca {
		key := self.xTree[node]
		if searchKey <= key {
			yLeftNew = descendLeft(yLeftNew, self.rankSelectStructures[node])
			yRightNew = descendLeft(yRightNew, self.rankSelectStructures[node])
			node = 2*node + 1
		} else {
			yLeftNew = descendRight(yLeftNew, self.rankSelectStructures[node])
			yRightNew = descendRight(yRightNew, self.rankSelectStructures[node])
			node = 2*node + 2
		}
	}
	return yLeftNew, yRightNew
}

func (self *RangeSearchAdvanced) Query(bottomLeft, topRight Point) []int {
	var result = []int{}
	bottomLeftRank, topRightRank := self.getRankSpacePoints(bottomLeft, topRight)
	lca := lowestCommonAncestor(bottomLeftRank.x, topRightRank.x)
	yLeft, yRight := self.descendToLca(lca, bottomLeftRank.y, topRightRank.y)
	_, _ = yLeft, yRight
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
			self.bitArrays[node] = append(self.bitArrays[node], 0)
			self.ballInheritance[height] = append(self.ballInheritance[height], point.i)
			recursivelySearchAndAppend(2*node+1, height+1)
		} else {
			self.bitArrays[node] = append(self.bitArrays[node], 1)
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
	numberOfInternalNodes := len(self.xTree) / 2
	for i := 0; i < numberOfInternalNodes; i++ {
		if len(self.bitArrays[i]) > 0 {
			self.rankSelectStructures[i] = gorasp.NewRankSelectFast(self.bitArrays[i])
		}
	}
	sort.Sort(byXRank(self.pointsRankSpace))
}

func (self *RangeSearchAdvanced) initializeRankSelectBallInheritance() {
	heightWithoutLeaves := self.xTreeHeight - 1
	numberOfInternalNodes := len(self.xTree) / 2
	self.bitArrays = make([][]int, numberOfInternalNodes)
	self.rankSelectStructures = make([]gorasp.RankSelect, numberOfInternalNodes)
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
		xRank := sort.Search(len(self.points), func(i int) bool { return p.x <= xCoords[i] })
		yRank := sort.Search(len(self.points), func(i int) bool { return p.y <= yCoords[i] })
		self.pointsRankSpace[index].x = xRank
		self.pointsRankSpace[index].y = yRank
		self.pointsRankSpace[index].i = index
	}
	self.xCoords = xCoords
	self.yCoords = yCoords
}

func (self *RangeSearchAdvanced) Build() {
	self.makeRankSpace()
	self.makeTreeOnXAxis()
	self.buildRankSelectAndBallInheritance()
}

func NewRangeSearchAdvanced(points []Point) *RangeSearchAdvanced {
	result := new(RangeSearchAdvanced)
	result.points = points
	return result
}
