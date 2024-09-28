package console

import (
	"RogueUI/d100"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/textiles"
	"image/color"
	"strconv"
	"strings"
)

type SheetMode int

const (
	ModeView SheetMode = iota
	ModeCreate
)

type CharsheetViewer struct {
	*cview.Grid
	charName           string
	sheet              *d100.CharSheet
	close              func()
	nameAgeSexBar      *cview.TextView
	derivedStatsWindow *cview.TextView
	descriptionWindow  *cview.TextView
	statList           *cview.List
	skillList          *cview.List
	charPointsDisplay  *cview.TextView
	traitsList         *cview.List
	buttonBar          *cview.Grid
	mode               SheetMode
	conf               Confirmer

	virtuallySpentSkillPoints map[d100.Skill]int
	virtualFocus              int
}

type Confirmer interface {
	AskForConfirmation(title, message string, choice func(didConfirm bool))
}

func NewCharsheetViewer(name string, sheet *d100.CharSheet, close func()) *CharsheetViewer {
	c := &CharsheetViewer{
		Grid:                      cview.NewGrid(),
		sheet:                     sheet,
		close:                     close,
		charName:                  name,
		virtuallySpentSkillPoints: make(map[d100.Skill]int),
		virtualFocus:              0,
	}
	c.Grid.SetBorder(true)
	c.Grid.SetBorderColor(tcell.ColorGreen)
	c.Grid.SetBackgroundTransparent(false)
	c.Grid.SetBackgroundColor(tcell.ColorDefault)
	c.Grid.SetBorders(false)
	c.Grid.SetGap(0, 1)
	c.Grid.SetOffset(0, 0)
	c.Grid.SetPadding(0, 0, 0, 0)
	c.Grid.SetInputCapture(c.handleInput)
	c.SetMode()
	return c
}

func (c *CharsheetViewer) getVirtuallySpentSkillPoints() int {
	fakeTotal := 0
	for _, v := range c.virtuallySpentSkillPoints {
		fakeTotal += v
	}
	return fakeTotal
}

func (c *CharsheetViewer) SetConfirmer(conf Confirmer) {
	c.conf = conf
}

func (c *CharsheetViewer) SetMode() {
	c.mode = ModeView
	if c.sheet.HasStatPointsToSpend() || c.sheet.GetTagSkillCount() < 3 {
		c.mode = ModeCreate
	}
	c.setupUI()
	c.updateUIFromSheet()
}

func (c *CharsheetViewer) Draw(screen tcell.Screen) {
	w, h := screen.Size()
	if w == 80 || h == 25 {
		c.SetBorder(false)
	} else {
		c.SetBorder(true)
	}
	c.Grid.Draw(screen)
}

func (c *CharsheetViewer) tryAskForConfirmation(msg string, result func(didConfirm bool)) {
	if c.conf == nil {
		result(true)
		return
	}
	c.conf.AskForConfirmation("Sure?", msg, result)
}

func (c *CharsheetViewer) setupUI() {
	nameAgeSexBar := cview.NewTextView()
	nameAgeSexBar.SetBorder(false)
	nameAgeSexBar.SetScrollBarVisibility(cview.ScrollBarNever)
	nameAgeSexBar.SetScrollable(false)

	derivedStatsWindow := cview.NewTextView()
	derivedStatsWindow.SetBorder(true)
	derivedStatsWindow.SetScrollBarVisibility(cview.ScrollBarNever)
	derivedStatsWindow.SetScrollable(false)
	derivedStatsWindow.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
		if action == cview.MouseLeftClick {
			x, y := event.Position()
			// to local coordinates
			startX, startY, _, _ := derivedStatsWindow.GetInnerRect()
			x -= startX
			y -= startY
			// to list item index
			i := y

			if i < 0 || i >= int(d100.VisibleDerivedStatCount) {
				return action, event
			}

			derivedStat := (d100.DerivedStat)(i)

			_, mods := c.sheet.GetDerivedStatWithModInfo(derivedStat)

			derivedString := fmt.Sprintf("%s: %s\n%s", derivedStat.String(), derivedStat.Source(), modifiersToString(mods))

			c.descriptionWindow.SetText(derivedString)
		}
		return action, event
	})

	descriptionWindow := cview.NewTextView()
	descriptionWindow.SetScrollBarVisibility(cview.ScrollBarAuto)
	descriptionWindow.SetScrollable(true)
	descriptionWindow.SetBorder(true)
	descriptionWindow.SetWrap(true)
	descriptionWindow.SetWordWrap(true)

	statList := cview.NewList()
	statList.SetBorder(true)
	statList.ShowSecondaryText(false)
	statList.SetScrollBarVisibility(cview.ScrollBarNever)
	statList.SetHighlightDisabled(true)
	statList.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
		if action == cview.MouseLeftClick {
			x, y := event.Position()
			// to local coordinates
			startX, startY, _, _ := statList.GetInnerRect()
			x -= startX
			y -= startY
			// to list item index
			i := y
			isOnlyInfo := x < 6
			isPlus := x > 10

			if i < 0 || i >= int(d100.StatCount) {
				return action, event
			}

			stat := (d100.Stat)(i)

			_, mods := c.sheet.GetStatWithModInfo(stat)

			c.descriptionWindow.SetText(stat.String() + "\n" + modifiersToString(mods))

			if isOnlyInfo || c.mode == ModeView {
				return action, event
			}
			if isPlus {
				c.sheet.SpendStatPoint(stat)
			} else {
				c.sheet.RefundStatPoint(stat)
			}
			c.updateUIFromSheet()
		}
		return action, event
	})

	skillList := cview.NewList()
	skillList.SetBorder(true)
	skillList.ShowSecondaryText(false)
	skillList.SetScrollBarVisibility(cview.ScrollBarNever)
	skillList.SetTitle("Skills")
	//skillList.SetHighlightFullLine(true)
	skillList.SetHighlightDisabled(true)

	skillList.SetSelectedFunc(func(i int, item *cview.ListItem) {
		skill := d100.Skill(i)
		if c.mode == ModeCreate {
			c.toggleTagSkill(skill)
		}
	})
	if c.sheet.HasSkillPointsToSpend() {
		skillList.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
			if action == cview.MouseLeftClick {
				x, y := event.Position()
				// to local coordinates
				startX, startY, width, _ := skillList.GetInnerRect()
				x -= startX
				y -= startY
				// to list item index
				i := y
				// the last four characters are the plus button
				isPlus := x > width-6

				if i < 0 || i >= int(d100.SkillCount()) {
					return action, event
				}

				skill := (d100.Skill)(i)

				isOnlyInfo := x < 10

				_, mods := c.sheet.GetSkillWithModInfo(skill)

				skillDesc := fmt.Sprintf("%s: %s\n%s", skill.String(), skill.Source(), modifiersToString(mods))
				c.descriptionWindow.SetText(skillDesc)

				if isOnlyInfo {
					return action, event
				}

				if isPlus {
					c.increaseSkill(skill)
				} else {
					c.decreaseSkill(skill)
				}
			}
			return action, event
		})
	} else {
		skillList.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
			if action == cview.MouseLeftClick {
				x, y := event.Position()
				// to local coordinates
				startX, startY, _, _ := skillList.GetInnerRect()
				x -= startX
				y -= startY
				// to list item index
				i := y

				if i < 0 || i >= int(d100.SkillCount()) {
					return action, event
				}

				skill := (d100.Skill)(i)

				_, mods := c.sheet.GetSkillWithModInfo(skill)
				skillDesc := fmt.Sprintf("%s: %s\n%s", skill.String(), skill.Source(), modifiersToString(mods))
				c.descriptionWindow.SetText(skillDesc)
				return action, event
			}
			return action, event
		})
	}

	charPointsDisplay := cview.NewTextView()
	charPointsDisplay.SetBorder(true)
	charPointsDisplay.SetScrollBarVisibility(cview.ScrollBarNever)
	charPointsDisplay.SetScrollable(false)

	traitsList := cview.NewList()
	traitsList.SetBorder(true)
	traitsList.ShowSecondaryText(false)
	traitsList.SetScrollBarVisibility(cview.ScrollBarNever)
	traitsList.SetHighlightDisabled(true)

	buttonBar := cview.NewGrid()
	buttonBar.SetColumns(0, 0)
	buttonBar.SetRows(1)
	buttonBar.SetBorder(false)
	buttonBar.SetGap(0, 1)

	doneButton := cview.NewButton("Done")

	doneButton.SetSelectedFunc(c.onDoneButtonClicked)

	buttonBar.AddItem(doneButton, 0, 0, 1, 2, 0, 0, false)
	//buttonBar.AddItem(cancelButton, 0, 1, 1, 1, 0, 0, false)

	c.Grid.SetRows(1, 9, 5, 9, 1)
	c.Grid.SetColumns(18, 28, 34)

	c.Grid.AddItem(nameAgeSexBar, 0, 0, 1, 2, 0, 0, false)
	c.Grid.AddItem(statList, 1, 0, 1, 1, 0, 0, false)
	c.Grid.AddItem(charPointsDisplay, 2, 0, 1, 1, 0, 0, false)
	c.Grid.AddItem(traitsList, 3, 0, 2, 1, 0, 0, false)
	c.Grid.AddItem(derivedStatsWindow, 1, 1, 2, 1, 0, 0, false)
	c.Grid.AddItem(skillList, 1, 2, 3, 1, 0, 0, false)
	c.Grid.AddItem(descriptionWindow, 3, 1, 2, 1, 0, 0, false)
	c.Grid.AddItem(buttonBar, 4, 2, 1, 1, 0, 0, false)

	c.nameAgeSexBar = nameAgeSexBar
	c.derivedStatsWindow = derivedStatsWindow
	c.descriptionWindow = descriptionWindow
	c.statList = statList
	c.skillList = skillList
	c.charPointsDisplay = charPointsDisplay
	c.traitsList = traitsList
	c.buttonBar = buttonBar
}

func modifiersToString(mods []d100.Modifier) string {
	if len(mods) == 0 {
		return ""
	}
	modStrs := make([]string, len(mods))
	for i, mod := range mods {
		modStrs[i] = mod.Description()
	}
	return strings.Join(modStrs, "\n")
}

func (c *CharsheetViewer) decreaseSkill(skill d100.Skill) {
	if c.virtuallySpentSkillPoints[skill] > 0 {
		c.virtuallySpentSkillPoints[skill]--
	}
	c.updateUIFromSheet()
	c.skillList.SetCurrentItem(int(skill))
}

func (c *CharsheetViewer) increaseSkill(skill d100.Skill) {
	currentSkill := c.sheet.GetUnmodifiedSkill(skill) + c.getVirtuallySpentSkillPointsFor(skill)
	if c.getSkillPointsAvailable() > 0 && currentSkill < d100.SkillCap {
		c.virtuallySpentSkillPoints[skill]++
	}
	c.updateUIFromSheet()
	c.skillList.SetCurrentItem(int(skill))
}

func (c *CharsheetViewer) toggleTagSkill(skill d100.Skill) {
	if c.sheet.IsTagSkill(skill) {
		c.sheet.UntagSkill(skill)
	} else {
		c.sheet.TagSkill(skill)
	}
	c.updateUIFromSheet()
	c.skillList.SetCurrentItem(int(skill))
}

func (c *CharsheetViewer) Close() {
	c.close()
}
func (c *CharsheetViewer) onDoneButtonClicked() {
	if c.mode == ModeCreate && (!c.sheet.HasStatPointsToSpend() && c.sheet.GetTagSkillCount() == 3) {
		c.tryAskForConfirmation("Are you sure you want to finish character creation?\nYou won't be able to change your character's stats or tag skills after this.", func(didConfirm bool) {
			if didConfirm {
				c.Close()
			}
		})
		return
	}

	if c.getVirtuallySpentSkillPoints() > 0 {
		c.tryAskForConfirmation("Confirm spending skill points?", func(didConfirm bool) {
			if didConfirm {
				for skill, points := range c.virtuallySpentSkillPoints {
					c.sheet.SpendSkillPoints(skill, points)
				}
				c.Close()
			}
		})
		return
	}

	c.Close()
}
func (c *CharsheetViewer) updateUIFromSheet() {
	c.nameAgeSexBar.SetText(c.charName)

	var infoLines []string
	if c.mode == ModeCreate {
		tableRowsForCreateInfo := []fxtools.TableRow{
			{Columns: []string{"Stat Points:", strconv.Itoa(c.sheet.GetStatPointsToSpend())}},
			{Columns: []string{"Tag Skills:", fmt.Sprintf("%d/3", c.sheet.GetTagSkillCount())}},
		}
		infoLines = fxtools.TableLayout(tableRowsForCreateInfo, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignRight})
	} else if c.sheet.HasSkillPointsToSpend() {
		skillPointsToSpend := c.getSkillPointsAvailable()
		infoLines = []string{
			fmt.Sprintf("Skill Points: %d", skillPointsToSpend),
		}
	} else {
		tableRowsForLevelInfo := []fxtools.TableRow{
			{Columns: []string{"Level:", strconv.Itoa(c.sheet.GetLevel())}},
			{Columns: []string{"XP:", strconv.Itoa(c.sheet.GetCurrentXP())}},
			{Columns: []string{"Next:", strconv.Itoa(c.sheet.GetXPNeededForNextLevel())}},
		}
		infoLines = fxtools.TableLayout(tableRowsForLevelInfo, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignRight})
	}
	c.charPointsDisplay.SetText(strings.Join(infoLines, "\n"))

	c.statList.Clear()
	for i := 0; i < int(d100.StatCount); i++ {
		stat := (d100.Stat)(i)
		statVal, mods := c.sheet.GetStatWithModInfo(stat)
		statCC := "[-:-:-]"
		if c.mode == ModeCreate {
			if statVal < 5 {
				statCC = "[red:black:]"
			} else if statVal > 5 {
				statCC = "[green:black:]"
			}
		}

		statName := stat.ToShortString()

		statAdjust := " "
		if len(mods) > 0 {
			statAdjust = "*"
		}

		statLine := fmt.Sprintf("%s:%s%s%d[-:-:-]", statName, statAdjust, statCC, statVal)

		if c.mode == ModeCreate {
			plusButton := cview.Escape(fmt.Sprintf("[+]"))
			minusButton := cview.Escape(fmt.Sprintf("[-]"))
			minusCC := textiles.RGBAToColorCodes(color.RGBA{R: 255, G: 255, B: 255, A: 255}, color.RGBA{40, 0, 0, 255})
			plusCC := textiles.RGBAToColorCodes(color.RGBA{R: 255, G: 255, B: 255, A: 255}, color.RGBA{0, 40, 0, 255})
			adjustButtons := fmt.Sprintf(" %s%s[-:-:-] %s%s[-:-:-]", minusCC, minusButton, plusCC, plusButton)
			if statVal < 10 {
				statLine = fmt.Sprintf("%s %s", statLine, adjustButtons)
			} else {
				statLine = fmt.Sprintf("%s%s", statLine, adjustButtons)
			}
		}

		listItem := cview.NewListItem(statLine)
		c.statList.AddItem(listItem)
	}

	skillRows := make([]fxtools.TableRow, d100.SkillCount())
	for i := 0; i < int(d100.SkillCount()); i++ {
		skill := (d100.Skill)(i)

		skillVal, modifiers := c.sheet.GetSkillWithModInfo(skill)

		skillVal += c.getVirtuallySpentSkillPointsFor(skill)

		isTagged := c.sheet.IsTagSkill(skill)

		valueColumn := fmt.Sprintf("%d%%", skillVal)

		if len(modifiers) > 0 {
			valueColumn = "*" + valueColumn
		}

		skillColumns := []string{
			skill.String(),
			valueColumn,
		}

		if c.mode == ModeCreate {
			tagIcon := cview.Escape("[ ]")
			if isTagged {
				tagIcon = cview.Escape("[X]")
			}
			skillColumns = append(skillColumns, tagIcon)
		} else if c.sheet.HasSkillPointsToSpend() {
			plusButton := cview.Escape(fmt.Sprintf("[+]"))
			minusButton := cview.Escape(fmt.Sprintf("[-]"))
			minusCC := textiles.RGBAToColorCodes(color.RGBA{R: 255, G: 255, B: 255, A: 255}, color.RGBA{40, 0, 0, 255})
			plusCC := textiles.RGBAToColorCodes(color.RGBA{R: 255, G: 255, B: 255, A: 255}, color.RGBA{0, 40, 0, 255})
			divider := " "
			if c.getVirtuallySpentSkillPointsFor(skill) > 0 {
				divider = "!"
			}
			adjustButtons := fmt.Sprintf(" %s%s[-:-:-]%s%s%s[-:-:-]", minusCC, minusButton, divider, plusCC, plusButton)
			skillColumns = append(skillColumns, adjustButtons)
		}

		skillRows[i] = fxtools.TableRow{
			Columns: skillColumns,
		}
	}

	alignments := []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignRight}
	if c.mode == ModeCreate || c.sheet.HasSkillPointsToSpend() {
		alignments = append(alignments, fxtools.AlignRight)
	}
	c.skillList.Clear()
	skillsAsTable := fxtools.TableLayout(skillRows, alignments)
	for i, row := range skillsAsTable {
		skill := (d100.Skill)(i)
		if c.sheet.IsTagSkill(skill) {
			skillCC := "[-:-:]"
			if c.mode == ModeCreate {
				skillCC = "[black:yellow:]"
			} else {
				skillCC = "[yellow:black:]"
			}
			row = fmt.Sprintf("%s%s", skillCC, row)
		}
		listItem := cview.NewListItem(row)
		c.skillList.AddItem(listItem)
	}

	derivedRows := make([]fxtools.TableRow, d100.VisibleDerivedStatCount)
	for i := 0; i < int(d100.VisibleDerivedStatCount); i++ {
		derivedStat := (d100.DerivedStat)(i)
		derivedStatVal := c.sheet.GetDerivedStat(derivedStat)
		derivedRows[i] = fxtools.TableRow{
			Columns: []string{
				derivedStat.String(),
				strconv.Itoa(derivedStatVal),
			},
		}
	}
	derivedStr := strings.Join(fxtools.TableLayout(derivedRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignRight}), "\n")
	c.derivedStatsWindow.SetText(strings.TrimSpace(derivedStr))
	/*
	   c.statList.AddItem()
	   c.skillList.AddItem()

	   c.charPointsDisplay.SetText(c.sheet.CharPointsDisplay)

	   c.traitsList.SetItems(c.sheet.TraitsList)

	*/
	c.traitsList.Clear()
	c.traitsList.AddItem(cview.NewListItem("No Traits"))
	c.traitsList.AddItem(cview.NewListItem("No Perks"))
}

func (c *CharsheetViewer) getSkillPointsAvailable() int {
	skillPointsToSpend := c.sheet.GetSkillPointsToSpend()
	skillPointsToSpend -= c.getVirtuallySpentSkillPoints()
	return skillPointsToSpend
}

func (c *CharsheetViewer) getVirtuallySpentSkillPointsFor(skill d100.Skill) int {
	if v, ok := c.virtuallySpentSkillPoints[skill]; ok {
		isTagged := c.sheet.IsTagSkill(skill)
		if isTagged {
			return v * 2
		}
		return v
	}
	return 0
}

func (c *CharsheetViewer) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEscape {
		c.onDoneButtonClicked()
		return nil
	}
	if c.mode == ModeView && !c.sheet.HasSkillPointsToSpend() {
		return nil
	}
	switch event.Key() {
	case tcell.KeyDown:
		c.moveSelectionVertically(1)
	case tcell.KeyUp:
		c.moveSelectionVertically(-1)
	case tcell.KeyRight:
		c.moveSelectionHorizontally(1)
	case tcell.KeyLeft:
		c.moveSelectionHorizontally(-1)
	case tcell.KeyEnter:
		c.confirmSelection()
	}
	return nil
}

func (c *CharsheetViewer) moveSelectionVertically(direction int) {
	if c.mode == ModeCreate {
		// stats & skills
		if c.virtualFocus == 0 { // nothing selected
			if direction > 0 {
				c.focusStats(0)
			} else {
				c.focusStats(int(d100.StatCount) - 1)
			}
		} else if c.virtualFocus == 1 { // stats selected
			newIndex := (c.statList.GetCurrentItemIndex() + direction) % int(d100.StatCount)
			c.statList.SetCurrentItem(newIndex)
		} else if c.virtualFocus == 2 { // skills selected
			newIndex := (c.skillList.GetCurrentItemIndex() + direction) % int(d100.SkillCount())
			c.skillList.SetCurrentItem(newIndex)
		}
		return
	}
	if c.virtualFocus == 0 {
		if direction > 0 {
			c.focusSkills(0)
		} else {
			c.focusSkills(int(d100.SkillCount()) - 1)
		}
	} else {
		newIndex := (c.skillList.GetCurrentItemIndex() + direction) % int(d100.SkillCount())
		c.skillList.SetCurrentItem(newIndex)
		return
	}
}

func (c *CharsheetViewer) moveSelectionHorizontally(direction int) {
	if c.mode == ModeCreate {
		if c.virtualFocus == 0 {
			if direction > 0 {
				// skills
				c.focusSkills(0)
			} else {
				// stats
				c.focusStats(0)
			}
		} else if c.virtualFocus == 1 {
			// stats selected
			if direction > 0 {
				// skills
				c.focusSkills(0)
			} else {
				index := c.statList.GetCurrentItemIndex()
				selectedStat := d100.Stat(index)
				c.sheet.RefundStatPoint(selectedStat)
				c.updateUIFromSheet()
				c.statList.SetCurrentItem(index)
			}
		} else if c.virtualFocus == 2 {
			// skills selected
			if direction <= 0 {
				// stats
				c.focusStats(0)
			}
		}

		return
	}

	if c.virtualFocus == 0 {
		c.focusSkills(0)
	} else {
		selectedSkill := d100.Skill(c.skillList.GetCurrentItemIndex())
		if direction > 0 {
			c.increaseSkill(selectedSkill)
		} else {
			c.decreaseSkill(selectedSkill)
		}
	}
}

func (c *CharsheetViewer) focusStats(selectedIndex int) {
	c.virtualFocus = 1
	c.statList.SetHighlightDisabled(false)
	c.statList.SetCurrentItem(selectedIndex)

	c.skillList.SetHighlightDisabled(true)
}
func (c *CharsheetViewer) focusSkills(index int) {
	c.virtualFocus = 2
	c.skillList.SetHighlightDisabled(false)
	c.skillList.SetCurrentItem(index)

	c.statList.SetHighlightDisabled(true)
}

func (c *CharsheetViewer) confirmSelection() {
	if c.mode == ModeCreate {
		if c.virtualFocus == 1 {
			index := c.statList.GetCurrentItemIndex()
			selectedStat := d100.Stat(index)
			c.sheet.SpendStatPoint(selectedStat)
			c.updateUIFromSheet()
			c.statList.SetCurrentItem(index)
		} else if c.virtualFocus == 2 {
			index := c.skillList.GetCurrentItemIndex()
			selectedSkill := d100.Skill(index)
			c.toggleTagSkill(selectedSkill)
		}
	}
}
