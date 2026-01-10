package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// 创建临时配置文件
	tmpConfig := `server:
  port: 8080
  host: "0.0.0.0"

game:
  punishment_duration_min: 1.0
  punishment_duration_max: 10.0
  move_duration: 0.5
  draw_duration: 1.0

waveforms:
  default: "0A0A0A0A0A0A0A0A"
  pulse: "00000000FFFFFFFF00000000"
`
	tmpFile, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(tmpConfig); err != nil {
		t.Fatalf("写入临时配置文件失败: %v", err)
	}
	tmpFile.Close()

	// 测试加载配置
	err = Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证配置
	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("获取配置失败，返回 nil")
	}

	// 检查服务器配置
	if cfg.Server.Port != 8080 {
		t.Errorf("期望端口为 8080，实际为 %d", cfg.Server.Port)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("期望主机为 0.0.0.0，实际为 %s", cfg.Server.Host)
	}

	// 检查游戏配置
	if cfg.Game.MoveDuration != 0.5 {
		t.Errorf("期望 move_duration 为 0.5，实际为 %f", cfg.Game.MoveDuration)
	}

	// 检查波形配置
	if cfg.Waveforms.Default != "0A0A0A0A0A0A0A0A" {
		t.Errorf("期望默认波形为 0A0A0A0A0A0A0A0A，实际为 %s", cfg.Waveforms.Default)
	}
}

func TestLoadInvalidPath(t *testing.T) {
	err := Load("/invalid/path/config.yml")
	if err == nil {
		t.Error("期望加载不存在的文件时返回错误，但没有返回")
	}
}
