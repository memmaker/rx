package console

import (
	"RogueUI/gridmap"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
)

type LaserAnimation struct {
	*BaseAnimation
	drawables  map[geometry.Point]textiles.TextIcon
	lights     []*gridmap.LightSource
	framesLeft int
	lightColor fxtools.HDRColor
}

func NewLaserAnimation(path []geometry.Point, laserColor fxtools.HDRColor, done func()) *LaserAnimation {
	drawables := make(map[geometry.Point]textiles.TextIcon)
	lights := make([]*gridmap.LightSource, len(path))
	var directionOfLine geometry.Point
	var lineIcon textiles.TextIcon
	lineIcon.Fg = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for i, pos := range path {
		if i == 0 {
			directionOfLine = path[i+1].Sub(path[i])
		} else {
			directionOfLine = path[i].Sub(path[i-1])
		}
		lineIcon.Char = charFromDirection(directionOfLine)
		drawables[pos] = lineIcon

		// Add light source
		light := gridmap.LightSource{
			Pos:          pos,
			Radius:       1,
			Color:        laserColor,
			MaxIntensity: 1,
		}
		lights[i] = &light
	}

	return &LaserAnimation{
		BaseAnimation: &BaseAnimation{
			done: done,
		},
		drawables:  drawables,
		lights:     lights,
		lightColor: laserColor,
		framesLeft: 5,
	}
}

func charFromDirection(lineDir geometry.Point) rune {
	if lineDir.X == 0 && (lineDir.Y == 1 || lineDir.Y == -1) { // vertical
		return '|'
	} else if (lineDir.X == 1 || lineDir.X == -1) && lineDir.Y == 0 { // horizontal
		return '-'
	} else if (lineDir.X == 1 && lineDir.Y == 1) || (lineDir.X == -1 && lineDir.Y == -1) { // diagonal
		return '\\'
	} else if (lineDir.X == 1 && lineDir.Y == -1) || (lineDir.X == -1 && lineDir.Y == 1) { // diagonal
		return '/'
	}
	return ' '
}

func (p *LaserAnimation) isLowLightBeam() bool {
	return p.framesLeft == 5 || p.framesLeft == 1
}

func (p *LaserAnimation) isHiLightBeam() bool {
	return p.framesLeft == 4 || p.framesLeft == 2
}

func (p *LaserAnimation) isVeryBrightBeam() bool {
	return p.framesLeft == 3
}

func (p *LaserAnimation) GetLights() []*gridmap.LightSource {
	return fxtools.MapSlice(p.lights, func(light *gridmap.LightSource) *gridmap.LightSource {
		if p.isLowLightBeam() {
			light.Color = p.lightColor.MultiplyWithScalar(0.5)
		} else if p.isHiLightBeam() {
			light.Color = p.lightColor
		} else if p.isVeryBrightBeam() {
			light.Color = fxtools.HDRColor{1, 1, 1, 1}.MultiplyWithScalar(3)
		}
		return light
	})
}

func (p *LaserAnimation) GetPriority() int {
	return 1
}

func (p *LaserAnimation) GetDrawables() map[geometry.Point]textiles.TextIcon {
	return p.drawables
}

func (p *LaserAnimation) NextFrame() {
	if p.IsDone() {
		return
	}
	// next path index
	p.framesLeft = p.framesLeft - 1
	if p.framesLeft == 0 {
		p.onFinishedOrCancelled()
		return
	}
}

func (p *LaserAnimation) IsDone() bool {
	return p.finishedOrCancelled
}
