package util

import (
	"RogueUI/geometry"
	"image/color"
)

func GetLoopingFrameFromTick(tick uint64, delayBetweenFramesInSeconds float64, frameCount int) int32 {
	ticksPerInterval := secondsToTicksF(delayBetweenFramesInSeconds)
	intervallCount := float64(tick) / ticksPerInterval
	return int32(intervallCount) % int32(frameCount)
}

func GetPercentageFromTick(tick uint64, cycleTimeInSeconds float64) float64 {
	ticksPerInterval := secondsToTicksF(cycleTimeInSeconds)
	intervallCount := float64(tick) / ticksPerInterval
	// only return the fractional part
	return intervallCount - float64(int64(intervallCount))
}

func SecondsToTicks(seconds float64) uint64 {
	return uint64(secondsToTicksF(seconds))
}

func secondsToTicksF(seconds float64) float64 {
	return safeTPS() * seconds
}

func safeTPS() float64 {
	return 20
}

func LerpPoint(start, end geometry.Point, percent float64) geometry.Point {
	return geometry.Point{
		X: int(float64(start.X) + float64(end.X-start.X)*percent),
		Y: int(float64(start.Y) + float64(end.Y-start.Y)*percent),
	}
}

func LerpPointF(start, end geometry.PointF, percent float64) geometry.PointF {
	return geometry.PointF{
		X: start.X + (end.X-start.X)*percent,
		Y: start.Y + (end.Y-start.Y)*percent,
	}
}

func LerpColorRGBA(start, end color.RGBA, percent float64) color.RGBA {
	r1, g1, b1, a1 := start.R, start.G, start.B, start.A
	r2, g2, b2, a2 := end.R, end.G, end.B, end.A
	return color.RGBA{
		R: uint8(float64(r1) + float64(r2-r1)*percent),
		G: uint8(float64(g1) + float64(g2-g1)*percent),
		B: uint8(float64(b1) + float64(b2-b1)*percent),
		A: uint8(float64(a1) + float64(a2-a1)*percent),
	}
}

func SetBrightness(start color.RGBA, percent float64) color.RGBA {
	r1, g1, b1 := start.R, start.G, start.B
	return color.RGBA{
		R: uint8(float64(r1) * percent),
		G: uint8(float64(g1) * percent),
		B: uint8(float64(b1) * percent),
		A: uint8(255),
	}
}
