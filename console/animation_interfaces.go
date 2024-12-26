package console

import (
	"contractor/foundation"
	"contractor/gridmap"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
)

func (u *UI) GetAnimMove(actor foundation.ActorForUI, old geometry.Point, new geometry.Point) foundation.Animation {
	if u.settings.AnimationsEnabled && u.settings.AnimateMovement {
		return NewMovementAnimation(u.getIconForActor(actor), old, new, u.uiTheme.GetColorByName, nil)
	}
	return nil
}

func (u *UI) GetAnimQuickMove(actor foundation.ActorForUI, path []geometry.Point) foundation.Animation {
	if u.settings.AnimationsEnabled {
		animation := NewMovementAnimation(u.getIconForActor(actor), actor.Position(), path[len(path)-1], u.uiTheme.GetColorByName, nil)
		animation.EnableQuickMoveMode(path)
		animation.SetMapLookup(u.game.MapTileAt)
		return animation
	}
	return nil
}

func (u *UI) GetAnimMuzzleFlash(position geometry.Point, flashColor fxtools.HDRColor, radius int, bulletCount int, done func()) foundation.Animation {
	return NewMuzzleAnimation(position, flashColor, radius, bulletCount, done)
}

func (u *UI) GetAnimLaser(path []geometry.Point, lightColor fxtools.HDRColor, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}
	return NewLaserAnimation(path, lightColor, done)
}

func (u *UI) GetAnimBackgroundColor(position geometry.Point, colorName string, frameCount int, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}
	iconAtLocation, _ := u.mapLookup(position)
	bgColor := u.uiTheme.GetColorByName(colorName)
	return NewCoverAnimation(position, iconAtLocation.WithBg(bgColor), frameCount, done)
}

func (u *UI) GetAnimCover(loc geometry.Point, icon textiles.TextIcon, turns int, done func()) foundation.Animation {
	if u.settings.AnimationsEnabled && u.settings.AnimateMovement {
		return NewCoverAnimation(loc, icon, turns, done)
	}
	return nil
}

func (u *UI) GetAnimAttack(attacker, defender foundation.ActorForUI) foundation.Animation {
	return nil
}

func (u *UI) GetAnimDamage(spreadBlood func(mapPos geometry.Point), defenderPos geometry.Point, damage int, bullets int) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateDamage {
		return nil
	}
	bloodColors := []color.RGBA{
		u.uiTheme.palette.Get("red_5"),
		u.uiTheme.palette.Get("red_6"),
		u.uiTheme.palette.Get("red_7"),
		u.uiTheme.palette.Get("red_8"),
		u.uiTheme.palette.Get("red_9"),
		u.uiTheme.palette.Get("red_10"),
		u.uiTheme.palette.Get("red_11"),
		u.uiTheme.palette.Get("red_12"),
		u.uiTheme.palette.Get("red_13"),
		u.uiTheme.palette.Get("red_14"),
		u.uiTheme.palette.Get("red_15"),
	}
	animation := NewDamageAnimation(spreadBlood, defenderPos, u.game.GetPlayerPosition(), damage, bullets, bloodColors)
	return animation
}
func (u *UI) GetAnimTiles(positions []geometry.Point, frames []textiles.TextIcon, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}
	return NewTilesAnimation(positions, frames, done)
}

func (u *UI) GetAnimRadialReveal(position geometry.Point, dijkstra map[geometry.Point]int, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}

	animation := NewRadialAnimation(position, dijkstra, u.uiTheme.GetColorByName, u.mapLookup, done)
	animation.SetKeepDrawingCoveredGround(true)
	animation.SetUseIconColors(false)
	return animation
}

func (u *UI) GetAnimRadialAlert(position geometry.Point, dijkstra map[geometry.Point]int, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateEffects {
		return nil
	}
	lookup := func(loc geometry.Point) (textiles.TextIcon, bool) {
		return textiles.TextIcon{
			Char: '‼',
			Fg:   u.uiTheme.GetColorByName("Black"),
			Bg:   u.uiTheme.GetColorByName("Red_1"),
		}, true
	}
	animation := NewRadialAnimation(position, dijkstra, u.uiTheme.GetColorByName, lookup, done)
	animation.SetUseIconColors(true)
	return animation
}

func (u *UI) GetAnimTeleport(user foundation.ActorForUI, origin, targetPos geometry.Point, appearOnMap func()) (foundation.Animation, foundation.Animation) {
	originalIcon := u.getIconForActor(user)
	mapBackground := u.getMapTileBackgroundColor(origin)
	lightCyan := u.uiTheme.GetColorByName("LightCyan")
	white := u.uiTheme.GetColorByName("White")
	lightGray := u.uiTheme.GetColorByName("light_gray_5")
	vanishAnim := u.GetAnimTiles([]geometry.Point{origin}, []textiles.TextIcon{
		originalIcon.WithFg(white),
		originalIcon.WithFg(white),
		originalIcon.WithFg(lightCyan),
		{Char: '*', Fg: lightCyan, Bg: mapBackground},
		{Char: '*', Fg: lightCyan, Bg: mapBackground},
		{Char: '+', Fg: lightCyan, Bg: mapBackground},
		{Char: '+', Fg: lightCyan, Bg: mapBackground},
		{Char: '|', Fg: lightCyan, Bg: mapBackground},
		{Char: '|', Fg: lightCyan, Bg: mapBackground},
		{Char: '∙', Fg: lightCyan, Bg: mapBackground},
		{Char: '.', Fg: lightCyan, Bg: mapBackground},
		{Char: '.', Fg: lightGray, Bg: mapBackground},
		{Char: '.', Fg: u.uiTheme.GetColorByName("dark_gray_3"), Bg: mapBackground},
	}, nil)
	vanishAnim.RequestMapUpdateOnFinish()

	appearAnim := u.GetAnimAppearance(user, targetPos, appearOnMap)
	vanishAnim.SetFollowUp([]foundation.Animation{appearAnim})
	return vanishAnim, appearAnim
}

func (u *UI) GetAnimAppearance(actor foundation.ActorForUI, targetPos geometry.Point, done func()) foundation.Animation {
	originalIcon := u.getIconForActor(actor)
	mapBackground := u.getMapTileBackgroundColor(targetPos)
	lightCyan := u.uiTheme.GetColorByName("LightCyan")
	white := u.uiTheme.GetColorByName("White")
	lightGray := u.uiTheme.GetColorByName("light_gray_5")
	appearAnim := u.GetAnimTiles([]geometry.Point{targetPos}, []textiles.TextIcon{
		{Char: '.', Fg: u.uiTheme.GetColorByName("dark_gray_3"), Bg: mapBackground},
		{Char: '.', Fg: lightGray, Bg: mapBackground},
		{Char: '.', Fg: lightGray, Bg: mapBackground},
		{Char: '.', Fg: lightCyan, Bg: mapBackground},
		{Char: '∙', Fg: lightCyan, Bg: mapBackground},
		{Char: '|', Fg: lightCyan, Bg: mapBackground},
		{Char: '|', Fg: lightCyan, Bg: mapBackground},
		{Char: '+', Fg: lightCyan, Bg: mapBackground},
		{Char: '+', Fg: lightCyan, Bg: mapBackground},
		{Char: '*', Fg: lightCyan, Bg: mapBackground},
		{Char: '*', Fg: lightCyan, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: white, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: white, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: lightCyan, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: lightCyan, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: white, Bg: mapBackground},
		{Char: originalIcon.Char, Fg: white, Bg: mapBackground},
	}, done)
	return appearAnim
}

func (u *UI) GetAnimEvade(defender foundation.ActorForUI, done func()) foundation.Animation {
	actorIcon := u.getIconForActor(defender)
	return u.GetAnimTiles([]geometry.Point{defender.Position()}, []textiles.TextIcon{
		actorIcon.WithItalic(),
		actorIcon.WithItalic().WithBold(),
		actorIcon.WithItalic(),
	},
		done)
}

func (u *UI) GetAnimWakeUp(location geometry.Point, done func()) foundation.Animation {
	keepAllNeighbors := func(point geometry.Point) bool { return true }

	neigh := geometry.Neighbors{}
	cardinalNeighbors := neigh.Cardinal(location, keepAllNeighbors)
	diagonalNeighbors := neigh.Diagonal(location, keepAllNeighbors)

	wakeUpRunes := []rune("????")
	yellow := u.uiTheme.GetColorByName("Yellow_1")
	var prevAnim foundation.Animation
	var rootAnim foundation.Animation
	runeCount := len(wakeUpRunes)
	for i := 0; i < runeCount; i++ {

		cycleIcon := textiles.TextIcon{
			Char: wakeUpRunes[i],
			Fg:   u.uiTheme.GetUIColor(UIColorUIForeground),
			Bg:   u.uiTheme.GetColorByName("black"),
		}

		frames := []textiles.TextIcon{
			cycleIcon.WithFg(yellow),
			cycleIcon.WithFg(yellow),
		}

		var neighbors []geometry.Point
		if i%2 == 0 {
			neighbors = cardinalNeighbors
		} else {
			neighbors = diagonalNeighbors
		}
		var doneCall func()
		if i == runeCount-1 {
			doneCall = done
		}
		anim := u.GetAnimTiles(neighbors, frames, doneCall)
		if rootAnim == nil {
			rootAnim = anim
		}

		if prevAnim != nil {
			prevAnim.SetFollowUp([]foundation.Animation{anim})
		}

		prevAnim = anim
	}
	return rootAnim
}
func (u *UI) GetAnimConfuse(location geometry.Point, done func()) foundation.Animation {
	keepAllNeighbors := func(point geometry.Point) bool { return true }

	neigh := geometry.Neighbors{}
	cardinalNeighbors := neigh.Cardinal(location, keepAllNeighbors)
	diagonalNeighbors := neigh.Diagonal(location, keepAllNeighbors)

	confuseRune := []rune("?¿¡!")
	randomRune := func() rune {
		return confuseRune[rand.Intn(len(confuseRune))]
	}
	confuseColors := []color.RGBA{u.uiTheme.GetColorByName("LightMagenta"), u.uiTheme.GetColorByName("LightRed"), u.uiTheme.GetColorByName("Yellow_1"), u.uiTheme.GetColorByName("LightGreen"), u.uiTheme.GetColorByName("light_blue_3")}
	randomColor := func() color.RGBA {
		return confuseColors[rand.Intn(len(confuseColors))]
	}
	cycleCount := 4

	var prevAnim foundation.Animation
	var rootAnim foundation.Animation
	for i := 0; i < cycleCount; i++ {

		cycleIcon := textiles.TextIcon{
			Char: randomRune(),
			Fg:   u.uiTheme.GetUIColor(UIColorUIForeground),
			Bg:   u.uiTheme.GetUIColor(UIColorUIBackground),
		}

		frames := []textiles.TextIcon{
			cycleIcon.WithFg(randomColor()),
			cycleIcon.WithFg(randomColor()),
			cycleIcon.WithFg(randomColor()),
			cycleIcon.WithFg(randomColor()),
		}

		var neighbors []geometry.Point
		if i%2 == 0 {
			neighbors = cardinalNeighbors
		} else {
			neighbors = diagonalNeighbors
		}
		var doneCall func()
		if i == cycleCount-1 {
			doneCall = done
		}
		anim := u.GetAnimTiles(neighbors, frames, doneCall)
		if rootAnim == nil {
			rootAnim = anim
		}

		if prevAnim != nil {
			prevAnim.SetFollowUp([]foundation.Animation{anim})
		}

		prevAnim = anim
	}
	return rootAnim
}
func (u *UI) GetAnimBreath(path []geometry.Point, done func()) []foundation.Animation {
	var randomAroundPath []geometry.Point
	dest := path[len(path)-1]
	prevToLast := dest
	if len(path) > 1 {
		prevToLast = path[len(path)-2]
	}
	neigh := geometry.Neighbors{}
	aroundDest := neigh.Cardinal(dest, func(point geometry.Point) bool {
		return point != prevToLast
	})
	pathWithoutDest := path[:len(path)-1]
	aroundDest = append(aroundDest, dest)

	for _, point := range path {
		randomAroundPath = append(randomAroundPath, point.Add(geometry.Point{X: rand.Intn(3) - 1, Y: rand.Intn(3) - 1}))
		randomAroundPath = append(randomAroundPath, point.Add(geometry.Point{X: rand.Intn(3) - 1, Y: rand.Intn(3) - 1}))
	}

	smokeAnim := u.GetAnimTiles(randomAroundPath, []textiles.TextIcon{
		{Char: ' ', Fg: u.uiTheme.GetColorByName("Black"), Bg: u.uiTheme.GetColorByName("Black")},
		{Char: ' ', Fg: u.uiTheme.GetColorByName("Black"), Bg: u.uiTheme.GetColorByName("Black")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("dark_gray_3"), Bg: u.uiTheme.GetColorByName("Black")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("dark_gray_3")},
		{Char: '.', Fg: u.uiTheme.GetColorByName("light_gray_5")},
		{Char: '.', Fg: u.uiTheme.GetColorByName("light_gray_5")},
	}, nil)

	projAnim := NewTilesAnimation(pathWithoutDest, []textiles.TextIcon{
		{Char: '.', Fg: u.uiTheme.GetColorByName("White"), Bg: u.uiTheme.GetColorByName("Yellow_2")},
		{Char: '∙', Fg: u.uiTheme.GetColorByName("White"), Bg: u.uiTheme.GetColorByName("Yellow_2")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("Orange_4"), Bg: u.uiTheme.GetColorByName("Yellow_1")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("Yellow"), Bg: u.uiTheme.GetColorByName("Orange_4")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("Red"), Bg: u.uiTheme.GetColorByName("Orange_6")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("light_gray_5"), Bg: u.uiTheme.GetColorByName("Orange_8")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("dark_gray_3"), Bg: u.uiTheme.GetColorByName("Black")},
	}, done)
	projAnim.SetLightsOnAllTiles(&gridmap.LightSource{
		Pos:          geometry.Point{},
		Radius:       2,
		Color:        fxtools.NewColorFromRGBA(u.uiTheme.GetColorByName("Orange_4")),
		MaxIntensity: 0.9,
	})
	projAnim.SetFollowUp([]foundation.Animation{smokeAnim})

	ballAnim := NewTilesAnimation(aroundDest, []textiles.TextIcon{
		{Char: '*', Fg: u.uiTheme.GetColorByName("Yellow_2"), Bg: u.uiTheme.GetColorByName("Yellow_1")},
		{Char: '∙', Fg: u.uiTheme.GetColorByName("Yellow_2"), Bg: u.uiTheme.GetColorByName("Yellow_1")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("Yellow_1"), Bg: u.uiTheme.GetColorByName("Orange_4")},
		{Char: '∙', Fg: u.uiTheme.GetColorByName("Yellow"), Bg: u.uiTheme.GetColorByName("Orange_6")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("Red"), Bg: u.uiTheme.GetColorByName("Orange_8")},
		{Char: '∙', Fg: u.uiTheme.GetColorByName("Orange_6"), Bg: u.uiTheme.GetColorByName("Yellow_1")},
		{Char: '*', Fg: u.uiTheme.GetColorByName("Black"), Bg: u.uiTheme.GetColorByName("Yellow_1")},
	}, nil)
	ballAnim.SetLightsOnAllTiles(&gridmap.LightSource{
		Pos:          geometry.Point{},
		Radius:       2,
		Color:        fxtools.NewColorFromRGBA(u.uiTheme.GetColorByName("Orange_4")),
		MaxIntensity: 0.9,
	})
	return []foundation.Animation{ballAnim, projAnim}
}
func (u *UI) GetAnimVorpalizeWeapon(origin geometry.Point, done func()) []foundation.Animation {
	effectIcon := textiles.TextIcon{
		Char: '+',
		Fg:   u.uiTheme.GetColorByName("White"),
		Bg:   u.uiTheme.GetColorByName("Black"),
	}
	outmostPositions := geometry.CircleAround(origin, 2)
	outerPositions := geometry.CircleAround(origin, 1)

	animationInner := u.GetAnimTiles([]geometry.Point{origin}, []textiles.TextIcon{
		effectIcon.WithBg(u.uiTheme.GetColorByName("White")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("White")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("Black")),
	}, done)
	animationCenter := u.GetAnimTiles(outerPositions, []textiles.TextIcon{
		effectIcon.WithBg(u.uiTheme.GetColorByName("White")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("Black")),
	}, nil)

	animationOuter := u.GetAnimTiles(outmostPositions, []textiles.TextIcon{
		effectIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		effectIcon.WithBg(u.uiTheme.GetColorByName("Black")).WithFg(u.uiTheme.GetColorByName("Black")),
	}, nil)

	return []foundation.Animation{animationInner, animationCenter, animationOuter}
}
func (u *UI) GetAnimEnchantWeapon(player foundation.ActorForUI, location geometry.Point, done func()) foundation.Animation {
	playerIcon := u.getIconForActor(player)
	frames := []textiles.TextIcon{
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("LightCyan")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("LightCyan")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("LightCyan")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_blue_3")).WithFg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue_1")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("Blue")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
	}
	return u.GetAnimTiles([]geometry.Point{location}, frames, done)
}
func (u *UI) GetAnimEnchantArmor(player foundation.ActorForUI, location geometry.Point, done func()) foundation.Animation {
	playerIcon := u.getIconForActor(player)
	frames := []textiles.TextIcon{
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")).WithFg(u.uiTheme.GetColorByName("Black")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")).WithFg(u.uiTheme.GetColorByName("White")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")).WithFg(u.uiTheme.GetColorByName("White")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")).WithFg(u.uiTheme.GetColorByName("White")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("light_gray_5")).WithFg(u.uiTheme.GetColorByName("White")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
		playerIcon.WithBg(u.uiTheme.GetColorByName("dark_gray_3")).WithFg(u.uiTheme.GetColorByName("light_gray_5")),
	}

	return u.GetAnimTiles([]geometry.Point{location}, frames, done)
}
func (u *UI) GetAnimThrow(item foundation.Item, origin geometry.Point, target geometry.Point) (foundation.Animation, int) {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateProjectiles {
		return nil, 0
	}
	textIcon := item.GetIcon()

	return u.GetAnimProjectileWithIcon(textIcon, origin, target, nil)
}

func (u *UI) GetAnimProjectile(icon rune, fgColor string, origin geometry.Point, target geometry.Point, done func()) (foundation.Animation, int) {
	textIcon := textiles.TextIcon{
		Char: icon,
		Fg:   u.uiTheme.GetColorByName(fgColor),
	}
	return u.GetAnimProjectileWithIcon(textIcon, origin, target, done)
}
func (u *UI) GetAnimProjectileWithIcon(textIcon textiles.TextIcon, origin geometry.Point, target geometry.Point, done func()) (foundation.Animation, int) {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateProjectiles {
		return nil, 0
	}
	pathOfFlight := geometry.BresenhamLine(origin, target, func(x, y int) bool {
		return true
	})

	if len(pathOfFlight) == 0 {
		return nil, 0
	}

	return NewProjectileAnimation(pathOfFlight, textIcon, u.mapLookup, done), len(pathOfFlight)
}

func (u *UI) GetAnimProjectileWithTrail(leadIcon rune, colorNames []string, pathOfFlight []geometry.Point, done func()) (foundation.Animation, int) {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateProjectiles {
		return nil, 0
	}

	if len(pathOfFlight) == 0 {
		return nil, 0
	}

	var trailIcons []textiles.TextIcon

	for i, cName := range colorNames {
		if i == 0 {
			trailIcons = append(trailIcons, textiles.TextIcon{
				Char: leadIcon,
				Fg:   u.uiTheme.GetColorByName(cName),
			})
		} else {
			trailIcons = append(trailIcons, textiles.TextIcon{
				Char: '█',
				Fg:   u.uiTheme.GetColorByName(cName),
			})
		}
	}

	animation := NewProjectileAnimation(pathOfFlight, trailIcons[0], u.mapLookup, done)
	animation.SetTrail(trailIcons[1:])
	return animation, len(pathOfFlight)
}

func (u *UI) GetAnimProjectileWithLight(leadIcon rune, lightColorName string, pathOfFlight []geometry.Point, done func()) (foundation.Animation, int) {
	if !u.settings.AnimationsEnabled || !u.settings.AnimateProjectiles {
		return nil, 0
	}

	if len(pathOfFlight) == 0 {
		return nil, 0
	}

	lightColor := u.uiTheme.GetColorByName(lightColorName)
	leadTextIcon := textiles.TextIcon{
		Char: leadIcon,
		Fg:   lightColor,
	}

	animation := NewProjectileAnimation(pathOfFlight, leadTextIcon, u.mapLookup, done)
	animation.AddLightSource(&gridmap.LightSource{
		Pos:          pathOfFlight[0],
		Radius:       1,
		Color:        fxtools.NewColorFromRGBA(lightColor),
		MaxIntensity: 3,
	})
	return animation, len(pathOfFlight)
}

func (u *UI) GetAnimRadialExplosion(hitPositions map[geometry.Point]int, lightColor fxtools.HDRColor, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled {
		return nil
	}

	return NewRadialExplosionAnimation(hitPositions, lightColor, done)
}

func (u *UI) GetAnimExplosion(hitPositions []geometry.Point, done func()) foundation.Animation {
	if !u.settings.AnimationsEnabled {
		return nil
	}
	white := u.uiTheme.GetColorByName("white")
	yellow := u.uiTheme.GetColorByName("yellow_3")
	red := u.uiTheme.GetColorByName("red_8")
	orange := u.uiTheme.GetColorByName("orange_4")
	darkGray := u.uiTheme.GetColorByName("dark_gray_3")
	black := u.uiTheme.GetColorByName("Black")
	frames := []textiles.TextIcon{
		{Char: '.', Fg: white, Bg: yellow},
		{Char: '∙', Fg: yellow, Bg: orange},
		{Char: '*', Fg: orange, Bg: orange},
		{Char: '*', Fg: orange, Bg: red},
		{Char: '*', Fg: red, Bg: orange},
		{Char: '☼', Fg: darkGray, Bg: black},
	}

	animation := NewTilesAnimation(hitPositions, frames, done)
	animation.SetLightsOnAllTiles(&gridmap.LightSource{
		Radius:       2,
		Color:        fxtools.NewColorFromRGBA(orange),
		MaxIntensity: 0.9,
	})
	return animation
}

func (u *UI) GetAnimUncloakAtPosition(actor foundation.ActorForUI, uncloakLocation geometry.Point) (foundation.Animation, int) {
	actorIcon := u.getIconForActor(actor)
	tileIcon := u.getIconForMap(uncloakLocation)
	lightGray := u.uiTheme.GetColorByName("light_gray_5")
	darkGray := u.uiTheme.GetColorByName("dark_gray_3")
	black := u.uiTheme.GetColorByName("Black")
	frames := []textiles.TextIcon{
		tileIcon,
		tileIcon.WithFg(lightGray),
		tileIcon.WithFg(lightGray),
		tileIcon.WithFg(darkGray),
		tileIcon.WithFg(darkGray),
		tileIcon.WithFg(black),
		tileIcon.WithFg(black),
		actorIcon.WithFg(black),
		actorIcon.WithFg(darkGray),
		actorIcon.WithFg(darkGray),
		actorIcon.WithFg(lightGray),
		actorIcon.WithFg(lightGray),
		actorIcon,
	}
	uncloakAnim := u.GetAnimTiles([]geometry.Point{uncloakLocation}, frames, nil)
	return uncloakAnim, len(frames)
}
