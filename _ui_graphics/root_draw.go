package ui_graphics

import (
    "contractor/foundation"
    "contractor/gridmap"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/fxtools"
    "github.com/memmaker/go/geometry"
    "github.com/memmaker/go/textiles"
    "image/color"
)

func (u *UIController) Draw(screen *ebiten.Image) {
    u.tileRenderer.SetRenderTarget(screen)

    u.mapRenderer.Draw(u.worldTicks)

    u.drawActors()

    if !u.IsMouseCapturedByUI() {
        if u.targeter != nil {
            u.targeter.Draw(u.mapRenderer)
        } else {
            // default map selection rectangle
            u.mapRenderer.DrawOnMap(u.mousePosOnMap, u.configuration.TargetingTileIndex, color.White)
        }
    }

    u.drawUI()

    if u.message.ticksLeft > 0 {

        //u.tileRenderer.DrawTTFOnScreen(1, height+2, common.RemoveColorCodes(u.message.GetText()), color.Black)
        u.tileRenderer.DrawTTFOnScreenWithColorCodes(1, u.message.GetHeight(), u.message.GetColorcodedText())
        u.message.ticksLeft--
    }

    if u.currentTooltip != nil && !u.currentTooltip.IsNull() && u.ticksUntilTooltipAppears == 0 {
        u.currentTooltip.Draw()
    }
}

func (u *UIController) drawActors() {
    visibleMapOnScreen := u.mapRenderer.GetVisibleMap()
    state := u.gameState.GameState()
    player := state.GetPlayer()
    playerFieldOfView := player.GetFOV()

    enemyColor := defs.White
    actorAtlas := u.GetCurrentAtlas()

    for _, downedActor := range state.GetMap().DownedActors() {
        if !visibleMapOnScreen.Contains(downedActor.MapPosition()) || !playerFieldOfView.Visible(downedActor.MapPosition()) {
            continue
        }
        if !u.configuration.DrawTilesBehindActors && state.GetMap().IsActorAt(downedActor.MapPosition()) {
            continue
        }
        enemyColor = defs.OffWhite
        screenPos := u.mapRenderer.MapFloatToScreen(downedActor.DrawPosition())
        u.tileRenderer.DrawDefaultScaleTile(float64(screenPos.X), float64(screenPos.Y), actorAtlas, downedActor.Icon(u.worldTicks), u.gameState.GameState().ApplyLighting(downedActor.MapPosition(), enemyColor, u.worldTicks))
    }

    for _, actor := range state.GetMap().Actors() {
        if !visibleMapOnScreen.Contains(actor.MapPosition()) || !playerFieldOfView.Visible(actor.MapPosition()) {
            continue
        }
        if !actor.IsVisible() && !player.HasStatusEffect(common.StatusSeeInvisible) {
            continue
        }

        if u.renderingMode != GraphicRendering {
            enemyColor = actor.GetColor()
        }
        var drawColor color.Color
        drawColor = state.ApplyLighting(actor.MapPosition(), enemyColor, u.worldTicks)

        if actor.IsHurt() {
            percentage := actor.GetHurtTicksAsPercentage() // goes from 1.0 to 0.0
            drawColor = util.LerpColor(defs.Red, enemyColor, percentage)
        }
        if u.configuration.ActorBobbingEnabled {
            u.drawActorWithBobbingAnimation(actor, actorAtlas, drawColor)
        } else {
            screenPos := u.mapRenderer.MapFloatToScreen(actor.DrawPosition())
            flippedX := actor.FlipX() && u.configuration.FlipActorsHorizontally
            icon := u.getActorIcon(actor, u.worldTicks)
            u.tileRenderer.DrawDefaultScaleTileWithFlip(float64(screenPos.X), float64(screenPos.Y), actorAtlas, icon, drawColor, flippedX)
        }
    }

    for _, actor := range state.GetMap().Actors() {
        if !visibleMapOnScreen.Contains(actor.MapPosition()) || !playerFieldOfView.Visible(actor.MapPosition()) {
            continue
        }
        if !actor.GetStats().GetResource(stats.Health).IsAtMax() {
            u.drawHealthBar(actor)
        }
    }
}

func (u *UIController) drawActorWithBobbingAnimation(actor *common.Actor, atlas TextureAtlas, drawColor color.Color) {
    screenPos := u.mapRenderer.MapFloatToScreen(actor.DrawPosition())
    actorTickOffset := actor.MapPosition().X + actor.MapPosition().Y + len(actor.Name())
    tileSize := u.mapRenderer.GetScaledTileSize()
    ticksForAnimation := u.worldTicks + uint64(actorTickOffset)
    percentage := GetPercentageFromTick(ticksForAnimation, 1) // -> [0.0..1.0)  but we want [0.0..1.0..0.0]
    percentage *= 2
    if percentage > 1 {
        percentage = 1 - percentage
    }
    //u.tileRenderer.DrawDefaultScaleTile(float64(screenPos.X), float64(screenPos.Y), u.worldTexture, actor.Icon(u.worldTicks), color.White)
    scaleUp := geometry.PointF{X: 1, Y: 1}
    scaleDown := geometry.PointF{X: 1.05, Y: 0.95}
    currentScale := LerpPointF(scaleUp, scaleDown, percentage)
    yOffset := float64(tileSize.Y) * (1 - currentScale.Y)
    xOffset := (float64(tileSize.X) * (1 - currentScale.X)) / 2

    icon := u.getActorIcon(actor, ticksForAnimation)

    flippedX := actor.FlipX() && u.configuration.FlipActorsHorizontally
    u.tileRenderer.DrawScaledTileWithFlip(float64(screenPos.X)+xOffset, float64(screenPos.Y)+yOffset, atlas, icon, currentScale, drawColor, flippedX)
}

func (u *UIController) getActorIcon(actor foundation.ActorForUI, ticksForAnimation uint64) int32 {
    return int32(fxtools.UnicodeToCP437Byte(actor.GetIcon().Char))
}

func CodePage437IconAndColor(t textiles.TextIcon) (int32, color.RGBA) {
    return int32(fxtools.UnicodeToCP437Byte(t.Char)), t.Fg
}
