package main

import (
	"log"
	"os"

	"github.com/anon/tictactoe-dg-lab/internal/config"
	"github.com/anon/tictactoe-dg-lab/internal/server"
)

func main() {
	// 加载配置文件
	configPath := "config.yml"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}

	if err := config.Load(configPath); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 创建并启动服务器
	srv := server.New(config.GetConfig())
	if err := srv.Run(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
