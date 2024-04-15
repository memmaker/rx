package console

import (
	"RogueUI/foundation"
	"RogueUI/geometry"
	"RogueUI/util"
	"bufio"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"regexp"
	"strconv"
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

func (u *UI) setupCommandTable() {
	u.commandTable = make(map[string]func())

	u.commandTable["quit"] = u.application.Stop

	u.commandTable["inventory"] = u.game.OpenInventory
	u.commandTable["tactics"] = u.game.OpenTacticsMenu
	u.commandTable["character"] = u.ShowCharacterSheet
	u.commandTable["wizard"] = u.game.OpenWizardMenu
	u.commandTable["themes"] = u.OpenThemesMenu
	u.commandTable["log"] = u.ShowLog
	u.commandTable["monsters"] = u.ShowVisibleEnemies
	u.commandTable["items"] = u.ShowVisibleItems
	u.commandTable["help"] = u.ShowHelpScreen
	u.commandTable["log"] = u.ShowLog
	u.commandTable["wiz_ascend"] = u.game.Ascend
	u.commandTable["wiz_descend"] = u.game.Descend

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
	u.commandTable["use"] = u.game.ChooseItemForUseOrZap
	u.commandTable["launch"] = u.game.ChooseItemForMissileLaunch
	u.commandTable["aim"] = u.game.AimedShot
	u.commandTable["quick_shot"] = u.game.QuickShot
	u.commandTable["pickup"] = u.game.PickupItem
	u.commandTable["map_interaction"] = u.game.PlayerInteractWithMap
	u.commandTable["run_direction"] = u.ChooseDirectionForRun
	u.commandTable["wait"] = u.game.Wait
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
func (u *UI) loadKeyMap(filename string) {
	clear(u.keyTable)
	file := util.MustOpen(filename)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.Split(line, "->")
		if len(parts) == 2 {
			key := UIKeyFromString(strings.TrimSpace(parts[0]))
			command := strings.TrimSpace(parts[1])
			u.keyTable[key] = command
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
	if len(s) == 1 {
		return Letter([]rune(s)[0])
	}
	fKeyRegex, _ := regexp.Compile(`F(\d+)`)
	if fKeyRegex.MatchString(s) {
		matches := fKeyRegex.FindStringSubmatch(s)
		fKey, _ := strconv.Atoi(matches[1])
		return FunctionKey(tcell.KeyF1 + tcell.Key(fKey) - 1)
	}
	ctrlComboRegex, _ := regexp.Compile(`Ctrl\+(\w)`)
	if ctrlComboRegex.MatchString(s) {
		matches := ctrlComboRegex.FindStringSubmatch(s)
		letterIndex := strings.ToLower(matches[1])[0] - 'a'
		return CtrlCombo(tcell.Key(letterIndex) + tcell.KeyCtrlA)
	}
	s = strings.ToLower(s)
	nonPrintable := map[string]tcell.Key{
		"enter": tcell.KeyCR,
		"tab":   tcell.KeyTAB,
		"del":   tcell.KeyDEL,
	}
	if key, ok := nonPrintable[s]; ok {
		return NonPrintableKey(key)
	}
	return UIKey{name: "unknown"}
}
func FunctionKey(key tcell.Key) UIKey {
	return UIKey{key: key, name: tcell.KeyNames[key]}
}
func Letter(letter rune) UIKey {
	return UIKey{ch: letter, name: string(letter), key: tcell.KeyRune}
}
func CtrlCombo(key tcell.Key) UIKey {
	return UIKey{ch: rune(key), mod: tcell.ModCtrl, key: key, name: tcell.KeyNames[key]}
}
func NonPrintableKey(key tcell.Key) UIKey {
	return UIKey{ch: rune(key), key: key, name: tcell.KeyNames[key]}
}
func (k UIKey) String() string {
	return fmt.Sprintf("UIKey{mod: %d, key: %d, ch: %d, name: %s}", k.mod, k.key, k.ch, k.name)
}
