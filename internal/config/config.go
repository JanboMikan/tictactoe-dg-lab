package config

import (
	"log"

	"github.com/spf13/viper"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// GameConfig 游戏配置
type GameConfig struct {
	PunishmentDurationMin float64 `mapstructure:"punishment_duration_min"`
	PunishmentDurationMax float64 `mapstructure:"punishment_duration_max"`
	MoveDuration          float64 `mapstructure:"move_duration"`
	DrawDuration          float64 `mapstructure:"draw_duration"`
}

// WaveformsConfig 波形配置
type WaveformsConfig struct {
	Default string `mapstructure:"default"`
	Pulse   string `mapstructure:"pulse"`
}

// Config 全局配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Game      GameConfig      `mapstructure:"game"`
	Waveforms WaveformsConfig `mapstructure:"waveforms"`
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// Load 加载配置文件
func Load(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v", err)
		return err
	}

	// 解析配置
	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		log.Printf("解析配置文件失败: %v", err)
		return err
	}

	log.Printf("配置加载成功: %+v", GlobalConfig)
	return nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return GlobalConfig
}
