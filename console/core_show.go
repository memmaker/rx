package console

import (
    "contractor/foundation"
    "fmt"
    "github.com/gdamore/tcell/v2"
    "github.com/memmaker/go/cview"
    "github.com/memmaker/go/fxtools"
    "github.com/memmaker/go/textiles"
    "image/color"
    "math/rand"
    "strconv"
    "strings"
)

// Updating stuff already on the screen

func (u *UI) UpdateLogWindow() {
    textColor := u.uiTheme.GetUIColor(UIColorUIForeground)
    hiLiteColor := u.uiTheme.GetUIColor(UIColorTextForegroundHighlighted)
    // last 100 lines
    logMessages := u.game.GetLog()
    if len(logMessages) > 100 {
        logMessages = logMessages[len(logMessages)-100:]
    }
    var asColoredStrings []string
    for i, message := range logMessages {
        fadePercent := fxtools.Clamp(0.2, 1.0, float64(i+1)/float64(len(logMessages)))
        asColoredStrings = append(asColoredStrings, ToColoredText(message, fadePercent, textColor, hiLiteColor))
    }

    u.setColoredText(u.messageLabel, strings.Join(asColoredStrings, "\n"))
    u.messageLabel.ScrollToEnd()
}

func (u *UI) UpdateInventory() {
    items := u.game.GetInventoryForUI()
    if len(items) == 0 {
        u.rightPanel.Clear()
        return
    }
    longest := longestInventoryLineWithoutColorCodes(items)

    var getItemName func(item foundation.Item, isEquipped bool) string

    if !u.isRightPanelWidthAtLeast(longest) {
        if u.getRightPanelWidth() == 0 {
            return
        }
        getItemName = func(item foundation.Item, isEquipped bool) string {
            itemIcon := item.GetIcon().WithFg(u.uiTheme.GetInventoryItemColor(item.GetCategory())).WithBg(u.uiTheme.GetUIColor(UIColorUIBackground))
            if isEquipped {
                itemIcon = itemIcon.Reversed()
            }
            iconString := IconAsString(itemIcon)
            return iconString
        }
    } else {
        getItemName = func(item foundation.Item, isEquipped bool) string {
            nameWithColorsAndShortcut := item.InventoryNameWithColorsAndShortcut(u.uiTheme.GetInventoryItemColorCode(item.GetCategory()))
            if isEquipped {
                nameWithColorsAndShortcut = nameWithColorsAndShortcut[:2] + "+" + nameWithColorsAndShortcut[3:]
            }
            appendString := RightPadColored(nameWithColorsAndShortcut, longest)
            return appendString
        }
    }

    var asString []string
    for _, item := range items {
        isEquipped := u.game.IsEquipped(item)
        appendString := getItemName(item, isEquipped)
        asString = append(asString, appendString)
    }
    u.rightPanel.SetTextAlign(cview.AlignRight)
    u.rightPanel.SetText("\n" + strings.Join(asString, "\n"))
}

func (u *UI) UpdateVisibleActors() {
    visibleEnemies := u.game.GetVisibleActors()
    //longest := longestInventoryLineWithoutColorCodes(visibleEnemies)
    var asString []string
    for _, enemy := range visibleEnemies {
        icon := u.getIconForActor(enemy)
        iconBackground := color.RGBA{A: 255}
        if u.game.IsActorHostileTowardsPlayer(enemy) {
            iconBackground = u.uiTheme.GetColorByName("Red_13")
        } else if u.game.IsActorAlliedWithPlayer(enemy) {
            iconBackground = u.uiTheme.GetColorByName("Green_6")
        }
        iconColor := textiles.RGBAToColorCodes(icon.Fg, iconBackground)
        iconString := fmt.Sprintf("%s%s[-:-]", iconColor, string(icon.Char))
        hp, hpMax := enemy.GetHitPoints(), enemy.GetHitPointsMax()
        asPercent := float64(hp) / float64(hpMax)

        hallucinating := u.isPlayerHallucinating()
        if hallucinating {
            asPercent = rand.Float64()
        }
        barIcon := '*'
        if enemy.HasFlag(foundation.FlagSleep) {
            barIcon = 'z'
        }
        hpBarString := fmt.Sprintf("[%s]", RuneBarFromPercent(barIcon, asPercent, 5, [3]color.RGBA{u.uiTheme.GetColorByName("Green_4"), u.uiTheme.GetColorByName("Yellow_4"), u.uiTheme.GetColorByName("Red_4")}))
        name := enemy.Name()

        canAttack := u.game.CanActorAttackNextTurn(enemy)
        debugInfo := "" //fmt.Sprintf(" (%d, TTM: %d, TFA: %d)", enemy.TimeEnergy(), enemy.TimeNeededForMovement(), enemy.TimeNeededForActions())
        if canAttack {
            debugInfo = " (A!)"
        }
        enemyLine := fmt.Sprintf(" %s %s %s %s ", iconString, hpBarString, name, debugInfo)
        if enemy == u.hoveredActor {
            bgCode := textiles.RGBAToBgColorCode(u.uiTheme.GetColorByName("forest_green_13"))
            enemyLine = fmt.Sprintf("%s %s%s %s %s %s [:-]", bgCode, iconString, bgCode, hpBarString, name, debugInfo)
            asString = append([]string{enemyLine}, asString...)
        } else {
            asString = append(asString, enemyLine)
        }
    }
    u.lowerRightPanel.SetText(strings.Join(asString, "\n"))
    u.lowerRightPanel.ScrollToBeginning()
}

func (u *UI) isStatusBarMultiLine() bool {
    _, h := u.application.GetScreenSize()
    _, hNeeded := u.settings.GetMinTerminalSize()
    return h >= hNeeded+1
}

func (u *UI) UpdateStats() {
    statusValues := u.game.GetHudStats()
    flags := u.game.GetHudFlags()
    if len(statusValues) == 0 {
        return
    }

    equippedItem, isEquipped := u.game.GetItemInMainHand()

    multiLine := u.isStatusBarMultiLine()

    itemName := "| bare hands |"
    if isEquipped {
        if multiLine {
            itemName = "| " + equippedItem.LongNameWithColors(textiles.RGBAToFgColorCode(u.uiTheme.GetInventoryItemColor(equippedItem.GetCategory()))) + " |"
        } else {
            itemName = "| " + equippedItem.ShortNameWithColors(textiles.RGBAToFgColorCode(u.uiTheme.GetInventoryItemColor(equippedItem.GetCategory()))) + " |"
        }

    }

    turns := statusValues.GetInt(foundation.HudTurnsTaken)

    statusBarLine := u.getLowerStatusBar(statusValues, flags, multiLine, itemName)

    if multiLine {
        hp := statusValues.GetInt(foundation.HudHitPoints)
        hpMax := statusValues.GetInt(foundation.HudHitPointsMax)

        playerBar := FullColorBarFromPercent(hp, hpMax, 11, [3]color.RGBA{u.uiTheme.GetColorByName("Green_4"), u.uiTheme.GetColorByName("Yellow_4"), u.uiTheme.GetColorByName("dark_red_1")}, u.uiTheme.GetColorByName("dark_gray_3"), u.uiTheme.GetColorByName("White"), u.uiTheme.GetColorByName("Black"))
        hpBarStr := fmt.Sprintf("HP [%s]", playerBar)

        apCurrent := statusValues.GetInt(foundation.HudActionPoints)
        apMax := statusValues.GetInt(foundation.HudActionPointsMax)

        // display as bar

        actionBarContent := RuneBarWithColor(
            '!',
            apCurrent,
            apMax,
            u.uiTheme.GetColorByName("neon_blue_1"),
            u.uiTheme.GetColorByName("neon_blue_3"),
            u.uiTheme.GetColorByName("dark_gray_3"),
        )

        apBarStr := fmt.Sprintf("AP [%s]", actionBarContent)

        longFlags := PlayerFlagStringLong(flags)

        width, _ := u.application.GetScreenSize()

        mapFriendlyName := u.game.GetMapDisplayName()

        topStatusLine := fmt.Sprintf("%s %s %s | %s | T: %d", hpBarStr, apBarStr, longFlags, mapFriendlyName, turns)

        if cview.TaggedStringWidth(topStatusLine) > width {
            shortFlags := PlayerFlagStringShort(flags)
            topStatusLine = fmt.Sprintf("%s %s %s", hpBarStr, apBarStr, shortFlags)
        }

        topStatusLine = expandToWidth(topStatusLine, width)
        statusBarLine = fmt.Sprintf("%s\n%s", topStatusLine, statusBarLine)
    }

    u.statusBar.SetText(statusBarLine)

    if !u.isAnimationFrame {
        u.lastHudStats = statusValues
    }
}

func (u *UI) ShowGiveAndTakeContainer(leftName string, leftItems []foundation.Item, rightName string, rightItems []foundation.Item, transferToLeft func(itemTaken foundation.Item, stackCount int), transferToRight func(itemTaken foundation.Item, stackCount int), takeAll func()) {
    leftPanel := "leftModal"
    rightPanel := "rightModal"
    var leftMenuItems []foundation.MenuItem
    var rightMenuItems []foundation.MenuItem
    var leftMenuLabels []string
    var rightMenuLabels []string
    if len(leftItems) > 0 {
        leftMenuLabels = u.uiTheme.menuLabelsFor(leftItems)
    }
    if len(rightItems) > 0 {
        rightMenuLabels = u.uiTheme.menuLabelsFor(rightItems)
    }

    closeContainer := func() {
        u.pages.RemovePanel(leftPanel)
        u.pages.RemovePanel(rightPanel)
        u.popFocus()
    }

    for index, i := range leftItems {
        item := i
        leftMenuItems = append(leftMenuItems, foundation.MenuItem{
            Name: leftMenuLabels[index],
            Action: func() {
                closeContainer()
                if item.IsMultipleStacks() {
                    u.openAmountWidget(item.Name(), item.GetStackSize(), func(amount int) {
                        transferToRight(item, amount)
                    })
                } else {
                    transferToRight(item, 1)
                }
                u.application.QueueUpdateDraw(u.UpdateLogWindow)
            },
        })
    }
    for index, i := range rightItems {
        item := i
        rightMenuItems = append(rightMenuItems, foundation.MenuItem{
            Name: rightMenuLabels[index],
            Action: func() {
                closeContainer()
                if item.IsMultipleStacks() {
                    u.openAmountWidget(item.Name(), item.GetStackSize(), func(amount int) {
                        transferToLeft(item, amount)
                    })
                } else {
                    transferToLeft(item, 1)
                }
                u.application.QueueUpdateDraw(u.UpdateLogWindow)
            },
        })
    }

    leftMenu, longestItemLeft := u.createSimpleMenu(leftPanel, leftMenuItems)
    longestItemLeft = max(longestItemLeft, len(leftName))
    leftWidth := longestItemLeft + 2

    leftMenu.SetTitle(leftName)
    leftMenu.SetSelectedFocusOnly(true)

    if len(rightItems) > 0 {
        keyForTakeAll := u.GetKeysForCommandAsString(KeyLayerMain, "pickup")
        u.Print(foundation.HiLite("Press %s to take all items", keyForTakeAll))
    }

    rightMenu, longestItemRight := u.createSimpleMenu(rightPanel, rightMenuItems)
    longestItemRight = max(longestItemRight, len(rightName))
    rightWidth := longestItemRight + 2
    rightMenu.SetTitle(rightName)
    rightMenu.ShowFocus(true)
    rightMenu.SetSelectedFocusOnly(true)

    screenWidth, screenHeight := u.application.GetScreen().Size()

    height := min(screenHeight-4, max(len(leftItems)+2, len(rightItems)+2))

    leftHasScrollbar := len(leftItems) > height-2
    rightHasScrollbar := len(rightItems) > height-2

    centerGap := 4
    equalWidth := max(leftWidth, rightWidth)
    if screenWidth <= (equalWidth*2)+centerGap+4 {
        centerGap = 0
    }
    //borderPadding := 2

    //screenRemaining := screenWidth - centerGap - (2 * borderPadding)
    leftWidth = equalWidth
    rightWidth = equalWidth

    if leftHasScrollbar {
        leftWidth += 2
    }

    if rightHasScrollbar {
        rightWidth += 2
    }

    halfCenterGap := centerGap / 2
    centerScreen := screenWidth / 2

    leftListStart := centerScreen - leftWidth - halfCenterGap
    rightListStart := leftListStart + leftWidth + centerGap
    leftMenu.ShowFocus(true)
    leftMenu.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
        if action == cview.MouseRightClick {
            if !leftMenu.InRect(event.Position()) && !rightMenu.InRect(event.Position()) {
                u.application.QueueUpdateDraw(closeContainer)
                return action, nil
            }
        }
        if action == cview.MouseMove {
            if leftMenu.InRect(event.Position()) {
                u.application.SetFocus(leftMenu)
            } else if rightMenu.InRect(event.Position()) {
                u.application.SetFocus(rightMenu)
            }
        }
        return action, event
    })
    leftMenu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if event.Key() == tcell.KeyEscape {
            closeContainer()
            return nil
        }
        uiKey := toUIKey(event)
        command := u.getCommandForKey(uiKey)
        if command == "west" || command == "east" {
            u.application.SetFocus(rightMenu)
            return nil
        } else if command == "wait" {
            closeContainer()
            return nil
        } else if command == "north" {
            return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
        } else if command == "south" {
            return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
        }
        return event
    })
    rightMenu.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
        if action == cview.MouseRightClick {
            if !leftMenu.InRect(event.Position()) && !rightMenu.InRect(event.Position()) {
                u.application.QueueUpdateDraw(closeContainer)
                return action, nil
            }
        }
        if action == cview.MouseMove {
            if leftMenu.InRect(event.Position()) {
                u.application.SetFocus(leftMenu)
            } else if rightMenu.InRect(event.Position()) {
                u.application.SetFocus(rightMenu)
            }
        }
        return action, event
    })
    rightMenu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if event.Key() == tcell.KeyEscape {
            closeContainer()
            return nil
        }
        uiKey := toUIKey(event)
        command := u.getCommandForKey(uiKey)
        if command == "pickup" {
            closeContainer()
            takeAll()
            u.application.QueueUpdateDraw(u.UpdateLogWindow)
            return nil
        } else if command == "west" || command == "east" {
            u.application.SetFocus(leftMenu)
            return nil
        } else if command == "wait" {
            closeContainer()
            return nil
        } else if command == "north" {
            return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
        } else if command == "south" {
            return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
        }
        return event
    })

    // place both lists vertically centered
    // but side by side

    leftMenu.SetRect(leftListStart, 2, leftWidth, height)
    rightMenu.SetRect(rightListStart, 2, rightWidth, height)

    u.pages.AddPanel("leftModal", leftMenu, false, true)
    u.pages.AddPanel("rightModal", rightMenu, false, true)

    u.pushFocus()

    u.application.SetBeforeFocusFunc(nil)

    if len(rightItems) > 0 {
        u.application.SetFocus(rightMenu)
    } else {
        u.application.SetFocus(leftMenu)
    }

    u.application.SetBeforeFocusFunc(func(p cview.Primitive) bool {
        if p == leftMenu || p == rightMenu {
            return true
        }
        return false
    })
}

func (u *UI) ShowTakeOnlyContainer(name string, containedItems []foundation.Item, transfer func(item foundation.Item)) {
    var menuItems []foundation.MenuItem
    menuLabels := u.uiTheme.menuLabelsFor(containedItems)
    for index, i := range containedItems {
        item := i
        menuItems = append(menuItems, foundation.MenuItem{
            Name: menuLabels[index],
            Action: func() {
                transfer(item)
                u.application.QueueUpdateDraw(u.UpdateLogWindow)
            },
            CloseMenus: true,
        })
    }

    menu := u.openSimpleMenu(menuItems, nil)
    menu.SetTitle(name)
    keyForTakeAll := u.GetKeysForCommandAsString(KeyLayerMain, "pickup")
    u.Print(foundation.HiLite("Press %s to take all items", keyForTakeAll))
    originalCapture := menu.GetInputCapture() // will manage pressing escape
    menu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        uiKey := toUIKey(event)
        command := u.getCommandForKey(uiKey)
        if command == "pickup" {
            for _, item := range containedItems {
                transfer(item)
            }
            originalCapture(EscapeKeyEvent())
            u.application.QueueUpdateDraw(u.UpdateLogWindow)
            return nil
        }
        return originalCapture(event)
    })
}

func (u *UI) OpenKeyMapper(layer KeyLayer) {
    var commandMenu []foundation.MenuItem

    for key, c := range u.keyTable[layer] {
        command := c
        line := fmt.Sprintf("%s - %s", key.name, command)
        commandMenu = append(commandMenu, foundation.MenuItem{
            Name: line,
            Action: func() {
                u.remapCommand(layer, command)
                u.OpenKeyMapper(layer)
            },
        })
    }

    u.OpenMenu(commandMenu)
}

func (u *UI) remapCommand(layer KeyLayer, command string) {
    u.Print(foundation.Msg("Press the key you want to bind to this command"))
    key := u.getPressedKey()
    u.keyTable[layer][key] = command
    u.Print(foundation.Msg(fmt.Sprintf("Bound %s to %s", key.name, command)))
}

func (u *UI) ShowHelpScreen() {
    u.openDirectoryAsTopics("help")
}

// Open Naming

func (u *UI) OpenVendorMenu(title string, itemsForSale []foundation.Item, buyItem func(ui foundation.Item, amount int, price int), inspect func(item foundation.Item), onClose func()) {
    var menuItems []foundation.MenuItem
    var tableRows []fxtools.TableRow
    for _, i := range itemsForSale {
        item := i
        price := i.GetPrice()
        itemName := item.InventoryNameWithColors(u.uiTheme.GetInventoryItemColorCode(item.GetCategory()))
        tableRows = append(tableRows, fxtools.NewTableRow(itemName, fmt.Sprintf("%d sat", price)))
    }
    rendered := fxtools.TableLayoutLastRight(tableRows)
    for index, line := range rendered {
        item := itemsForSale[index]
        tooExpensive := item.GetPrice() > u.game.PlayerGold()
        if tooExpensive {
            line = fmt.Sprintf("%s%s[-]", textiles.RGBAToFgColorCode(u.uiTheme.palette.Get("dark_gray_3")), string(cview.StripTags([]byte(line), true, true)))
        }
        menuItems = append(menuItems, foundation.MenuItem{
            Name: line,
            Action: func() {
                if tooExpensive {
                    u.Print(foundation.Msg("You can't afford that."))
                    return
                }
                if item.IsMultipleStacks() {
                    u.openAmountWidget(item.Name(), item.GetStackSize(), func(amount int) {
                        buyItem(item, amount, item.GetPrice()*amount)
                    })
                } else {
                    buyItem(item, 1, item.GetPrice())
                }
            },
            CloseMenus: !tooExpensive,
        })
    }
    menu := u.openSimpleMenu(menuItems, onClose)
    menu.SetTitle(title)
    origCapture := menu.GetInputCapture()
    menu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        uiKey := toUIKey(event)
        command := u.getCommandForKey(uiKey)
        if command == "run_direction" {
            index := menu.GetCurrentItemIndex()
            if index >= 0 && index < len(itemsForSale) {
                inspect(itemsForSale[index])
                return nil
            }
        }
        if origCapture != nil {
            return origCapture(event)
        }
        return event
    })
}

func (u *UI) ShowTextFile(fileName string) {
    lines := fxtools.ReadFile(fileName)
    u.OpenTextWindow(lines)
}
func (u *UI) OpenTextWindow(description string) {
    u.openTextModal(description)
}

func (u *UI) ShowTextFileFullscreen(filename string, onClose func()) {
    lines := fxtools.ReadFileAsLines(filename)
    textView := cview.NewTextView()
    textView.SetBorder(false)
    u.setColoredText(textView, strings.Join(lines, "\n"))

    panelName := "main"
    if u.pages.HasPanel("main") {
        panelName = "fullscreen"
    }

    textView.SetInputCapture(u.popOnAnyKeyWithNotification(panelName, onClose))
    u.pages.AddPanel(panelName, textView, true, true)
    u.pages.ShowPanel(panelName)
    u.application.SetFocus(textView)
}

func (u *UI) openTextModal(description string) *cview.TextView {
    panelName := "textModal"
    textView := u.newTextModal(description)
    w, h := widthAndHeightFromString(description)
    u.makeCenteredModal(panelName, textView, w, h)
    originalInputCapture := textView.GetInputCapture()
    textView.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
        if action == cview.MouseLeftClick || action == cview.MouseRightClick {
            u.popPanel(panelName)
            return cview.MouseLeftClick, nil
        }
        return action, event
    })
    textView.SetInputCapture(u.directionalWrapper(func(event *tcell.EventKey) *tcell.EventKey {
        if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyEnter {
            u.popPanel(panelName)
            return nil
        }
        if originalInputCapture != nil {
            return originalInputCapture(event)
        }
        return event
    }))
    return textView
}

func (u *UI) newTextModal(description string) *cview.TextView {
    textView := cview.NewTextView()
    textView.SetWrap(true)
    textView.SetWrapWidth(0)
    textView.SetWordWrap(true)
    textView.SetBorder(true)

    textView.SetTextColor(toTcellColor(u.uiTheme.GetColorByName("light_gray_5")))
    textView.SetBackgroundColor(toTcellColor(u.uiTheme.GetColorByName("black")))

    textView.SetBorderColor(u.uiTheme.GetUIColorForTcell(UIColorBorderForeground))

    u.setColoredText(textView, description)
    return textView
}

// Show Naming

func (u *UI) ShowItemOverlay() {
    listOfItems := u.game.GetVisibleItems()

    if len(listOfItems) == 0 {
        u.Print(foundation.Msg("No items in sight"))
        return
    }
    u.mapOverlay.ClearAll()

    for _, items := range listOfItems {
        u.mapOverlay.TryAddOverlay(items.Position(), items.Name(), u.GetMapWindowGridSize(), u.game.IsSomethingInterestingAtLoc)
    }
}

func (u *UI) ShowVisibleActors() {
    listOfEnemies := u.game.GetVisibleActors()
    if len(listOfEnemies) == 0 {
        u.Print(foundation.Msg("No actors in sight"))
        return

    }
    tableRows := make([]fxtools.TableRow, len(listOfEnemies)+1)
    header := fxtools.NewTableRow(
        "Icon",
        "Name",
        "HP",
        "Dmg",
        "Armor",
    )
    tableRows[0] = header
    for i, enemy := range listOfEnemies {
        row := fxtools.NewTableRow(
            string(u.getIconForActor(enemy).Char),
            enemy.Name(),
            strconv.Itoa(enemy.GetHitPoints()),
            enemy.GetMainHandDamageAsString(),
            enemy.GetArmorProtectionString(),
        )
        tableRows[i+1] = row
    }
    layout := fxtools.TableLayout(tableRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft, fxtools.AlignRight, fxtools.AlignRight, fxtools.AlignRight})
    u.OpenTextWindow(strings.Join(layout, "\n"))
}

func (u *UI) ShowVisibleItems() {
    listOfItems := u.game.GetVisibleItems()
    if len(listOfItems) == 0 {
        u.Print(foundation.Msg("No items in sight"))
        return
    }
    var infoTexts strings.Builder
    infoTexts.WriteString("Items in sight:\n")
    longestLine := 0
    for i, item := range listOfItems {
        info := item.InventoryNameWithColors(u.uiTheme.GetInventoryItemColorCode(item.GetCategory()))
        icon := item.GetIcon()
        char := icon.Char
        info = fmt.Sprintf(" %s%c[-:-] - %s", textiles.RGBAToColorCodes(icon.Fg, icon.Bg), char, info)
        longestLine = max(longestLine, cview.TaggedStringWidth(info))
        infoTexts.WriteString(info)
        if i < len(listOfItems)-1 {
            infoTexts.WriteString("\n")
        }
    }
    if u.isRightPanelWidthAtLeast(longestLine) {
        u.rightPanel.SetTextAlign(cview.AlignLeft)
        u.rightPanel.SetText(infoTexts.String())
    } else {
        u.OpenTextWindow(infoTexts.String())
    }
}

func (u *UI) ShowMonsterInfo(monster foundation.ActorForUI) {
    monsterInfo := monster.GetDetailInfo()
    u.openTextModal(monsterInfo)
}
func (u *UI) ShowLog() {
    logLines := u.game.GetLog()
    if len(logLines) == 0 {
        u.Print(foundation.Msg("No log entries"))
        return
    }
    textColor := u.uiTheme.GetUIColor(UIColorUIForeground)
    hiLiteColor := u.uiTheme.GetUIColor(UIColorTextForegroundHighlighted)

    var sb strings.Builder
    for i, line := range logLines {
        text := ToColoredText(line, 1.0, textColor, hiLiteColor)
        sb.WriteString(text)
        if i < len(logLines)-1 {
            sb.WriteString("\n")
        }
    }
    textView := u.openTextModal(sb.String())
    textView.ScrollToEnd()
    textView.SetTitle("Log Messages")
}

func (u *UI) ShowActorOverlay() {
    listOfEnemies := u.game.GetVisibleActors()

    if len(listOfEnemies) == 0 {
        u.Print(foundation.Msg("No actors in sight"))
        return
    }
    u.mapOverlay.ClearAll()

    for _, enemy := range listOfEnemies {
        u.mapOverlay.TryAddOverlay(enemy.Position(), enemy.Name(), u.GetMapWindowGridSize(), u.game.IsSomethingInterestingAtLoc)
    }
}
