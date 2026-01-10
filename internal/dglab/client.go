package dglab

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写入超时时间
	writeWait = 10 * time.Second

	// Pong等待时间（应大于pongWait）
	pongWait = 60 * time.Second

	// Ping发送间隔（必须小于pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 8192
)

// WritePump 从hub向WebSocket连接写入消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub关闭了通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将队列中的其他消息也一并发送
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump 从WebSocket连接读取消息并处理
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	c.Conn.SetReadLimit(maxMessageSize)

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[DG-LAB Client] Error: %v", err)
			}
			break
		}

		// 解析JSON消息
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("[DG-LAB Client] Invalid JSON: %v", err)
			// 发送错误响应
			errorMsg := Message{
				Type:     TypeError,
				ClientID: "",
				TargetID: "",
				Message:  "403", // 非标准JSON格式
			}
			data, _ := json.Marshal(errorMsg)
			c.Send <- data
			continue
		}

		log.Printf("[DG-LAB Client %s] Received: type=%s, msg=%s", c.ID, msg.Type, msg.Message)

		// 处理不同类型的消息
		switch msg.Type {
		case TypeBind:
			// 处理绑定请求
			c.Hub.HandleBind(c.ID, msg)

		case TypeMsg:
			// 转发业务消息
			c.handleMessage(msg)

		case TypeHeartbeat:
			// 心跳响应（客户端可能发送心跳，我们简单记录）
			log.Printf("[DG-LAB Client %s] Heartbeat received", c.ID)

		default:
			log.Printf("[DG-LAB Client %s] Unknown message type: %s", c.ID, msg.Type)
		}
	}
}

// handleMessage 处理业务消息（msg类型）
func (c *Client) handleMessage(msg Message) {
	c.Hub.mu.RLock()
	defer c.Hub.mu.RUnlock()

	// 验证发送者权限（确保发送者是消息中声明的clientId或targetId之一）
	if msg.ClientID != c.ID && msg.TargetID != c.ID {
		log.Printf("[DG-LAB Client] Unauthorized message from %s", c.ID)
		errorMsg := Message{
			Type:     TypeError,
			ClientID: "",
			TargetID: "",
			Message:  "404", // 未找到收信人
		}
		data, _ := json.Marshal(errorMsg)
		c.Send <- data
		return
	}

	// 确定接收方ID
	var recipientID string
	if msg.ClientID == c.ID {
		recipientID = msg.TargetID
	} else {
		recipientID = msg.ClientID
	}

	// 查找接收方
	recipient, exists := c.Hub.clients[recipientID]
	if !exists {
		log.Printf("[DG-LAB Client] Recipient not found: %s", recipientID)
		errorMsg := Message{
			Type:     TypeError,
			ClientID: msg.ClientID,
			TargetID: msg.TargetID,
			Message:  "404", // 未找到收信人
		}
		data, _ := json.Marshal(errorMsg)
		c.Send <- data
		return
	}

	// 转发消息
	data, _ := json.Marshal(msg)
	if len(data) > 1950 {
		log.Printf("[DG-LAB Client] Message too long: %d bytes", len(data))
		errorMsg := Message{
			Type:     TypeError,
			ClientID: msg.ClientID,
			TargetID: msg.TargetID,
			Message:  "405", // 消息长度超过1950
		}
		errData, _ := json.Marshal(errorMsg)
		c.Send <- errData
		return
	}

	recipient.Send <- data
	log.Printf("[DG-LAB Client] Message forwarded from %s to %s", c.ID, recipientID)
}
