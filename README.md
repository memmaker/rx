# Millenium RPG

### Player Skills

	small_guns
	big_guns
	energy_weapons
	unarmed
	melee_weapons
	throwing
	doctor
	sneak
	lockpick
	steal
	traps
	science
	repair
	speech
	barter
	gambling
	outdoorsman

### Dialogue File Format Overview

A dialogue file defines conversations through sections called `OpeningBranch` and `Nodes`.

#### **OpeningBranch Section**
- **cond**: A condition expression to determine if this branch is selected as the starting point.
- **goto**: The name of the node to go to if the condition is met.

#### **Nodes Section**
- **name**: Identifier for the conversation node.
- **npc**: Text spoken by the NPC at this node.
- **effect**: Action triggered at this node (e.g., `EndConversation`).
- **o_text**: Text for a player's dialogue option.
- **o_cond**: Condition to display this option.
- **o_goto**: Node to go to if the option is chosen.
- **o_test**: Skill check condition to determine the success of an option.
- **o_succ**: Node to go to on test success.
- **o_fail**: Node to go to on test failure.

#### **Conversation Flow**
1. **OpeningBranch**: Determines starting node based on conditions.
2. **Nodes**: NPC speaks, player chooses dialogue, tests may occur.
3. **Effects**: Triggered actions or state changes.

This format supports branching dialogues that adapt to the player’s decisions.

#### Dialogue Effects Summary

1. **EndHostility**: Sets the NPC to neutral toward the player.
2. **EndConversation**: Ends the current dialogue.
3. **Transition(mapName, locationName)**: Moves the player to a specified map and location.
4. **RemoveItem(itemName)**: Removes an item from the player's inventory.
5. **Hacking(terminalID, difficulty, flagName, followUpNode)**: Starts a hacking mini-game; success sets a flag and continues the dialogue.
6. **SetFlag(flagName)**: Sets a specified game flag to `true`.
7. **ClearFlag(flagName)**: Clears (sets to `false`) a specified game flag.

#### Supported Conditional Functions

1. **HasFlag(flagName)**: Checks if a specific game flag is set to `true`.
2. **HasItem(itemName)**: Checks if the player has an item by the given name in their inventory.
3. **GetSkill(skillName)**: Retrieves the player's skill level for the specified skill.
4. **RollSkill(skillName, modifier)**: Performs a skill check with an optional modifier; returns `true` on success.
5. **IsMap(mapName)**: Checks if the player is currently on the specified map.
6. **IsInCombat(npcName)**: Checks if the specified NPC is currently engaged in combat.
7. **IsInCombatWithPlayer(npcName)**: Checks if the specified NPC is hostile toward the player.
