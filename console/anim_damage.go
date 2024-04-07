package console

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"image/color"
	"strconv"
)

type DamageAnimation struct {
	*BaseAnimation
	pos       geometry.Point
	damage    int
	drawables map[geometry.Point]foundation.TextIcon
	ticksLeft int
}

func NewDamageAnimation(defenderPos geometry.Point, damage int) *DamageAnimation {
	damageRune := '!'
	fgColor := color.RGBA{R: 240, G: 20, B: 20, A: 255}
	bgColor := color.RGBA{R: 40, A: 255}
	if damage == 0 {
		damageRune = '-'
		fgColor = color.RGBA{R: 200, G: 200, B: 200, A: 255}
		bgColor = color.RGBA{R: 40, G: 40, B: 40, A: 255}
	} else if damage < 10 {
		asRunes := []rune(strconv.Itoa(damage))
		damageRune = asRunes[0]
	}
	return &DamageAnimation{
		BaseAnimation: &BaseAnimation{},
		pos:           defenderPos,
		damage:        damage,
		ticksLeft:     4,
		drawables: map[geometry.Point]foundation.TextIcon{
			defenderPos: {
				Rune: damageRune,
				Fg:   fgColor,
				Bg:   bgColor,
			},
		},
	}
}

func (d *DamageAnimation) Cancel() {
	d.ticksLeft = 0
	d.BaseAnimation.Cancel()
}

func (d *DamageAnimation) GetPriority() int {
	return 1
}

func (d *DamageAnimation) GetDrawables() map[geometry.Point]foundation.TextIcon {
	return d.drawables
}

func (d *DamageAnimation) NextFrame() {
	if d.ticksLeft > 0 {
		d.ticksLeft--
	}
	if d.ticksLeft == 0 {
		d.onFinishedOrCancelled()
	}
	if d.ticksLeft == 2 {
		d.drawables[d.pos] = d.drawables[d.pos].Reversed()
	}
}

func (d *DamageAnimation) IsDone() bool {
	return d.ticksLeft <= 0
}

func (d *DamageAnimation) SetDoneCallback(done func()) {
	d.done = done
}
