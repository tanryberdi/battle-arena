package client

import (
	"encoding/json"
	"log"
	"net"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	"battle-arena/pkg/game"
)

// GameClient represents the game client
type GameClient struct {
	conn       net.Conn
	characters map[string]*game.Character
	playerID   string
	window     *pixelgl.Window
	imd        *imdraw.IMDraw
	mu         sync.RWMutex
}

// NewGameClient creates a new game client
func NewGameClient(serverAddr string) (*GameClient, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	return &GameClient{
		conn:       conn,
		characters: make(map[string]*game.Character),
	}, nil
}

// Start initializes and starts the game client
func (c *GameClient) Start() error {
	cfg := pixelgl.WindowConfig{
		Title:  "Battle Arena",
		Bounds: pixel.R(0, 0, float64(game.WorldWidth), float64(game.WorldHeight)),
		VSync:  true,
	}

	var err error
	c.window, err = pixelgl.NewWindow(cfg)
	if err != nil {
		return err
	}

	c.imd = imdraw.New(nil)

	// Start receiving game state updates
	go c.receiveUpdates()

	// Main game loop
	for !c.window.Closed() {
		c.handleInput()
		c.draw()
		c.window.Update()
	}

	return nil
}

func (c *GameClient) receiveUpdates() {
	decoder := json.NewDecoder(c.conn)
	for {
		var state map[string]interface{}
		if err := decoder.Decode(&state); err != nil {
			log.Printf("Error decoding state: %v", err)
			return
		}

		// Update characters
		if chars, ok := state["characters"].(map[string]interface{}); ok {
			c.mu.Lock()
			for id, charData := range chars {
				if charMap, ok := charData.(map[string]interface{}); ok {
					// If this is our first update, set our player ID
					if c.playerID == "" {
						c.playerID = id
					}

					// Update or create character
					char := &game.Character{
						ID: id,
						Position: game.Position{
							X: charMap["position"].(map[string]interface{})["x"].(float64),
							Y: charMap["position"].(map[string]interface{})["y"].(float64),
						},
						Health:    charMap["health"].(float64),
						MaxHealth: charMap["maxHealth"].(float64),
						Class:     game.CharacterClass(int(charMap["class"].(float64))),
					}
					c.characters[id] = char
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *GameClient) handleInput() {
	if c.window.Pressed(pixelgl.MouseButtonLeft) {
		mousePos := c.window.MousePosition()
		cmd := struct {
			Type string  `json:"type"`
			X    float64 `json:"x"`
			Y    float64 `json:"y"`
		}{
			Type: "move",
			X:    mousePos.X,
			Y:    mousePos.Y,
		}

		data, err := json.Marshal(cmd)
		if err == nil {
			c.conn.Write(data)
		}
	}
}

func (c *GameClient) draw() {
	c.window.Clear(colornames.Black)
	c.imd.Clear()

	// Draw world border
	c.imd.Color = pixel.RGB(0.2, 0.2, 0.2)
	c.imd.Push(pixel.V(0, 0), pixel.V(float64(game.WorldWidth), float64(game.WorldHeight)))
	c.imd.Rectangle(1)

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Draw all characters
	for id, char := range c.characters {
		isPlayer := id == c.playerID
		pos := pixel.V(char.Position.X, char.Position.Y)

		// Draw character body
		if isPlayer {
			c.imd.Color = pixel.RGB(0, 1, 0) // Green for player
		} else {
			c.imd.Color = pixel.RGB(1, 0, 0) // Red for enemies
		}

		// Draw character circle
		c.imd.Push(pos)
		c.imd.Circle(10, 0)

		// Draw health bar
		healthWidth := 30.0
		healthHeight := 3.0
		healthPos := pos.Add(pixel.V(-healthWidth/2, 15))

		// Health bar background
		c.imd.Color = pixel.RGB(0.3, 0.3, 0.3)
		c.imd.Push(healthPos, healthPos.Add(pixel.V(healthWidth, healthHeight)))
		c.imd.Rectangle(0)

		// Health bar fill
		healthPerc := char.Health / char.MaxHealth
		c.imd.Color = pixel.RGB(1-healthPerc, healthPerc, 0)
		c.imd.Push(healthPos, healthPos.Add(pixel.V(healthWidth*healthPerc, healthHeight)))
		c.imd.Rectangle(0)

		// Draw character class indicator
		switch char.Class {
		case game.Warrior:
			c.drawAttackRange(pos, game.WarriorAttackRange)
		case game.Mage:
			c.drawAttackRange(pos, game.MageAttackRange)
		}
	}

	c.imd.Draw(c.window)
}

func (c *GameClient) drawAttackRange(pos pixel.Vec, radius float64) {
	c.imd.Color = pixel.RGB(0.2, 0.2, 0.2)
	c.imd.Push(pos)
	c.imd.Circle(radius, 1)
}
