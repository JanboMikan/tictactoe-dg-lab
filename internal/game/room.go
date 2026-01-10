package game

import (
	"encoding/json"
	"log"
)

// MakeMove 玩家落子
func (r *Room) MakeMove(player *Player, position int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查游戏是否已结束
	if r.GameOver {
		return ErrGameOver
	}

	// 检查是否轮到该玩家
	if player.Symbol != r.Turn {
		return ErrNotYourTurn
	}

	// 检查位置是否合法
	if position < 0 || position > 8 {
		return ErrInvalidMove
	}

	// 检查位置是否已被占用
	if r.Board[position] != 0 {
		return ErrPositionOccupied
	}

	// 落子
	r.Board[position] = player.Symbol
	log.Printf("[Room %s] Player %s (Symbol %d) moved to position %d", r.ID, player.Name, player.Symbol, position)

	// 检查胜负
	r.checkWin()

	// 切换回合（如果游戏未结束）
	if !r.GameOver {
		if r.Turn == 1 {
			r.Turn = 2
		} else {
			r.Turn = 1
		}
	}

	return nil
}

// checkWin 检查游戏是否结束（胜利或平局）
func (r *Room) checkWin() {
	// 所有可能的获胜组合
	winPatterns := [][]int{
		{0, 1, 2}, // 横排
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6}, // 竖排
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8}, // 对角线
		{2, 4, 6},
	}

	// 检查是否有人获胜
	for _, pattern := range winPatterns {
		a, b, c := pattern[0], pattern[1], pattern[2]
		if r.Board[a] != 0 && r.Board[a] == r.Board[b] && r.Board[b] == r.Board[c] {
			r.GameOver = true
			r.Winner = r.Board[a]
			r.WinningLine = pattern
			log.Printf("[Room %s] Game over! Winner: Symbol %d, Line: %v", r.ID, r.Winner, r.WinningLine)
			return
		}
	}

	// 检查是否平局（棋盘已满）
	full := true
	for _, cell := range r.Board {
		if cell == 0 {
			full = false
			break
		}
	}
	if full {
		r.GameOver = true
		r.Winner = 0 // 0 表示平局
		log.Printf("[Room %s] Game over! Draw!", r.ID)
	}
}

// Broadcast 向房间内所有玩家广播消息
func (r *Room) Broadcast(msg *Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[Room %s] Failed to marshal broadcast message: %v", r.ID, err)
		return
	}

	if r.PlayerX != nil && r.PlayerX.Conn != nil {
		select {
		case r.PlayerX.Send <- data:
		default:
			log.Printf("[Room %s] Failed to send to PlayerX %s (channel full)", r.ID, r.PlayerX.Name)
		}
	}

	if r.PlayerO != nil && r.PlayerO.Conn != nil {
		select {
		case r.PlayerO.Send <- data:
		default:
			log.Printf("[Room %s] Failed to send to PlayerO %s (channel full)", r.ID, r.PlayerO.Name)
		}
	}
}

// BroadcastRoomState 广播当前房间状态
func (r *Room) BroadcastRoomState() {
	r.mu.RLock()

	players := make(map[string]*PlayerInfo)

	if r.PlayerX != nil {
		players[r.PlayerX.Name] = &PlayerInfo{
			Connected:    r.PlayerX.Conn != nil,
			DeviceActive: r.PlayerX.DGLabClientID != "",
		}
	}

	if r.PlayerO != nil {
		players[r.PlayerO.Name] = &PlayerInfo{
			Connected:    r.PlayerO.Conn != nil,
			DeviceActive: r.PlayerO.DGLabClientID != "",
		}
	}

	// 获取当前回合的玩家名字
	turnName := ""
	if r.Turn == 1 && r.PlayerX != nil {
		turnName = r.PlayerX.Name
	} else if r.Turn == 2 && r.PlayerO != nil {
		turnName = r.PlayerO.Name
	}

	// 复制棋盘数据
	board := make([]int, 9)
	copy(board, r.Board[:])

	msg := &Message{
		Type:    TypeRoomState,
		Board:   board,
		Turn:    turnName,
		Players: players,
	}

	// 释放锁后再广播
	r.mu.RUnlock()
	r.Broadcast(msg)
}

// BroadcastGameOver 广播游戏结束消息
func (r *Room) BroadcastGameOver() {
	r.mu.RLock()

	winnerName := ""
	if r.Winner == 1 && r.PlayerX != nil {
		winnerName = r.PlayerX.Name
	} else if r.Winner == 2 && r.PlayerO != nil {
		winnerName = r.PlayerO.Name
	}
	// Winner == 0 表示平局，winnerName 为空

	msg := &Message{
		Type:   TypeGameOver,
		Winner: winnerName,
		Line:   r.WinningLine,
	}

	// 释放锁后再广播
	r.mu.RUnlock()
	r.Broadcast(msg)
}

// Reset 重置房间（重新开始游戏）
func (r *Room) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Board = [9]int{}
	r.Turn = 1
	r.GameOver = false
	r.Winner = 0
	r.WinningLine = nil

	log.Printf("[Room %s] Room reset", r.ID)
}

// RemovePlayer 移除玩家
func (r *Room) RemovePlayer(player *Player) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.PlayerX == player {
		r.PlayerX = nil
		log.Printf("[Room %s] PlayerX removed", r.ID)
	} else if r.PlayerO == player {
		r.PlayerO = nil
		log.Printf("[Room %s] PlayerO removed", r.ID)
	}
}

// IsEmpty 房间是否为空
func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.PlayerX == nil && r.PlayerO == nil
}
