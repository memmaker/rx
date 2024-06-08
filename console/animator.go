package console

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"cmp"
	"slices"
)

type TextAnimation interface {
	GetPriority() int
	GetDrawables() map[geometry.Point]foundation.TextIcon
	NextFrame()
	IsDone() bool
	GetFollowUp() []foundation.Animation
	Cancel()
	IsRequestingMapStateUpdate() bool
	GetAudioCue() string
}
type AudioCuePlayer interface {
	PlayCue(cueName string)
}
type Animator struct {
	animationState    map[geometry.Point]foundation.TextIcon
	runningAnimations []TextAnimation
	audioCuePlayer    AudioCuePlayer
}

func NewAnimator() *Animator {
	return &Animator{
		animationState: make(map[geometry.Point]foundation.TextIcon),
	}
}

func (a *Animator) SetAudioCuePlayer(player AudioCuePlayer) {
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

	for _, animation := range a.runningAnimations {
		for pos, icon := range animation.GetDrawables() {
			a.animationState[pos] = icon
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
