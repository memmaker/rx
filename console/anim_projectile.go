package console

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
)

type BaseAnimation struct {
	finishedOrCancelled bool
	followUp            []foundation.Animation
	done                func()
	calledDone          bool
	requestMapUpdate    bool
	audioCue            string
}

func (p *BaseAnimation) Cancel() {
	p.onFinishedOrCancelled()
}
func (p *BaseAnimation) GetAudioCue() string {
	return p.audioCue
}
func (p *BaseAnimation) SetAudioCue(cueName string) {
	p.audioCue = cueName
}
func (p *BaseAnimation) IsRequestingMapStateUpdate() bool {
	return p.requestMapUpdate
}
func (p *BaseAnimation) RequestMapUpdateOnFinish() {
	p.requestMapUpdate = true
}
func (p *BaseAnimation) onFinishedOrCancelled() {
	p.finishedOrCancelled = true
	if !p.calledDone && p.done != nil {
		p.done()
		p.calledDone = true
	}
}
func (p *BaseAnimation) SetFollowUp(animations []foundation.Animation) {
	p.SetOrAppendFollowUp(animations)
}

func (p *BaseAnimation) SetOrAppendFollowUp(animations []foundation.Animation) {
	if p.followUp == nil {
		p.followUp = animations
	} else {
		p.followUp = append(p.followUp, animations...)
	}
}

func (p *BaseAnimation) GetFollowUp() []foundation.Animation {
	return p.followUp
}

type ProjectileAnimation struct {
	*BaseAnimation
	path             []geometry.Point
	icon             textiles.TextIcon
	drawables        map[geometry.Point]textiles.TextIcon
	currentPathIndex int
	lookup           func(loc geometry.Point) (textiles.TextIcon, bool)
	trail            []textiles.TextIcon
}

func NewProjectileAnimation(path []geometry.Point, icon textiles.TextIcon, lookup func(loc geometry.Point) (textiles.TextIcon, bool), done func()) *ProjectileAnimation {
	return &ProjectileAnimation{
		BaseAnimation: &BaseAnimation{
			done: done,
		},
		path:   path,
		icon:   icon,
		lookup: lookup,
		drawables: map[geometry.Point]textiles.TextIcon{
			path[0]: icon,
		},
	}
}

func (p *ProjectileAnimation) GetPriority() int {
	return 1
}

func (p *ProjectileAnimation) GetDrawables() map[geometry.Point]textiles.TextIcon {
	return p.drawables
}

func (p *ProjectileAnimation) NextFrame() {
	if p.IsDone() {
		return
	}

	// next path index
	clear(p.drawables)
	p.currentPathIndex = p.currentPathIndex + 1
	if p.currentPathIndex >= len(p.path) {
		p.onFinishedOrCancelled()
		return
	}
	drawIcon := p.icon
	if p.icon.Char < 0 && p.lookup != nil {
		icon, exists := p.lookup(p.path[p.currentPathIndex])
		if exists {
			drawIcon = icon.WithBg(p.icon.Bg)
		}
	}
	p.drawables[p.path[p.currentPathIndex]] = drawIcon

	if p.currentPathIndex > 0 && p.trail != nil {
		trailLength := min(len(p.trail), p.currentPathIndex)
		for i := 0; i < trailLength; i++ {
			pathPos := p.path[p.currentPathIndex-i-1]
			trailIcon := p.trail[i]
			if trailIcon.Char < 0 && p.lookup != nil {
				icon, exists := p.lookup(pathPos)
				if exists {
					trailIcon = icon.WithBg(trailIcon.Bg)
				}
			}
			p.drawables[pathPos] = trailIcon
		}
	}
}

func (p *ProjectileAnimation) IsDone() bool {
	return p.currentPathIndex > len(p.path)-1 || p.finishedOrCancelled
}

func (p *ProjectileAnimation) SetTrail(icons []textiles.TextIcon) {
	p.trail = icons
}
