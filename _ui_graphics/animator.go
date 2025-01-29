package ui_graphics

type Animator struct {
}

func (a Animator) CancelAllAnimations() {

}

func (a Animator) Tick() {

}

func NewAnimator() *Animator {
    return &Animator{}
}
