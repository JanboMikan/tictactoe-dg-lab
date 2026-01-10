package game

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestRoomCreation 测试房间创建
func TestRoomCreation(t *testing.T) {
	manager := NewRoomManager()
	room := manager.CreateRoom()

	if room.ID == "" {
		t.Error("Room ID should not be empty")
	}

	if room.Turn != 1 {
		t.Errorf("Expected Turn to be 1, got %d", room.Turn)
	}

	if room.GameOver {
		t.Error("New room should not be game over")
	}

	if room.IsFull() {
		t.Error("New room should not be full")
	}
}

// TestJoinRoom 测试加入房间
func TestJoinRoom(t *testing.T) {
	manager := NewRoomManager()
	room := manager.CreateRoom()

	player1 := &Player{
		Name: "Alice",
		Send: make(chan []byte, 256),
	}

	player2 := &Player{
		Name: "Bob",
		Send: make(chan []byte, 256),
	}

	// 第一个玩家加入
	err := manager.JoinRoom(room.ID, player1)
	if err != nil {
		t.Fatalf("Failed to join room: %v", err)
	}

	if player1.Symbol != 1 {
		t.Errorf("Expected player1 symbol to be 1, got %d", player1.Symbol)
	}

	if room.PlayerX != player1 {
		t.Error("Player1 should be PlayerX")
	}

	// 第二个玩家加入
	err = manager.JoinRoom(room.ID, player2)
	if err != nil {
		t.Fatalf("Failed to join room: %v", err)
	}

	if player2.Symbol != 2 {
		t.Errorf("Expected player2 symbol to be 2, got %d", player2.Symbol)
	}

	if room.PlayerO != player2 {
		t.Error("Player2 should be PlayerO")
	}

	if !room.IsFull() {
		t.Error("Room should be full after two players join")
	}

	// 尝试第三个玩家加入（应该失败）
	player3 := &Player{
		Name: "Charlie",
		Send: make(chan []byte, 256),
	}

	err = manager.JoinRoom(room.ID, player3)
	if err != ErrRoomFull {
		t.Errorf("Expected ErrRoomFull, got %v", err)
	}
}

// TestMakeMove 测试落子逻辑
func TestMakeMove(t *testing.T) {
	room := &Room{
		ID:    "test",
		Board: [9]int{},
		Turn:  1,
	}

	player1 := &Player{
		Name:   "Alice",
		Symbol: 1,
	}

	player2 := &Player{
		Name:   "Bob",
		Symbol: 2,
	}

	room.PlayerX = player1
	room.PlayerO = player2

	// Player1 落子
	err := room.MakeMove(player1, 0)
	if err != nil {
		t.Fatalf("MakeMove failed: %v", err)
	}

	if room.Board[0] != 1 {
		t.Error("Position 0 should be occupied by player 1")
	}

	if room.Turn != 2 {
		t.Errorf("Turn should switch to 2, got %d", room.Turn)
	}

	// Player2 尝试落在已占用的位置（应该失败）
	err = room.MakeMove(player2, 0)
	if err != ErrPositionOccupied {
		t.Errorf("Expected ErrPositionOccupied, got %v", err)
	}

	// Player1 尝试连续落子（应该失败）
	err = room.MakeMove(player1, 1)
	if err != ErrNotYourTurn {
		t.Errorf("Expected ErrNotYourTurn, got %v", err)
	}

	// Player2 正常落子
	err = room.MakeMove(player2, 1)
	if err != nil {
		t.Fatalf("MakeMove failed: %v", err)
	}

	if room.Board[1] != 2 {
		t.Error("Position 1 should be occupied by player 2")
	}
}

// TestCheckWin 测试胜负判定
func TestCheckWin(t *testing.T) {
	tests := []struct {
		name        string
		board       [9]int
		expectedWin bool
		winner      int
		line        []int
	}{
		{
			name:        "Horizontal win - top row",
			board:       [9]int{1, 1, 1, 2, 2, 0, 0, 0, 0},
			expectedWin: true,
			winner:      1,
			line:        []int{0, 1, 2},
		},
		{
			name:        "Vertical win - left column",
			board:       [9]int{1, 2, 0, 1, 2, 0, 1, 0, 0},
			expectedWin: true,
			winner:      1,
			line:        []int{0, 3, 6},
		},
		{
			name:        "Diagonal win - top-left to bottom-right",
			board:       [9]int{1, 2, 0, 0, 1, 2, 0, 0, 1},
			expectedWin: true,
			winner:      1,
			line:        []int{0, 4, 8},
		},
		{
			name:        "Draw - board full, no winner",
			board:       [9]int{1, 2, 1, 2, 1, 2, 2, 1, 2},
			expectedWin: true,
			winner:      0,
			line:        nil,
		},
		{
			name:        "Game not over",
			board:       [9]int{1, 2, 0, 0, 0, 0, 0, 0, 0},
			expectedWin: false,
			winner:      0,
			line:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := &Room{
				ID:    "test",
				Board: tt.board,
				Turn:  1,
			}

			room.checkWin()

			if room.GameOver != tt.expectedWin {
				t.Errorf("Expected GameOver=%v, got %v", tt.expectedWin, room.GameOver)
			}

			if room.Winner != tt.winner {
				t.Errorf("Expected Winner=%d, got %d", tt.winner, room.Winner)
			}

			if tt.line != nil {
				if len(room.WinningLine) != len(tt.line) {
					t.Errorf("Expected WinningLine length %d, got %d", len(tt.line), len(room.WinningLine))
				} else {
					for i, pos := range tt.line {
						if room.WinningLine[i] != pos {
							t.Errorf("WinningLine[%d]: expected %d, got %d", i, pos, room.WinningLine[i])
						}
					}
				}
			}
		})
	}
}

// TestPlayerConfig 测试玩家配置验证
func TestPlayerConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *PlayerConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &PlayerConfig{
				SafeMin:      10,
				SafeMax:      60,
				MoveStrength: 20,
				DrawStrength: 40,
			},
			wantErr: false,
		},
		{
			name: "Invalid - SafeMin >= SafeMax",
			config: &PlayerConfig{
				SafeMin:      60,
				SafeMax:      60,
				MoveStrength: 60,
				DrawStrength: 60,
			},
			wantErr: true,
		},
		{
			name: "Invalid - MoveStrength out of range",
			config: &PlayerConfig{
				SafeMin:      10,
				SafeMax:      60,
				MoveStrength: 5, // less than SafeMin
				DrawStrength: 40,
			},
			wantErr: true,
		},
		{
			name: "Invalid - DrawStrength out of range",
			config: &PlayerConfig{
				SafeMin:      10,
				SafeMax:      60,
				MoveStrength: 20,
				DrawStrength: 70, // greater than SafeMax
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRoomManager 测试房间管理器
func TestRoomManager(t *testing.T) {
	manager := NewRoomManager()

	// 测试创建房间
	room1 := manager.CreateRoom()
	room2 := manager.CreateRoom()

	if room1.ID == room2.ID {
		t.Error("Room IDs should be unique")
	}

	if manager.GetRoomCount() != 2 {
		t.Errorf("Expected 2 rooms, got %d", manager.GetRoomCount())
	}

	// 测试获取房间
	fetchedRoom, err := manager.GetRoom(room1.ID)
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	if fetchedRoom.ID != room1.ID {
		t.Error("Fetched room ID doesn't match")
	}

	// 测试获取不存在的房间
	_, err = manager.GetRoom("nonexistent")
	if err != ErrRoomNotFound {
		t.Errorf("Expected ErrRoomNotFound, got %v", err)
	}

	// 测试删除房间
	manager.DeleteRoom(room1.ID)
	if manager.GetRoomCount() != 1 {
		t.Errorf("Expected 1 room after deletion, got %d", manager.GetRoomCount())
	}

	_, err = manager.GetRoom(room1.ID)
	if err != ErrRoomNotFound {
		t.Error("Deleted room should not be found")
	}
}

// TestCleanEmptyRooms 测试清理空房间
func TestCleanEmptyRooms(t *testing.T) {
	manager := NewRoomManager()

	// 创建一个旧房间（模拟超过10分钟）
	oldRoom := manager.CreateRoom()
	oldRoom.CreatedAt = time.Now().Add(-11 * time.Minute)

	// 创建一个新房间
	newRoom := manager.CreateRoom()

	if manager.GetRoomCount() != 2 {
		t.Fatalf("Expected 2 rooms, got %d", manager.GetRoomCount())
	}

	// 清理空房间
	manager.CleanEmptyRooms()

	// 旧房间应该被删除
	if manager.GetRoomCount() != 1 {
		t.Errorf("Expected 1 room after cleanup, got %d", manager.GetRoomCount())
	}

	// 新房间应该还在
	_, err := manager.GetRoom(newRoom.ID)
	if err != nil {
		t.Error("New room should still exist after cleanup")
	}

	// 旧房间应该不在了
	_, err = manager.GetRoom(oldRoom.ID)
	if err != ErrRoomNotFound {
		t.Error("Old room should be deleted")
	}
}

// MockConn 模拟 WebSocket 连接（用于测试）
type MockConn struct {
	websocket.Conn
}

func (m *MockConn) WriteMessage(messageType int, data []byte) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}
