package ui_console

import (
	"contractor/foundation"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"unicode"
)

func (u *UI) ForceUIRedraw() {
	screen := u.application.GetScreen()

	u.application.Lock()
	defer u.application.Unlock()

	u.rightPanel.Draw(screen)
	u.lowerRightPanel.Draw(screen)
	u.messageLabel.Draw(screen)
	u.statusBar.Draw(screen)
}
func (u *UI) updateMapScrollPosition() {
	if !u.isMapScrolling() {
		u.mapScroll = geometry.PointZero
		return
	}
	currentScrollOffset := u.mapScroll
	minDistToBorders := 5

	windowWidth, windowHeight := u.settings.MapWidth, u.settings.MapHeight
	mapSize := u.game.GetMapSize()
	playerPos := u.game.GetPlayerPosition()

	if playerPos.X-currentScrollOffset.X < minDistToBorders {
		u.mapScroll.X = max(0, playerPos.X-minDistToBorders)
	} else if playerPos.X-currentScrollOffset.X > windowWidth-minDistToBorders {
		// Player is too far right, scroll to keep player within border
		u.mapScroll.X = min(mapSize.X-windowWidth, playerPos.X+minDistToBorders-windowWidth)
	}
	if playerPos.Y-currentScrollOffset.Y < minDistToBorders {
		u.mapScroll.Y = max(0, playerPos.Y-minDistToBorders)
	} else if playerPos.Y-currentScrollOffset.Y > windowHeight-minDistToBorders {
		// Player is too far down, scroll to keep player within border
		u.mapScroll.Y = min(mapSize.Y-windowHeight, playerPos.Y+minDistToBorders-windowHeight)
	}
}

func (u *UI) drawMap(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	if !u.game.IsPlayerAndMapInitialized() {
		return x, y, width, height
	}

	defaultMapStyle := u.uiTheme.GetMapDefaultStyle()

	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {

			mapPosX := col - x + u.mapScroll.X
			mapPosY := row - y + u.mapScroll.Y

			mapPos := geometry.Point{X: mapPosX, Y: mapPosY}

			ch, style := u.renderMapPosition(mapPos, u.isAnimationFrame, defaultMapStyle)

			screen.SetContent(col, row, ch, nil, style)
		}
	}
	if u.showCursor {
		screen.ShowCursor(u.game.GetPlayerPosition().X+x, u.game.GetPlayerPosition().Y+y)
	}
	// Space for other content.
	return x, y, width, height
}

func (u *UI) updateLastFrame() {
	// iterate the map and force and update of the last frame
	for y := 0; y < u.settings.MapHeight; y++ {
		for x := 0; x < u.settings.MapWidth; x++ {
			pos := geometry.Point{X: x, Y: y}
			u.renderMapPosition(pos, false, u.uiTheme.GetMapDefaultStyle())
		}
	}
}

func (u *UI) renderMapPosition(mapPos geometry.Point, isAnimationFrame bool, style tcell.Style) (rune, tcell.Style) {
	var ch rune
	var textIcon textiles.TextIcon
	isPositionAnimated := u.animator.isPositionAnimated(mapPos)
	foundIcon := false

	if animIcon, exists := u.animator.animationState[mapPos]; isAnimationFrame && exists {
		foundIcon = true
		if animIcon.HasBackground() {
			textIcon = animIcon
		} else {
			mapIcon, _ := u.mapLookup(mapPos)
			textIcon = animIcon.WithBg(mapIcon.Bg)
		}
	} else if u.mapOverlay.IsSet(mapPos.X, mapPos.Y) {
		textIcon = u.mapOverlay.Get(mapPos.X, mapPos.Y)
		foundIcon = true
	} else {
		textIcon, foundIcon = u.mapLookup(mapPos)
	}

	var fg, bg color.RGBA

	if foundIcon {
		ch, fg, bg = textIcon.Char, textIcon.Fg, textIcon.Bg
		style = style.Attributes(textIcon.Attributes)
	} else {
		ch = ' '
		fg = u.uiTheme.GetUIColor(UIColorUIForeground)
		bg = u.uiTheme.GetUIColor(UIColorUIBackground)
	}

	if bgColor, hasColor := u.tileColors[mapPos]; hasColor {
		existingBG := fxtools.NewColorFromRGBA(bg)
		scaleFactor := 1.0
		brightness := u.lightAt(mapPos).Brightness()
		if brightness < 0.4 {
			scaleFactor = 8.0
		} else if brightness < 0.8 {
			scaleFactor = 4.0
		} else if brightness > 1 {
			scaleFactor = 0.8
		}
		bg = existingBG.Multiply(bgColor.MultiplyWithScalar(scaleFactor)).ToRGBA()
	}

	style = style.Foreground(tcell.NewRGBColor(int32(applyGamma(fg.R, u.gamma)), int32(applyGamma(fg.G, u.gamma)), int32(applyGamma(fg.B, u.gamma))))
	style = style.Background(tcell.NewRGBColor(int32(applyGamma(bg.R, u.gamma)), int32(applyGamma(bg.G, u.gamma)), int32(applyGamma(bg.B, u.gamma))))

	if isAnimationFrame && !isPositionAnimated {
		ch = u.lastFrameIcons[mapPos]
		style = u.lastFrameStyle[mapPos]
	}

	if !isAnimationFrame {
		u.lastFrameStyle[mapPos] = style
		u.lastFrameIcons[mapPos] = ch
	}

	if targetChar, ok := u.targetingTiles[mapPos]; (u.state.IsTargeting()) && ok {
		if mapPos == u.targetPos {
			style = style.Background(tcell.ColorGreen)
		} else {
			ch = targetChar
			style = style.Foreground(tcell.ColorGreen)
		}

	}
	return ch, style
}

func (u *UI) mapLookup(loc geometry.Point) (textiles.TextIcon, bool) {
	if u.game.IsVisibleToPlayer(loc) {
		mapIcon, found := u.visibleLookup(loc)
		return mapIcon, found
	} else if u.game.IsExplored(loc) {
		mapIcon, found := u.exploredLookup(loc)
		mapIcon.Fg = darken(desaturate(mapIcon.Fg))
		mapIcon.Bg = darken(desaturate(mapIcon.Bg))
		return mapIcon, found
	}
	return textiles.TextIcon{}, false
}

func (u *UI) visibleLookup(loc geometry.Point) (textiles.TextIcon, bool) {
	conditionalBackgroundWrapper := func(i textiles.TextIcon) textiles.TextIcon {
		if i.HasBackground() {
			return i
		}
		mapIcon := u.game.MapTileAt(loc)
		return i.WithBg(mapIcon.Bg)
	}
	var icon textiles.TextIcon
	entityType := u.game.PlayerViewAt(loc)
	switch entityType {
	case foundation.EntityTypeActor:
		actor := u.game.ActorAt(loc)
		icon = u.getIconForActor(actor)
	case foundation.EntityTypeDownedActor:
		actor := u.game.DownedActorAt(loc)
		icon = u.getIconForActor(actor)
	case foundation.EntityTypeItem:
		item := u.game.ItemAt(loc)
		icon = conditionalBackgroundWrapper(item.GetIcon())
	case foundation.EntityTypeObject:
		object := u.game.ObjectAt(loc)
		icon = conditionalBackgroundWrapper(object.GetIcon())
	default:
		icon = u.game.MapTileAt(loc)
	}
	fgWithLight, bgWithLight := u.ApplyLighting(loc, icon.Fg, icon.Bg)
	icon.Fg = fgWithLight
	icon.Bg = bgWithLight
	return icon, true
}
func (u *UI) ApplyLighting(p geometry.Point, fg, bg color.RGBA) (color.RGBA, color.RGBA) {
	lightAtCell := u.lightAt(p)
	fgWithLight := applyLightToMaterial(lightAtCell, fg)
	bgWithLight := applyLightToMaterial(lightAtCell, bg)
	return fgWithLight.ToRGBA(), bgWithLight.ToRGBA()
}

func (u *UI) getIconForActor(actor foundation.ActorForUI) textiles.TextIcon {
	mapIconHere := u.game.MapTileAt(actor.Position())

	isHallucinating := u.isPlayerHallucinating()
	if isHallucinating {
		randomLetter := rune('A' + rand.Intn(26))
		if rand.Intn(2) == 0 {
			randomLetter = unicode.ToLower(randomLetter)
		}
		return textiles.TextIcon{
			Char: randomLetter,
			Fg:   u.uiTheme.GetRandomColor(),
			Bg:   mapIconHere.Bg,
		}
	}

	var backGroundColor color.RGBA

	if actor.HasFlag(foundation.FlagHeld) {
		return textiles.TextIcon{
			Char: actor.GetIcon().Char,
			Fg:   u.uiTheme.GetColorByName("Blue_1"),
			Bg:   u.uiTheme.GetColorByName("White"),
		}
	} else {
		backGroundColor = mapIconHere.Bg
	}
	if !actor.IsAlive() {
		return actor.TextIcon(backGroundColor).WithRune('%')
	}

	if actor.HasFlag(foundation.FlagActiveCamouflage) {
		var fgColor color.RGBA
		if u.game.TurnCount()%2 == 0 {
			fgColor = fxtools.LerpColorRGBA(mapIconHere.Bg, mapIconHere.Fg, 0.6)
		} else {
			fgColor = fxtools.LerpColorRGBA(mapIconHere.Bg, u.uiTheme.GetColorByName("White"), 0.1)
		}
		return textiles.TextIcon{
			Char: actor.GetIcon().Char,
			Fg:   fgColor,
			Bg:   backGroundColor,
		}
	}

	return actor.TextIcon(backGroundColor)
}

func (u *UI) getMapTileBackgroundColor(loc geometry.Point) color.RGBA {
	icon := u.game.MapTileAt(loc)
	return applyLightToMaterial(u.lightAt(loc), icon.Bg).ToRGBA()
}

func (u *UI) lightAt(loc geometry.Point) fxtools.HDRColor {
	if u.isAnimationFrame {
		return u.game.LightAt(loc).Add(u.animator.lightAt(loc))
	}
	return u.game.LightAt(loc)
}

func (u *UI) exploredLookup(loc geometry.Point) (textiles.TextIcon, bool) {
	conditionalBackgroundWrapper := func(i textiles.TextIcon) textiles.TextIcon {
		if i.HasBackground() {
			return i
		}
		mapIcon := u.game.MapTileAt(loc)
		return i.WithBg(mapIcon.Bg)
	}
	var icon textiles.TextIcon
	objectAtLoc := u.game.ObjectAt(loc)
	if objectAtLoc != nil {
		icon = conditionalBackgroundWrapper(objectAtLoc.GetIcon())
	} else {
		icon = u.game.MapTileAt(loc)
	}
	fgWithLight, bgWithLight := u.ApplyLighting(loc, icon.Fg, icon.Bg)
	icon.Fg = fgWithLight
	icon.Bg = bgWithLight
	return icon, true
}
