package foundation

import (
	"RogueUI/special"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"strings"
)

// Actions that the User Interface can trigger on the game
type GameForUI interface {
	// init
	UIRunning()
	UIReady()

	/* Direct Player Control */

	// ManualMovePlayer Single Step in any Direction
	ManualMovePlayer(direction geometry.CompassDirection)
	// RunPlayer Start or continue running in a direction
	RunPlayer(direction geometry.CompassDirection, isStarting bool) bool
	RunPlayerPath()
	// Do stuff

	PlayerPickupItem()
	EquipToggle(item ItemForUI)
	DropItemFromInventory(item ItemForUI)
	PlayerApplyItem(item ItemForUI)
	Wait()

	PlayerRangedAttack()
	PlayerQuickRangedAttack()

	PlayerReloadWeapon()
	SwitchWeapons()
	CycleTargetMode()
	PlayerApplySkill()

	CheckTransition() // up/down stairs..
	PlayerInteractInDirection(direction geometry.CompassDirection)
	PlayerInteractAtPosition(pos geometry.Point)

	OpenContextMenuFor(pos geometry.Point) bool
	OpenTacticsMenu()
	OpenJournal()
	OpenRestMenu()
	ShowDateTime()

	// State Queries
	GetPlayerName() string
	GetPlayerCharSheet() *special.CharSheet
	GetPlayerPosition() geometry.Point
	GetCharacterSheet() string
	IsPlayerOverEncumbered() bool

	GetBodyPartsAndHitChances(targeted ActorForUI) []fxtools.Tuple3[special.BodyPart, bool, int]
	GetRangedChanceToHitForUI(target ActorForUI) int

	GetHudStats() map[HudValue]int
	GetHudFlags() map[special.ActorFlag]int
	GetMapInfo(pos geometry.Point) HiLiteString
	LightAt(p geometry.Point) fxtools.HDRColor

	GetInventoryForUI() []ItemForUI

	GetVisibleActors() []ActorForUI
	GetVisibleItems() []ItemForUI
	GetLog() []HiLiteString

	GetItemInMainHand() (ItemForUI, bool)
	GetMapDisplayName() string

	IsSomethingInterestingAtLoc(position geometry.Point) bool
	IsSomethingBlockingTargetingAtLoc(point geometry.Point) bool

	// Inventory Management
	OpenInventory()
	OpenAmmoInventory()
	OpenRepairMenu()

	ChooseItemForDrop()
	ChooseItemForThrow()
	ChooseItemForEat()
	ChooseItemForApply()

	ChooseWeaponForWield()
	ChooseArmorForWear()

	ChooseArmorToTakeOff()

	IsEquipped(item ItemForUI) bool

	// Game State
	Reset()

	// Map Drawing
	IsExplored(loc geometry.Point) bool
	IsVisibleToPlayer(loc geometry.Point) bool
	IsInteractionAt(position geometry.Point) bool

	TopEntityAt(loc geometry.Point) EntityType

	MapAt(loc geometry.Point) textiles.TextIcon
	ItemAt(loc geometry.Point) ItemForUI
	ObjectAt(loc geometry.Point) ObjectForUI
	ActorAt(loc geometry.Point) ActorForUI
	DownedActorAt(loc geometry.Point) ActorForUI

	// Level up choices

	// Wizard
	OpenWizardMenu()
	WizardAdvanceTime()
}

type PlayerMoveMode int

const (
	PlayerMoveModeManual PlayerMoveMode = iota
	PlayerMoveModeRun
	PlayerMoveModePath
)

type MoveInfo struct {
	Direction geometry.CompassDirection
	OldPos    geometry.Point
	NewPos    geometry.Point
	Mode      PlayerMoveMode
}

// Actions that the game can trigger on the User Interface
type GameUI interface {
	// init
	SetGame(game GameForUI)
	StartGameLoop()
	InitDungeonUI(palette textiles.ColorPalette, inventoryColors map[ItemCategory]color.RGBA)

	// Basics / Debug
	AskForString(prompt string, prefill string, result func(entered string))
	GetKeybindingsAsString(command string) string

	// Notification of state changes
	UpdateStats()
	UpdateInventory()
	UpdateLogWindow()
	UpdateVisibleActors()

	// Targeting
	SelectTarget(onSelected func(targetPos geometry.Point))
	SelectDirection(onSelected func(direction geometry.CompassDirection))
	SelectBodyPart(previousAim special.BodyPart, onSelected func(victim ActorForUI, hitZone special.BodyPart))

	// Menus / Modals / Windows
	OpenInventoryForManagement(stack []ItemForUI)
	OpenInventoryForSelection(stack []ItemForUI, prompt string, onSelected func(item ItemForUI))
	OpenTextWindow(description string)
	ShowTextFileFullscreen(filename string, onClose func())
	OpenMenu(actions []MenuItem)
	OpenKeypad(correctSequence []rune, onCompletion func(success bool))
	OpenVendorMenu(itemsForSale []fxtools.Tuple[ItemForUI, int], buyItem func(ui ItemForUI, price int))
	ShowGameOver(score ScoreInfo, highScores []ScoreInfo)
	ShowTakeOnlyContainer(name string, containedItems []ItemForUI, transfer func(ui ItemForUI))
	ShowGiveAndTakeContainer(leftName string, leftItems []ItemForUI, rightName string, rightItems []ItemForUI, transferToLeft func(itemTaken ItemForUI, amount int), transferToRight func(itemTaken ItemForUI, amount int))
	OpenAimedShotPicker(actorAt ActorForUI, previousAim special.BodyPart, onSelected func(victim ActorForUI, hitZone special.BodyPart))

	// Auto Move Callback
	AfterPlayerMoved(moveInfo MoveInfo)

	// Animations

	// AddAnimations takes a list of list of animations.
	// Each list contains animations that should be played in parallel.
	// The lists are played in order.
	AddAnimations(animations []Animation)
	AnimatePending() (cancelled bool)
	SkipAnimations()
	GetAnimThrow(item ItemForUI, origin geometry.Point, target geometry.Point) (Animation, int)
	GetAnimDamage(spreadBlood func(mapPos geometry.Point), actorPos geometry.Point, damage int, done func()) Animation
	GetAnimMove(actor ActorForUI, old geometry.Point, new geometry.Point) Animation
	GetAnimQuickMove(actor ActorForUI, path []geometry.Point) Animation
	GetAnimAttack(attacker, defender ActorForUI) Animation
	GetAnimMuzzleFlash(position geometry.Point, flashColor fxtools.HDRColor, radius int, bulletCount int, done func()) Animation

	// GetAnimProjectile won't draw a rune for the projectile if the icon's rune is negative
	GetAnimProjectile(icon rune, colorName string, origin geometry.Point, dest geometry.Point, done func()) (Animation, int)
	GetAnimProjectileWithTrail(leadIcon rune, colorNames []string, path []geometry.Point, done func()) (Animation, int)
	GetAnimProjectileWithLight(leadIcon rune, lightColorName string, pathOfFlight []geometry.Point, done func()) (Animation, int)
	GetAnimTiles(positions []geometry.Point, frames []textiles.TextIcon, done func()) Animation
	GetAnimTeleport(actor ActorForUI, origin geometry.Point, targetPos geometry.Point, appearOnMap func()) (vanishAnim, appearAnim Animation)
	GetAnimRadialReveal(position geometry.Point, dijkstra map[geometry.Point]int, done func()) Animation
	GetAnimRadialAlert(position geometry.Point, dijkstra map[geometry.Point]int, done func()) Animation
	GetAnimUncloakAtPosition(actor ActorForUI, position geometry.Point) (Animation, int)
	GetAnimExplosion(points []geometry.Point, done func()) Animation
	GetAnimEnchantArmor(actor ActorForUI, position geometry.Point, done func()) Animation
	GetAnimEnchantWeapon(actor ActorForUI, position geometry.Point, done func()) Animation
	GetAnimVorpalizeWeapon(origin geometry.Point, done func()) []Animation
	GetAnimConfuse(position geometry.Point, done func()) Animation
	GetAnimBreath(flight []geometry.Point, done func()) Animation
	GetAnimBackgroundColor(position geometry.Point, colorName string, frameCount int, done func()) Animation
	GetAnimAppearance(actor ActorForUI, position geometry.Point, done func()) Animation
	GetAnimWakeUp(position geometry.Point, done func()) Animation
	GetAnimEvade(defender ActorForUI, done func()) Animation

	PlayMusic(fileName string)
	PlayCue(cue string)
	SetConversationState(text string, options []MenuItem, conversationPartner ChatterSource, isTerminal bool)
	CloseConversation()
	StartHackingGame(identifier uint64, difficulty Difficulty, previousGuesses []string, onCompletion func(previousGuesses []string, success InteractionResult))
	StartLockpickGame(difficulty Difficulty, getLockpickCount func() int, removeLockpick func(), onCompletion func(result InteractionResult))
	SetColors(palette textiles.ColorPalette, colors map[ItemCategory]color.RGBA)
	TryAddChatter(victim ChatterSource, text string) bool
	FadeToBlack()
	FadeFromBlack()
	AskForConfirmation(title string, message string, onConfirm func(didConfirm bool))
}

type Animation interface {
	IsDone() bool
	SetFollowUp([]Animation)
	RequestMapUpdateOnFinish()
	SetAudioCue(cueName string)
}

type MenuItem struct {
	Name       string
	Action     func()
	CloseMenus bool
}

type UIStat struct {
	DisplayName          string
	CurrentValue         int
	MaxValue             int
	MaxLenOfValueDisplay int
}

type ScoreInfo struct {
	PlayerName         string
	DescriptiveMessage string
	Escaped            bool
	Gold               int
}

type EntityType int

const (
	EntityTypeWorldTile EntityType = iota
	EntityTypeActor
	EntityTypeDownedActor
	EntityTypeItem
	EntityTypeObject
	EntityTypeOther
)

type InteractionResult uint8

func (r InteractionResult) String() string {
	switch r {
	case Success:
		return "Success"
	case Failure:
		return "Failure"
	case Cancel:
		return "Cancel"
	}
	return "Unknown"
}

const (
	Success InteractionResult = iota
	Failure
	Cancel
)

type Difficulty uint8

func (d Difficulty) String() string {
	switch d {
	case VeryEasy:
		return "quite simple"
	case Easy:
		return "simplistic"
	case Medium:
		return "average"
	case Hard:
		return "challenging"
	case VeryHard:
		return "very complex"
	}
	return "Unknown"
}

func (d Difficulty) GetRollModifier() int {
	switch d {
	case VeryEasy:
		return 10
	case Easy:
		return 5
	case Medium:
		return 0
	case Hard:
		return -20
	case VeryHard:
		return -40
	}
	return 0
}

func (d Difficulty) GetStrength() int {
	switch d {
	case VeryEasy:
		return 20
	case Easy:
		return 40
	case Medium:
		return 60
	case Hard:
		return 80
	case VeryHard:
		return 100
	}
	return 3
}

const (
	VeryEasy Difficulty = iota
	Easy
	Medium
	Hard
	VeryHard
)

func DifficultyFromString(difficulty string) Difficulty {
	difficulty = strings.ToLower(difficulty)
	switch difficulty {
	case "veryeasy":
		return VeryEasy
	case "easy":
		return Easy
	case "medium":
		return Medium
	case "hard":
		return Hard
	case "veryhard":
		return VeryHard
	}
	panic("Unknown difficulty: " + difficulty)
	return Medium
}

type ChatterSource interface {
	Name() string
	Position() geometry.Point
	Icon() textiles.TextIcon
}
type AudioCuePlayer interface {
	PlayCue(cueName string)
}
