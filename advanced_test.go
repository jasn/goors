package goors

import (
	"fmt"
	"math"
	"math/rand"
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
	// Input drawn on a piece of paper.
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

	a := Point{1.9, 3.9}
	b := Point{6.5, 9.2}
	result := structure.Query(a, b)
	for _, v := range result {
		fmt.Println(v, "->", points[v])
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
	noData := -1
	correctArray := []int{3, 1, 5, 0, 2, 4, noData, 0, 1, 2, 3, 4, 5, noData, noData}
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

func setupFourElements() RangeSearch {
	points := []Point{{0.0, 0.0}, {5.0, 5.0}, {10.0, 10.0}, {15.0, 15.0}}

	ds := NewRangeSearchAdvanced(points)
	ds.Build()
	return ds
}

func TestFourElementsEmptyResult(t *testing.T) {
	ds := setupFourElements()

	bottomLeft := Point{8.0, 8.0}
	topRight := Point{9.0, 9.0}
	result := ds.Query(bottomLeft, topRight)

	if len(result) != 0 {
		fmt.Println("Error, expected no results, but received", len(result))
		t.Fail()
	}
}

func TestFourElementsOneResultLeftLeaf(t *testing.T) {
	ds := setupFourElements()

	bottomLeft := Point{9.0, 9.0}
	topRight := Point{11.0, 11.0}
	result := ds.Query(bottomLeft, topRight)
	if len(result) != 1 || result[0] != 2 {
		fmt.Println(result)
		t.Fail()
	}
}

func TestFourElementsOneResultRightLeaf(t *testing.T) {
	ds := setupFourElements()

	bottomLeft := Point{14.0, 14.0}
	topRight := Point{16.0, 16.0}
	result := ds.Query(bottomLeft, topRight)
	if len(result) != 1 || result[0] != 3 {
		fmt.Println(result)
		t.Fail()
	}
}

func testFourElementsReportAll(t *testing.T) {
	ds := setupFourElements()

	bottomLeft := Point{0.0, 0.0}
	topRight := Point{15.0, 15.0}
	result := ds.Query(bottomLeft, topRight)
	if len(result) != 4 {
		fmt.Println("Expected 4 results, but received only", len(result))
		t.Fail()
	}

	for i := 0; i < len(result); i++ {
		found := false
		for i, v := range result {
			if v == i {
				found = true
			}
		}
		if !found {
			fmt.Println("Did not find point", i)
			t.Fail()
		}
	}
}

func testFourElementsOneResultRightLeafInMiddle(t *testing.T) {
	ds := setupFourElements()

	bottomLeft := Point{4.0, 4.0}
	topRight := Point{8.0, 8.0}
	result := ds.Query(bottomLeft, topRight)
	if len(result) != 1 {
		fmt.Println("Expected 1 results, but received only", len(result))
		t.Fail()
	}
}

func TestTenElementsReportLast(t *testing.T) {
	points := make([]Point, 10)
	for i := 0; i < 10; i++ {
		points[i] = Point{float64(i), float64(i)}
	}
	ds := NewRangeSearchAdvanced(points)
	ds.Build()
	bottomLeft := Point{8.5, 8.5}
	topRight := Point{9.5, 9.5}
	result := ds.Query(bottomLeft, topRight)
	if len(result) != 1 {
		fmt.Println("Expected 1 results, but received only", len(result))
		t.Fail()
	}
}

func TestTenElementsReportSixth(t *testing.T) {
	points := make([]Point, 10)
	for i := 0; i < 10; i++ {
		points[i] = Point{float64(i), float64(i)}
	}
	ds := NewRangeSearchAdvanced(points)
	ds.Build()
	bottomLeft := Point{4.5, 4.5}
	topRight := Point{5.5, 5.5}
	result := ds.Query(bottomLeft, topRight)
	if len(result) != 1 || result[0] != 5 {
		fmt.Println("Expected 1 results, but received only", len(result))
		t.Fail()
	}
}

func TestRandom(t *testing.T) {
	size := 1234
	points := make([]Point, size)
	rand.Seed(42)
	for i := 0; i < size; i++ {
		points[i] = Point{float64(rand.Float32()), float64(rand.Float32())}
	}
	doPrint := false
	if doPrint {
		fmt.Println("[")
		for _, v := range points {
			fmt.Println("(", v.x, ",", v.y, "),")
		}
		fmt.Println("]")
	}
	dsAdvanced := NewRangeSearchAdvanced(points)
	dsSimple := NewRangeSearchSimple(points)

	dsSimple.Build()
	dsAdvanced.Build()

	numberOfQueries := 1000

	for i := 0; i < numberOfQueries; i++ {
		x1 := float64(rand.Float32())
		x2 := float64(rand.Float32())
		y1 := float64(rand.Float32())
		y2 := float64(rand.Float32())

		bottomLeft := Point{math.Min(x1, x2), math.Min(y1, y2)}
		topRight := Point{math.Max(x1, x2), math.Max(y1, y2)}
		if doPrint {
			fmt.Println("querying:", bottomLeft, topRight, "iteration:", i)
		}
		resultAdvanced := dsAdvanced.Query(bottomLeft, topRight)
		resultSimple := dsSimple.Query(bottomLeft, topRight)
		if len(resultSimple) != len(resultAdvanced) {
			fmt.Println("Simple and Advanced report different amount of elements")
			fmt.Println("len(simple):", len(resultSimple), " len(advanced):", len(resultAdvanced))
			fmt.Println("simple:")
			for _, v := range resultSimple {
				fmt.Println(points[v])
			}
			fmt.Println("advanced:")
			for _, v := range resultAdvanced {
				fmt.Println(points[v])
			}
			t.Fail()
		}
		for _, v := range resultSimple {
			found := false
			for _, v2 := range resultAdvanced {
				if v == v2 {
					found = true
				}
			}
			if !found {
				fmt.Println("did not find ", v)
				t.Fail()
			}
		}
	}
}
