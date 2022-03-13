title: Go 内存模型
date: 2022-03-13 16:44:02
summary: 
---

并发场景下最讨厌的是对变量的读写操作隐含着违反直觉的结果。这时候，就需要语言的**内存模型**了。所有语言的内存模型都在讲着一件事情，什么操作才能让写变量的结果被另一个线程/协程读到。

幸好 Go 一直是一个设计的很“简单”的语言，它的内存模型需要记住的结论很少，我们很容易就写出基于 goroutine 的并发安全程序。

先解释一个经典的名词 `happens-before`：如果一个事件的发生在另一个事件之前，那么第一个事件的结果必须影响到第二个事件。`happens-before` 听上去理应如此，但由于现代处理器的多核、指令重排，编译器优化等等因素，`happens-before` 不是天然成立的。

接下来看一下 Go 内存模型规定了哪些操作是确定的：

1. 如果包 `p` 导入了包 `q`，那么包 `q` 的 `init` 函数 `happens-before` 包 `p` 的任何代码。
2. 所有包的 `init` 函数 `happens-before` `main.main` 函数。
3. 用 `go` 语句启动新的 `goroutine`，该语句本身 `happens-before` 启动的这个新 `goroutine`。
4. `channel` 上的发送 `happens-before` 该 `channel` 上的接收。
5. 关闭 `channel` 的操作 `happens-before` 在 `channel` 上读到零值（零值意味着 `channel` 关闭）。
6. 无缓冲 `channel` 的接收 `happens-before` 该 `channel` 上的发送完成。
7. 容量为 C 的 `channel` 接收到第 k 个元素 `happens-before` 该 `channel` 第 k+C 个元素发送完成。
8. 两个变量 n 和 m，n 出现在 m 的前面，若使用互斥锁 `sync.Mutex` 或 `sync.RWMutex` 保护，那么变量 n 的 `l.Unlock` `happens-before` 变量 m 的 `l.Lock()`。
9. 如果有多个对 `once.Do(f)` 的调用，那么第一个调用 `f()` 函数 `happens-before` 任何其他 `once.Do(f)` 的调用返回。
10. `sync.Cond` 的 `Broadcast` 和 `Signal` 调用 `happens-before` 阻塞中 `Wait` 调用。
11. 对于 `Sync.Map`，`Load`, `LoadAndDelete` 和 `LoadOrStore` 都是读操作。`Delete`, `LoadAndDelete` 和 `Store` 都是写操作。`LoadOrStore` 如果 `loaded` 返回 `false`，那也是写操作。`Sync.Map` 的所有写操作 `happens-before` 读操作。
12. 对于 `Sync.Pool`，`Put` `happens-before` `Get`，`New` `happens-before` `Get`。
13. 对于 `Sync.WaitGroup`，`Done` `happens-before` 任何一个阻塞中的 `Wait`。
14. `sync.atomic` 包是一些列原子操作的合集，可用于不同 goroutine 间的同步。如果原子操作 B 的观察到原子操作 A 的结果，那么 A `happens-before` B。一旦使用了原子操作，那么程序便拥有了**一致**的执行顺序。