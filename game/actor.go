package game

import (
	"RogueUI/d100"
	"RogueUI/foundation"
	"RogueUI/fsmai"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"image/color"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type ActorStance uint8

const (
	Standing ActorStance = iota
	Crawling
	Mounted
	KnockedDown
)

type Actor struct {
	internalName string
	name         string
	charSheet    *d100.CharSheet
	position     geometry.Point

	inventory *Inventory
	vendorInv *Inventory
	equipment *Equipment

	schedule *Schedule

	statusFlags *foundation.ActorFlags
	activeGoal  ActorGoal
	stance      ActorStance

	intrinsicZapEffects []string
	intrinsicUseEffects []string

	icon         textiles.TextIcon
	sizeModifier int
	timeEnergy   int
	body         d100.BodyStructure
	bodyDamage   map[d100.BodyPart]int

	dialogueFile string
	chatterFile  string
	teamName     string

	enemyActors map[string]bool
	enemyTeams  map[string]bool

	isAggressive bool

	audioBaseName string

	xp int

	SpawnPosition           geometry.Point
	currentPathBlockedCount int
	currentPath             []geometry.Point
	currentPathIndex        int
	bodyAugmentations       map[CyberWare]bool
	temporaryStatChanges    []*TemporaryStatChange

	DijkstraMap map[geometry.Point]int
	FoV         map[geometry.Point]bool
	FSM         *ActorFSM
}

func (a *Actor) TimeNeededForAttack() int {
	weapon, hasWeapon := a.GetEquipment().GetMainHandWeapon()
	if !hasWeapon {
		return a.TimeNeededForActions()
	}
	return weapon.GetCurrentAttackMode().TUCost
}

func (a *Actor) TimeEnergy() int {
	return a.timeEnergy
}
func (a *Actor) AddCyberWare(ware CyberWare) {
	a.bodyAugmentations[ware] = false
}

func (a *Actor) RemoveCyberWare(ware CyberWare) {
	delete(a.bodyAugmentations, ware)
}

func (a *Actor) HasCyberWare(ware CyberWare) bool {
	_, hasAug := a.bodyAugmentations[ware]
	return hasAug
}

func (a *Actor) IsCyberWareActive(ware CyberWare) bool {
	if active, hasAug := a.bodyAugmentations[ware]; hasAug {
		return active
	}
	return false
}

func (a *Actor) SetCyberWareActive(ware CyberWare, active bool) {
	a.bodyAugmentations[ware] = active
}

func (a *Actor) GetState() fsmai.StateName {
	return a.FSM.State()
}

func (a *Actor) GetBodyPartIndex(aim d100.BodyPart) int {
	structure := a.body
	for i, part := range structure {
		if part == aim {
			return i
		}
	}
	return -1
}

func (a *Actor) GetBodyPart(index int) d100.BodyPart {
	structure := a.body
	if index < 0 || index >= len(structure) {
		return d100.Body
	}
	return structure[index]
}

// GobEncode encodes the Actor struct into a byte slice.
func (a *Actor) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// Encode each field of the struct
	err := encoder.Encode(a.internalName)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.name)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.charSheet)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.position)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.inventory)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.equipment)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.statusFlags)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.intrinsicZapEffects)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.intrinsicUseEffects)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.icon)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.sizeModifier)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.timeEnergy)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.body)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.FSM)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.dialogueFile)
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(a.chatterFile)
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(a.teamName)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.enemyActors)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.enemyTeams)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.xp)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(a.bodyDamage)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GobDecode decodes a byte slice into an Actor struct.
func (a *Actor) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)

	// Decode each field of the struct
	err := decoder.Decode(&a.internalName)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.name)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.charSheet)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.position)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.inventory)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.equipment)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.statusFlags)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.intrinsicZapEffects)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.intrinsicUseEffects)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.icon)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.sizeModifier)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.timeEnergy)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.body)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.FSM)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.dialogueFile)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.chatterFile)
	if err != nil {
		return err
	}

	err = decoder.Decode(&a.teamName)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.enemyActors)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.enemyTeams)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.xp)
	if err != nil {
		return err
	}
	err = decoder.Decode(&a.bodyDamage)
	if err != nil {
		return err
	}

	return nil
}

func NewPlayer(name string, icon textiles.TextIcon, character *d100.CharSheet) *Actor {
	player := NewActor()
	player.SetCharSheet(character)
	player.SetDisplayName(name)
	player.SetIcon(icon)
	player.SetInternalName("player")
	player.teamName = "player"
	return player
}

var NoGoal = ActorGoal{}

func NewActor() *Actor {
	sheet := d100.NewCharSheet()

	a := &Actor{
		name: "Unknown",
		icon: textiles.TextIcon{
			Char: '0',
			Fg:   color.RGBA{255, 255, 255, 255},
			Bg:   color.RGBA{0, 0, 0, 255},
		},
		equipment:         NewEquipment(),
		charSheet:         sheet,
		bodyAugmentations: make(map[CyberWare]bool),
		body:              d100.HumanBodyParts,
		bodyDamage:        make(map[d100.BodyPart]int),
		statusFlags:       foundation.NewActorFlags(),
		enemyActors:       make(map[string]bool),
		enemyTeams:        make(map[string]bool),
		activeGoal:        NoGoal,
		audioBaseName:     "human_male",
		FoV:               make(map[geometry.Point]bool),
	}
	a.inventory = NewInventory(23, a.Position)

	// Hookup automatic un-equip on removal
	a.GetInventory().SetOnBeforeRemove(a.GetEquipment().UnEquip)

	return a
}

func (a *Actor) SetDialogueFile(scriptName string) {
	a.dialogueFile = scriptName
}

func (a *Actor) GetBodyPartsAndHitChances(baseHitChance int, isMelee bool) []fxtools.Tuple3[d100.BodyPart, bool, int] {
	var result []fxtools.Tuple3[d100.BodyPart, bool, int]
	for _, part := range a.body {
		penalty := part.AimPenalty()
		if isMelee {
			penalty /= 2
		}
		hitChanceOnBodyPart := baseHitChance + penalty
		result = append(result, fxtools.Tuple3[d100.BodyPart, bool, int]{Item1: part, Item2: a.IsCrippled(part), Item3: hitChanceOnBodyPart})
	}
	return result
}

type CapModifier struct {
	maxValue    int
	doesApply   func() bool
	persistent  bool
	description string
}

func (c CapModifier) Description() string {
	return c.description
}

func (c CapModifier) Apply(i int) int {
	if i > c.maxValue {
		return c.maxValue
	}
	return i
}

func (c CapModifier) IsApplicable() bool {
	return c.doesApply()
}

func (c CapModifier) IsPersistent() bool {
	return c.persistent
}

func (c CapModifier) SortOrder() int {
	return 3
}

type FlatModifier struct {
	flatMod     int
	doesApply   func() bool
	persistent  bool
	description string
}

func (f FlatModifier) Description() string {
	return f.description
}

func (f FlatModifier) Apply(i int) int {
	return i + f.flatMod
}

func (f FlatModifier) IsApplicable() bool {
	return f.doesApply()
}

func (f FlatModifier) IsPersistent() bool {
	return f.persistent
}

func (f FlatModifier) SortOrder() int {
	return 1
}

type PercentageModifier struct {
	factor      float64
	doesApply   func() bool
	persistent  bool
	description string
}

func (p PercentageModifier) Description() string {
	return p.description
}

func (p PercentageModifier) Apply(i int) int {
	return int(float64(i) * p.factor)
}

func (p PercentageModifier) IsApplicable() bool {
	return p.doesApply()
}

func (p PercentageModifier) IsPersistent() bool {
	return p.persistent
}

func (p PercentageModifier) SortOrder() int {
	return 0
}

func ModHalveWhen(reason string, isInjured func() bool) PercentageModifier {
	return PercentageModifier{
		factor:      0.5,
		doesApply:   isInjured,
		persistent:  true,
		description: fmt.Sprintf("1/2 - %s", reason),
	}
}

func (a *Actor) Icon() textiles.TextIcon {
	if a.IsSleeping() || a.IsKnockedDown() {
		originalRune := a.icon.Char
		asLower := strings.ToLower(string(originalRune))
		return a.icon.WithRune([]rune(asLower)[0])
	}
	return a.icon
}
func (a *Actor) GetListInfo() string {
	hp := a.charSheet.GetHitPoints()
	hpMax := a.charSheet.GetHitPointsMax()
	damage := a.GetMainHandDamageAsString()
	return fmt.Sprintf("%s HP: %d/%d Dmg: %s DR: %d", a.name, hp, hpMax, damage, a.GetDamageResistance())
}

func (a *Actor) GetMainHandDamageAsString() string {
	item, hasMainHandItem := a.GetEquipment().GetMainHandWeapon()
	damage := strconv.Itoa(a.GetMeleeDamageBonus())
	if hasMainHandItem && item.IsWeapon() {
		damage = item.GetWeaponDamage().ShortString()
	}
	return damage
}

func (a *Actor) IsStunned() bool {
	return a.HasFlag(foundation.FlagStun)
}

func (a *Actor) Position() geometry.Point {
	return a.position
}
func (a *Actor) SetPosition(pos geometry.Point) {
	a.position = pos
}

func (a *Actor) Name() string {
	if a.HasFlag(foundation.FlagInvisible) {
		return "something"
	}
	return a.name
}

func (a *Actor) IsVisible(playerCanSeeInvisible bool) bool {
	return !a.HasFlag(foundation.FlagInvisible) || playerCanSeeInvisible
}

func (a *Actor) GetInventory() *Inventory {
	return a.inventory
}

func (a *Actor) GetEquipment() *Equipment {
	return a.equipment
}

func (a *Actor) GetDamageResistance() int {
	return 0
}

func (a *Actor) GetFlags() *foundation.ActorFlags {
	return a.statusFlags
}

func (a *Actor) IsAlive() bool {
	return a.charSheet.IsAlive()
}

func (a *Actor) HasFlag(flag foundation.ActorFlag) bool {
	return a.statusFlags.IsSet(flag) || a.GetEquipment().ContainsFlag(flag)
}

func (a *Actor) TakeDamage(dmg SourcedDamage) (didCripple bool) {
	wasHeavilyInjured := a.IsHeavilyInjured()
	if a.HasFlag(foundation.FlagZombie) && dmg.DamageType != DamageTypeExplosive { // explosive damage works as usual
		// headshots with normal damage kill zombies instantly if the damage is high enough
		if dmg.DamageType == DamageTypeNormal && dmg.DamageAmount > 10 {
			if dmg.BodyPart == d100.Head || dmg.BodyPart == d100.Eyes {
				currentHitPoints := a.GetHitPoints()
				a.charSheet.TakeRawDamage(currentHitPoints)
			} else if dmg.BodyPart == d100.Legs || dmg.BodyPart == d100.Arms {
				return a.addDamageToBodyPart(dmg) // still able to cripple
			}
		}
		return
	}
	a.charSheet.TakeRawDamage(dmg.DamageAmount)

	if !wasHeavilyInjured && a.IsHeavilyInjured() && dmg.Attacker != nil && a.FSM != nil {
		a.FSM.SendEvent(NewHeavilyInjuredEvent(dmg.Attacker))
	}

	return a.addDamageToBodyPart(dmg)
}

func (a *Actor) addDamageToBodyPart(dmg SourcedDamage) (didCripple bool) {
	wasCrippled := a.IsCrippled(dmg.BodyPart)
	a.bodyDamage[dmg.BodyPart] += dmg.DamageAmount
	return !wasCrippled && a.IsCrippled(dmg.BodyPart)
}

func (a *Actor) IsCrippled(part d100.BodyPart) bool {
	return a.bodyDamage[part] > part.DamageForCrippled(a.GetHitPointsMax())
}

func (a *Actor) IsSleeping() bool {
	return a.HasFlag(foundation.FlagSleep)
}

func (a *Actor) WakeUp() {
	a.statusFlags.Unset(foundation.FlagSleep)
}

func (a *Actor) SetIntrinsicZapEffects(effects []string) {
	a.intrinsicZapEffects = effects
}

func (a *Actor) SetIntrinsicUseEffects(effects []string) {
	a.intrinsicUseEffects = effects
}

func (a *Actor) GetIntrinsicZapEffects() []string {
	return a.intrinsicZapEffects
}

func (a *Actor) GetIntrinsicUseEffects() []string {
	return a.intrinsicUseEffects
}
func (a *Actor) RemoveLevelStatusEffects() {
	a.statusFlags.Unset(foundation.FlagSeeFood)
	a.statusFlags.Unset(foundation.FlagSeeMagic)
	a.statusFlags.Unset(foundation.FlagSeeTraps)
}

func (a *Actor) Heal(amount int) {
	a.charSheet.Heal(amount)
}

func (a *Actor) GetInternalName() string {
	return a.internalName
}

func (a *Actor) SetInternalName(name string) {
	a.internalName = name
}
func (a *Actor) GetHitPoints() int {
	return a.charSheet.GetHitPoints()
}

func (a *Actor) GetHitPointsMax() int {
	return a.charSheet.GetHitPointsMax()
}

func (a *Actor) IsInjured() bool {
	hpMax := a.GetHitPointsMax()
	hpCurrent := a.GetHitPoints()
	belowOneThirdHitPoints := hpCurrent < hpMax/3
	return belowOneThirdHitPoints
}
func (a *Actor) IsFatigued() bool {
	fpMax := a.charSheet.GetActionPointsMax()
	fpCurrent := a.charSheet.GetActionPoints()
	belowOneThird := fpCurrent < fpMax/3
	return belowOneThird
}

func (a *Actor) TextIcon(bg color.RGBA) textiles.TextIcon {
	return a.Icon().WithBg(bg)
}

func (a *Actor) GetDetailInfo() string {

	var result []string
	result = append(result, fmt.Sprintf("Name: %s", a.Name()))
	result = a.appendStateInfo(result)
	// melee attack
	statRows := []fxtools.TableRow{
		fxtools.TableRow{Columns: []string{"Str:", fmt.Sprintf("%d", a.charSheet.GetStat(d100.Strength))}},
		fxtools.TableRow{Columns: []string{"Per:", fmt.Sprintf("%d", a.charSheet.GetStat(d100.Perception))}},
		fxtools.TableRow{Columns: []string{"End:", fmt.Sprintf("%d", a.charSheet.GetStat(d100.Endurance))}},
		fxtools.TableRow{Columns: []string{"Cha:", fmt.Sprintf("%d", a.charSheet.GetStat(d100.Charisma))}},
		fxtools.TableRow{Columns: []string{"Int:", fmt.Sprintf("%d", a.charSheet.GetStat(d100.Intelligence))}},
		fxtools.TableRow{Columns: []string{"Agi:", fmt.Sprintf("%d", a.charSheet.GetStat(d100.Agility))}},
	}

	derivedStatRows := []fxtools.TableRow{
		{Columns: []string{"HP:", fmt.Sprintf("%d/%d", a.GetHitPoints(), a.GetHitPointsMax())}},
		{Columns: []string{"AP:", fmt.Sprintf("%d/%d", a.charSheet.GetActionPoints(), a.charSheet.GetActionPointsMax())}},
		{Columns: []string{"Speed:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(d100.Speed))}},
		{Columns: []string{"Dodge:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(d100.Dodge))}},
		{Columns: []string{"Crit. Chance:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(d100.CriticalChance))}},
		{Columns: []string{"Carry Weight:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(d100.CarryWeight))}},
	}

	resistanceRows := []fxtools.TableRow{
		{Columns: []string{"Physical:", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(d100.DamageResistance))}},
		{Columns: []string{"Poison :", fmt.Sprintf("%d", a.charSheet.GetDerivedStat(d100.PoisonResistance))}},
	}

	var skillRows []fxtools.TableRow
	for skillNo := 0; skillNo < d100.SkillCount(); skillNo++ {
		skill := d100.Skill(skillNo)
		skillRows = append(skillRows, fxtools.TableRow{Columns: []string{skill.String() + ":", fmt.Sprintf("%d", a.charSheet.GetSkill(skill))}})
	}

	statLines := fxtools.TableLayout(statRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	skillLines := fxtools.TableLayout(skillRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	derivedLines := fxtools.TableLayout(derivedStatRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	resistanceLines := fxtools.TableLayout(resistanceRows, []fxtools.TextAlignment{fxtools.AlignLeft, fxtools.AlignLeft})
	result = append(result, "", "> Stats:")
	result = append(result, statLines...)
	result = append(result, "", "> Derived Stats:")
	result = append(result, derivedLines...)
	result = append(result, "", "> Resistances:")
	result = append(result, resistanceLines...)
	result = append(result, "", "> Skills:")
	result = append(result, skillLines...)
	return strings.Join(result, "\n")
}

func (a *Actor) GetEncumbrance() int {
	if !a.GetEquipment().HasArmorEquipped() {
		return 0
	}
	armorEncumbrance := a.GetEquipment().GetEncumbranceFromArmor()
	return armorEncumbrance
}

func (a *Actor) GetSizeModifier() int {
	return a.sizeModifier
}

func (a *Actor) SetSizeModifier(modifier int) {
	a.sizeModifier = modifier
}
func (a *Actor) HasGold(price int) bool {
	return a.GetInventory().HasItemWithNameAndCount("gold", price)
}

func (a *Actor) RemoveGold(price int) []foundation.Item {
	return a.GetInventory().RemoveItemsByNameAndCount("gold", price)
}

func (a *Actor) GetGold() int {
	gold := a.GetInventory().GetItemByName("gold")
	if gold == nil {
		return 0
	}
	return gold.StackSize()
}

func (a *Actor) IsWounded() bool {
	return a.GetHitPoints() < a.GetHitPointsMax()
}

func (a *Actor) IsHungry() bool {
	return a.statusFlags.Get(foundation.FlagHunger) > 0
}

func (a *Actor) IsStarving() bool {
	return a.statusFlags.Get(foundation.FlagHunger) > 1
}

func (a *Actor) Satiate() {
	a.statusFlags.Unset(foundation.FlagHunger)
	a.statusFlags.Unset(foundation.FlagTurnsSinceEating)
}

func (a *Actor) SetSleeping() {
	flags := a.GetFlags()
	flags.Set(foundation.FlagSleep)
}

func (a *Actor) IsBlind() bool {
	return a.HasFlag(foundation.FlagBlind)
}

func (a *Actor) AddTimeEnergy(timeSpent int) {
	a.timeEnergy += timeSpent
}

func (a *Actor) HasEnergyForActions() bool {
	return a.timeEnergy >= a.maximalTimeNeededForActions()
}

func (a *Actor) maximalTimeNeededForActions() int {
	return max(a.TimeNeededForActions(), a.TimeNeededForMovement())
}

func (a *Actor) minimalTimeNeededForAction() int {
	return min(a.TimeNeededForActions(), a.TimeNeededForMovement())
}
func (a *Actor) timeNeededForMeleeAttack() int {
	if meleeWeapon, hasWeapon := a.GetEquipment().GetMainHandWeapon(); hasWeapon {
		return meleeWeapon.GetCurrentAttackMode().TUCost
	}
	return a.TimeNeededForActions()
}

func (a *Actor) TimeNeededForMovement() int {
	return 100 / a.MovementSpeed()
}

func (a *Actor) MovementSpeed() int {
	speed := a.GetBasicSpeed()

	if a.HasFlag(foundation.FlagRunning) {
		speed *= 6
	}

	if a.IsCrippled(d100.Legs) {
		speed = max(1, speed/2)
	}
	if a.IsOverEncumbered() {
		speed = max(1, speed/2)
	}
	speed = max(1, speed-a.GetEncumbrance())
	return speed
}

func (a *Actor) TimeNeededForActions() int {
	speed := a.GetBasicSpeed()
	if a.IsCrippled(d100.Arms) {
		speed = max(1, speed-2)
	}
	if a.IsCrippled(d100.Eyes) {
		speed = max(1, speed-1)
	}
	speed = max(1, speed-a.GetEncumbrance())
	timeNeeded := 100 / speed
	return timeNeeded
}

func (a *Actor) SpendTimeEnergy(amount int) {
	a.timeEnergy -= amount
}

func (a *Actor) AfterTurn() []string {
	a.GetEquipment().AfterTurn()
	a.decrementStatusEffectCounters()
	wornOff := a.decrementTemporaryStatChanges()
	if a.HasFlag(foundation.FlagRunning) {
		sheet := a.GetCharSheet()
		if sheet.GetActionPoints() > 0 {
			sheet.LooseActionPoints(1)
		} else {
			a.UnsetFlag(foundation.FlagRunning)
		}
	}
	return wornOff
}

func (a *Actor) decrementStatusEffectCounters() {
	flags := a.GetFlags()
	flags.Decrement(foundation.FlagHaste)
	flags.Decrement(foundation.FlagSlow)
	flags.Decrement(foundation.FlagConfused)
	flags.Decrement(foundation.FlagFly)
	flags.Decrement(foundation.FlagSeeInvisible)
	flags.Decrement(foundation.FlagHallucinating)
}

func (a *Actor) decrementTemporaryStatChanges() []string {
	var wornOffEffects []string
	for i := len(a.temporaryStatChanges) - 1; i >= 0; i-- {
		statChange := a.temporaryStatChanges[i]
		statChange.TurnsLeft--
		if statChange.TurnsLeft <= 0 {
			wornOffEffects = append(wornOffEffects, statChange.Name)
			a.temporaryStatChanges = append(a.temporaryStatChanges[:i], a.temporaryStatChanges[i+1:]...)
		}
	}
	return wornOffEffects
}
func (a *Actor) GetBasicSpeed() int {
	return max(1, a.charSheet.GetDerivedStat(d100.Speed))
}

func (a *Actor) GetCharSheet() *d100.CharSheet {
	return a.charSheet
}

func (a *Actor) Kill() {
	a.charSheet.Kill()
}

func (a *Actor) HasKey(identifier string) bool {
	return a.GetInventory().HasKey(identifier)
}

func (a *Actor) IsInCombat() bool {
	return a.HasActiveGoal() && a.activeGoal.IsCombatGoal()
}

func (a *Actor) SetDisplayName(name string) {
	a.name = name
}

func (a *Actor) GetDialogueFile() string {
	return a.dialogueFile
}

func (a *Actor) tryEquipWeapon() {
	if !a.GetEquipment().HasWeaponEquipped() {
		weapon := a.GetInventory().GetBestWeapon()
		if weapon != nil {
			a.GetEquipment().Equip(weapon)
		}
	}
}

func (a *Actor) tryEquipRangedWeapon() {
	if !a.GetEquipment().HasRangedWeaponEquipped() {
		weapon := a.GetInventory().GetBestRangedWeapon()
		if weapon != nil {
			a.GetEquipment().Equip(weapon)
		}
	}
}
func (a *Actor) tryEquipMeleeWeapon() {
	if !a.GetEquipment().HasMeleeWeaponEquipped() {
		weapon := a.GetInventory().GetBestMeleeWeapon()
		if weapon != nil {
			a.GetEquipment().Equip(weapon)
		}
	}
}

func (a *Actor) GetHitAudioCue(isMelee bool) string {
	audioName := a.getAudioName()
	hitType := "HIT"
	if isMelee {
		hitType = "MELEE_HIT"
	}
	return fmt.Sprintf("critters/%s/%s", audioName, hitType)
}

func (a *Actor) GetDeathAudioCue() string {
	audioName := a.getAudioName()
	return fmt.Sprintf("critters/%s/FALLING", audioName)
}
func (a *Actor) GetDeathCriticalAudioCue(mode TargetingMode, damageType DamageType) string {
	audioName := a.getAudioName()
	actionName := "FALLING"
	switch damageType {
	case DamageTypeNormal:
		switch mode {
		case TargetingModeFireBurst:
			actionName = "PERFORATED_DEATH"
		default:
			if rand.Intn(2) == 0 {
				actionName = "HOLE_IN_BODY"
			} else {
				actionName = "RIPPING_APART"
			}
		}
	case DamageTypeLaser:
		actionName = "SLICE_IN_TWO"
	case DamageTypeFire:
		if rand.Intn(2) == 0 { // TODO: not both always available, fallbacks or tests needed..
			actionName = "BURNED"
		} else {
			actionName = "BURNING_DANCE"
		}
	case DamageTypeExplosive:
		actionName = "BLOW_EXPLOSION"
	case DamageTypeElectrical:
		if rand.Intn(2) == 0 {
			actionName = "ELECTRIC_BURNED"
		} else {
			actionName = "ELECTRIC_BURNED_TO_ASHES"
		}
	case DamageTypePlasma:
		actionName = "MELTDOWN"
	default:
		actionName = "FALLING"
	}
	return fmt.Sprintf("critters/%s/%s", audioName, actionName)
}
func (a *Actor) GetDodgedAudioCue() string {
	audioName := a.getAudioName()
	return fmt.Sprintf("critters/%s/DODGE", audioName)
}
func (a *Actor) GetMeleeDamageBonus() int {
	return a.charSheet.GetDerivedStat(d100.MeleeDamageBonus)
}
func (a *Actor) GetMeleeAudioCue(isKick bool) string {
	audioName := a.getAudioName()
	hitType := "PUNCH"
	if isKick {
		hitType = "KICK"
	}
	return fmt.Sprintf("critters/%s/%s", audioName, hitType)
}
func (a *Actor) getAudioName() string {
	return a.audioBaseName
}

func (a *Actor) GetTeam() string {
	return a.teamName
}

func (a *Actor) AddToEnemyActors(name string) {
	if a.internalName == name {
		return
	}
	a.enemyActors[name] = true
}

func (a *Actor) AddToEnemyTeams(name string) {
	if a.teamName == name {
		return
	}
	a.enemyTeams[name] = true
}

func (a *Actor) IsHostileTowards(attacker *Actor) bool {
	if a.activeGoal.IsHostilityTowards(attacker) {
		return true
	}
	if a.IsAggressive() && attacker.teamName != a.teamName {
		return true
	}
	if _, exists := a.enemyActors[attacker.GetInternalName()]; exists {
		return true
	}
	if _, exists := a.enemyTeams[attacker.GetTeam()]; exists {
		return true
	}
	return false
}

func (a *Actor) IsPanicking() bool {
	return a.FSM.State() == fsmai.StatePanic
}

func (a *Actor) LookInfo() string {
	if !a.IsAlive() {
		return fmt.Sprintf("%s (dead)", a.Name())
	}
	if a.IsSleeping() {
		return fmt.Sprintf("%s (sleeping)", a.Name())
	}
	if a.IsKnockedDown() {
		return fmt.Sprintf("%s (knocked down)", a.Name())
	}

	displayName := a.Name()
	if a.IsWounded() {
		displayName = fmt.Sprintf("%s (%s)", a.Name(), a.injuredString())
	}

	return a.ActionDescription(displayName) + "\n" + a.OutfitDescription()
}

func (a *Actor) ActionDescription(displayName string) string {
	if a.HasActiveGoal() {
		return a.activeGoal.Description(displayName)
	}
	if a.schedule != nil && a.schedule.LastSlotID.Index != -1 {
		return a.schedule.CurrentTimeSlot().Description(displayName)
	}
	return fmt.Sprintf("%s is standing there", displayName)
}

func (a *Actor) OutfitDescription() string {
	armor := a.GetEquipment().GetArmor()
	helmet := a.GetEquipment().GetHelmet()
	clothes := ""
	if armor == nil && helmet == nil {
		clothes = fmt.Sprintf("%s is not wearing anything", a.Name())
	}
	if armor != nil && helmet != nil {
		clothes = fmt.Sprintf("%s is wearing %s and %s", a.Name(), armor.Name(), helmet.Name())
	}
	if armor != nil {
		clothes = fmt.Sprintf("%s is wearing %s", a.Name(), armor.Name())
	}
	if helmet != nil {
		clothes = fmt.Sprintf("%s is wearing %s", a.Name(), helmet.Name())
	}

	hands := ""
	item, hasItem := a.GetEquipment().GetMainHandItem()
	if hasItem {
		hands = fmt.Sprintf("%s is holding %s", a.Name(), item.Name())
	} else {
		hands = fmt.Sprintf("%s is not holding anything", a.Name())
	}

	return fmt.Sprintf("%s\n%s", clothes, hands)
}

func (a *Actor) HasDialogue() bool {
	return a.dialogueFile != ""
}

func (a *Actor) HasStealableItems() bool {
	return a.GetInventory().HasStealableItems(a.GetEquipment().IsNotEquipped)
}

func (a *Actor) IsKnockedDown() bool {
	return a.HasFlag(foundation.FlagKnockedDown)
}

func (a *Actor) injuredString() string {
	percent := a.GetHitPoints() * 100 / a.GetHitPointsMax()
	if percent < 10 {
		return "near death"
	}
	if percent < 25 {
		return "severely wounded"
	}
	if percent < 50 {
		return "badly injured"
	}
	if percent < 75 {
		return "injured"
	}
	return "scratched"
}

func (a *Actor) RemoveEnemy(other *Actor) {
	delete(a.enemyActors, other.GetInternalName())
}

func (a *Actor) SetIcon(icon textiles.TextIcon) {
	a.icon = icon
}

func (a *Actor) SetCharSheet(character *d100.CharSheet) {
	a.charSheet = character
}

func (a *Actor) SetXP(xp int) {
	a.xp = xp
}

func (a *Actor) ToRecord() recfile.Record {
	actorRecord := append(recfile.Record{
		recfile.Field{Name: "Name", Value: a.name},
		recfile.Field{Name: "InternalName", Value: a.internalName},
		recfile.Field{Name: "Icon", Value: string(a.icon.Char)},
		recfile.Field{Name: "Fg", Value: recfile.RGBStr(a.icon.Fg)},
		recfile.Field{Name: "Bg", Value: recfile.RGBStr(a.icon.Bg)},
		recfile.Field{Name: "DialogueFile", Value: a.dialogueFile},
		recfile.Field{Name: "Team", Value: a.teamName},
		recfile.Field{Name: "XP", Value: recfile.IntStr(a.xp)},
	}, a.charSheet.ToRecord()...)
	return actorRecord
}

func (a *Actor) ActOnGoal(g *GameState) (fsmai.TransitionEvent, int) {
	if a.activeGoal.IsEmpty() {
		return fsmai.NoEvent, 0
	}
	if a.activeGoal.Achieved(g, a) {
		a.activeGoal = NoGoal
		return fsmai.NoEvent, 0
	}
	event, tuSpent := a.activeGoal.Action(g, a)
	if a.activeGoal.Achieved(g, a) {
		a.activeGoal = NoGoal
	}
	return event, tuSpent
}

func (a *Actor) HasActiveGoal() bool {
	return !a.activeGoal.IsEmpty()
}

func (a *Actor) GetMeleeTUCost() int {
	if meleeWeapon, hasWeapon := a.GetEquipment().GetMainHandWeapon(); hasWeapon {
		return meleeWeapon.GetCurrentAttackMode().TUCost
	}
	return a.TimeNeededForActions()
}

func (a *Actor) SetGoal(goal ActorGoal) {
	a.activeGoal = goal
}

func (a *Actor) GetWeaponRange() int {
	if rangedWeapon, hasWeapon := a.GetEquipment().GetRangedWeapon(); hasWeapon {
		return rangedWeapon.GetCurrentAttackMode().MaxRange
	}
	return 1
}
func (a *Actor) GetMaxRepairQuality() d100.Percentage {
	repairSkill := a.GetCharSheet().GetSkill(d100.SkillForRepairs)
	return d100.Percentage(min(100, int(20+(float64(repairSkill)*0.5))))
}

func (a *Actor) GetRepairQuality(qualityOne, qualityTwo d100.Percentage) d100.Percentage {
	var lower, higher float64
	if qualityOne < qualityTwo {
		lower = float64(qualityOne)
		higher = float64(qualityTwo)
	} else {
		lower = float64(qualityTwo)
		higher = float64(qualityOne)
	}
	repairSkill := float64(a.GetCharSheet().GetSkill(d100.SkillForRepairs))
	newQuality := 5 + higher + (0.05 * lower) + (0.15 * repairSkill)
	return min(a.GetMaxRepairQuality(), d100.Percentage(newQuality))
}

func (a *Actor) getMoveTowards(g *GameState, pos geometry.Point) geometry.Point {
	if a.Position() == pos {
		return a.Position()
	}

	if !a.hasPathTo(pos) {
		a.calcAndSetPath(g, pos)
	}

	if a.currentPathIndex < 0 || a.currentPathIndex >= len(a.currentPath) {
		return a.Position()
	}

	nextStep := a.currentPath[a.currentPathIndex]
	if !g.currentMap().IsWalkableFor(nextStep, a) {
		a.calcAndSetPath(g, pos)
		if a.currentPathIndex == -1 {
			return a.Position()
		}
		nextStep = a.currentPath[a.currentPathIndex]
	}
	a.currentPathBlockedCount = 0
	a.currentPathIndex++
	return nextStep

}

func (a *Actor) getMoveTowardsActor(g *GameState, other *Actor) geometry.Point {
	moveDist := g.currentMap().MoveDistance(a.Position(), other.Position())
	if moveDist <= 1 {
		return a.Position()
	}

	nextStep := g.currentMap().GetMoveOnOtherDijkstraMap(a.Position(), true, other.DijkstraMap)

	if !g.currentMap().IsWalkableFor(nextStep, a) {
		a.currentPathBlockedCount++
		if a.currentPathBlockedCount <= 3 {
			return a.Position()
		}
	}
	a.currentPathBlockedCount = 0
	a.currentPathIndex++
	return nextStep
}

func (a *Actor) getMoveAwayFromActor(g *GameState, other *Actor) geometry.Point {
	nextStep := g.currentMap().GetMoveOnOtherDijkstraMap(a.Position(), false, other.DijkstraMap)

	if !g.currentMap().IsWalkableFor(nextStep, a) {
		a.currentPathBlockedCount++
		if a.currentPathBlockedCount <= 3 {
			return a.Position()
		}
	}
	a.currentPathBlockedCount = 0
	a.currentPathIndex++
	return nextStep
}

func (a *Actor) calcAndSetPath(g *GameState, pos geometry.Point) {
	a.currentPath = nil
	a.currentPathBlockedCount = 0
	calcPath := g.currentMap().GetJPSPath(a.Position(), pos, func(point geometry.Point) bool {
		return g.currentMap().IsWalkableFor(point, a)
	})
	if len(calcPath) == 0 || (len(calcPath) == 1 && calcPath[0] == a.Position()) {
		a.currentPathIndex = -1
	} else {
		a.currentPathIndex = 0
		a.currentPath = calcPath
	}
}
func (a *Actor) cannotFindPath() bool {
	return a.currentPathIndex == -1
}
func (a *Actor) hasPathTo(pos geometry.Point) bool {
	if a.currentPath == nil || len(a.currentPath) == 0 {
		return false
	}
	if a.currentPathIndex < 0 || a.currentPathIndex >= len(a.currentPath) {
		return false
	}
	targetOfPath := a.currentPath[len(a.currentPath)-1]
	isNear := geometry.DistanceChebyshev(targetOfPath, pos) <= 1
	return isNear
}

func (a *Actor) GetMaxThrowRange() int {
	strength := a.GetCharSheet().GetStat(d100.Strength)
	return strength * 2
}

func (a *Actor) GetXP() int {
	return a.xp
}

func (a *Actor) TryEquipRangedWeaponFirst() {
	a.tryEquipRangedWeapon()
	if !a.GetEquipment().HasRangedWeaponInMainHand() {
		a.tryEquipWeapon()
	}
}

func (a *Actor) IsOverEncumbered() bool {
	carryWeight := a.GetCharSheet().GetDerivedStat(d100.CarryWeight)
	totalWeight := a.GetInventory().GetTotalWeight()
	return totalWeight > carryWeight
}

func (a *Actor) SetStance(stance ActorStance) {
	a.stance = stance
}

func (a *Actor) RemoveGoal() {
	a.activeGoal = NoGoal
}

func (a *Actor) IsAlliedWith(player *Actor) bool {
	return a.teamName == player.teamName
}

func (a *Actor) SetChatterFile(value string) {
	a.chatterFile = value
}

func (a *Actor) UnsetFlag(flag foundation.ActorFlag) {
	a.statusFlags.Unset(flag)
}

func (a *Actor) SetFlag(flag foundation.ActorFlag) {
	a.statusFlags.Set(flag)
}

func (a *Actor) GetTemporarySkillModifiers(skill d100.Skill) []d100.Modifier {
	var result []d100.Modifier
	for _, statChange := range a.temporaryStatChanges {
		if value, exists := statChange.SkillChanges[skill]; exists {
			result = append(result, d100.DefaultModifier{
				Source:    statChange.Name,
				Modifier:  value,
				Order:     1,
				IsPercent: true,
				Suffix:    fmt.Sprintf("(%d turns left)", statChange.TurnsLeft),
			})
		}
	}
	return result
}

func (a *Actor) GetTemporaryStatModifiers(stat d100.Stat) []d100.Modifier {
	var result []d100.Modifier
	for _, statChange := range a.temporaryStatChanges {
		if value, exists := statChange.StatChanges[stat]; exists {
			result = append(result, d100.DefaultModifier{
				Source:   statChange.Name,
				Modifier: value,
				Order:    1,
				Suffix:   fmt.Sprintf("(%d turns left)", statChange.TurnsLeft),
			})
		}
	}
	return result
}

func (a *Actor) GetTemporaryDerivedStatModifiers(stat d100.DerivedStat) []d100.Modifier {
	var result []d100.Modifier
	for _, statChange := range a.temporaryStatChanges {
		if value, exists := statChange.DerivedStatChanges[stat]; exists {
			result = append(result, d100.DefaultModifier{
				Source:   statChange.Name,
				Modifier: value,
				Order:    1,
				Suffix:   fmt.Sprintf("(%d turns left)", statChange.TurnsLeft),
			})
		}
	}
	return result
}

func (a *Actor) AddTemporaryStatChange(change *TemporaryStatChange) {
	for i, existingChange := range a.temporaryStatChanges {
		if existingChange.Name == change.Name {
			a.temporaryStatChanges[i] = change
			return
		}
	}
	a.temporaryStatChanges = append(a.temporaryStatChanges, change)
}

func (a *Actor) GetVendorInventory() *Inventory {
	return a.vendorInv
}

func (a *Actor) GetMeleeSkillUsed() d100.Skill {
	if meleeWeapon, hasWeapon := a.GetEquipment().GetMainHandWeapon(); hasWeapon {
		return meleeWeapon.GetSkillUsed()
	}
	return d100.SkillForUnarmed
}

func (a *Actor) DetectionRange() int {
	return a.GetCharSheet().GetStat(d100.Perception) + 2
}

func (a *Actor) CanSee(pos geometry.Point) bool {
	if geometry.DistanceChebyshev(a.Position(), pos) <= 1 {
		return true
	}
	return a.inFov(pos)
}

func (a *Actor) ResetFov() {
	clear(a.FoV)
}
func (a *Actor) inFov(pos geometry.Point) bool {
	value, exists := a.FoV[pos]
	return exists && value // new

}

func (a *Actor) Visibles() []geometry.Point {
	var visibles []geometry.Point
	for pos, value := range a.FoV {
		if value {
			visibles = append(visibles, pos)
		}
	}
	return visibles

}

func (a *Actor) SetVisible(point geometry.Point) {
	a.FoV[point] = true
}

func (a *Actor) GetAllNeighbors() []geometry.Point {
	var neighbors []geometry.Point
	for pos, dist := range a.DijkstraMap {
		if dist == 10 || dist == 14 {
			neighbors = append(neighbors, pos)
		}
	}
	return neighbors
}

func (a *Actor) MoveToNextTimeSlot(time time.Time) (TimeSlot, bool) {
	if a.schedule == nil {
		return TimeSlot{}, false
	}
	return a.schedule.MoveToNextTimeSlot(time)
}

func (a *Actor) appendStateInfo(result []string) []string {
	fsmState := a.FSM.currentBehavior.AssociatedState().ToString()
	goal := a.activeGoal.Description(a.name)
	result = append(result, fmt.Sprintf("State: %s", fsmState))
	result = append(result, fmt.Sprintf("Goal: %s", goal))
	return result
}

func (a *Actor) IsHeavilyInjured() bool {
	return a.GetHitPoints() < a.GetHitPointsMax()/3
}

func (a *Actor) IsAggressive() bool {
	return a.isAggressive
}

func (a *Actor) SetNeutral() {
	a.isAggressive = false
}

func (a *Actor) SetAggressive() {
	a.isAggressive = true
}

func (a *Actor) HasPerk(perk d100.Perk) bool {
	return a.charSheet.HasPerk(perk)
}

func (a *Actor) GetPerkLevel(perk d100.Perk) int {
	return a.charSheet.GetPerkLevel(perk)
}

type StatChange struct {
	StatChanges        map[d100.Stat]int
	SkillChanges       map[d100.Skill]int
	DerivedStatChanges map[d100.DerivedStat]int
}

type TemporaryStatChange struct {
	StatChange
	Name      string
	TurnsLeft int
}

type ActorGoal struct {
	Action             func(g *GameState, a *Actor) (fsmai.TransitionEvent, int)
	Achieved           func(g *GameState, a *Actor) bool
	ActionFormatString string
	IsCombat           bool
	IsCombatTarget     func(attacker *Actor) bool
}

func (g ActorGoal) IsEmpty() bool {
	return g.Action == nil && g.Achieved == nil
}

func (g ActorGoal) Description(actorName string) string {
	if g.IsEmpty() {
		return "no goal"
	}
	return fmt.Sprintf(g.ActionFormatString, actorName)
}

func (g ActorGoal) IsCombatGoal() bool {
	return g.IsCombat
}

func (g ActorGoal) IsHostilityTowards(attacker *Actor) bool {
	if !g.IsCombatGoal() {
		return false
	}
	return g.IsCombatTarget(attacker)
}

func GoalMoveToSpawn() ActorGoal {
	return ActorGoal{
		Action: func(g *GameState, a *Actor) (fsmai.TransitionEvent, int) {
			targetPos := a.SpawnPosition
			return moveTowards(g, a, targetPos)
		},
		Achieved: func(g *GameState, a *Actor) bool {
			return a.Position() == a.SpawnPosition
		},
		ActionFormatString: "%s is moving",
	}
}

func GoalMoveIntoShootingRange(target *Actor) ActorGoal {
	return ActorGoal{
		Action: func(g *GameState, a *Actor) (fsmai.TransitionEvent, int) {
			return moveTowardsActor(g, a, target)
		},
		Achieved: func(g *GameState, a *Actor) bool {
			return g.IsInShootingRange(a, target)
		},
		ActionFormatString: "%s is moving aggressively",
		IsCombat:           true,
		IsCombatTarget:     IsActor(target),
	}
}

func IsActor(target *Actor) func(actor *Actor) bool {
	return func(actor *Actor) bool { return actor == target }
}

func GoalFleeFromActor(actor *Actor, threat *Actor) ActorGoal {
	return ActorGoal{
		Action: func(g *GameState, a *Actor) (fsmai.TransitionEvent, int) {
			return moveAwayFromActor(g, a, threat)
		},
		Achieved: func(g *GameState, a *Actor) bool {
			return !threat.IsAlive() || !actor.IsAlive()
		},
		ActionFormatString: "%s is fleeing",
	}
}

func GoalKillActor(attacker *Actor, victim *Actor) ActorGoal {
	return ActorGoal{
		Action: func(g *GameState, a *Actor) (fsmai.TransitionEvent, int) {
			return tryKill(g, a, victim)
		},
		Achieved: func(g *GameState, a *Actor) bool {
			return !victim.IsAlive() || !attacker.IsAlive()
		},
		ActionFormatString: "%s is attacking",
		IsCombat:           true,
		IsCombatTarget:     IsActor(victim),
	}
}

func GoalMoveToLocation(loc geometry.Point) ActorGoal {
	return ActorGoal{
		Action: func(g *GameState, a *Actor) (fsmai.TransitionEvent, int) {
			return moveTowards(g, a, loc)
		},
		Achieved: func(g *GameState, a *Actor) bool {
			return a.Position() == loc
		},
		ActionFormatString: "%s is moving",
	}
}
