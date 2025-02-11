package game

import (
	"log"
	"math"
	"sync"
	"time"
)

const (
	WorldWidth  = 800
	WorldHeight = 600

	WarriorMaxHealth   = 150.0
	WarriorAttackPower = 30.0
	WarriorSpeed       = 3.0
	WarriorAttackRange = 50.0

	MageMaxHealth   = 100.0
	MageAttackPower = 40.0
	MageSpeed       = 2.5
	MageAttackRange = 200.0
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (p Position) Distance(other Position) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

type CharacterClass int

const (
	Warrior CharacterClass = iota
	Mage
)

type Character struct {
	ID            string         `json:"id"`
	Class         CharacterClass `json:"class"`
	Position      Position       `json:"position"`
	Health        float64        `json:"health"`
	MaxHealth     float64        `json:"maxHealth"`
	AttackPower   float64        `json:"attackPower"`
	MovementSpeed float64        `json:"movementSpeed"`
	LastAttack    time.Time      `json:"-"`
	mu            sync.RWMutex   `json:"-"`
}

func NewCharacter(id string, class CharacterClass, pos Position) *Character {
	c := &Character{
		ID:       id,
		Class:    class,
		Position: pos,
	}

	switch class {
	case Warrior:
		c.MaxHealth = WarriorMaxHealth
		c.AttackPower = WarriorAttackPower
		c.MovementSpeed = WarriorSpeed
	case Mage:
		c.MaxHealth = MageMaxHealth
		c.AttackPower = MageAttackPower
		c.MovementSpeed = MageSpeed
	}

	c.Health = c.MaxHealth
	return c
}

func (c *Character) SetPosition(pos Position) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clamp position within world bounds
	c.Position.X = math.Max(0, math.Min(WorldWidth, pos.X))
	c.Position.Y = math.Max(0, math.Min(WorldHeight, pos.Y))

	log.Printf("Character %s position set to (%f, %f)", c.ID, c.Position.X, c.Position.Y)
}

func (c *Character) GetPosition() Position {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Position
}

func (c *Character) TakeDamage(amount float64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Health -= amount
	if c.Health <= 0 {
		c.Health = 0
		return true // Character died
	}
	return false
}
