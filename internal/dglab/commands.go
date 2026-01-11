package dglab

import (
	"fmt"
	"log"
	"time"
)

// Channel 定义通道类型
type Channel int

const (
	ChannelA Channel = 1 // A通道
	ChannelB Channel = 2 // B通道
)

// StrengthMode 定义强度控制模式
type StrengthMode int

const (
	ModeDecrease StrengthMode = 0 // 减少强度
	ModeIncrease StrengthMode = 1 // 增加强度
	ModeSet      StrengthMode = 2 // 设置为指定数值
)

// SendStrength 发送强度控制指令
// channel: 通道 (1=A, 2=B)
// mode: 模式 (0=减, 1=加, 2=设置)
// value: 数值 (0-200)
func (h *Hub) SendStrength(clientID string, channel Channel, mode StrengthMode, value int) error {
	// 验证参数
	if channel != ChannelA && channel != ChannelB {
		return fmt.Errorf("invalid channel: %d", channel)
	}
	if value < 0 || value > 200 {
		return fmt.Errorf("invalid strength value: %d (must be 0-200)", value)
	}
	if mode < ModeDecrease || mode > ModeSet {
		return fmt.Errorf("invalid mode: %d", mode)
	}

	// 构造指令: strength-{通道}+{模式}+{数值}
	message := fmt.Sprintf("strength-%d+%d+%d", channel, mode, value)
	log.Printf("[DG-LAB Commands] Sending strength command: %s", message)

	return h.SendCommand(clientID, message)
}

// SendPulse 发送波形数据
// channel: 通道 ("A" 或 "B")
// hexData: HEX数组，每个元素为8字节HEX码，代表100ms数据
// 注意：数组最大长度100 (10秒)，APP队列最大缓存500 (50秒)
// 如果数据量超过单次消息限制（1950字节），会自动分批发送
func (h *Hub) SendPulse(clientID string, channel string, hexData []string) error {
	// 验证通道
	if channel != "A" && channel != "B" {
		return fmt.Errorf("invalid channel: %s (must be 'A' or 'B')", channel)
	}

	// 验证数组长度
	if len(hexData) == 0 {
		return fmt.Errorf("hexData is empty")
	}
	if len(hexData) > 100 {
		return fmt.Errorf("hexData too long: %d (max 100)", len(hexData))
	}

	// 计算单批最多能发送多少个波形
	// 每个波形约27字节（含引号和逗号），JSON固定部分约200字节
	// 为保险起见，限制为30个波形/批（约1000字节，远小于1950字节限制）
	const batchSize = 30

	// 如果数据量小于批次大小，直接发送
	if len(hexData) <= batchSize {
		return h.sendPulseBatch(clientID, channel, hexData)
	}

	// 分批发送
	log.Printf("[DG-LAB Commands] Sending pulse in batches: channel=%s, totalData=%d, batchSize=%d",
		channel, len(hexData), batchSize)

	for i := 0; i < len(hexData); i += batchSize {
		end := i + batchSize
		if end > len(hexData) {
			end = len(hexData)
		}

		batch := hexData[i:end]
		err := h.sendPulseBatch(clientID, channel, batch)
		if err != nil {
			log.Printf("[DG-LAB Commands] Failed to send batch %d-%d: %v", i, end, err)
			return fmt.Errorf("failed to send batch %d-%d: %w", i, end, err)
		}

		log.Printf("[DG-LAB Commands] Sent batch %d-%d successfully", i, end)

		// 批次之间添加短暂延迟（100ms），避免消息拥堵
		if end < len(hexData) {
			time.Sleep(100 * time.Millisecond)
		}
	}

	log.Printf("[DG-LAB Commands] All pulse batches sent successfully: channel=%s, totalData=%d",
		channel, len(hexData))
	return nil
}

// sendPulseBatch 发送单批波形数据（内部使用）
func (h *Hub) sendPulseBatch(clientID string, channel string, hexData []string) error {
	// 构造JSON数组字符串
	hexArrayStr := "["
	for i, hex := range hexData {
		if i > 0 {
			hexArrayStr += ","
		}
		hexArrayStr += fmt.Sprintf("\"%s\"", hex)
	}
	hexArrayStr += "]"

	// 构造指令: pulse-{通道}:{HEX数组}
	message := fmt.Sprintf("pulse-%s:%s", channel, hexArrayStr)
	log.Printf("[DG-LAB Commands] Sending pulse batch: channel=%s, dataLen=%d", channel, len(hexData))

	return h.SendCommand(clientID, message)
}

// ClearQueue 清空指定通道的波形队列
// channel: 通道 (1=A, 2=B)
func (h *Hub) ClearQueue(clientID string, channel Channel) error {
	// 验证参数
	if channel != ChannelA && channel != ChannelB {
		return fmt.Errorf("invalid channel: %d", channel)
	}

	// 构造指令: clear-{通道}
	message := fmt.Sprintf("clear-%d", channel)
	log.Printf("[DG-LAB Commands] Sending clear command: %s", message)

	return h.SendCommand(clientID, message)
}

// SendStrengthQuick 快捷方法：增加或减少强度（默认变化量为5）
func (h *Hub) SendStrengthQuick(clientID string, channel Channel, increase bool) error {
	mode := ModeDecrease
	if increase {
		mode = ModeIncrease
	}
	return h.SendStrength(clientID, channel, mode, 5)
}

// SendStrengthSet 快捷方法：设置强度到指定值
func (h *Hub) SendStrengthSet(clientID string, channel Channel, value int) error {
	return h.SendStrength(clientID, channel, ModeSet, value)
}

// SendStrengthZero 快捷方法：将强度归零
func (h *Hub) SendStrengthZero(clientID string, channel Channel) error {
	return h.SendStrength(clientID, channel, ModeSet, 0)
}
