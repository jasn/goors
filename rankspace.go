package goors

type pointRankPerm struct {
	x, y, i int
}

type byXRank []pointRankPerm

func (self byXRank) Len() int {
	return len(self)
}

func (self byXRank) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self byXRank) Less(i, j int) bool {
	return self[i].x < self[j].x
}

type byYRank []pointRankPerm

func (self byYRank) Len() int {
	return len(self)
}

func (self byYRank) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self byYRank) Less(i, j int) bool {
	return self[i].y < self[j].y
}
