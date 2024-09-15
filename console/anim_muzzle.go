package console

import (
	"RogueUI/gridmap"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
)

type MuzzleAnimation struct {
	*BaseAnimation
	framesLeft int
	light      *gridmap.LightSource
}

func NewMuzzleAnimation(position geometry.Point, lightColor fxtools.HDRColor, radius int, bulletCount int, done func()) *MuzzleAnimation {
	bulletCount = min(bulletCount, 3)
	return &MuzzleAnimation{
		BaseAnimation: &BaseAnimation{
			done: done,
		},

		framesLeft: bulletCount * 2,
		light: &gridmap.LightSource{
			Pos:          position,
			Radius:       radius,
			Color:        lightColor,
			MaxIntensity: 1,
		},
	}
}

func (p *MuzzleAnimation) GetLights() []*gridmap.LightSource {
	if p.framesLeft%2 == 1 {
		return nil
	}
	return []*gridmap.LightSource{p.light}
}

func (p *MuzzleAnimation) GetPriority() int {
	return 1
}

func (p *MuzzleAnimation) GetDrawables() map[geometry.Point]textiles.TextIcon {
	return nil
}

func (p *MuzzleAnimation) NextFrame() {
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

func (p *MuzzleAnimation) IsDone() bool {
	return p.framesLeft == 0 || p.finishedOrCancelled
}
