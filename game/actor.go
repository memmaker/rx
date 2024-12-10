package game

import (
	"contractor/d100"
	"contractor/foundation"
	"contractor/fsmai"
	"contractor/gridmap"
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
	UID gridmap.ActorID

	CharSheet *d100.CharSheet

	InternalName string
	DisplayName  string

	RawPosition geometry.Point
	Stance      ActorStance

	Icon          textiles.TextIcon
	SizeModifier  int
	RawTimeEnergy int
	Body          d100.BodyStructure
	BodyDamage    map[d100.BodyPart]int

	Inventory *Inventory

	StatusFlags          *foundation.ActorFlags
	TemporaryStatChanges []*TemporaryStatChange

	FSM        *ActorFSM
	Schedule   *Schedule
	ActiveGoal ActorGoal

	IntrinsicZapEffects []string
	IntrinsicUseEffects []string

	DialogueFile string
	ChatterFile  string
	TeamName     string

	EnemyActors map[string]bool
	EnemyTeams  map[string]bool

	Aggressive   bool
	GuardingZone string

	AudioBaseName string

	XP int

	SpawnPosition           geometry.Point
	CurrentPathBlockedCount int
	CurrentPath             []geometry.Point
	CurrentPathIndex        int

	BodyAugmentations map[CyberWare]bool
	AugmentToggled    func(CyberWare, bool)

	DijkstraMap map[geometry.Point]int
	FoV         map[geometry.Point]bool

	OffersCyberWare []fxtools.Tuple[CyberWare, int]
	VendorInv       *Inventory
}

func (a *Actor) SetAugmentToggledHandler(f func(CyberWare, bool)) {
	a.AugmentToggled = f
}

func (a *Actor) GetArmorProtectionString() string {
	if !a.GetInventory().HasArmorEquipped() {
		return ""
	}
	armor := a.GetInventory().GetArmor()
	return armor.GetArmorProtectionValueAsString()
}

func (a *Actor) TimeNeededForAttack() int {
	weapon, hasWeapon := a.GetInventory().GetEquippedWeapon()
	if !hasWeapon {
		return a.TimeNeededForActions()
	}
	return weapon.GetCurrentAttackMode().TUCost
}

func (a *Actor) TimeEnergy() int {
	return a.RawTimeEnergy
}
func (a *Actor) AddCyberWare(ware CyberWare) {
	a.BodyAugmentations[ware] = false
}

func (a *Actor) RemoveCyberWare(ware CyberWare) {
	delete(a.BodyAugmentations, ware)
}

func (a *Actor) HasCyberWare(ware CyberWare) bool {
	_, hasAug := a.BodyAugmentations[ware]
	return hasAug
}

func (a *Actor) IsCyberWareActive(ware CyberWare) bool {
	if active, hasAug := a.BodyAugmentations[ware]; hasAug {
		return active
	}
	return false
}

func (a *Actor) SetCyberWareActive(ware CyberWare, active bool) {
	a.BodyAugmentations[ware] = active
	if a.AugmentToggled != nil {
		a.AugmentToggled(ware, active)
	}
}

func (a *Actor) GetState() fsmai.StateName {
	return a.FSM.State()
}

func (a *Actor) GetBodyPartIndex(aim d100.BodyPart) int {
	structure := a.Body
	for i, part := range structure {
		if part == aim {
			return i
		}
	}
	return -1
}

func (a *Actor) GetBodyPart(index int) d100.BodyPart {
	structure := a.Body
	if index < 0 || index >= len(structure) {
		return d100.Body
	}
	return structure[index]
}

func NewPlayer(name string, icon textiles.TextIcon, character *d100.CharSheet) *Actor {
	player := NewActor()
	player.SetCharSheet(character)
	player.SetDisplayName(name)
	player.SetIcon(icon)
	player.SetInternalName("player")
	player.TeamName = "player"
	return player
}

func NewActor() *Actor {
	sheet := d100.NewCharSheet()

	a := &Actor{
		UID:         gridmap.NextActorID(),
		DisplayName: "Unknown",
		Icon: textiles.TextIcon{
			Char: '0',
			Fg:   color.RGBA{255, 255, 255, 255},
			Bg:   color.RGBA{0, 0, 0, 255},
		},
		CharSheet:         sheet,
		BodyAugmentations: make(map[CyberWare]bool),
		Body:              d100.HumanBodyParts,
		BodyDamage:        make(map[d100.BodyPart]int),
		StatusFlags:       foundation.NewActorFlags(),
		EnemyActors:       make(map[string]bool),
		EnemyTeams:        make(map[string]bool),
		AudioBaseName:     "human_male",
		FoV:               make(map[geometry.Point]bool),
	}
	a.Inventory = NewInventory(23)

	return a
}

func (a *Actor) SetDialogueFile(scriptName string) {
	a.DialogueFile = scriptName
}

func (a *Actor) GetBodyPartsAndHitChances(baseHitChance int, isMelee bool) []fxtools.Tuple3[d100.BodyPart, bool, int] {
	var result []fxtools.Tuple3[d100.BodyPart, bool, int]
	for _, part := range a.Body {
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

func (a *Actor) IsOpenCarryWeapon() bool {
	if a.GetInventory().HasWeaponEquipped() {
		return true
	}

	allNonHiddenWeaponsInInventory := a.GetInventory().StackedItemsWithFilter(func(item foundation.Item) bool {
		if !item.IsWeapon() {
			return false
		}
		weapon := item.(*Weapon)
		return weapon.RelativeSize != SizeHidden
	})
	if len(allNonHiddenWeaponsInInventory) == 0 {
		return false
	}
	if !a.GetInventory().HasArmorEquipped() {
		return true
	}
	armor := a.GetInventory().GetArmor()

	asWeapons := fxtools.MapSlice(allNonHiddenWeaponsInInventory, func(item foundation.Item) *Weapon {
		return item.(*Weapon)
	})

	canConceal := armor.CanConceal(asWeapons)
	return !canConceal
}

func (a *Actor) GetIcon() textiles.TextIcon {
	if a.IsSleeping() || a.IsKnockedDown() {
		originalRune := a.Icon.Char
		asLower := strings.ToLower(string(originalRune))
		return a.Icon.WithRune([]rune(asLower)[0])
	}
	return a.Icon
}
func (a *Actor) GetListInfo() string {
	hp := a.CharSheet.GetHitPoints()
	hpMax := a.CharSheet.GetHitPointsMax()
	damage := a.GetMainHandDamageAsString()
	return fmt.Sprintf("%s HP: %d/%d Dmg: %s Armor: %s", a.DisplayName, hp, hpMax, damage, a.GetArmorProtectionString())
}

func (a *Actor) GetMainHandDamageAsString() string {
	item, hasMainHandItem := a.GetInventory().GetEquippedWeapon()
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
	return a.RawPosition
}
func (a *Actor) SetPosition(pos geometry.Point) {
	a.RawPosition = pos
}

func (a *Actor) Name() string {
	if a.HasFlag(foundation.FlagActiveCamouflage) {
		return "something"
	}
	return a.DisplayName
}

func (a *Actor) IsVisible(playerCanSeeInvisible bool) bool {
	return !a.HasFlag(foundation.FlagActiveCamouflage) || playerCanSeeInvisible
}

func (a *Actor) GetInventory() *Inventory {
	return a.Inventory
}

func (a *Actor) GetFlags() *foundation.ActorFlags {
	return a.StatusFlags
}

func (a *Actor) IsAlive() bool {
	return a.CharSheet.IsAlive()
}

func (a *Actor) HasFlag(flag foundation.ActorFlag) bool {
	return a.StatusFlags.IsSet(flag) || a.GetInventory().EquipmentContainsFlag(flag)
}

func (a *Actor) TakeDamage(dmg SourcedDamage) (didCripple bool) {
	wasHeavilyInjured := a.IsHeavilyInjured()
	if a.HasFlag(foundation.FlagZombie) && dmg.DamageType != DamageTypeExplosive { // explosive damage works as usual
		// headshots with normal damage kill zombies instantly if the damage is high enough
		if dmg.DamageType == DamageTypeNormal && dmg.DamageAmount > 10 {
			if dmg.BodyPart == d100.Head || dmg.BodyPart == d100.Eyes {
				currentHitPoints := a.GetHitPoints()
				a.CharSheet.TakeRawDamage(currentHitPoints)
			} else if dmg.BodyPart == d100.Legs || dmg.BodyPart == d100.Arms {
				return a.addDamageToBodyPart(dmg) // still able to cripple
			}
		}
		return
	}

	a.CharSheet.TakeRawDamage(dmg.DamageAmount)

	if dmg.IsKillingBlow {
		a.CharSheet.Kill()
	}

	if !wasHeavilyInjured && a.IsHeavilyInjured() && dmg.Attacker != nil && a.FSM != nil {
		a.FSM.SendEvent(NewHeavilyInjuredEvent(dmg.Attacker))
	}

	for _, statusFlag := range dmg.ApplyStatus {
		a.GetFlags().Increment(statusFlag)
	}

	return a.addDamageToBodyPart(dmg)
}

func (a *Actor) addDamageToBodyPart(dmg SourcedDamage) (didCripple bool) {
	wasCrippled := a.IsCrippled(dmg.BodyPart)

	if dmg.IsCrippling {
		a.Cripple(dmg.BodyPart)
	} else {
		a.BodyDamage[dmg.BodyPart] += dmg.DamageAmount
	}

	return !wasCrippled && a.IsCrippled(dmg.BodyPart)
}

func (a *Actor) Cripple(part d100.BodyPart) {
	a.BodyDamage[part] = part.DamageForCrippled(a.GetHitPointsMax()) + 1
}

func (a *Actor) IsCrippled(part d100.BodyPart) bool {
	return a.BodyDamage[part] > part.DamageForCrippled(a.GetHitPointsMax())
}

func (a *Actor) IsSleeping() bool {
	return a.HasFlag(foundation.FlagSleep)
}

func (a *Actor) WakeUp() {
	a.StatusFlags.Unset(foundation.FlagSleep)
}

func (a *Actor) SetIntrinsicZapEffects(effects []string) {
	a.IntrinsicZapEffects = effects
}

func (a *Actor) SetIntrinsicUseEffects(effects []string) {
	a.IntrinsicUseEffects = effects
}

func (a *Actor) GetIntrinsicZapEffects() []string {
	return a.IntrinsicZapEffects
}

func (a *Actor) GetIntrinsicUseEffects() []string {
	return a.IntrinsicUseEffects
}
func (a *Actor) RemoveLevelStatusEffects() {
	a.StatusFlags.Unset(foundation.FlagSeeFood)
	a.StatusFlags.Unset(foundation.FlagSeeMagic)
	a.StatusFlags.Unset(foundation.FlagSeeTraps)
}

func (a *Actor) Heal(amount int) {
	a.CharSheet.Heal(amount)
}

func (a *Actor) GetInternalName() string {
	return a.InternalName
}

func (a *Actor) SetInternalName(name string) {
	a.InternalName = name
}
func (a *Actor) GetHitPoints() int {
	return a.CharSheet.GetHitPoints()
}

func (a *Actor) GetHitPointsMax() int {
	return a.CharSheet.GetHitPointsMax()
}

func (a *Actor) IsInjured() bool {
	hpMax := a.GetHitPointsMax()
	hpCurrent := a.GetHitPoints()
	belowOneThirdHitPoints := hpCurrent < hpMax/3
	return belowOneThirdHitPoints
}
func (a *Actor) IsFatigued() bool {
	fpMax := a.CharSheet.GetActionPointsMax()
	fpCurrent := a.CharSheet.GetActionPoints()
	belowOneThird := fpCurrent < fpMax/3
	return belowOneThird
}

func (a *Actor) TextIcon(bg color.RGBA) textiles.TextIcon {
	return a.GetIcon().WithBg(bg)
}

func (a *Actor) GetDetailInfo() string {

	var result []string
	result = append(result, fmt.Sprintf("Name: %s", a.Name()))
	result = a.appendStateInfo(result)
	// melee attack
	statRows := []fxtools.TableRow{
		fxtools.TableRow{Columns: []string{"Str:", fmt.Sprintf("%d", a.CharSheet.GetStat(d100.Strength))}},
		fxtools.TableRow{Columns: []string{"Per:", fmt.Sprintf("%d", a.CharSheet.GetStat(d100.Perception))}},
		fxtools.TableRow{Columns: []string{"End:", fmt.Sprintf("%d", a.CharSheet.GetStat(d100.Endurance))}},
		fxtools.TableRow{Columns: []string{"Coo:", fmt.Sprintf("%d", a.CharSheet.GetStat(d100.Cool))}},
		fxtools.TableRow{Columns: []string{"Int:", fmt.Sprintf("%d", a.CharSheet.GetStat(d100.Intelligence))}},
		fxtools.TableRow{Columns: []string{"Agi:", fmt.Sprintf("%d", a.CharSheet.GetStat(d100.Agility))}},
	}

	derivedStatRows := []fxtools.TableRow{
		{Columns: []string{"HP:", fmt.Sprintf("%d/%d", a.GetHitPoints(), a.GetHitPointsMax())}},
		{Columns: []string{"AP:", fmt.Sprintf("%d/%d", a.CharSheet.GetActionPoints(), a.CharSheet.GetActionPointsMax())}},
		{Columns: []string{"Speed:", fmt.Sprintf("%d", a.CharSheet.GetDerivedStat(d100.Speed))}},
		{Columns: []string{"Dodge:", fmt.Sprintf("%d", a.CharSheet.GetDerivedStat(d100.Dodge))}},
		{Columns: []string{"Crit. Chance:", fmt.Sprintf("%d", a.CharSheet.GetDerivedStat(d100.CriticalChance))}},
		{Columns: []string{"Carry Weight:", fmt.Sprintf("%d", a.CharSheet.GetDerivedStat(d100.CarryWeight))}},
		{Columns: []string{"Max Repair:", fmt.Sprintf("%d%%", int(a.GetMaxRepairQuality()))}},
	}

	resistanceRows := []fxtools.TableRow{
		{Columns: []string{"Physical:", fmt.Sprintf("%d", a.CharSheet.GetDerivedStat(d100.DamageResistance))}},
		{Columns: []string{"Poison :", fmt.Sprintf("%d", a.CharSheet.GetDerivedStat(d100.PoisonResistance))}},
	}

	var skillRows []fxtools.TableRow
	for skillNo := 0; skillNo < d100.SkillCount(); skillNo++ {
		skill := d100.Skill(skillNo)
		skillRows = append(skillRows, fxtools.TableRow{Columns: []string{skill.String() + ":", fmt.Sprintf("%d", a.CharSheet.GetSkill(skill))}})
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
	if !a.GetInventory().HasArmorEquipped() {
		return 0
	}
	armorEncumbrance := a.GetInventory().GetEncumbranceFromArmor()
	return armorEncumbrance
}

func (a *Actor) GetSizeModifier() int {
	return a.SizeModifier
}

func (a *Actor) SetSizeModifier(modifier int) {
	a.SizeModifier = modifier
}
func (a *Actor) HasGold(price int) bool {
	return a.GetInventory().HasItemWithNameAndCount("gold", price)
}

func (a *Actor) RemoveGold(price int) foundation.Item {
	count := a.GetInventory().RemoveItemsByNameAndCount("gold", price)
	if len(count) == 0 {
		return nil
	}
	return count[0]
}

func (a *Actor) GetGold() int {
	gold := a.GetInventory().GetItemByName("gold")
	if gold == nil {
		return 0
	}
	return gold.GetStackSize()
}

func (a *Actor) IsWounded() bool {
	return a.GetHitPoints() < a.GetHitPointsMax()
}

func (a *Actor) SetSleeping() {
	flags := a.GetFlags()
	flags.Set(foundation.FlagSleep)
}

func (a *Actor) IsBlind() bool {
	return a.HasFlag(foundation.FlagBlind)
}

func (a *Actor) AddTimeEnergy(timeSpent int) {
	a.RawTimeEnergy += timeSpent
}

func (a *Actor) HasEnergyForActions() bool {
	return a.RawTimeEnergy >= a.maximalTimeNeededForActions()
}

func (a *Actor) maximalTimeNeededForActions() int {
	return max(a.TimeNeededForActions(), a.TimeNeededForMovement())
}

func (a *Actor) minimalTimeNeededForAction() int {
	return min(a.TimeNeededForActions(), a.TimeNeededForMovement())
}
func (a *Actor) timeNeededForMeleeAttack() int {
	if meleeWeapon, hasWeapon := a.GetInventory().GetEquippedWeapon(); hasWeapon {
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
	if a.HasFlag(foundation.FlagKnockedDown) {
		return speed / 2
	}
	speed = max(1, speed-a.GetEncumbrance())
	timeNeeded := 100 / speed
	return timeNeeded
}

func (a *Actor) SpendTimeEnergy(amount int) {
	a.RawTimeEnergy -= amount
}

func (a *Actor) AfterTurn() []foundation.HiLiteString {
	a.GetInventory().AfterTurn()
	a.decrementStatusEffectCounters()

	wornOff := a.decrementTemporaryStatChanges()

	if a.HasFlag(foundation.FlagRunning) {
		sheet := a.GetCharSheet()
		if sheet.GetActionPoints() > 0 {
			sheet.LooseActionPoints(1)
		} else {
			a.UnsetFlag(foundation.FlagRunning)
			wornOff = append(wornOff, foundation.HiLite("You are no longer running."))
		}
	}

	if a.IsCyberWareActive(CyberWareThermopticCamouflage) {
		sheet := a.GetCharSheet()
		if sheet.GetActionPoints() > 0 {
			sheet.LooseActionPoints(1)
		} else {
			a.SetCyberWareActive(CyberWareThermopticCamouflage, false)
			a.GetFlags().Unset(foundation.FlagActiveCamouflage)
			wornOff = append(wornOff, foundation.HiLite("Your thermoptic camouflage has been disabled."))
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

func (a *Actor) decrementTemporaryStatChanges() []foundation.HiLiteString {
	var wornOffEffects []foundation.HiLiteString
	for i := len(a.TemporaryStatChanges) - 1; i >= 0; i-- {
		statChange := a.TemporaryStatChanges[i]
		statChange.TurnsLeft--
		if statChange.TurnsLeft < 0 {
			wornOffEffects = append(wornOffEffects, foundation.HiLite("%s has worn off.", statChange.Name))
			a.TemporaryStatChanges = append(a.TemporaryStatChanges[:i], a.TemporaryStatChanges[i+1:]...)
		}
	}
	return wornOffEffects
}
func (a *Actor) GetBasicSpeed() int {
	return max(1, a.CharSheet.GetDerivedStat(d100.Speed))
}

func (a *Actor) GetCharSheet() *d100.CharSheet {
	return a.CharSheet
}

func (a *Actor) Kill() {
	a.CharSheet.Kill()
}

func (a *Actor) HasKey(identifier string) bool {
	return a.GetInventory().HasKey(identifier)
}

func (a *Actor) IsInCombat() bool {
	return a.HasActiveGoal() && a.ActiveGoal.IsCombatGoal()
}

func (a *Actor) SetDisplayName(name string) {
	a.DisplayName = name
}

func (a *Actor) GetDialogueFile() string {
	return a.DialogueFile
}

func (a *Actor) tryEquipWeapon() {
	if !a.GetInventory().HasWeaponEquipped() {
		weapon := a.GetInventory().GetBestWeapon()
		if weapon != nil {
			a.GetInventory().Equip(weapon)
		}
	}
}

func (a *Actor) tryEquipRangedWeapon() {
	if !a.GetInventory().HasRangedWeaponEquipped() {
		weapon := a.GetInventory().GetBestRangedWeapon()
		if weapon != nil {
			a.GetInventory().Equip(weapon)
		}
	}
}
func (a *Actor) tryEquipMeleeWeapon() {
	if !a.GetInventory().HasMeleeWeaponEquipped() {
		weapon := a.GetInventory().GetBestMeleeWeapon()
		if weapon != nil {
			a.GetInventory().Equip(weapon)
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
	return a.CharSheet.GetDerivedStat(d100.MeleeDamageBonus)
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
	return a.AudioBaseName
}

func (a *Actor) GetTeam() string {
	return a.TeamName
}

func (a *Actor) AddToEnemyActors(name string) {
	if a.InternalName == name {
		return
	}
	a.EnemyActors[name] = true
}

func (a *Actor) AddToEnemyTeams(name string) {
	if a.TeamName == name {
		return
	}
	a.EnemyTeams[name] = true
}

func (a *Actor) IsHostileTowards(attacker *Actor) bool {
	if a.HasActiveGoal() && a.ActiveGoal.IsHostilityTowards(attacker) {
		return true
	}
	if a.IsAggressive() && attacker.TeamName != a.TeamName {
		return true
	}
	if _, exists := a.EnemyActors[attacker.GetInternalName()]; exists {
		return true
	}
	if _, exists := a.EnemyTeams[attacker.GetTeam()]; exists {
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

	return displayName + "\n" + a.ActionDescription() + "\n" + a.OutfitDescription()
}

func (a *Actor) ActionDescription() string {
	action := "just standing there"
	if a.HasActiveGoal() {
		action = a.ActiveGoal.Description()
	}
	if a.Schedule != nil && a.Schedule.LastSlotID.Index != -1 {
		action = a.Schedule.CurrentTimeSlot().Description()
	}
	// Activity
	// Clothing
	// Equipped
	return fmt.Sprintf("Activity: [white]%s[-]", action)
}

func (a *Actor) OutfitDescription() string {
	armor := a.GetInventory().GetArmor()
	helmet := a.GetInventory().GetHelmet()
	clothes := ""
	hasClothes := true
	if armor == nil && helmet == nil {
		clothes = fmt.Sprintf("nothing")
		hasClothes = false
	} else if armor != nil && helmet != nil {
		clothes = fmt.Sprintf("%s and %s", armor.Name(), helmet.Name())
	} else if armor != nil && helmet == nil {
		clothes = fmt.Sprintf("%s", armor.Name())
	} else if helmet != nil && armor == nil {
		clothes = fmt.Sprintf("%s", helmet.Name())
	}
	outFitStyle := foundation.FashionStyleLowLife
	if hasClothes {
		outFitStyle = a.OutfitStyle()
	}

	clothes = fmt.Sprintf("%s (%s)", clothes, outFitStyle.String())

	hands := ""
	item, hasItem := a.GetInventory().GetMainHandItem()

	if hasItem {
		hands = fmt.Sprintf("%s", item.Name())
	} else {
		if a.IsOpenCarryWeapon() {
			return fmt.Sprintf("Clothing: [white]%s[-]\nCarrying: [white]weapons[-]", clothes)
		}
		return fmt.Sprintf("Clothing: [white]%s[-]", clothes)
	}

	return fmt.Sprintf("Clothing: [white]%s[-]\nEquipped: [white]%s[-]", clothes, hands)
}

func (a *Actor) HasDialogue() bool {
	return a.DialogueFile != ""
}

func (a *Actor) HasStealableItems() bool {
	return a.GetInventory().HasStealableItems(a.GetInventory().IsNotEquipped)
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
	delete(a.EnemyActors, other.GetInternalName())
}

func (a *Actor) SetIcon(icon textiles.TextIcon) {
	a.Icon = icon
}

func (a *Actor) SetCharSheet(character *d100.CharSheet) {
	a.CharSheet = character
}

func (a *Actor) SetXP(xp int) {
	a.XP = xp
}

func (a *Actor) ToRecord() recfile.Record {
	actorRecord := append(recfile.Record{
		recfile.Field{Name: "Name", Value: a.DisplayName},
		recfile.Field{Name: "GetInternalName", Value: a.InternalName},
		recfile.Field{Name: "GetIcon", Value: string(a.Icon.Char)},
		recfile.Field{Name: "Fg", Value: recfile.RGBStr(a.Icon.Fg)},
		recfile.Field{Name: "Bg", Value: recfile.RGBStr(a.Icon.Bg)},
		recfile.Field{Name: "DialogueFile", Value: a.DialogueFile},
		recfile.Field{Name: "Team", Value: a.TeamName},
		recfile.Field{Name: "XP", Value: recfile.IntStr(a.XP)},
	}, a.CharSheet.ToRecord()...)
	return actorRecord
}

func (a *Actor) ActOnGoal(g *GameState) (fsmai.TransitionEvent, int) {
	if a.ActiveGoal == nil {
		return fsmai.NoEvent, 0
	}
	if a.ActiveGoal.Achieved(g, a) {
		a.ActiveGoal = nil
		return fsmai.NoEvent, 0
	}
	event, tuSpent := a.ActiveGoal.Action(g, a)
	if a.ActiveGoal.Achieved(g, a) {
		a.ActiveGoal = nil
	}
	return event, tuSpent
}

func (a *Actor) HasActiveGoal() bool {
	return a.ActiveGoal != nil
}

func (a *Actor) GetMeleeTUCost() int {
	if meleeWeapon, hasWeapon := a.GetInventory().GetEquippedWeapon(); hasWeapon {
		return meleeWeapon.GetCurrentAttackMode().TUCost
	}
	return a.TimeNeededForActions()
}

func (a *Actor) SetGoal(goal ActorGoal) {
	a.ActiveGoal = goal
}

func (a *Actor) GetWeaponRange() int {
	if rangedWeapon, hasWeapon := a.GetInventory().GetRangedWeapon(); hasWeapon {
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

	if a.CurrentPathIndex < 0 || a.CurrentPathIndex >= len(a.CurrentPath) {
		return a.Position()
	}

	nextStep := a.CurrentPath[a.CurrentPathIndex]
	if !g.currentMap().IsWalkableFor(nextStep, a) {
		a.calcAndSetPath(g, pos)
		if a.CurrentPathIndex == -1 {
			return a.Position()
		}
		nextStep = a.CurrentPath[a.CurrentPathIndex]
	}
	a.CurrentPathBlockedCount = 0
	a.CurrentPathIndex++
	return nextStep

}

func (a *Actor) getMoveTowardsActor(g *GameState, other *Actor, maxDist int) geometry.Point {
	moveDist := g.currentMap().MoveDistance(a.Position(), other.Position())
	if moveDist <= maxDist {
		return a.Position()
	}

	nextStep := g.currentMap().GetMoveOnOtherDijkstraMap(a.Position(), true, other.DijkstraMap)

	if !g.currentMap().IsWalkableFor(nextStep, a) {
		a.CurrentPathBlockedCount++
		if a.CurrentPathBlockedCount <= 3 {
			return a.Position()
		}
	}
	a.CurrentPathBlockedCount = 0
	return nextStep
}

func (a *Actor) getMoveAwayFromActor(g *GameState, other *Actor) geometry.Point {
	nextStep := g.currentMap().GetMoveOnOtherDijkstraMap(a.Position(), false, other.DijkstraMap)

	if !g.currentMap().IsWalkableFor(nextStep, a) {
		a.CurrentPathBlockedCount++
		if a.CurrentPathBlockedCount <= 3 {
			return a.Position()
		}
	}
	a.CurrentPathBlockedCount = 0
	a.CurrentPathIndex++
	return nextStep
}

func (a *Actor) calcAndSetPath(g *GameState, pos geometry.Point) {
	a.CurrentPath = nil
	a.CurrentPathBlockedCount = 0
	calcPath := g.currentMap().GetJPSPath(a.Position(), pos, func(point geometry.Point) bool {
		return g.currentMap().IsWalkableFor(point, a)
	})
	if len(calcPath) == 0 || (len(calcPath) == 1 && calcPath[0] == a.Position()) {
		a.CurrentPathIndex = -1
	} else {
		a.CurrentPathIndex = 0
		a.CurrentPath = calcPath
	}
}
func (a *Actor) cannotFindPath() bool {
	return a.CurrentPathIndex == -1
}
func (a *Actor) hasPathTo(pos geometry.Point) bool {
	if a.CurrentPath == nil || len(a.CurrentPath) == 0 {
		return false
	}
	if a.CurrentPathIndex < 0 || a.CurrentPathIndex >= len(a.CurrentPath) {
		return false
	}
	targetOfPath := a.CurrentPath[len(a.CurrentPath)-1]
	isNear := geometry.DistanceChebyshev(targetOfPath, pos) <= 1
	return isNear
}

func (a *Actor) GetMaxThrowRange() int {
	strength := a.GetCharSheet().GetStat(d100.Strength)
	return strength * 2
}

func (a *Actor) GetXP() int {
	return a.XP
}

func (a *Actor) TryEquipRangedWeaponFirst() {
	a.tryEquipRangedWeapon()
	if !a.GetInventory().HasRangedWeaponEquipped() {
		a.tryEquipWeapon()
	}
}

func (a *Actor) IsOverEncumbered() bool {
	carryWeight := a.GetCharSheet().GetDerivedStat(d100.CarryWeight)
	totalWeight := a.GetInventory().GetTotalWeight()
	return totalWeight > carryWeight
}

func (a *Actor) SetStance(stance ActorStance) {
	a.Stance = stance
}

func (a *Actor) RemoveGoal() {
	a.ActiveGoal = nil
}

func (a *Actor) IsAlliedWith(player *Actor) bool {
	return a.TeamName == player.TeamName
}

func (a *Actor) SetChatterFile(value string) {
	a.ChatterFile = value
}

func (a *Actor) UnsetFlag(flag foundation.ActorFlag) {
	a.StatusFlags.Unset(flag)
}

func (a *Actor) SetFlag(flag foundation.ActorFlag) {
	a.StatusFlags.Set(flag)
}

func (a *Actor) GetTemporarySkillModifiers(skill d100.Skill) []d100.Modifier {
	var result []d100.Modifier
	for _, statChange := range a.TemporaryStatChanges {
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

	if skill == d100.SkillForSneak && a.HasFlag(foundation.FlagActiveCamouflage) {
		result = append(result, d100.DefaultModifier{
			Source:    "Active Camouflage",
			Modifier:  75,
			Order:     1,
			IsPercent: true,
		})
	}

	if skill == d100.SkillForSmoothTalking {
		debuff := 0
		if a.HasFlag(foundation.FlagHunger) {
			debuff = -10
		} else if a.HasFlag(foundation.FlagStarving) {
			debuff = -20
		}
		if debuff != 0 {
			result = append(result, d100.DefaultModifier{
				Source:    "Hunger",
				Modifier:  debuff,
				Order:     1,
				IsPercent: true,
			})
		}
	}

	return result
}
func (a *Actor) secondaryInit() {
	a.attachHooksToActor()

	a.GetInventory().SetName(fmt.Sprintf("%s's Inventory", a.Name()))
}

func (a *Actor) attachHooksToActor() {
	inventory := a.GetInventory()

	equipment := a.GetInventory()

	// Hookup inventory item owner position tracking
	inventory.SetCarrierPosition(a.Position)
	inventory.EnsurePositionHandlers()

	if a.VendorInv != nil {
		a.VendorInv.SetCarrierPosition(a.Position)
		a.VendorInv.EnsurePositionHandlers()
	}

	// Hookup automatic un-equip on removal
	inventory.SetOnBeforeRemove(equipment.UnEquip)

	// Hookup Modifiers for CharSheet
	a.GetCharSheet().SetSkillModifierHandler(func(skill d100.Skill) []d100.Modifier {
		modsFromItems := a.GetInventory().GetSkillModifiersFromItems(skill)
		modsFromEquipment := equipment.GetSkillModifiersFromEquippedItems(skill)
		modsFromActiveEffects := a.GetTemporarySkillModifiers(skill)
		return append(append(modsFromItems, modsFromEquipment...), modsFromActiveEffects...)
	})

	a.GetCharSheet().SetStatModifierHandler(func(stat d100.Stat) []d100.Modifier {
		modsFromItems := a.GetInventory().GetStatModifiersFromItems(stat)
		modsFromEquipment := equipment.GetStatModifiersFromEquippedItems(stat)
		modsFromActiveEffects := a.GetTemporaryStatModifiers(stat)
		return append(append(modsFromItems, modsFromEquipment...), modsFromActiveEffects...)
	})

	a.GetCharSheet().SetDerivedStatModifierHandler(func(stat d100.DerivedStat) []d100.Modifier {
		modsFromItems := a.GetInventory().GetDerivedStatModifiersFromItems(stat)
		modsFromEquipment := equipment.GetDerivedStatModifiersFromEquippedItems(stat)
		modsFromActiveEffects := a.GetTemporaryDerivedStatModifiers(stat)
		return append(append(modsFromItems, modsFromEquipment...), modsFromActiveEffects...)
	})
}
func (a *Actor) GetTemporaryStatModifiers(stat d100.Stat) []d100.Modifier {
	var result []d100.Modifier
	for _, statChange := range a.TemporaryStatChanges {
		if value, exists := statChange.StatChanges[stat]; exists {
			result = append(result, d100.DefaultModifier{
				Source:   statChange.Name,
				Modifier: value,
				Order:    1,
				Suffix:   fmt.Sprintf("(%d turns left)", statChange.TurnsLeft),
			})
		}
	}

	if stat == d100.Strength {
		debuff := 0
		if a.HasFlag(foundation.FlagHunger) {
			debuff = -1
		} else if a.HasFlag(foundation.FlagStarving) {
			debuff = -2
		}
		if debuff != 0 {
			result = append(result, d100.DefaultModifier{
				Source:    "Hunger",
				Modifier:  debuff,
				Order:     1,
				IsPercent: false,
			})
		}
	}
	return result
}

func (a *Actor) GetTemporaryDerivedStatModifiers(stat d100.DerivedStat) []d100.Modifier {
	var result []d100.Modifier
	for _, statChange := range a.TemporaryStatChanges {
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
	for i, existingChange := range a.TemporaryStatChanges {
		if existingChange.Name == change.Name {
			a.TemporaryStatChanges[i] = change
			return
		}
	}
	a.TemporaryStatChanges = append(a.TemporaryStatChanges, change)
}

func (a *Actor) GetVendorInventory() *Inventory {
	return a.VendorInv
}

func (a *Actor) GetMeleeSkillUsed() d100.Skill {
	if meleeWeapon, hasWeapon := a.GetInventory().GetEquippedWeapon(); hasWeapon {
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
	if a.Schedule == nil {
		return TimeSlot{}, false
	}
	return a.Schedule.MoveToNextTimeSlot(time)
}

func (a *Actor) appendStateInfo(result []string) []string {
	fsmState := a.FSM.currentBehavior.AssociatedState().ToString()
	goal := "none"
	if a.ActiveGoal != nil {
		goal = a.ActiveGoal.Description()
	}
	schedule := "none"
	if a.Schedule != nil {
		schedule = a.Schedule.String()
	}
	result = append(result, fmt.Sprintf("Schedule: %s", schedule))
	result = append(result, fmt.Sprintf("State: %s", fsmState))
	result = append(result, fmt.Sprintf("Goal: %s", goal))
	return result
}

func (a *Actor) IsHeavilyInjured() bool {
	return a.GetHitPoints() < a.GetHitPointsMax()/3
}

func (a *Actor) IsAggressive() bool {
	return a.Aggressive
}

func (a *Actor) SetNeutral() {
	a.Aggressive = false
}

func (a *Actor) SetAggressive() {
	a.Aggressive = true
}

func (a *Actor) HasPerk(perk d100.Perk) bool {
	return a.CharSheet.HasPerk(perk)
}

func (a *Actor) GetPerkLevel(perk d100.Perk) int {
	return a.CharSheet.GetPerkLevel(perk)
}

func (a *Actor) IsGuarding(zone string) bool {
	if a.GuardingZone == "" || zone == "" {
		return false
	}
	return a.GuardingZone == zone
}

func (a *Actor) CanDetect(pos geometry.Point) bool {
	return geometry.Distance(a.Position(), pos) <= float64(a.DetectionRange())
}

func (a *Actor) OutfitStyle() foundation.FashionStyle {
	armor := a.GetInventory().GetArmor()
	helmet := a.GetInventory().GetHelmet()
	if armor == nil { // naked == low life
		return foundation.FashionStyleLowLife
	}
	if helmet == nil {
		return armor.FashionStyle
	}

	if armor.FashionStyle == helmet.FashionStyle {
		return armor.FashionStyle
	}

	return min(armor.FashionStyle, helmet.FashionStyle)
}

func (a *Actor) GetArmorString() string {
	if !a.GetInventory().HasArmorEquipped() {
		return fmt.Sprintf("no clothes (low-life)")
	}
	armor := a.GetInventory().GetArmor()
	return fmt.Sprintf("%s (%s)", armor.InventoryName(), a.OutfitStyle().String())
}

func (a *Actor) HasActionPoints() bool {
	return a.GetCharSheet().GetActionPoints() > 0
}

func (a *Actor) IsIdle() bool {
	idleState := a.FSM.State() == fsmai.StateNeutral || a.FSM.State() == fsmai.StateAggressive
	noGoal := !a.HasActiveGoal()
	return idleState && noGoal
}

func (a *Actor) HasWatch() bool {
	if a.HasCyberWare(CyberWareClock) {
		return true
	}
	return a.GetInventory().HasWatch()
}

func (a *Actor) InitWithGameState(g *GameState) {
	a.secondaryInit()
	// FSM
	defaultState := fsmai.StateNeutral
	if a.Aggressive {
		defaultState = fsmai.StateAggressive
	}
	a.FSM.RestoreState(g, a, defaultState, DefaultBehaviorFactory)
}

func (a *Actor) ID() gridmap.ActorID {
	return a.UID
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

func IsActor(target *Actor) func(actor *Actor) bool {
	return func(actor *Actor) bool { return actor == target }
}
