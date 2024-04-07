package foundation

import (
	"RogueUI/geometry"
	"RogueUI/rpg"
	"RogueUI/util"
)

// Actions that the User Interface can trigger on the game
type GameForUI interface {
	// Init

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
	PlayerUseOrZapItem(item ItemForUI)
	AimedShot()
	QuickShot()

	PlayerInteractWithMap() // up/down stairs..

	// State Queries

	GetPlayerPosition() geometry.Point
	GetCharacterSheet() []string

	GetHudStats() (map[HudValue]int, []string)
	GetMapInfo(pos geometry.Point) HiLiteString

	GetInventory() []ItemForUI

	GetVisibleEnemies() []ActorForUI
	GetVisibleItems() []ItemForUI
	GetLog() []HiLiteString

	IsSomethingInterestingAtLoc(position geometry.Point) bool
	IsSomethingBlockingTargetingAtLoc(point geometry.Point) bool


	// Inventory Management
	OpenInventory()
	ChooseItemForThrow()
	ChooseItemForUseOrZap()
	ChooseItemForMissileLaunch()

	IsEquipped(item ItemForUI) bool

	// Special
	OpenWizardMenu()

	// Game State
	Reset()

	// Map Drawing
	IsExplored(loc geometry.Point) bool
	IsLit(pos geometry.Point) bool
	IsVisibleToPlayer(loc geometry.Point) bool

	EntityAt(loc geometry.Point) EntityType

	MapAt(loc geometry.Point) TileType
	ItemAt(loc geometry.Point) ItemForUI
	ObjectAt(loc geometry.Point) ObjectCategory
	ActorAt(loc geometry.Point) ActorForUI


	// Level up choices
	IncreaseAttributeLevel(stat rpg.Stat)
	IncreaseSkillLevel(skill rpg.SkillName)


	// Wizard
	Descend()
	Ascend()

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
	// Init
	SetGame(game GameForUI)
	StartGameLoop()
	InitDungeonUI()

	// Notification of state changes
	UpdateStats()
	UpdateInventory()
	UpdateLogWindow()
	UpdateVisibleEnemies()

	// Targeting
	SelectTarget(origin geometry.Point, onSelected func(targetPos geometry.Point))

	// Menus / Modals / Windows
	OpenInventoryForManagement(stack []ItemForUI)
	OpenInventoryForSelection(stack []ItemForUI, prompt string, onSelected func(item ItemForUI))
	OpenTextWindow(description []string)
	ShowTextFileFullscreen(filename string, onClose func())
	OpenMenu(actions []MenuItem)
	OpenVendorMenu(itemsForSale []util.Tuple[ItemForUI, int], buyItem func(ui ItemForUI, price int))
	ShowGameOver(score ScoreInfo, highScores []ScoreInfo)

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
	GetAnimTiles(positions []geometry.Point, frames []TextIcon, done func()) Animation
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
	GetAnimAppearance(actor ActorForUI,position geometry.Point, done func()) Animation

}

type Animation interface {
	IsDone() bool
	SetFollowUp([]Animation)
	RequestMapUpdateOnFinish()
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
	PlayerName   string
	Score        int
	MaxLevel     int
	CauseOfDeath string
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

type TileType string

const (
	TileEmpty                         TileType = "empty"
	TileFloor                                  = "TileFloor"
	TileWall                                   = "TileWall"
	TileRoomFloor                              = "TileRoomFloor"
	TileRoomWallHorizontal                     = "TileRoomWallHorizontal"
	TileRoomWallVertical                       = "TileRoomWallVertical"
	TileRoomWallCornerTopLeft                  = "TileRoomWallCornerTopLeft"
	TileRoomWallCornerTopRight                 = "TileRoomWallCornerTopRight"
	TileRoomWallCornerBottomRight              = "TileRoomWallCornerBottomRight"
	TileRoomWallCornerBottomLeft               = "TileRoomWallCornerBottomLeft"
	TileCorridorFloor                          = "TileCorridorFloor"
	TileCorridorWall                           = "TileCorridorWall"
	TileCorridorWallHorizontal                 = "TileCorridorWallHorizontal"
	TileCorridorWallVertical                   = "TileCorridorWallVertical"
	TileCorridorWallCornerTopLeft              = "TileCorridorWallCornerTopLeft"
	TileCorridorWallCornerTopRight             = "TileCorridorWallCornerTopRight"
	TileCorridorWallCornerBottomRight          = "TileCorridorWallCornerBottomRight"
	TileCorridorWallCornerBottomLeft           = "TileCorridorWallCornerBottomLeft"
	TileWallTJunctionTop                       = "TileWallTJunctionTop"
	TileWallTJunctionRight                     = "TileWallTJunctionRight"
	TileWallTJunctionBottom                    = "TileWallTJunctionBottom"
	TileWallTJunctionLeft                      = "TileWallTJunctionLeft"
	TileDoorOpen                               = "TileDoorOpen"
	TileDoorClosed                             = "TileDoorClosed"
	TileDoorBroken                             = "TileDoorBroken"
	TileDoorLocked                             = "TileDoorLocked"
	TileStairsUp                               = "TileStairsUp"
	TileStairsDown                             = "TileStairsDown"
	TileMountain                               = "TileMountain"
	TileGrass                                  = "TileGrass"
	TileTree                                   = "TileTree"
	TileWater                                  = "TileWater"
	TileLava                                   = "TileLava"
	TileChasm                                  = "TileChasm"
	TileVendorGeneral                          = "TileVendorGeneral"
	TileVendorWeapons                          = "TileVendorWeapons"
	TileVendorArmor                            = "TileVendorArmor"
	TileVendorAlchemist                        = "TileVendorAlchemist"
)
