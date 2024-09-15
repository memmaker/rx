package console

import (
	"RogueUI/gridmap"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
)

type RadialExplosionAnimation struct {
	*BaseAnimation
	drawables    map[geometry.Point]textiles.TextIcon
	currentFrame int
	dijkstra     map[geometry.Point]int

	closedPositions map[geometry.Point]bool
	maxDist         int

	lightColor fxtools.HDRColor

	lights []*gridmap.LightSource
}

func NewRadialExplosionAnimation(dijkstra map[geometry.Point]int, lightColor fxtools.HDRColor, done func()) *RadialExplosionAnimation {
	drawables := map[geometry.Point]textiles.TextIcon{}
	maxDist := 0
	for _, dist := range dijkstra {
		if dist > maxDist {
			maxDist = dist
		}
	}
	return &RadialExplosionAnimation{
		BaseAnimation: &BaseAnimation{
			done: done,
		},
		drawables:       drawables,
		dijkstra:        dijkstra,
		maxDist:         maxDist,
		lightColor:      lightColor,
		closedPositions: make(map[geometry.Point]bool),
	}
}

func (p *RadialExplosionAnimation) GetLights() []*gridmap.LightSource {
	return p.lights
}

func (p *RadialExplosionAnimation) GetPriority() int {
	return 1
}

func (p *RadialExplosionAnimation) GetDrawables() map[geometry.Point]textiles.TextIcon {
	return p.drawables
}
func (p *RadialExplosionAnimation) curDist() int {
	return p.currentFrame * 10
}
func (p *RadialExplosionAnimation) isClosed(pos geometry.Point) bool {
	_, ok := p.closedPositions[pos]
	return ok
}
func (p *RadialExplosionAnimation) curPositions() []geometry.Point {
	var positions []geometry.Point
	for pos, dist := range p.dijkstra {
		if dist <= p.curDist() && !p.isClosed(pos) {
			positions = append(positions, pos)
		}
	}
	return positions
}

func (p *RadialExplosionAnimation) NextFrame() {
	if p.IsDone() {
		p.onFinishedOrCancelled()
		return
	}
	// next path index
	clear(p.drawables)
	p.lights = nil

	p.currentFrame = p.currentFrame + 1

	positions := p.curPositions()
	for _, pos := range positions {
		dist := p.dijkstra[pos]
		p.drawables[pos] = p.edgeIconForFrame(dist)
		p.lights = append(p.lights, &gridmap.LightSource{
			Pos:          pos,
			Color:        p.lightColor,
			Radius:       1,
			MaxIntensity: 1,
		})
	}

	for pos, _ := range p.closedPositions {
		dist := p.dijkstra[pos]
		p.drawables[pos] = p.centerIconForFrame(dist)
	}

	for _, pos := range positions {
		p.closedPositions[pos] = true
	}

	if p.IsDone() {
		p.onFinishedOrCancelled()
		return
	}
}

func (p *RadialExplosionAnimation) IsDone() bool {
	return p.curDist() >= p.maxDist && len(p.closedPositions) == len(p.dijkstra)
}

func (p *RadialExplosionAnimation) edgeIconForFrame(dist int) textiles.TextIcon {
	return textiles.TextIcon{Char: '*', Fg: p.lightColor.ToRGBA()}
}

func (p *RadialExplosionAnimation) centerIconForFrame(dist int) textiles.TextIcon {
	// ·•*☼○+•·
	chars := []rune{'·', '•', '*', '☼', '○', '+', '•', '·'}
	return textiles.TextIcon{Char: chars[p.currentFrame%len(chars)], Fg: p.lightColor.ToRGBA()}
}
