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

func NewDamageAnimation(defenderPos geometry.Point, playerPos geometry.Point, damage int) *DamageAnimation {
	primary := '!'
	fgColor := color.RGBA{R: 240, G: 20, B: 20, A: 255}
	bgColor := color.RGBA{R: 40, A: 255}
	drawables := make(map[geometry.Point]foundation.TextIcon)
	if damage == 0 {
		primary = '-'
		fgColor = color.RGBA{R: 200, G: 200, B: 200, A: 255}
		bgColor = color.RGBA{R: 40, G: 40, B: 40, A: 255}
	} else if damage < 10 {
		asRunes := []rune(strconv.Itoa(damage))
		primary = asRunes[0]
	} else if damage >= 10 {
		asRunes := []rune(strconv.Itoa(damage))
		primary = asRunes[0]
		secondary := asRunes[1]
		secondaryPos := defenderPos.Add(geometry.Point{X: 1, Y: 0})
		if secondaryPos == playerPos {
			secondaryPos = defenderPos.Add(geometry.Point{X: -1, Y: 0})
			primary = asRunes[1]
			secondary = asRunes[0]
		}
		if damage >= 100 {
			primary = '!'
			secondary = '!'
		}
		drawables[secondaryPos] = foundation.TextIcon{
			Rune: secondary,
			Fg:   fgColor,
			Bg:   bgColor,
		}
	}
	drawables[defenderPos] = foundation.TextIcon{
		Rune: primary,
		Fg:   fgColor,
		Bg:   bgColor,
	}

	return &DamageAnimation{
		BaseAnimation: &BaseAnimation{},
		pos:           defenderPos,
		damage:        damage,
		ticksLeft:     3,
		drawables:     drawables,
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
}

func (d *DamageAnimation) IsDone() bool {
	return d.ticksLeft <= 0
}

func (d *DamageAnimation) SetDoneCallback(done func()) {
	d.done = done
}
