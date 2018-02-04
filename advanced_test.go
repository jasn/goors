package goors

import (
	"fmt"
	"testing"
)

func testDescend(lca, yLeftExpected, yRightExpected int, structure *RangeSearchAdvanced) bool {
	success := true
	yLeft, yRight := structure.descendToLca(lca, 3, 9)
	if yLeft != yLeftExpected {
		fmt.Println("Incorrect descendToLca(", lca, "). yLeft =", yLeft, " expected", yLeftExpected, ".")
		success = false
	}
	if yRight != yRightExpected {
		fmt.Println("Incorrect descendToLca(", lca, "). yRight =", yRight, " expected", yRightExpected, ".")
		success = false
	}
	return success
}

func TestSomething(t *testing.T) {
	points := []Point{
		{3., 4.},
		{5., 5.},
		{4.5, 7.},
		{10., 3.},
		{9., 8.},
		{2., 9.5},
		{4., 10.},
		{2.5, 2.},
		{8., 6.},
		{7., 3.5},
		{6., 9.},
	}
	structure := NewRangeSearchAdvanced(points)
	structure.Build()

	bitArray := structure.bitArrays[1]
	correctBits := []int{0, 1, 0, 1, 1, 1, 0, 0}
	for i, bit := range bitArray {
		if bit != correctBits[i] {
			fmt.Println(i, "th bit incorrect. Expected", bitArray[i], "received", bit)
			t.Fail()
		}
	}
	lcas := []int{1, 4, 9, 2}
	yLeftExpecteds := []int{2, 1, 0, 1}
	yRightExpecteds := []int{6, 4, 2, 3}
	for i := 0; i < len(lcas); i++ {
		if !testDescend(lcas[i], yLeftExpecteds[i], yRightExpecteds[i], structure) {
			t.Fail()
		}
	}
}

func TestLowestCommonAncestor(t *testing.T) {
	lefts := []int{1, 3, 4, 8, 9, 19}
	rights := []int{2, 4, 6, 9, 11, 22}
	answers := []int{0, 1, 0, 1, 0, 4}

	for i := 0; i < len(lefts); i++ {
		lca := lowestCommonAncestor(lefts[i], rights[i])
		if answers[i] != lca {
			fmt.Println("error: LCA(", lefts[i], ",", rights[i], ") returned", lca, " but expected", answers[i])
			t.Fail()
		}
	}
}

func TestMakeTreeOnXAxis(t *testing.T) {
	n := 6
	points := make([]pointRankPerm, n)
	for i := 0; i < n; i++ {
		points[i].x = i
	}

	ps := []Point{}
	ds := NewRangeSearchAdvanced(ps)

	ds.pointsRankSpace = points
	ds.makeTreeOnXAxis()
	xTree := ds.xTree
	height := ds.xTreeHeight
	if height != 4 {
		fmt.Println("Expected height 4 but received height", height)
		t.Fail()
	}
	maxInt := 1<<31 - 1
	correctArray := []int{3, 1, 5, 0, 2, 4, maxInt, 0, 1, 2, 3, 4, 5, maxInt, maxInt}
	correct := true
	for i, v := range correctArray {
		if xTree[i] != v {
			t.Fail()
			fmt.Println("xTree is not correct")
			correct = false
		}
	}
	if !correct {
		fmt.Println("xTree:")
		fmt.Println(xTree)
		fmt.Println("correct:")
		fmt.Println(correctArray)
	}
}
