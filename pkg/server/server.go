package server

import (
	"bufio"
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"battle-arena/pkg/game"

	"github.com/google/uuid"
)

// GameServer represents the main server instance
type GameServer struct {
	World   *game.World
	clients map[string]net.Conn
	mu      sync.RWMutex
}

// Message represents a client-server message
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// MoveCommand represents a movement command from client
type MoveCommand struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NewGameServer creates a new game server instance
func NewGameServer() *GameServer {
	return &GameServer{
		World:   game.NewWorld(),
		clients: make(map[string]net.Conn),
	}
}

// Start begins the server on the specified port
func (s *GameServer) Start(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	log.Printf("Game server listening on port %s", port)

	go s.gameLoop()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *GameServer) handleConnection(conn net.Conn) {
	playerID := uuid.New().String()
	log.Printf("New player connected: %s", playerID)

	// Create new character
	char := game.NewCharacter(
		playerID,
		game.CharacterClass(rand.Intn(2)),
		s.World.GetRandomSpawnPosition(),
	)

	s.World.AddCharacter(char)

	s.mu.Lock()
	s.clients[playerID] = conn
	s.mu.Unlock()

	s.sendGameState(conn)

	// Handle incoming messages
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var cmd struct {
			Type string  `json:"type"`
			X    float64 `json:"x"`
			Y    float64 `json:"y"`
		}

		if err := json.Unmarshal(scanner.Bytes(), &cmd); err != nil {
			log.Printf("Error decoding message: %v", err)
			continue
		}

		log.Printf("Received command from %s: %+v", playerID, cmd)

		switch cmd.Type {
		case "move":
			if char := s.World.GetCharacter(playerID); char != nil {
				target := game.Position{X: cmd.X, Y: cmd.Y}
				log.Printf("Moving player %s to position: (%f, %f)", playerID, target.X, target.Y)
				char.UpdatePosition(target, 16*time.Millisecond)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from connection: %v", err)
	}

	log.Printf("Player disconnected: %s", playerID)
	s.mu.Lock()
	delete(s.clients, playerID)
	s.mu.Unlock()
	s.World.RemoveCharacter(playerID)
	conn.Close()
}

func (s *GameServer) gameLoop() {
	ticker := time.NewTicker(16 * time.Millisecond)
	for range ticker.C {
		s.World.ProcessCombat()
		s.broadcastGameState()
	}
}

func (s *GameServer) broadcastGameState() {
	state := make(map[string]interface{})
	state["characters"] = s.World.GetCharacters()

	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	s.mu.RLock()
	for _, conn := range s.clients {
		conn.Write(data)
	}
	s.mu.RUnlock()
}

func (s *GameServer) sendGameState(conn net.Conn) {
	state := make(map[string]interface{})
	state["characters"] = s.World.GetCharacters()

	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	conn.Write(data)
}
