package game

import (
	"contractor/foundation"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"regexp"
	"strconv"
	"strings"
)

type ToggleState struct {
	Label       string
	StateChange *govaluate.EvaluableExpression
	Icon        rune
}

type StateChanger struct {
	*BaseObject
	onBump       func(actor *Actor)
	onDamage     func(dmg SourcedDamage) []foundation.Animation
	canBeChanged bool
	states       []ToggleState
	currentState int
}

func (g *GameState) NewStateChanger(record recfile.Record, resolver func(objType string) textiles.TextIcon) *StateChanger {
	box := &StateChanger{
		BaseObject: &BaseObject{
			Category:      foundation.ObjectStateChanger,
			UseCustomIcon: true,
			InternalName:  "state_changer",
			iconForObject: resolver,
		},
		canBeChanged: true,
	}
	// state_0_label: "State 1"
	// state_0_change: GiveItem("item1")
	// state_1_label: "State 2"
	// state_1_change: GiveItem("item2")
	labelRegex := regexp.MustCompile(`^state_(\d+)_label$`)
	changeRegex := regexp.MustCompile(`^state_(\d+)_change$`)
	runeRegex := regexp.MustCompile(`^state_(\d+)_rune$`)
	for _, field := range record {
		lower := strings.ToLower(field.Name)
		switch lower {
		case "category":
			box.Category = foundation.ObjectCategoryFromString(field.Value)
			box.CustomIcon = box.iconForObject(field.Value)
		case "position":
			box.RawPosition, _ = geometry.NewPointFromEncodedString(field.Value)
		case "description":
			box.DisplayName = field.Value
		default:
			if labelRegex.MatchString(lower) {
				matches := labelRegex.FindStringSubmatch(lower)
				index, _ := strconv.Atoi(matches[1])
				if index >= len(box.states) {
					box.states = append(box.states, ToggleState{})
				}
				box.states[index].Label = field.Value
			} else if changeRegex.MatchString(lower) {
				matches := changeRegex.FindStringSubmatch(lower)
				index, _ := strconv.Atoi(matches[1])
				if index >= len(box.states) {
					box.states = append(box.states, ToggleState{})
				}
				expression, parseErr := govaluate.NewEvaluableExpressionWithFunctions(field.Value, g.GetScriptFuncs())
				if parseErr != nil {
					panic(parseErr)
				}
				box.states[index].StateChange = expression
			} else if runeRegex.MatchString(lower) {
				matches := runeRegex.FindStringSubmatch(lower)
				index, _ := strconv.Atoi(matches[1])
				if index >= len(box.states) {
					box.states = append(box.states, ToggleState{})
				}
				box.states[index].Icon = []rune(field.Value)[0]
			}
		}
	}

	box.SetWalkable(false)
	box.SetHidden(false)
	box.SetTransparent(false)

	box.InitWithGameState(g)
	return box
}
func (b *StateChanger) GetIcon() textiles.TextIcon {
	if len(b.states) == 0 {
		return b.BaseObject.GetIcon()
	}
	if len(b.states) == 1 {
		if b.canBeChanged {
			return b.BaseObject.GetIcon()
		} else {
			return b.BaseObject.GetIcon().WithRune(b.states[0].Icon)
		}
	}
	return b.BaseObject.GetIcon().WithRune(b.states[b.currentState].Icon)
}
func (b *StateChanger) InitWithGameState(g *GameState) {
	if len(b.states) > 0 {
		b.DisplayName = b.states[b.currentState].Label
	}
	b.onBump = func(actor *Actor) {
		if !b.canBeChanged {
			return
		}
		if len(b.states) == 0 {
			return
		}
		if len(b.states) == 1 {
			b.states[0].StateChange.Evaluate(nil)
			b.DisplayName = b.states[0].Label
			b.canBeChanged = false
			return
		}
		if len(b.states) == 2 {
			newState := 1 - b.currentState
			b.states[newState].StateChange.Evaluate(nil)
			b.DisplayName = b.states[newState].Label
			b.currentState = newState
			return
		}
		if len(b.states) > 2 {
			var menuItems []foundation.MenuItem
			for i, state := range b.states {
				menuItems = append(menuItems, foundation.MenuItem{
					Name: state.Label,
					Action: func() {
						state.StateChange.Evaluate(nil)
						b.DisplayName = state.Label
						b.currentState = i
					},
					CloseMenus: true,
				})
			}
			g.ui.OpenMenu(menuItems)
			return
		}
	}
	b.onDamage = func(dmg SourcedDamage) []foundation.Animation {
		return nil
	}
}
func (b *StateChanger) ToRecord() recfile.Record {
	return recfile.Record{
		{Name: "category", Value: b.Category.String()},
		{Name: "description", Value: b.DisplayName},
		{Name: "position", Value: b.RawPosition.Encode()},
	}
}

func (b *StateChanger) OnBump(actor *Actor) {
	b.onBump(actor)
}

func (b *StateChanger) OnDamage(dmg SourcedDamage) []foundation.Animation {
	return b.onDamage(dmg)
}
