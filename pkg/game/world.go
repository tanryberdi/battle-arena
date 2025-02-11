package game

import (
	"math/rand"
	"sync"
)

// World represents the game world
type World struct {
	Characters map[string]*Character
	mu         sync.RWMutex
}

// NewWorld creates a new game world
func NewWorld() *World {
	return &World{
		Characters: make(map[string]*Character),
	}
}

// AddCharacter adds a character to the world
func (w *World) AddCharacter(c *Character) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Characters[c.ID] = c
}

// RemoveCharacter removes a character from the world
func (w *World) RemoveCharacter(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.Characters, id)
}

// GetCharacter gets a character by ID
func (w *World) GetCharacter(id string) *Character {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.Characters[id]
}

// GetCharacters returns a copy of all characters
func (w *World) GetCharacters() map[string]*Character {
	w.mu.RLock()
	defer w.mu.RUnlock()

	chars := make(map[string]*Character)
	for id, char := range w.Characters {
		chars[id] = char
	}
	return chars
}

// GetRandomSpawnPosition returns a random position in the world
func (w *World) GetRandomSpawnPosition() Position {
	return Position{
		X: rand.Float64() * WorldWidth,
		Y: rand.Float64() * WorldHeight,
	}
}

// ProcessCombat handles combat between characters
func (w *World) ProcessCombat() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, attacker := range w.Characters {
		for _, target := range w.Characters {
			if attacker.ID == target.ID {
				continue
			}

			distance := attacker.Position.Distance(target.Position)

			var attackRange float64
			switch attacker.Class {
			case Warrior:
				attackRange = WarriorAttackRange
			case Mage:
				attackRange = MageAttackRange
			}

			if distance <= attackRange {
				target.TakeDamage(attacker.AttackPower)

				if target.Health <= 0 {
					target.Position = w.GetRandomSpawnPosition()
					target.Health = target.MaxHealth
				}
			}
		}
	}
}
