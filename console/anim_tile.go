package console

import (
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
)

type TilesAnimation struct {
	*BaseAnimation
	positions    []geometry.Point
	frames       []textiles.TextIcon
	drawables    map[geometry.Point]textiles.TextIcon
	currentFrame int
}

func NewTilesAnimation(positions []geometry.Point, icons []textiles.TextIcon, done func()) *TilesAnimation {
	drawables := map[geometry.Point]textiles.TextIcon{}
	for _, pos := range positions {
		drawables[pos] = icons[0]
	}
	return &TilesAnimation{
		BaseAnimation: &BaseAnimation{
			done: done,
		},
		positions: positions,
		drawables: drawables,
		frames:    icons,
	}
}

func (p *TilesAnimation) GetPriority() int {
	return 1
}

func (p *TilesAnimation) GetDrawables() map[geometry.Point]textiles.TextIcon {
	return p.drawables
}

func (p *TilesAnimation) NextFrame() {
	if p.IsDone() {
		return
	}

	// next path index
	clear(p.drawables)
	p.currentFrame = p.currentFrame + 1
	if p.currentFrame >= len(p.frames) {
		p.onFinishedOrCancelled()
		return
	}
	for _, pos := range p.positions {
		p.drawables[pos] = p.frames[p.currentFrame]
	}
}

func (p *TilesAnimation) IsDone() bool {
	return p.currentFrame > len(p.frames)-1 || p.finishedOrCancelled
}
