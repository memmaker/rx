package foundation

import (
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"strings"
)

// Actions that the User Interface can trigger on the game
type GameForUI interface {
	// init

	UIReady()

	/* Direct Player Control */

	// ManualMovePlayer Single Step in any Direction
	ManualMovePlayer(direction geometry.CompassDirection)
	// RunPlayer Start or continue running in a direction
	RunPlayer(direction geometry.CompassDirection, isStarting bool) bool

	// Do stuff

	PickupItem()
	EquipToggle(item ItemForUI)
	DropItem(item ItemForUI)
	PlayerApplyItem(item ItemForUI)
	Wait()

	PlayerRangedAttack()
	PlayerQuickRangedAttack()

	ReloadWeapon()
	SwitchWeapons()
	CycleTargetMode()
	PlayerApplySkill()

	PlayerInteractWithMap() // up/down stairs..
	PlayerInteractInDirection(direction geometry.CompassDirection)

	OpenTacticsMenu()

	// State Queries
	GetPlayerPosition() geometry.Point
	GetCharacterSheet() []string

	GetBodyPartsAndHitChances(targeted ActorForUI) []fxtools.Tuple[string, int]
	GetRangedChanceToHitForUI(target ActorForUI) int

	GetHudStats() map[HudValue]int
	GetHudFlags() map[ActorFlag]int
	GetMapInfo(pos geometry.Point) HiLiteString

	GetInventory() []ItemForUI

	GetVisibleEnemies() []ActorForUI
	GetVisibleItems() []ItemForUI
	GetLog() []HiLiteString

	IsSomethingInterestingAtLoc(position geometry.Point) bool
	IsSomethingBlockingTargetingAtLoc(point geometry.Point) bool

	// Inventory Management
	OpenInventory()
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
	IsLit(pos geometry.Point) bool
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
	GetRandomEnemyName() string
	GetItemInMainHand() (ItemForUI, bool)
	OpenContextMenuFor(pos geometry.Point)
}

type PlayerMoveMode int

const (
	PlayerMoveModeManual PlayerMoveMode = iota
	PlayerMoveModeRun
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

	// Notification of state changes
	UpdateStats()
	UpdateInventory()
	UpdateLogWindow()
	UpdateVisibleEnemies()

	// Targeting
	SelectTarget(onSelected func(targetPos geometry.Point, hitZone int))
	SelectDirection(onSelected func(direction geometry.CompassDirection))
	SelectBodyPart(onSelected func(victim ActorForUI, hitZone int))

	// Menus / Modals / Windows
	OpenInventoryForManagement(stack []ItemForUI)
	OpenInventoryForSelection(stack []ItemForUI, prompt string, onSelected func(item ItemForUI))
	OpenTextWindow(description []string)
	ShowTextFileFullscreen(filename string, onClose func())
	OpenMenu(actions []MenuItem)
	OpenKeypad(correctSequence []rune, onCompletion func(success bool))
	OpenVendorMenu(itemsForSale []fxtools.Tuple[ItemForUI, int], buyItem func(ui ItemForUI, price int))
	ShowGameOver(score ScoreInfo, highScores []ScoreInfo)
	ShowContainer(name string, containedItems []ItemForUI, transfer func(ui ItemForUI))

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
	GetAnimDamage(actorPos geometry.Point, damage int, done func()) Animation
	GetAnimMove(actor ActorForUI, old geometry.Point, new geometry.Point) Animation
	GetAnimQuickMove(actor ActorForUI, path []geometry.Point) Animation
	GetAnimAttack(attacker, defender ActorForUI) Animation
	// GetAnimProjectile won't draw a rune for the projectile if the icon's rune is negative
	GetAnimProjectile(icon rune, colorName string, origin geometry.Point, dest geometry.Point, done func()) (Animation, int)
	GetAnimProjectileWithTrail(leadIcon rune, colorNames []string, path []geometry.Point, done func()) (Animation, int)
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
	SetConversationState(text string, options []MenuItem, conversationPartnerName string, isTerminal bool)
	CloseConversation()
	StartHackingGame(identifier uint64, difficulty Difficulty, previousGuesses []string, onCompletion func(previousGuesses []string, success InteractionResult))
	StartLockpickGame(difficulty Difficulty, getLockpickCount func() int, removeLockpick func(), onCompletion func(result InteractionResult))
	SetColors(palette textiles.ColorPalette, colors map[ItemCategory]color.RGBA)
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

type CheckResult struct {
	Success bool
	Crit    bool
}

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
