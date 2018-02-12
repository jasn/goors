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
	queryResult          []int
	queryTopRight        Point
	queryBottomLeft      Point
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

func isLeaf(n int, self *RangeSearchAdvanced) bool {
	return n >= len(self.xTree)/2
}

func (self *RangeSearchAdvanced) reportAll(node, yLeft, yRight int) {
	if isLeaf(node, self) {
		indexInSortedPoints := node - len(self.xTree)/2
		point := self.points[self.pointsRankSpace[indexInSortedPoints].i]
		if isContained(self.queryBottomLeft, self.queryTopRight, point) {
			self.queryResult = append(self.queryResult, self.pointsRankSpace[indexInSortedPoints].i)
		}
		return
	}

	for i := yLeft; i < yRight; i++ {
		self.queryResult = append(self.queryResult, self.ballInheritance[node][i])
	}
	return
}

func (self *RangeSearchAdvanced) reportLeftHanging(node, yLeft, yRight, xRankMax int) {
	if isLeaf(node, self) {
		if self.xTree[node] == -1 {
			return
		}
		index := node - len(self.xTree)/2
		point := self.pointsRankSpace[index]
		if point.x < xRankMax && point.y >= yLeft && yLeft < yRight {
			self.queryResult = append(self.queryResult, point.i)
		}
		return
	}
	rightChild := 2*node + 2
	leftChild := 2*node + 1

	keyOfMe := self.xTree[node]
	if keyOfMe == -1 {
		return
	}
	if xRankMax > keyOfMe {
		// report left childs everything.
		yLeftTmp := descendLeft(yLeft, self.rankSelectStructures[node])
		yRightTmp := descendLeft(yRight, self.rankSelectStructures[node])
		self.reportAll(leftChild, yLeftTmp, yRightTmp)

		// then descend right.
		yLeftNew := descendRight(yLeft, self.rankSelectStructures[node])
		yRightNew := descendRight(yRight, self.rankSelectStructures[node])
		self.reportLeftHanging(rightChild, yLeftNew, yRightNew, xRankMax)
	} else {
		// descendRight and do the same again.
		yLeftNew := descendLeft(yLeft, self.rankSelectStructures[node])
		yRightNew := descendLeft(yRight, self.rankSelectStructures[node])
		self.reportLeftHanging(leftChild, yLeftNew, yRightNew, xRankMax)
	}
}

func (self *RangeSearchAdvanced) reportRightHanging(node, yLeft, yRight, xRankMin int) {
	if isLeaf(node, self) {
		index := node - len(self.xTree)/2
		point := self.pointsRankSpace[index]
		if point.x >= xRankMin && point.y >= yLeft && yLeft < yRight {
			self.queryResult = append(self.queryResult, point.i)
		}
		return
	}

	rightChild := 2*node + 2
	leftChild := 2*node + 1

	keyOfMe := self.xTree[node]
	if keyOfMe == -1 {
		return
	}
	if xRankMin <= keyOfMe {
		// report right childs everything.
		yLeftTmp := descendRight(yLeft, self.rankSelectStructures[node])
		yRightTmp := descendRight(yRight, self.rankSelectStructures[node])
		self.reportAll(rightChild, yLeftTmp, yRightTmp)

		// then descend left.
		yLeftNew := descendLeft(yLeft, self.rankSelectStructures[node])
		yRightNew := descendLeft(yRight, self.rankSelectStructures[node])
		self.reportRightHanging(leftChild, yLeftNew, yRightNew, xRankMin)
	} else {
		// descendRight and do the same again.
		yLeftNew := descendRight(yLeft, self.rankSelectStructures[node])
		yRightNew := descendRight(yRight, self.rankSelectStructures[node])
		self.reportRightHanging(rightChild, yLeftNew, yRightNew, xRankMin)
	}
}

func (self *RangeSearchAdvanced) descendToLca(lca, yLeft, yRight int) (int, int) {
	searchKey := self.xTree[lca]
	yLeftNew := yLeft
	yRightNew := yRight
	node := 0
	for node != lca && node < len(self.xTree) {
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

func (self *RangeSearchAdvanced) branchLeftReport(node, yLeft, yRight, xMinRank int) {
	leftChild := 2*node + 1
	yLeftNew := descendLeft(yLeft, self.rankSelectStructures[node])
	yRightNew := descendLeft(yRight, self.rankSelectStructures[node])
	self.reportRightHanging(leftChild, yLeftNew, yRightNew, xMinRank)
}

func (self *RangeSearchAdvanced) branchRightReport(node, yLeft, yRight, xMaxRank int) {
	rightChild := 2*node + 2
	yLeftNew := descendRight(yLeft, self.rankSelectStructures[node])
	yRightNew := descendRight(yRight, self.rankSelectStructures[node])
	self.reportLeftHanging(rightChild, yLeftNew, yRightNew, xMaxRank)
}

func isContained(bottomLeft, topRight, point Point) bool {
	return point.x >= bottomLeft.x && point.x <= topRight.x && point.y >= bottomLeft.y && point.y <= topRight.y
}

// Checks if both x-coordinates ended up in the same leaf.
// can happen either at the very end meaning the rightLeafIndex is one past the end of the array
func bothXCoordinatesInSameLeaf(leafIndexLeft, leafIndexRight, onePastLastLeafIndex int) bool {
	caseOne := leafIndexLeft == leafIndexRight
	caseTwo := leafIndexRight == onePastLastLeafIndex && leafIndexLeft == onePastLastLeafIndex-1
	return caseOne || caseTwo
}

func (self *RangeSearchAdvanced) Query(bottomLeft, topRight Point) []int {
	self.queryBottomLeft = bottomLeft
	self.queryTopRight = topRight
	bottomLeftRank, topRightRank := self.getRankSpacePoints(bottomLeft, topRight)
	leafIndexLeft := len(self.xTree)/2 + bottomLeftRank.x
	leafIndexRight := len(self.xTree)/2 + topRightRank.x

	onePastLastLeafIndex := len(self.xTree)/2 + len(self.points)
	if bothXCoordinatesInSameLeaf(leafIndexLeft, leafIndexRight, onePastLastLeafIndex) {
		if leafIndexLeft >= onePastLastLeafIndex {
			result := make([]int, 0)
			return result
		}
		idx := leafIndexLeft - len(self.xTree)/2
		point := self.points[self.pointsRankSpace[idx].i]
		if isContained(bottomLeft, topRight, point) {
			result := make([]int, 1)
			result[0] = self.pointsRankSpace[idx].i
			return result
		} else {
			result := make([]int, 0)
			return result
		}
	}

	var lca int = 0
	if leafIndexRight < len(self.xTree) {
		lca = lowestCommonAncestor(leafIndexLeft, leafIndexRight)
	} else {
		lca = lowestCommonAncestor(leafIndexLeft, leafIndexRight-1)
	}
	yLeft, yRight := self.descendToLca(lca, bottomLeftRank.y, topRightRank.y)

	self.branchLeftReport(lca, yLeft, yRight, bottomLeftRank.x)
	self.branchRightReport(lca, yLeft, yRight, topRightRank.x)

	result := make([]int, len(self.queryResult))
	copy(result, self.queryResult)
	self.queryResult = make([]int, 0, 16)
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
			self.ballInheritance[node] = append(self.ballInheritance[node], point.i)
			recursivelySearchAndAppend(2*node+1, height+1)
		} else {
			self.bitArrays[node] = append(self.bitArrays[node], 1)
			self.ballInheritance[node] = append(self.ballInheritance[node], point.i)
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
	numberOfInternalNodes := len(self.xTree) / 2
	self.bitArrays = make([][]int, numberOfInternalNodes)
	self.rankSelectStructures = make([]gorasp.RankSelect, numberOfInternalNodes)
	self.ballInheritance = make([][]int, numberOfInternalNodes)
}

func setLeavesOfXTree(xTree []int, pointsRankSpace []pointRankPerm) []int {
	arrayLength := len(xTree)
	leafsStartAt := arrayLength / 2
	for i, p := range pointsRankSpace {
		index := i + leafsStartAt
		xTree[index] = p.x
	}

	leafsEndAt := leafsStartAt + len(pointsRankSpace)
	noDataValue := -1
	for i := leafsEndAt; i < len(xTree); i++ {
		xTree[i] = noDataValue
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
		// put -1 as key of internal nodes where no descendant leaf contains input data.
		if xTree[2*i+1] == -1 && xTree[2*i+2] == -1 {
			xTree[i] = -1
		} else {
			key := maxLeftSubTree(i)
			if key != -1 {
				xTree[i] = key
			} else {
				xTree[i] = int(1<<31 - 1)
			}
		}
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
	self.queryResult = make([]int, 0, 16)
}

func NewRangeSearchAdvanced(points []Point) *RangeSearchAdvanced {
	result := new(RangeSearchAdvanced)
	result.points = points
	return result
}
