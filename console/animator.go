package console

import (
	"RogueUI/foundation"
	"RogueUI/gridmap"
	"cmp"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"slices"
)

type TextAnimation interface {
	GetPriority() int
	GetDrawables() map[geometry.Point]textiles.TextIcon
	NextFrame()
	IsDone() bool
	GetFollowUp() []foundation.Animation
	Cancel()
	IsRequestingMapStateUpdate() bool
	GetAudioCue() string
	GetLights() []*gridmap.LightSource
}
type Animator struct {
	animationState    map[geometry.Point]textiles.TextIcon
	lightState        map[geometry.Point]fxtools.HDRColor
	runningAnimations []TextAnimation
	audioCuePlayer    foundation.AudioCuePlayer
}

func NewAnimator() *Animator {
	return &Animator{
		animationState: make(map[geometry.Point]textiles.TextIcon),
		lightState:     make(map[geometry.Point]fxtools.HDRColor),
	}
}

func (a *Animator) SetAudioCuePlayer(player foundation.AudioCuePlayer) {
	a.audioCuePlayer = player
}

func (a *Animator) AddAnimation(animation TextAnimation) {
	a.runningAnimations = append(a.runningAnimations, animation)
	a.tryPlayAudioFor(animation)
	slices.SortStableFunc(a.runningAnimations, func(i, j TextAnimation) int {
		return cmp.Compare(i.GetPriority(), j.GetPriority())
	})
}

func (a *Animator) tryPlayAudioFor(animation TextAnimation) {
	if a.audioCuePlayer != nil && animation.GetAudioCue() != "" {
		a.audioCuePlayer.PlayCue(animation.GetAudioCue())
	}
}

func (a *Animator) Tick() (shouldUpdateMapState bool) {
	mapStateNeedsUpdate := false
	for i := len(a.runningAnimations) - 1; i >= 0; i-- {
		currentAnim := a.runningAnimations[i]
		if currentAnim.IsDone() {
			followUp := currentAnim.GetFollowUp()
			if currentAnim.IsRequestingMapStateUpdate() {
				mapStateNeedsUpdate = true
			}
			a.runningAnimations = append(a.runningAnimations[:i], a.runningAnimations[i+1:]...)
			for _, followUpAnim := range followUp {
				if textAnim, isText := followUpAnim.(TextAnimation); isText && textAnim != nil {
					a.runningAnimations = append(a.runningAnimations, textAnim)
					a.tryPlayAudioFor(textAnim)
				}
			}
		}
	}

	slices.SortStableFunc(a.runningAnimations, func(i, j TextAnimation) int {
		return cmp.Compare(i.GetPriority(), j.GetPriority())
	})

	clear(a.animationState)
	clear(a.lightState)

	for _, animation := range a.runningAnimations {
		for pos, icon := range animation.GetDrawables() {
			a.animationState[pos] = icon
		}
		if animation.GetLights() != nil {
			a.updateDynamicLight(animation.GetLights())
		}
		animation.NextFrame()
	}
	return mapStateNeedsUpdate
}

func (a *Animator) CancelAll() {
	for _, animation := range a.runningAnimations {
		cancelRecursive(animation)
	}
	a.runningAnimations = nil
	clear(a.animationState)
}

func (a *Animator) updateDynamicLight(lights []*gridmap.LightSource) {
	for _, light := range lights {
		radius := light.Radius
		for x := -radius; x <= radius; x++ {
			for y := -radius; y <= radius; y++ {
				if x*x+y*y > radius*radius {
					continue
				}
				pos := geometry.Point{X: x, Y: y}.Add(light.Pos)
				existingLight := a.lightAt(pos)
				thisLight := light.Color.MultiplyWithScalar(light.MaxIntensity)
				mixedLight := existingLight.Add(thisLight)
				a.lightState[pos] = mixedLight
			}
		}
	}
}

func (a *Animator) lightAt(pos geometry.Point) fxtools.HDRColor {
	color, exists := a.lightState[pos]
	if !exists {
		return fxtools.HDRColor{A: 1}
	}
	return color
}

func (a *Animator) isPositionAnimated(pos geometry.Point) bool {
	if _, exists := a.animationState[pos]; exists {
		return true
	}
	if _, exists := a.lightState[pos]; exists {
		return true
	}
	return false
}

func cancelRecursive(animation TextAnimation) {
	animation.Cancel()
	for _, followUp := range animation.GetFollowUp() {
		textAnim, isText := followUp.(TextAnimation)
		if !isText {
			continue
		}
		cancelRecursive(textAnim)
	}
}
