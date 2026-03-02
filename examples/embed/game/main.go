// Game Simulation Example
// Demonstrates embedding Logos for game scripting with custom game objects

package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/codetesla51/logos/logos"
)

// Game state
type Player struct {
	Name   string
	Health int
	Mana   int
	Level  int
	XP     int
}

type Enemy struct {
	Name   string
	Health int
	Damage int
}

var player Player
var enemies []Enemy

func main() {
	rand.Seed(time.Now().UnixNano())

	// Initialize game state
	player = Player{
		Name:   "Hero",
		Health: 100,
		Mana:   50,
		Level:  1,
		XP:     0,
	}

	enemies = []Enemy{
		{Name: "Goblin", Health: 30, Damage: 5},
		{Name: "Orc", Health: 50, Damage: 10},
		{Name: "Dragon", Health: 100, Damage: 25},
	}

	// Create sandboxed VM for game scripts
	vm := logos.NewWithConfig(logos.SandboxConfig{
		AllowFileIO:  false,
		AllowNetwork: false,
		AllowShell:   false,
		AllowExit:    false,
	})

	// Register game functions
	vm.Register("getPlayer", func(args ...logos.Object) logos.Object {
		return &logos.Table{Pairs: map[string]logos.Object{
			"STRING:name":   &logos.String{Value: player.Name},
			"STRING:health": &logos.Integer{Value: int64(player.Health)},
			"STRING:mana":   &logos.Integer{Value: int64(player.Mana)},
			"STRING:level":  &logos.Integer{Value: int64(player.Level)},
			"STRING:xp":     &logos.Integer{Value: int64(player.XP)},
		}}
	})

	vm.Register("setPlayerName", func(args ...logos.Object) logos.Object {
		if len(args) < 1 {
			return &logos.String{Value: "error: name required"}
		}
		name := args[0].(*logos.String).Value
		player.Name = name
		return &logos.String{Value: "Player renamed to " + name}
	})

	vm.Register("heal", func(args ...logos.Object) logos.Object {
		amount := int64(20)
		if len(args) > 0 {
			amount = args[0].(*logos.Integer).Value
		}
		if player.Mana < 10 {
			return &logos.String{Value: "Not enough mana!"}
		}
		player.Mana -= 10
		player.Health += int(amount)
		if player.Health > 100 {
			player.Health = 100
		}
		return &logos.String{Value: fmt.Sprintf("Healed %d HP. Health: %d, Mana: %d", amount, player.Health, player.Mana)}
	})

	vm.Register("attack", func(args ...logos.Object) logos.Object {
		if len(args) < 1 {
			return &logos.String{Value: "error: enemy index required"}
		}
		idx := int(args[0].(*logos.Integer).Value)
		if idx < 0 || idx >= len(enemies) {
			return &logos.String{Value: "Invalid enemy index"}
		}

		enemy := &enemies[idx]
		damage := 10 + rand.Intn(10)
		enemy.Health -= damage

		result := fmt.Sprintf("You hit %s for %d damage!", enemy.Name, damage)
		if enemy.Health <= 0 {
			xpGain := 20 + rand.Intn(10)
			player.XP += xpGain
			result += fmt.Sprintf(" %s defeated! +%d XP", enemy.Name, xpGain)

			if player.XP >= player.Level*50 {
				player.Level++
				player.Health = 100
				player.Mana = 50
				result += fmt.Sprintf(" LEVEL UP! Now level %d!", player.Level)
			}
		} else {
			// Enemy counterattack
			player.Health -= enemy.Damage
			result += fmt.Sprintf(" %s hits back for %d damage!", enemy.Name, enemy.Damage)
		}

		return &logos.String{Value: result}
	})

	vm.Register("getEnemies", func(args ...logos.Object) logos.Object {
		elements := make([]logos.Object, len(enemies))
		for i, e := range enemies {
			elements[i] = &logos.Table{Pairs: map[string]logos.Object{
				"STRING:name":   &logos.String{Value: e.Name},
				"STRING:health": &logos.Integer{Value: int64(e.Health)},
				"STRING:damage": &logos.Integer{Value: int64(e.Damage)},
			}}
		}
		return &logos.Array{Elements: elements}
	})

	vm.Register("spawnEnemy", func(args ...logos.Object) logos.Object {
		if len(args) < 3 {
			return &logos.String{Value: "error: name, health, damage required"}
		}
		name := args[0].(*logos.String).Value
		health := int(args[1].(*logos.Integer).Value)
		damage := int(args[2].(*logos.Integer).Value)

		enemies = append(enemies, Enemy{Name: name, Health: health, Damage: damage})
		return &logos.String{Value: fmt.Sprintf("Spawned %s (HP: %d, DMG: %d)", name, health, damage)}
	})

	// Game script
	script := `
// Game simulation script
print("=== Welcome to the Arena ===")
print("")

// Set player name
setPlayerName("Warrior")

// Show player stats
let p = getPlayer()
print("Player: " + p["name"])
print("Health: " + toStr(p["health"]) + " | Mana: " + toStr(p["mana"]) + " | Level: " + toStr(p["level"]))
print("")

// List enemies
print("Enemies in the arena:")
let enemyList = getEnemies()
let i = 0
for enemy in enemyList {
    print("  [" + toStr(i) + "] " + enemy["name"] + " (HP: " + toStr(enemy["health"]) + ")")
    i = i + 1
}
print("")

// Battle sequence
print("--- Battle Begins! ---")
print("")

// Attack the goblin
print(attack(0))
print(attack(0))
print(attack(0))
print(attack(0))
print("")

// Heal up
print(heal(30))
print("")

// Spawn a new enemy
print(spawnEnemy("Troll", 60, 15))
print("")

// Final stats
let finalP = getPlayer()
print("--- Final Stats ---")
print("Level: " + toStr(finalP["level"]) + " | XP: " + toStr(finalP["xp"]))
print("Health: " + toStr(finalP["health"]) + " | Mana: " + toStr(finalP["mana"]))
`

	fmt.Println("Running game script...")
	fmt.Println("")

	err := vm.Run(script)
	if err != nil {
		fmt.Println("Script error:", err)
	}
}
