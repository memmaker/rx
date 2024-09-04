package console

import (
	"RogueUI/foundation"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
)

type MuzzleAnimation struct {
	*BaseAnimation
}

func (u *UI) GetAnimMuzzleFlash(position geometry.Point, flashColor fxtools.HDRColor, radius int, bulletCount int, done func()) foundation.Animation {
	color := u.GetAnimBackgroundColor(position, "White", bulletCount*4, done)
	return color
}
