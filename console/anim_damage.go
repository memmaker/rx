package console

import (
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"strconv"
)

type DamageAnimation struct {
	*BaseAnimation
	pos              geometry.Point
	damage           int
	drawables        map[geometry.Point]textiles.TextIcon
	ticksLeft        int
	makeBloody       func(mapPos geometry.Point)
	bloodPerTick     int
	secondaryDrawPos geometry.Point
	bloodColors      []color.RGBA
}

func NewDamageAnimation(blood func(mapPos geometry.Point), defenderPos geometry.Point, playerPos geometry.Point, damage int, bulletMultiplier int, bloodColors []color.RGBA) *DamageAnimation {
	primary := '!'
	fgColor := color.RGBA{R: 240, G: 20, B: 20, A: 255}
	bgColor := color.RGBA{R: 40, A: 255}
	var secondaryDrawPos geometry.Point
	drawables := make(map[geometry.Point]textiles.TextIcon)
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
		drawables[secondaryPos] = textiles.TextIcon{
			Char: secondary,
			Fg:   fgColor,
			Bg:   bgColor,
		}
		secondaryDrawPos = secondaryPos
	}
	drawables[defenderPos] = textiles.TextIcon{
		Char: primary,
		Fg:   fgColor,
		Bg:   bgColor,
	}

	// for every 10 points of damage we want 1 blood call
	// spread around the 3 ticks we got
	// damage between 1 - 300
	// bullet multi: 1 or 3
	bloodAmount := damage / 5                                // 1-60
	bloodPerTick := max(1, bloodAmount/(6*bulletMultiplier)) // 1-10

	return &DamageAnimation{
		BaseAnimation:    &BaseAnimation{},
		pos:              defenderPos,
		damage:           damage,
		drawables:        drawables,
		ticksLeft:        3 * bulletMultiplier,
		makeBloody:       blood,
		bloodPerTick:     bloodPerTick,
		secondaryDrawPos: secondaryDrawPos,
		bloodColors:      bloodColors,
	}
}

func (d *DamageAnimation) Cancel() {
	d.ticksLeft = 0
	d.BaseAnimation.Cancel()
}

func (d *DamageAnimation) GetPriority() int {
	return 1
}

func (d *DamageAnimation) GetDrawables() map[geometry.Point]textiles.TextIcon {
	drawables := d.drawables
	neigh := geometry.Neighbors{}
	neighbors := neigh.All(d.pos, func(point geometry.Point) bool {
		if d.damage < 10 {
			return true
		}
		return point != d.secondaryDrawPos
	})

	for i := 0; i < d.bloodPerTick; i++ {
		randomNeighbor := neighbors[rand.Intn(len(neighbors))]
		icon := textiles.TextIcon{
			Char: []rune{'.', ',', ';', ':'}[rand.Intn(4)],
			Fg:   d.bloodColors[rand.Intn(len(d.bloodColors))],
		}
		if rand.Intn(4) == 0 {
			icon.Bg = d.bloodColors[rand.Intn(len(d.bloodColors))]
		}
		drawables[randomNeighbor] = icon
	}
	return drawables
}

func (d *DamageAnimation) NextFrame() {
	if d.ticksLeft > 0 {
		d.ticksLeft--
	}
	for i := 0; i < d.bloodPerTick; i++ {
		d.makeBloody(d.pos)
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
