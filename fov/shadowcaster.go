package fov

import (
	"math"
)

type Fraction struct {
	Numerator, Denominator int
}

func NewFraction(numerator, denominator int) Fraction {
	return Fraction{numerator, denominator}
}

func (f Fraction) Value() float64 {
	return float64(f.Numerator) / float64(f.Denominator)
}

func ComputeFov(origin [2]int, isBlocking func(int, int) bool, markVisible func(int, int)) {

	markVisible(origin[0], origin[1])

	for i := 0; i < 4; i++ {
		quadrant := NewQuadrant(i, origin)

		reveal := func(tile [2]int) {
			x, y := quadrant.Transform(tile)
			markVisible(x, y)
		}

		isWall := func(tile *[2]int) bool {
			if tile == nil {
				return false
			}
			x, y := quadrant.Transform(*tile)
			return isBlocking(x, y)
		}

		isFloor := func(tile *[2]int) bool {
			if tile == nil {
				return false
			}
			x, y := quadrant.Transform(*tile)
			return !isBlocking(x, y)
		}

		var scan func(row Row)
		scan = func(row Row) {
			var prevTile *[2]int
			for _, tile := range row.Tiles() {
				if isWall(&tile) || isSymmetric(row, tile) {
					reveal(tile)
				}
				if prevTile != nil && isWall(prevTile) && isFloor(&tile) {
					row.startSlope = slope(tile)
				}
				if prevTile != nil && isFloor(prevTile) && isWall(&tile) {
					nextRow := row.Next()
					nextRow.endSlope = slope(tile)
					scan(nextRow)
				}
				prevTile = &tile
			}
			if prevTile != nil && isFloor(prevTile) {
				scan(row.Next())
			}
		}

		firstRow := NewRow(1, NewFraction(-1, 1), NewFraction(1, 1))
		scan(firstRow)
	}
}

type Quadrant struct {
	cardinal int
	ox, oy   int
}

const (
	North = 0
	East  = 1
	South = 2
	West  = 3
)

func NewQuadrant(cardinal int, origin [2]int) Quadrant {
	return Quadrant{cardinal: cardinal, ox: origin[0], oy: origin[1]}
}

func (q Quadrant) Transform(tile [2]int) (int, int) {
	row, col := tile[0], tile[1]
	switch q.cardinal {
	case North:
		return q.ox + col, q.oy - row
	case South:
		return q.ox + col, q.oy + row
	case East:
		return q.ox + row, q.oy + col
	case West:
		return q.ox - row, q.oy + col
	}
	return 0, 0
}

type Row struct {
	depth      int
	startSlope Fraction
	endSlope   Fraction
}

func NewRow(depth int, startSlope, endSlope Fraction) Row {
	return Row{depth: depth, startSlope: startSlope, endSlope: endSlope}
}

func (r Row) Tiles() [][2]int {
	minCol := roundTiesUp(float64(r.depth) * r.startSlope.Value())
	maxCol := roundTiesDown(float64(r.depth) * r.endSlope.Value())
	var tiles [][2]int
	for col := minCol; col <= maxCol; col++ {
		tiles = append(tiles, [2]int{r.depth, col})
	}
	return tiles
}

func (r Row) Next() Row {
	return NewRow(r.depth+1, r.startSlope, r.endSlope)
}

func slope(tile [2]int) Fraction {
	rowDepth, col := tile[0], tile[1]
	return NewFraction(2*col-1, 2*rowDepth)
}

func isSymmetric(row Row, tile [2]int) bool {
	_, col := tile[0], tile[1]
	return float64(col) >= float64(row.depth)*row.startSlope.Value() && float64(col) <= float64(row.depth)*row.endSlope.Value()
}

func roundTiesUp(n float64) int {
	return int(math.Floor(n + 0.5))
}

func roundTiesDown(n float64) int {
	return int(math.Ceil(n - 0.5))
}
