# 郊狼+井字棋

郊狼（DB-LAB）是一款支持物联网的情趣电击器，输出包括A、B两个通道，用户可以自定义每个通道的波形、强度。官方支持websocket连接。

开发时必须遵循 <important_rules>：

<important_rules>
- 使用yarn管理js相关依赖
- 每完成一个重大功能或里程碑后，必须检查 memory-bank/arch.md 与当前实际实现不同的地方，进行更新， **并且为当前已实现的内容增加具体的描述**
- 每完成 impl-plan.md 的其中一步后必须更新文件中的todolist
- 无论是前端还是后端，必须拆分文件、按模块化的思路开发
- 对功能进行任何操作前（增删改），必须使用tree工具查看目录结构，看是否与预期的匹配
- 必须编写单元测试，每一个模块代码增加和修改后要运行单元测试
- 能用yarn装的第三方库尽量就用第三方库，避免自己造轮子
- 要有完善的日志输出，方便debug
- 添加依赖的时候，不要手动编辑包管理器文件（如package.json、go.mod等），应该使用安装命令（如yarn add、go get）等获取最新版本，包括框架的init等操作也是要用对应的命令
</important_rules>

以下 <documents> 中的文档每次编程前必读：

<documents>
- @memory-bank/dg-lab/dg-lab.md // 郊狼websocket开发文档
- @memory-bank/arch.md // 项目架构文档
- @memory-bank/impl-plan.md // 项目实施计划
</documents>