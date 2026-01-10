# DG-LAB Tic-Tac-Toe (Shock-Tac-Toe) 项目架构文档

## 1. 项目概述
本项目是一个基于 Web 的在线井字棋（Tic-Tac-Toe）游戏，集成了 **DG-LAB 郊狼 V3** 脉冲主机的控制功能。
**核心玩法**：
- 用户创建/加入房间进行对战。
- 每次落子、平局、输掉比赛都会触发不同强度的物理反馈（震动/脉冲）。
- 赢家可以在游戏结束后，在限定范围内自定义强度和时长，向输家发送“惩罚”。
- 支持用户自定义安全强度范围（Min-Max）。

## 2. 技术栈 (Tech Stack)

### 2.1 前端 (Frontend)
*   **Framework**: React + Vite + yarn
*   **Language**: TypeScript
*   **UI Library**: Material UI (MUI) v5 - 遵循 Material Design 风格。
*   **State Management**: React Context 或 Zustand (用于管理 WebSocket 和游戏状态)。
*   **Utilities**:
    *   `qrcode.react`: 生成连接郊狼的二维码。
    *   `react-hot-toast`: 用于显示震动触发提示和游戏通知。
    *   `react-use-websocket`: (可选) 简化 WS 连接管理。

### 2.2 后端 (Backend)
*   **Language**: Go (Golang) >= 1.21
*   **Web Framework**: 标准库 `net/http` 或 `Gin` (推荐 Gin 以简化路由)。
*   **WebSocket**: `github.com/gorilla/websocket`
*   **Configuration**: `github.com/spf13/viper` (解析 YAML)。
*   **ID Generation**: `github.com/google/uuid`

---

## 3. 系统架构与数据流

系统由 **HTTP/WS 服务器**、**Web 前端**、**DG-LAB APP** 三部分组成。Go 后端同时承担“游戏服务器”和“DG-LAB Socket 服务端”的角色。

### 3.1 核心实体关系
1.  **Game Client (Web)**: 玩家操作界面。
2.  **DGLAB Client (Phone)**: 运行 DG-LAB APP 的手机，连接郊狼主机。
3.  **Server**: 维护游戏房间状态，同时维护 DGLAB 的 Socket 连接映射。

### 3.2 标识符映射逻辑
为了让游戏事件触发脉冲，服务器需要知道哪个玩家对应哪个郊狼设备：
1.  **Web 端** 生成一个 `dglab_client_id` (UUID)。
2.  **Web 端** 通过 WS 将此 ID 发送给 **Game Room** (绑定玩家 -> ID)。
3.  **Web 端** 展示包含此 ID 的二维码。
4.  **Phone 端** 扫码连接 **Server** 的 `/ws/dglab` 端口。
5.  **Server** 记录 `dglab_client_id` <-> `dglab_target_app_id` 的映射。
6.  当 **Game Room** 判定玩家 A 需要震动 -> 查找玩家 A 的 `dglab_client_id` -> 查找对应的 `dglab_target_app_id` -> 发送指令。

---

## 4. 业务逻辑规范

### 4.1 配置文件 (`config.yml`)
所有系统级参数必须可配置。

```yaml
server:
  port: 8080
  host: "0.0.0.0" # 适配容器化或云部署

game:
  # 输掉比赛时，惩罚持续时间的可选范围 (秒)
  punishment_duration_min: 1.0
  punishment_duration_max: 10.0
  
  # 自动触发震动的持续时间 (秒)
  move_duration: 0.5
  draw_duration: 1.0

waveforms:
  # 波形数据 (HEX String, 每8字节100ms)
  # 可以在此处定义多种波形，代码中通过 key 调用
  default: "0A0A0A0A0A0A0A0A" 
  pulse: "00000000FFFFFFFF00000000"
```

### 4.2 强度与安全限制逻辑
每个玩家在前端设置自己的偏好，这些数据保存在后端房间的玩家状态中。

**玩家设置项 (User Settings):**
*   `safe_min` (0-100): 最小能感知的强度。
*   `safe_max` (0-100): 最大能承受的强度。
*   `move_strength` (int): 落子时的强度值 (需在 safe_min ~ safe_max 之间)。
*   `draw_strength` (int): 平局时的强度值 (需在 safe_min ~ safe_max 之间)。

**惩罚计算公式 (输家视角):**
赢家选择 `intensity_percent` (1% - 100%)。
输家实际受到的强度 `actual_strength` 计算如下：
```go
range = loser.safe_max - loser.safe_min
actual_strength = loser.safe_min + (range * intensity_percent / 100)
```
*注意：发送给郊狼的最终数值通常是 0-200，需将 0-100 映射到设备协议的范围。*

### 4.3 触发场景矩阵

| 场景 | 触发对象 | 强度来源 | 持续时间 | 波形 |
| :--- | :--- | :--- | :--- | :--- |
| **玩家落子** | 落子方 | 玩家配置 `move_strength` | Config `move_duration` | Config `default` |
| **平局** | 双方 | 玩家配置 `draw_strength` | Config `draw_duration` | Config `default` |
| **分出胜负** | 输家 | 赢家设定 % -> 映射公式 | 赢家设定 (Config范围内) | Config `pulse` |

---

### 举个例子说明流程

假设配置如下：
*   `loss_duration`: 2秒
*   `punishment_duration_min`: 1秒
*   `punishment_duration_max`: 5秒

**游戏过程：**
1.  **A 赢了，B 输了。**
2.  **阶段一（自动）**：B 的设备**立即**震动 **2秒**（`loss_duration`）。此时 B 知道自己输了。
3.  **阶段二（手动）**：A 的屏幕上出现一个滑块，范围是 **1秒 到 5秒**。
4.  A 觉得刚才 B 下棋太慢了，想惩罚一下，于是把滑块拖到 **4秒**，点击“发送”。
5.  B 的设备再次震动 **4秒**。

**总结：**
*   `loss_duration` 是**裁判的哨声**（固定）。
*   `duration min/max` 是**赢家的奖励范围**（可变）。

---

## 5. 接口与协议设计

### 5.1 WebSocket 路由
*   `/ws/game`: 处理游戏逻辑、房间管理、聊天、玩家配置同步。
*   `/ws/dglab`: 处理 DG-LAB APP 的连接、绑定、心跳 (遵循 `dg-lab.md` 协议)。

### 5.2 游戏协议 (JSON Payload)

**Client -> Server:**
```json
// 1. 创建/加入
{ "type": "join_room", "room_id": "1234", "player_name": "Alice" }

// 2. 更新设备绑定ID (前端生成ID后通知后端)
{ "type": "update_dglab_id", "dglab_client_id": "uuid-..." }

// 3. 更新安全配置
{ 
  "type": "update_config", 
  "config": { "safe_min": 10, "safe_max": 60, "move_strength": 15, "draw_strength": 30 } 
}

// 4. 落子
{ "type": "move", "position": 4 } // 0-8

// 5. 发送惩罚 (仅赢家可用)
{ "type": "punish", "percent": 80, "duration": 5 }
```

**Server -> Client:**
```json
// 1. 房间状态更新 (广播)
{ 
  "type": "room_state", 
  "board": [0,0,1...], 
  "turn": "Alice", 
  "players": {
      "Alice": { "connected": true, "device_active": true },
      "Bob": { "connected": true, "device_active": false }
  }
}

// 2. 游戏结束
{ "type": "game_over", "winner": "Alice", "line": [0,1,2] }

// 3. 震动通知 (用于前端 Toast)
{ 
  "type": "shock_event", 
  "target": "Bob", 
  "intensity": 45, 
  "reason": "move" // move, draw, loss, punish
}
```

---

## 6. 前端 UI 设计 (Material Design)

### 6.1 页面流
1.  **Home**: 输入昵称 -> [创建房间] / [输入房间号加入]。
2.  **Game Room**:
    *   **AppBar**: 房间号显示，右侧显示“设置”图标。
    *   **Player Info Area**: 显示双方名字、头像。
        *   关键：名字旁边显示 **"郊狼图标"**。
        *   图标绿色：设备已连接。
        *   图标灰色：设备未连接。
        *   图标震动动画：正在接收脉冲。
    *   **Board Area**: 3x3 棋盘，响应式布局。
    *   **Action Area**:
        *   未连接时：显示 [Connect Toy] 按钮 -> 打开 Dialog 显示二维码。
        *   游戏结束且胜利时：显示 [Punish Panel]。
    *   **Punish Panel**:
        *   Slider 1: 强度 (1% - 100%)。
        *   Slider 2: 时长 (Config.min - Config.max)。
        *   Button: [SEND SHOCK]。

### 6.2 设置弹窗 (Settings Dialog)
*   **Range Slider**: 设置 `safe_min` 和 `safe_max`。
*   **Slider**: 设置 `move_strength` (受限于 Range)。
*   **Slider**: 设置 `draw_strength` (受限于 Range)。

---

## 7. 后端实现细节 (Go)

### 7.1 目录结构（规划）
```text
/
├── config.yml
├── main.go            # 入口
├── internal/
│   ├── config/        # Viper 配置加载
│   ├── game/          # 井字棋逻辑, 房间管理
│   ├── dglab/         # DG-LAB Socket 服务实现 (参考 dg-lab.md)
│   └── server/        # HTTP/WS 路由处理
└── web/               # React 前端产物
```

### 7.1.1 当前实现的目录结构（Phase 2 完成）
```text
/
├── config.yml                  # 配置文件（server, game, waveforms）
├── go.mod                      # Go 模块依赖管理
├── go.sum                      # 依赖校验和
├── tictactoe-server           # 编译后的服务器二进制文件
├── cmd/
│   └── main.go                # 服务器主入口，加载配置并启动 HTTP 服务
├── internal/
│   ├── config/
│   │   ├── config.go          # 配置加载模块（使用 viper）
│   │   └── config_test.go     # 配置模块单元测试
│   ├── dglab/                 # DG-LAB Socket 服务模块 ✅
│   │   ├── types.go           # 消息类型、Client、Binding 结构定义
│   │   ├── hub.go             # Hub: 管理所有连接、绑定关系、心跳
│   │   ├── client.go          # Client: ReadPump/WritePump 处理 WS 读写
│   │   ├── handler.go         # WebSocket HTTP 升级处理器
│   │   ├── commands.go        # 指令封装: SendStrength, SendPulse, ClearQueue
│   │   └── dglab_test.go      # 单元测试（覆盖注册、绑定、指令发送）
│   ├── game/                  # 游戏逻辑模块（待实现）
│   └── server/
│       ├── server.go          # HTTP 服务器（Gin，CORS，/ping，/ws/dglab 路由）
│       └── server_test.go     # 服务器模块单元测试
└── web/                       # React 前端项目
    ├── src/
    │   ├── App.tsx            # 主应用组件（简化版，使用 MUI）
    │   ├── main.tsx           # React 入口文件
    │   └── assets/            # 静态资源
    ├── public/                # 公共资源
    ├── package.json           # 前端依赖配置
    ├── vite.config.ts         # Vite 配置
    └── tsconfig.json          # TypeScript 配置
```

**已实现的功能（Phase 1-3）：**
- ✅ Go 后端基础架构：配置加载、HTTP 服务、CORS 支持
- ✅ 配置文件系统：支持 YAML 格式配置，包含服务器、游戏、波形参数
- ✅ 单元测试：config、server、dglab 和 game 模块均有测试覆盖
- ✅ React 前端基础：Vite + TypeScript + MUI 框架搭建完成
- ✅ 健康检查：`/ping` 端点用于服务存活性检测
- ✅ **DG-LAB WebSocket 服务完整实现**：
  - **Hub 管理**: 客户端注册/注销、绑定关系维护、心跳保活
  - **WebSocket 路由**: `/ws/dglab` 端点，支持 APP 连接
  - **握手协议**: 自动分配 UUID、发送初始绑定消息
  - **绑定逻辑**: 处理 APP 的 bind 请求，建立控制端-APP 映射
  - **指令系统**:
    - `SendStrength`: 强度控制（增/减/设置，通道A/B）
    - `SendPulse`: 波形数据下发（支持HEX数组）
    - `ClearQueue`: 清空波形队列
  - **错误处理**: 完整的错误码系统（200/400/401/402/404/405）
  - **并发安全**: 使用 sync.RWMutex 保护共享数据
  - **单元测试覆盖**: 8个测试用例，覆盖注册、绑定、指令发送、参数验证
- ✅ **Game WebSocket 服务完整实现** (Phase 3):
  - **Hub 管理**: 房间管理、玩家连接、消息处理
  - **WebSocket 路由**: `/ws/game` 端点，支持游戏客户端连接
  - **房间系统**: 创建房间、加入房间、自动清理空房间
  - **游戏逻辑**: 完整的井字棋规则、胜负判定、平局判定
  - **消息协议**: join_room, update_dglab_id, update_config, move, punish
  - **状态同步**: 实时广播房间状态和游戏结果
  - **玩家配置**: 支持自定义安全强度范围
  - **并发安全**: 使用 sync.RWMutex 保护共享数据
  - **单元测试覆盖**: 7个测试用例，覆盖房间管理、游戏逻辑、配置验证

**待实现的模块：**
- ⏳ 前端游戏界面和 WebSocket 通信 (Phase 4)
- ⏳ 游戏与 DG-LAB 硬件联动 (Phase 5)

### 7.2 DG-LAB 服务模块 (`internal/dglab`) ✅ 已实现

**模块职责**：
- 实现完整的 DG-LAB WebSocket 通信协议（参考 `dg-lab.md`）
- 管理控制端（Web）和 APP 端的连接与绑定
- 提供高级 API 供游戏模块调用

**核心组件**：

1. **types.go** - 类型定义
   - `Message`: 通信消息结构（type, clientId, targetId, message）
   - `MessageType`: 消息类型常量（bind, msg, heartbeat, break, error）
   - `Client`: WebSocket 客户端封装
   - `Binding`: 绑定关系结构

2. **hub.go** - 连接管理中心
   - `Hub`: 核心管理器，维护所有连接和绑定关系
   - `Run()`: 主事件循环（处理注册/注销/心跳）
   - `RegisterClient()`: 注册新客户端并分配 UUID
   - `HandleBind()`: 处理绑定请求（注意字段语义，详见文档）
   - `SendCommand()`: 向绑定的 APP 发送指令
   - **并发安全**: 使用 `sync.RWMutex` 保护共享数据

3. **client.go** - 客户端读写循环
   - `ReadPump()`: 从 WebSocket 读取消息并分发处理
   - `WritePump()`: 向 WebSocket 写入消息（带 Ping/Pong）
   - 自动处理消息验证、转发、错误响应

4. **handler.go** - HTTP 升级处理
   - `HandleWebSocket()`: 返回 Gin 路由处理器
   - 升级 HTTP 连接为 WebSocket
   - 启动客户端读写协程

5. **commands.go** - 指令封装层
   - `SendStrength(clientID, channel, mode, value)`: 强度控制
   - `SendPulse(clientID, channel, hexData)`: 波形下发
   - `ClearQueue(clientID, channel)`: 清空队列
   - 快捷方法: `SendStrengthQuick`, `SendStrengthSet`, `SendStrengthZero`
   - 参数验证和格式化

**与游戏模块的集成方式**：
```go
// 获取 Hub 实例
hub := server.GetDGLabHub()

// 发送震动指令
hub.SendStrength(playerDGLabID, dglab.ChannelA, dglab.ModeSet, 50)
hub.SendPulse(playerDGLabID, "A", []string{"0A0A0A0A00000000", "..."})
hub.ClearQueue(playerDGLabID, dglab.ChannelB)
```

**测试覆盖**（dglab_test.go）：
- ✅ Hub 创建和客户端注册
- ✅ 绑定流程（控制端-APP 配对）
- ✅ 指令发送和消息转发
- ✅ 强度/波形/清空指令格式验证
- ✅ 参数边界检查（无效通道、超范围值）

**注意事项**：
- WebSocket 升级直接由 Go 处理，无需 Nginx 代理
- 生产环境使用 WSS（需配置 TLS）
- 心跳间隔 60 秒，Ping/Pong 保活机制
- 消息最大长度 1950 字节（协议限制）

### 7.3 游戏模块 (`internal/game`) ✅ Phase 3 完成

**模块职责**：
- 实现完整的井字棋游戏逻辑
- 管理房间和玩家连接
- 处理游戏 WebSocket 通信协议
- 提供玩家配置管理功能

**核心组件**：

1. **types.go** - 类型定义
   - `Message`: 游戏消息结构（type, room_id, player_name, position, config 等）
   - `MessageType`: 消息类型常量（join_room, update_dglab_id, update_config, move, punish, room_state, game_over, shock_event）
   - `Player`: 玩家结构（name, conn, dglab_client_id, config, symbol, send channel）
   - `Room`: 房间结构（id, board, turn, players, game_over, winner）
   - `PlayerConfig`: 玩家配置（safe_min, safe_max, move_strength, draw_strength）
   - `PlayerInfo`: 广播用的玩家信息（connected, device_active）

2. **room.go** - 房间游戏逻辑
   - `MakeMove(player, position)`: 落子逻辑，验证合法性，切换回合
   - `checkWin()`: 胜负判定，支持8种获胜模式和平局检测
   - `Broadcast(msg)`: 向房间内所有玩家广播消息
   - `BroadcastRoomState()`: 广播当前房间状态（棋盘、回合、玩家状态）
   - `BroadcastGameOver()`: 广播游戏结束消息
   - `Reset()`: 重置房间（重新开始游戏）
   - `RemovePlayer(player)`: 移除玩家
   - 辅助方法：`GetPlayerBySymbol`, `GetPlayerByName`, `GetOpponent`, `IsFull`, `IsEmpty`

3. **manager.go** - 房间管理器
   - `RoomManager`: 维护所有房间的映射关系
   - `CreateRoom()`: 创建新房间，自动生成6位房间ID
   - `GetRoom(roomID)`: 获取指定房间
   - `JoinRoom(roomID, player)`: 加入房间，自动分配 X/O 符号
   - `DeleteRoom(roomID)`: 删除房间
   - `CleanEmptyRooms()`: 定期清理空闲超过10分钟的空房间
   - **并发安全**: 使用 `sync.RWMutex` 保护房间映射

4. **client.go** - 客户端读写循环
   - `ReadPump(hub)`: 从 WebSocket 读取消息并发送到 Hub 处理
   - `WritePump()`: 向 WebSocket 写入消息（带 Ping/Pong 保活）
   - `SendError(msg)`: 发送错误消息给玩家
   - `UpdateDGLabID(clientID)`: 更新玩家的 DG-LAB 客户端 ID
   - `UpdateConfig(config)`: 更新并验证玩家配置
   - `GetDGLabID()`: 获取玩家的 DG-LAB 客户端 ID

5. **hub.go** - 游戏连接管理中心
   - `Hub`: 核心管理器，维护房间管理器和玩家连接
   - `Run()`: 主事件循环（处理注册/注销/消息/定时清理）
   - `handleRegister(player)`: 处理玩家注册
   - `handleUnregister(player)`: 处理玩家断开（广播状态、清理空房间）
   - `processMessage(pm)`: 消息路由和处理
   - `handleJoinRoom(player, msg)`: 处理加入/创建房间请求
   - `handleUpdateDGLabID(player, msg)`: 处理 DG-LAB ID 更新
   - `handleUpdateConfig(player, msg)`: 处理配置更新
   - `handleMove(player, msg)`: 处理落子，触发状态广播
   - `handlePunish(player, msg)`: 处理惩罚请求（验证权限）

6. **handler.go** - HTTP 升级处理
   - `HandleWebSocket(hub)`: 返回 Gin 路由处理器
   - 升级 HTTP 连接为 WebSocket
   - 创建玩家实例并启动读写协程

7. **errors.go** - 错误定义
   - 房间错误：`ErrRoomNotFound`, `ErrRoomFull`, `ErrPlayerNotFound`
   - 游戏错误：`ErrGameOver`, `ErrNotYourTurn`, `ErrInvalidMove`, `ErrPositionOccupied`
   - 配置错误：`ErrInvalidConfig`
   - 权限错误：`ErrNotWinner`

**消息协议**：

客户端 -> 服务器：
- `join_room`: 创建或加入房间（可选 room_id）
- `update_dglab_id`: 更新 DG-LAB 客户端 ID
- `update_config`: 更新玩家配置（safe_min, safe_max, move_strength, draw_strength）
- `move`: 落子（position: 0-8）
- `punish`: 发送惩罚（仅赢家可用，需提供 percent 和 duration）

服务器 -> 客户端：
- `room_state`: 房间状态更新（board, turn, players）
- `game_over`: 游戏结束（winner, line）
- `shock_event`: 震动通知（target, intensity, reason）
- `error`: 错误消息

**游戏流程**：

1. **加入房间**：
   - 玩家发送 `join_room` 消息（可选房间 ID）
   - 服务器创建新房间或加入现有房间
   - 自动分配 X（先手）或 O（后手）符号
   - 广播房间状态给所有玩家

2. **游戏进行**：
   - 玩家轮流发送 `move` 消息
   - 服务器验证合法性（回合、位置）
   - 更新棋盘并检查胜负
   - 广播最新房间状态

3. **游戏结束**：
   - 检测到胜利或平局
   - 广播 `game_over` 消息
   - 赢家可发送 `punish` 请求

**与 DG-LAB 模块的集成**（Phase 5 实现）：
```go
// 在 game hub 中获取 dglab hub 实例
dglabHub := server.GetDGLabHub()

// 发送震动（落子）
playerDGLabID := player.GetDGLabID()
dglabHub.SendStrength(playerDGLabID, dglab.ChannelA, dglab.ModeSet, player.Config.MoveStrength)

// 发送震动（惩罚）
loserDGLabID := loser.GetDGLabID()
strength := calculatePunishmentStrength(loser.Config, percent)
dglabHub.SendPulse(loserDGLabID, "A", waveformData)
```

**测试覆盖**（game_test.go）：
- ✅ 房间创建和基本属性
- ✅ 玩家加入房间（分配符号、房间满员）
- ✅ 落子逻辑（合法性、回合切换、位置占用）
- ✅ 胜负判定（横排、竖排、对角线、平局）
- ✅ 玩家配置验证（范围检查、一致性）
- ✅ 房间管理器（创建、获取、删除、唯一性）
- ✅ 空房间清理（时间阈值）

**注意事项**：
- 所有共享数据使用 `sync.RWMutex` 保护，避免竞态条件
- 广播方法在释放锁后再调用 `Broadcast`，避免死锁
- 房间 ID 使用 UUID 前6位，足够随机且易于输入
- 玩家断开连接时自动广播状态更新
- 空房间定期清理，避免内存泄漏

---

## 8. 开发步骤建议

1.  **Backend Base**: 初始化 Go 项目，配置 Viper 读取 `config.yml`，搭建基础 Gin/HTTP 服务。
2.  **DG-LAB Service**: 实现 `internal/dglab`，确保手机 APP 能扫码连接，并能通过 Postman/Curl 触发测试震动。
3.  **Game Logic**: 实现 WebSocket 游戏房间逻辑（不含震动），完成 React 前端的基础对战功能。
4.  **Integration**: 将 Game 事件与 DG-LAB Service 联通。
    *   前端生成 UUID 传给后端。
    *   前端生成二维码。
    *   后端在游戏事件发生时查找 ID 并触发震动。
5.  **UI Polish**: 应用 Material Design，添加 Toast 提示，完善设置面板和惩罚面板。
6.  **Testing**: 
    *   测试单人连接、双人连接、一方未连接的情况（需确保程序不崩溃）。
    *   测试安全范围限制是否生效。
