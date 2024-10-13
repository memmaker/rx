package foundation

import (
	"RogueUI/d100"
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

	SetIronMan()
	IsIronMan() bool

	/* Direct Player Control */

	// ManualMovePlayer Single Step in any Direction
	ManualMovePlayer(direction geometry.CompassDirection)
	// RunPlayer Start or continue running in a direction
	RunPlayer(direction geometry.CompassDirection, isStarting bool) bool
	RunPlayerPath() bool
	// Do stuff
	PlayerPickupItem()
	EquipToggle(item Item)
	DropItemFromInventory(item Item)
	PlayerApplyItem(item Item)
	PlayerExamineItem(item Item)
	PlayerDropItem(item Item)

	PlayerToggleRun()
	PlayerToggleSneak()
	PlayerQuip()
	Wait()

	PlayerRangedAttack()
	PlayerQuickRangedAttack()

	PlayerReloadWeapon()
	CycleTargetMode()
	PlayerApplySkill()

	CheckTransition() // up/down stairs..
	PlayerInteractInDirection(direction geometry.CompassDirection)
	PlayerInteractAtPosition(pos geometry.Point)

	OpenContextMenuFor(pos geometry.Point) bool
	OpenContextMenuForItem(item Item, done func())

	OpenTacticsMenu()
	OpenJournal()
	OpenRestMenu()
	OpenPerkSelection(done func())
	ShowDateTime()

	LoadGame(fromDir string)
	SaveGame(toDir string)

	// State Queries
	IsPlayerAndMapInitialized() bool
	GetPlayerName() string
	GetPlayerCharSheet() *d100.CharSheet
	GetPlayerPosition() geometry.Point
	GetCharacterSheet() string
	IsPlayerOverEncumbered() bool

	CanActorAttackNextTurn(enemy ActorForUI) bool

	GetBodyPartsAndHitChances(targeted ActorForUI) []fxtools.Tuple3[d100.BodyPart, bool, int]

	GetHudStats() map[HudValue]int
	GetHudFlags() map[ActorFlag]int
	GetMapInfo(pos geometry.Point) HiLiteString
	LightAt(p geometry.Point) fxtools.HDRColor

	GetInventoryForUI() []Item

	GetVisibleActors() []ActorForUI
	GetVisibleItems() []Item
	GetLog() []HiLiteString
	GetFlightPath(origin geometry.Point, pos geometry.Point) []geometry.Point

	IsActorHostileTowardsPlayer(enemy ActorForUI) bool
	IsActorAlliedWithPlayer(ally ActorForUI) bool

	GetItemInMainHand() (Item, bool)
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

	IsEquipped(item Item) bool

	// Game State
	Reset()

	// Map Drawing
	IsExplored(loc geometry.Point) bool
	IsVisibleToPlayer(loc geometry.Point) bool
	IsInteractionAt(position geometry.Point) bool

	TopEntityAt(loc geometry.Point) EntityType

	MapAt(loc geometry.Point) textiles.TextIcon
	ItemAt(loc geometry.Point) Item
	ObjectAt(loc geometry.Point) ObjectForUI
	ActorAt(loc geometry.Point) ActorForUI
	DownedActorAt(loc geometry.Point) ActorForUI

	// Level up choices

	// Wizard
	OpenWizardMenu()
	WizardAdvanceTime()

	DebugClickAt(pos geometry.Point)
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
	QuitGame()

	// Notification of state changes
	UpdateStats()
	UpdateInventory()
	UpdateLogWindow()
	UpdateVisibleActors()

	// Targeting
	SelectTarget(getCth func(target ActorForUI) AttackInfo, onSelected func(targetPos geometry.Point))
	SelectDirection(onSelected func(direction geometry.CompassDirection))
	SelectBodyPart(previousAim d100.BodyPart, onSelected func(victim ActorForUI, hitZone d100.BodyPart))

	// Menus / Modals / Windows
	OpenInventoryForManagement(stack []Item)
	OpenInventoryForSelection(stack []Item, prompt string, onSelected func(item Item))
	OpenTextWindow(description string)
	ShowTextFileFullscreen(filename string, onClose func())
	OpenMenu(actions []MenuItem)
	OpenMenuWithTitle(title string, actions []MenuItem)
	OpenKeypad(specialAction string, correctSequence []rune, onSpecialAction func() bool, onCompletion func(success bool))
	OpenVendorMenu(title string, itemsForSale []Item, buyItem func(ui Item, price int), onClose func())
	ShowGameOver(score ScoreInfo, highScores []ScoreInfo)
	ShowTakeOnlyContainer(name string, containedItems []Item, transfer func(ui Item))
	ShowGiveAndTakeContainer(leftName string, leftItems []Item, rightName string, rightItems []Item, transferToLeft func(itemTaken Item, amount int), transferToRight func(itemTaken Item, amount int), takeAll func())
	OpenAimedShotPicker(actorAt ActorForUI, previousAim d100.BodyPart, onSelected func(victim ActorForUI, hitZone d100.BodyPart))

	SaveGame()
	LoadGame()
	// Auto Move Callback
	AfterPlayerMoved(moveInfo MoveInfo)

	// Animations

	// AddAnimations takes a list of list of animations.
	// Each list contains animations that should be played in parallel.
	// The lists are played in order.
	AddAnimations(animations []Animation)
	AnimatePending() (cancelled bool)
	SkipAnimations()
	GetAnimThrow(item Item, origin geometry.Point, target geometry.Point) (Animation, int)
	GetAnimDamage(spreadBlood func(mapPos geometry.Point), actorPos geometry.Point, damage int, bullets int, done func()) Animation
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
	GetAnimRadialExplosion(points map[geometry.Point]int, lightColor fxtools.HDRColor, done func()) Animation
	GetAnimEnchantArmor(actor ActorForUI, position geometry.Point, done func()) Animation
	GetAnimEnchantWeapon(actor ActorForUI, position geometry.Point, done func()) Animation
	GetAnimVorpalizeWeapon(origin geometry.Point, done func()) []Animation
	GetAnimConfuse(position geometry.Point, done func()) Animation
	GetAnimBreath(flight []geometry.Point, done func()) []Animation
	GetAnimBackgroundColor(position geometry.Point, colorName string, frameCount int, done func()) Animation
	GetAnimAppearance(actor ActorForUI, position geometry.Point, done func()) Animation
	GetAnimWakeUp(position geometry.Point, done func()) Animation
	GetAnimEvade(defender ActorForUI, done func()) Animation
	GetAnimLaser(path []geometry.Point, lightColor fxtools.HDRColor, done func()) Animation

	PlayMusic(fileName string)
	PlayCue(cue string)
	SetConversationState(text string, options []MenuItem, conversationPartner ChatterSource, isTerminal bool)
	CloseConversation()
	SetColors(palette textiles.ColorPalette, colors map[ItemCategory]color.RGBA)
	TryAddChatter(victim ChatterSource, text string) bool
	FadeToBlack()
	FadeFromBlack()
	SetSneakOverlay(overlay map[geometry.Point]fxtools.HDRColor)

	AskForConfirmation(title string, message string, onConfirm func(didConfirm bool))
	ForceUIRedraw()
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

func (d Difficulty) LockReductionFactor() float64 {
	switch d {
	case VeryEasy:
		return 1.5
	case Easy:
		return 1
	case Medium:
		return 0.5
	case Hard:
		return 0.35
	case VeryHard:
		return 0.1
	}
	return 1
}

func (d Difficulty) EPicksNeeded() int {
	switch d {
	case VeryEasy:
		return 10
	case Easy:
		return 20
	case Medium:
		return 30
	case Hard:
		return 40
	case VeryHard:
		return 50
	}
	return 1
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

type RangedAttackInfo struct {
	CtH             int
	Mods            d100.CombatModifiers
	SkillUsed       d100.Skill
	SkillBaseValue  int
	DamageBaseValue fxtools.Interval
	Victim          ActorForUI
}

func (r RangedAttackInfo) Skill() d100.Skill {
	return r.SkillUsed
}

func (r RangedAttackInfo) SkillBase() int {
	return r.SkillBaseValue
}

func (r RangedAttackInfo) HitChance() int {
	return r.CtH
}

func (r RangedAttackInfo) DamageBase() fxtools.Interval {
	return r.DamageBaseValue
}

func (r RangedAttackInfo) Modifiers() d100.CombatModifiers {
	return r.Mods
}

func (r RangedAttackInfo) Defender() ActorForUI {
	return r.Victim
}

func (r RangedAttackInfo) DamageWithMods() fxtools.Interval {
	return r.Mods.DamageMods.ApplyForInterval(r.DamageBaseValue)
}

type AttackInfo interface {
	Skill() d100.Skill
	SkillBase() int
	HitChance() int
	DamageBase() fxtools.Interval
	DamageWithMods() fxtools.Interval
	Modifiers() d100.CombatModifiers
	Defender() ActorForUI
}
