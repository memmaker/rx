package console

import (
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
)

type TransitionAnimation struct {
	*BaseAnimation
	drawables    map[geometry.Point]textiles.TextIcon
	currentFrame int
	bounds       geometry.Rect
	lookup       func(loc geometry.Point) (textiles.TextIcon, bool)
	getColor     func(colorName string) color.RGBA
}

func NewTransitionAnimation(target geometry.Rect, getColor func(colorName string) color.RGBA, lookup func(loc geometry.Point) (textiles.TextIcon, bool), done func()) *TransitionAnimation {
	drawables := map[geometry.Point]textiles.TextIcon{}
	target.Iter(func(pos geometry.Point) {
		icon, exists := lookup(pos)
		if exists {
			drawables[pos] = icon
		}
	})

	return &TransitionAnimation{
		BaseAnimation: &BaseAnimation{
			done: done,
		},
		bounds:    target,
		drawables: drawables,
		lookup:    lookup,
		getColor:  getColor,
	}
}

func (p *TransitionAnimation) GetPriority() int {
	return 1
}

func (p *TransitionAnimation) GetDrawables() map[geometry.Point]textiles.TextIcon {
	return p.drawables
}

func (p *TransitionAnimation) NextFrame() {
	if p.IsDone() {
		return
	}
	blackColor := p.getColor("Black")
	black := textiles.TextIcon{Char: ' ', Fg: blackColor, Bg: blackColor}
	// next path index
	p.currentFrame = p.currentFrame + 1
	if p.IsDone() {
		p.onFinishedOrCancelled()
		return
	}
	//clear(p.drawables)
	shrunkBounds := p.bounds.Shift(p.currentFrame, p.currentFrame, -p.currentFrame, -p.currentFrame)

	p.bounds.Iter(func(pos geometry.Point) {
		if !shrunkBounds.Contains(pos) {
			p.drawables[pos] = black
		}
	})
}

func (p *TransitionAnimation) IsDone() bool {
	return p.currentFrame > p.bounds.Size().X/2 || p.currentFrame > p.bounds.Size().Y/2 || p.finishedOrCancelled
}
