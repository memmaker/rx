package util

import (
	"RogueUI/geometry"
	"math"
)

// Draw circle of radius r using given putPixel function; use simple
// square root method.
func DrawCircleSqrt(originX, originY, r int, putPixel func(x, y int)) {
	putP := func(x, y int) {
		putPixel(x+originX, y+originY)
	}
	y := r
	rsq := r * r
	for x := 0; x <= y; x++ {
		// Just calculate y = sqrt(r^2 - x^2)
		y := int(math.Round(math.Sqrt(float64(rsq - x*x))))
		putP(x, y)
		putP(y, x)
		putP(-x, y)
		putP(-y, x)
		putP(x, -y)
		putP(y, -x)
		putP(-x, -y)
		putP(-y, -x)
	}
}
func DrawCircleSqrtSegmented(originX, originY, r int, rotation float64, putPixel func(x, y int)) {
	y := r
	rsq := r * r

	dirVec := geometry.Point{X: 0, Y: -r}
	dirVec = geometry.RotateVector(dirVec, rotation)
	putP := func(x, y int) {
		if AbsInt(dirVec.X-x) <= 1 && AbsInt(dirVec.Y-y) <= 1 {
			return
		}
		putPixel(x+originX, y+originY)
	}
	for x := 0; x <= y; x++ {
		// Just calculate y = sqrt(r^2 - x^2)
		y := int(math.Round(math.Sqrt(float64(rsq - x*x))))
		putP(x, y)
		putP(y, x)
		putP(-x, y)
		putP(-y, x)
		putP(x, -y)
		putP(y, -x)
		putP(-x, -y)
		putP(-y, -x)
	}
}

// Draw circle of radius r using given putPixel function; use
// Bresenham-ish method with no sqrt and only integer math.
func drawCircleInt(r int, putPixel func(x, y int)) {
	x := 0
	y := r
	xsq := 0
	rsq := r * r
	ysq := rsq
	// Loop x from 0 to the line x==y. Start y at r and each time
	// around the loop either keep it the same or decrement it.
	for x <= y {
		putPixel(x, y)
		putPixel(y, x)
		putPixel(-x, y)
		putPixel(-y, x)
		putPixel(x, -y)
		putPixel(y, -x)
		putPixel(-x, -y)
		putPixel(-y, -x)

		// New x^2 = (x+1)^2 = x^2 + 2x + 1
		xsq = xsq + 2*x + 1
		x++
		// Potential new y^2 = (y-1)^2 = y^2 - 2y + 1
		y1sq := ysq - 2*y + 1
		// Choose y or y-1, whichever gives smallest error
		a := xsq + ysq
		b := xsq + y1sq
		if a-rsq >= rsq-b {
			y--
			ysq = y1sq
		}
	}
}
