# DG-LAB Tic-Tac-Toe 实施计划 (Implementation Plan)

本计划旨在指导开发 "DG-LAB 郊狼井字棋" 项目。请按照以下阶段顺序执行。

## Phase 1: 项目初始化与基础架构 (Initialization) ✅
目标：搭建前后端基础框架，确保配置加载和 HTTP 服务正常运行。

- [x] **后端项目结构搭建**
    - [x] 初始化 Go Module (`go mod init`).
    - [x] 创建目录结构: `cmd`, `internal/config`, `internal/server`, `internal/game`, `internal/dglab`.
- [x] **配置文件模块 (`internal/config`)**
    - [x] 创建 `config.yml` 模板 (包含 server, game, waveforms 配置项).
    - [x] 使用 `viper` 实现配置加载与解析.
    - [x] 定义 Go Struct 映射配置文件的结构.
- [x] **基础 HTTP 服务 (`internal/server`)**
    - [x] 使用 `Gin` 启动 Web Server.
    - [x] 配置 CORS 中间件 (允许前端跨域调试).
    - [x] 编写一个 `/ping` 路由测试服务存活.
- [x] **前端项目初始化**
    - [x] 使用 Vite 创建 React + TypeScript 项目.
    - [x] 安装依赖: `@mui/material`, `@emotion/react`, `@emotion/styled`, `react-router-dom`.
    - [x] 清理默认模板，确保 `yarn dev` 正常运行.

## Phase 2: DG-LAB Socket 服务核心 (IoT Core) ✅
目标：实现与郊狼 APP 的通信协议，确保能独立控制设备。

- [x] **WebSocket 基础 (`internal/dglab`)**
    - [x] 定义 `Client` 结构体 (存储 WS 连接).
    - [x] 实现 `/ws/dglab` 路由处理函数.
    - [x] 实现 `Hub` 管理所有 APP 连接.
- [x] **握手与绑定协议**
    - [x] 实现 `bind` 消息处理: 解析 `clientId` (Web端ID) 和 `targetId` (APP端ID).
    - [x] 建立 `clientId` -> `Client` 的映射关系.
    - [x] 实现心跳机制 (`heartbeat`) 保持连接.
- [x] **指令发送模块**
    - [x] 实现 `strength` 指令封装 (通道+模式+数值).
    - [x] 实现 `pulse` 波形下发逻辑 (分段发送 HEX 数据).
    - [x] 实现 `clear` 清空队列指令.
    - [x] **单元测试**: 编写一个 Go Test，模拟 APP 连接并验证服务器是否返回了正确的握手响应.

## Phase 3: 游戏核心逻辑 (Game Engine) ✅
目标：实现井字棋逻辑和房间管理，暂不涉及硬件控制。

- [x] **房间管理 (`internal/game`)**
    - [x] 定义 `Room` 和 `Player` 结构体.
    - [x] 实现 `RoomManager`: 创建房间、加入房间、查找房间.
- [x] **游戏 WebSocket 协议**
    - [x] 实现 `/ws/game` 路由.
    - [x] 定义消息类型: `join_room`, `move`, `game_over`, `room_state`.
- [x] **游戏算法**
    - [x] 实现 `MakeMove(pos int)`: 更新棋盘.
    - [x] 实现 `CheckWin()`: 判定胜负或平局.
    - [x] 实现广播机制: 每次状态变化向房间内所有玩家推送 `room_state`.
- [x] **单元测试**
    - [x] 测试房间创建和加入
    - [x] 测试落子逻辑和胜负判定
    - [x] 测试玩家配置验证
    - [x] 测试房间管理器功能


## Phase 4: 前端核心功能开发 (Frontend Core) ✅
目标：完成游戏界面和交互，实现与后端的纯软件通信。

- [x] **UI 框架搭建 (MUI)**
    - [x] 创建 `Layout` 组件.
    - [x] 创建 `HomePage`: 输入昵称、创建/加入房间.
    - [x] 创建 `GameRoom`: 基础布局 (AppBar, PlayerInfo, Board).
- [x] **游戏逻辑对接**
    - [x] 封装 WebSocket Hook (处理连接、断线重连).
    - [x] 实现 `Board` 组件: 渲染 3x3 格子，点击发送 `move` 指令.
    - [x] 状态同步: 根据后端推送的 `room_state` 更新界面.
- [x] **二维码生成**
    - [x] 引入 `qrcode.react`.
    - [x] 生成 UUID (`dglab_client_id`).
    - [x] 渲染连接二维码: `https://...#DGLAB-SOCKET#wss://.../<uuid>`.

## Phase 5: 系统集成与硬件联动 (Integration) ✅
目标：将游戏事件与 DG-LAB 服务打通，实现自动震动。

- [x] **ID 关联**
    - [x] 前端: 连接 Game WS 后，立即发送 `update_dglab_id` 消息，上传生成的 UUID.
    - [x] 后端: 在 `Player` 结构中存储 `DGLabClientID`.
- [x] **触发器实现 (`internal/game` -> `internal/dglab`)**
    - [x] **落子震动**: 在 `handleMove` 成功后，调用 `triggerMoveShock` 发送震动.
    - [x] **平局震动**: 在 `checkWin` 返回 Draw 时，调用 `triggerDrawShock` 向双方发送震动.
    - [x] **输赢震动**: 游戏结束时，通过 `triggerGameOverShock` 处理（当前留待惩罚机制）.
    - [x] **惩罚震动**: 在 `handlePunish` 中调用 `triggerPunishmentShock` 发送自定义强度和时长的震动.
- [x] **状态反馈 UI**
    - [x] 后端: 添加 `dglab.Hub.IsDeviceConnected` 方法检查设备在线状态.
    - [x] 后端: 通过 `room.BroadcastRoomState` 传递设备状态检查函数.
    - [x] 前端: 在玩家头像旁显示连接状态图标 (绿/灰) - 已实现（Phase 4）.
    - [x] 前端: 接收 `shock_event` 时弹出 Toast 提示 - 已实现（Phase 4）.
- [x] **设备断开连接处理**（补充完善）
    - [x] 后端: 在 `dglab.Hub` 添加 `OnDeviceDisconnect` 回调字段
    - [x] 后端: 在设备断开时调用回调通知游戏 Hub 更新状态
    - [x] 后端: 在 `game.Hub` 添加 `NotifyDeviceDisconnected` 方法
    - [x] 后端: 在 `server.go` 设置 `OnDeviceDisconnect` 回调
    - [x] 前端: 监听设备状态变化，显示断开连接的 Toast 提示

## Phase 6: 高级功能与配置 (Refinement) ✅
目标：实现用户自定义设置和赢家惩罚机制。

- [x] **用户配置 (Settings)**
    - [x] 前端: 创建 `SettingsDialog`.
        - [x] 双滑块: `Safe Min` - `Safe Max`.
        - [x] 单滑块: `Move Strength`, `Draw Strength`.
    - [x] 后端: 处理 `update_config` 消息，保存到 Player Session (Phase 3 已实现).
    - [x] **核心逻辑**: 修改震动发送函数，将逻辑强度映射到用户的 `Safe Range` (Phase 5 已实现).
- [x] **惩罚机制 (Punishment)**
    - [x] 后端: 实现 `punish` 接口 (校验发起者是否为 Winner) (Phase 5 已实现).
    - [x] 后端: 计算惩罚强度 (基于 Loser 的 Safe Range 和 Winner 的百分比) (Phase 5 已实现).
    - [x] 前端: 仅在胜利且游戏结束状态下显示 `PunishPanel`.
        - [x] 强度滑块 (1-100%).
        - [x] 时间滑块 (读取 Config 中的 Min/Max).

## Phase 7: 测试与部署 (Testing & Deployment)
目标：确保系统稳定，边界条件处理得当。

- [ ] **边界测试**
    - [ ] 测试 A 连接设备，B 未连接设备时的游戏流程 (不应报错).
    - [ ] 测试断网重连机制.
    - [ ] 测试输入非法的配置数值 (如 Min > Max).
- [ ] **波形调试**
    - [ ] 在 `config.yml` 中调整 HEX 波形，找到最适合 "轻微震动" 和 "强力惩罚" 的手感.
- [ ] **构建**
    - [ ] 编写 `Dockerfile` (多阶段构建: Build React -> Build Go).
    - [ ] 编写 `docker-compose.yml` (包含 Server 和 Config 挂载).

