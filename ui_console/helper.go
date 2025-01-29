package ui_console

import (
	"contractor/foundation"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type InputCapturer interface {
	SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey)
	GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey
	SetRect(x int, y int, w int, h int)
}

type Focuser interface {
	pushFocus()
	popFocus()
}

type FocusInfo struct {
	Primitive   cview.Primitive
	BeforeFocus func(cview.Primitive) bool
}

type UIState int

func (s UIState) IsTargeting() bool {
	return s == StateTargeting || s == StateLookTargeting || s == StateTargetingBodyPart
}

const (
	StateNormal UIState = iota
	StateLookTargeting
	StateTargeting
	StateTargetingBodyPart
)

type IntroState int

const (
	NoIntro IntroState = iota
	Fading
	TitleScreen
	MainMenu
	InGame
)

type UILifeCycler interface {
	StartGameLoop(settings *foundation.Configuration, application *cview.Application, afterScreenReady func())
	QuitGame(application *cview.Application)
}

type InputPrimitive interface {
	cview.Primitive
	SetInputCapture(f func(event *tcell.EventKey) *tcell.EventKey)
	GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey
}

func EscapeKeyEvent() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
}

func RightPadColored(s string, pLen int) string {
	width := cview.TaggedStringWidth(s)
	return s + strings.Repeat(" ", pLen-width)
}

func toLinesOfText(highScores []foundation.ScoreInfo) []string {
	scoreTable := []string{
		"= Top 10 Contractors =",
		"",
	}
	for i, highScore := range highScores {
		if i == 10 {
			break
		}
		scoreLine := ""
		if highScore.Escaped {
			scoreLine = fmt.Sprintf("[#c9c54d::b]%d. %s: %d sat, %s[-:-:-]", i+1, highScore.PlayerName, highScore.Gold, highScore.DescriptiveMessage)
		} else {
			scoreLine = fmt.Sprintf("%d. %s: %d sat, CoD: %s", i+1, highScore.PlayerName, highScore.Gold, highScore.DescriptiveMessage)
		}
		scoreTable = append(scoreTable, scoreLine)
	}
	return scoreTable
}

func chooseSubDirMenuItems(savegameBaseDirectory string, onSubDirConfirmed func(savegameSubdir string)) []foundation.MenuItem {
	entries, readErr := os.ReadDir(savegameBaseDirectory)
	if readErr != nil {
		return nil
	}
	var menuItems []foundation.MenuItem
	for _, entry := range entries {
		if entry.IsDir() {
			subDir := entry.Name()
			menuItems = append(menuItems, foundation.MenuItem{
				Name: subDir,
				Action: func() {
					onSubDirConfirmed(filepath.Join(savegameBaseDirectory, subDir))
				},
				CloseMenus: true,
			})
		}
	}
	return menuItems
}

func applyLightToMaterial(lightAtCell fxtools.HDRColor, material color.RGBA) fxtools.HDRColor {
	return lightAtCell.Multiply(fxtools.NewRGBColorFromBytes(material.R, material.G, material.B))
}

func applyGamma(colorChannel uint8, gamma float64) uint8 {
	colorAsFloat := float64(colorChannel) / 255.0
	gammaCorrected := fxtools.Clamp(0, 1, math.Pow(colorAsFloat, gamma))
	asEightBit := uint8(gammaCorrected * 255.0)
	return asEightBit
}

func tcellColorToRGBA(tColor tcell.Color) color.RGBA {
	rF, gF, bF := tColor.RGB()
	return color.RGBA{R: uint8(rF), G: uint8(gF), B: uint8(bF), A: 255}
}

func IconAsString(icon textiles.TextIcon) string {
	code := textiles.RGBAToColorCodes(icon.Fg, icon.Bg)
	return fmt.Sprintf("%s%s[-:-]", code, string(icon.Char))
}

func PlayerFlagStringLong(flags map[foundation.ActorFlag]int) string {
	var flagStrings []string
	for flag := foundation.ActorFlag(0); flag < foundation.FlagCount; flag++ {
		if !flag.ShowInHud() {
			continue
		}
		if count, ok := flags[flag]; ok {
			var flagLine string
			if count > 1 {
				flagLine = fmt.Sprintf("%s(%d)", flag.String(), count)
			} else {
				flagLine = fmt.Sprintf("%s", flag.String())
			}

			flagStrings = append(flagStrings, flagLine)
		}
	}
	return strings.Join(flagStrings, " | ")
}

func PlayerFlagStringShort(flags map[foundation.ActorFlag]int) string {
	var flagStrings []string
	for flag := foundation.ActorFlag(0); flag < foundation.FlagCount; flag++ {
		if !flag.ShowInHud() {
			continue
		}
		if count, ok := flags[flag]; ok {
			var flagLine string
			if count > 1 {
				flagLine = fmt.Sprintf("%s(%d)", flag.StringShort(), count)
			} else {
				flagLine = fmt.Sprintf("%s", flag.StringShort())
			}

			flagStrings = append(flagStrings, flagLine)
		}
	}
	return strings.Join(flagStrings, " ")
}

func expandToWidth(statusStr string, width int) string {
	statusWidth := cview.TaggedStringWidth(statusStr)
	if statusWidth < width {
		statusStr = fxtools.RightPadCount(statusStr, width-statusWidth)
	}
	return statusStr
}

func ToColoredText(h foundation.HiLiteString, intensity float64, textColor color.RGBA, hiLiteColor color.RGBA) string {
	if h.IsEmpty() {
		return ""
	}
	if intensity < 1.0 {
		textColor = fxtools.SetBrightness(textColor, intensity)
		hiLiteColor = fxtools.SetBrightness(hiLiteColor, intensity)
	}
	if len(h.Value) == 0 {
		return h.AppendRepetitions(fmt.Sprintf("%s%s", textiles.RGBAToFgColorCode(hiLiteColor), h.FormatString))
	}
	textColorCode := textiles.RGBAToFgColorCode(textColor)
	if h.FormatString == "" {
		return h.AppendRepetitions(fmt.Sprintf("%s%s", textColorCode, h.Value[0]))
	}
	hiLiteColorCode := textiles.RGBAToFgColorCode(hiLiteColor)
	anyValues := make([]interface{}, len(h.Value)+1)
	anyValues[0] = textColorCode
	for i, v := range h.Value {
		anyValues[i+1] = fmt.Sprintf("%s%s%s", hiLiteColorCode, v, textColorCode)
	}
	return h.AppendRepetitions(fmt.Sprintf("%s"+h.FormatString, anyValues...))
}

func FullColorBarFromPercent(currentVal int, maxVal int, width int, gradient [3]color.RGBA, inactive color.RGBA, lightText color.RGBA, darkText color.RGBA) string {
	percent := float64(currentVal) / float64(maxVal)
	colorChangeIndex := int(math.Round(percent * float64(width)))

	contrastColor := func(ourColor color.RGBA) color.RGBA {
		if isMoreLightThanDark(ourColor) {
			return darkText
		}
		return lightText
	}

	colorCode := textiles.RGBAToColorCodes(contrastColor(gradient[0]), gradient[0])
	if percent < 0.50 {
		colorCode = textiles.RGBAToColorCodes(contrastColor(gradient[2]), gradient[2])
	} else if percent < 0.75 {
		colorCode = textiles.RGBAToColorCodes(contrastColor(gradient[1]), gradient[1])
	}
	darkGrayCode := textiles.RGBAToColorCodes(contrastColor(inactive), inactive)

	valString := fmt.Sprintf("%d/%d", currentVal, maxVal)
	xForCenter := (width - len(valString)) / 2
	prefix := strings.Repeat(" ", xForCenter)
	suffix := strings.Repeat(" ", width-len(valString)-xForCenter)
	barString := fmt.Sprintf("%s%s%s", prefix, valString, suffix)

	if colorChangeIndex > len(barString) {
		colorChangeIndex = len(barString) - 1
	} else if colorChangeIndex < 0 {
		colorChangeIndex = 0
	}

	barString = colorCode + barString[:colorChangeIndex] + darkGrayCode + barString[colorChangeIndex:] + "[-:-]"
	return barString
}
func RuneBarWithColor(icon rune, current int, max int, fgColor color.RGBA, bgColor color.RGBA, inactiveColor color.RGBA) string {
	colorCode := textiles.RGBAToColorCodes(fgColor, bgColor)
	darkGrayCode := textiles.RGBAToBgColorCode(inactiveColor)
	return colorCode + strings.Repeat(string(icon), current) + "[-:-]" + darkGrayCode + strings.Repeat(" ", max-current) + "[-:-]"
}

func RuneBarFromPercent(icon rune, percent float64, width int, stageColors [3]color.RGBA) string {
	repeats := int(math.Round(percent * float64(width)))
	colorCode := textiles.RGBAToFgColorCode(stageColors[0])
	if percent < 0.50 {
		colorCode = textiles.RGBAToFgColorCode(stageColors[2])
	} else if percent < 0.75 {
		colorCode = textiles.RGBAToFgColorCode(stageColors[1])
	}
	return colorCode + strings.Repeat(string(icon), repeats) + "[-]" + strings.Repeat(" ", width-repeats)
}
func desaturate(fg color.RGBA) color.RGBA {
	gray := uint8((uint16(fg.R) + uint16(fg.G) + uint16(fg.B)) / 3)
	return color.RGBA{R: gray, G: gray, B: gray, A: fg.A}
}

func darken(fg color.RGBA) color.RGBA {
	return color.RGBA{
		R: fg.R / 2,
		G: fg.G / 2,
		B: fg.B / 2,
		A: fg.A,
	}
}
func toTcellColor(rgba color.RGBA) tcell.Color {
	return tcell.NewRGBColor(int32(rgba.R), int32(rgba.G), int32(rgba.B))
}

func runeToDirection(r rune) geometry.CompassDirection {
	switch r {
	case '8':
		fallthrough
	case 'w':
		return geometry.North
	case '2':
		fallthrough
	case 's':
		return geometry.South
	case '4':
		fallthrough
	case 'a':
		return geometry.West
	case '6':
		fallthrough
	case 'd':
		return geometry.East
	case '7':
		return geometry.NorthWest
	case '9':
		return geometry.NorthEast
	case '1':
		return geometry.SouthWest
	case '3':
		return geometry.SouthEast
	}
	return geometry.North
}

func upperRuneToDirection(r rune) geometry.CompassDirection {
	switch r {
	case 'W':
		return geometry.North
	case 'S':
		return geometry.South
	case 'A':
		return geometry.West
	case 'D':
		return geometry.East
	}
	return geometry.North
}

func directionToRune(dir geometry.CompassDirection) rune {
	switch dir {
	case geometry.North:
		return '8'
	case geometry.South:
		return '2'
	case geometry.West:
		return '4'
	case geometry.East:
		return '6'
	case geometry.NorthWest:
		return '7'
	case geometry.NorthEast:
		return '9'
	case geometry.SouthWest:
		return '1'
	case geometry.SouthEast:
		return '3'
	}
	return 'w'
}

func offsetVertically(primitive cview.Primitive, amount int) {
	x, y, w, h := primitive.GetRect()
	primitive.SetRect(x, y+amount, w, h)
}

func FadeToWhite(app *cview.Application, animDelay time.Duration, stepSize int) {
	screen := app.GetScreen()
	duration := 10 * time.Millisecond
	var waited time.Duration
	for i := 0; i < 100; i++ {
		if !lightenScreen(screen, stepSize) {
			return
		}
		screen.Show()
		waited = 0
		for waited < animDelay {
			if screen.HasPendingEvent() {
				ev := screen.PollEvent()
				if _, ok := ev.(*tcell.EventKey); ok {
					return
				}
			}

			time.Sleep(duration)
			waited += duration
		}
	}
	return
}

func lightenScreen(screen tcell.Screen, size int) bool {
	w, h := screen.Size()
	workLeft := false
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			workDone := lightenScreenLocation(screen, x, y, int32(size))
			if workDone {
				workLeft = true
			}
		}
	}
	return workLeft
}

func lightenScreenLocation(screen tcell.Screen, x int, y int, amount int32) bool {
	icon, _, style, _ := screen.GetContent(x, y)
	fg, bg, _ := style.Decompose()
	fR, fG, fB := fg.RGB()
	bR, bG, bB := bg.RGB()
	hadWorkLeft := fR < 255 || fG < 255 || fB < 255 || bR < 255 || bG < 255 || bB < 255
	newFG := tcell.NewRGBColor(min(255, fR+amount), min(255, fG+amount), min(255, fB+amount))
	newBG := tcell.NewRGBColor(min(255, bR+amount), min(255, bG+amount), min(255, bB+amount))
	screen.SetContent(x, y, icon, nil, style.Background(newBG).Foreground(newFG))
	return hadWorkLeft
}

func wrapPrimitivesSideBySide(p, q cview.Primitive, width, height int) cview.Primitive {
	containerLeft := cview.NewFlex()
	containerLeft.SetDirection(cview.FlexRow)
	containerLeft.AddItem(nil, 0, 1, false)
	containerLeft.AddItem(p, height, 1, false)
	containerLeft.AddItem(nil, 0, 1, false)

	containerRight := cview.NewFlex()
	containerRight.SetDirection(cview.FlexRow)
	containerRight.AddItem(nil, 0, 1, false)
	containerRight.AddItem(q, height, 1, true)
	containerRight.AddItem(nil, 0, 1, false)

	flex := cview.NewFlex()
	flex.AddItem(containerLeft, width, 1, false)
	flex.AddItem(nil, 0, 1, false)
	flex.AddItem(containerRight, width, 1, true)
	return flex
}

func wrapPrimitivesTopToBottom(p, q cview.Primitive) cview.Primitive {

	container := cview.NewFlex()
	container.SetDirection(cview.FlexRow)
	container.AddItem(p, 0, 1, false)
	container.AddItem(q, 0, 1, true)

	return container
}

func widthAndHeightFromString(description string) (int, int) {
	longest := 0
	height := 0
	for _, line := range strings.Split(description, "\n") {
		withoutColors := cview.TaggedStringWidth(line)
		longest = max(longest, withoutColors)
		height++
	}
	return longest, height
}

func longestInventoryLineWithoutColorCodes(items []foundation.Item) int {
	longest := 0
	for _, line := range items {
		withoutColors := line.DisplayLength()
		longest = max(longest, withoutColors)
	}
	return longest
}

func drawBackgroundAndBorderWithTitleForInventory(screen tcell.Screen, x int, y int, width int, height int, title string, style tcell.Style, runes []rune) {
	horizontal := runes[0]
	vertical := runes[1]
	topLeft := runes[2]
	topRight := runes[3]
	//bottomRight := runes[4]
	bottomLeft := runes[5]
	fg, _, _ := style.Decompose()
	// fill the background
	for i := x; i < x+width; i++ {
		for j := y; j < y+height; j++ {
			screen.SetContent(i, j, ' ', nil, style)
		}
	}

	// Draw the corners
	cview.Print(screen, []byte(string(topLeft)), x, y, width, cview.AlignLeft, fg)
	cview.Print(screen, []byte(string(topRight)), x+width-1, y, width, cview.AlignLeft, fg)
	cview.Print(screen, []byte(string(bottomLeft)), x, y+height-1, width, cview.AlignLeft, fg)

	// center title
	startTitleX := x + 1 + (width-1-len(title))/2
	endTitleX := startTitleX + len(title)
	// Draw the horizontal borders
	for i := x + 1; i < x+width-1; i++ {
		if title != "" && i >= startTitleX && i < endTitleX {
			cview.Print(screen, []byte(string(title[i-startTitleX])), i, y, width, cview.AlignLeft, fg)
		} else {
			cview.Print(screen, []byte(string(horizontal)), i, y, width, cview.AlignLeft, fg)
		}
	}

	// Draw the vertical borders
	for i := y + 1; i < y+height-1; i++ {
		cview.Print(screen, []byte(string(vertical)), x, i, width, cview.AlignLeft, fg)
	}
}

func drawBackgroundAndBorderWithTitle(screen tcell.Screen, x int, y int, width int, height int, title string, style tcell.Style, runes []rune) {
	horizontal := runes[0]
	vertical := runes[1]
	topLeft := runes[2]
	topRight := runes[3]
	bottomRight := runes[4]
	bottomLeft := runes[5]
	fg, _, _ := style.Decompose()
	// fill the background
	for i := x; i < x+width; i++ {
		for j := y; j < y+height; j++ {
			screen.SetContent(i, j, ' ', nil, style)
		}
	}

	// Draw the corners
	cview.Print(screen, []byte(string(topLeft)), x, y, width, cview.AlignLeft, fg)
	cview.Print(screen, []byte(string(topRight)), x+width-1, y, width, cview.AlignLeft, fg)
	cview.Print(screen, []byte(string(bottomRight)), x+width-1, y+height-1, width, cview.AlignLeft, fg)
	cview.Print(screen, []byte(string(bottomLeft)), x, y+height-1, width, cview.AlignLeft, fg)

	// center title
	startTitleX := x + 1 + (width-1-len(title))/2
	endTitleX := startTitleX + len(title)
	// Draw the horizontal borders
	for i := x + 1; i < x+width-1; i++ {
		if title != "" && i >= startTitleX && i < endTitleX {
			cview.Print(screen, []byte(string(title[i-startTitleX])), i, y, width, cview.AlignLeft, fg)
		} else {
			cview.Print(screen, []byte(string(horizontal)), i, y, width, cview.AlignLeft, fg)
		}
		cview.Print(screen, []byte(string(horizontal)), i, y+height-1, width, cview.AlignLeft, fg)
	}

	// Draw the vertical borders
	for i := y + 1; i < y+height-1; i++ {
		cview.Print(screen, []byte(string(vertical)), x, i, width, cview.AlignLeft, fg)
		cview.Print(screen, []byte(string(vertical)), x+width-1, i, width, cview.AlignLeft, fg)
	}
}

func iterateBorderTiles(x int, y int, width int, height int, runes []rune, setTile func(r rune, x, y int)) {
	horizontal := runes[0]
	vertical := runes[1]
	topLeft := runes[2]
	topRight := runes[3]
	bottomRight := runes[4]
	bottomLeft := runes[5]
	// fill the background

	// Draw the corners
	setTile(topLeft, x, y)
	setTile(topRight, x+width-1, y)
	setTile(bottomRight, x+width-1, y+height-1)
	setTile(bottomLeft, x, y+height-1)

	// center title
	// Draw the horizontal borders
	for i := x + 1; i < x+width-1; i++ {
		setTile(horizontal, i, y)
		setTile(horizontal, i, y+height-1)
	}

	// Draw the vertical borders
	for i := y + 1; i < y+height-1; i++ {
		setTile(vertical, x, i)
		setTile(vertical, x+width-1, i)
	}
}

func setListItemsFromMenuItemsWithNumbers(list *cview.List, menuItems []foundation.MenuItem) int {
	list.Clear()
	longestItem := 0
	for index, a := range menuItems {
		action := a
		listItem := cview.NewListItem(action.Name)
		// we need the runes 0-9
		asRune := '0' + rune(index+1)
		listItem.SetShortcut(asRune)
		list.AddItem(listItem)
		itemLength := cview.TaggedStringWidth(action.Name) + 4
		longestItem = max(longestItem, itemLength)
	}
	return longestItem
}
func setListItemsFromMenuItems(list *cview.List, menuItems []foundation.MenuItem) int {
	list.Clear()
	longestItem := 0
	for index, a := range menuItems {
		action := a
		shortcut := foundation.ShortCutFromIndex(index)
		listItem := cview.NewListItem(action.Name)
		listItem.SetShortcut(shortcut)
		list.AddItem(listItem)
		itemLength := cview.TaggedStringWidth(action.Name) + 4
		longestItem = max(longestItem, itemLength)
	}
	return longestItem
}
