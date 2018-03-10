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

// when processing a query we receive floats, but the rest of our structure uses rank space
// This function finds the corresponding rank-space coordinates for the query range.
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

// This is a neat trick to find the LCA of two nodes on the *same* level (that detail is important!)
// It of course only works when we zero-index and use a heap-layout.
func lowestCommonAncestor(left, right int) int {
	xor := uint64((left + 1) ^ (right + 1))
	zeros := bits.LeadingZeros64(xor)
	shift := uint(64 - zeros)
	return ((left + 1) >> shift) - 1
}

// currentIndex denotes a y-rank at the current node.
// when descending we want to maintain an interval [l,r] such that all y-coordinates to be reported fall in that range.
// this function is used when descending to the left.
func descendLeft(currentIndex int, rankSelectStruct gorasp.RankSelect) int {
	onesLeft := rankSelectStruct.RankOfIndex(currentIndex)
	zeros := currentIndex - int(onesLeft)
	return zeros
}

// similar to descendLeft
func descendRight(currentIndex int, rankSelectStruct gorasp.RankSelect) int {
	onesLeft := rankSelectStruct.RankOfIndex(currentIndex)
	return int(onesLeft)
}

// the last half of the array self.xTree are leaves.
// this function tests if a node represented by n is in fact a leaf.
func isLeaf(n int, self *RangeSearchAdvanced) bool {
	return n >= len(self.xTree)/2
}

// Reports everything hanging at or below node, with y-ranks [yLeft, yRight[ (half open interval).
func (self *RangeSearchAdvanced) reportAll(node, yLeft, yRight int) []int {
	if isLeaf(node, self) {
		indexInSortedPoints := node - len(self.xTree)/2
		if yLeft < yRight {
			return []int{self.pointsRankSpace[indexInSortedPoints].i}
		}
		return []int{}
	}

	result := make([]int, yRight-yLeft)
	for i := yLeft; i < yRight; i++ {
		result[i-yLeft] = self.ballInheritance[node][i]
	}
	return result
}

// appends the smaller list to the bigger list.
func merge(left, right []int) []int {
	if len(right) > len(left) {
		for _, v := range left {
			right = append(right, v)
		}
		return right
	} else {
		for _, v := range right {
			left = append(left, v)
		}
		return left
	}
}

// After finding the lca, this function is called with node=lca's right child.
// This function then keeps descending toward the node with key xRankMax, while
// reporting all subtrees that are strictly to the left.
func (self *RangeSearchAdvanced) reportLeftHanging(node, yLeft, yRight, xRankMax int) []int {
	if isLeaf(node, self) {
		if self.xTree[node] == -1 {
			return []int{}
		}
		index := node - len(self.xTree)/2
		point := self.pointsRankSpace[index]
		if point.x < xRankMax && point.y >= yLeft && yLeft < yRight {
			return []int{point.i}
		}
		return []int{}
	}
	rightChild := 2*node + 2
	leftChild := 2*node + 1

	keyOfMe := self.xTree[node]
	if keyOfMe == -1 {
		return []int{}
	}
	if xRankMax > keyOfMe {
		// report left childs everything.
		yLeftTmp := descendLeft(yLeft, self.rankSelectStructures[node])
		yRightTmp := descendLeft(yRight, self.rankSelectStructures[node])
		result_left := self.reportAll(leftChild, yLeftTmp, yRightTmp)

		// then descend right.
		yLeftNew := descendRight(yLeft, self.rankSelectStructures[node])
		yRightNew := descendRight(yRight, self.rankSelectStructures[node])
		result_right := self.reportLeftHanging(rightChild, yLeftNew, yRightNew, xRankMax)
		return merge(result_left, result_right)
	} else {
		// descendRight and do the same again.
		yLeftNew := descendLeft(yLeft, self.rankSelectStructures[node])
		yRightNew := descendLeft(yRight, self.rankSelectStructures[node])
		return self.reportLeftHanging(leftChild, yLeftNew, yRightNew, xRankMax)
	}
}

// symmetric to reportLeftHanging
func (self *RangeSearchAdvanced) reportRightHanging(node, yLeft, yRight, xRankMin int) []int {
	if isLeaf(node, self) {
		index := node - len(self.xTree)/2
		point := self.pointsRankSpace[index]
		if point.x >= xRankMin && point.y >= yLeft && yLeft < yRight {
			return []int{point.i}
		}
		return []int{}
	}

	rightChild := 2*node + 2
	leftChild := 2*node + 1

	keyOfMe := self.xTree[node]
	if keyOfMe == -1 {
		return []int{}
	}

	if xRankMin <= keyOfMe {
		// report right childs everything.
		yLeftTmp := descendRight(yLeft, self.rankSelectStructures[node])
		yRightTmp := descendRight(yRight, self.rankSelectStructures[node])
		result_left := self.reportAll(rightChild, yLeftTmp, yRightTmp)

		// then descend left.
		yLeftNew := descendLeft(yLeft, self.rankSelectStructures[node])
		yRightNew := descendLeft(yRight, self.rankSelectStructures[node])
		result_right := self.reportRightHanging(leftChild, yLeftNew, yRightNew, xRankMin)
		return merge(result_left, result_right)
	} else {
		// descendRight and do the same again.
		yLeftNew := descendRight(yLeft, self.rankSelectStructures[node])
		yRightNew := descendRight(yRight, self.rankSelectStructures[node])
		return self.reportRightHanging(rightChild, yLeftNew, yRightNew, xRankMin)
	}
}

// This function computes the y-rank-interval at a node lca, given that at the root the interval is [yLeft, yRight[ (half open).
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

// convenience function, called with node=lca when processing a query.
// Initiates the search towards the lower x-coordinate in the query range.
func (self *RangeSearchAdvanced) branchLeftReport(node, yLeft, yRight, xMinRank int) []int {
	leftChild := 2*node + 1
	yLeftNew := descendLeft(yLeft, self.rankSelectStructures[node])
	yRightNew := descendLeft(yRight, self.rankSelectStructures[node])
	return self.reportRightHanging(leftChild, yLeftNew, yRightNew, xMinRank)
}

// symmetric to branchLeftReport
func (self *RangeSearchAdvanced) branchRightReport(node, yLeft, yRight, xMaxRank int) []int {
	rightChild := 2*node + 2
	yLeftNew := descendRight(yLeft, self.rankSelectStructures[node])
	yRightNew := descendRight(yRight, self.rankSelectStructures[node])
	return self.reportLeftHanging(rightChild, yLeftNew, yRightNew, xMaxRank)
}

// determines if point is contained in the rectangle defined by bottomLeft, topRight.
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

// The query algorithm for the structure.
// Assumes bottomLeft, is in fact less than topRight on both the x and y coordinates.
// return a slice of indices, each is an index into self.points, which is in the order it was given to the constructor.
func (self *RangeSearchAdvanced) Query(bottomLeft, topRight Point) []int {
	bottomLeftRank, topRightRank := self.getRankSpacePoints(bottomLeft, topRight)
	leafIndexLeft := len(self.xTree)/2 + bottomLeftRank.x
	leafIndexRight := len(self.xTree)/2 + topRightRank.x

	onePastLastLeafIndex := len(self.xTree)/2 + len(self.points)
	// special case
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

	// general case
	var lca int = 0
	if leafIndexRight < len(self.xTree) {
		lca = lowestCommonAncestor(leafIndexLeft, leafIndexRight)
	} else {
		lca = lowestCommonAncestor(leafIndexLeft, leafIndexRight-1)
	}
	yLeft, yRight := self.descendToLca(lca, bottomLeftRank.y, topRightRank.y)

	result := self.branchLeftReport(lca, yLeft, yRight, bottomLeftRank.x)
	result = append(
		result,
		self.branchRightReport(lca, yLeft, yRight, topRightRank.x)...,
	)
	return result
}

// helper function used to build ball-inheritace and bit arrays.
// This function is called with elements in self.pointsRankspace by increasing y-rank.
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

// Helper function for building the bit arrays and ball-inheritance structure.
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

// helper function to initialize all the important arrays.
func (self *RangeSearchAdvanced) initializeRankSelectBallInheritance() {
	numberOfInternalNodes := len(self.xTree) / 2
	self.bitArrays = make([][]int, numberOfInternalNodes)
	self.rankSelectStructures = make([]gorasp.RankSelect, numberOfInternalNodes)
	self.ballInheritance = make([][]int, numberOfInternalNodes)
}

// Set the keys of the leaves appropriately.
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

// Set the keys of internal nodes appropriately.
// We define the key of an internal node to be the largest key in its left subtree.
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

// Builds the xTree.
func (self *RangeSearchAdvanced) makeTreeOnXAxis() {
	arrayLength := 2*getNextPowerOfTwo(len(self.pointsRankSpace)) - 1
	self.xTree = make([]int, arrayLength)
	sort.Sort(byXRank(self.pointsRankSpace))

	self.xTree = setLeavesOfXTree(self.xTree, self.pointsRankSpace)

	self.xTree = setInternalNodesOfXTree(self.xTree)

	self.setXTreeHeight()
}

// compute the height of xTree and set the field xTreeHeight.
// Height is here defined as the number of nodes from root-to-leaf including both.
func (self *RangeSearchAdvanced) setXTreeHeight() {
	arrayLength := len(self.xTree)
	height := uint(1)
	for 1<<height < arrayLength {
		height++
	}
	self.xTreeHeight = int(height)
}

// Since our inputs are floats, we reduce everything to rankspace first.
// This is done by sorting and binary searching each point.
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

// This function must be called before any Query can be called.
func (self *RangeSearchAdvanced) Build() {
	self.makeRankSpace()
	self.makeTreeOnXAxis()
	self.buildRankSelectAndBallInheritance()
}

// Constructor: takes a slice of points. These are the points we want to build the structure on.
func NewRangeSearchAdvanced(points []Point) *RangeSearchAdvanced {
	result := new(RangeSearchAdvanced)
	result.points = points
	return result
}
