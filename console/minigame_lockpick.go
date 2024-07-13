package console

import (
	"RogueUI/cview"
	"RogueUI/foundation"
	"github.com/gdamore/tcell/v2"
	"math/rand"
	"slices"
)

type Pattern []int

func (p Pattern) Sorted() Pattern {
	sorted := make(Pattern, len(p))
	copy(sorted, p)
	slices.Sort(sorted)
	return sorted
}

func (p Pattern) OffsetBy(offset int) Pattern {
	patternCopy := slices.Clone(p)
	for i := 0; i < len(p); i++ {
		patternCopy[i] = (p[i] + offset) % 36
	}
	patternCopy = patternCopy.Sorted()
	return patternCopy
}

func (p Pattern) PickedBy(pick Pattern) Pattern {
	newPattern := p
	for i := 0; i < len(pick); i++ {
		for j := 0; j < len(newPattern); j++ {
			if pick[i] == newPattern[j] {
				newPattern = append(newPattern[:j], newPattern[j+1:]...)
				break
			}
		}
	}
	return newPattern
}

type LockpickGame struct {
	*cview.Box
	style                 tcell.Style
	closeFunc             func(result foundation.InteractionResult)
	borders               cview.BorderDef
	audioPlayer           AudioCuePlayer
	currentlySelectedPick int
	mouseX                int

	lockPatterns []Pattern
	pickPatterns []Pattern
	pickOffsets  []int
	seed         int64
	destroyPick  func()
	getPickCount func() int
	difficulty   foundation.Difficulty
}

func NewLockpickGame(seed int64, difficulty foundation.Difficulty, pickCount func() int, destroyPick func(), close func(result foundation.InteractionResult)) *LockpickGame {
	box := cview.NewBox()
	//box.SetBorder(true)
	//box.SetTitle("Lockpick")
	greenStyle := tcell.StyleDefault.Foreground(tcell.ColorForestGreen).Background(tcell.ColorBlack)
	h := &LockpickGame{
		Box:          box,
		style:        greenStyle,
		closeFunc:    close,
		seed:         seed,
		getPickCount: pickCount,
		destroyPick:  destroyPick,
		difficulty:   difficulty,
		borders: cview.BorderDef{
			Horizontal:  '─',
			Vertical:    '│',
			TopLeft:     '╭',
			TopRight:    '╮',
			BottomLeft:  '╰',
			BottomRight: '╯',
		},
	}

	box.SetBorder(false)
	box.SetDrawFunc(h.drawInside)
	box.SetBackgroundTransparent(false)
	box.SetInputCapture(h.handleInput)
	box.SetMouseCapture(h.handleMouse)

	h.initPatterns()
	return h
}

func (lg *LockpickGame) initPatterns() {
	rnd := rand.New(rand.NewSource(lg.seed))
	// a pattern is 36 chars long, so it contains indices between 0 and 35

	diff := lg.difficulty
	var minPickSize, lockPatternCount, minLockSize, maxLockSize int
	switch diff {
	default: // very easy
		minPickSize = 2
		lockPatternCount = 2
		minLockSize = 3
		maxLockSize = 5
	case foundation.Easy:
		minPickSize = 2
		lockPatternCount = 2
		minLockSize = 4
		maxLockSize = 6
	case foundation.Medium:
		minPickSize = 3
		lockPatternCount = 3
		minLockSize = 4
		maxLockSize = 7
	case foundation.Hard:
		minPickSize = 3
		lockPatternCount = 4
		minLockSize = 5
		maxLockSize = 8
	case foundation.VeryHard:
		minPickSize = 4
		lockPatternCount = 4
		minLockSize = 6
		maxLockSize = 8
	}

	var pickPatterns []Pattern
	lockPatterns := make([]Pattern, lockPatternCount)
	var allIndices []int

	for i := 0; i < lockPatternCount; i++ {
		// a lock pattern contains 4-8 indices
		indexCount := rnd.Intn(maxLockSize-minLockSize+1) + minLockSize
		lockPatterns[i] = make(Pattern, indexCount)
		allowedIndices := rnd.Perm(36)
		for j := 0; j < indexCount; j++ {
			lockIndex := allowedIndices[j]
			lockPatterns[i][j] = lockIndex
			allIndices = append(allIndices, lockIndex)
		}
		lockPatterns[i] = lockPatterns[i].Sorted()
		currentPattern := slices.Clone(lockPatterns[i])
		if len(currentPattern) < minPickSize*2 {
			// cannot split further
			pickPatterns = append(pickPatterns, currentPattern)
			continue
		}
		for len(currentPattern) >= minPickSize*2 {
			maxSplit := len(currentPattern) - minPickSize
			rndRange := maxSplit - minPickSize + 1
			splitCount := minPickSize
			if rndRange > 0 {
				splitCount += rnd.Intn(rndRange)
			}
			// random split
			var pickPattern Pattern
			for j := 0; j < splitCount; j++ {
				permutedIdx := rnd.Intn(len(currentPattern))
				pickPattern = append(pickPattern, currentPattern[permutedIdx])
				currentPattern = append(currentPattern[:permutedIdx], currentPattern[permutedIdx+1:]...)
			}
			pickPatterns = append(pickPatterns, pickPattern.Sorted())
			if len(currentPattern) < minPickSize*2 {
				// cannot split further
				pickPatterns = append(pickPatterns, currentPattern)
			}
		}
	}

	offsets := make([]int, len(pickPatterns))
	for i := 0; i < len(pickPatterns); i++ {
		offsets[i] = rnd.Intn(35) + 1
	}

	lg.lockPatterns = lockPatterns
	lg.pickPatterns = pickPatterns
	lg.pickOffsets = offsets
	lg.currentlySelectedPick = 0
}

func (lg *LockpickGame) drawPattern(screen tcell.Screen, x int, y int, offset int, setChar, unsetChar rune, pattern Pattern, style tcell.Style) {
	var patternCopy Pattern
	if offset != 0 {
		patternCopy = pattern.OffsetBy(offset)
	} else {
		patternCopy = pattern
	}
	currentIndex := 0
	nextSetIndex := patternCopy[currentIndex]

	// a pattern is always 36 chars
	for i := 0; i < 36; i++ {
		if i == nextSetIndex {
			screen.SetContent(x+i, y, setChar, nil, style)
			currentIndex++
			if currentIndex >= len(patternCopy) {
				nextSetIndex = -1
			} else {
				nextSetIndex = patternCopy[currentIndex]
			}
		} else {
			screen.SetContent(x+i, y, unsetChar, nil, style)
		}
	}
}

func (lg *LockpickGame) SetAmberStyle() {
	lg.style = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
}
func (lg *LockpickGame) playCue(cue string) {
	if lg.audioPlayer != nil {
		lg.audioPlayer.PlayCue(cue)
	}
}
func (lg *LockpickGame) drawInside(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	if width < 40 || height < 10 {
		return x, y, width, height
	}

	cview.DrawBox(screen, x, y, x+width-1, y+height-1, lg.style, lg.borders, ' ')

	// draw all lock patterns, from bottom to top
	for i, lockPattern := range lg.lockPatterns {
		yPos := y + height - 2 - i*2
		lg.drawPattern(screen, x+2, yPos, 0, '_', '█', lockPattern, lg.style)
	}

	// draw the currently selected pick right above the lock patterns
	if lg.currentlySelectedPick >= 0 && lg.currentlySelectedPick < len(lg.pickPatterns) {
		drawStyle := lg.style
		yPos := y + height - 4 - len(lg.lockPatterns)*2
		lg.drawPattern(screen, x+2, yPos, lg.pickOffsets[lg.currentlySelectedPick], '█', '=', lg.pickPatterns[lg.currentlySelectedPick], drawStyle.Foreground(tcell.ColorYellow))
	}

	// draw all pick patterns / on the right
	for i, pickPattern := range lg.pickPatterns {
		pickOffset := lg.pickOffsets[i]
		drawStyle := lg.style
		if i == lg.currentlySelectedPick {
			drawStyle = drawStyle.Foreground(tcell.ColorYellow)
		}
		yPos := y + height - 2 - i*2
		lg.drawPattern(screen, x+42, yPos, pickOffset, '█', '=', pickPattern, drawStyle)
	}

	return x, y, width, height
}

func (lg *LockpickGame) handleInput(event *tcell.EventKey) *tcell.EventKey {
	//uikey := toUIKey(event)
	switch event.Key() {
	case tcell.KeyDown:
		lg.currentlySelectedPick--
		if lg.currentlySelectedPick < 0 {
			lg.currentlySelectedPick = len(lg.pickPatterns) - 1
		}
	case tcell.KeyUp:
		lg.currentlySelectedPick++
		if lg.currentlySelectedPick >= len(lg.pickPatterns) {
			lg.currentlySelectedPick = 0
		}
	case tcell.KeyLeft:
		if lg.currentlySelectedPick >= 0 && lg.currentlySelectedPick < len(lg.pickPatterns) {
			lg.pickOffsets[lg.currentlySelectedPick]--
			if lg.pickOffsets[lg.currentlySelectedPick] < 0 {
				lg.pickOffsets[lg.currentlySelectedPick] = 35
			}
		}
	case tcell.KeyRight:
		if lg.currentlySelectedPick >= 0 && lg.currentlySelectedPick < len(lg.pickPatterns) {
			lg.pickOffsets[lg.currentlySelectedPick]++
			if lg.pickOffsets[lg.currentlySelectedPick] >= 36 {
				lg.pickOffsets[lg.currentlySelectedPick] = 0
			}
		}
	case tcell.KeyEnter:
		if lg.currentlySelectedPick >= 0 && lg.currentlySelectedPick < len(lg.pickPatterns) {
			lg.confirmPick()
		}
	case tcell.KeyDEL:
		if lg.getPickCount() > 0 {
			lg.destroyPick()
			lg.initPatterns()
		}
	case tcell.KeyEsc:
		lg.closeFunc(foundation.Cancel)
	}
	return event
}
func (lg *LockpickGame) SetAudioPlayer(player AudioCuePlayer) {
	lg.audioPlayer = player
}

func (lg *LockpickGame) handleMouse(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
	rawMouseX, _ := event.Position()

	if rawMouseX != lg.mouseX {
		lg.mouseX = rawMouseX
		return cview.MouseLeftClick, event
	}

	return action, event
}

func (lg *LockpickGame) confirmPick() {
	lastLockIndex := len(lg.lockPatterns) - 1
	lastLockPattern := lg.lockPatterns[lastLockIndex]
	pickPattern := lg.pickPatterns[lg.currentlySelectedPick]
	offset := lg.pickOffsets[lg.currentlySelectedPick]
	lockPatternAfterPicking := lastLockPattern.PickedBy(pickPattern.OffsetBy(offset))
	if len(lockPatternAfterPicking) == 0 {
		lg.lockPatterns = lg.lockPatterns[:lastLockIndex]
		if len(lg.lockPatterns) == 0 {
			lg.closeFunc(foundation.Success)
		}
	} else {
		lg.lockPatterns[lastLockIndex] = lockPatternAfterPicking
	}
	lg.pickPatterns = append(lg.pickPatterns[:lg.currentlySelectedPick], lg.pickPatterns[lg.currentlySelectedPick+1:]...)
	lg.pickOffsets = append(lg.pickOffsets[:lg.currentlySelectedPick], lg.pickOffsets[lg.currentlySelectedPick+1:]...)
	if lg.currentlySelectedPick >= len(lg.pickPatterns) {
		lg.currentlySelectedPick = len(lg.pickPatterns) - 1
	}
}
