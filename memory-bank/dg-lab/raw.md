# README

## SOCKET 控制-控制端开源

### 更新

time : 2025-03-15

desc:

新增 ws 服务核心方法讲解-JavaScript

### 更新

time : 2024-09-014

desc :

新增 QA：

[English_QA](QA/Websocket_open_source_QA_English.txt)
[Chinese_QA](QA/Websocket_open_source_QA_Chinese.txt)

### 说明

SOCKET 控制功能，是 DG-LAB APP 通过 Socket 服务连接到外部第三方控制端，控制端通过 SOCKET 向 APP 发送数据指令使郊狼进行脉冲输出的功能。开发者可以通过网页，游戏，脚本或其他终端在局域网环境或公网环境中对郊狼进行控制。

该功能仅支持 郊狼脉冲主机 3.0

### 项目

我们提供的官网示例分为两部分，前端控制部分(逻辑控制，数据展示，行为操作，指令数据生成等)和 SOCKET 后端部分(关系绑定，数据转发等)。

我们设计的方案是 N(APP 终端)-SOCKET 服务-N(第三方终端)的 N 对 N 模式，方便开发者制作的控制端可以同时多人使用。

### 项目结构

/socket/BackEnd(Node) -> SOCKET 控制后端代码，部署文档可见 /socket/BackEnd(Node)/document.txt

/socket/FrontEnd(Html+Css+Js) -> SOCKET 控制前端代码，部署文档可见 /socket/FrontEnd(Html+Css+Js)/document.txt

### 两端连接流程

由于我们设计的方案是 N 对 N 的模式，所以两端需要通过关系绑定的流程来连接到一起。

### 流程图

```
以下是这两张图片的完整纯文字描述。

---

### **图片一：系统架构与数据流向图**

这张图展示了一个基于 Socket 服务的星型网络架构，主要分为上、中、下三层。

**1. 上层：外部终端（客户端组 A）**
包含多个并列的外部设备，它们通过网络连接指向中间的 Socket 服务。
*   **外部终端1**：ID = 0001
*   **外部终端2**：ID = 0002
*   **......** （表示省略的中间终端）
*   **外部终端N**：ID = 000N

**2. 中层：SOCKET服务（核心服务端）**
这是一个大的矩形框，内部包含核心逻辑组件和数据结构：
*   **功能组件（左侧）：**
    *   **数据转发**：负责处理数据的传输。
    *   **心跳包**：负责维持连接的活性。
*   **核心数据结构（右侧）：关系绑定Map**
    *   这是一个键值对表格，用于存储连接映射关系。
    *   **表头**：KEY (外部终端ID) | VALUE (APP终端ID)
    *   **数据行1**：0001 | 1001
    *   **数据行2**：0002 | 1002
    *   **数据行3**：...... | ......
    *   **数据行4**：0003 | 100N
*   **底层逻辑组件（底部）：**
    *   **消息转发关系**
    *   **客户端计时器关系**

**3. 下层：DG-LAB APP（客户端组 B）**
包含多个并列的手机端应用，它们通过网络连接指向中间的 Socket 服务。
*   **DG-LAB APP1**：ID = 1001
*   **DG-LAB APP2**：ID = 1002
*   **......** （表示省略的中间APP）
*   **DG-LAB APP3**：ID = 100N

**连接关系总结：**
*   上层的“外部终端”与下层的“DG-LAB APP”通过中层的“SOCKET服务”进行通信。
*   Socket 服务通过“关系绑定Map”将特定的外部终端 ID（如 0001）与特定的 APP ID（如 1001）进行一一对应，从而实现定向的数据转发。

---

### **图片二：两端连接流程（时序图）**

这张图详细描述了第三方终端、Socket 服务和 DG-LAB APP 三者之间的交互和连接建立过程。

**涉及对象（从左至右）：**
1.  **第三方终端**
2.  **SOCKET服务**
3.  **DG-LAB APP**

**详细流程步骤：**

1.  **连接发起（第三方终端）：**
    *   第三方终端向 SOCKET服务 发送请求：**连接SOCKET服务**。

2.  **ID生成（SOCKET服务）：**
    *   SOCKET服务内部执行：**生成唯一ID，将ID与ws（WebSocket）对象绑定**。

3.  **ID返回（SOCKET服务 → 第三方终端）：**
    *   SOCKET服务向第三方终端发送：**返回ID**。

4.  **二维码生成（第三方终端）：**
    *   第三方终端内部执行：**存储ID，生成二维码(包含ID信息)**。

5.  **扫码获取ID（DG-LAB APP ↔ 第三方终端）：**
    *   DG-LAB APP 执行动作：**APP扫描终端生成的二维码，并存储终端的ID**。（箭头由 APP 指向 终端，表示扫描交互）。

6.  **连接发起（DG-LAB APP）：**
    *   DG-LAB APP 向 SOCKET服务 发送请求：**连接SOCKET服务**。

7.  **ID生成（SOCKET服务）：**
    *   SOCKET服务内部执行：**生成唯一ID，将ID与ws对象绑定**。

8.  **ID返回（SOCKET服务 → DG-LAB APP）：**
    *   SOCKET服务向 DG-LAB APP 发送：**返回ID**。

9.  **存储ID（DG-LAB APP）：**
    *   DG-LAB APP 内部执行：**存储ID**（指存储刚才服务端返回的APP自身的ID）。

10. **发送绑定请求（DG-LAB APP → SOCKET服务）：**
    *   DG-LAB APP 向 SOCKET服务 发送请求：**发送绑定请求 (包含终端ID和APP ID)**。

11. **建立绑定（SOCKET服务）：**
    *   SOCKET服务内部执行：**将终端ID和APP ID形成绑定关系**。

12. **通知绑定结果（SOCKET服务 → 双端）：**
    *   **路径A**：SOCKET服务向第三方终端发送：**返回绑定结果**。
    *   **路径B**：SOCKET服务向 DG-LAB APP 发送：**返回绑定结果**。

13. **执行UI逻辑（双端）：**
    *   **第三方终端**收到结果后执行：**根据绑定结果执行UI逻辑**。
    *   **DG-LAB APP**收到结果后执行：**根据绑定结果执行UI逻辑**。
```

### APP 收信协议

#### 总则

1. 所有的消息全部都是 json 格式
2. json 格式: {"type":"xxx","clientId":"xxx","targetId":"xxx","message":"xxx"}
3. type 指令:
   1. heartbeat -> 心跳包数据
   2. bind -> ID 关系绑定
   3. msg -> 波形下发/强度变化/队列清空等数据指令
   4. break -> 连接断开
   5. error -> 服务错误
4. clientID: 第三方终端 ID
5. targetId: APP ID
6. message: 消息/指令
7. json 数据的字符最大长度为 1950，若超过该长度，APP 收到数据将会丢弃该消息
8. 除 SOCKET 连接时由 SOCKET 向终端返回 ID 的 json 数据 targetId 可以为空外，其他所有指令都必须且仅包含 type,clientId,targetId,message 这 4 个 key，并且 value 不能为空
9. SOCKET 服务生成的 ID 必须保证唯一，长度推荐 32 位(uuidV4)

#### 关系绑定

1. SOCKET 通道与终端绑定：终端或 APP 连接 SOCKET 服务后，生成唯一 ID，并与终端或 APP 的 websocket 对象绑定存储在 Map 中，向终端或 APP 返回 ID

   SOCKET 向终端或 APP 返回的数据: {"type":"bind","clientId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","targetId":"","message":"targetId"}

   终端或 APP 收到 type = bind，message = targetID 时，表明为 SOCKET 服务返回的 clientId 为当前终端或 APP 的 ID，本地保存。

2. 两边终端的关系绑定: DG-LAB APP 将两边终端的 ID 发送给 SOCKET 服务后，服务将两个 ID 绑定存储在 Map 中

   APP 向上发送的 ID 数据: {"type":"bind","clientId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","targetId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","message":"DGLAB"}

   SOCKET 服务收到 type = bind，message = DGLAB，且 clientId，targetId 不为空时，会将 clientId(第三方终端 ID)和 targetId(APP ID)进行绑定。

3. 绑定结果由 SOCKET 服务下发绑定关系的两个 ID 对应的终端或 APP

   SOCKET 下发的结果数据: {"type":"bind","clientId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","targetId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","message":"200"}

   终端或 APP 收到 type = bind，message = 200(或其他指定数据，详细请见错误码)时，执行对应 UI 逻辑

#### 接收强度数据

APP 中的通道强度或强度上限变化时，会向上同步当前最新的通道强度和强度上限。

APP 向上发送强度数据: {"type":"msg","clientId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","targetId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","message":"strength-x+x+x+x"}

SOCKET 根据对应的 ID 将 json 转发给第三方终端，终端收到 type = msg，message = strength-x+x+x+x 的数据时，更新 UI(更新最新的设备通道强度和强度上限)

指令解释:

1. strength-A 通道强度+B 通道强度+A 强度上限+B 强度上限
2. 通道强度和强度上限的值范围在 0 ～ 200

举例：

数据：{"type":"msg","clientId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","targetId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","message":"strength-11+7+100+35"}

解释：strength-11+7+100+35 表示：当前设备 A 通道强度=11，B 通道强度=7，A 通道强度上限=100，B 通道强度上限=35

#### 强度操作

第三方终端要修改设备通道强度时，发送指定的 json 指令。

终端向下发送强度操作数据: {"type":"msg","clientId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","targetId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","message":"strength-x+x+x"}

SOCKET 服务根据对应的 ID 将 json 转发给 APP，APP 收到 type = msg，message = strength-x+x+x 的数据时，执行指定强度变化操作

指令解释:

1. strength-通道+强度变化模式+数值
2. 通道: 1 - A 通道；2 - B 通道
3. 强度变化模式: 0 - 通道强度减少；1 - 通道强度增加；2 - 通道强度变化为指定数值
4. 数值: 范围在(0 ~ 200)的整型

举例：

1. A 通道强度+5 -> strength-1+1+5
2. B 通道强度归零 -> strength-2+2+0
3. B 通道强度-20 -> strength-2+0+20
4. A 通道强度指定为 35 -> strength-1+2+35

- Tips 指令必须严格按照协议编辑，任何非法的指令都会在 APP 端丢弃，不会执行

#### 波形操作

第三方终端要下发通道波形数据时，发送指定的 json 指令

终端向下发送波形数据: {"type":"msg","clientId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","targetId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","message":"pulse-x:[\"xxxxxxxxxxxxxxxx\",\"xxxxxxxxxxxxxxxx\",......,\"xxxxxxxxxxxxxxxx\"]"}

SOCKET 服务根据对应的 ID 将 json 转发给 APP，APP 收到 type = msg，message = pulse-x:[] 的数据时，执行波形输出操作

指令解释:

1. pulse-通道:[波形数据,波形数据,......,波形数据]
2. 通道: A - A 通道；B - B 通道
3. 数据[波形数据,波形数据,......,波形数据]: 数组最大长度为 100，若超出范围则 APP 会丢弃全部数据
4. 波形数据必须是 8 字节的 HEX(16 进制)形式。波形数据详情请参考 [郊狼情趣脉冲主机 V3 的蓝牙协议](../coyote/v3/README_V3.md)

- Tips 每条波形数据代表了 100ms 的数据，所以若每次发送的数据有 10 条，那么就是 1s 的数据，由于网络有一定延时，若要保证波形输出的连续性，建议波形数据的发送间隔略微小于波形数据的时间长度(< 1s)
- Tips 数组最大长度为 100,也就是最多放置 10s 的数据，另外 APP 中的波形队列最大长度为 500，即为 50s 的数据，若后接收到的数据无法全部放入波形队列，多余的部分会丢弃。所以谨慎考虑您的数据长度和数据发送间隔

#### 清空波形队列

APP 中的波形执行是基于波形队列，遵循先进先出的原则，并且队列可以缓存 500 条波形数据(50s 的数据)。

当波形队列中还有尚未执行完的波形数据时，第三方终端希望立刻执行新的波形数据，则需要先将对应通道的波形队列执行清空操作后，再发送波形数据，即可实现立刻执行新的波形数据的需求。

终端向下发送清空波形队列数据: {"type":"msg","clientId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","targetId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","message":"clear-x"}

SOCKET 服务根据对应的 ID 将 json 转发给 APP，APP 收到 type = msg，message = clear-x 的数据时，执行指定通道波形队列清空操作

指令解释:

1. clear-通道
2. 通道: 1 - A 通道；2 - B 通道

- Tips 建议清空波形队列指令下发后，设定一个时间间隔后再下发新的波形数据，避免由于网络波动等原因导致 清空队列指令晚于波形数据执行造成波形数据丢失 的情况

#### APP 反馈

APP 中有多个不同形状的图标按钮，点击可以上发当前按下按钮的指令，第三方终端可以拟定不同形状图标代表的感受状态。

APP 向上发送强度数据: {"type":"msg","clientId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","targetId":"xxxx-xxxxxxxxx-xxxxx-xxxxx-xx","message":"feedback-x"}

SOCKET 根据对应的 ID 将 json 转发给第三方终端，终端收到 type = msg，message = feedback-x 的数据时，更新 UI(显示 APP 用户的反馈)

指令解释:

1. feedback-index
2. index: A 通道 5 个按钮(从左至右)的角标为:0,1,2,3,4;B 通道 5 个按钮(从左至右)的角标为:5,6,7,8,9

- Tips 您可以在自己开发的终端自由拟定每个形状代表了 APP 用户的某种反馈

#### 前端协议(重要)

如果您希望自己开发前端但完全使用我们的后端代码，那么您的前端协议与以上内容有所不同。

<b>请注意：前端协议的消息不能直接发送到 app，会导致无法解析。App 实际收到的消息请看前半部分 APP 收信协议内容解释</b>

1. 强度操作：

   type : 1 -> 通道强度减少; 2 -> 通道强度增加; 3 -> 通道强度归零 ;4 -> 通道强度指定为某个值

   strength: 强度值变化量/指定强度值(当 type 为 1 或 2 时，该值会被强制设置为 1)

   message: 'set channel' 固定不变

   channel: 1 -> A 通道; 2 -> B 通道

   clientId: 终端 ID

   targetId: APP ID

   A 通道强度减 5 : { type : 1,strength: 5,message : 'set channel',channel:1,clientId:xxxx-xxxxxxxxx-xxxxx-xxxxx-xx,targetId:xxxx-xxxxxxxxx-xxxxx-xxxxx-xx }

   B 通道强度加 1 : { type : 2,strength: 1,message : 'set channel',channel:2,clientId:xxxx-xxxxxxxxx-xxxxx-xxxxx-xx,targetId:xxxx-xxxxxxxxx-xxxxx-xxxxx-xx }

   B 通道强度变 0 : { type : 3,strength: 0,message : 'set channel',channel:2,clientId:xxxx-xxxxxxxxx-xxxxx-xxxxx-xx,targetId:xxxx-xxxxxxxxx-xxxxx-xxxxx-xx }

   B 通道强度变 10 : { type : 4,strength: 10,message : 'set channel',channel:2,clientId:xxxx-xxxxxxxxx-xxxxx-xxxxx-xx,targetId:xxxx-xxxxxxxxx-xxxxx-xxxxx-xx }

2. 波形数据:

   后端代码中默认波形数据发送间隔为 200ms，您可以根据您的波形数据来调整后端的波形数据发送间隔(修改后端代码 timeSpace 的变量值)

   type : clientMsg 固定不变

   message : A 通道波形数据(16 进制 HEX 数组 json,具体见上面的协议说明)

   message2 : B 通道波形数据(16 进制 HEX 数组 json,具体见上面的协议说明)

   time1 : A 通道波形数据持续发送时长

   time2 : B 通道波形数据持续发送时长

   clientId: 终端 ID

   targetId: APP ID

3. 清空波形队列:

   type : msg 固定不变

   message: clear-1 -> 清除 A 通道波形队列; clear-2 -> 清除 B 通道波形队列

   clientId: 终端 ID

   targetId: APP ID

#### 终端二维码

第三方终端的二维码必须按照协议指定方式来生成，否则 APP 将无法识别该二维码

第三方终端需要先连接 SOCKET 服务，并收到服务返回的终端 ID，并存储。

二维码内容为: https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#xxxxxxxxx

内容解释:

1. 二维码必须包含我们的 APP 官网下载地址: https://www.dungeon-lab.com/app-download.php
2. 二维码必须包含标签: DGLAB-SOCKET
3. 二维码必须包含 SOCKET 服务地址,且含有终端 ID 信息,且服务地址与 ID 信息之间不得再有其他内容

   举例：

   1. 正确 -> wss://ws.dungeon-lab.cn/xxxx-xxxxxxxxx-xxxxx-xxxxx-xx
   2. 错误 -> wss://ws.dungeon-lab.cn/xxxx/xxxx-xxxxxxxxx-xxxxx-xxxxx-xx

4. 二维码有且仅有两个#来分割 1.2.3.提到的内容，否则 APP 将无法识别内容
5. 二维码除以上描述的必须包含的内容外，不可再涉及其他内容，否则 APP 可能无法识别

#### 错误码

200 - 成功

209 - 对方客户端已断开

210 - 二维码中没有有效的 clientID

211 - socket 连接上了，但服务器迟迟不下发 app 端的 id 来绑定

400 - 此 id 已被其他客户端绑定关系

401 - 要绑定的目标客户端不存在

402 - 收信方和寄信方不是绑定关系

403 - 发送的内容不是标准 json 对象

404 - 未找到收信人（离线）

405 - 下发的 message 长度大于 1950

500 - 服务器内部异常

> 如有问题，请咨询service@dungeon-lab.com 或 发起 issues


---


# ws服务核心方法讲解-JavaScript

## 前端链接 Websocket 服务器核心方法

```JavaScript
var connectionId = ""; // 前端页面在本次通信里的唯一ID

var targetWSId = ""; // app在本次通信里的唯一ID

let followAStrength = false; //跟随AB软上限

let followBStrength = false;

var wsConn = null; // 全局ws链接对象

function connectWs() {
    wsConn = new WebSocket("ws://12.34.56.78:9999/"); // 内容请改成您的ws服务器地址

    // ws是一个长链接，所以官方定义了几个状态方便你处理信息，onopen事件是ws链接建立成功之后自动调用的，这里我们只打印状态
    wsConn.onopen = function (event) {
        console.log("WebSocket连接已建立");
    };

    // 接下来我们定义通信协议,ws的消息是通过长链接在链接双方之间互相发送，所以需要我们主动定义通信协议
    wsConn.onmessage = function (event) {
        var message = null;
        try {
            // 获取消息内容
            message = JSON.parse(event.data);
        }
        catch (e) {
            // 消息不符合JSON格式异常处理
            console.log(event.data);
            return;
        }

        // 根据 message.type 进行不同的处理，我们定义了消息体格式，类型都是字符串{type, clientId, targetId, message}
        switch (message.type) {
            case 'bind':// 链接上第一件事就是绑定，首先让前端页面和服务器绑定一个本次通信的唯一id
                if (!message.targetId) {
                    // 链接创建时，ws服务器生成一个本次通信的id，通过cliendId传给前端（这个clientId是给前端使用的）
                    connectionId = message.clientId; // 获取 clientId
                    console.log("收到clientId：" + message.clientId);
                    qrcodeImg.clear();
                    // 通过qrcode.min.js库生成一个二维码，之后通过app扫描来和服务器创建链接（当app扫描这个二维码之后，服务器端会知道app需要和前端页面进行绑定，生成一个新的targetId返回给app，并在服务器里绑定这一对clientId和targetId，在本次连接中作为唯一通讯合法性鉴权的标志，之后每次二者互发消息时，targetId和clientId都必须携带，服务器会鉴定是否为合法消息，防止他人非法向app和前端发送消息，避免被恶意修改强度
                    qrcodeImg.makeCode("https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#ws://12.34.56.78:9999/" + connectionId);
                    //qrcodeImg.makeCode("https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#ws://192.168.3.235:9999/" + connectionId);
                }
                else {
                    if (message.clientId != connectionId) {
                        alert('收到不正确的target消息' + message.message)
                        return;
                    }
                    // 当app扫描了二维码之后，服务器完成了targetId的创建，需要通知前端已完成绑定，前端也保存targetId，然后就可以开始正常通信了（记得在每次发送消息的时候携带上targetId和clientId，把波形内容/强度设置等信息放在message里，type设置为msg）
                    targetWSId = message.targetId;
                    console.log("收到targetId: " + message.targetId + "msg: " + message.message);
                    hideqrcode();
                }
                break;
            case 'break':
                // app断开时，服务器通知前端，结束本次游戏
                if (message.targetId != targetWSId)
                    return;
                showToast("对方已断开，code:" + message.message)
                location.reload();
                break;
            case 'error':
                // 服务器出现异常情况和流程错误时，提醒前端
                if (message.targetId != targetWSId)
                    return;
                console.log(message); // 输出错误信息到控制台
                showToast(message.message); // 弹出错误提示框，显示错误消息
                break;
            case 'msg':
                // 正式通讯开始，消息的编码格式请查看socket/readme.md 中的 APP 收信协议
                // 先定义一个空数组来存储结果
                const result = [];
                if (message.message.includes("strength")) {
                    const numbers = message.message.match(/\d+/g).map(Number);
                    result.push({ type: "strength", numbers });
                    document.getElementById("channel-a").innerText = numbers[0]; //解析a通道强度
                    document.getElementById("channel-b").innerText = numbers[1];//解析b通道强度
                    document.getElementById("soft-a").innerText = numbers[2];//解析a通道强度软上限
                    document.getElementById("soft-b").innerText = numbers[3];//解析b通道强度软上限

                    if (followAStrength && numbers[2] !== numbers[0]) {
                        //开启跟随软上限设置  当收到和缓存不同的软上限值时触发自动设置
                        softAStrength = numbers[2]; // 保存 避免重复发信
                        const data1 = { type: 4, message: `strength-1+2+${numbers[2]}` }
                        sendWsMsg(data1);
                    }
                    if (followBStrength && numbers[3] !== numbers[1]) {
                        softBStrength = numbers[3]
                        const data2 = { type: 4, message: `strength-2+2+${numbers[3]}` }
                        sendWsMsg(data2);
                    }
                }
                else if (message.message.includes("feedback")) {
                    showSuccessToast(feedBackMsg[message.message]);
                }
                break;
            case 'heartbeat':
                //心跳包，用于监听本次通信是否网络不稳定而异常断开
                console.log("收到心跳");
                if (targetWSId !== '') {
                    // 已连接上
                    const light = document.getElementById("status-light");
                    light.style.color = '#00ff37';

                    // 1秒后将颜色设置回 #ffe99d
                    setTimeout(() => {
                        light.style.color = '#ffe99d';
                    }, 1000);
                }
                break;
            default:
                console.log("收到其他消息：" + JSON.stringify(message)); // 输出其他类型的消息到控制台
                break;
        }
    };

    wsConn.onerror = function (event) {
        console.error("WebSocket连接发生错误");
        // websocket提供的方法之一，在这里处理连接错误的情况
    };

    wsConn.onclose = function (event) {
        // websocket提供的方法之一，在这里处理连接关闭之后的操作，比如重置页面设置
        showToast("连接已断开");
    };
}
```

## 后端 Websocket 服务核心方法

```JavaScript
// 必须引入的ws链接库
const WebSocket = require('ws');
const { v4: uuidv4 } = require('uuid');

// 储存所有已连接的client，这包括前端和app，在后续消息转发时需要检查连接是否存在
const clients = new Map();

// 存储通讯关系，clientId是key，targetId是value, 本示例中为允许前端和app建立链接，一对一关系
// 如果要设置成一对N关系，请您修改Map的保存策略（将targetId对应的关系设置成数组而不是键值）
const relations = new Map();

const punishmentDuration = 5; //默认发送时间5秒

const punishmentTime = 1; // 默认一秒发送1次

// 存储客户端和发送计时器关系，每个客户端在发送波形消息时都有一个计时器，这个计时器用于波形数据的下发，比如每秒发送1次，一共发送5秒
const clientTimers = new Map();

// 定义心跳消息，告知所有已连接的client，服务器正常工作，若长时间未收到则自动断开（表示网络异常）
const heartbeatMsg = {
    type: "heartbeat",
    clientId: "",
    targetId: "",
    message: "200"
};

// 定义心跳定时器，独立于发信计时器，只要ws服务启动，此定时器就开始持续工作
let heartbeatInterval;

const wss = new WebSocket.Server({ port: 9999 }); // 定义链接端口，根据需求选择自己服务器上的可用端口

wss.on('connection', function connection(ws) {
    // ws提供的方法，链接时自动调用
    // 生成唯一的标识符clientId
    const clientId = uuidv4();

    console.log('新的 WebSocket 连接已建立，标识符为:', clientId);

    //存储这个clientId
    clients.set(clientId, ws);

    // 发送标识符给客户端（格式固定，双方都必须获取才可以进行后续通信：比如浏览器和APP，服务器仅作为一个状态管理者和消息转发工具）
    ws.send(JSON.stringify({ type: 'bind', clientId, message: 'targetId', targetId: '' }));

    // 服务器监听双方发信并处理消息
    ws.on('message', function incoming(message) {
        console.log("收到消息：" + message)
        let data = null;
        try {
            data = JSON.parse(message);
        }
        catch (e) {
            // 非JSON格式数据处理
            ws.send(JSON.stringify({ type: 'msg', clientId: "", targetId: "", message: '403' }))
            return;
        }

        // 非法消息来源拒绝，clientId和targetId并非绑定关系
        if (clients.get(data.clientId) !== ws && clients.get(data.targetId) !== ws) {
            ws.send(JSON.stringify({ type: 'msg', clientId: "", targetId: "", message: '404' }))
            return;
        }

        if (data.type && data.clientId && data.message && data.targetId) {
            // 优先处理clientId和targetId的绑定关系
            const { clientId, targetId, message, type } = data;
            switch (data.type) {
                case "bind":
                    // 服务器下发绑定关系
                    if (clients.has(clientId) && clients.has(targetId)) {
                        // relations的双方都不存在这俩id才能绑定，防止app绑定多个前端
                        if (![clientId, targetId].some(id => relations.has(id) || [...relations.values()].includes(id))) {
                            relations.set(clientId, targetId)
                            const client = clients.get(clientId);
                            const sendData = { clientId, targetId, message: "200", type: "bind" }
                            ws.send(JSON.stringify(sendData));
                            client.send(JSON.stringify(sendData));
                        }
                        else {
                            // 此id已被绑定 拒绝再次绑定
                            const data = { type: "bind", clientId, targetId, message: "400" }
                            ws.send(JSON.stringify(data))
                            return;
                        }
                    } else {
                        const sendData = { clientId, targetId, message: "401", type: "bind" }
                        ws.send(JSON.stringify(sendData));
                        return;
                    }
                    break;
                     // 正式通讯开始，消息的编码格式请查看socket/readme.md 中的 APP 收信协议
                case 1:
                case 2:
                case 3:
                    // clientId请求调节targetId的强度，服务器审核链接合法后下发APP强度调节
                    if (invalidRelation(cliendId, targetId, ws)) return; // 鉴定是否为绑定关系
                        const client = clients.get(targetId);
                        const sendType = data.type - 1;
                        const sendChannel = data.channel ? data.channel : 1;
                        const sendStrength = data.type >= 3 ? data.strength : 1 //增加模式强度改成1
                        const msg = "strength-" + sendChannel + "+" + sendType + "+" + sendStrength;
                        const sendData = { type: "msg", clientId, targetId, message: msg }
                        client.send(JSON.stringify(sendData));
                    break;
                case 4:
                    // clientId请求指定targetId的强度，服务器审核链接合法后下发指定APP强度
                    if (invalidRelation(cliendId, targetId, ws)) return; // 鉴定是否为绑定关系

                        const client = clients.get(targetId);
                        const sendData = { type: "msg", clientId, targetId, message }
                        client.send(JSON.stringify(sendData));

                    break;
                case "clientMsg":
                    // clientId发送给targetId的波形消息，服务器审核链接合法后下发给客户端的消息
                    if (invalidRelation(cliendId, targetId, ws)) return; // 鉴定是否为绑定关系

                    if (!data.channel) {
                        // 240531.现在必须指定通道(允许一次只覆盖一个正在播放的波形)
                        const data = { type: "error", clientId, targetId, message: "406-channel is empty" }
                        ws.send(JSON.stringify(data))
                        return;
                    }

                        //消息体 默认最少一个波形消息
                        let sendtime = data.time ? data.time : punishmentDuration; // AB通道的执行时间
                        const target = clients.get(targetId); //发送到目标app
                        const sendData = { type: "msg", clientId, targetId, message: "pulse-" + data.message }
                        let totalSends = punishmentTime * sendtime;
                        const timeSpace = 1000 / punishmentTime;

                        if (clientTimers.has(clientId + "-" + data.channel)) {
                            // A通道计时器尚未工作完毕, 清除计时器且发送清除APP队列消息，延迟150ms重新发送新数据
                            // 新消息覆盖旧消息逻辑，在多次触发波形输出的情况下，新的波形会覆盖旧的波形
                            console.log("通道" + data.channel + "覆盖消息发送中，总消息数：" + totalSends + "持续时间A：" + sendtime)
                            ws.send("当前通道" + data.channel + "有正在发送的消息，覆盖之前的消息")

                            const timerId = clientTimers.get(clientId + "-" + data.channel);
                            clearInterval(timerId); // 清除定时器
                            clientTimers.delete(clientId + "-" + data.channel); // 清除 Map 中的对应项

                            // 由于App中存在波形队列，保证波形的播放顺序正确，因此新波形覆盖旧波形之前需要发送APP波形队列清除指令
                            switch (data.channel) {
                                case "A":
                                    const clearDataA = { clientId, targetId, message: "clear-1", type: "msg" }
                                    target.send(JSON.stringify(clearDataA));
                                    break;

                                case "B":
                                    const clearDataB = { clientId, targetId, message: "clear-2", type: "msg" }
                                    target.send(JSON.stringify(clearDataB));
                                    break;
                                default:
                                    break;
                            }

                            setTimeout(() => {
                                delaySendMsg(clientId, ws, target, sendData, totalSends, timeSpace, data.channel);
                            }, 150);
                        }
                        else {
                            // 如果不存在未发完的波形消息，无需清除波形队列，直接发送
                            delaySendMsg(clientId, ws, target, sendData, totalSends, timeSpace, data.channel);
                            console.log("通道" + data.channel +"消息发送中，总消息数：" + totalSends + "持续时间：" + sendtime)
                        }

                    break;

                default:
                    // 未定义的其他消息，一般用作提示消息
                   if (invalidRelation(cliendId, targetId, ws)) return; // 鉴定是否为绑定关系

                        const client = clients.get(clientId);
                        const sendData = { type, clientId, targetId, message }
                        client.send(JSON.stringify(sendData));

                    break;
            }
        }
    });

    ws.on('close', function close() {
        // 连接关闭时，清除对应的 clientId 和 WebSocket 实例
        console.log('WebSocket 连接已关闭');
        // 遍历 clients Map，找到并删除对应的 clientId 条目
        let clientId = '';
        clients.forEach((value, key) => {
            if (value === ws) {
                // 拿到断开的客户端id
                clientId = key;
            }
        });
        console.log("断开的client id:" + clientId)
        relations.forEach((value, key) => {
            if (key === clientId) {
                //网页断开 通知app
                let appid = relations.get(key)
                let appClient = clients.get(appid)
                const data = { type: "break", clientId, targetId: appid, message: "209" }
                appClient.send(JSON.stringify(data))
                appClient.close(); // 关闭当前 WebSocket 连接
                relations.delete(key); // 清除关系
                console.log("对方掉线，关闭" + appid);
            }
            else if (value === clientId) {
                // app断开 通知网页
                let webClient = clients.get(key)
                const data = { type: "break", clientId: key, targetId: clientId, message: "209" }
                webClient.send(JSON.stringify(data))
                webClient.close(); // 关闭当前 WebSocket 连接
                relations.delete(key); // 清除关系
                console.log("对方掉线，关闭" + clientId);
            }
        })
        clients.delete(clientId); //清除ws客户端
        console.log("已清除" + clientId + " ,当前size: " + clients.size)
    });

    ws.on('error', function (error) {
        // 错误处理
        console.error('WebSocket 异常:', error.message);
        // 在此通知用户异常，通过 WebSocket 发送消息给双方
        let clientId = '';
        // 查找当前 WebSocket 实例对应的 clientId
        for (const [key, value] of clients.entries()) {
            if (value === ws) {
                clientId = key;
                break;
            }
        }
        if (!clientId) {
            console.error('无法找到对应的 clientId');
            return;
        }
        // 构造错误消息
        const errorMessage = 'WebSocket 异常: ' + error.message;

        relations.forEach((value, key) => {
            // 遍历关系 Map，找到并通知没掉线的那一方
            if (key === clientId) {
                // 通知app
                let appid = relations.get(key)
                let appClient = clients.get(appid)
                const data = { type: "error", clientId: clientId, targetId: appid, message: "500" }
                appClient.send(JSON.stringify(data))
            }
            if (value === clientId) {
                // 通知网页
                let webClient = clients.get(key)
                const data = { type: "error", clientId: key, targetId: clientId, message: errorMessage }
                webClient.send(JSON.stringify(data))
            }
        })
    });

    // 启动心跳定时器（如果尚未启动）
    if (!heartbeatInterval) {
        heartbeatInterval = setInterval(() => {
            // 遍历 clients Map（大于0个链接），向每个客户端发送心跳消息
            if (clients.size > 0) {
                console.log(relations.size, clients.size, '发送心跳消息：' + new Date().toLocaleString());
                clients.forEach((client, clientId) => {
                    heartbeatMsg.clientId = clientId;
                    heartbeatMsg.targetId = relations.get(clientId) || '';
                    client.send(JSON.stringify(heartbeatMsg));
                });
            }
        }, 60 * 1000); // 每分钟发送一次心跳消息
    }
});

function delaySendMsg(clientId, client, target, sendData, totalSends, timeSpace, channel) {
    // 发信计时器 通道会分别发送不同的消息和不同的数量 必须等全部发送完才会取消这个消息 新消息可以覆盖
    // 波形消息由这个计时器来控制按时间发送，波形长度1秒，比如默认输出波形5秒，就需要按顺序向app发送5次，timeSpace设置为1000ms
    // 如果您在前端定义的波形长度不是1秒，那么您就需要控制这个计时器发信的延迟timeSpace，防止波形被覆盖播放

    target.send(JSON.stringify(sendData)); // 计时器开始，立即发送第一次通道的消息
    totalSends--; // 发信总数
    if (totalSends > 0) {
        return new Promise((resolve, reject) => {
            // 按设定频率发送消息给特定的客户端
            const timerId = setInterval(() => {
                if (totalSends > 0) {
                    target.send(JSON.stringify(sendData));
                    totalSends--;
                }
                // 如果达到发信总数，则停止定时器
                if (totalSends <= 0) {
                    clearInterval(timerId);
                    client.send("发送完毕")
                    clientTimers.delete(clientId); // 删除对应的定时器
                    resolve();
                }
            }, timeSpace); // 每隔频率倒数触发一次定时器，一般和波形长度一致

            // 存储clientId与其对应的timerId和波形通道，为了下次收信时确认是否还有正在发送的波形，若有则覆盖，防止多个计时器争抢发信
            clientTimers.set(clientId + "-" + channel, timerId);
        });
    }
}

function invalidRelation(cliendId, targetId, ws) {
    // 关系合法性鉴定，clientId和targetId必须存在于客户端集合中，且在relation集合中绑定了关系
    if (relations.get(clientId) !== targetId) {
        const data = { type: "bind", clientId, targetId, message: "402" }
        ws.send(JSON.stringify(data))
        return true;
    }
    if (!clients.has(clientId) || !clients.has(targetId)) {
        console.log(`未找到匹配的客户端，clientId: ${clientId}`);
        const data = { type: "bind", clientId, targetId, message: "404" }
        ws.send(JSON.stringify(data))
        return true;
    }
    return false;
}

```

---

# Websocket_open_source_QA_Chinese.txt

【中文】
这篇文档简单的给您创建属于您自己的websocket站点以及解决一些websoket相关的疑惑

Q：DG-LAB APP中的Websocket可以做什么？它如何与Coyote v3一起工作？

A：Websocket是一个很常用的消息转发协议，首先，APP与Coyote V3之间是通过蓝牙消息进行通信的，无法通过APP之外的方式发送波形（即便是通过远程控制），因此我们考虑到各种DIY玩法和需求，设计了一个协议让您可以自己搭建Websocket服务器与APP进行网络连接，然后将符合格式的数据通过Websocket向APP发送，再由APP解析发送给Coyote V3执行。我们写的示范游戏就是一个经典例子，（https://www.dungeon-lab.com/t-rex-runner/index.html），为了完成这个示范，我们搭建了一个测试用的websocket服务器（wss://ws.dungeon-lab.cn/），考虑到您的隐私和安全性，我们不推荐您使用我们的测试服务器来搭建自己的Websocket应用，因为它并不十分可靠。在您打开游戏的首页时，它会自动运行JS脚本向websocket服务器请求属于自己的uuid（这个uuid是为了让服务器识别每个客户端分别是谁，在每次通讯中这都是不一样的），这之后网页会将链接生成一个二维码，您使用APP首页的socket扫码功能就可以扫描它并建立链接。（实际上，APP也向服务器请求了一次，获取到了自己这次聊天中的uuid，服务器会将两者的聊天关系绑定成1对1模式，尽管您可以打开许多个网页和手机APP分别链接到websocket服务器，但您会发现，他们彼此直接的通讯是互不干扰的。）

Q：是否可以将两台coyote v3主机通过websocket链接起来？

A：这需要您有一定的代码编写能力，我们目前的websocket设计逻辑是1对1模式，也就是一个coyote v3主机先通过蓝牙连接手机之后，手机再和websocket服务器连接，通过websocket服务器从网络向手机APP传达指令，最后才由APP转发指令给coyote主机执行。如果您使用我们提供的模板代码，那么您会发现在网页中生成的QR code实际上包含一段uuid，而您使用手机APP扫描这个QR code之后，WS服务器也会为手机生成一个uuid并在WS服务器中存储他们为一对Key-Value关系，在您关闭此次连接之前，这个关系可以让您的APP和WS服务器保持严格的1对1通讯。（即使您开启了多个网页，并分别使用不同的手机连接上不同的coyote v3主机，再扫描QR code建立websocket连接，手机A的发送的数据也不会影响手机B）

Q：如何获取QRcode？如何在我的服务器上安装示范游戏项目？

A：实际上，QRcode是包含了websocket服务器地址信息的，我们目前在官网上提供了一个游戏示范（https://www.dungeon-lab.com/t-rex-runner/index.html），为了完成这个示范，我们搭建了一个测试用的websocket服务器。您看到文档里的QRcode前缀（https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#）是必须存在的，否则APP无法识别出连接的可用性，而在这个前缀之后的内容，比如说您的ws服务器地址是ws://testws.test.com或者wss://testws.test.com，他们都是合法的，WSS协议更加安全。如何知道自己的ws服务器地址需要您查询一下搜索引擎中的相关内容，这需要一定的编码能力。确定好您的websocket服务器地址后，您可以遵循如下步骤尝试在自己的websocket服务器上安装好示范代码：

1、在服务器上安装node.js之后，下载我们的websocketNode.js脚本来运行（地址是https://github.com/DG-LAB-OPENSOURCE/DG-LAB-OPENSOURCE/tree/main/socket/BackEnd(Node)）。
2、在服务器上部署好我们提供的html项目（地址是https://github.com/DG-LAB-OPENSOURCE/DG-LAB-OPENSOURCE/tree/main/socket/FrontEnd(Html%2BCss%2BJs)）。
3、您还需要修改我们提供的FrontEnd(Html+Css+Js)/wsConnection.js这个文件中关于
wsConn = new WebSocket("ws://12.34.56.78:9999/");将它改成您的websocket服务器地址（ws或者wss协议都是可以的，我们推荐使用wss，它更安全）。
4、检查第3步的文件里还有一段代码，qrcodeImg.makeCode("https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#ws://12.34.56.78:9999/" + connectionId); 这里面的地址ws://12.34.56.78:9999/需要修改成您的websocket地址。
5.保存上述修改之后，并使用node来启动websocket服务之后，打开您的网站页面（假如您已经完成了第二步并且能够从互联网或者局域网上正确访问它），您就可以点击在网页右上角的connection获取到属于您自己的QRcode了，实际上打开网页之后，网页会自动运行wsConnection.js脚本以创建一个和websocket服务器的连接，并获取一个本次websocket通讯中属于这个网页的唯一uuid。
6、使用手机APP扫描网页显示的QRcode，APP会链接到websocket服务器获取第二个uuid，这是APP在本次通信中的唯一标识，服务器中运行的websocketNode.js会将第五步和第六步中两个uuid绑定成一个通信关系，其中任何一方关闭websocket链接时，服务器都会结束这次websocket通信。
7、当网页或者APP向websocket服务器发送消息时，websocket服务器就会根据第六步的绑定关系来查询自己接收到的消息需要转发给谁，是的websocket本质上是一个消息转发器。

Q：如何通过一个手机同时控制两台coyote v3主机？

A：这需要您修改我们的websocketNode.js文件，您可以很快的掌握它的内容，它使用JavaScript开发，核心逻辑是每个uuid对应一个设备，而我们的设计理念并没有规定您必须让websocket服务器和app必须是1对1通讯的，您可以通过修改代码逻辑，并重新运行node.js服务器，让每个APP扫描QRcode之后都进入一个“聊天室”，这里我们通过一个比喻来解释：我们提供的代码样例是两个设备通过websocket服务器建立了一个1对1的私聊频道，那么您的需求就是把这个结构改造成一个公共聊天室，或者说1对N的聊天室，那么您可以修改websocketNode.js，让websocket接收到消息时，群发给任何您希望它发送的设备。比如说您可以建立一个Map来存储所有的手机APP扫描后获取到的通讯uuid，然后在收到网页发送过来的消息时（没错，您可以根据uuid识别出这条消息是来自于您的网页而不是手机），将这条消息转发给Map里存储的所有手机APP对应的uuid（只需要遍历这个Map并使用send方法依次向它们对应的client发送消息）

Q：可以在不使用二维码扫描的情况下使用Websocket吗？

A：很遗憾，我们的APP必须通过扫描QRcode才能链接上您的websocket服务器，您必须有一个可以显示QRcode的页面让APP扫描才能链接上，因为coyote v3 只能通过蓝牙链接到手机APP上，当前我们还没有提供直接输入websocket地址的UI界面。


---

# 示例代码

## backend

```javascript
const WebSocket = require('ws');
const { v4: uuidv4 } = require('uuid');

// 储存已连接的用户及其标识
const clients = new Map();

// 存储消息关系
const relations = new Map();

const punishmentDuration = 5; //默认发送时间1秒

const punishmentTime = 1; // 默认一秒发送1次

// 存储客户端和发送计时器关系
const clientTimers = new Map();

// 定义心跳消息
const heartbeatMsg = {
    type: "heartbeat",
    clientId: "",
    targetId: "",
    message: "200"
};

// 定义定时器
let heartbeatInterval;

const wss = new WebSocket.Server({ port: 9999 });

wss.on('connection', function connection(ws) {
    // 生成唯一的标识符
    const clientId = uuidv4();

    console.log('新的 WebSocket 连接已建立，标识符为:', clientId);

    //存储
    clients.set(clientId, ws);

    // 发送标识符给客户端（格式固定，双方都必须获取才可以进行后续通信：比如浏览器和APP）
    ws.send(JSON.stringify({ type: 'bind', clientId, message: 'targetId', targetId: '' }));

    // 监听发信
    ws.on('message', function incoming(message) {
        console.log("收到消息：" + message)
        let data = null;
        try {
            data = JSON.parse(message);
        }
        catch (e) {
            // 非JSON数据处理
            ws.send(JSON.stringify({ type: 'msg', clientId: "", targetId: "", message: '403' }))
            return;
        }

        // 非法消息来源拒绝
        if (clients.get(data.clientId) !== ws && clients.get(data.targetId) !== ws) {
            ws.send(JSON.stringify({ type: 'msg', clientId: "", targetId: "", message: '404' }))
            return;
        }

        if (data.type && data.clientId && data.message && data.targetId) {
            // 优先处理绑定关系
            const { clientId, targetId, message, type } = data;
            switch (data.type) {
                case "bind":
                    // 服务器下发绑定关系
                    if (clients.has(clientId) && clients.has(targetId)) {
                        // relations的双方都不存在这俩id
                        if (![clientId, targetId].some(id => relations.has(id) || [...relations.values()].includes(id))) {
                            relations.set(clientId, targetId)
                            const client = clients.get(clientId);
                            const sendData = { clientId, targetId, message: "200", type: "bind" }
                            ws.send(JSON.stringify(sendData));
                            client.send(JSON.stringify(sendData));
                        }
                        else {
                            const data = { type: "bind", clientId, targetId, message: "400" }
                            ws.send(JSON.stringify(data))
                            return;
                        }
                    } else {
                        const sendData = { clientId, targetId, message: "401", type: "bind" }
                        ws.send(JSON.stringify(sendData));
                        return;
                    }
                    break;
                case 1:
                case 2:
                case 3:
                    // 服务器下发APP强度调节
                    if (relations.get(clientId) !== targetId) {
                        const data = { type: "bind", clientId, targetId, message: "402" }
                        ws.send(JSON.stringify(data))
                        return;
                    }
                    if (clients.has(targetId)) {
                        const client = clients.get(targetId);
                        const sendType = data.type - 1;
                        const sendChannel = data.channel ? data.channel : 1;
                        const sendStrength = data.type >= 3 ? data.strength : 1 //增加模式强度改成1
                        const msg = "strength-" + sendChannel + "+" + sendType + "+" + sendStrength;
                        const sendData = { type: "msg", clientId, targetId, message: msg }
                        client.send(JSON.stringify(sendData));
                    }
                    break;
                case 4:
                    // 服务器下发指定APP强度
                    if (relations.get(clientId) !== targetId) {
                        const data = { type: "bind", clientId, targetId, message: "402" }
                        ws.send(JSON.stringify(data))
                        return;
                    }
                    if (clients.has(targetId)) {
                        const client = clients.get(targetId);
                        const sendData = { type: "msg", clientId, targetId, message }
                        client.send(JSON.stringify(sendData));
                    }
                    break;
                case "clientMsg":
                    // 服务端下发给客户端的消息
                    if (relations.get(clientId) !== targetId) {
                        const data = { type: "bind", clientId, targetId, message: "402" }
                        ws.send(JSON.stringify(data))
                        return;
                    }
                    if (!data.channel) {
                        // 240531.现在必须指定通道(允许一次只覆盖一个正在播放的波形)
                        const data = { type: "error", clientId, targetId, message: "406-channel is empty" }
                        ws.send(JSON.stringify(data))
                        return;
                    }
                    if (clients.has(targetId)) {
                        //消息体 默认最少一个消息
                        let sendtime = data.time ? data.time : punishmentDuration; // AB通道的执行时间
                        const target = clients.get(targetId); //发送目标
                        const sendData = { type: "msg", clientId, targetId, message: "pulse-" + data.message }
                        let totalSends = punishmentTime * sendtime;
                        const timeSpace = 1000 / punishmentTime;

                        if (clientTimers.has(clientId + "-" + data.channel)) {
                            // A通道计时器尚未工作完毕, 清除计时器且发送清除APP队列消息，延迟150ms重新发送新数据
                            // 新消息覆盖旧消息逻辑
                            console.log("通道" + data.channel + "覆盖消息发送中，总消息数：" + totalSends + "持续时间A：" + sendtime)
                            ws.send("当前通道" + data.channel + "有正在发送的消息，覆盖之前的消息")

                            const timerId = clientTimers.get(clientId + "-" + data.channel);
                            clearInterval(timerId); // 清除定时器
                            clientTimers.delete(clientId + "-" + data.channel); // 清除 Map 中的对应项

                            // 发送APP波形队列清除指令
                            switch (data.channel) {
                                case "A":
                                    const clearDataA = { clientId, targetId, message: "clear-1", type: "msg" }
                                    target.send(JSON.stringify(clearDataA));
                                    break;

                                case "B":
                                    const clearDataB = { clientId, targetId, message: "clear-2", type: "msg" }
                                    target.send(JSON.stringify(clearDataB));
                                    break;
                                default:
                                    break;
                            }

                            setTimeout(() => {
                                delaySendMsg(clientId, ws, target, sendData, totalSends, timeSpace, data.channel);
                            }, 150);
                        } 
                        else {
                            // 不存在未发完的消息 直接发送
                            delaySendMsg(clientId, ws, target, sendData, totalSends, timeSpace, data.channel);
                            console.log("通道" + data.channel +"消息发送中，总消息数：" + totalSends + "持续时间：" + sendtime)
                        }
                    } else {
                        console.log(`未找到匹配的客户端，clientId: ${clientId}`);
                        const sendData = { clientId, targetId, message: "404", type: "msg" }
                        ws.send(JSON.stringify(sendData));
                    }
                    break;
                default:
                    // 未定义的普通消息
                    if (relations.get(clientId) !== targetId) {
                        const data = { type: "bind", clientId, targetId, message: "402" }
                        ws.send(JSON.stringify(data))
                        return;
                    }
                    if (clients.has(clientId)) {
                        const client = clients.get(clientId);
                        const sendData = { type, clientId, targetId, message }
                        client.send(JSON.stringify(sendData));
                    } else {
                        // 未找到匹配的客户端
                        const sendData = { clientId, targetId, message: "404", type: "msg" }
                        ws.send(JSON.stringify(sendData));
                    }
                    break;
            }
        }
    });

    ws.on('close', function close() {
        // 连接关闭时，清除对应的 clientId 和 WebSocket 实例
        console.log('WebSocket 连接已关闭');
        // 遍历 clients Map，找到并删除对应的 clientId 条目
        let clientId = '';
        clients.forEach((value, key) => {
            if (value === ws) {
                // 拿到断开的客户端id
                clientId = key;
            }
        });
        console.log("断开的client id:" + clientId)
        relations.forEach((value, key) => {
            if (key === clientId) {
                //网页断开 通知app
                let appid = relations.get(key)
                let appClient = clients.get(appid)
                const data = { type: "break", clientId, targetId: appid, message: "209" }
                appClient.send(JSON.stringify(data))
                appClient.close(); // 关闭当前 WebSocket 连接
                relations.delete(key); // 清除关系
                console.log("对方掉线，关闭" + appid);
            }
            else if (value === clientId) {
                // app断开 通知网页
                let webClient = clients.get(key)
                const data = { type: "break", clientId: key, targetId: clientId, message: "209" }
                webClient.send(JSON.stringify(data))
                webClient.close(); // 关闭当前 WebSocket 连接
                relations.delete(key); // 清除关系
                console.log("对方掉线，关闭" + clientId);
            }
        })
        clients.delete(clientId); //清除ws客户端
        console.log("已清除" + clientId + " ,当前size: " + clients.size)
    });

    ws.on('error', function (error) {
        // 错误处理
        console.error('WebSocket 异常:', error.message);
        // 在此通知用户异常，通过 WebSocket 发送消息给双方
        let clientId = '';
        // 查找当前 WebSocket 实例对应的 clientId
        for (const [key, value] of clients.entries()) {
            if (value === ws) {
                clientId = key;
                break;
            }
        }
        if (!clientId) {
            console.error('无法找到对应的 clientId');
            return;
        }
        // 构造错误消息
        const errorMessage = 'WebSocket 异常: ' + error.message;

        relations.forEach((value, key) => {
            // 遍历关系 Map，找到并通知没掉线的那一方
            if (key === clientId) {
                // 通知app
                let appid = relations.get(key)
                let appClient = clients.get(appid)
                const data = { type: "error", clientId: clientId, targetId: appid, message: "500" }
                appClient.send(JSON.stringify(data))
            }
            if (value === clientId) {
                // 通知网页
                let webClient = clients.get(key)
                const data = { type: "error", clientId: key, targetId: clientId, message: errorMessage }
                webClient.send(JSON.stringify(data))
            }
        })
    });

    // 启动心跳定时器（如果尚未启动）
    if (!heartbeatInterval) {
        heartbeatInterval = setInterval(() => {
            // 遍历 clients Map（大于0个链接），向每个客户端发送心跳消息
            if (clients.size > 0) {
                console.log(relations.size, clients.size, '发送心跳消息：' + new Date().toLocaleString());
                clients.forEach((client, clientId) => {
                    heartbeatMsg.clientId = clientId;
                    heartbeatMsg.targetId = relations.get(clientId) || '';
                    client.send(JSON.stringify(heartbeatMsg));
                });
            }
        }, 60 * 1000); // 每分钟发送一次心跳消息
    }
});

function delaySendMsg(clientId, client, target, sendData, totalSends, timeSpace, channel) {
    // 发信计时器 通道会分别发送不同的消息和不同的数量 必须等全部发送完才会取消这个消息 新消息可以覆盖
    target.send(JSON.stringify(sendData)); //立即发送一次通道的消息
    totalSends--;
    if (totalSends > 0) {
        return new Promise((resolve, reject) => {
            // 按频率发送消息给特定的客户端
            const timerId = setInterval(() => {
                if (totalSends > 0) {
                    target.send(JSON.stringify(sendData));
                    totalSends--;
                }
                // 如果达到发送次数上限，则停止定时器
                if (totalSends <= 0) {
                    clearInterval(timerId);
                    client.send("发送完毕")
                    clientTimers.delete(clientId); // 删除对应的定时器
                    resolve();
                }
            }, timeSpace); // 每隔频率倒数触发一次定时器

            // 存储clientId与其对应的timerId和通道
            clientTimers.set(clientId + "-" + channel, timerId);
        });
    }
}
```

## frontend


wsConnection.js:

```javascript
var connectionId = ""; // 从接口获取的连接标识符

var targetWSId = ""; // 发送目标

var fangdou = 500; //500毫秒防抖

var fangdouSetTimeOut; // 防抖定时器

let followAStrength = false; //跟随AB软上限

let followBStrength = false;

var wsConn = null; // 全局ws链接

const feedBackMsg = {
    "feedback-0": "A通道：○",
    "feedback-1": "A通道：△",
    "feedback-2": "A通道：□",
    "feedback-3": "A通道：☆",
    "feedback-4": "A通道：⬡",
    "feedback-5": "B通道：○",
    "feedback-6": "B通道：△",
    "feedback-7": "B通道：□",
    "feedback-8": "B通道：☆",
    "feedback-9": "B通道：⬡",
}

const waveData = {
    "1": `["0A0A0A0A00000000","0A0A0A0A0A0A0A0A","0A0A0A0A14141414","0A0A0A0A1E1E1E1E","0A0A0A0A28282828","0A0A0A0A32323232","0A0A0A0A3C3C3C3C","0A0A0A0A46464646","0A0A0A0A50505050","0A0A0A0A5A5A5A5A","0A0A0A0A64646464"]`,
    "2": `["0A0A0A0A00000000","0D0D0D0D0F0F0F0F","101010101E1E1E1E","1313131332323232","1616161641414141","1A1A1A1A50505050","1D1D1D1D64646464","202020205A5A5A5A","2323232350505050","262626264B4B4B4B","2A2A2A2A41414141"]`,
    "3": `["4A4A4A4A64646464","4545454564646464","4040404064646464","3B3B3B3B64646464","3636363664646464","3232323264646464","2D2D2D2D64646464","2828282864646464","2323232364646464","1E1E1E1E64646464","1A1A1A1A64646464"]`
}

function connectWs() {
    wsConn = new WebSocket("ws://12.34.56.78:9999/");
    //wsConn = new WebSocket("ws://localhost:9999/");
    wsConn.onopen = function (event) {
        console.log("WebSocket连接已建立");
    };

    wsConn.onmessage = function (event) {
        var message = null;
        try {
            message = JSON.parse(event.data);
        }
        catch (e) {
            console.log(event.data);
            return;
        }

        // 根据 message.type 进行不同的处理
        switch (message.type) {
            case 'bind':
                if (!message.targetId) {
                    //初次连接获取网页wsid
                    connectionId = message.clientId; // 获取 clientId
                    console.log("收到clientId：" + message.clientId);
                    qrcodeImg.clear();
                    qrcodeImg.makeCode("https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#ws://12.34.56.78:9999/" + connectionId);
                    //qrcodeImg.makeCode("https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#ws://192.168.3.235:9999/" + connectionId);
                }
                else {
                    if (message.clientId != connectionId) {
                        alert('收到不正确的target消息' + message.message)
                        return;
                    }
                    targetWSId = message.targetId;
                    document.getElementById("status").innerText = "已连接";
                    document.getElementById("status").classList.remove("red");
                    document.getElementById("status-light").classList.remove("red");
                    document.getElementById("status-btn").innerText = "断开";
                    document.getElementById("status-btn").classList.add("red-background");
                    console.log("收到targetId: " + message.targetId + "msg: " + message.message);
                    hideqrcode();
                }
                break;
            case 'break':
                //对方断开
                if (message.targetId != targetWSId)
                    return;
                showToast("对方已断开，code:" + message.message)
                location.reload();
                break;
            case 'error':
                if (message.targetId != targetWSId)
                    return;
                console.log(message); // 输出错误信息到控制台
                showToast(message.message); // 弹出错误提示框，显示错误消息
                break;
            case 'msg':
                // 定义一个空数组来存储结果
                const result = [];
                if (message.message.includes("strength")) {
                    const numbers = message.message.match(/\d+/g).map(Number);
                    result.push({ type: "strength", numbers });
                    document.getElementById("channel-a").innerText = numbers[0];
                    document.getElementById("channel-b").innerText = numbers[1];
                    document.getElementById("soft-a").innerText = numbers[2];
                    document.getElementById("soft-b").innerText = numbers[3];

                    if (followAStrength && numbers[2] !== numbers[0]) {
                        //开启跟随软上限  当收到和缓存不同的软上限值时触发自动设置
                        softAStrength = numbers[2]; // 保存 避免重复发信
                        const data1 = { type: 4, message: `strength-1+2+${numbers[2]}` }
                        sendWsMsg(data1);
                    }
                    if (followBStrength && numbers[3] !== numbers[1]) {
                        softBStrength = numbers[3]
                        const data2 = { type: 4, message: `strength-2+2+${numbers[3]}` }
                        sendWsMsg(data2);
                    }
                }
                else if (message.message.includes("feedback")) {
                    showSuccessToast(feedBackMsg[message.message]);
                }
                break;
            case 'heartbeat':
                //心跳包
                console.log("收到心跳");
                if (targetWSId !== '') {
                    // 已连接上
                    const light = document.getElementById("status-light");
                    light.style.color = '#00ff37';

                    // 1秒后将颜色设置回 #ffe99d
                    setTimeout(() => {
                        light.style.color = '#ffe99d';
                    }, 1000);
                }
                break;
            default:
                console.log("收到其他消息：" + JSON.stringify(message)); // 输出其他类型的消息到控制台
                break;
        }
    };

    wsConn.onerror = function (event) {
        console.error("WebSocket连接发生错误");
        // 在这里处理连接错误的情况
    };

    wsConn.onclose = function (event) {
        showToast("连接已断开");
    };
}

// 自动链接
connectWs();

function sendWsMsg(messageObj) {
    messageObj.clientId = connectionId;
    messageObj.targetId = targetWSId;
    if (!messageObj.hasOwnProperty('type'))
        messageObj.type = "msg";
    wsConn.send(JSON.stringify((messageObj)));
}

function toggleSwitch(id) {
    const element = document.getElementById(id);
    element.classList.toggle("switch-on");
    element.classList.toggle("switch-off");
}

function addOrIncrease(type, channelIndex, strength) {
    // 1 减少一  2 增加一  3 设置到
    // channel:1-A    2-B
    // 获取当前频道元素和当前值
    const channelElement = document.getElementById(channelIndex === 1 ? "channel-a" : "channel-b");
    let currentValue = parseInt(channelElement.innerText);

    // 如果是设置操作
    if (type === 3) {
        currentValue = 0; //固定为0
    }
    // 减少一
    else if (type === 1) {
        currentValue = Math.max(currentValue - strength, 0);
    }
    // 增加一
    else if (type === 2) {
        currentValue = Math.min(currentValue + strength, 200);
    }

    // 构造消息对象并发送
    const data = { type, strength: currentValue, message: "set channel", channel: channelIndex };
    console.log(data)
    sendWsMsg(data);
}

function clearAB(channelIndex) {
    const data = { type: 4, message: "clear-" + channelIndex }
    sendWsMsg(data);
}

function autoAddStrength(channelId, inputId, currentId, follow) {
    // 检查是否开启跟随软上限
    if (!follow) {
        let addStrength = parseInt(document.getElementById(inputId).value, 10);
        let currentStrength = parseInt(document.getElementById(currentId).innerText, 10);
        let setTo = addStrength + currentStrength;
        if (addStrength > 0) {
            const data = { type: 4, message: `strength-${channelId}+2+${setTo}` }
            sendWsMsg(data);
        }
    }
}

function sendCustomMsg() {
    if (fangdouSetTimeOut) {
        return;
    }

    autoAddStrength(1, "failed-a", "channel-a", followAStrength); // 给A通道加强度
    autoAddStrength(2, "failed-b", "channel-b", followBStrength); // 给B通道加强度

    const selectA = document.getElementById("wave-a").value;
    const selectB = document.getElementById("wave-b").value;
    const timeA = parseInt(document.getElementById("time-a").value, 10);
    const timeB = parseInt(document.getElementById("time-b").value, 10);

    const msg1 = `A:${waveData[selectA]}`;
    const msg2 = `B:${waveData[selectB]}`;

    const dataA = { type: "clientMsg", message: msg1, time: timeA, channel: "A" }
    const dataB = { type: "clientMsg", message: msg2, time: timeB, channel: "B" }
    sendWsMsg(dataA)
    sendWsMsg(dataB)

    fangdouSetTimeOut = setTimeout(() => {
        clearTimeout(fangdouSetTimeOut);
        fangdouSetTimeOut = null;
    }, fangdou);

}

function showToast(message) {
    let notyf = new Notyf();
    // Display a success notification
    //notyf.success(message);

    notyf.error(message);
}

function showSuccessToast(message) {
    let notyf = new Notyf();
    notyf.success(message);
}

function toggleSwitch(id) {
    // 获取开关元素 并切换开关状态
    const container = document.getElementById(id);
    container.classList.toggle('on');
    const switch1State = container.classList.contains('on');
    followAStrength = id === 'toggle1' ? switch1State : followAStrength;
    followBStrength = id === 'toggle2' ? switch1State : followBStrength;

    const currentStrength = parseInt(document.getElementById(id === 'toggle1' ? 'channel-a' : 'channel-b').innerText);
    const currentSoft = parseInt(document.getElementById(id === 'toggle1' ? 'soft-a' : 'soft-b').innerText);

    console.log(switch1State + '@' + currentStrength + '@' + currentSoft)

    if (switch1State && currentStrength !== currentSoft) {
        //马上判断是否和软上限符合
        console.log('不符合 马上变化')
        const channel = id === 'toggle1' ? 1 : 2;
        const data = { type: 4, message: `strength-${channel}+2+${currentSoft}` }
        sendWsMsg(data);
    }
}

function connectOrDisconn() {
    // 如果未连接则显示二维码
    if (wsConn && targetWSId === '') {
        showqrcode();
        return;
    } else {
        wsConn.close();
        showToast("已断开连接");
        location.reload();
    }
}
```

index.html

```html
<!doctype html>
<html>

<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0,maximum-scale=1.0, user-scalable=no">
    <title>T-Rex Runner</title>
    <link rel="stylesheet" href="notyf.min.css">
    <script src="notyf.min.js"></script>
    <script src="qrcode.min.js"></script>

    <link rel="stylesheet" href="index.css?v=2404221746">
    <script src="index.js?v=2404221746"></script>
    <script src="wsConnection.js?v=2404221746"></script>
</head>

<body id="t" class="offline">
    <div id="qrcode-overlay">
        <div id="qrcode-container">
            <div id="qrcode-text">
                <p>使用DG-LAB APP扫码建立WebSocket链接</p>
            </div>
            <div id="qrcode"></div>
            <div class="close-qrcode" onclick="hideqrcode()">关闭</div>
        </div>
    </div>
    <div class="header-container">
        <div class="dg-controller">
            <div class="btn-container">
                <!-- 小提示消息框 -->
                <div class="tooltip" id="tooltip"></div>

                <span>A通道: 当前强度:</span>
                <span id="channel-a">0</span>
                <button onclick="addOrIncrease(1, 1, 1)"> 强度-1 </button>
                <button onclick="addOrIncrease(2, 1, 1)"> 强度+1 </button>
                <button onclick="addOrIncrease(3, 1)"> 强度置0 </button>
                <span>软上限: </span><span id="soft-a">0</span>
                <span>强度跟随软上限</span>
                <img src="question.svg" class="question-img" id="question1" />
                <div class="toggle-container" id="toggle1" onclick="toggleSwitch('toggle1')">
                    <div class="toggle-switch"></div>
                </div>
                <span>波形</span>
                <select id="wave-a">
                    <option value="1">波形A</option>
                    <option value="2">波形B</option>
                    <option value="3">波形C</option>
                </select>
                <span>失败增加强度</span>
                <img src="question.svg" class="question-img" id="question2" />
                <select id="failed-a">
                    <option value="0" selected>+0</option>
                    <option value="1">+1</option>
                    <option value="2">+2</option>
                    <option value="3">+3</option>
                    <option value="4">+4</option>
                    <option value="5">+5</option>
                    <option value="6">+6</option>
                    <option value="7">+7</option>
                    <option value="8">+8</option>
                    <option value="9">+9</option>
                    <option value="10">+10</option>
                </select>
                <span>输出时长</span>
                <select id="time-a">
                    <option value="1">1秒</option>
                    <option value="2">2秒</option>
                    <option value="3">3秒</option>
                    <option value="4">4秒</option>
                    <option value="5" selected>5秒</option>
                    <option value="6">6秒</option>
                    <option value="7">7秒</option>
                    <option value="8">8秒</option>
                    <option value="9">9秒</option>
                    <option value="10">10秒</option>
                </select>
                <button onclick="clearAB(1)"> 清除当前波形队列 </button>
                <img src="question.svg" class="question-img" id="question3" />
            </div>
            <div class="btn-container">
                <span>B通道: 当前强度:</span>
                <span id="channel-b">0</span>
                <button onclick="addOrIncrease(1, 2, 1)"> 强度-1 </button>
                <button onclick="addOrIncrease(2, 2, 1)"> 强度+1 </button>
                <button onclick="addOrIncrease(3, 2)"> 强度置0 </button>
                <span>软上限: </span><span id="soft-b">0</span>
                <span>强度跟随软上限</span>
                <img src="question.svg" class="question-img" id="question1" />
                <div class="toggle-container" id="toggle2" onclick="toggleSwitch('toggle2')">
                    <div class="toggle-switch"></div>
                </div>
                <span>波形</span>
                <select id="wave-b">
                    <option value="1">波形A</option>
                    <option value="2">波形B</option>
                    <option value="3">波形C</option>
                </select>
                <span>失败增加强度</span>
                <img src="question.svg" class="question-img" id="question2" />
                <select id="failed-b">
                    <option value="0" selected>+0</option>
                    <option value="1">+1</option>
                    <option value="2">+2</option>
                    <option value="3">+3</option>
                    <option value="4">+4</option>
                    <option value="5">+5</option>
                    <option value="6">+6</option>
                    <option value="7">+7</option>
                    <option value="8">+8</option>
                    <option value="9">+9</option>
                    <option value="10">+10</option>
                </select>
                <span>输出时长</span>
                <select id="time-b">
                    <option value="1">1秒</option>
                    <option value="2">2秒</option>
                    <option value="3">3秒</option>
                    <option value="4">4秒</option>
                    <option value="5" selected>5秒</option>
                    <option value="6">6秒</option>
                    <option value="7">7秒</option>
                    <option value="8">8秒</option>
                    <option value="9">9秒</option>
                    <option value="10">10秒</option>
                </select>
                <button onclick="clearAB(2)"> 清除当前波形队列 </button>
                <img src="question.svg" class="question-img" id="question3" />
            </div>
        </div>
        <div class="status-container">
            <div class="status">
                <span>当前状态: </span>
                <span id="status" class="red">未连接</span>
                <span id="status-light" class="red"> ●</span>
            </div>
            <div class="connect-btn">
                <button onclick="showInfo()">使用说明</button>
                <button onclick="connectOrDisconn()" id="status-btn">连接</button>
            </div>
        </div>
        <div class="information-overlay" id="information-overlay">
            <div class="information">
                <h2>使用说明</h2>
                <p>
                    这是郊狼Socket控制官方demo，<span class="notify">本功能目前只支持郊狼3.0连接</span>。
                    <br><br>
                    Socket控制模式下，手机APP通过Socket服务连接到第三方控制端，并将收到的数据转发给郊狼。从而允许开发者通过网页、游戏或其他终端对郊狼进行控制。和远程控制类似，您也可以为本模式设置强度上限保护。
                    <br><br>
                    当前页面是我们提供的Web游戏demo，当小恐龙碰到障碍物时，就会控制郊狼输出脉冲。
                    <br><br>
                    要开始游戏，首先点击右上角<span class="notify">【连接】</span>按钮生成二维码。然后在APP中选择右下角<span class="notify">【Socket控制】</span>模式，连接设备后点击<span class="notify">【连接socket服务器】</span>按钮扫码完成连接。
                    <br><br>
                    我们已经在GitHub上开源相关协议和示例代码，以便开发者们自行搭建服务器进行控制，详情请访问链接：<span class="notify"><a href="https://github.com/DG-LAB-OPENSOURCE/DG-LAB-OPENSOURCE">点此跳转GitHub</a></span>
                    <br><br>
                    需要注意的是，Socket控制协议是开放的，我们无法对第三方控制端进行审核。<span class="notify"><strong>请不要扫描来源不明的Socket二维码</strong></span>。在您准备进行连接前，<span class="notify"><strong>请确保您已经设置好强度上限保护</strong></span>。
                </p>
                <div class="info-close">
                    <button class="info-close-btn" onclick="closeInfo()">我知道了</button>
                </div>
            </div>
        </div>
    </div>
    <div id="main-frame-error" class="interstitial-wrapper">
        <div id="main-content">
            <div class="icon icon-offline" alt=""></div>
        </div>
        <div id="offline-resources">
            <img id="offline-resources-1x" src="assets/default_100_percent/100-offline-sprite.png">
            <img id="offline-resources-2x" src="assets/default_200_percent/200-offline-sprite.png">
            <template id="audio-resources">
                <audio id="offline-sound-press"
                    src="data:audio/mpeg;base64,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"></audio>
                <audio id="offline-sound-hit"
                    src="data:audio/mpeg;base64,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"></audio>
                <audio id="offline-sound-reached"
                    src="data:audio/mpeg;base64,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"></audio>
            </template>
        </div>
    </div>
    <div class="intro-game">
        <div id="messageBox" class="sendmessage">
            <h1 style="text-align: center;font-family: 'Open Sans', sans-serif;">按空格开始游戏</h1>
        </div>
        <div class="game-title">
            <div class="game-tips" id="game-tips">游戏规则</div>
            <button class="tips-hide" onclick="hideOrShowTips()" id="tip-btn">隐藏</button>
        </div>
        <div class="game-content">
            <h4>操作：空格：跳跃/开始游戏，S键：趴下，P：停止游戏<br><br>每当小恐龙碰到障碍物时，执行屏幕上方设置的波形和强度</h4>
            <P>强度设置：点击屏幕上方A/B通道强度增减按钮设置强度，点击[时长]下拉菜单选择发送时长</P>
            <p>提示：Websocket链接断开时游戏会直接结束，游戏异常请刷新页面</p>
        </div>
    </div>
</body>

<script>
    var qrcode = document.getElementById("qrcode-overlay");
    var qrcodeImg = new QRCode(document.getElementById("qrcode"), "https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#ws://39.108.168.199:9999/"); // 全局二维码
    let gamePanelInit = true;

    document.onkeydown = function (evt) {
        evt = evt || window.event;
        if (evt.keyCode == 32 && wsConn != null && targetWSId !== "" && gamePanelInit) {
            evt.preventDefault();
            document.getElementById("messageBox").querySelector('h1').textContent = '';
            document.querySelector('.offline .runner-container').style.top = '10px';
            gamePanelInit = false;
        }
        else if (evt.keyCode == 32) {
            // 阻止默认的空格键行为
            evt.preventDefault();
        }
    };
    function hideqrcode() {
        qrcode.style.visibility = "hidden";
    }
    function showqrcode() {
        qrcode.style.visibility = "visible";
    }

    const questionIcons = document.querySelectorAll('.question-img'); // 获取所有提示图标
    const tooltip = document.getElementById('tooltip');

    const tipsMsg = {
        'question1': '当此通道软上限强度变化时，自动把此通道强度设置为软上限强度。(注意：开启后失败增加强度功能失效)',
        'question2': '当小恐龙碰到障碍物时，此通道强度自动增加设置的值。(默认为+0) 注意：开启强度跟随软上限功能后此功能失效。',
        'question3': '清除当前APP内此通道正在执行的波形数据，避免新的波形数据被之前的覆盖。'
    }

    // 鼠标悬停事件
    questionIcons.forEach(icon => {
        icon.addEventListener('mouseover', function (event) {
            const x = event.clientX;
            const y = event.clientY;

            // 根据当前图标ID设置相应的提示文本
            tooltip.innerText = `${tipsMsg[icon.id]}`;

            tooltip.style.display = 'block';
            tooltip.style.left = x + 'px';
            tooltip.style.top = y + 'px';
        });

        // 鼠标离开事件
        icon.addEventListener('mouseout', function (event) {
            tooltip.style.display = 'none'; // 隐藏提示文字
        });
    });

    function showInfo() {
        document.getElementById("information-overlay").style.visibility = "visible";
    }

    function closeInfo() {
        document.getElementById("information-overlay").style.visibility = "hidden";
    }

    function hideOrShowTips() {
        const tips = document.querySelector('.game-content');
        if (tips.style.display === 'none') {
            tips.style.display = 'block';
            document.getElementById('tip-btn').textContent = '隐藏';
            document.getElementById('messageBox').style.display = 'block';
        } else {
            tips.style.display = 'none';
            document.getElementById('tip-btn').textContent = '显示';
            document.getElementById('messageBox').style.display = 'none';
        }
    }

</script>

</html>
```