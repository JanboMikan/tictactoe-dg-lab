package dglab

import (
	"github.com/gorilla/websocket"
)

// MessageType 定义消息类型常量
type MessageType string

const (
	TypeHeartbeat MessageType = "heartbeat" // 心跳包
	TypeBind      MessageType = "bind"      // 绑定关系
	TypeMsg       MessageType = "msg"       // 业务指令
	TypeBreak     MessageType = "break"     // 断开连接
	TypeError     MessageType = "error"     // 错误信息
)

// Message 定义DG-LAB通信协议的基础消息结构
type Message struct {
	Type     MessageType `json:"type"`               // 消息类型
	ClientID string      `json:"clientId"`           // 控制端ID
	TargetID string      `json:"targetId"`           // APP端ID
	Message  string      `json:"message"`            // 具体指令或数据内容
}

// Client 代表一个WebSocket连接的客户端
type Client struct {
	ID   string          // 客户端唯一标识符 (UUID)
	Conn *websocket.Conn // WebSocket连接
	Send chan []byte     // 用于发送消息的缓冲通道
	Hub  *Hub            // 引用所属的Hub
}

// Binding 代表控制端和APP端的绑定关系
type Binding struct {
	ClientID string // 控制端ID (Web端生成的UUID)
	TargetID string // APP端ID (服务器分配的UUID)
}
