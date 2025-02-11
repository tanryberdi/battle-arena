# Battle Arena Game

A simple 2D multiplayer battle arena game written in Go, where players control warriors and mages in real-time combat.

## Features

- Real-time multiplayer combat
- Two character classes (warrior and mage)
- Different attack ranges and combat mechanics
- Health system with automatic respawning
- Dual control scheme (keyboard and mouse)

# Character Classes

## Warrior

- High health (150 HP)
- Close combat range (50 units)
- Physical damage dealer
- High movement speed

## Mage

- Lower Health (100 HP)
- Long attack range (200 units)
- Magical damage dealer
- Slower movement speed

# Controls

## Movement

- Arrow keys (up, down, left, right)
- WASD keys (W, A, S, D)
- Left (right) mouse click to move to cursor position

# Prerequisites

- Go 1.21 or higher
- github.com/faiface/pixel
- github.com/google/uuid
- golang.org/x/image

# Installation

1. Clone the repository

```bash
git clone https://github.com/tanryberdi/battle-arena
cd battle-arena
```

2. Install dependencies

```bash
go mod tidy
```

3. Build the project

```bash
make build
```

# Running the game

1. Start the server

```bash
make run-server
# Or with custom port:
make run-server PORT=9000
```

2. Start the client

```bash
make run-client
# Or with custom port:
make run-client PORT=9000
```

# Game mechanics

- Characters automatically attack enemies within their range
- Health bars display above each character
- Players respawn automatically when killed
- Green circle indicates your character
- Red circles indicate other players
- Circle around characters shows attack range

# Project Structure

```
battle-arena/
├── cmd/
│   ├── client/
│   │   └── main.go
│   └── server/
│       └── main.go
├── pkg/
│   ├── game/
│   │   ├── types.go
│   │   └── world.go
│   ├── server/
│   │   └── server.go
│   └── client/
│       └── client.go
├── Makefile
├── go.mod
└── README.md
```
