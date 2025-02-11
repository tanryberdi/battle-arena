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

type GameServer struct {
	World   *game.World
	clients map[string]net.Conn
	mu      sync.RWMutex
}

func NewGameServer() *GameServer {
	return &GameServer{
		World:   game.NewWorld(),
		clients: make(map[string]net.Conn),
	}
}

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

	// Create character
	char := game.NewCharacter(
		playerID,
		game.CharacterClass(rand.Intn(2)),
		s.World.GetRandomSpawnPosition(),
	)

	s.World.AddCharacter(char)

	s.mu.Lock()
	s.clients[playerID] = conn
	s.mu.Unlock()

	// Send initial state
	s.sendGameState(conn)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Received raw message from %s: %s", playerID, line)

		var cmd struct {
			Type string  `json:"type"`
			X    float64 `json:"x"`
			Y    float64 `json:"y"`
		}

		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			log.Printf("Error decoding message from %s: %v", playerID, err)
			continue
		}

		log.Printf("Decoded command from %s: %+v", playerID, cmd)

		if cmd.Type == "move" {
			char := s.World.GetCharacter(playerID)
			if char != nil {
				newPos := game.Position{X: cmd.X, Y: cmd.Y}
				log.Printf("Moving player %s to position: (%f, %f)", playerID, newPos.X, newPos.Y)

				char.SetPosition(newPos)
				log.Printf("Player %s new position: (%f, %f)", playerID, char.GetPosition().X, char.GetPosition().Y)
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
	ticker := time.NewTicker(time.Second / 60) // 60 FPS
	defer ticker.Stop()

	for range ticker.C {
		s.World.ProcessCombat()
		s.broadcastGameState()
	}
}

func (s *GameServer) broadcastGameState() {
	state := map[string]interface{}{
		"characters": s.World.GetCharacters(),
	}

	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("Error marshaling game state: %v", err)
		return
	}

	s.mu.RLock()
	for id, conn := range s.clients {
		if _, err := conn.Write(append(data, '\n')); err != nil {
			log.Printf("Error sending state to player %s: %v", id, err)
		}
	}
	s.mu.RUnlock()
}

func (s *GameServer) sendGameState(conn net.Conn) {
	state := map[string]interface{}{
		"characters": s.World.GetCharacters(),
	}

	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("Error marshaling game state: %v", err)
		return
	}

	if _, err := conn.Write(append(data, '\n')); err != nil {
		log.Printf("Error sending initial state: %v", err)
	}
}
