package console

import (
    "github.com/memmaker/go/geometry"
    "github.com/memmaker/go/textiles"
)

type CoverAnimation struct {
    *BaseAnimation
    framesLeft int
    drawables  map[geometry.Point]textiles.TextIcon
}

func NewCoverAnimation(position geometry.Point, icon textiles.TextIcon, turnCount int, done func()) *CoverAnimation {
    drawables := map[geometry.Point]textiles.TextIcon{}
    drawables[position] = icon
    return &CoverAnimation{
        BaseAnimation: &BaseAnimation{
            done: done,
        },
        framesLeft: turnCount,
        drawables:  drawables,
    }
}

func (p *CoverAnimation) GetPriority() int {
    return 1
}

func (p *CoverAnimation) GetDrawables() map[geometry.Point]textiles.TextIcon {
    return p.drawables
}

func (p *CoverAnimation) NextFrame() {
    if p.IsDone() {
        return
    }

    // next path index
    p.framesLeft = p.framesLeft - 1
    if p.IsDone() {
        p.onFinishedOrCancelled()
        return
    }
}

func (p *CoverAnimation) IsDone() bool {
    return p.framesLeft == 0 || p.finishedOrCancelled
}
