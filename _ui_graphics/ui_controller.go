package ui_graphics

import (
    "contractor/foundation"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/memmaker/go/geometry"
    "golang.org/x/image/font"
    "golang.org/x/image/font/opentype"
    "image/color"
    "io"
    "math"
    "slices"
    "strings"
)

type TilemapAnimInfo struct {
    FrameCount     int
    DelayInSeconds float64
}

type UIWidget interface {
    InputReceiver
    Draw()
    ShouldClose() bool
    Layout()
}

type UIRenderer interface {
    DrawStringOnGrid(gridX int, gridY int, text string, color color.Color)
    DrawCharOnGrid(gridX int, gridY int, char rune, textColor color.Color)

    DrawOnScreen(screenX int, screenY int, icon int32)
    DrawCharOnScreen(screenX int, screenY int, char rune, scale float64, textColor color.Color)
    DrawOnScreenWithScale(screenX int, screenY int, icon int32, scale float64)
    DrawOnMap(highlight geometry.Point, icon int32, color color.Color)
    DrawOnGrid(gridX int, gridY int, icon int32)
    DrawOnFontGrid(gridX int, gridY int, icon int32)
    GetFontScale() float64
    GetFontGridSize() geometry.Point
    DrawTTFOnScreen(screenX float64, screenY float64, line string, color color.Color)
    MeasureString(line string) (float64, float64)
    DrawDefaultBorder(topLeftScreen geometry.Point, size geometry.Point)
    DrawColoredBorder(topLeftScreen geometry.Point, size geometry.Point, fillColor, borderColor color.Color)
    DrawColoredRect(topLeftScreen geometry.Point, size geometry.Point, fillColor color.Color)
    DrawTTFOnScreenWithColorCodes(screenX float64, screenY float64, textWithColorCodes []ColoredTextPart)
    DrawImageOnScreen(screenX int, screenY int, size geometry.Point, image *ebiten.Image)

    ParseColorCodedText(string) []ColoredTextPart
}

func NewUIController(tileRenderer *TileRenderer, context foundation.GameForUI, config *foundation.Configuration) *UIController {
    u := &UIController{
        tileRenderer:   tileRenderer,
        screenSize:     config.WindowSize,
        configuration:  config,
        gameState:      context,
        deviceDPIScale: ebiten.Monitor().DeviceScaleFactor(),
    }
    u.setTileScale(4)
    return u
}

type UIConfig struct {
    DrawTilesBehindActors  bool
    DrawTilesBehindItems   bool
    DrawTilesBehindObjects bool
    ActorBobbingEnabled    bool
    ItemNamesInShop        ItemNamingScheme
    KeepPlayerCentered     bool
}

type UIController struct {
    modal        UIWidget
    overlays     []UIWidget
    tileRenderer *TileRenderer
    mapRenderer  *MapRenderer

    gameState                   foundation.GameForUI
    isMouseCapturedByUI         bool
    shiftIsBeingPressedThisTick bool

    currentTooltip           Tooltip
    ticksUntilTooltipAppears uint64
    screenSize               geometry.Point
    uiBarWidth               int
    elevatedOverlay          UIWidget

    statsLabel    *BasicLabel
    levelUpButton *ButtonOverlay
    inventory     *IconOverlay
    skills        *IconOverlay
    timer         *IconOverlay
    miniMap       *MinimapOverlay
    message       PrintMessage

    configuration *foundation.Configuration

    setToExactFit    bool
    isFullScreen     bool
    logHook          func(string)
    worldTicks       uint64
    shouldQuitGame   bool
    mousePosInPixels geometry.Point

    targeter                   *Targeter
    tilemapAnimations          map[int32]TilemapAnimInfo
    deviceDPIScale             float64
    tileScale                  float64
    globalFlashColorTicksLeft  int
    globalFlashColorTicksStart int
    globalScaleColor           color.Color
    disableFogOfWar            bool
    mousePosOnMap              geometry.Point
    msg                        func(message string) // this might be a smell already? Should the UI really generate messages to the player?
    globalFadeColorTicksLeft   int
    globalFadeColorTicksStart  int
    fadeFromBlack              bool
    keepBlack                  bool
}

func (u *UIController) FadeToBlack() {
    u.globalFadeColorTicksLeft = int(SecondsToTicks(2))
    u.globalFadeColorTicksStart = u.globalFadeColorTicksLeft
    u.fadeFromBlack = false
}

func (u *UIController) FadeFromBlack() {
    u.globalFadeColorTicksLeft = int(SecondsToTicks(2))
    u.globalFadeColorTicksStart = u.globalFadeColorTicksLeft
    u.fadeFromBlack = true
}

func (u *UIController) IsPlayerForcingAggression() bool {
    return u.shiftIsBeingPressedThisTick
}

func (u *UIController) GetVisibleMap() geometry.Rect {
    return u.mapRenderer.GetVisibleMap()
}

func (u *UIController) ShowLogMessage(message string) {
    if message == "" {
        return
    }
    rawText := RemoveColorCodes(message)
    u.message.rawText = rawText
    u.message.textForRenderer = u.tileRenderer.ParseColorCodedText(message)
    ticks := SecondsToTicks(3)
    u.message.ticksLeft = ticks
    u.message.ticksStart = ticks
}

func GetAutoFitRectWithExtraSpace[T fmt.Stringer](text []T, screenSize geometry.Point, lineDist int, padding geometry.Point, measureString func(line string) (float64, float64)) (geometry.Point, PositionInfo) {
    posInfo := SizeNeededForText(measureString, text, lineDist)
    posInfo.SizeInPixels.X += padding.X
    posInfo.SizeInPixels.Y += padding.Y

    startX := (screenSize.X - posInfo.SizeInPixels.X) / 2
    //endX := startX + width

    startY := (screenSize.Y - posInfo.SizeInPixels.Y) / 2
    //endY := startY + height

    topLeft := geometry.Point{X: startX, Y: startY}
    //bottomRight := geometry.Point{X: endX, Y: endY}
    return topLeft, posInfo
}

func (u *UIController) GetXPosAndHeightForIconText(text []string) (int, int, int) {
    //TODO
    return 0, 0, 0
}

func (u *UIController) ChangeMapZoom(sign int) {

    change := 0.25 * float64(sign)
    minScale := 0.25
    maxScale := 16.0

    currentScale := u.tileScale
    if sign > 0 && u.setToExactFit {
        // round to next 0.25
        currentScale = math.Floor(currentScale*4) / 4
        u.setToExactFit = false
    }
    targetScale := Clamp(minScale, maxScale, currentScale+change)

    mapSize := u.gameState.GetMapSize()
    tileSize := u.configuration.TileSize()

    targetMapSize := geometry.Point{
        X: int(float64(mapSize.X) * float64(tileSize.X) * targetScale),
        Y: int(float64(mapSize.Y) * float64(tileSize.Y) * targetScale),
    }

    windowSize := u.GetDeviceIndependentScreenSize()
    if u.HasOverlays() {
        windowSize = windowSize.Sub(geometry.Point{X: u.uiBarWidth, Y: 0})
    }

    centerMapPos := u.mapRenderer.GetVisibleMap().Center()

    tooSmall := targetMapSize.X < windowSize.X && targetMapSize.Y < windowSize.Y
    if tooSmall {
        u.disableScrolling()
        return
    }

    u.setTileScale(targetScale)
    u.mapRenderer.CenterOn(centerMapPos)
}

func (u *UIController) disableScrolling() {
    u.setToExactFit = true
    state := u.gameState
    mapSize := state.GetMapSize()
    tileSize := u.configuration.TileSize()

    windowSize := u.GetDeviceIndependentScreenSize()
    if u.HasOverlays() {
        windowSize = windowSize.Sub(geometry.Point{X: u.uiBarWidth, Y: 0})
    }
    // calculate exact scale to fit
    mapSizeW := float64(mapSize.X) * float64(tileSize.X)
    mapSizeH := float64(mapSize.Y) * float64(tileSize.Y)
    targetScaleH := float64(windowSize.X) / mapSizeW
    targetScaleV := float64(windowSize.Y) / mapSizeH
    targetScale := min(targetScaleH, targetScaleV)
    u.mapRenderer.ResetScrolling()

    u.setTileScale(targetScale)
}

func (u *UIController) GetTopModal() UIWidget {
    return u.modal
}

func (u *UIController) IsUIModalVisible() bool {
    return u.modal != nil
}
func (u *UIController) PushWidget(widget UIWidget) {
    u.modal = widget
}

func (u *UIController) PopWidget() {
    u.modal = nil
}

func (u *UIController) OnCommand(command CommandType) bool {
    if u.IsUIModalVisible() && u.GetTopModal().OnCommand(command) {
        return true
    }
    for _, overlay := range u.overlays {
        if handled := overlay.OnCommand(command); handled {
            return true
        }
    }
    return false
}

func (u *UIController) OnMouseClicked(button ebiten.MouseButton, screenX int, screenY int) bool {
    u.isMouseCapturedByUI = false

    if u.IsUIModalVisible() && u.GetTopModal().OnMouseClicked(button, screenX, screenY) {
        return true
    }

    for _, overlay := range u.overlays {
        if handled := overlay.OnMouseClicked(button, screenX, screenY); handled {
            u.isMouseCapturedByUI = true
            return true
        }
    }

    return false
}

func (u *UIController) OnMouseMoved(screenX int, screenY int) (bool, Tooltip) {
    u.isMouseCapturedByUI = false

    for _, overlay := range u.overlays {
        if handled, toolTip := overlay.OnMouseMoved(screenX, screenY); handled {
            u.isMouseCapturedByUI = true
            return handled, toolTip
        }
    }

    if !u.IsUIModalVisible() {
        return false, EmptyTooltip{}
    }

    u.isMouseCapturedByUI = true

    return u.GetTopModal().OnMouseMoved(screenX, screenY)
}

func (u *UIController) IsMouseCapturedByUI() bool {
    return u.isMouseCapturedByUI
}

func (u *UIController) OnMouseWheel(x int, y int, dy float64) bool {
    if !u.IsUIModalVisible() {
        return false
    }
    return u.GetTopModal().OnMouseWheel(x, y, dy)
}

func (u *UIController) drawUI() {

    defer u.drawTooltip()
    for _, overlay := range u.overlays {
        if overlay == u.elevatedOverlay {
            continue
        }
        overlay.Draw()
    }
    if !u.IsUIModalVisible() && u.elevatedOverlay == nil {
        return
    }
    renderer := u.GetRenderer()
    renderer.DrawColoredRect(geometry.Point{X: 0, Y: 0}, u.screenSize, color.RGBA{A: 140})

    if u.elevatedOverlay != nil {
        u.elevatedOverlay.Draw()
    }

    if !u.IsUIModalVisible() {
        return
    }

    if u.GetTopModal().ShouldClose() {
        u.PopWidget()
    } else {
        u.GetTopModal().Draw()
    }
}

func (u *UIController) drawTooltip() {
    if u.currentTooltip != nil && u.ticksUntilTooltipAppears == 0 {
        u.currentTooltip.Draw()
    }
}

func (u *UIController) OpenMenu(menuItems []MenuItem) {
    menu := NewBasicMenu(u)
    menu.SetItems(menuItems)
    menu.OnMouseMoved(
        u.mousePosInPixels.X,
        u.mousePosInPixels.Y,
    )
    u.PushWidget(menu)
}
func (u *UIController) GetRenderer() *TileRenderer {
    return u.tileRenderer
}

func (u *UIController) OpenTextWindow(buffer []LabelText) {
    if len(buffer) == 0 {
        return
    }
    textWindow := NewCenteredTextLabel(u)
    textWindow.SetCanBeClosed(true)
    textWindow.SetText(buffer)
    textWindow.SetBorderPadding(geometry.Point{X: 15, Y: 15})
    u.PushWidget(textWindow)
}
func (u *UIController) OpenInventoryOverlay(pos, tileSize geometry.Point, dataSource IconDataSource) *IconOverlay {
    inventory := NewIconOverlay(u, pos, tileSize, dataSource)
    inventory.SetShowEmptySlots(true)
    inventory.SetShowCounter(true)
    u.PushOverlay(inventory)
    return inventory
}

func (u *UIController) OpenInventoryModal(pos, tileSize geometry.Point, dataSource IconDataSource) *IconOverlay {
    inventory := NewIconOverlay(u, pos, tileSize, dataSource)
    inventory.SetShowEmptySlots(true)
    inventory.SetShowCounter(true)
    inventory.SetCanBeClosed(true)
    u.PushWidget(inventory)
    return inventory
}

func (u *UIController) OpenInventoryListModal(tileSize geometry.Point, dataSource ListDataSource) *IconList {
    inventory := NewIconList(u, tileSize, dataSource)
    u.PushWidget(inventory)
    return inventory
}

func (u *UIController) OpenSkillsOverlay(pos, tileSize geometry.Point, dataSource IconDataSource) *IconOverlay {
    skills := NewIconOverlay(u, pos, tileSize, dataSource)
    skills.SetShowEmptySlots(false)
    u.PushOverlay(skills)
    return skills
}
func (u *UIController) OpenTimerOverlay(pos, tileSize geometry.Point, dataSource IconDataSource) *IconOverlay {
    timer := NewIconOverlay(u, pos, tileSize, dataSource)
    timer.SetShowEmptySlots(false)
    timer.SetShowCounter(true)
    u.PushOverlay(timer)
    return timer
}

func (u *UIController) AddButton(offsetFromTopRightCorner, tileSize geometry.Point, icon int32, onClick func()) *ButtonOverlay {
    button := NewIconButton(u, offsetFromTopRightCorner, tileSize, icon)
    button.SetAction(onClick)
    u.PushOverlay(button)
    return button
}

func (u *UIController) SetUIBarWidth(uiBarWidth int) *ColoredRect {
    u.uiBarWidth = uiBarWidth
    background := NewColoredRect(u, uiBarWidth)
    background.SetFillColor(color.Black)
    u.PrependOverlay(background)
    return background
}

func (u *UIController) AddMinimap(offsetFromTopRightCorner geometry.Point, maxHeight, availableWidth int, miniMapImage *ebiten.Image) *MinimapOverlay {
    minimap := NewMiniMap(u, offsetFromTopRightCorner, maxHeight, availableWidth, miniMapImage)
    u.InsertOverlayAtIndex(1, minimap)
    return minimap
}

func (u *UIController) AddLabel(pos geometry.Point) *BasicLabel {
    label := NewTextLabel(u, pos)
    u.PushOverlay(label)
    return label
}

func (u *UIController) PushOverlay(widget UIWidget) {
    u.overlays = append(u.overlays, widget)
}

func (u *UIController) PrependOverlay(widget UIWidget) {
    u.overlays = append([]UIWidget{widget}, u.overlays...)
}

func (u *UIController) InsertOverlayAtIndex(index int, widget UIWidget) {
    u.overlays = append(u.overlays[:index], append([]UIWidget{widget}, u.overlays[index:]...)...)
}

func (u *UIController) ShowTooltipWithDelay(tip Tooltip) {
    u.handleTooltipWithDelay(tip, 0.5)
}

func (u *UIController) handleTooltipWithDelay(tooltip Tooltip, delay float64) {
    if tooltip.IsNull() {
        u.currentTooltip = nil
    } else {
        u.currentTooltip = tooltip
        u.ticksUntilTooltipAppears = SecondsToTicks(delay)
    }
}

func (u *UIController) HideTooltip() {
    u.currentTooltip = nil
}

func (u *UIController) PopAll() {
    u.modal = nil
}

func (u *UIController) SetTileMapAnimations(mapping map[int32]TilemapAnimInfo) {
    u.tilemapAnimations = mapping
}

func (u *UIController) GetScreenSize() geometry.Point {
    return u.screenSize
}

func (u *UIController) OnScreenSizeChanged() {
    u.deviceDPIScale = ebiten.Monitor().DeviceScaleFactor()
    newScreenSize := u.GetDeviceIndependentScreenSize()

    for _, overlay := range u.overlays {
        if overlay != nil {
            overlay.Layout()
        }
    }
    screenSizeForMap := newScreenSize
    if u.HasOverlays() {
        screenSizeForMap = newScreenSize.Sub(geometry.Point{X: u.uiBarWidth, Y: 0})
    }
    u.tileRenderer.SetDeviceScale(u.deviceDPIScale)
    u.mapRenderer.OnScreenSizeChanged(screenSizeForMap)
    u.miniMap.Layout()
    if u.setToExactFit {
        u.disableScrolling()
    }
}

func (u *UIController) ClearAllOverlays() {
    u.overlays = []UIWidget{}
}

func (u *UIController) HasOverlays() bool {
    return len(u.overlays) > 0
}

func (u *UIController) GetUIBarWidth() int {
    return u.uiBarWidth
}

func (u *UIController) OnKeyPressed(key ebiten.Key, shiftPressed bool) {
    if !u.IsUIModalVisible() {
        return
    }
    if keyListener, isKeylistener := u.GetTopModal().(Keylistener); isKeylistener {
        keyListener.OnKeyPressed(key, shiftPressed)
    }
}

func (u *UIController) ElevateOverlay(overlay UIWidget) {
    u.elevatedOverlay = overlay
}

func (u *UIController) StopElevation() {
    u.elevatedOverlay = nil
}

func (u *UIController) ParseColorCodedText(colorCodedText string) []ColoredTextPart {
    return u.tileRenderer.ParseColorCodedText(colorCodedText)
}

/*
	func (u *UIController) OpenStash(playerInv *Inventory, stashInTown *Inventory, drop func(item *Item)) {
		player := u.gameState.GameState().GetPlayer()
		toStashHandler := func(uiItem UIElement) {
			playerEquip := player.GetEquipment()
			itemStack := uiItem.(common.InventoryStack)
			firstItem := itemStack[0]
			if playerEquip.IsEquipped(firstItem) { // unequip first..
				u.gameState.PlayerAction().UnequipItem(firstItem)
			}
			playerInv.Remove(firstItem)
			if u.shiftIsBeingPressedThisTick {
				drop(firstItem)
			} else {
				stashInTown.Add(firstItem)
			}
		}

		dataSource := IconDataSource{
			IsHighlighted: func(element UIElement) bool {
				return false
			},
			GetElements: func() []UIElement {
				return sortStacksForUI(stashInTown.StackedItems())
			},
			OnSelection: func(uiItem UIElement) {
				itemStack := uiItem.(common.InventoryStack)
				firstItem := itemStack[0]
				stashInTown.Remove(firstItem)

				if u.shiftIsBeingPressedThisTick { // drop
					firstItem.DroppedByPlayer()
					u.gameState.PlayerAction().DropItem(firstItem)
				} else { // move
					playerInv.Add(firstItem)
				}
			},
		}

		tileSize := u.worldTexture.GetTileSize()

		screenSize := u.GetDeviceIndependentScreenSize()
		invBounds := u.inventory.GetBounds()
		paddingX := 20
		offsetFromBottomRightCorner := geometry.Point{X: screenSize.X - invBounds.Max.X + invBounds.Size().X + paddingX, Y: screenSize.Y - invBounds.Max.Y} //- invBounds.Size().X - paddingX, Y: invBounds.Min.Y}

		stashInv := u.OpenInventoryModal(offsetFromBottomRightCorner, tileSize, dataSource)
		u.ElevateOverlay(u.inventory)

		oldHandler := u.inventory.GetOnSelectedHandler()
		u.inventory.SetOnSelectedHandler(toStashHandler)

		onClose := func() { // revert inventory..
			u.StopElevation()
			u.inventory.SetOnSelectedHandler(oldHandler)
		}

		stashInv.SetOnCloseHandler(onClose)
	}
*/
func (u *UIController) setupUI() {
    u.overlays = []UIWidget{}
    u.modal = nil

    paddingLeftAndRight := 6
    u.inventory = u.addInventoryOverlay(geometry.Point{X: paddingLeftAndRight, Y: 10})
    inventoryBoxSize := u.inventory.GetBounds().Size()

    u.uiBarWidth = inventoryBoxSize.X + (2 * paddingLeftAndRight)
    u.SetUIBarWidth(u.uiBarWidth)
    xOffset := u.uiBarWidth - paddingLeftAndRight
    u.statsLabel = u.AddLabel(geometry.Point{X: xOffset, Y: paddingLeftAndRight})
    u.statsLabel.SetText(ToLabelText([]string{ // hacky way to measure the right bounding box..
        "Level 1",
        "MoreInfo",
        "Health: 999",
        "Mana: 999",
        "Gold: 999",
    }))
    statsBounds := u.statsLabel.GetBounds()
    levelUpButtonY := statsBounds.Max.Y + paddingLeftAndRight*2

    u.levelUpButton = u.addLevelUpButton(geometry.Point{X: xOffset, Y: levelUpButtonY})
    buttonBounds := u.levelUpButton.GetBounds()

    u.skills = u.addSkillsOverlay(geometry.Point{X: paddingLeftAndRight, Y: 180})
    skillBounds := u.skills.GetBounds()

    topYForMap := buttonBounds.Max.Y + paddingLeftAndRight
    bottomYForMap := skillBounds.Min.Y - paddingLeftAndRight
    maxHeightForMap := bottomYForMap - topYForMap

    u.miniMap = u.addMiniMap(geometry.Point{X: u.uiBarWidth, Y: topYForMap}, maxHeightForMap, paddingLeftAndRight)

    u.timer = u.addTimerOverlay(geometry.Point{X: paddingLeftAndRight, Y: 360})

    u.OnScreenSizeChanged()
}

func (u *UIController) addInventoryOverlay(pixelEdgeFromBottomRight geometry.Point) *IconOverlay {
    state := u.gameState
    isEquipped := func(uiItem UIElement) bool {
        player := state.GetPlayer()
        itemStack := uiItem.(common.InventoryStack)
        firstItem := itemStack[0]
        return player.GetEquipment().IsEquipped(firstItem)
    }

    dataSource := IconDataSource{
        IsHighlighted: isEquipped,
        GetElements: func() []UIElement {
            return sortStacksForUI(state.GetPlayer().GetInventory().StackedItems())
        },
        OnSelection: u.onItemInInventoryClicked,
    }
    tileSize := u.configuration.TileSize()
    return u.OpenInventoryOverlay(pixelEdgeFromBottomRight, tileSize, dataSource)
}

func firstItemFromUIStack(uiItem UIElement) foundation.Item {
    var firstItem foundation.Item

    if item, isItem := uiItem.(foundation.Item); isItem {
        firstItem = item
    } else if textStack, isTextualStack := uiItem.(*common.TextualStack); isTextualStack {
        firstItem = textStack.InventoryStack[0]
    } else if itemStack, isItemStack := uiItem.(common.InventoryStack); isItemStack {
        firstItem = itemStack[0]
    }

    return firstItem
}

func (u *UIController) onItemInInventoryClicked(uiItem UIElement) {
    state := u.gameState.GameState()
    if !state.GetPlayer().IsAlive() {
        return
    }
    act := u.gameState.PlayerAction()
    // shift state
    u.shiftIsBeingPressedThisTick = ebiten.IsKeyPressed(ebiten.KeyShift)
    typedItem := firstItemFromUIStack(uiItem)

    player := state.GetPlayer()

    wasEquipped := player.GetEquipment().IsEquipped(typedItem)
    if u.shiftIsBeingPressedThisTick { // drop
        act.DropItem(typedItem)
        if wasEquipped {
            u.ShowFullStats()
        }
        return
    }

    // equip
    if typedItem.IsEquippable() {
        if wasEquipped {
            act.UnequipItem(typedItem)
        } else {
            act.EquipItem(typedItem)
        }
        u.ShowFullStats()
        return
    }

    // use
    contextActions := u.gameState.PlayerAction().GetContextActionsForItem(typedItem)
    if len(contextActions) == 0 {
        return // maybe open a description? or at least a menu with drop or sth?
    }

    if len(contextActions) == 1 {
        actionFunc := contextActions[0].Action
        actionFunc()
        // hmm, this will trigger even if unsuc
        return
    }
    u.OpenMenu(contextActions)
}
func (u *UIController) ShowFullStats() {
    u.levelUpButton.Hide()
    u.miniMap.Hide()
    label := u.statsLabel

    player := u.gameState.GameState().GetPlayer()
    label.SetText(player.GetFullStats())
}

func (u *UIController) addLevelUpButton(offset geometry.Point) *ButtonOverlay {
    button := u.AddButton(offset, u.worldTexture.GetTileSize(), 141, u.openLevelUpMenu)
    button.SetTooltip("Level up")
    button.Hide()
    return button
}
func (u *UIController) BeginPlayerSellingTo(vendor *common.Actor) {
    player := u.gameState.GameState().GetPlayer()
    playerInventory := player.GetInventory()
    msg := u.msg
    if playerInventory.IsEmpty() {
        msg("You have nothing to sell")
        return
    }
    var sellItems []ShopMenuItem
    textParser := u.tileRenderer.ParseColorCodedText

    for _, itemStack := range playerInventory.StackedItems() {
        typedItem := itemStack[0]
        sellPrice := typedItem.BuyPrice() / 2

        itemName := typedItem.Name()
        if u.configuration.ItemNamesInShop == common.ItemNamingSchemeEnchantments {
            itemName = typedItem.PlusName()
        } else if u.configuration.ItemNamesInShop == common.ItemNamingSchemeBandStyle {
            itemName = typedItem.BandStyleName()
        }
        sellItems = append(sellItems, ShopMenuItem{
            ShortCut: playerInventory.GetLetterForItemStack(itemStack),
            Icon:     typedItem.InventoryIcon(),
            Tooltip:  typedItem.GetTooltipLines(textParser),
            TextColumns: []LabelText{
                {
                    Text:               itemName,
                    TextWithColorCodes: textParser(itemName),
                    UseColorCodes:      true,
                },
                {
                    Text:               fmt.Sprintf("%d", sellPrice),
                    TextWithColorCodes: textParser(fmt.Sprintf("%d", sellPrice)),
                    UseColorCodes:      true,
                },
            },
            Action: func() {
                playerEquip := player.GetEquipment()
                if playerEquip.IsEquipped(typedItem) {
                    u.gameState.PlayerAction().UnequipItem(typedItem)
                }
                playerInventory.Remove(typedItem)
                vendor.GetInventory().Add(typedItem)
                player.GetStats().IncrementBy(stats.Gold, sellPrice)
                //u.sendInventoryToWindowClient()
                msg(fmt.Sprintf("You sold [:darkText]%s[:white] for [:darkText]%d[:white] gold", typedItem.Name(), sellPrice))
                u.BeginPlayerSellingTo(vendor)
            },
        })
    }
    goldAmount := player.GetStats().GetAttribute(stats.Gold)
    menu := NewShopMenu(u, u.worldTexture.GetTileSize())
    u.PushWidget(menu)
    menu.SetHeader("You have:", fmt.Sprintf("%d gold", goldAmount))
    menu.SetItems(sellItems, []Alignment{AlignmentLeft, AlignmentRight})
}
func (u *UIController) openLevelUpMenu() {
    player := u.gameState.GameState().GetPlayer()
    if !player.CanLevelUp() {
        return
    }
    audioPlayer := u.gameState.AudioPlayer
    u.levelUpButton.Hide()
    u.ShowFullStats()
    u.OpenMenu([]MenuItem{
        {
            MainText: "STR",
            Action: func() {
                player.LevelUp(stats.Strength)
                if player.CanLevelUp() {
                    audioPlayer.PlayCue("levelup")
                    u.openLevelUpMenu()
                } else {
                    u.levelUpButton.Hide()
                    u.ShowFullStats()
                }
            },
        },
        {
            MainText: "DEX",
            Action: func() {
                player.LevelUp(stats.Dexterity)
                if player.CanLevelUp() {
                    audioPlayer.PlayCue("levelup")
                    u.openLevelUpMenu()
                } else {
                    u.levelUpButton.Hide()
                    u.ShowFullStats()
                }
            },
        },
        {
            MainText: "INT",
            Action: func() {
                player.LevelUp(stats.Intelligence)
                if player.CanLevelUp() {
                    audioPlayer.PlayCue("levelup")
                    u.openLevelUpMenu()
                } else {
                    u.levelUpButton.Hide()
                    u.ShowFullStats()
                }
            },
        },
    })
}
func (u *UIController) addMiniMap(offset geometry.Point, maxHeight int, horizontalPadding int) *MinimapOverlay {
    miniMap := u.AddMinimap(offset.Sub(geometry.Point{X: horizontalPadding}), maxHeight, offset.X-(horizontalPadding*2), u.mapRenderer.GetMiniMap())
    miniMap.SetOnMapClickedHandler(u.onMapClicked)
    return miniMap
}

func (u *UIController) addSkillsOverlay(pixelEdgeFromBottomRight geometry.Point) *IconOverlay {
    tileSize := u.worldTexture.GetTileSize()
    state := u.gameState.GameState()
    getList := func() []UIElement {
        player := state.GetPlayer()
        items := make([]UIElement, 0)
        for _, item := range u.GetActiveSkillsForUI(player) {
            items = append(items, item)
        }
        return items
    }
    isEquipped := func(item UIElement) bool {
        return false // TODO:
    }

    dataSource := IconDataSource{
        IsHighlighted: isEquipped,
        GetElements:   getList,
        OnSelection: func(element UIElement) {
            switch typedElement := element.(type) {
            case *common.UntargetedAction:
                u.gameState.PlayerAction().ActivateUntargetedAction(typedElement, common.EffectSource(typedElement.GetInternalName()))
            case *common.TargetedAction:
                u.gameState.PlayerAction().BeginActionTargeting(typedElement)
            }
        },
    }

    return u.OpenSkillsOverlay(pixelEdgeFromBottomRight, tileSize, dataSource)
}
func (u *UIController) GetActiveSkillsForUI(player *common.Actor) []UIElement {
    var activeSkills []UIElement
    for _, skill := range player.GetTargetedSkills() {
        activeSkills = append(activeSkills, skill)
    }
    for _, skill := range player.GetUntargetedSkills() {
        activeSkills = append(activeSkills, skill)
    }
    return activeSkills
}

func (u *UIController) addTimerOverlay(pixelEdgeFromBottomRight geometry.Point) *IconOverlay {
    tileSize := u.worldTexture.GetTileSize()

    isEquipped := func(item UIElement) bool {
        return false // TODO:
    }

    dataSource := IconDataSource{
        IsHighlighted: isEquipped,
        GetElements:   u.getCurrentTimedEffects,
        OnSelection: func(element UIElement) {
        },
    }

    return u.OpenTimerOverlay(pixelEdgeFromBottomRight, tileSize, dataSource)
}
func (u *UIController) getCurrentTimedEffects() []UIElement {
    var result []UIElement
    for _, timedBeing := range u.gameState.GameState().GetTimeKeeper().GetBeingsByFilter(func(being common.TimedBeing) bool {
        switch being.(type) {
        case *common.Actor:
            return false
        case *common.UntargetedAction: // this is used to delay activation of spells
            return true
        case *common.TimedModifier: // these are temporary buffs/debuffs to a characters' stats
            return true
        default:
            return false
        }
    }) {
        result = append(result, timedBeing.(UIElement))
    }

    for statusName, turnsLeft := range u.gameState.GameState().GetPlayer().GetStatusEffects().AllEntries() {
        result = append(result, NewStatusEffectWidget(statusName, turnsLeft))
    }
    for _, statusEffect := range u.gameState.GameState().GetPlayer().GetParametrizedStatusEffects() {
        result = append(result, NewStatusEffectWidget(statusEffect.GetName(), statusEffect.TurnsLeft()))
    }
    return result
}
func (u *UIController) SetTTFFont(file io.ReaderAt, size float64) {
    tt, err := opentype.ParseReaderAt(file)
    if err != nil {
        println(err.Error())
        return
    }
    dpi := 72 * u.deviceDPIScale
    fontFace, faceErr := opentype.NewFace(tt, &opentype.FaceOptions{
        Size:    size,
        DPI:     dpi,
        Hinting: font.HintingVertical,
    })
    if faceErr != nil {
        println(faceErr.Error())
    }
    //mplusBigFont = text.FaceWithLineHeight(mplusBigFont, 54) // adjust line height
    u.tileRenderer.SetTTF(fontFace)
}

func (u *UIController) OnMapChanged() {
    gridSize := u.worldTexture.GetTileSize()
    mapWindow := NewMapWindow(
        u.GetDeviceIndependentScreenSize().Sub(geometry.Point{X: u.uiBarWidth, Y: 0}),
        u.gameState.GetMapSize(),
        gridSize,
        u.GetTileScaleFactor,
        u.StaticMapLookup)

    u.mapRenderer = NewMapRenderer(u.tileRenderer, mapWindow)
    u.mapRenderer.SetMiniMapColorHandler(u.getMiniMapColor)

    u.setupUI()

    u.OnPlayerTurnTaken()
}
func (u *UIController) getMiniMapColor(mapPos geometry.Point, tileColor *[4]byte) {
    state := u.gameState.GameState()
    currentMap := state.GetMap()
    if state.GetPlayer().MapPosition() == mapPos {
        *tileColor = [4]byte{100, 255, 100, 255}
        return
    }
    if !currentMap.IsExplored(mapPos) {
        *tileColor = [4]byte{10, 10, 10, 255}
        return
    }
    if currentMap.IsStairsAt(mapPos) {
        *tileColor = [4]byte{235, 52, 229, 255}
        return
    }
    if !state.GetPlayer().CanSee(mapPos) {
        if currentMap.IsSpecialAt(mapPos, gridmap.SpecialTileWater) {
            if !state.GetFlags().IsSet(fmt.Sprintf("all_clear:%s", defs.SpecialLevelPoisonedWaterSupply)) {
                //199, G: 251, B: 161
                *tileColor = [4]byte{199 / 2, 251 / 2, 161 / 2, 255}
                return
            }
            *tileColor = [4]byte{86 / 2, 112 / 2, 194 / 2, 255}
            return
        }
        if currentMap.IsWalkable(mapPos) {
            *tileColor = [4]byte{35, 30, 51, 255}
        } else {
            *tileColor = [4]byte{85, 95, 112, 255}
        }
        return
    }
    if currentMap.IsActorAt(mapPos) {
        actorAt := currentMap.ActorAt(mapPos)
        if actorAt.IsHostileTo(state.GetPlayer()) {
            *tileColor = [4]byte{255, 0, 0, 255}
            return
        }
        *tileColor = [4]byte{220, 120, 75, 255}
        return
    }

    if currentMap.IsSpecialAt(mapPos, gridmap.SpecialTileWater) {
        if !state.GetFlags().IsSet(fmt.Sprintf("all_clear:%s", defs.SpecialLevelPoisonedWaterSupply)) {
            //199, G: 251, B: 161
            *tileColor = [4]byte{199, 251, 161, 255}
            return
        }
        *tileColor = [4]byte{86, 112, 194, 255}
        return
    }

    if currentMap.IsWalkable(mapPos) {
        *tileColor = [4]byte{77, 60, 102, 255}
    } else {
        *tileColor = [4]byte{170, 191, 224, 255}
    }
}

func prepend[T any](slice []T, s T) []T {
    return append([]T{s}, slice...)
}
func (u *UIController) StaticMapLookup(x int, y int, tick uint64) []CellDrawInfo {
    mapPos := geometry.Point{X: x, Y: y}
    state := u.gameState
    mapTile := state(mapPos)

    if currentMap == nil || !currentMap.Contains(mapPos) {
        return []CellDrawInfo{}
    }

    if !currentMap.IsExplored(mapPos) && !u.disableFogOfWar {
        return []CellDrawInfo{}
    }

    var drawColor color.RGBA = color.RGBA{255, 255, 255, 255}

    tileAt := currentMap.GetCell(mapPos).TileType
    worldIcon, worldTileColor := CodePage437IconAndColor(tileAt)
    if tileAt.Special == gridmap.SpecialTileWater {
        worldTileColor = state.GetWaterColor()
    }

    // animate the worldIcon
    if u.renderingMode == GraphicRendering && (state.GetPlayer().CanSee(mapPos) || u.disableFogOfWar) {
        if !u.configuration.DrawTilesBehindActors && currentMap.IsActorAt(mapPos) {
            return []CellDrawInfo{}
        }
        if info, ok := u.tilemapAnimations[worldIcon]; ok {
            // if it is, then calculate the current frame based on the tick
            frameOffset := util.GetLoopingFrameFromTick(tick+1000-uint64(x*17)-uint64(y*19), info.DelayInSeconds, info.FrameCount)
            worldIcon += frameOffset
        }
    }
    // fake lighting
    if !u.disableFogOfWar {
        drawColor = state.ApplyLighting(mapPos, drawColor, tick)
        worldTileColor = state.ApplyLighting(mapPos, worldTileColor, tick)
    }

    var drawInfo []CellDrawInfo

    // 1. items
    if itemAt, isItemAt := currentMap.TryGetItemAt(mapPos); isItemAt && itemAt.IsDrawn() && (!itemAt.IsHidden() || u.gameState.GameState().GetFlags().IsSet(string(common.FlagHasTrueSight))) {
        itemIcon := itemAt.Icon(u.worldTicks)
        drawItem := CellDrawInfo{
            Atlas: u.worldTexture,
            Icon:  itemIcon,
            Color: drawColor,
        }
        drawInfo = prepend(drawInfo, drawItem)
        if !u.configuration.DrawTilesBehindItems {
            return drawInfo
        }
    }

    // 2. objects
    if currentMap.IsObjectAt(mapPos) {
        objectAt := currentMap.ObjectAt(mapPos)
        objIcon := objectAt.Icon(u.worldTicks)
        drawObject := renderer.CellDrawInfo{
            Atlas: u.worldTexture,
            Icon:  objIcon,
            Color: drawColor,
        }
        drawInfo = prepend(drawInfo, drawObject)
        if !u.configuration.DrawTilesBehindObjects {
            return drawInfo
        }
    }

    // 2. decals
    if decalIcon, exists := currentMap.GetDecal(mapPos); exists {
        drawInfo = prepend(drawInfo, CellDrawInfo{
            Atlas: u.worldTexture,
            Icon:  decalIcon,
            Color: color.White,
        })
    }

    // 1. world
    drawInfo = prepend(drawInfo, CellDrawInfo{
        Atlas: u.GetCurrentAtlas(),
        Icon:  worldIcon,
        Color: worldTileColor,
    })

    return drawInfo
}

func (u *UIController) GetDeviceIndependentScreenSize() geometry.Point {
    if u.isFullScreen {
        w, h := ebiten.Monitor().Size()
        return geometry.Point{X: w, Y: h}
    }
    return u.screenSize
}

func (u *UIController) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    panic("should use layoutf")
}

func (u *UIController) LayoutF(outsideWidth, outsideHeight float64) (screenWidth, screenHeight float64) {
    //u.deviceDPIScale = ebiten.DeviceScaleFactor()
    intWidth := int(outsideWidth)
    intHeight := int(outsideHeight)
    if u.screenSize.X != intWidth || u.screenSize.Y != intHeight {
        u.screenSize = geometry.Point{X: intWidth, Y: intHeight}
        u.OnScreenSizeChanged()
    }
    return outsideWidth * u.deviceDPIScale, outsideHeight * u.deviceDPIScale
}

func (u *UIController) GetTileScaleFactor() float64 {
    return u.tileScale
}

func (u *UIController) Cancel() {
    if u.IsUIModalVisible() {
        u.OnCommand(PlayerCommandCancel)
        return
    }
    if u.targeter != nil {
        u.CancelTargeting()
        return
    }
    if u.inventory.IsActive() {
        u.inventory.Deactivate()
        return
    }
}

func (u *UIController) CancelTargeting() {
    u.targeter = nil
}

func (u *UIController) BeginThrowTargeting(targetRules common.TargetRules, onSelected func(attacker *common.Actor, target geometry.Point)) {
    state := u.gameState.GameState()
    if !state.GetPlayer().IsAlive() {
        return
    }
    act := u.gameState.PlayerAction()
    act.CancelAutoMove()
    msg := u.msg
    u.targeter = NewTargeter(u, targetRules, &common.HitEffect{
        Activate: func(attacker *common.Actor, target geometry.Point, hitPositions []geometry.Point) {
            u.CancelTargeting()

            origin := state.GetPlayer().MapPosition()
            if origin == target {
                msg("You can't attack yourself like that")
                return
            }
            u.gameState.Animator().CancelAllAnimations()
            act.CancelAutoMove()
            onSelected(attacker, target)
        },
        Description: []LabelText{
            LabelTextFromString(fmt.Sprintf("Throw")),
        },
    })
}
func (u *UIController) BeginTargeting(targetRules common.TargetRules, onHit *common.HitEffect) {
    u.targeter = NewTargeter(u, targetRules, onHit)
}

func (u *UIController) ShowGameOverMenu() {
    u.OpenMenu([]MenuItem{
        {
            MainText: "Try again",
            Action: func() {
                u.gameState.Reset()
            },
        },
        {
            MainText: "Change class",
            Action: func() {
                u.OpenClassSelectionMenu()
            },
        },
        {
            MainText: "Quit",
            Action: func() {
                u.shouldQuitGame = true
            },
        },
    })
}

/*
func (u *UIController) OpenPlayerStashInTown() {
	state := u.gameState.GameState()
	gui := u
	stashInTown := state.GetPlayerStashInTown()
	playerInv := state.GetPlayer().GetInventory()
	//playerEquip := g.state.GetPlayer().GetEquipment()

	// NEW
	onDrop := func(item foundation.Item) {
		item.DroppedByPlayer()
		state.AddItemToMapWithDistribution(item, state.GetPlayer().MapPosition())
	}
	gui.OpenStash(playerInv, stashInTown, onDrop)
	// on close of the stash modal, restore the onclick handler of the inventory..
}

func (u *UIController) OpenClassSelectionMenu() {
	var items []MenuItem
	for _, c := range u.gameState.Definitions.ClassDefinitions {
		classDef := c
		items = append(items, MenuItem{
			MainText: classDef.Name,
			Action: func() {
				u.gameState.ResetGame(classDef)
			},
		})
	}
	u.OpenMenu(items)
}
*/

func (u *UIController) OnPlayerTurnTaken() {
    u.UpdateMiniMap()
    u.UpdateStatsLabel()
    player := u.gameState.GameState().GetPlayer()

    if (!u.setToExactFit && !u.configuration.KeepPlayerCentered) || u.configuration.KeepPlayerCentered {
        distFromBorder := 4
        playerPos := player.MapPosition()
        visibleMap := u.mapRenderer.GetVisibleMap()
        visibleMap = visibleMap.Shift(distFromBorder, distFromBorder, -distFromBorder, -distFromBorder)
        if !visibleMap.Contains(playerPos) {
            u.mapRenderer.CenterOnFloat(player.DrawPosition())
        }
    }
}

func (u *UIController) UpdateStatsLabel() {

    //statLines := append(mapNameParts, statusParts...)

    //label.SetText(statLines)
    u.miniMap.Show()
}
func (u *UIController) ShowLog() {
    buffer := u.gameState.GetLog()
    if len(buffer) > 20 {
        buffer = buffer[len(buffer)-20:]
    }
    u.OpenTextWindow(toGradientLabelText(u.tileRenderer.ParseColorCodedText, buffer))
    //logWindow.ScrollToBottom()
}

func (u *UIController) OpenListInventory() {
    noFilter := func(element foundation.Item) bool { return true }
    u.OpenListInventoryForSelection(noFilter, func(invItemKey rune) {
        inv := u.gameState.GetInventoryForUI()
        u.onItemInInventoryClicked(inv.GetItemByLetter(invItemKey))
    })
}

func (u *UIController) FlashScreen(color color.Color, seconds float64) {
    u.globalScaleColor = color
    u.globalFlashColorTicksLeft = int(SecondsToTicks(seconds))
    u.globalFlashColorTicksStart = int(SecondsToTicks(seconds))
}

func (u *UIController) OpenListInventoryForSelection(filter func(element foundation.Item) bool, onSelected func(invItemKey rune)) {
    dataSource := ListDataSource{
        GetElements: func() []IconListElement {
            inventory := u.gameState.GetInventoryForUI()
            typed := SortStacksForTextUITyped(inventory.StackedItemsWithFilter(filter), inventory.GetLetterForItemStack, onSelected)
            var result []IconListElement
            for _, item := range typed {
                result = append(result, item)
            }
            return result
        },
    }
    u.PopAll()
    u.OpenInventoryListModal(u.configuration.TileSize(), dataSource)
}

type ColorCodeParser func(draw string) []ColoredTextPart

func toGradientLabelText(textParser ColorCodeParser, buffer []string) []LabelText {
    // we want each item to be less white than the one before
    intensityFromIndex := func(index int) uint8 {
        maxIntensity := uint8(255)
        minIntensity := uint8(100)
        maxIndex := len(buffer) - 1
        percentage := 1.0 - (float64(index) / float64(maxIndex))
        return uint8(float64(maxIntensity) - (float64(maxIntensity)-float64(minIntensity))*percentage)
    }
    var result []LabelText
    for i, line := range buffer {
        fadedBaseColor := color.RGBA{R: intensityFromIndex(i), G: intensityFromIndex(i), B: intensityFromIndex(i), A: 255}
        baseColorCode := RGBAToColorCode(fadedBaseColor)
        result = append(result, LabelText{
            UseColorCodes:      true,
            TextWithColorCodes: textParser(baseColorCode + line),
            Text:               RemoveColorCodes(line),
            TextColor:          fadedBaseColor,
        })
    }
    return result
}

/*
	func sortStacksForUI(stacks []common.InventoryStack) []UIElement {
		items := make([]UIElement, 0)
		common.SortInventory(stacks)
		for _, item := range stacks {
			items = append(items, item)
		}
		return items
	}

	func (u *UIController) OpenShopMenu(vendor *common.Actor) {
		u.openShopMenuWithHandler(vendor.GetInventory(), func(item foundation.Item) {
			u.tryToBuy(item, vendor)
		})
	}

	func (u *UIController) openShopMenuWithHandler(inventory *common.Inventory, handler func(item foundation.Item)) {
		if inventory.IsEmpty() {
			return
		}
		player := u.gameState.GameState().GetPlayer()
		goldAmount := player.GetStats().GetAttribute(stats.Gold)

		var menuItems []ShopMenuItem
		inventoryStacks := inventory.StackedItems()
		common.SortInventory(inventoryStacks)
		textParser := u.ParseColorCodedText

		for itemIndex, i := range inventoryStacks {

			item := i[0]
			cost := item.BuyPrice()
			cannotPayForItem := cost > goldAmount

			colorCodeForPrice := "darkText"
			if cannotPayForItem {
				colorCodeForPrice = "negative"
			}
			itemName := item.Name()
			if u.configuration.ItemNamesInShop == common.ItemNamingSchemeEnchantments {
				itemName = item.PlusName()
			} else if u.configuration.ItemNamesInShop == common.ItemNamingSchemeBandStyle {
				itemName = item.BandStyleName()
			}
			if len(i) > 1 {
				itemName = fmt.Sprintf("%s (x%d)", itemName, len(i))
			}
			nameString := fmt.Sprintf("%s", itemName)
			//nameStringWithColor := fmt.Sprintf("%s", itemName)
			priceString := fmt.Sprintf("%d", cost)
			priceStringWithColor := fmt.Sprintf("[:%s]%d", colorCodeForPrice, cost)
			// shortcuts start with 0 => 'a'
			shortCut := rune('a') + rune(itemIndex)

			menuItems = append(menuItems, ShopMenuItem{
				ShortCut: shortCut,
				Tooltip:  item.GetTooltipLines(textParser),
				Icon:     item.InventoryIcon(),
				TextColumns: []LabelText{
					{
						Text: nameString,
						//TextWithColorCodes: textParser(nameStringWithColor),
						UseColorCodes: false,
						TextColor:     item.TintColor(),
					},
					{
						Text:               priceString,
						TextWithColorCodes: textParser(priceStringWithColor),
						UseColorCodes:      true,
					},
				},
				Action: func() {
					handler(item)
				},
			})
		}

		menu := NewShopMenu(u, u.worldTexture.GetTileSize())
		u.PushWidget(menu)
		menu.SetHeader("You have:", fmt.Sprintf("%d gold", goldAmount))
		menu.SetItems(menuItems, []Alignment{AlignmentLeft, AlignmentRight})
	}

	func (u *UIController) tryToBuy(item foundation.Item, seller *common.Actor) {
		player := u.gameState.GameState().GetPlayer()
		goldAmount := player.GetStats().GetAttribute(stats.Gold)
		msg := u.msg
		cost := item.BuyPrice()
		cannotPayForItem := cost > goldAmount

		if cannotPayForItem {
			msg("You don't have enough gold")
			u.OpenShopMenu(seller)
			return
		}
		player.GetStats().DecrementBy(stats.Gold, cost)
		seller.GetStats().IncrementBy(stats.Gold, cost)

		alwaysItems := u.gameState.Definitions.VendorDefinitions[string(seller.GetVendorType())].Always
		seller.GetInventory().Remove(item)
		player.GetInventory().Add(item)
		itemName := item.GetInternalName()
		if slices.Contains(alwaysItems, itemName) {
			seller.GetInventory().Add(item.Clone())
		}

		msg(fmt.Sprintf("You bought [:darkText]%s[:white] for [:darkText]%d[:white] gold", item.Name(), cost))
		u.OpenShopMenu(seller)
	}

	func (u *UIController) showJournal() {
		state := u.gameState.GameState()
		if !state.IsStoryModeEnabled() {
			return
		}
		entries := u.gameState.GameState().GetQuestLog(u.tileRenderer.ParseColorCodedText)
		u.OpenTextWindow(entries)
	}

	func (u *UIController) showFlags() {
		state := u.gameState.GameState()
		asStrings := state.GetFlags().AsStrings()
		u.OpenTextWindow(ToLabelText(asStrings))
		//logWindow.ScrollToBottom()
	}
*/
func (u *UIController) setTileScale(scale float64) {
    u.tileScale = scale
    u.tileRenderer.SetTileScale(scale)
}

func (u *UIController) UpdateMiniMap() {
    u.mapRenderer.UpdateMiniMap()
}

func (u *UIController) SetPlayerMessage(message func(message string)) {
    u.msg = message
}

func (u *UIController) drawHealthBar(curHP, maxHP int, drawPosition geometry.PointF, isPlayer bool) {
    health := curHP
    maxHealth := maxHP
    healthPercent := float64(health) / float64(maxHealth)

    gridSize := u.configuration.TileSize()
    tileScale := u.GetTileScaleFactor()
    pos := drawPosition
    screenPos := u.mapRenderer.MapFloatToScreen(pos)
    healthBarHeight := 2
    healthBarWidth := float64(gridSize.X) * tileScale

    healthBarX := screenPos.X
    healthBarY := float64(screenPos.Y) + (float64(gridSize.Y) * tileScale) - float64(healthBarHeight)

    healthBarColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}
    backgroundColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

    if !isPlayer {
        backgroundColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}
        healthBarColor = redToGreen(healthPercent)
    }

    topLeft := geometry.Point{X: int(healthBarX), Y: int(healthBarY)}
    size := geometry.Point{X: int(healthBarWidth), Y: healthBarHeight}
    barSize := geometry.Point{X: int(healthBarWidth * healthPercent), Y: healthBarHeight}
    u.tileRenderer.DrawColoredRect(topLeft, size, backgroundColor)
    u.tileRenderer.DrawColoredRect(topLeft, barSize, healthBarColor)
}

func redToGreen(percent float64) color.RGBA {
    // 0 -> red, 0.5 -> yellow, 1 -> green
    red := uint8(255 * (1 - percent))
    green := uint8(255 * percent)
    return color.RGBA{R: red, G: green, B: 0, A: 255}
}
