# DG-LAB 郊狼 (Coyote) V3 Socket 控制开发文档

## 1. 项目概述

**DG-LAB Socket 控制** 允许开发者通过 Socket 服务连接到 DG-LAB APP，从而对 **郊狼脉冲主机 3.0 (Coyote V3)** 进行远程控制。
系统采用 **N(控制端) - Socket 服务 - N(APP端)** 的架构模式。

### 1.1 核心术语
*   **控制端 (Client/Frontend)**：发起控制指令的终端（网页、脚本、游戏等）。
*   **Socket 服务 (Server/Backend)**：中间件，负责维护连接、绑定关系、转发指令。
*   **APP 端 (Target/DG-LAB APP)**：安装在手机上的 DG-LAB 应用，通过蓝牙连接郊狼主机，作为指令的执行网关。

### 1.2 系统架构与数据流
```mermaid
graph LR
    A[控制端 (网页/脚本)] -- WebSocket --> B(Socket 服务端)
    C[DG-LAB APP] -- WebSocket --> B
    C -- 蓝牙 BLE --> D[郊狼主机 V3]
    
    subgraph "数据流向"
    A ->> B: 发送控制指令 (JSON)
    B ->> C: 转发指令
    C ->> D: 执行脉冲/强度变化
    C ->> B: 反馈状态 (强度/按键)
    B ->> A: 转发反馈
    end
```

---

## 2. 连接与绑定流程 (Handshake)

连接过程依赖于 **二维码** 进行配对。必须严格遵循以下时序：

1.  **控制端连接**：控制端连接 Socket 服务。
2.  **ID 分配**：Socket 服务生成唯一的 `clientId` 并返回给控制端。
3.  **生成二维码**：控制端使用特定格式生成二维码（包含 Socket 地址和 `clientId`）。
4.  **APP 扫码**：用户使用 DG-LAB APP 扫描二维码。
5.  **APP 连接**：APP 解析二维码，连接 Socket 服务，服务器分配 `targetId` (APP 的 ID)。
6.  **发送绑定请求**：APP 向服务器发送 `bind` 指令，携带从二维码获取的 `clientId`。
7.  **建立绑定**：服务器将 `clientId` 和 `targetId` 绑定为一对一（或一对多）关系。
8.  **通知结果**：服务器向双方发送绑定成功消息 (200 OK)。

**⚠️ 重要：APP 绑定消息的字段含义**

当 APP 发送绑定请求时，消息格式为：
```json
{
  "type": "bind",
  "clientId": "从二维码中获取的控制端ID",
  "targetId": "APP自己的ID（服务器分配的）",
  "message": "DGLAB"
}
```

**关键点：**
- `clientId` 字段：包含从二维码中解析出的**控制端 ID**（不是 APP 自己的 ID）
- `targetId` 字段：包含 **APP 自己的 ID**（服务器在连接时分配的）

**服务器端处理逻辑：**
```javascript
function handleBind(ws, clientId, targetId) {
    // 检查 clientId 是否匹配预期的控制端 ID
    if (clientId === controlClientId) {  // 注意：检查 clientId，不是 targetId！
        appTargetId = targetId;  // targetId 才是 APP 的 ID
        relations.set(controlClientId, appTargetId);
        // 发送绑定成功消息...
    }
}
```

这是一个常见的混淆点：很容易错误地认为 `clientId` 是发送者（APP）的 ID，但实际上它是从二维码中获取的控制端 ID。

### 2.1 二维码生成规则
二维码内容必须严格符合以下格式，否则 APP 无法识别：
```text
https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#<Socket服务地址>/<clientId>
```
*   **前缀**：`https://www.dungeon-lab.com/app-download.php` (固定)
*   **标签**：`DGLAB-SOCKET` (固定)
*   **分隔符**：使用两个 `#` 分割这三部分。
*   **Socket地址**：例如 `wss://ws.dungeon-lab.cn` 或 `ws://192.168.1.100:9999`。
*   **ID**：Socket 服务分配给控制端的 `clientId`。

**示例**：
`https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#wss://my-server.com/1234-5678-uuid`

---

## 3. 通信协议 (Protocol)

所有消息均为 **JSON** 格式。

### 3.1 基础消息结构
```json
{
  "type": "string",       // 消息类型: bind, msg, heartbeat, break, error
  "clientId": "string",   // 控制端 ID
  "targetId": "string",   // APP 端 ID
  "message": "string"     // 具体指令或数据内容
}
```
*   **字符限制**：JSON 字符串最大长度 **1950**，超过将被 APP 丢弃。
*   **ID 格式**：推荐使用 UUID v4 (32位)。

### 3.2 消息类型 (Type) 定义

| Type | 描述 | 发送方 | 接收方 | 备注 |
| :--- | :--- | :--- | :--- | :--- |
| `heartbeat` | 心跳包 | Server | Client/APP | 保持连接活跃 |
| `bind` | 绑定关系 | Server/APP | Server/Client | 建立连接握手 |
| `msg` | 业务指令 | Client/APP | Server/APP | 波形、强度、反馈等核心数据 |
| `break` | 断开连接 | Server | Client/APP | 通知对方已掉线 |
| `error` | 错误信息 | Server | Client/APP | 异常处理 |

---

## 4. 业务指令详解 (type: "msg")

这是核心控制部分。**注意：以下指令内容均放置在 JSON 的 `message` 字段中。**

### 4.1 强度控制 (控制端 -> APP)
控制郊狼 A/B 通道的强度变化。

**格式**：`strength-{通道}+{模式}+{数值}`

*   **通道**：`1` (A通道), `2` (B通道)
*   **模式**：
    *   `0`: 减少强度
    *   `1`: 增加强度
    *   `2`: 设置为指定数值
*   **数值**：`0-200` 的整数

**示例**：
*   A通道强度+5: `strength-1+1+5`
*   B通道强度归零: `strength-2+2+0`
*   B通道强度-20: `strength-2+0+20`
*   A通道设为35: `strength-1+2+35`

### 4.2 波形下发 (控制端 -> APP)
下发具体的脉冲波形数据。

**格式**：`pulse-{通道}:[HEX数组]`

*   **通道**：`A` (A通道), `B` (B通道)
*   **HEX数组**：JSON 数组字符串，每个元素为 8 字节 HEX 码 (代表 100ms 数据)。
*   **限制**：数组最大长度 100 (10秒)。APP 队列最大缓存 500 (50秒)。

**示例**：
`pulse-A:["0A0A0A0A00000000","0A0A0A0A0A0A0A0A"]`

### 4.3 清空队列 (控制端 -> APP)
在发送新波形前，建议清空原有队列以实现“立即执行”。

**格式**：`clear-{通道}`

*   **通道**：`1` (A通道), `2` (B通道)

**示例**：`clear-1`

### 4.4 接收强度反馈 (APP -> 控制端)
APP 端强度发生变化时会主动上报。

**格式**：`strength-{A强度}+{B强度}+{A上限}+{B上限}`

**示例**：`strength-11+7+100+35`
(A强度11, B强度7, A软上限100, B软上限35)

### 4.5 接收按键反馈 (APP -> 控制端)
用户点击 APP 上的图形按钮时触发。

**格式**：`feedback-{index}`

*   **index**：
    *   A通道按钮: `0, 1, 2, 3, 4`
    *   B通道按钮: `5, 6, 7, 8, 9`

---

## 5. 前端与后端交互协议 (特定于示例代码)

如果你使用官方提供的 `WebSocketNode.js` 作为后端，前端发送给后端的消息格式有特殊封装，后端会将其转换为上述的标准 APP 协议。

### 5.1 强度操作 (Frontend -> Backend)
```json
{
  "type": 1, // 1:减, 2:加, 3:归零, 4:指定值
  "strength": 5, // 变化量或目标值
  "message": "set channel",
  "channel": 1, // 1:A, 2:B
  "clientId": "...",
  "targetId": "..."
}
```

### 5.2 波形发送 (Frontend -> Backend)
后端包含定时器逻辑，负责将长波形分段发送。
```json
{
  "type": "clientMsg",
  "message": "...", // A通道波形数据 (不带 pulse-前缀)
  "message2": "...", // B通道波形数据 (可选，示例代码中通常分开发送)
  "time": 5, // 持续发送时长(秒)，后端会根据此时间循环发送波形
  "channel": "A", // "A" 或 "B"
  "clientId": "...",
  "targetId": "..."
}
```

---

## 6. 错误码定义

| 代码 | 含义 |
| :--- | :--- |
| **200** | 成功 |
| **209** | 对方客户端已断开 |
| **210** | 二维码中没有有效的 clientID |
| **211** | Socket已连接，但服务器未下发绑定ID |
| **400** | 此 ID 已被绑定 |
| **401** | 目标客户端不存在 |
| **402** | 双方未建立绑定关系 |
| **403** | 非标准 JSON 格式 |
| **404** | 未找到收信人（离线） |
| **405** | 消息长度超过 1950 |
| **500** | 服务器内部异常 |

---

## 7. 完整参考实现代码

### 7.1 后端代码 (Node.js)
功能：WebSocket 服务、ID 生成、关系绑定、心跳保活、波形定时分发。

```javascript
/**
 * DG-LAB Socket Backend (Node.js)
 * 依赖: npm install ws uuid
 */
const WebSocket = require('ws');
const { v4: uuidv4 } = require('uuid');

// 存储连接对象: Map<clientId, WebSocket>
const clients = new Map();
// 存储绑定关系: Map<clientId, targetId> (1对1)
const relations = new Map();
// 存储波形发送定时器: Map<clientId-channel, intervalId>
const clientTimers = new Map();

// 配置
const PORT = 9999;
const PULSE_SEND_RATE = 1; // 每秒发送次数
const DEFAULT_DURATION = 5; // 默认持续时间(秒)

// 心跳包模板
const heartbeatMsg = { type: "heartbeat", clientId: "", targetId: "", message: "200" };

const wss = new WebSocket.Server({ port: PORT });
console.log(`DG-LAB Socket Server started on port ${PORT}`);

wss.on('connection', function connection(ws) {
    const clientId = uuidv4();
    console.log('New connection:', clientId);
    clients.set(clientId, ws);

    // 1. 发送 ID 给客户端 (握手第一步)
    ws.send(JSON.stringify({ type: 'bind', clientId, message: 'targetId', targetId: '' }));

    ws.on('message', function incoming(messageStr) {
        console.log("Received:", messageStr.toString());
        let data = null;
        try {
            data = JSON.parse(messageStr);
        } catch (e) {
            ws.send(JSON.stringify({ type: 'msg', clientId: "", targetId: "", message: '403' }));
            return;
        }

        const { type, clientId: senderId, targetId, message } = data;

        // 安全检查：发送者必须是连接持有者
        if (clients.get(senderId) !== ws && clients.get(targetId) !== ws) {
            ws.send(JSON.stringify({ type: 'msg', clientId: "", targetId: "", message: '404' }));
            return;
        }

        // 核心消息处理
        switch (type) {
            case "bind":
                handleBind(ws, senderId, targetId);
                break;
            
            // 前端协议转换：强度控制 (Type 1-4)
            case 1: // 减
            case 2: // 加
            case 3: // 归零
            case 4: // 指定
                handleStrengthControl(ws, data);
                break;

            // 前端协议转换：波形发送
            case "clientMsg":
                handlePulseSend(ws, data);
                break;

            // 通用消息转发 (如 clear 指令)
            default:
                forwardMessage(ws, data);
                break;
        }
    });

    ws.on('close', () => handleDisconnect(ws, clientId));
    ws.on('error', (err) => console.error('WS Error:', err));
});

// --- 业务逻辑函数 ---

function handleBind(ws, clientId, targetId) {
    // ⚠️ 重要：参数含义
    // clientId: 从 APP 发送的消息中提取的控制端ID（从二维码获取的）
    // targetId: 从 APP 发送的消息中提取的 APP 自己的ID
    // ws: 发送此消息的 WebSocket 连接（即 APP 的连接）

    // 检查 ID 是否存在
    if (clients.has(clientId) && clients.has(targetId)) {
        // 检查是否已被绑定
        if (!isBound(clientId) && !isBound(targetId)) {
            relations.set(clientId, targetId);
            const successMsg = { type: "bind", clientId, targetId, message: "200" };
            // 向 APP 发送成功消息
            ws.send(JSON.stringify(successMsg));
            // 如果控制端也有 WebSocket 连接，也通知控制端
            // 注意：在最小实现中，控制端可能没有真实的 WebSocket 连接
            if (clients.has(clientId)) {
                clients.get(clientId).send(JSON.stringify(successMsg));
            }
            console.log(`Bound ${clientId} to ${targetId}`);
        } else {
            ws.send(JSON.stringify({ type: "bind", clientId, targetId, message: "400" }));
        }
    } else {
        ws.send(JSON.stringify({ type: "bind", clientId, targetId, message: "401" }));
    }
}

function handleStrengthControl(ws, data) {
    const { clientId, targetId, type, strength, channel } = data;
    if (relations.get(clientId) !== targetId) {
        ws.send(JSON.stringify({ type: "bind", clientId, targetId, message: "402" }));
        return;
    }

    if (clients.has(targetId)) {
        const target = clients.get(targetId);
        // 转换协议: strength-通道+模式+数值
        // 前端 type: 1(减), 2(加), 3(归零), 4(指定)
        // APP 模式: 0(减), 1(加), 2(指定)
        const appMode = type - 1; 
        const appChannel = channel || 1;
        const appStrength = type >= 3 ? strength : 1; // 加减模式下数值通常固定为1，或由前端传入

        const msgStr = `strength-${appChannel}+${appMode}+${appStrength}`;
        target.send(JSON.stringify({ type: "msg", clientId, targetId, message: msgStr }));
    }
}

function handlePulseSend(ws, data) {
    const { clientId, targetId, message, channel, time } = data;
    
    if (relations.get(clientId) !== targetId) {
        ws.send(JSON.stringify({ type: "bind", clientId, targetId, message: "402" }));
        return;
    }
    if (!channel) {
        ws.send(JSON.stringify({ type: "error", clientId, targetId, message: "406-channel empty" }));
        return;
    }

    if (clients.has(targetId)) {
        const target = clients.get(targetId);
        const duration = time || DEFAULT_DURATION;
        const sendData = { type: "msg", clientId, targetId, message: "pulse-" + message };
        
        let totalSends = PULSE_SEND_RATE * duration;
        const timeSpace = 1000 / PULSE_SEND_RATE;
        const timerKey = `${clientId}-${channel}`;

        // 如果该通道已有定时器，先清除并发送 clear 指令
        if (clientTimers.has(timerKey)) {
            console.log(`Overwriting pulse on channel ${channel}`);
            clearInterval(clientTimers.get(timerKey));
            clientTimers.delete(timerKey);

            // 发送清空指令
            const clearMsg = channel === "A" ? "clear-1" : "clear-2";
            target.send(JSON.stringify({ type: "msg", clientId, targetId, message: clearMsg }));

            // 延迟发送新波形，防止网络乱序
            setTimeout(() => {
                startPulseTimer(clientId, ws, target, sendData, totalSends, timeSpace, channel);
            }, 150);
        } else {
            startPulseTimer(clientId, ws, target, sendData, totalSends, timeSpace, channel);
        }
    } else {
        ws.send(JSON.stringify({ type: "msg", clientId, targetId, message: "404" }));
    }
}

function startPulseTimer(clientId, sourceWs, targetWs, sendData, totalSends, timeSpace, channel) {
    // 立即发送第一次
    targetWs.send(JSON.stringify(sendData));
    totalSends--;

    if (totalSends > 0) {
        const timerId = setInterval(() => {
            if (totalSends > 0) {
                targetWs.send(JSON.stringify(sendData));
                totalSends--;
            } else {
                clearInterval(timerId);
                clientTimers.delete(`${clientId}-${channel}`);
                sourceWs.send(JSON.stringify({type: "msg", message: "发送完毕"})); // 可选反馈
            }
        }, timeSpace);
        clientTimers.set(`${clientId}-${channel}`, timerId);
    }
}

function forwardMessage(ws, data) {
    const { clientId, targetId } = data;
    if (relations.get(clientId) !== targetId) {
        ws.send(JSON.stringify({ type: "bind", clientId, targetId, message: "402" }));
        return;
    }
    if (clients.has(targetId)) { // 如果是 APP 发来的，targetId 实际上是 ClientId，反之亦然，需根据业务调整
         // 简单转发逻辑：根据 relations 查找对方
         // 注意：标准实现中 data.targetId 应该是接收方。
         const receiver = clients.get(targetId);
         if(receiver) receiver.send(JSON.stringify(data));
    }
}

function handleDisconnect(ws, disconnectedId) {
    console.log('Disconnected:', disconnectedId);
    let partnerId = null;

    // 查找配对伙伴
    if (relations.has(disconnectedId)) {
        partnerId = relations.get(disconnectedId);
        relations.delete(disconnectedId);
    } else {
        // 反向查找
        for (let [key, val] of relations.entries()) {
            if (val === disconnectedId) {
                partnerId = key;
                relations.delete(key);
                break;
            }
        }
    }

    // 通知伙伴
    if (partnerId && clients.has(partnerId)) {
        const partnerWs = clients.get(partnerId);
        partnerWs.send(JSON.stringify({ type: "break", clientId: disconnectedId, targetId: partnerId, message: "209" }));
        // 业务决定是否关闭伙伴连接，通常保持开启等待重连或刷新
        // partnerWs.close(); 
    }

    clients.delete(disconnectedId);
    // 清理定时器
    clientTimers.forEach((val, key) => {
        if (key.startsWith(disconnectedId)) clearInterval(val);
    });
}

function isBound(id) {
    return relations.has(id) || [...relations.values()].includes(id);
}

// 心跳定时器
setInterval(() => {
    if (clients.size > 0) {
        clients.forEach((ws, id) => {
            if(ws.readyState === WebSocket.OPEN) {
                heartbeatMsg.clientId = id;
                heartbeatMsg.targetId = relations.get(id) || "";
                ws.send(JSON.stringify(heartbeatMsg));
            }
        });
    }
}, 60000);
```

### 7.2 前端核心封装 (JavaScript)
功能：连接 WS、生成二维码、处理反馈、封装发送指令。

```javascript
/**
 * DG-LAB Socket Frontend Client
 * 依赖: qrcode.min.js
 */

let wsConn = null;
let connectionId = ""; // 本机 ID
let targetWSId = "";   // APP ID
const WS_URL = "ws://localhost:9999/"; // 修改为实际地址

// 初始化连接
function initSocket() {
    wsConn = new WebSocket(WS_URL);
    
    wsConn.onopen = () => console.log("WS Connected");
    
    wsConn.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        handleMessage(msg);
    };
    
    wsConn.onclose = () => alert("连接断开");
}

function handleMessage(msg) {
    switch (msg.type) {
        case 'bind':
            if (!msg.targetId) {
                // 1. 收到服务器分配的 ID，生成二维码
                connectionId = msg.clientId;
                generateQRCode(connectionId);
            } else {
                // 2. 收到绑定成功通知
                if (msg.clientId !== connectionId) return; // 安全校验
                targetWSId = msg.targetId;
                console.log("绑定成功，APP ID:", targetWSId);
                document.getElementById("qrcode").style.display = "none"; // 隐藏二维码
            }
            break;
            
        case 'msg':
            // 处理 APP 反馈 (强度/按键)
            if (msg.message.includes("strength")) {
                // 格式: strength-A+B+LimitA+LimitB
                const nums = msg.message.match(/\d+/g).map(Number);
                updateStrengthUI(nums[0], nums[1]);
            } else if (msg.message.includes("feedback")) {
                console.log("收到按键反馈:", msg.message);
            }
            break;
            
        case 'break':
            alert("APP 已断开连接");
            location.reload();
            break;
    }
}

function generateQRCode(clientId) {
    const qrContent = `https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#${WS_URL}${clientId}`;
    // 使用 qrcode.js 生成
    new QRCode(document.getElementById("qrcode"), qrContent);
}

// --- 发送指令封装 ---

/**
 * 发送波形
 * @param {string} channel "A" or "B"
 * @param {string} hexDataStr JSON数组字符串 '["0A...","0B..."]'
 * @param {number} duration 持续秒数
 */
function sendWave(channel, hexDataStr, duration = 5) {
    if (!targetWSId) return;
    const payload = {
        type: "clientMsg",
        message: hexDataStr, // 注意：后端代码会自动添加 "pulse-" 前缀
        time: duration,
        channel: channel,
        clientId: connectionId,
        targetId: targetWSId
    };
    wsConn.send(JSON.stringify(payload));
}

/**
 * 调整强度
 * @param {number} channel 1(A) or 2(B)
 * @param {number} type 1:减, 2:加, 3:归零, 4:设置
 * @param {number} val 变化量或目标值
 */
function setStrength(channel, type, val = 1) {
    if (!targetWSId) return;
    const payload = {
        type: type,
        strength: val,
        message: "set channel",
        channel: channel,
        clientId: connectionId,
        targetId: targetWSId
    };
    wsConn.send(JSON.stringify(payload));
}

/**
 * 清空队列
 */
function clearQueue(channelIdx) {
    if (!targetWSId) return;
    const payload = {
        type: "msg",
        message: `clear-${channelIdx}`,
        clientId: connectionId,
        targetId: targetWSId
    };
    wsConn.send(JSON.stringify(payload));
}

// 启动
initSocket();
```

---

## 8. 开发注意事项

1.  **波形数据格式**：波形必须是 HEX 字符串数组。每条数据代表 100ms。如果数组长度为 10，则总时长为 1秒。
2.  **队列管理**：APP 内部有 500 条数据的队列缓冲区。如果发送频率过高或不清除队列，会导致波形延迟执行。建议在切换不同波形模式时，先发送 `clear` 指令。
3.  **心跳保活**：Socket 连接必须维持心跳，否则 APP 或服务端可能会判定超时断开。
4.  **安全性**：当前的 Socket 协议是明文传输（JSON），建议在生产环境使用 `wss://` (SSL) 并增加额外的鉴权逻辑（虽然 APP 端协议固定，但可以在建立连接前增加 Token 校验）。
5.  **1对N 扩展**：虽然官方示例是 1对1，但可以通过修改 Server 端 `relations` 数据结构（如 `Map<clientId, Set<targetId>>`）来实现群控功能。