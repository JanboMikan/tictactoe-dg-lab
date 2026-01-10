package game

import (
	"log"
	"time"
)

// Hub 游戏连接管理中心
type Hub struct {
	roomManager   *RoomManager
	handleMessage chan *PlayerMessage
	register      chan *Player
	unregister    chan *Player
}

// PlayerMessage 玩家消息（内部使用）
type PlayerMessage struct {
	Player  *Player
	Message *Message
}

// NewHub 创建游戏 Hub
func NewHub() *Hub {
	return &Hub{
		roomManager:   NewRoomManager(),
		handleMessage: make(chan *PlayerMessage, 256),
		register:      make(chan *Player, 256),
		unregister:    make(chan *Player, 256),
	}
}

// Run 运行 Hub 主循环
func (h *Hub) Run() {
	// 启动定期清理空房间的任务
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case player := <-h.register:
			h.handleRegister(player)

		case player := <-h.unregister:
			h.handleUnregister(player)

		case pm := <-h.handleMessage:
			h.processMessage(pm)

		case <-ticker.C:
			h.roomManager.CleanEmptyRooms()
		}
	}
}

// handleRegister 处理玩家注册
func (h *Hub) handleRegister(player *Player) {
	log.Printf("[Hub] Player %s registered", player.Name)
}

// handleUnregister 处理玩家断开连接
func (h *Hub) handleUnregister(player *Player) {
	if player.Room != nil {
		log.Printf("[Hub] Player %s disconnected from room %s", player.Name, player.Room.ID)

		// 通知房间内其他玩家
		player.Room.BroadcastRoomState()

		// 如果房间为空，删除房间
		if player.Room.IsEmpty() {
			h.roomManager.DeleteRoom(player.Room.ID)
		}
	}

	// 关闭发送通道
	if player.Send != nil {
		close(player.Send)
	}

	log.Printf("[Hub] Player %s unregistered", player.Name)
}

// processMessage 处理玩家消息
func (h *Hub) processMessage(pm *PlayerMessage) {
	player := pm.Player
	msg := pm.Message

	switch msg.Type {
	case TypeJoinRoom:
		h.handleJoinRoom(player, msg)

	case TypeUpdateDGLabID:
		h.handleUpdateDGLabID(player, msg)

	case TypeUpdateConfig:
		h.handleUpdateConfig(player, msg)

	case TypeMove:
		h.handleMove(player, msg)

	case TypePunish:
		h.handlePunish(player, msg)

	default:
		log.Printf("[Hub] Unknown message type from player %s: %s", player.Name, msg.Type)
		player.SendError("Unknown message type")
	}
}

// handleJoinRoom 处理加入房间请求
func (h *Hub) handleJoinRoom(player *Player, msg *Message) {
	var room *Room
	var err error

	// 如果提供了房间 ID，尝试加入现有房间
	if msg.RoomID != "" {
		room, err = h.roomManager.GetRoom(msg.RoomID)
		if err != nil {
			log.Printf("[Hub] Player %s failed to join room %s: %v", player.Name, msg.RoomID, err)
			player.SendError("Room not found")
			return
		}
	} else {
		// 否则创建新房间
		room = h.roomManager.CreateRoom()
		log.Printf("[Hub] Player %s created new room %s", player.Name, room.ID)
	}

	// 设置玩家名字
	if msg.PlayerName != "" {
		player.Name = msg.PlayerName
	}

	// 加入房间
	err = h.roomManager.JoinRoom(room.ID, player)
	if err != nil {
		log.Printf("[Hub] Player %s failed to join room %s: %v", player.Name, room.ID, err)
		player.SendError(err.Error())
		return
	}

	// 设置默认配置
	player.Config = DefaultPlayerConfig()

	log.Printf("[Hub] Player %s joined room %s as Symbol %d", player.Name, room.ID, player.Symbol)

	// 广播房间状态
	room.BroadcastRoomState()
}

// handleUpdateDGLabID 处理更新 DG-LAB ID 请求
func (h *Hub) handleUpdateDGLabID(player *Player, msg *Message) {
	player.UpdateDGLabID(msg.DGLabClientID)

	// 广播房间状态（更新设备连接状态）
	if player.Room != nil {
		player.Room.BroadcastRoomState()
	}
}

// handleUpdateConfig 处理更新配置请求
func (h *Hub) handleUpdateConfig(player *Player, msg *Message) {
	if msg.Config == nil {
		player.SendError("Config is required")
		return
	}

	err := player.UpdateConfig(msg.Config)
	if err != nil {
		log.Printf("[Hub] Player %s failed to update config: %v", player.Name, err)
		player.SendError(err.Error())
		return
	}

	log.Printf("[Hub] Player %s updated config successfully", player.Name)
}

// handleMove 处理落子请求
func (h *Hub) handleMove(player *Player, msg *Message) {
	if player.Room == nil {
		player.SendError("Not in a room")
		return
	}

	room := player.Room

	// 执行落子
	err := room.MakeMove(player, msg.Position)
	if err != nil {
		log.Printf("[Hub] Player %s move failed: %v", player.Name, err)
		player.SendError(err.Error())
		return
	}

	log.Printf("[Hub] Player %s moved to position %d in room %s", player.Name, msg.Position, room.ID)

	// 广播房间状态
	room.BroadcastRoomState()

	// 如果游戏结束，广播游戏结束消息
	if room.GameOver {
		room.BroadcastGameOver()
	}
}

// handlePunish 处理惩罚请求
func (h *Hub) handlePunish(player *Player, msg *Message) {
	if player.Room == nil {
		player.SendError("Not in a room")
		return
	}

	room := player.Room

	// 检查游戏是否已结束
	if !room.GameOver {
		player.SendError("Game is not over yet")
		return
	}

	// 检查是否是赢家
	if room.Winner != player.Symbol {
		player.SendError(ErrNotWinner.Error())
		return
	}

	// 验证参数
	if msg.Percent < 1 || msg.Percent > 100 {
		player.SendError("Percent must be between 1 and 100")
		return
	}

	// 注意：实际的惩罚逻辑（发送震动）将在 Phase 5 实现
	// 这里只是记录日志
	log.Printf("[Hub] Player %s requested punishment: %d%%, %.1fs", player.Name, msg.Percent, msg.Duration)
}

// GetRoomManager 获取房间管理器（用于测试）
func (h *Hub) GetRoomManager() *RoomManager {
	return h.roomManager
}
