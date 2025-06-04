title: Python asyncio 与 Go 的横向比较
date: 2025-06-04 10:56:33

---

Python asyncio 出现也十年了，这十年，主流语言都在朝这个方向探索，即在用户态模拟线程的概念，让使用者以类似同步线程的方式编写并发代码。抛开各个底层的线程、协程、纤程、绿色线程等等名词，其实，对上层的使用者来讲，各个语言是非常相似的。这也说明，工业界在并发编程上达成了某些共识，这些共识并不是性能层面的共识，而是一种使用模式的共识。虽然我的并发编程经验只能算 so so，但我始终认为，对大多是人来讲，并发编程的关键绝不在什么高性能，而在于你是否能设计出一个易于理解，不易出错的模式。就像在多个线程中通信的时候，共享内存的出错概率远远大于消息队列。

Python asyncio 这种用户态模拟的 coroutine 就是一个革命性的并发编程方式。在这之前，通常是基于单线程 + 同步非阻塞 + 多路复用来实现高并发服务的，这种模式下，业务逻辑是分散在多个 callback 中，让人头痛。asyncio 让使用者能继续使用多线程编程的习惯，业务代码在视觉上都是同步的，清晰明了。我认为 Go 语言一开始就走了一条非常正确的道路，它的并发编程模式从一开始到现在就没怎么变过。

下面来看看 Python asyncio 和 Go 的类似之处。

启动一个 goroutine 运行：

```go
go func()
```

而 Python 启动 coroutine 运行：

```py
asyncio.create_task(func())
```

Go 中 goroutine 之间推荐的通信机制为为 `chan`：

```go
ch := make(chan int)     // 无缓冲 channel
ch := make(chan int, 10) // 缓冲大小为 10 的 channel

ch <- 42      // 发送数据到 channel
value := <-ch // 从 channel 接收数据
close(ch)     // 关闭 channel
```

Python 中 coroutine 之间也可用异步队列 `asyncio.Queue` 通信：

```py
q = asyncio.Queue(0)     # 无限长度的队列，和 Go 不同
q = asyncio.Queue(10)    # 长度为 10 的队列

await q.put(42)          # 发送数据到队列
value = await q.get()    # 从队列获取数据
q.shutdown()             # 关闭队列
```

需要注意，0 在 Python `asyncio.Queue` 中代表无限长度的队列，和 Go 有很大的区别。

Go 中等待一组 goroutine 都结束：

```go
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
    }
}

// 阻塞
wg.Wait()
```

Python 中等待一组 coroutine 都结束：

```py
async with asyncio.TaskGroup() as tg:
    for i in range(10):
        tg.create_task(func())
```

`asyncio.TaskGroup`，是 `Python 3.11` 才加入的，因为 Python 有上下文管理器，语法更简练。

Go 中的 `context.Context` 兼具多种功能，是一个跨 goroutine 存储信息，传递控制的对象：

```go
// 存储信息
func parent() {
    ctx := context.Value(context.Background(), "key", "value")

    go child(ctx)
}

func child(ctx context.Context) {
   val := ctx.Value("key")
}

// 控制
func parent1() {
    ctx1, cancel := context.WithCancel(context.Background())
    go child1(ctx1)

    // 让 child1 直接结束 
    cancel() 

    // 若 child1 超时，则不再等待，结束
    ctx2 := context.WithTimeout(context.Background(), 5 * time.Second)
    child1(ctx2)
}

func child1(ctx context.Context) {
    select {
    case <-ctx.Done():
    }
}
```

Python 中没有直接对应的东西，和 `context.Value` 类似的是 `contextvars.ContextVar`，这玩意又和多线程编程中的**线程本地变量**极其相似：

```py
request_id = contextvars.ContextVar('request_id')

async def handle_request(id):
    # 为每个请求设置不同的 request_id
    request_id.set(id)
    
    # 即使在嵌套的协程中也能正确获取
    await asyncio.sleep(0.1)
    print(f"Request {request_id.get()} processed")

# 同时运行多个请求处理
await asyncio.gather(
    handle_request(1),
    handle_request(2),
    handle_request(3)
)
```

而控制 coroutine 又有多种方式，比如取消一个正在运行的 coroutine：

```py
async def my_coroutine():
    try:
        while True:
            await asyncio.sleep(1)
    except asyncio.CancelledError:
        print("Coroutine was cancelled!")
        raise

# 创建任务
task = asyncio.create_task(my_coroutine())

# 取消任务
task.cancel()
```

又比如超时退出：

```py
# 最多等 2 秒
await asyncio.wait_for(my_coroutine(), timeout=2)
```