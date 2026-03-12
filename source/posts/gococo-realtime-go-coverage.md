title: "gococo: 从零构建实时 Go 覆盖率可视化工具"
date: 2026-03-11 10:00:00
summary: "开源一款全新的 Go 实时覆盖率收集与可视化工具 gococo，流式覆盖率事件 + Web UI 实时展示"
---

## 前言

几年前我写过一篇[如何收集 Go 的实时覆盖率](/posts/how-to-collect-coverage-in-go/)，介绍了 goc 的原理。goc 是一款优秀的工具，但它诞生于 Go 1.16 时代，设计上有一些历史包袱。

这次我从零重写了一款新工具 —— **gococo**（Go Coverage Collection Tools），保留了 goc 的核心思想（源码级插桩 + 实时收集），但在架构上做了全面革新：

- **流式事件**：不再轮询，覆盖率事件发生即推送
- **Web UI 实时可视化**：源码级高亮，命中的行实时发光
- **计数器快照**：精确捕获 `init()` 和 `main()` 阶段的覆盖率
- **零外部依赖**：纯 Go server + 前端嵌入二进制，单文件部署

项目地址：**[https://github.com/gococo/gococo](https://github.com/gococo/gococo)**

![gococo 实时覆盖率演示](/img/posts/gococo/demo.gif)

## 回顾：Go 的覆盖率原理

Go 的覆盖率收集本质上是**源码级插桩**。编译器把代码划分为若干基本块（basic block），在每个基本块的入口插入计数器自增语句。例如：

```go
func classify(n int) string {
    GoCover.Count[0]++
    if n <= 0 {
        GoCover.Count[1]++
        return "non-positive"
    }
    GoCover.Count[2]++
    return "positive"
}
```

`go test -cover` 正是这么做的。但它有一个根本限制：**覆盖率只在程序退出后才能拿到**。对于长时间运行的服务（HTTP server、微服务），你必须停掉进程才能得到报告。

goc 解决了这个问题，它为每个 main 包注入了一个 HTTP API，可以在运行时拉取计数器。但 goc 的方案是"拉"模式——客户端主动请求覆盖率，时效性取决于轮询频率。

**gococo 的做法完全不同：推模式，事件级粒度。**

## gococo 的架构

```
┌──────────────┐   instrument    ┌──────────────────┐
│  Go Project  │ ──────────────► │ Instrumented Bin  │
└──────────────┘   gococo build  └────────┬─────────┘
                                          │ events (HTTP stream)
                                          ▼
                                 ┌──────────────────┐   SSE
                                 │  gococo server   │ ──────► Web UI
                                 └──────────────────┘
```

三个角色：

1. **`gococo build`** —— 编译时插桩
2. **Instrumented Binary** —— 运行时推送事件
3. **`gococo server`** —— 接收事件 + 服务 Web UI

下面逐一展开。

## 插桩：AST 重写

和 goc 一样，gococo 使用 Go 的 `go/ast` + `go/parser` 解析源码，找到每个基本块并注入代码。但注入的内容不同：

```go
GococoCov_RAND_FILEIDX[blockIdx]++; GococoEmit_RAND(fileIdx, blockIdx);
```

每条插桩语句做两件事：

1. **计数器自增**：`GococoCov_RAND_FILEIDX[blockIdx]++`，这是覆盖率的 ground truth，永远不会丢失
2. **事件发射**：`GococoEmit_RAND(fileIdx, blockIdx)`，向 channel 发送一个事件

事件发射函数使用 `select/default` 实现非阻塞写入：

```go
func GococoEmit_RAND(fileIdx int, blockIdx int) {
    if !gococoEnabled_RAND { return }
    select {
    case gococoCh_RAND <- &GococoBlock_RAND{FileIdx: fileIdx, BlockIdx: blockIdx}:
    default: // channel 满了就丢弃，不阻塞业务逻辑
    }
}
```

channel 容量为 8192，平衡了实时性和内存开销。即使 channel 满了导致事件丢失，计数器仍然是准确的。

### 为什么需要两套机制？

- **计数器**：保证准确性。`init()` 和 `main()` 启动期间可能还没建立网络连接，事件会丢失，但计数器不会。
- **事件流**：保证实时性。每次代码块被执行，UI 都能"看到"。

这是 gococo 和 goc 的一个关键区别：goc 只有计数器（拉模式），gococo 同时有计数器和事件流（推模式）。

### dot import 的妙用

插桩后的代码直接引用 `GococoCov_RAND_0`、`GococoEmit_RAND` 等符号，这些符号定义在独立的 `gococodef` 包中。为了让插桩代码不需要包名前缀，gococo 使用了 dot import：

```go
import . "module/gococodef"
```

这样所有导出符号直接进入当前包的命名空间，插桩代码就可以直接写 `GococoCov_RAND_0[i]++` 而无需 `gococodef.GococoCov_RAND_0[i]++`。

## Agent：同步注册 + 异步推流

gococo 在每个 main 包中注入一个 `init()` 函数作为 agent：

```go
func init() {
    host := "127.0.0.1:7778"
    if env := os.Getenv("GOCOCO_HOST"); env != "" {
        host = env
    }

    // 同步注册：连不上就退出
    agentID := registerAgent(host)
    registerBlocks(host, agentID)

    // 异步推流
    go runStreaming(host, agentID)
}
```

几个设计决策：

### 1. 同步注册，fail-fast

`registerAgent()` 最多重试 10 次，全部失败则 `os.Exit(1)`。这是有意为之——如果 server 没有启动，运行插桩后的二进制没有意义，不如立即报错让用户知道。

### 2. Block 元数据预注册

Agent 启动时会发送所有基本块的位置信息（文件名、起止行列、语句数），即使这些块尚未被执行。这样 server 从一开始就知道"总共有多少代码"，覆盖率分母是准确的。

### 3. 计数器快照

Agent 在启动 500ms 后发送一次计数器快照：

```go
func runStreaming(host string, agentID string) {
    time.Sleep(500 * time.Millisecond)
    sendCounterSnapshot(host, agentID)
    // ... then start event streaming
}
```

为什么要等 500ms？因为 `init()` 函数在 `main()` 之前执行，如果快照发得太早，`main()` 的启动逻辑还没跑完，覆盖率数据不完整。500ms 的延迟给了 `main()` 足够的启动时间。

### 4. Chunked HTTP 推流

事件通过 HTTP chunked POST 持续推送到 server，使用 `io.Pipe` + `bufio.Writer` 实现：

```go
func streamEvents(host string, agentID string) error {
    pr, pw := io.Pipe()
    go func() {
        bw := bufio.NewWriter(pw)
        for {
            select {
            case block := <-eventChan:
                // 写入事件：seq|ts|goroutineID|file|blockIdx|sl|sc|el|ec|stmts
                fmt.Fprintf(bw, "%d|%d|%d|%s|%d|%d|%d|%d|%d|%d\n", ...)
            case <-ticker.C:
                bw.Flush() // 每 100ms flush 一次
            }
        }
    }()

    req, _ := http.NewRequest("POST", url, pr)
    req.Header.Set("Transfer-Encoding", "chunked")
    http.DefaultClient.Do(req)
}
```

每个事件携带 goroutine ID，这让 server 可以追踪"哪个 goroutine 执行了哪段代码"。

## Server：接收 + 广播 + 提供 Web UI

gococo server 是一个纯 Go HTTP server，核心逻辑：

1. **接收 agent 事件**：解析 chunked POST 中的覆盖率事件
2. **更新 block state**：维护每个代码块的命中次数和最后命中时间
3. **SSE 广播**：通过 Server-Sent Events 实时推送给所有 Web UI 客户端
4. **服务 Web UI**：前端通过 `go:embed` 嵌入二进制，无需额外部署

Web UI 使用 React + TypeScript 构建，主要功能：

- 源码级覆盖率展示，命中的行实时高亮
- 每行显示命中次数和最后执行时间（如 `x12 18:05:22 (3s ago)`）
- 文件树展示每个文件的覆盖率百分比
- Goroutine 执行流面板，显示每个 goroutine 正在执行的代码

## 快速体验

```bash
# 安装
go install github.com/gococo/gococo/cmd/gococo@latest

# 在项目目录下启动 server
cd /path/to/your/project
gococo server

# 另一个终端：插桩编译
gococo build -o ./myapp-instrumented .

# 运行
./myapp-instrumented

# 打开浏览器
open http://127.0.0.1:7778
```

## gococo vs goc

|  | goc                          | gococo                  |
|-------|-------|-------|
| 覆盖率获取 | 拉模式（HTTP API 轮询） | 推模式（流式事件） |
| 可视化 | 无（输出 coverprofile） | 内置 Web UI，实时高亮 |
| init/main 覆盖 | 可能遗漏 | 计数器快照捕获 |
| Goroutine 追踪 | 无 | 每个事件携带 goroutine ID |
| 部署 | server + 被测服务需互相可达 | 只需被测服务能访问 server |
| 前端 | 无 | 嵌入二进制，零依赖 |

## 总结

gococo 的核心创新在于**双通道架构**：计数器保证精确性，事件流保证实时性。这让我们可以在不影响程序性能的前提下，实现真正的实时覆盖率可视化。

项目完全开源，欢迎 star 和 contribute：

**[https://github.com/gococo/gococo](https://github.com/gococo/gococo)**
