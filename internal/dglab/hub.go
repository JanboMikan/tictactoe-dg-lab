package dglab

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Hub 管理所有WebSocket连接和绑定关系
type Hub struct {
	// 已注册的客户端 (ID -> Client)
	clients map[string]*Client

	// 绑定关系: clientId (控制端ID) -> targetId (APP端ID)
	bindings map[string]string

	// 注册新客户端
	register chan *Client

	// 注销客户端
	unregister chan *Client

	// 处理接收到的消息
	broadcast chan []byte

	// 保护并发访问的互斥锁
	mu sync.RWMutex
}

// NewHub 创建一个新的Hub实例
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		bindings:   make(map[string]string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
	}
}

// Run 启动Hub的主循环，处理客户端注册、注销和消息广播
func (h *Hub) Run() {
	// 启动心跳定时器（每60秒发送一次）
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("[DG-LAB Hub] Client registered: %s", client.ID)

			// 发送初始握手消息（分配ID）
			msg := Message{
				Type:     TypeBind,
				ClientID: client.ID,
				TargetID: "",
				Message:  "targetId",
			}
			data, _ := json.Marshal(msg)
			client.Send <- data

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)

				// 查找并通知绑定的伙伴
				partnerID := h.findPartner(client.ID)
				if partnerID != "" {
					h.removeBinding(client.ID)

					// 通知伙伴连接已断开
					if partner, exists := h.clients[partnerID]; exists {
						breakMsg := Message{
							Type:     TypeBreak,
							ClientID: client.ID,
							TargetID: partnerID,
							Message:  "209", // 对方客户端已断开
						}
						data, _ := json.Marshal(breakMsg)
						partner.Send <- data
					}
				}

				log.Printf("[DG-LAB Hub] Client unregistered: %s", client.ID)
			}
			h.mu.Unlock()

		case <-ticker.C:
			// 发送心跳包给所有连接的客户端
			h.sendHeartbeat()
		}
	}
}

// RegisterClient 生成新客户端ID并注册
func (h *Hub) RegisterClient(client *Client) string {
	clientID := uuid.New().String()
	client.ID = clientID
	client.Hub = h
	h.register <- client
	return clientID
}

// UnregisterClient 注销客户端
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// HandleBind 处理绑定请求
// 重要：根据dg-lab.md，APP发送的bind消息中
// - clientId 字段包含从二维码获取的控制端ID
// - targetId 字段包含APP自己的ID
func (h *Hub) HandleBind(senderID string, msg Message) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	clientID := msg.ClientID // 从二维码中获取的控制端ID
	targetID := msg.TargetID // APP自己的ID

	log.Printf("[DG-LAB Hub] Bind request: clientId=%s, targetId=%s, sender=%s",
		clientID, targetID, senderID)

	// 检查两个ID是否都存在
	clientExists := h.clients[clientID] != nil
	targetExists := h.clients[targetID] != nil

	if !clientExists || !targetExists {
		log.Printf("[DG-LAB Hub] Bind failed: client or target not found")
		return h.sendError(senderID, clientID, targetID, "401") // 目标客户端不存在
	}

	// 检查是否已被绑定
	if h.isBound(clientID) || h.isBound(targetID) {
		log.Printf("[DG-LAB Hub] Bind failed: already bound")
		return h.sendError(senderID, clientID, targetID, "400") // 此ID已被绑定
	}

	// 建立绑定关系
	h.bindings[clientID] = targetID
	log.Printf("[DG-LAB Hub] Binding established: %s <-> %s", clientID, targetID)

	// 向双方发送绑定成功消息
	successMsg := Message{
		Type:     TypeBind,
		ClientID: clientID,
		TargetID: targetID,
		Message:  "200", // 成功
	}
	data, _ := json.Marshal(successMsg)

	// 发送给APP
	if target, ok := h.clients[targetID]; ok {
		target.Send <- data
	}

	// 发送给控制端
	if client, ok := h.clients[clientID]; ok {
		client.Send <- data
	}

	return nil
}

// SendCommand 向指定的DG-LAB设备发送控制指令
// clientID: 控制端ID (从Web端传来)
// message: 指令内容 (如 "strength-1+2+50" 或 "pulse-A:[...]")
func (h *Hub) SendCommand(clientID, message string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 查找绑定的APP端
	targetID, exists := h.bindings[clientID]
	if !exists {
		log.Printf("[DG-LAB Hub] SendCommand failed: no binding for clientID=%s", clientID)
		h.sendError(clientID, clientID, "", "402") // 双方未建立绑定关系
		return fmt.Errorf("no binding for clientID=%s", clientID)
	}

	target, ok := h.clients[targetID]
	if !ok {
		log.Printf("[DG-LAB Hub] SendCommand failed: target not connected, targetID=%s", targetID)
		h.sendError(clientID, clientID, targetID, "404") // 未找到收信人
		return fmt.Errorf("target not connected: %s", targetID)
	}

	// 构造并发送消息
	msg := Message{
		Type:     TypeMsg,
		ClientID: clientID,
		TargetID: targetID,
		Message:  message,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[DG-LAB Hub] SendCommand marshal error: %v", err)
		return err
	}

	if len(data) > 1950 {
		log.Printf("[DG-LAB Hub] SendCommand failed: message too long (%d bytes)", len(data))
		h.sendError(clientID, clientID, targetID, "405") // 消息长度超过1950
		return fmt.Errorf("message too long: %d bytes", len(data))
	}

	target.Send <- data
	log.Printf("[DG-LAB Hub] Command sent to %s: %s", targetID, message)
	return nil
}

// sendHeartbeat 向所有连接的客户端发送心跳包
func (h *Hub) sendHeartbeat() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for id, client := range h.clients {
		targetID := ""
		if tid, ok := h.bindings[id]; ok {
			targetID = tid
		} else {
			// 反向查找
			for cid, tid := range h.bindings {
				if tid == id {
					targetID = cid
					break
				}
			}
		}

		msg := Message{
			Type:     TypeHeartbeat,
			ClientID: id,
			TargetID: targetID,
			Message:  "200",
		}
		data, _ := json.Marshal(msg)

		select {
		case client.Send <- data:
		default:
			// 发送通道已满，跳过
		}
	}
}

// sendError 发送错误消息给指定客户端
func (h *Hub) sendError(senderID, clientID, targetID, errorCode string) error {
	msg := Message{
		Type:     TypeBind, // 绑定相关错误使用bind类型
		ClientID: clientID,
		TargetID: targetID,
		Message:  errorCode,
	}
	data, _ := json.Marshal(msg)

	if client, ok := h.clients[senderID]; ok {
		client.Send <- data
	}
	return nil
}

// findPartner 查找客户端的绑定伙伴ID
func (h *Hub) findPartner(id string) string {
	// 作为clientId查找
	if targetID, ok := h.bindings[id]; ok {
		return targetID
	}
	// 作为targetId反向查找
	for clientID, targetID := range h.bindings {
		if targetID == id {
			return clientID
		}
	}
	return ""
}

// removeBinding 移除绑定关系
func (h *Hub) removeBinding(id string) {
	// 作为clientId删除
	if _, ok := h.bindings[id]; ok {
		delete(h.bindings, id)
		return
	}
	// 作为targetId反向删除
	for clientID, targetID := range h.bindings {
		if targetID == id {
			delete(h.bindings, clientID)
			return
		}
	}
}

// isBound 检查ID是否已被绑定
func (h *Hub) isBound(id string) bool {
	// 检查是否作为clientId存在
	if _, ok := h.bindings[id]; ok {
		return true
	}
	// 检查是否作为targetId存在
	for _, targetID := range h.bindings {
		if targetID == id {
			return true
		}
	}
	return false
}

// IsDeviceConnected 检查指定的客户端ID是否有绑定的设备且设备在线
// clientID: 控制端ID (即玩家的 DGLabClientID)
// 返回 true 表示设备已连接并在线
func (h *Hub) IsDeviceConnected(clientID string) bool {
	if clientID == "" {
		return false
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// 检查是否有绑定关系
	targetID, exists := h.bindings[clientID]
	if !exists {
		return false
	}

	// 检查APP端是否在线
	_, online := h.clients[targetID]
	return online
}
