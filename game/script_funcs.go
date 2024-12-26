package game

import (
    "contractor/d100"
    "contractor/foundation"
    "contractor/fsmai"
    "github.com/Knetic/govaluate"
    "github.com/memmaker/go/geometry"
    "strconv"
    "strings"
    "time"
)

func (g *GameState) GetScriptFuncs() map[string]govaluate.ExpressionFunction {
    return map[string]govaluate.ExpressionFunction{
        // Player Only
        "IsWounded": func(args ...interface{}) (interface{}, error) {
            return g.Player.IsWounded(), nil
        },
        "Skill": func(args ...interface{}) (interface{}, error) {
            skillName := args[0].(string)
            skillValue := g.Player.GetCharSheet().GetSkill(d100.SkillFromString(skillName))
            return (float64)(skillValue), nil
        },
        "RollSkill": func(args ...interface{}) (interface{}, error) {
            skillName := args[0].(string)
            diff := d100.Medium
            if len(args) > 1 {
                diff = d100.DifficultyFromString(args[1].(string))
            }
            result := g.Player.GetCharSheet().SkillRollVsDiff(d100.SkillFromString(skillName), diff)
            return (bool)(result.Success), nil
        },
        // o_test: DialogueCheck(NPC, 'intimidate')
        "DialogueCheck": func(args ...interface{}) (interface{}, error) {
            opponent := args[0].(*Actor)
            skillName := args[0].(string)
            skill := d100.SkillFromString(skillName)
            skillValue := g.Player.GetCharSheet().GetSkill(skill)
            mods := g.getDialogueCheckMods(g.Player, opponent, skill, nil)
            chance := mods.Apply(skillValue)
            result := d100.SuccessRoll(d100.Percentage(chance), 0)
            return (bool)(result.Success), nil
        },

        // Player Inventory & Equipment
        "RemoveItem": func(args ...interface{}) (interface{}, error) {
            count := 1
            itemName := args[0].(string)
            if len(args) > 1 {
                count = int(args[1].(float64))
            }
            removedItem := g.Player.GetInventory().RemoveItemsByNameAndCount(itemName, count)
            if len(removedItem) > 0 {
                g.msg(foundation.HiLite("%s removed.", removedItem[0].Name()))
            }
            return nil, nil
        },
        // eg. StackTransForTo(NPC, 'gold', 500)
        "StackTransferTo": func(args ...interface{}) (interface{}, error) {
            count := 1
            targetActor := args[0].(*Actor)
            itemName := args[1].(string)
            if len(args) > 2 {
                count = int(args[2].(float64))
            }
            removedItem := g.Player.GetInventory().RemoveItemsByNameAndCount(itemName, count)
            if len(removedItem) > 0 {
                targetActor.GetInventory().AddItems(removedItem)
                g.msg(foundation.HiLite("%s removed.", removedItem[0].Name()))
            }
            return nil, nil
        },

        "HasItem": func(args ...interface{}) (interface{}, error) {
            itemName := args[0].(string)
            count := 1
            if len(args) > 1 {
                count = int(args[1].(float64))
            }
            return g.Player.GetInventory().HasItemWithNameAndCount(itemName, count), nil
        },
        "HasArmorEquipped": func(args ...interface{}) (interface{}, error) {
            return g.Player.GetInventory().HasArmorEquipped(), nil
        },
        "HasVisibleWeapon": func(args ...interface{}) (interface{}, error) {
            return g.Player.IsOpenCarryWeapon(), nil
        },
        "HasArmorEquippedWithName": func(args ...interface{}) (interface{}, error) {
            armorName := args[0].(string)
            return g.Player.GetInventory().HasArmorWithNameEquipped(armorName), nil
        },
        "HasWeaponEquippedWithName": func(args ...interface{}) (interface{}, error) {
            weaponName := args[0].(string)
            mainHandItem, hasMainHandItem := g.Player.GetInventory().GetMainHandItem()
            if !hasMainHandItem {
                return false, nil
            }
            return mainHandItem.GetInternalName() == weaponName, nil
        },

        // Global Queries & Actions
        "UserFunc": func(args ...interface{}) (interface{}, error) {
            funcName := args[0].(string)
            expression, exists := g.userFunctions[funcName]
            if !exists {
                return nil, nil
            }
            return expression.Evaluate(nil)
        },
        // Flags
        "HasFlag": func(args ...interface{}) (interface{}, error) {
            flagName := args[0].(string)
            return (bool)(g.gameFlags.HasFlag(flagName)), nil
        },
        "SetFlag": func(args ...interface{}) (interface{}, error) {
            flagName := args[0].(string)
            g.gameFlags.SetFlag(flagName)
            return nil, nil
        },
        "ClearFlag": func(args ...interface{}) (interface{}, error) {
            flagName := args[0].(string)
            g.gameFlags.ClearFlag(flagName)
            return nil, nil
        },
        "GetFlag": func(args ...interface{}) (interface{}, error) {
            flagName := args[0].(string)
            value := g.gameFlags.Get(flagName)
            return (float64)(value), nil
        },
        "IsMap": func(args ...interface{}) (interface{}, error) {
            mapName := args[0].(string)
            return g.currentMap().GetName() == mapName, nil
        },
        "RandomNumberCode": func(args ...interface{}) (interface{}, error) {
            nameOfDoor := args[0].(string)
            code := g.getRandomNumberCode(nameOfDoor)
            return string(code), nil
        },
        // Time / Turns
        "Turns": func(args ...interface{}) (interface{}, error) {
            return (float64)(g.TurnsTaken()), nil
        },
        "IsTurnsAfter": func(args ...interface{}) (interface{}, error) {
            namedTime := args[0].(string)
            turns := args[1].(float64)
            return g.IsTurnsAfter(namedTime, int(turns)), nil
        },
        "IsMinutesAfter": func(args ...interface{}) (interface{}, error) {
            namedTime := args[0].(string)
            minutes := args[1].(float64)
            return g.IsMinutesAfter(namedTime, int(minutes)), nil
        },
        "IsHoursAfter": func(args ...interface{}) (interface{}, error) {
            namedTime := args[0].(string)
            hours := args[1].(float64)
            return g.IsHoursAfter(namedTime, int(hours)), nil
        },
        "IsDaysAfter": func(args ...interface{}) (interface{}, error) {
            namedTime := args[0].(string)
            days := args[1].(float64)
            return g.IsDaysAfter(namedTime, int(days)), nil
        },

        // Scripts
        "IsScriptRunning": func(args ...interface{}) (interface{}, error) {
            scriptName := args[0].(string)
            return g.scriptRunner.IsScriptRunning(scriptName), nil
        },
        "RunScript": func(args ...interface{}) (interface{}, error) {
            scriptName := args[0].(string)
            g.RunScriptByName(scriptName)
            return nil, nil
        },
        "RunScriptOnMap": func(args ...interface{}) (interface{}, error) {
            mapName := args[0].(string)
            scriptName := args[1].(string)
            g.RunScriptOnMap(mapName, scriptName)
            return nil, nil
        },
        "StopScript": func(args ...interface{}) (interface{}, error) {
            scriptName := args[0].(string)
            g.scriptRunner.StopScript(g.currentMap().GetName(), scriptName)
            return nil, nil
        },
        "RestartScript": func(args ...interface{}) (interface{}, error) {
            scriptName := args[0].(string)
            g.scriptRunner.StopScript(g.currentMap().GetName(), scriptName)
            g.RunScriptByName(scriptName)
            return nil, nil
        },
        "RunScriptKill": func(args ...interface{}) (interface{}, error) {
            killerName := args[0].(string)
            victimName := args[1].(string)
            killer := g.actorWithName(killerName)
            victim := g.actorWithName(victimName)
            g.msg(foundation.HiLite("%s kills %s.", killerName, victimName))
            killScript := g.NewScriptKill(killer, victim)
            g.RunScript(killScript)
            return nil, nil
        },
        "RunScriptLeave": func(args ...interface{}) (interface{}, error) {
            argZero := args[0]
            var leaver *Actor
            if nameOfLeaver, isString := argZero.(string); isString {
                leaver = g.actorWithName(nameOfLeaver)
            } else if actor, isActor := argZero.(*Actor); isActor {
                leaver = actor
            }

            isRunning := false
            if len(args) > 1 {
                isRunning = args[1].(bool)
            }

            killScript := g.NewScriptLeaveMap(leaver, isRunning)
            g.RunScript(killScript)
            return nil, nil
        },
        "RunScriptLeaveAt": func(args ...interface{}) (interface{}, error) {
            argZero := args[0]
            var leaver *Actor
            if nameOfLeaver, isString := argZero.(string); isString {
                leaver = g.actorWithName(nameOfLeaver)
            } else if actor, isActor := argZero.(*Actor); isActor {
                leaver = actor
            }

            transitionLocation := args[1].(string)
            isRunning := false
            if len(args) > 2 {
                isRunning = args[2].(bool)
            }

            killScript := g.NewScriptLeaveMapAtLocation(leaver, isRunning, transitionLocation)
            g.RunScript(killScript)
            return nil, nil
        },

        // Query Containers
        "ContainerWithName": func(args ...interface{}) (interface{}, error) {
            containerName := args[0].(string)
            containers := g.currentMap().GetFilteredObjects(func(c Object) bool {
                return c.GetInternalName() == containerName
            })
            if len(containers) > 0 {
                return containers[0], nil
            }
            return nil, nil
        },

        "IsItemInContainer": func(args ...interface{}) (interface{}, error) {
            container := args[0].(*Container)
            count := 1
            itemName := args[1].(string)
            if len(args) > 2 {
                count = int(args[2].(float64))
            }
            return container.HasItemsWithName(itemName, count), nil
        },

        // Query Actors
        "PlayerInZone": func(args ...interface{}) (interface{}, error) {
            zoneName := args[0].(string)
            return g.currentMap().IsZoneAt(g.Player.Position(), zoneName), nil
        },
        "ActorWithName": func(args ...interface{}) (interface{}, error) {
            actorName := args[0].(string)
            actors := g.currentMap().GetFilteredActors(func(a *Actor) bool {
                return a.GetInternalName() == actorName
            })
            if len(actors) > 0 {
                return actors[0], nil
            }
            return nil, nil
        },
        "HasActorFlag": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            flagName := args[1].(string)
            return actor.HasFlag(foundation.ActorFlagFromString(flagName)), nil
        },
        "ItemCountZoneRecursiveSinglePrefix": func(args ...interface{}) (interface{}, error) {
            itemPrefix := args[0].(string)
            mapName := args[1].(string)
            zoneName := args[2].(string)
            count := 0
            g.IterateItemsOnMapZoneRecursively(mapName, zoneName, func(item foundation.Item) {
                if item.GetStackSize() == 1 &&
                    strings.HasPrefix(item.GetInternalName(), itemPrefix) {
                    count++
                }
            })
            return (float64)(count), nil
        },
        "ItemCountFactionInventorySinglePrefix": func(args ...interface{}) (interface{}, error) {
            itemPrefix := args[0].(string)
            mapName := args[1].(string)
            factionName := args[2].(string)
            count := 0
            g.IterateItemsInAllInventories(mapName, func(owner *Actor, item foundation.Item) {
                if owner.GetTeam() == factionName &&
                    item.GetStackSize() == 1 &&
                    strings.HasPrefix(item.GetInternalName(), itemPrefix) {
                    count++
                }
            })
            return (float64)(count), nil
        },
        "IsActorWounded": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            return actor.IsWounded(), nil
        },
        "IsActorInShootingRange": func(args ...interface{}) (interface{}, error) {
            attacker := args[0].(*Actor)
            defender := args[1].(*Actor)
            return g.IsInShootingRange(attacker, defender), nil
        },
        "IsActorAtNamedLocation": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            locName := args[1].(string)
            loc := g.currentMap().GetNamedLocation(locName)
            return actor.Position() == loc, nil
        },
        "IsActorAbleToReach": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            locName := args[1].(string)
            loc := g.currentMap().GetNamedLocation(locName)
            if loc == actor.Position() {
                return true, nil
            }
            pathToDest := g.currentMap().GetJPSPath(actor.Position(), loc, func(point geometry.Point) bool {
                return g.currentMap().IsWalkableFor(point, actor)
            })
            if len(pathToDest) == 0 {
                return false, nil
            }

            lastPos := pathToDest[len(pathToDest)-1]
            return lastPos == loc, nil
        },
        "IsActorDead": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            return !actor.IsAlive(), nil
        },
        "IsActorInCombat": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            if actor.IsInCombat() {
                return true, nil
            }
            return false, nil
        },
        "IsActorInTalkingRange": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            target := args[1].(*Actor)
            return g.IsInTalkingRange(actor, target), nil
        },
        "IsActorInCombatWithPlayer": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            if actor.IsHostileTowards(g.Player) {
                return true, nil
            }
            return false, nil
        },
        // Actor Actions,
        "ActorDropItem": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            count := 1
            itemName := args[1].(string)
            if len(args) > 2 {
                count = int(args[2].(float64))
            }
            removedItems := actor.GetInventory().RemoveItemsByNameAndCount(itemName, count)
            for _, item := range removedItems {
                g.addItemToMap(item, actor.Position())
            }
            if len(removedItems) > 0 {
                first := removedItems[0]
                displayName := first.Name()
                g.msg(foundation.HiLite("%s dropped %s.", actor.Name(), displayName))
            }
            return nil, nil
        },
        "ActorTransition": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            currentMap := g.currentMap()
            transition, exists := currentMap.GetTransitionAt(actor.Position())
            if !exists {
                return nil, nil
            }
            g.actorTransition(currentMap, actor, transition)
            return nil, nil
        },
        "PlayerAddCyberware": func(args ...interface{}) (interface{}, error) {
            cyberwareName := args[0].(string)
            g.playerAddCyberware(NewCyberWareFromString(cyberwareName))
            return nil, nil
        },
        // Global Actions
        "SaveTimeNow": func(args ...interface{}) (interface{}, error) {
            nameForTime := args[0].(string)
            g.SaveTimeNow(nameForTime)
            return nil, nil
        },
        "AdvanceTimeByMinutes": func(args ...interface{}) (interface{}, error) {
            minutes := int(args[0].(float64))
            g.advanceTime(time.Minute * time.Duration(minutes))
            return nil, nil
        },
        "AddChatter": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            chatter := args[1].(string)
            g.tryAddChatter(actor, chatter)
            return nil, nil
        },
        "Hilite": func(args ...interface{}) (interface{}, error) {
            text := args[0].(string)
            g.msg(foundation.HiLite(text))
            return nil, nil
        },

        // Actors Goals
        "MoveTo": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            locName := args[1].(string)
            loc := g.currentMap().GetNamedLocation(locName)
            actor.FSM.SetState(fsmai.StateScripted, LocationEvent{Event: fsmai.EventNone, Location: loc})
            return nil, nil
        },

        // Container Actions
        "ContainerRemoveItem": func(args ...interface{}) (interface{}, error) {
            container := args[0].(*Container)
            itemName := args[1].(string)
            count := 1
            if len(args) > 2 {
                count = int(args[2].(float64))
            }
            return container.RemoveItemsWithName(itemName, count), nil
        },
        "ContainerAddItem": func(args ...interface{}) (interface{}, error) {
            container := args[0].(*Container)
            if item, isItem := args[1].(foundation.Item); isItem {
                container.AddItem(item)
                return nil, nil
            } else if items, isItems := args[1].([]foundation.Item); isItems {
                container.AddItems(items)
                return nil, nil
            }
            newItem := g.NewItemFromString(args[1].(string))
            container.AddItem(newItem)
            return nil, nil
        },
        "ActorRemoveItem": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            count := 1
            itemName := args[1].(string)
            if len(args) > 2 {
                count = int(args[2].(float64))
            }
            removedItems := actor.GetInventory().RemoveItemsByNameAndCount(itemName, count)
            return removedItems, nil
        },
        "ActorAddItem": func(args ...interface{}) (interface{}, error) {
            actor := args[0].(*Actor)
            if item, isItem := args[1].(foundation.Item); isItem {
                actor.GetInventory().AddItem(item)
                return nil, nil
            } else if items, isItems := args[1].([]foundation.Item); isItems {
                actor.GetInventory().AddItems(items)
                return nil, nil
            }
            newItem := g.NewItemFromString(args[1].(string))
            actor.GetInventory().AddItem(newItem)
            return nil, nil
        },
        "PlayerAddItem": func(args ...interface{}) (interface{}, error) {
            newItem := g.NewItemFromString(args[1].(string))
            g.Player.GetInventory().AddItem(newItem)
            g.msg(foundation.HiLite("%s received.", newItem.Name()))
            return nil, nil
        },
        "PlayerAddGold": func(args ...interface{}) (interface{}, error) {
            goldAmount := int(args[0].(float64))
            g.Player.GetInventory().AddItem(g.NewGold(goldAmount))
            g.msg(foundation.HiLite("%s sat received.", strconv.Itoa(goldAmount)))
            return nil, nil
        },
    }
}
