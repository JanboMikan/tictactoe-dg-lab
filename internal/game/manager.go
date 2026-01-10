package game

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RoomManager 房间管理器
type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

// NewRoomManager 创建房间管理器
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

// CreateRoom 创建新房间
func (m *RoomManager) CreateRoom() *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	roomID := generateRoomID()
	room := &Room{
		ID:        roomID,
		Board:     [9]int{},
		Turn:      1, // X 先手
		CreatedAt: time.Now(),
	}

	m.rooms[roomID] = room
	log.Printf("[RoomManager] Room %s created", roomID)
	return room
}

// CreateRoomWithID 创建指定ID的新房间
func (m *RoomManager) CreateRoomWithID(roomID string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查房间是否已存在
	if _, exists := m.rooms[roomID]; exists {
		log.Printf("[RoomManager] Room %s already exists", roomID)
		return m.rooms[roomID]
	}

	room := &Room{
		ID:        roomID,
		Board:     [9]int{},
		Turn:      1, // X 先手
		CreatedAt: time.Now(),
	}

	m.rooms[roomID] = room
	log.Printf("[RoomManager] Room %s created with specified ID", roomID)
	return room
}

// GetRoom 获取房间
func (m *RoomManager) GetRoom(roomID string) (*Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, exists := m.rooms[roomID]
	if !exists {
		return nil, ErrRoomNotFound
	}
	return room, nil
}

// JoinRoom 加入房间
func (m *RoomManager) JoinRoom(roomID string, player *Player) error {
	room, err := m.GetRoom(roomID)
	if err != nil {
		return err
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// 检查房间是否已满（直接检查字段，因为已持有锁）
	if room.PlayerX != nil && room.PlayerO != nil {
		return ErrRoomFull
	}

	// 分配玩家位置
	if room.PlayerX == nil {
		room.PlayerX = player
		player.Symbol = 1
		player.Room = room
		log.Printf("[Room %s] Player %s joined as X", roomID, player.Name)
	} else if room.PlayerO == nil {
		room.PlayerO = player
		player.Symbol = 2
		player.Room = room
		log.Printf("[Room %s] Player %s joined as O", roomID, player.Name)
	} else {
		return ErrRoomFull
	}

	return nil
}

// DeleteRoom 删除房间
func (m *RoomManager) DeleteRoom(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, exists := m.rooms[roomID]; exists {
		delete(m.rooms, roomID)
		log.Printf("[RoomManager] Room %s deleted (PlayerX: %v, PlayerO: %v)",
			roomID, room.PlayerX != nil, room.PlayerO != nil)
	}
}

// GetRoomCount 获取房间数量
func (m *RoomManager) GetRoomCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// CleanEmptyRooms 清理空房间（定时任务）
func (m *RoomManager) CleanEmptyRooms() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for roomID, room := range m.rooms {
		if room.IsEmpty() {
			// 如果房间空闲超过 10 分钟，删除
			if time.Since(room.CreatedAt) > 10*time.Minute {
				delete(m.rooms, roomID)
				log.Printf("[RoomManager] Empty room %s cleaned up", roomID)
			}
		}
	}
}

// generateRoomID 生成房间 ID（6位随机数字）
func generateRoomID() string {
	// 使用 UUID 的前6个字符作为房间 ID
	id := uuid.New().String()
	return id[:6]
}
