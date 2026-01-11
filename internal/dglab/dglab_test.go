package dglab

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewHub 测试创建新Hub
func TestNewHub(t *testing.T) {
	hub := NewHub()
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.bindings)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
}

// TestHubRegisterClient 测试客户端注册
func TestHubRegisterClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// 创建模拟客户端
	client := &Client{
		Conn: nil, // 测试中不需要真实连接
		Send: make(chan []byte, 256),
	}

	// 注册客户端
	clientID := hub.RegisterClient(client)

	// 等待注册完成
	time.Sleep(100 * time.Millisecond)

	// 验证客户端ID已生成
	assert.NotEmpty(t, clientID)
	assert.Equal(t, clientID, client.ID)

	// 验证客户端在Hub中
	hub.mu.RLock()
	_, exists := hub.clients[clientID]
	hub.mu.RUnlock()
	assert.True(t, exists)

	// 验证收到了初始握手消息
	select {
	case msg := <-client.Send:
		var message Message
		err := json.Unmarshal(msg, &message)
		assert.NoError(t, err)
		assert.Equal(t, TypeBind, message.Type)
		assert.Equal(t, clientID, message.ClientID)
		assert.Equal(t, "targetId", message.Message)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial handshake message")
	}
}

// TestHubHandleBind 测试绑定逻辑
func TestHubHandleBind(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// 创建两个模拟客户端（控制端和APP端）
	controlClient := &Client{
		Send: make(chan []byte, 256),
	}
	appClient := &Client{
		Send: make(chan []byte, 256),
	}

	// 注册两个客户端
	controlID := hub.RegisterClient(controlClient)
	appID := hub.RegisterClient(appClient)

	// 等待注册完成
	time.Sleep(100 * time.Millisecond)

	// 清空初始握手消息
	<-controlClient.Send
	<-appClient.Send

	// APP发送绑定请求
	bindMsg := Message{
		Type:     TypeBind,
		ClientID: controlID, // 从二维码获取的控制端ID
		TargetID: appID,     // APP自己的ID
		Message:  "DGLAB",
	}

	err := hub.HandleBind(appID, bindMsg)
	assert.NoError(t, err)

	// 等待处理完成
	time.Sleep(100 * time.Millisecond)

	// 验证绑定关系已建立
	hub.mu.RLock()
	targetID, exists := hub.bindings[controlID]
	hub.mu.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, appID, targetID)

	// 验证两个客户端都收到了绑定成功消息
	select {
	case msg := <-controlClient.Send:
		var message Message
		err := json.Unmarshal(msg, &message)
		assert.NoError(t, err)
		assert.Equal(t, TypeBind, message.Type)
		assert.Equal(t, "200", message.Message)
	case <-time.After(1 * time.Second):
		t.Fatal("Control client didn't receive bind success message")
	}

	select {
	case msg := <-appClient.Send:
		var message Message
		err := json.Unmarshal(msg, &message)
		assert.NoError(t, err)
		assert.Equal(t, TypeBind, message.Type)
		assert.Equal(t, "200", message.Message)
	case <-time.After(1 * time.Second):
		t.Fatal("App client didn't receive bind success message")
	}
}

// TestSendCommand 测试发送指令
func TestSendCommand(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// 创建并注册两个客户端
	controlClient := &Client{
		Send: make(chan []byte, 256),
	}
	appClient := &Client{
		Send: make(chan []byte, 256),
	}

	controlID := hub.RegisterClient(controlClient)
	appID := hub.RegisterClient(appClient)

	time.Sleep(100 * time.Millisecond)

	// 清空初始消息
	<-controlClient.Send
	<-appClient.Send

	// 建立绑定
	hub.mu.Lock()
	hub.bindings[controlID] = appID
	hub.mu.Unlock()

	// 发送指令
	testMessage := "strength-1+2+50"
	err := hub.SendCommand(controlID, testMessage)
	assert.NoError(t, err)

	// 验证APP收到了指令
	select {
	case msg := <-appClient.Send:
		var message Message
		err := json.Unmarshal(msg, &message)
		assert.NoError(t, err)
		assert.Equal(t, TypeMsg, message.Type)
		assert.Equal(t, controlID, message.ClientID)
		assert.Equal(t, appID, message.TargetID)
		assert.Equal(t, testMessage, message.Message)
	case <-time.After(1 * time.Second):
		t.Fatal("App client didn't receive command")
	}
}

// TestSendStrength 测试强度控制指令
func TestSendStrength(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// 创建并注册客户端
	controlClient := &Client{Send: make(chan []byte, 256)}
	appClient := &Client{Send: make(chan []byte, 256)}

	controlID := hub.RegisterClient(controlClient)
	appID := hub.RegisterClient(appClient)
	time.Sleep(100 * time.Millisecond)

	// 清空初始消息
	<-controlClient.Send
	<-appClient.Send

	// 建立绑定
	hub.mu.Lock()
	hub.bindings[controlID] = appID
	hub.mu.Unlock()

	// 测试发送强度指令
	err := hub.SendStrength(controlID, ChannelA, ModeSet, 50)
	assert.NoError(t, err)

	// 验证指令格式
	select {
	case msg := <-appClient.Send:
		var message Message
		err := json.Unmarshal(msg, &message)
		assert.NoError(t, err)
		assert.Equal(t, "strength-1+2+50", message.Message)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for strength command")
	}
}

// TestSendPulse 测试波形下发指令
func TestSendPulse(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// 创建并注册客户端
	controlClient := &Client{Send: make(chan []byte, 256)}
	appClient := &Client{Send: make(chan []byte, 256)}

	controlID := hub.RegisterClient(controlClient)
	appID := hub.RegisterClient(appClient)
	time.Sleep(100 * time.Millisecond)

	// 清空初始消息
	<-controlClient.Send
	<-appClient.Send

	// 建立绑定
	hub.mu.Lock()
	hub.bindings[controlID] = appID
	hub.mu.Unlock()

	// 测试发送波形数据
	hexData := []string{"0A0A0A0A00000000", "0A0A0A0A0A0A0A0A"}
	err := hub.SendPulse(controlID, "A", hexData)
	assert.NoError(t, err)

	// 验证指令格式
	select {
	case msg := <-appClient.Send:
		var message Message
		err := json.Unmarshal(msg, &message)
		assert.NoError(t, err)
		assert.Contains(t, message.Message, "pulse-A:")
		assert.Contains(t, message.Message, "0A0A0A0A00000000")
		assert.Contains(t, message.Message, "0A0A0A0A0A0A0A0A")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for pulse command")
	}
}

// TestClearQueue 测试清空队列指令
func TestClearQueue(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// 创建并注册客户端
	controlClient := &Client{Send: make(chan []byte, 256)}
	appClient := &Client{Send: make(chan []byte, 256)}

	controlID := hub.RegisterClient(controlClient)
	appID := hub.RegisterClient(appClient)
	time.Sleep(100 * time.Millisecond)

	// 清空初始消息
	<-controlClient.Send
	<-appClient.Send

	// 建立绑定
	hub.mu.Lock()
	hub.bindings[controlID] = appID
	hub.mu.Unlock()

	// 测试清空队列
	err := hub.ClearQueue(controlID, ChannelA)
	assert.NoError(t, err)

	// 验证指令格式
	select {
	case msg := <-appClient.Send:
		var message Message
		err := json.Unmarshal(msg, &message)
		assert.NoError(t, err)
		assert.Equal(t, "clear-1", message.Message)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for clear command")
	}
}

// TestMessageValidation 测试消息验证
func TestMessageValidation(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	controlClient := &Client{Send: make(chan []byte, 256)}
	controlID := hub.RegisterClient(controlClient)
	time.Sleep(100 * time.Millisecond)
	<-controlClient.Send

	// 测试发送给不存在的目标
	err := hub.SendCommand(controlID, "test")
	assert.Error(t, err) // 应该返回错误，因为没有绑定

	// 测试无效的强度值
	err = hub.SendStrength(controlID, ChannelA, ModeSet, 300)
	assert.Error(t, err)

	// 测试无效的通道
	err = hub.SendStrength(controlID, Channel(99), ModeSet, 50)
	assert.Error(t, err)

	// 测试无效的波形通道
	err = hub.SendPulse(controlID, "X", []string{"0A0A0A0A00000000"})
	assert.Error(t, err)

	// 测试空波形数据
	err = hub.SendPulse(controlID, "A", []string{})
	assert.Error(t, err)
}

// TestSendPulseBatching 测试大量波形数据的分批发送
func TestSendPulseBatching(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// 创建并注册客户端
	controlClient := &Client{Send: make(chan []byte, 256)}
	appClient := &Client{Send: make(chan []byte, 256)}

	controlID := hub.RegisterClient(controlClient)
	appID := hub.RegisterClient(appClient)
	time.Sleep(100 * time.Millisecond)

	// 清空初始消息
	<-controlClient.Send
	<-appClient.Send

	// 建立绑定
	hub.mu.Lock()
	hub.bindings[controlID] = appID
	hub.mu.Unlock()

	// 测试发送100个波形数据（会触发分批）
	hexData := make([]string, 100)
	for i := 0; i < 100; i++ {
		hexData[i] = "1E1E1E1E3C3C3C3C" // 16字符的波形数据（正确格式：8字节）
	}

	err := hub.SendPulse(controlID, "A", hexData)
	assert.NoError(t, err)

	// 验证收到了多个批次的消息
	// 批次大小是30，所以100个数据应该分成4批：30+30+30+10
	batchCount := 0
	totalDataReceived := 0

	// 等待所有批次（最多5秒）
	timeout := time.After(5 * time.Second)
	for batchCount < 4 {
		select {
		case msg := <-appClient.Send:
			var message Message
			err := json.Unmarshal(msg, &message)
			assert.NoError(t, err)
			assert.Contains(t, message.Message, "pulse-A:")

			// 计算这批数据的数量（简单统计逗号数量+1）
			dataCount := 1
			for _, ch := range message.Message {
				if ch == ',' {
					dataCount++
				}
			}
			totalDataReceived += dataCount
			batchCount++

			t.Logf("Received batch %d with %d waveforms", batchCount, dataCount)

		case <-timeout:
			t.Fatalf("Timeout waiting for batches, received %d batches", batchCount)
		}
	}

	assert.Equal(t, 4, batchCount, "Should receive 4 batches")
	assert.Equal(t, 100, totalDataReceived, "Should receive 100 total waveforms")
}
