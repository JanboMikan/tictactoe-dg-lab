/**
 * DG-LAB 郊狼控制最小实现
 * 运行后输出 ws 链接和 uuid，通过命令行输入数字调整强度
 */
const WebSocket = require('ws');
const { v4: uuidv4 } = require('uuid');
const readline = require('readline');

// 配置
const PORT = 9999;

// 存储连接对象: Map<clientId, WebSocket>
const clients = new Map();
// 存储绑定关系: Map<clientId, targetId> (控制端 -> APP端)
const relations = new Map();

// 控制端和APP端的ID
let controlClientId = '';
let appTargetId = '';

const wss = new WebSocket.Server({ port: PORT });

console.log('='.repeat(60));
console.log('DG-LAB 郊狼控制服务启动');
console.log('='.repeat(60));

// 创建控制端连接（模拟）
controlClientId = uuidv4();

// 输出连接信息
const wsUrl = `ws://localhost:${PORT}`;
console.log('\n连接信息：');
console.log(`WebSocket URL: ${wsUrl}`);
console.log(`Client ID: ${controlClientId}`);
console.log('\n请使用 DG-LAB APP 扫描包含以下内容的二维码：');
console.log(`https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#${wsUrl}/${controlClientId}`);
console.log('\n等待 APP 连接...\n');
console.log('='.repeat(60));

// WebSocket 服务器处理
wss.on('connection', function connection(ws) {
    const clientId = uuidv4();
    console.log(`\n[连接] 新连接建立: ${clientId}`);
    clients.set(clientId, ws);

    // 发送 ID 给客户端（APP端）
    ws.send(JSON.stringify({
        type: 'bind',
        clientId,
        message: 'targetId',
        targetId: ''
    }));

    ws.on('message', function incoming(messageStr) {
        console.log(`\n[收到消息] ${messageStr.toString()}`);
        let data = null;

        try {
            data = JSON.parse(messageStr);
        } catch (e) {
            console.error('[错误] 非法 JSON 格式');
            ws.send(JSON.stringify({
                type: 'msg',
                clientId: '',
                targetId: '',
                message: '403'
            }));
            return;
        }

        const { type, clientId: senderId, targetId, message } = data;

        // 处理绑定请求
        if (type === 'bind') {
            handleBind(ws, senderId, targetId);
        }
        // 处理反馈消息（APP -> 控制端）
        else if (type === 'msg') {
            if (message && message.includes('strength')) {
                console.log(`[强度反馈] ${message}`);
            } else if (message && message.includes('feedback')) {
                console.log(`[按键反馈] ${message}`);
            }
        }
    });

    ws.on('close', () => {
        console.log(`\n[断开] 连接断开: ${clientId}`);
        handleDisconnect(ws, clientId);
    });

    ws.on('error', (err) => {
        console.error(`[错误] WebSocket 错误:`, err.message);
    });
});

// 处理绑定
function handleBind(ws, clientId, targetId) {
    // APP 发起绑定请求
    // clientId 是从二维码获取的控制端 ID，targetId 是 APP 自己的 ID
    if (clientId === controlClientId) {
        appTargetId = targetId;
        relations.set(controlClientId, appTargetId);

        const successMsg = {
            type: 'bind',
            clientId: controlClientId,
            targetId: appTargetId,
            message: '200'
        };

        ws.send(JSON.stringify(successMsg));

        console.log(`\n✓ 绑定成功!`);
        console.log(`控制端 ID: ${controlClientId}`);
        console.log(`APP ID: ${appTargetId}`);
        console.log('\n现在可以输入命令控制设备：');
        console.log('  输入数字 0-200 设置 A 通道强度');
        console.log('  输入 a<数字> 设置 A 通道强度（如 a50）');
        console.log('  输入 b<数字> 设置 B 通道强度（如 b50）');
        console.log('  输入 wave 发送基础波形');
        console.log('  输入 clear 清空队列');
        console.log('  输入 quit 退出');
        console.log('='.repeat(60) + '\n');

        // 启动命令行交互
        startCLI();
    }
}

// 处理断开连接
function handleDisconnect(ws, disconnectedId) {
    if (disconnectedId === appTargetId) {
        console.log('\n[警告] APP 已断开连接');
        appTargetId = '';
    }

    clients.delete(disconnectedId);

    // 清理绑定关系
    for (let [key, val] of relations.entries()) {
        if (key === disconnectedId || val === disconnectedId) {
            relations.delete(key);
        }
    }
}

// 发送消息到 APP
function sendToApp(message) {
    if (!appTargetId || !clients.has(appTargetId)) {
        console.log('[错误] APP 未连接');
        return false;
    }

    const appWs = clients.get(appTargetId);
    const payload = {
        type: 'msg',
        clientId: controlClientId,
        targetId: appTargetId,
        message: message
    };

    appWs.send(JSON.stringify(payload));
    console.log(`[发送] ${message}`);
    return true;
}

// 设置强度
function setStrength(channel, value) {
    // 使用模式 2（设置为指定值）
    const message = `strength-${channel}+2+${value}`;
    return sendToApp(message);
}

// 发送基础波形
function sendBasicWave(channel = 'A') {
    // 创建一个简单的脉冲波形：100ms 的重复模式
    // 每个字节代表约 12.5ms，8 字节 = 100ms
    const basicPattern = [
        '0A0A0A0A00000000',  // 渐强-停止
        '0F0F0F0F00000000',  // 更强-停止
        '0A0A0A0A0A0A0A0A',  // 持续中等强度
        '05050505050505050',  // 持续弱强度
        '0F0F0F0F0F0F0F0F',  // 持续强脉冲
        '0A050A0500000000',  // 交替模式
    ];

    const message = `pulse-${channel}:${JSON.stringify(basicPattern)}`;
    return sendToApp(message);
}

// 清空队列
function clearQueue(channel) {
    const channelNum = channel === 'A' ? 1 : 2;
    const message = `clear-${channelNum}`;
    return sendToApp(message);
}

// 命令行交互
function startCLI() {
    const rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout,
        prompt: 'dglab> '
    });

    rl.prompt();

    rl.on('line', (line) => {
        const input = line.trim().toLowerCase();

        if (!input) {
            rl.prompt();
            return;
        }

        if (input === 'quit' || input === 'exit') {
            console.log('正在退出...');
            process.exit(0);
        }
        else if (input === 'wave') {
            console.log('[命令] 发送基础波形到 A 通道');
            sendBasicWave('A');
        }
        else if (input.startsWith('wave ')) {
            const channel = input.split(' ')[1].toUpperCase();
            if (channel === 'A' || channel === 'B') {
                console.log(`[命令] 发送基础波形到 ${channel} 通道`);
                sendBasicWave(channel);
            }
        }
        else if (input === 'clear' || input === 'clear a') {
            console.log('[命令] 清空 A 通道队列');
            clearQueue('A');
        }
        else if (input === 'clear b') {
            console.log('[命令] 清空 B 通道队列');
            clearQueue('B');
        }
        else if (input.startsWith('a')) {
            // a50 格式
            const value = parseInt(input.substring(1));
            if (!isNaN(value) && value >= 0 && value <= 200) {
                console.log(`[命令] 设置 A 通道强度为 ${value}`);
                setStrength(1, value);
            } else {
                console.log('[错误] 强度值必须在 0-200 之间');
            }
        }
        else if (input.startsWith('b')) {
            // b50 格式
            const value = parseInt(input.substring(1));
            if (!isNaN(value) && value >= 0 && value <= 200) {
                console.log(`[命令] 设置 B 通道强度为 ${value}`);
                setStrength(2, value);
            } else {
                console.log('[错误] 强度值必须在 0-200 之间');
            }
        }
        else if (!isNaN(parseInt(input))) {
            // 纯数字，默认设置 A 通道
            const value = parseInt(input);
            if (value >= 0 && value <= 200) {
                console.log(`[命令] 设置 A 通道强度为 ${value}`);
                setStrength(1, value);
            } else {
                console.log('[错误] 强度值必须在 0-200 之间');
            }
        }
        else {
            console.log('[帮助] 可用命令：');
            console.log('  <数字>        - 设置 A 通道强度 (0-200)');
            console.log('  a<数字>       - 设置 A 通道强度');
            console.log('  b<数字>       - 设置 B 通道强度');
            console.log('  wave [a|b]    - 发送基础波形（默认 A 通道）');
            console.log('  clear [a|b]   - 清空队列（默认 A 通道）');
            console.log('  quit          - 退出程序');
        }

        rl.prompt();
    });

    rl.on('close', () => {
        console.log('\n再见！');
        process.exit(0);
    });
}

// 心跳保持
setInterval(() => {
    if (clients.size > 0) {
        const heartbeatMsg = {
            type: 'heartbeat',
            clientId: controlClientId,
            targetId: appTargetId || '',
            message: '200'
        };

        clients.forEach((ws, id) => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify(heartbeatMsg));
            }
        });
    }
}, 60000); // 每 60 秒

// 优雅退出
process.on('SIGINT', () => {
    console.log('\n\n正在关闭服务器...');
    wss.close(() => {
        console.log('服务器已关闭');
        process.exit(0);
    });
});
