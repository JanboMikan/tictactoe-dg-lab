package game

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写超时
	writeWait = 10 * time.Second

	// Pong 超时
	pongWait = 60 * time.Second

	// Ping 间隔
	pingPeriod = (pongWait * 9) / 10

	// 消息缓冲区大小
	maxMessageSize = 8192
)

// ReadPump 从 WebSocket 读取消息
func (p *Player) ReadPump(hub *Hub) {
	defer func() {
		hub.unregister <- p
		if p.Conn != nil {
			p.Conn.Close()
		}
	}()

	p.Conn.SetReadDeadline(time.Now().Add(pongWait))
	p.Conn.SetPongHandler(func(string) error {
		p.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := p.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Player %s] WebSocket error: %v", p.Name, err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("[Player %s] Failed to unmarshal message: %v", p.Name, err)
			p.SendError("Invalid message format")
			continue
		}

		// 处理消息
		hub.handleMessage <- &PlayerMessage{
			Player:  p,
			Message: &msg,
		}
	}
}

// WritePump 向 WebSocket 写入消息
func (p *Player) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if p.Conn != nil {
			p.Conn.Close()
		}
	}()

	for {
		select {
		case message, ok := <-p.Send:
			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了通道
				p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := p.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将队列中的其他消息也一起发送
			n := len(p.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-p.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendError 向玩家发送错误消息
func (p *Player) SendError(errorMsg string) {
	msg := &Message{
		Type:    TypeError,
		Message: errorMsg,
	}
	data, _ := json.Marshal(msg)
	select {
	case p.Send <- data:
	default:
		log.Printf("[Player %s] Failed to send error (channel full)", p.Name)
	}
}

// UpdateDGLabID 更新玩家的 DG-LAB 客户端 ID
func (p *Player) UpdateDGLabID(clientID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.DGLabClientID = clientID
	log.Printf("[Player %s] DG-LAB Client ID updated: %s", p.Name, clientID)
}

// UpdateConfig 更新玩家配置
func (p *Player) UpdateConfig(config *PlayerConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Config = config
	log.Printf("[Player %s] Config updated: %+v", p.Name, config)
	return nil
}

// GetDGLabID 获取玩家的 DG-LAB 客户端 ID
func (p *Player) GetDGLabID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.DGLabClientID
}
