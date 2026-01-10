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
*   **Framework**: React + Vite
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

### 7.1 目录结构
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

### 7.2 DG-LAB 服务模块 (`internal/dglab`)
*   必须实现 `dg-lab.md` 中的 `bind`, `msg`, `heartbeat` 处理。
*   提供一个 `SendCommand(dglabClientId string, strength int, duration int, waveformKey string)` 方法供 `game` 模块调用。
*   **注意**: 这里的 Socket 服务不需要 Nginx 代理，直接由 Go 处理。但在生产环境若有 HTTPS，需确保 WS 升级为 WSS。

### 7.3 游戏循环模块 (`internal/game`)
*   **Hub**: 管理所有房间。
*   **Room**:
    *   包含 `PlayerA`, `PlayerB`。
    *   包含 `Board` [9]int。
    *   处理 `Move` 逻辑。
    *   **关键**: 当 Move 合法时，调用 `dglab.SendCommand` 给当前玩家发送轻微震动。
    *   **关键**: 游戏结束判定输赢后，调用 `dglab.SendCommand` 给输家发送强力震动。

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
