package client

import (
	"encoding/json"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	"battle-arena/pkg/game"
)

type GameClient struct {
	conn       net.Conn
	characters map[string]*game.Character
	playerID   string
	window     *pixelgl.Window
	imd        *imdraw.IMDraw
	mu         sync.RWMutex
}

func NewGameClient(serverAddr string) (*GameClient, error) {
	log.Printf("Connecting to server at %s", serverAddr)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to server successfully")

	return &GameClient{
		conn:       conn,
		characters: make(map[string]*game.Character),
	}, nil
}

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

	// Game loop
	ticker := time.NewTicker(time.Second / 60) // 60 FPS
	defer ticker.Stop()

	log.Printf("Starting game loop")
	for !c.window.Closed() {
		c.handleInput()
		c.draw()
		c.window.Update()
		<-ticker.C
	}

	return nil
}

func (c *GameClient) handleInput() {
	// Handle mouse movement
	if c.window.JustPressed(pixelgl.MouseButtonLeft) {
		mousePos := c.window.MousePosition()
		log.Printf("Mouse clicked at position: %v", mousePos)

		moveCmd := struct {
			Type string  `json:"type"`
			X    float64 `json:"x"`
			Y    float64 `json:"y"`
		}{
			Type: "move",
			X:    mousePos.X,
			Y:    mousePos.Y,
		}

		data, err := json.Marshal(moveCmd)
		if err != nil {
			log.Printf("Error marshaling move command: %v", err)
			return
		}
		data = append(data, '\n')
		c.conn.Write(data)
	}

	// Handle keyboard movement
	if c.playerID != "" {
		c.mu.RLock()
		char, exists := c.characters[c.playerID]
		c.mu.RUnlock()

		if exists {
			var newPos game.Position
			moveSpeed := 5.0 // Movement speed per frame
			moved := false

			// Get current position
			newPos = char.Position

			// Check keyboard input
			if c.window.Pressed(pixelgl.KeyLeft) || c.window.Pressed(pixelgl.KeyA) {
				newPos.X -= moveSpeed
				moved = true
			}
			if c.window.Pressed(pixelgl.KeyRight) || c.window.Pressed(pixelgl.KeyD) {
				newPos.X += moveSpeed
				moved = true
			}
			if c.window.Pressed(pixelgl.KeyUp) || c.window.Pressed(pixelgl.KeyW) {
				newPos.Y += moveSpeed
				moved = true
			}
			if c.window.Pressed(pixelgl.KeyDown) || c.window.Pressed(pixelgl.KeyS) {
				newPos.Y -= moveSpeed
				moved = true
			}

			// If any movement key was pressed, send the move command
			if moved {
				// Clamp position to world bounds
				newPos.X = math.Max(0, math.Min(game.WorldWidth, newPos.X))
				newPos.Y = math.Max(0, math.Min(game.WorldHeight, newPos.Y))

				moveCmd := struct {
					Type string  `json:"type"`
					X    float64 `json:"x"`
					Y    float64 `json:"y"`
				}{
					Type: "move",
					X:    newPos.X,
					Y:    newPos.Y,
				}

				data, err := json.Marshal(moveCmd)
				if err != nil {
					log.Printf("Error marshaling move command: %v", err)
					return
				}
				data = append(data, '\n')
				c.conn.Write(data)
			}
		}
	}
}

func (c *GameClient) receiveUpdates() {
	decoder := json.NewDecoder(c.conn)
	for {
		var state map[string]interface{}
		if err := decoder.Decode(&state); err != nil {
			log.Printf("Error decoding state: %v", err)
			return
		}

		if chars, ok := state["characters"].(map[string]interface{}); ok {
			c.mu.Lock()
			for id, charData := range chars {
				if charMap, ok := charData.(map[string]interface{}); ok {
					if c.playerID == "" {
						c.playerID = id
						log.Printf("Received player ID: %s", id)
					}

					pos := charMap["position"].(map[string]interface{})
					char := &game.Character{
						ID: id,
						Position: game.Position{
							X: pos["x"].(float64),
							Y: pos["y"].(float64),
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

func (c *GameClient) draw() {
	c.window.Clear(colornames.Black)
	c.imd.Clear()

	// Draw world border
	c.imd.Color = pixel.RGB(0.2, 0.2, 0.2)
	c.imd.Push(pixel.V(0, 0), pixel.V(float64(game.WorldWidth), float64(game.WorldHeight)))
	c.imd.Rectangle(1)

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Draw characters
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
	}

	c.imd.Draw(c.window)
}
