package ui_graphics

import (
    "contractor/gridmap"
    "github.com/memmaker/go/geometry"
    "image/color"
)

type AoEResolver struct {
    GetHitPositions func(origin, target geometry.Point) []geometry.Point
    Description     []LabelText
}
type TargetRulesProvider struct {
    GetTargetRules func(attacker *Actor, hitPositions *AoEResolver, currentMap *gridmap.GridMap[*Actor, *Item, Object]) TargetRules
    Description    []LabelText
}

type TargetRules interface {
    OnTargetChanged(target geometry.Point) TargetInfo
    GetPreSelectedTarget() geometry.Point
    GetAttacker() *Actor
}
type TargetInfo struct {
    AdjustedTarget geometry.Point
    Highlights     []TargetHighlight
    Hitpositions   []geometry.Point
    IsValid        bool
    InfoText       string
}
type TargetHighlight struct {
    Pos   geometry.Point
    Color color.Color
}
type Targeter struct {
    rules        TargetRules
    hitEffect    *common.HitEffect
    isDone       bool
    uiController *UIController
    targetInfo   TargetInfo
}

func (t *Targeter) IsDone() bool {
    return t.isDone
}

func (t *Targeter) OnCommand(command CommandType) bool {
    switch command {
    case PlayerCommandCancel:
        t.isDone = true
        return true
    case PlayerCommandConfirm:
        if t.HasValidTarget() {
            t.Activate()
        } else {
            t.isDone = true
            return true
        }
    case PlayerCommandUp:
        nextTargetPos := t.targetInfo.AdjustedTarget.Add(geometry.Point{Y: -1})
        t.updateTarget(nextTargetPos)
        return true
    case PlayerCommandDown:
        nextTargetPos := t.targetInfo.AdjustedTarget.Add(geometry.Point{Y: 1})
        t.updateTarget(nextTargetPos)
        return true
    case PlayerCommandLeft:
        nextTargetPos := t.targetInfo.AdjustedTarget.Add(geometry.Point{X: -1})
        t.updateTarget(nextTargetPos)
        return true
    case PlayerCommandRight:
        nextTargetPos := t.targetInfo.AdjustedTarget.Add(geometry.Point{X: 1})
        t.updateTarget(nextTargetPos)
        return true
    case PlayerCommandUpRight:
        nextTargetPos := t.targetInfo.AdjustedTarget.Add(geometry.Point{X: 1, Y: -1})
        t.updateTarget(nextTargetPos)
        return true
    case PlayerCommandUpLeft:
        nextTargetPos := t.targetInfo.AdjustedTarget.Add(geometry.Point{X: -1, Y: -1})
        t.updateTarget(nextTargetPos)
        return true
    case PlayerCommandDownRight:
        nextTargetPos := t.targetInfo.AdjustedTarget.Add(geometry.Point{X: 1, Y: 1})
        t.updateTarget(nextTargetPos)
        return true
    case PlayerCommandDownLeft:
        nextTargetPos := t.targetInfo.AdjustedTarget.Add(geometry.Point{X: -1, Y: 1})
        t.updateTarget(nextTargetPos)
        return true

    }
    return false
}

func (t *Targeter) OnMouseClicked(mapPos geometry.Point) bool {
    if t.IsDone() {
        return false
    }
    t.updateTarget(mapPos)
    return t.Activate()
}

func (t *Targeter) OnMouseMoved(mapPos geometry.Point) (bool, Tooltip) { // THIS IS THE PROBLEM
    if mapPos == t.targetInfo.AdjustedTarget {
        return false, EmptyTooltip{}
    }
    t.updateTarget(mapPos)
    return true, EmptyTooltip{}
}

func (t *Targeter) Draw(renderer *MapRenderer) {
    if t.IsDone() {
        return
    }
    for _, highlight := range t.targetInfo.Highlights {
        if highlight.Pos == t.targetInfo.AdjustedTarget && t.targetInfo.InfoText != "" {
            screenOffset := geometry.Point{X: 0, Y: 0}
            renderer.DrawStringOnMapWithOffset(t.targetInfo.AdjustedTarget, screenOffset, t.targetInfo.InfoText, defs.OffWhite)
        }
        renderer.DrawOnMap(highlight.Pos, defs.IconTargeting, highlight.Color)
    }
}

func (t *Targeter) updateTarget(pos geometry.Point) {
    t.targetInfo = t.rules.OnTargetChanged(pos)
}

func (t *Targeter) HasValidTarget() bool {
    return t.targetInfo.IsValid
}

func (t *Targeter) Activate() bool {
    if t.HasValidTarget() {
        t.hitEffect.Activate(t.rules.GetAttacker(), t.targetInfo.AdjustedTarget, t.targetInfo.Hitpositions)
        t.isDone = true
        return true
    }
    return false
}

func NewTargeter(uiController *UIController, rules TargetRules, onTargetSelected *common.HitEffect) *Targeter {
    t := &Targeter{
        rules:        rules,
        uiController: uiController,
        hitEffect:    onTargetSelected,
        targetInfo:   rules.OnTargetChanged(rules.GetPreSelectedTarget()),
    }
    t.updateTarget(rules.GetPreSelectedTarget())
    return t
}
