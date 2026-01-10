package game

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（生产环境应该限制）
	},
}

// HandleWebSocket 处理游戏 WebSocket 连接
func HandleWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 升级 HTTP 连接为 WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[Game WS] Failed to upgrade connection: %v", err)
			return
		}

		// 创建新玩家
		player := &Player{
			Name: "Player",  // 默认名字，会在 join_room 时更新
			Conn: conn,
			Send: make(chan []byte, 256),
		}

		// 注册玩家到 Hub
		hub.register <- player

		log.Printf("[Game WS] New player connected from %s", conn.RemoteAddr())

		// 启动读写循环
		go player.WritePump()
		go player.ReadPump(hub)
	}
}
