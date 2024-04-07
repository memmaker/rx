package console

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"image/color"
)

type RadialAnimation struct {
	*BaseAnimation
	drawables    map[geometry.Point]foundation.TextIcon
	currentFrame int
	origin       geometry.Point
	dijkstra     map[geometry.Point]int
	lookup       func(loc geometry.Point) (foundation.TextIcon, bool)

	closedPositions          map[geometry.Point]bool
	maxDist                  int
	keepDrawingCoveredGround bool
	useIconColors            bool
	getColor                 func(colorName string) color.RGBA
}

func NewRadialAnimation(origin geometry.Point, dijkstra map[geometry.Point]int, getColor func(colorName string) color.RGBA, lookup func(loc geometry.Point) (foundation.TextIcon, bool), done func()) *RadialAnimation {
	drawables := map[geometry.Point]foundation.TextIcon{}
	maxDist := 0
	for _, dist := range dijkstra {
		if dist > maxDist {
			maxDist = dist
		}
	}
	return &RadialAnimation{
		BaseAnimation: &BaseAnimation{
			done: done,
		},
		drawables:       drawables,
		origin:          origin,
		dijkstra:        dijkstra,
		lookup:          lookup,
		maxDist:         maxDist,
		getColor:        getColor,
		closedPositions: make(map[geometry.Point]bool),
	}
}
func (p *RadialAnimation) SetUseIconColors(useIconColors bool) {
	p.useIconColors = useIconColors
}
func (p *RadialAnimation) SetKeepDrawingCoveredGround(keepDrawingCoveredGround bool) {
	p.keepDrawingCoveredGround = keepDrawingCoveredGround
}

func (p *RadialAnimation) GetPriority() int {
	return 1
}

func (p *RadialAnimation) GetDrawables() map[geometry.Point]foundation.TextIcon {
	return p.drawables
}
func (p *RadialAnimation) curDist() int {
	return p.currentFrame * 10
}
func (p *RadialAnimation) isClosed(pos geometry.Point) bool {
	_, ok := p.closedPositions[pos]
	return ok
}
func (p *RadialAnimation) curPositions() []geometry.Point {
	var positions []geometry.Point
	for pos, dist := range p.dijkstra {
		if dist <= p.curDist() && !p.isClosed(pos) {
			positions = append(positions, pos)
		}
	}
	return positions
}

func (p *RadialAnimation) NextFrame() {
	if p.IsDone() {
		p.onFinishedOrCancelled()
		return
	}
	// next path index
	clear(p.drawables)
	p.currentFrame = p.currentFrame + 1

	positions := p.curPositions()
	for _, pos := range positions {
		icon, ok := p.lookup(pos)
		if !ok {
			continue
		}
		if p.useIconColors {
			p.drawables[pos] = icon
		} else {
			p.drawables[pos] = icon.Reversed().WithFg(p.getColor("LightCyan"))
		}
	}

	if p.keepDrawingCoveredGround {
		for pos, _ := range p.closedPositions {
			icon, ok := p.lookup(pos)
			if !ok {
				continue
			}
			dist := p.dijkstra[pos]
			if dist <= p.curDist() && dist > p.curDist()-35 {
				lightGray := p.getColor("LightGray")
				black := p.getColor("Black")
				p.drawables[pos] = icon.WithFg(lightGray).WithBg(black)
			} else {
				p.drawables[pos] = icon
			}
		}
	}

	for _, pos := range positions {
		p.closedPositions[pos] = true
	}

	if p.IsDone() {
		p.onFinishedOrCancelled()
		return
	}
}

func (p *RadialAnimation) IsDone() bool {
	return p.curDist() >= p.maxDist && len(p.closedPositions) == len(p.dijkstra)
}
