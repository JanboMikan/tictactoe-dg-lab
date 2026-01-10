package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anon/tictactoe-dg-lab/internal/config"
)

func TestPingHandler(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
	}

	// 创建服务器
	server := New(cfg)

	// 创建测试请求
	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}

	// 记录响应
	rr := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(rr, req)

	// 检查状态码
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("期望状态码 %d，实际为 %d", http.StatusOK, status)
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 检查响应内容
	if msg, ok := response["message"].(string); !ok || msg != "pong" {
		t.Errorf("期望消息为 'pong'，实际为 %v", response["message"])
	}

	if _, ok := response["time"].(float64); !ok {
		t.Error("响应中应包含 time 字段")
	}
}

func TestCORS(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
	}

	// 创建服务器
	server := New(cfg)

	// 创建 OPTIONS 预检请求
	req, err := http.NewRequest("OPTIONS", "/ping", nil)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	// 记录响应
	rr := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(rr, req)

	// 检查 CORS 头
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin == "" {
		t.Error("响应中应包含 Access-Control-Allow-Origin 头")
	}

	if methods := rr.Header().Get("Access-Control-Allow-Methods"); methods == "" {
		t.Error("响应中应包含 Access-Control-Allow-Methods 头")
	}
}
