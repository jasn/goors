package goors

import (
	"fmt"
	"testing"
)

func TestmakeTreeOnXAxis(t *testing.T) {
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
