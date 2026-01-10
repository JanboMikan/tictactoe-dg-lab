package game

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType 定义游戏消息类型
type MessageType string

const (
	// Client -> Server
	TypeJoinRoom      MessageType = "join_room"
	TypeUpdateDGLabID MessageType = "update_dglab_id"
	TypeUpdateConfig  MessageType = "update_config"
	TypeMove          MessageType = "move"
	TypePunish        MessageType = "punish"

	// Server -> Client
	TypeRoomState  MessageType = "room_state"
	TypeGameOver   MessageType = "game_over"
	TypeShockEvent MessageType = "shock_event"
	TypeError      MessageType = "error"
)

// Message 定义游戏 WebSocket 消息格式
type Message struct {
	Type          MessageType   `json:"type"`
	RoomID        string        `json:"room_id,omitempty"`
	PlayerName    string        `json:"player_name,omitempty"`
	DGLabClientID string        `json:"dglab_client_id,omitempty"`
	Position      int           `json:"position,omitempty"`
	Config        *PlayerConfig `json:"config,omitempty"`
	Percent       int           `json:"percent,omitempty"`
	Duration      float64       `json:"duration,omitempty"`
	Message       string        `json:"message,omitempty"`
	Error         string        `json:"error,omitempty"` // 用于错误消息

	// 以下字段仅用于 Server -> Client 消息
	Board     []int                  `json:"board,omitempty"`
	Turn      string                 `json:"turn,omitempty"`
	Players   map[string]*PlayerInfo `json:"players,omitempty"`
	Winner    string                 `json:"winner,omitempty"`
	Line      []int                  `json:"line,omitempty"`
	Target    string                 `json:"target,omitempty"`
	Intensity int                    `json:"intensity,omitempty"`
	Reason    string                 `json:"reason,omitempty"`
}

// PlayerConfig 玩家配置
type PlayerConfig struct {
	SafeMin      int `json:"safe_min"`      // 0-100
	SafeMax      int `json:"safe_max"`      // 0-100
	MoveStrength int `json:"move_strength"` // safe_min ~ safe_max
	DrawStrength int `json:"draw_strength"` // safe_min ~ safe_max
}

// PlayerInfo 用于广播的玩家信息
type PlayerInfo struct {
	Connected    bool `json:"connected"`
	DeviceActive bool `json:"device_active"`
}

// Player 玩家
type Player struct {
	Name          string
	Conn          *websocket.Conn
	DGLabClientID string
	Config        *PlayerConfig
	Symbol        int // 1: X (先手), 2: O (后手)
	Send          chan []byte
	Room          *Room
	mu            sync.RWMutex
}

// Room 房间
type Room struct {
	ID          string
	Board       [9]int // 0: empty, 1: X, 2: O
	Turn        int    // 1 或 2，表示当前轮到谁
	PlayerX     *Player
	PlayerO     *Player
	GameOver    bool
	Winner      int   // 0: 平局, 1: X胜, 2: O胜
	WinningLine []int // 获胜的三个位置
	CreatedAt   time.Time
	mu          sync.RWMutex
}

// GetPlayerBySymbol 根据符号获取玩家
func (r *Room) GetPlayerBySymbol(symbol int) *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if symbol == 1 {
		return r.PlayerX
	}
	return r.PlayerO
}

// GetPlayerByName 根据名字获取玩家
func (r *Room) GetPlayerByName(name string) *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.PlayerX != nil && r.PlayerX.Name == name {
		return r.PlayerX
	}
	if r.PlayerO != nil && r.PlayerO.Name == name {
		return r.PlayerO
	}
	return nil
}

// GetOpponent 获取对手
func (r *Room) GetOpponent(player *Player) *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if player == r.PlayerX {
		return r.PlayerO
	}
	return r.PlayerX
}

// IsFull 房间是否已满
func (r *Room) IsFull() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.PlayerX != nil && r.PlayerO != nil
}

// GetPlayerCount 获取房间内玩家数量
func (r *Room) GetPlayerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	if r.PlayerX != nil {
		count++
	}
	if r.PlayerO != nil {
		count++
	}
	return count
}

// DefaultPlayerConfig 返回默认的玩家配置
func DefaultPlayerConfig() *PlayerConfig {
	return &PlayerConfig{
		SafeMin:      10,
		SafeMax:      30,
		MoveStrength: 10,
		DrawStrength: 15,
	}
}

// Validate 验证玩家配置的合法性
func (c *PlayerConfig) Validate() error {
	if c.SafeMin < 0 || c.SafeMin > 100 {
		return ErrInvalidConfig
	}
	if c.SafeMax < 0 || c.SafeMax > 100 {
		return ErrInvalidConfig
	}
	if c.SafeMin >= c.SafeMax {
		return ErrInvalidConfig
	}
	if c.MoveStrength < c.SafeMin || c.MoveStrength > c.SafeMax {
		return ErrInvalidConfig
	}
	if c.DrawStrength < c.SafeMin || c.DrawStrength > c.SafeMax {
		return ErrInvalidConfig
	}
	return nil
}
