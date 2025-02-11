package game

import (
	"math"
	"sync"
	"time"
)

// Position represents a point in 2D space
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Distance calculates the distance to another position
func (p Position) Distance(other Position) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// CharacterClass represents different character types
type CharacterClass int

const (
	Warrior CharacterClass = iota
	Mage
)

// Game world constants
const (
	WorldWidth  = 800
	WorldHeight = 600

	WarriorAttackRange = 50.0
	MageAttackRange    = 200.0
)

// Character represents a player in the game
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

// NewCharacter creates a new character
func NewCharacter(id string, class CharacterClass, pos Position) *Character {
	c := &Character{
		ID:       id,
		Class:    class,
		Position: pos,
	}

	switch class {
	case Warrior:
		c.MaxHealth = 150
		c.AttackPower = 30
		c.MovementSpeed = 3
	case Mage:
		c.MaxHealth = 100
		c.AttackPower = 40
		c.MovementSpeed = 2.5
	}

	c.Health = c.MaxHealth
	return c
}

// TakeDamage applies damage to the character
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

// UpdatePosition moves the character
func (c *Character) UpdatePosition(target Position, delta time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	dx := target.X - c.Position.X
	dy := target.Y - c.Position.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance < 0.1 {
		return
	}

	dx /= distance
	dy /= distance

	moveAmount := c.MovementSpeed * delta.Seconds()
	if moveAmount > distance {
		moveAmount = distance
	}

	c.Position.X += dx * moveAmount
	c.Position.Y += dy * moveAmount
}
