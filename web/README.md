# DG-LAB Tic-Tac-Toe Frontend

前端使用 React + TypeScript + Vite + MUI 构建。

## 开发

```bash
# 安装依赖
yarn install

# 启动开发服务器
yarn dev

# 构建生产版本
yarn build

# 预览生产版本
yarn preview
```

## 环境配置

复制 `.env.example` 为 `.env` 并配置 WebSocket 服务器端口：

```bash
cp .env.example .env
```

默认配置：
- `VITE_WS_PORT=8080` - WebSocket 服务器端口

## 功能特性

- ✅ 首页：输入昵称、创建/加入房间
- ✅ 游戏房间：实时井字棋对战
- ✅ 棋盘交互：点击落子，实时更新
- ✅ 玩家状态：显示在线状态和设备连接状态
- ✅ 二维码生成：用于连接 DG-LAB 设备
- ✅ WebSocket 通信：自动重连机制
- ✅ 游戏通知：使用 toast 显示游戏事件

## 项目结构

```
src/
├── App.tsx                    # 主应用组件，React Router 配置
├── main.tsx                   # React 入口文件
├── components/                # UI 组件
│   ├── Layout/               # 全局布局
│   ├── HomePage/             # 首页
│   ├── GameRoom/             # 游戏房间
│   ├── Board/                # 井字棋棋盘
│   └── QRCodeDialog/         # DG-LAB 设备连接二维码
├── hooks/
│   └── useGameWebSocket.ts   # WebSocket Hook
├── types/
│   └── game.ts               # TypeScript 类型定义
└── utils/                    # 工具函数
```

## 技术栈

- **React 19** - UI 框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **Material UI v7** - UI 组件库
- **React Router v7** - 路由管理
- **react-hot-toast** - 通知系统
- **qrcode.react** - 二维码生成
- **uuid** - UUID 生成
