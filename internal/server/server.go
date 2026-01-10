package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anon/tictactoe-dg-lab/internal/config"
	"github.com/anon/tictactoe-dg-lab/internal/dglab"
	"github.com/anon/tictactoe-dg-lab/internal/game"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server HTTP 服务器
type Server struct {
	router    *gin.Engine
	config    *config.Config
	dglabHub  *dglab.Hub // DG-LAB WebSocket Hub
	gameHub   *game.Hub  // Game WebSocket Hub
}

// New 创建新的服务器实例
func New(cfg *config.Config) *Server {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// 配置 CORS 中间件
	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// 创建DG-LAB Hub并启动
	dglabHub := dglab.NewHub()
	go dglabHub.Run()
	log.Println("[Server] DG-LAB Hub started")

	// 创建Game Hub并启动（注入 DGLab Hub 和 Config）
	gameHub := game.NewHub(dglabHub, cfg)
	go gameHub.Run()
	log.Println("[Server] Game Hub started")

	// 设置DG-LAB设备绑定成功的回调
	// 当设备绑定成功时，通知游戏Hub广播状态更新
	dglabHub.OnBindSuccess = func(clientID string) {
		log.Printf("[Server] Device bind success callback for clientID: %s", clientID)
		gameHub.NotifyDeviceConnected(clientID)
	}

	// 设置DG-LAB设备断开连接的回调
	// 当设备断开连接时，通知游戏Hub广播状态更新
	dglabHub.OnDeviceDisconnect = func(clientID string) {
		log.Printf("[Server] Device disconnect callback for clientID: %s", clientID)
		gameHub.NotifyDeviceDisconnected(clientID)
	}

	server := &Server{
		router:   router,
		config:   cfg,
		dglabHub: dglabHub,
		gameHub:  gameHub,
	}

	// 注册路由
	server.registerRoutes()

	return server
}

// registerRoutes 注册路由
func (s *Server) registerRoutes() {
	// 健康检查路由
	s.router.GET("/ping", s.pingHandler)

	// DG-LAB WebSocket 路由
	// 支持两种路径：/ws/dglab 和 /ws/dglab/:clientId
	// APP 扫描二维码后会连接到 /ws/dglab/:clientId 路径
	s.router.GET("/ws/dglab", dglab.HandleWebSocket(s.dglabHub))
	s.router.GET("/ws/dglab/:clientId", dglab.HandleWebSocket(s.dglabHub))

	// Game WebSocket 路由
	s.router.GET("/ws/game", game.HandleWebSocket(s.gameHub))
}

// pingHandler 健康检查处理器
func (s *Server) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"time":    time.Now().Unix(),
	})
}

// Run 启动服务器
func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	log.Printf("服务器启动在: %s", addr)

	return s.router.Run(addr)
}

// GetRouter 获取路由器（用于测试）
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// GetDGLabHub 获取DG-LAB Hub（用于游戏逻辑模块调用）
func (s *Server) GetDGLabHub() *dglab.Hub {
	return s.dglabHub
}

// GetGameHub 获取Game Hub（用于测试和其他模块调用）
func (s *Server) GetGameHub() *game.Hub {
	return s.gameHub
}
