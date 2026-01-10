package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anon/tictactoe-dg-lab/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server HTTP 服务器
type Server struct {
	router *gin.Engine
	config *config.Config
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

	server := &Server{
		router: router,
		config: cfg,
	}

	// 注册路由
	server.registerRoutes()

	return server
}

// registerRoutes 注册路由
func (s *Server) registerRoutes() {
	// 健康检查路由
	s.router.GET("/ping", s.pingHandler)

	// TODO: 后续添加 WebSocket 路由
	// s.router.GET("/ws/game", s.gameWSHandler)
	// s.router.GET("/ws/dglab", s.dglabWSHandler)
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
