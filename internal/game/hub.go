package game

import (
	"log"
	"time"

	"github.com/anon/tictactoe-dg-lab/internal/config"
	"github.com/anon/tictactoe-dg-lab/internal/dglab"
)

// Hub 游戏连接管理中心
type Hub struct {
	roomManager   *RoomManager
	handleMessage chan *PlayerMessage
	register      chan *Player
	unregister    chan *Player
	dglabHub      *dglab.Hub     // DG-LAB WebSocket Hub
	config        *config.Config // 配置
}

// PlayerMessage 玩家消息（内部使用）
type PlayerMessage struct {
	Player  *Player
	Message *Message
}

// NewHub 创建游戏 Hub
func NewHub(dglabHub *dglab.Hub, cfg *config.Config) *Hub {
	return &Hub{
		roomManager:   NewRoomManager(),
		handleMessage: make(chan *PlayerMessage, 256),
		register:      make(chan *Player, 256),
		unregister:    make(chan *Player, 256),
		dglabHub:      dglabHub,
		config:        cfg,
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

		// 先从房间中移除玩家（这样广播时就不会向已断开的玩家发送消息）
		player.Room.RemovePlayer(player)

		// 通知房间内其他玩家
		player.Room.BroadcastRoomState(h.dglabHub.IsDeviceConnected)

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

	// 如果提供了房间 ID，尝试加入现有房间，如果不存在则创建
	if msg.RoomID != "" {
		room, err = h.roomManager.GetRoom(msg.RoomID)
		if err == ErrRoomNotFound {
			// 房间不存在，创建新房间
			room = h.roomManager.CreateRoomWithID(msg.RoomID)
			log.Printf("[Hub] Player %s created new room %s with specified ID", player.Name, room.ID)
		} else if err != nil {
			// 其他错误
			log.Printf("[Hub] Player %s failed to get room %s: %v", player.Name, msg.RoomID, err)
			player.SendError(err.Error())
			return
		} else {
			log.Printf("[Hub] Player %s joining existing room %s", player.Name, room.ID)
		}
	} else {
		// 否则创建新房间（随机ID）
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
	room.BroadcastRoomState(h.dglabHub.IsDeviceConnected)
}

// handleUpdateDGLabID 处理更新 DG-LAB ID 请求
func (h *Hub) handleUpdateDGLabID(player *Player, msg *Message) {
	player.UpdateDGLabID(msg.DGLabClientID)

	// 在 DG-LAB Hub 中预注册这个虚拟客户端
	// 这样当 APP 扫码连接并发送 bind 请求时，Hub 能够找到这个 clientID
	if msg.DGLabClientID != "" {
		h.dglabHub.PreRegisterClient(msg.DGLabClientID)
	}

	// 广播房间状态（更新设备连接状态）
	if player.Room != nil {
		player.Room.BroadcastRoomState(h.dglabHub.IsDeviceConnected)
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

	// 触发落子震动（移动前触发，避免游戏结束后被覆盖）
	h.triggerMoveShock(player)

	// 广播房间状态
	room.BroadcastRoomState(h.dglabHub.IsDeviceConnected)

	// 如果游戏结束，处理游戏结束逻辑
	if room.GameOver {
		room.BroadcastGameOver()
		h.triggerGameOverShock(room)
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

	// 验证持续时间
	if msg.Duration < h.config.Game.PunishmentDurationMin || msg.Duration > h.config.Game.PunishmentDurationMax {
		player.SendError("Duration out of range")
		return
	}

	// 触发惩罚震动
	loser := room.GetOpponent(player)
	if loser != nil {
		h.triggerPunishmentShock(loser, msg.Percent, msg.Duration)
	}

	log.Printf("[Hub] Player %s sent punishment to opponent: %d%%, %.1fs", player.Name, msg.Percent, msg.Duration)
}

// GetRoomManager 获取房间管理器（用于测试）
func (h *Hub) GetRoomManager() *RoomManager {
	return h.roomManager
}

// triggerMoveShock 触发落子震动
func (h *Hub) triggerMoveShock(player *Player) {
	if player.GetDGLabID() == "" {
		log.Printf("[Hub] Player %s has no DG-LAB device connected, skipping move shock", player.Name)
		return
	}

	// 获取玩家配置的强度 (0-100)，需要映射到设备强度 (0-200)
	strength := mapStrengthToDevice(player.Config.MoveStrength)

	// 计算需要的波形数据量（每个波形数据是100ms）
	duration := h.config.Game.MoveDuration
	waveformCount := int(duration / 0.1) // 0.1秒 = 100ms
	if waveformCount < 1 {
		waveformCount = 1
	}

	// 生成波形数组（重复配置中的default波形）
	waveformData := make([]string, waveformCount)
	for i := 0; i < waveformCount; i++ {
		waveformData[i] = h.config.Waveforms.Default
	}

	clientID := player.GetDGLabID()

	// 清空A和B通道的队列
	h.dglabHub.ClearQueue(clientID, dglab.ChannelA)
	h.dglabHub.ClearQueue(clientID, dglab.ChannelB)

	// 延迟150ms，确保清空指令先被处理，避免网络乱序
	time.Sleep(150 * time.Millisecond)

	// 设置A通道强度
	err := h.dglabHub.SendStrengthSet(clientID, dglab.ChannelA, strength)
	if err != nil {
		log.Printf("[Hub] Failed to set strength for A channel for player %s: %v", player.Name, err)
		return
	}

	// 设置B通道强度
	err = h.dglabHub.SendStrengthSet(clientID, dglab.ChannelB, strength)
	if err != nil {
		log.Printf("[Hub] Failed to set strength for B channel for player %s: %v", player.Name, err)
		return
	}

	// 发送波形到A通道
	err = h.dglabHub.SendPulse(clientID, "A", waveformData)
	if err != nil {
		log.Printf("[Hub] Failed to send move shock pulse to A channel for player %s: %v", player.Name, err)
		return
	}

	// 发送波形到B通道
	err = h.dglabHub.SendPulse(clientID, "B", waveformData)
	if err != nil {
		log.Printf("[Hub] Failed to send move shock pulse to B channel for player %s: %v", player.Name, err)
		return
	}

	log.Printf("[Hub] Sent move shock to player %s: strength=%d (device=%d), duration=%.1fs, channels=A+B",
		player.Name, player.Config.MoveStrength, strength, duration)

	// 广播震动事件通知
	if player.Room != nil {
		player.Room.Broadcast(&Message{
			Type:      TypeShockEvent,
			Target:    player.Name,
			Intensity: player.Config.MoveStrength,
			Reason:    "move",
		})
	}
}

// triggerGameOverShock 触发游戏结束震动
func (h *Hub) triggerGameOverShock(room *Room) {
	if room.Winner == 0 {
		// 平局，向双方发送平局震动
		h.triggerDrawShock(room)
	} else {
		// 有赢家，向输家发送失败震动（暂不实现，留待惩罚机制）
		// 这里可以添加一个默认的"输了"震动
		loser := room.GetPlayerBySymbol(3 - room.Winner) // 1->2, 2->1
		if loser != nil && loser.GetDGLabID() != "" {
			// 可以选择发送一个默认的失败震动
			log.Printf("[Hub] Player %s lost the game", loser.Name)
		}
	}
}

// triggerDrawShock 触发平局震动
func (h *Hub) triggerDrawShock(room *Room) {
	players := []*Player{room.PlayerX, room.PlayerO}

	// 计算需要的波形数据量
	duration := h.config.Game.DrawDuration
	waveformCount := int(duration / 0.1)
	if waveformCount < 1 {
		waveformCount = 1
	}

	// 生成波形数组
	waveformData := make([]string, waveformCount)
	for i := 0; i < waveformCount; i++ {
		waveformData[i] = h.config.Waveforms.Default
	}

	for _, player := range players {
		if player == nil || player.GetDGLabID() == "" {
			continue
		}

		// 获取玩家配置的平局强度
		strength := mapStrengthToDevice(player.Config.DrawStrength)
		clientID := player.GetDGLabID()

		// 清空A和B通道的队列
		h.dglabHub.ClearQueue(clientID, dglab.ChannelA)
		h.dglabHub.ClearQueue(clientID, dglab.ChannelB)

		// 延迟150ms，确保清空指令先被处理，避免网络乱序
		time.Sleep(150 * time.Millisecond)

		// 设置A通道强度
		err := h.dglabHub.SendStrengthSet(clientID, dglab.ChannelA, strength)
		if err != nil {
			log.Printf("[Hub] Failed to set strength for A channel for player %s: %v", player.Name, err)
			continue
		}

		// 设置B通道强度
		err = h.dglabHub.SendStrengthSet(clientID, dglab.ChannelB, strength)
		if err != nil {
			log.Printf("[Hub] Failed to set strength for B channel for player %s: %v", player.Name, err)
			continue
		}

		// 发送波形到A通道
		err = h.dglabHub.SendPulse(clientID, "A", waveformData)
		if err != nil {
			log.Printf("[Hub] Failed to send draw shock pulse to A channel for player %s: %v", player.Name, err)
			continue
		}

		// 发送波形到B通道
		err = h.dglabHub.SendPulse(clientID, "B", waveformData)
		if err != nil {
			log.Printf("[Hub] Failed to send draw shock pulse to B channel for player %s: %v", player.Name, err)
			continue
		}

		log.Printf("[Hub] Sent draw shock to player %s: strength=%d (device=%d), duration=%.1fs, channels=A+B",
			player.Name, player.Config.DrawStrength, strength, duration)

		// 广播震动事件通知
		room.Broadcast(&Message{
			Type:      TypeShockEvent,
			Target:    player.Name,
			Intensity: player.Config.DrawStrength,
			Reason:    "draw",
		})
	}
}

// triggerPunishmentShock 触发惩罚震动
func (h *Hub) triggerPunishmentShock(loser *Player, percent int, duration float64) {
	if loser.GetDGLabID() == "" {
		log.Printf("[Hub] Player %s has no DG-LAB device connected, skipping punishment", loser.Name)
		return
	}

	// 计算惩罚强度：基于输家的安全范围
	// actual_strength = safe_min + (safe_max - safe_min) * percent / 100
	strengthRange := loser.Config.SafeMax - loser.Config.SafeMin
	actualStrength := loser.Config.SafeMin + (strengthRange * percent / 100)

	// 映射到设备强度 (0-200)
	deviceStrength := mapStrengthToDevice(actualStrength)

	// 计算需要的波形数据量（根据duration参数）
	waveformCount := int(duration / 0.1)
	if waveformCount < 1 {
		waveformCount = 1
	}
	if waveformCount > 100 {
		waveformCount = 100 // DG-LAB限制：最大100个波形数据
	}

	// 生成波形数组（使用pulse波形，更强烈）
	waveformData := make([]string, waveformCount)
	for i := 0; i < waveformCount; i++ {
		waveformData[i] = h.config.Waveforms.Pulse
	}

	clientID := loser.GetDGLabID()

	// 清空A和B通道的队列
	h.dglabHub.ClearQueue(clientID, dglab.ChannelA)
	h.dglabHub.ClearQueue(clientID, dglab.ChannelB)

	// 延迟150ms，确保清空指令先被处理，避免网络乱序
	time.Sleep(150 * time.Millisecond)

	// 设置A通道强度
	err := h.dglabHub.SendStrengthSet(clientID, dglab.ChannelA, deviceStrength)
	if err != nil {
		log.Printf("[Hub] Failed to set strength for A channel for player %s: %v", loser.Name, err)
		return
	}

	// 设置B通道强度
	err = h.dglabHub.SendStrengthSet(clientID, dglab.ChannelB, deviceStrength)
	if err != nil {
		log.Printf("[Hub] Failed to set strength for B channel for player %s: %v", loser.Name, err)
		return
	}

	// 发送波形到A通道
	err = h.dglabHub.SendPulse(clientID, "A", waveformData)
	if err != nil {
		log.Printf("[Hub] Failed to send punishment shock pulse to A channel for player %s: %v", loser.Name, err)
		return
	}

	// 发送波形到B通道
	err = h.dglabHub.SendPulse(clientID, "B", waveformData)
	if err != nil {
		log.Printf("[Hub] Failed to send punishment shock pulse to B channel for player %s: %v", loser.Name, err)
		return
	}

	log.Printf("[Hub] Sent punishment shock to player %s: percent=%d%%, duration=%.1fs, strength=%d (device=%d), channels=A+B",
		loser.Name, percent, duration, actualStrength, deviceStrength)

	// 广播震动事件通知
	if loser.Room != nil {
		loser.Room.Broadcast(&Message{
			Type:      TypeShockEvent,
			Target:    loser.Name,
			Intensity: actualStrength,
			Reason:    "punish",
		})
	}
}

// mapStrengthToDevice 将用户强度 (0-100) 映射到设备强度 (0-200)
func mapStrengthToDevice(userStrength int) int {
	if userStrength < 0 {
		userStrength = 0
	}
	if userStrength > 100 {
		userStrength = 100
	}
	return userStrength * 2
}

// NotifyDeviceConnected 通知设备已连接
// 当DG-LAB设备成功绑定时调用，触发房间状态广播
func (h *Hub) NotifyDeviceConnected(clientID string) {
	log.Printf("[Hub] Device connected notification for clientID: %s", clientID)

	// 查找使用这个clientID的房间
	room := h.roomManager.FindRoomByDGLabID(clientID)
	if room == nil {
		log.Printf("[Hub] No room found for clientID: %s", clientID)
		return
	}

	log.Printf("[Hub] Found room %s for device %s, broadcasting state update", room.ID, clientID)
	// 广播房间状态更新
	room.BroadcastRoomState(h.dglabHub.IsDeviceConnected)
}

// NotifyDeviceDisconnected 通知设备已断开
// 当DG-LAB设备断开连接时调用，触发房间状态广播
func (h *Hub) NotifyDeviceDisconnected(clientID string) {
	log.Printf("[Hub] Device disconnected notification for clientID: %s", clientID)

	// 查找使用这个clientID的房间
	room := h.roomManager.FindRoomByDGLabID(clientID)
	if room == nil {
		log.Printf("[Hub] No room found for clientID: %s", clientID)
		return
	}

	log.Printf("[Hub] Found room %s for device %s, broadcasting state update after disconnect", room.ID, clientID)
	// 广播房间状态更新
	room.BroadcastRoomState(h.dglabHub.IsDeviceConnected)
}
