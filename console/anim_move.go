package console

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/util"
	"image/color"
)

type MovementAnimation struct {
	*BaseAnimation

	actor         foundation.ActorForUI
	originalPos   geometry.Point
	newPos        geometry.Point
	frameCount    int
	icon          foundation.TextIcon
	isQuickMove   bool
	quickMovePath []geometry.Point
	getColor      func(colorName string) color.RGBA
}

func NewMovementAnimation(actorIcon foundation.TextIcon, old, new geometry.Point, getColor func(colorName string) color.RGBA, done func()) *MovementAnimation {
	return &MovementAnimation{
		BaseAnimation: &BaseAnimation{
			done: done,
		},
		icon:        actorIcon,
		originalPos: old,
		newPos:      new,
		getColor:    getColor,
	}
}

func (p *MovementAnimation) Cancel() {
	p.onFinishedOrCancelled()
}

func (p *MovementAnimation) GetPriority() int {
	return 1
}

func (p *MovementAnimation) GetDrawables() map[geometry.Point]foundation.TextIcon {
	if p.isQuickMove {
		return p.quickMoveAnimation()
	}
	return p.normalAnimation()
}

func (p *MovementAnimation) normalAnimation() map[geometry.Point]foundation.TextIcon {
	if p.frameCount <= 2 {
		lightGray := p.getColor("LightGray")
		return map[geometry.Point]foundation.TextIcon{
			p.originalPos: p.icon.WithFg(lightGray),
			p.newPos:      p.icon,
		}
	}
	darkGray := p.getColor("DarkGray")
	return map[geometry.Point]foundation.TextIcon{
		p.originalPos: p.icon.WithFg(darkGray),
		p.newPos:      p.icon,
	}
}

func (p *MovementAnimation) NextFrame() {
	if p.IsDone() {
		return
	}

	// next path index
	p.frameCount++

	if p.IsDone() {
		p.onFinishedOrCancelled()
	}
}

func (p *MovementAnimation) IsDone() bool {
	return p.frameCount >= 5
}

func (p *MovementAnimation) EnableQuickMoveMode(path []geometry.Point) {
	p.isQuickMove = true
	p.quickMovePath = path
}

func (p *MovementAnimation) quickMoveAnimation() map[geometry.Point]foundation.TextIcon {
	// we want to draw one fading white line with background tiles
	// and the actor icon on top of it
	drawables := make(map[geometry.Point]foundation.TextIcon)
	for i, pos := range p.quickMovePath {
		if i == len(p.quickMovePath)-1 {
			drawables[pos] = p.icon
		} else {
			black := p.getColor("Black")
			white := p.getColor("White")
			percent := util.Clamp(0.1, 1.0, float64(i+1)/float64(len(p.quickMovePath)))
			lerpColorRGBA := util.LerpColorRGBA(black, white, percent)
			drawables[pos] = foundation.TextIcon{
				Rune: ' ',
				Fg:   lerpColorRGBA,
				Bg:   lerpColorRGBA,
			}
		}
	}
	return drawables
}
