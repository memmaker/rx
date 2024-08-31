package console

import (
	"RogueUI/foundation"
	"bufio"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"regexp"
	"strings"
)

func (u *UI) executePlayerCommand(command string) {
	if command == "" {
		return
	}
	if action, ok := u.commandTable[command]; ok {
		action()
	} else {
		u.Print(foundation.Msg(fmt.Sprintf("Unknown command: %s", command)))
	}
}

func directionFromCommand(command string) (geometry.CompassDirection, bool) {
	switch command {
	case "run_north":
		fallthrough
	case "north":
		return geometry.North, true
	case "run_south":
		fallthrough
	case "south":
		return geometry.South, true
	case "run_west":
		fallthrough
	case "west":
		return geometry.West, true
	case "run_east":
		fallthrough
	case "east":
		return geometry.East, true
	case "run_northwest":
		fallthrough
	case "northwest":
		return geometry.NorthWest, true
	case "run_northeast":
		fallthrough
	case "northeast":
		return geometry.NorthEast, true
	case "run_southwest":
		fallthrough
	case "southwest":
		return geometry.SouthWest, true
	case "run_southeast":
		fallthrough
	case "southeast":
		return geometry.SouthEast, true
	}
	return geometry.East, false
}

func (u *UI) setupCommandTable() {
	u.commandTable = make(map[string]func())

	u.commandTable["quit"] = u.application.Stop

	u.commandTable["inventory"] = u.game.OpenInventory
	u.commandTable["tactics"] = u.game.OpenTacticsMenu
	u.commandTable["character"] = u.ShowCharacterSheet
	u.commandTable["wizard"] = u.game.OpenWizardMenu
	u.commandTable["rest"] = u.game.OpenRestMenu
	u.commandTable["journal"] = u.game.OpenJournal
	u.commandTable["log"] = u.ShowLog
	u.commandTable["monsters"] = u.ShowVisibleEnemies
	u.commandTable["items"] = u.ShowVisibleItems
	u.commandTable["help"] = u.ShowHelpScreen
	u.commandTable["log"] = u.ShowLog

	u.commandTable["north"] = func() { u.game.ManualMovePlayer(geometry.North) }
	u.commandTable["south"] = func() { u.game.ManualMovePlayer(geometry.South) }
	u.commandTable["west"] = func() { u.game.ManualMovePlayer(geometry.West) }
	u.commandTable["east"] = func() { u.game.ManualMovePlayer(geometry.East) }
	u.commandTable["northwest"] = func() { u.game.ManualMovePlayer(geometry.NorthWest) }
	u.commandTable["northeast"] = func() { u.game.ManualMovePlayer(geometry.NorthEast) }
	u.commandTable["southwest"] = func() { u.game.ManualMovePlayer(geometry.SouthWest) }
	u.commandTable["southeast"] = func() { u.game.ManualMovePlayer(geometry.SouthEast) }

	// u.startAutoRun(direction)
	u.commandTable["run_north"] = func() { u.startAutoRun(geometry.North) }
	u.commandTable["run_south"] = func() { u.startAutoRun(geometry.South) }
	u.commandTable["run_west"] = func() { u.startAutoRun(geometry.West) }
	u.commandTable["run_east"] = func() { u.startAutoRun(geometry.East) }
	u.commandTable["run_northwest"] = func() { u.startAutoRun(geometry.NorthWest) }
	u.commandTable["run_northeast"] = func() { u.startAutoRun(geometry.NorthEast) }
	u.commandTable["run_southwest"] = func() { u.startAutoRun(geometry.SouthWest) }
	u.commandTable["run_southeast"] = func() { u.startAutoRun(geometry.SouthEast) }

	u.commandTable["look"] = u.LookTargeting

	u.commandTable["overlay_monsters"] = u.ShowEnemyOverlay
	u.commandTable["overlay_items"] = u.ShowItemOverlay

	u.commandTable["gamma_up"] = func() {
		u.gamma += 0.1
		u.Print(foundation.Msg(fmt.Sprintf("Gamma: %.1f", u.gamma)))
	}
	u.commandTable["gamma_down"] = func() {
		u.gamma -= 0.1
		u.Print(foundation.Msg(fmt.Sprintf("Gamma: %.1f", u.gamma)))
	}

	u.commandTable["toggle_cursor"] = func() {
		u.SetShowCursor(!u.showCursor)
	}

	u.commandTable["throw"] = u.game.ChooseItemForThrow

	//u.commandTable["quaff"] = u.game.ChooseItemForQuaff
	//u.commandTable["read"] = u.game.ChooseItemForRead
	//u.commandTable["zap"] = u.game.ChooseItemForZap
	//u.commandTable["use"] = u.game.ChooseItemForUse
	u.commandTable["wear"] = u.game.ChooseArmorForWear
	u.commandTable["take_off"] = u.game.ChooseArmorToTakeOff
	u.commandTable["wield"] = u.game.ChooseWeaponForWield
	//u.commandTable["ring_put_on"] = u.game.ChooseRingToPutOn
	//u.commandTable["ring_remove"] = u.game.ChooseRingToRemove
	u.commandTable["drop"] = u.game.ChooseItemForDrop
	u.commandTable["eat"] = u.game.ChooseItemForEat

	u.commandTable["apply"] = u.game.ChooseItemForApply
	//u.commandTable["launch"] = u.game.ChooseItemForMissileLaunch
	u.commandTable["attack"] = u.game.PlayerRangedAttack
	u.commandTable["quick_attack"] = u.game.PlayerQuickRangedAttack

	u.commandTable["switch_weapons"] = u.game.SwitchWeapons
	u.commandTable["cycle_target_mode"] = u.game.CycleTargetMode
	u.commandTable["apply_skill"] = u.game.PlayerApplySkill
	u.commandTable["reload_weapon"] = u.game.PlayerReloadWeapon

	//u.commandTable["targeted_shot"] = u.game.TargetedShot

	u.commandTable["pickup"] = u.game.PlayerPickupItem
	u.commandTable["map_interaction"] = u.GenericInteraction
	u.commandTable["run_direction"] = u.ChooseDirectionForRun
	u.commandTable["wait"] = u.game.Wait
	u.commandTable["show_key_bindings"] = u.showKeyBindings
	u.commandTable["open_pip_boy"] = u.openPipBoy

	u.commandTable["wiz_advance_time"] = u.game.WizardAdvanceTime
}

func (u *UI) showKeyBindings() {
	friendlyNames := map[string]string{
		"quit":              "Quit",
		"inventory":         "Inventory",
		"tactics":           "Tactics Menu",
		"rest":              "Rest Menu",
		"character":         "Character",
		"wizard":            "Wizard",
		"themes":            "Themes",
		"log":               "Log",
		"monsters":          "Monster List",
		"items":             "Item List",
		"help":              "Help",
		"show_key_bindings": "Key Bindings",
		"north":             "North",
		"south":             "South",
		"west":              "West",
		"east":              "East",
		"northwest":         "Northwest",
		"northeast":         "Northeast",
		"southwest":         "Southwest",
		"southeast":         "Southeast",
		"run_north":         "Run North",
		"run_south":         "Run South",
		"run_west":          "Run West",
		"run_east":          "Run East",
		"run_northwest":     "Run Northwest",
		"run_northeast":     "Run Northeast",
		"run_southwest":     "Run Southwest",
		"run_southeast":     "Run Southeast",
		"look":              "Look",
		"overlay_monsters":  "Overlay Monsters",
		"overlay_items":     "Overlay Items",
		"gamma_up":          "Gamma Up",
		"gamma_down":        "Gamma Down",
		"throw":             "Throw",
		"use":               "Use",
		"drop":              "Drop",
		"eat":               "Eat",
		"attack":            "Attack",
		"quick_attack":      "Quick Attack",
		"aim":               "Aim",
		"pickup":            "Pickup",
		"map_interaction":   "Map Interaction",
		"run_direction":     "Run Direction",
		"wait":              "Wait",
		"cycle_target_mode": "Cycle Weapon Mode",
	}

	leftColCommands := []string{
		"help",
		"show_key_bindings",
		"gamma_up",
		"gamma_down",
		"north",
		"south",
		"west",
		"east",
		"northwest",
		"northeast",
		"southwest",
		"southeast",
		"run_north",
		"run_south",
		"run_west",
		"run_east",
		"run_northwest",
		"run_northeast",
		"run_southwest",
		"run_southeast",
		"run_direction",
	}

	rightColCommands := []string{
		"map_interaction",
		"wait",
		"throw",
		"drop",
		"pickup",
		"cycle_target_mode",
		"attack",
		"quick_attack",
		"aim",
		"look",
		"inventory",
		"character",
		"tactics",
		"log",
		"rest",
		"monsters",
		"overlay_monsters",
		"items",
		"overlay_items",
		"wizard",
		"quit",
	}

	rowsOfTable := make([]fxtools.TableRow, 0)
	lines := max(len(leftColCommands), len(rightColCommands))
	for i := 0; i < lines; i++ {
		leftKeys := ""
		leftCmd := ""
		rightKeys := ""
		rightCmd := ""
		if i < len(leftColCommands) {
			leftCmd = leftColCommands[i]
			leftKeys = u.GetKeysForCommandAsString(KeyLayerMain, leftCmd)
		}
		if i < len(rightColCommands) {
			rightCmd = rightColCommands[i]
			rightKeys = u.GetKeysForCommandAsString(KeyLayerMain, rightCmd)
		}
		rowsOfTable = append(rowsOfTable, fxtools.TableRow{
			Columns: []string{" ", leftKeys, friendlyNames[leftCmd], "   ", rightKeys, friendlyNames[rightCmd]},
		})
	}
	commandTableRender := fxtools.TableLayout(rowsOfTable, []fxtools.TextAlignment{
		fxtools.AlignLeft,
		fxtools.AlignLeft,
		fxtools.AlignLeft,
		fxtools.AlignLeft,
		fxtools.AlignLeft,
		fxtools.AlignLeft,
	})

	u.OpenTextWindow(strings.Join(commandTableRender, "\n"))
}

func (u *UI) getPressedKey() UIKey {
	s := u.application.GetScreen()
	ev := s.PollEvent()
	keyEvent, isKeyEvent := ev.(*tcell.EventKey)
	for !isKeyEvent {
		ev = s.PollEvent()
		keyEvent, isKeyEvent = ev.(*tcell.EventKey)
	}
	key := toUIKey(keyEvent)
	return key
}

func toUIKey(keyEvent *tcell.EventKey) UIKey {
	name := tcell.KeyNames[keyEvent.Key()]
	if keyEvent.Key() == tcell.KeyRune {
		name = string(keyEvent.Rune())
	}
	key := UIKey{
		mod:  keyEvent.Modifiers(),
		key:  keyEvent.Key(),
		ch:   keyEvent.Rune(),
		name: name,
	}
	return key
}

type KeyLayer int

const (
	KeyLayerMain KeyLayer = iota
	KeyLayerDirectionalTargeting
	KeyLayerAdvancedTargeting
)

func (u *UI) loadKeyMap(filename string) {
	u.keyTable = map[KeyLayer]map[UIKey]string{
		KeyLayerMain:                 make(map[UIKey]string),
		KeyLayerDirectionalTargeting: make(map[UIKey]string),
		KeyLayerAdvancedTargeting:    make(map[UIKey]string),
	}

	file := fxtools.MustOpen(filename)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	currentLayer := KeyLayerMain
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if strings.HasPrefix(line, "%") {
			switch strings.TrimSpace(strings.TrimPrefix(line, "%")) {
			case "main":
				currentLayer = KeyLayerMain
			case "directional_targeting":
				currentLayer = KeyLayerDirectionalTargeting
			case "advanced_targeting":
				currentLayer = KeyLayerAdvancedTargeting
			}
			continue
		}
		parts := strings.Split(line, "->")
		if len(parts) == 2 {
			key := UIKeyFromString(strings.TrimSpace(parts[0]))
			command := strings.TrimSpace(parts[1])
			u.keyTable[currentLayer][key] = command
		}
	}
}

type UIKey struct {
	mod  tcell.ModMask
	key  tcell.Key
	ch   rune
	name string
}

func UIKeyFromString(s string) UIKey {
	if printableRune, ok := ParsePrintableKey(s); ok {
		return Letter(printableRune)
	}
	keyString := s
	var mods tcell.ModMask
	// not a single key
	// so it's either a combo or a function key or a non-printable key
	ctrlComboRegex, _ := regexp.Compile(`Ctrl\+(\w)`)
	if ctrlComboRegex.MatchString(s) {
		matches := ctrlComboRegex.FindStringSubmatch(s)
		keyString = matches[1]
		mods = tcell.ModCtrl
	}
	shiftComboRegex, _ := regexp.Compile(`Shift\+(\w)`)
	if shiftComboRegex.MatchString(s) {
		matches := shiftComboRegex.FindStringSubmatch(s)
		keyString = matches[1]
		//letterIndex := strings.ToLower(matches[1])[0] - 'a'
		//return CtrlCombo(tcell.Key(letterIndex) + tcell.KeyCtrlA)
		mods = tcell.ModShift
	}
	if printableRune, ok := ParsePrintableKey(keyString); ok {
		return LetterCombo(printableRune, mods)
	}
	if tcellKey, ok := ParseNonPrintableKey(keyString); ok {
		return NonPrintableKeyCombo(tcellKey, mods)
	}
	return UIKey{name: "unknown"}
}
func ParsePrintableKey(s string) (rune, bool) {
	trimmedLower := strings.TrimSpace(strings.ToLower(s))
	if trimmedLower == "space" {
		s = " "
	} else if trimmedLower == "enter" {
		s = "\n"
	} else if trimmedLower == "tab" {
		s = "\t"
	}
	if len(s) == 1 {
		return []rune(s)[0], true
	}
	return 0, false
}
func ParseNonPrintableKey(s string) (tcell.Key, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	nonPrintable := map[string]tcell.Key{
		"backtab": tcell.KeyBacktab,
		"esc":     tcell.KeyESC,
		"escape":  tcell.KeyEscape,
		"del":     tcell.KeyDEL,
		"f1":      tcell.KeyF1,
		"f2":      tcell.KeyF2,
		"f3":      tcell.KeyF3,
		"f4":      tcell.KeyF4,
		"f5":      tcell.KeyF5,
		"f6":      tcell.KeyF6,
		"f7":      tcell.KeyF7,
		"f8":      tcell.KeyF8,
		"f9":      tcell.KeyF9,
		"f10":     tcell.KeyF10,
		"f11":     tcell.KeyF11,
		"f12":     tcell.KeyF12,
	}
	if key, ok := nonPrintable[s]; ok {
		return key, true
	}
	return 0, false
}
func FunctionKey(key tcell.Key) UIKey {
	return UIKey{key: key, name: tcell.KeyNames[key]}
}
func Letter(letter rune) UIKey {
	keyName := string(letter)
	key := tcell.KeyRune
	if letter == '\n' {
		keyName = "Enter"
		key = tcell.KeyEnter
		letter = 13
	} else if letter == '\t' {
		keyName = "Tab"
		key = tcell.KeyTAB
		letter = 9
	}

	return UIKey{ch: letter, name: keyName, key: key}
}
func LetterCombo(letter rune, mod tcell.ModMask) UIKey {
	keyName := string(letter)
	key := tcell.KeyRune
	if letter == '\n' {
		keyName = "Enter"
		key = tcell.KeyEnter
		letter = 13
	} else if letter == '\t' {
		keyName = "Tab"
		key = tcell.KeyTAB
		letter = 9
	}
	return UIKey{ch: letter, name: keyName, key: key, mod: mod}
}
func CtrlCombo(key tcell.Key) UIKey {
	return UIKey{ch: rune(key), mod: tcell.ModCtrl, key: key, name: tcell.KeyNames[key]}
}
func ShiftCombo(key tcell.Key) UIKey {
	return UIKey{ch: rune(key), mod: tcell.ModShift, key: key, name: tcell.KeyNames[key]}
}
func NonPrintableKeyCombo(key tcell.Key, mod tcell.ModMask) UIKey {
	//ch := rune(key)
	return UIKey{key: key, name: tcell.KeyNames[key], mod: mod}
}
func (k UIKey) String() string {
	return fmt.Sprintf("UIKey{mod: %d, key: %d, ch: %d, name: %s}", k.mod, k.key, k.ch, k.name)
}
