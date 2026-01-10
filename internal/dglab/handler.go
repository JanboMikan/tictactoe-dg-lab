package dglab

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有来源（生产环境应该限制）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket 处理WebSocket连接请求
func HandleWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 升级HTTP连接为WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[DG-LAB Handler] WebSocket upgrade error: %v", err)
			return
		}

		// 创建新客户端
		client := &Client{
			Conn: conn,
			Send: make(chan []byte, 256),
			Hub:  hub,
		}

		// 注册客户端到Hub（Hub会生成ID并发送初始握手消息）
		clientID := hub.RegisterClient(client)

		log.Printf("[DG-LAB Handler] New WebSocket connection established: %s", clientID)

		// 启动读写协程
		go client.WritePump()
		go client.ReadPump()
	}
}
