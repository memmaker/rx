package ui_graphics

import (
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

func (u *UIController) Update() error {

    if u.shouldQuitGame {
        return ebiten.Termination
    }

    u.worldTicks++
    if u.ticksUntilTooltipAppears > 0 {
        u.ticksUntilTooltipAppears--
        //println("decrementing tooltip ticks", u.ticksUntilTooltipAppears)
    }

    u.handleGlobalColorScaling()

    u.handleMouseInput()

    u.handleKeyboardInput()

    if !u.setToExactFit && u.configuration.KeepPlayerCentered {
        u.mapRenderer.CenterOnFloat(u.gameState.GetPlayerPosition().ToPointF())
    }

    return nil
}

func (u *UIController) handleGlobalColorScaling() {
    if u.keepBlack {
        u.tileRenderer.SetGlobalScaleColor([4]float32{0, 0, 0, 1})
    } else if u.globalFlashColorTicksLeft > 0 {
        //percentage := float64(u.globalFlashColorTicksLeft) / float64(u.globalFlashColorTicksStart)
        //lerpedColor := util.LerpColor(color.White, u.globalFlashColor, percentage)
        u.tileRenderer.SetGlobalScaleColor(colorAsScaledFloat(u.globalScaleColor, 10.0))
        u.globalFlashColorTicksLeft--
    } else if u.globalFadeColorTicksLeft > 0 {

        percentageDone := float32(float64(u.globalFadeColorTicksLeft) / float64(u.globalFadeColorTicksStart)) // from 1 to 0
        if u.fadeFromBlack {
            percentageDone = 1 - percentageDone // from 0 to 1
        }
        u.tileRenderer.SetGlobalScaleColor([4]float32{percentageDone, percentageDone, percentageDone, 1})
        u.globalFadeColorTicksLeft--
        if u.globalFadeColorTicksLeft == 0 {
            if u.fadeFromBlack {
                u.keepBlack = false
            } else {
                u.keepBlack = true
            }
        }
    } else {
        u.tileRenderer.SetGlobalScaleColor([4]float32{1, 1, 1, 1})
    }
}

func colorAsScaledFloat(lerpedColor color.Color, scale float32) [4]float32 {
    r, g, b, a := lerpedColor.RGBA()
    return [4]float32{float32(r) * scale / 65535.0, float32(g) * scale / 65535.0, float32(b) * scale / 65535.0, float32(a) * scale / 65535.0}
}
