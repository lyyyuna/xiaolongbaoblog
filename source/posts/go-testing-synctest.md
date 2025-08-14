title: 使用 testing/synctest 测试并发代码
date: 2025-08-15 15:36:25
---


Go 语言的核心特性之一是其原生支持的并发编程能力。Goroutine 和 channel 作为语言原语，为编写并发程序提供了简洁高效的构建模块。

然而，并发程序的测试往往颇具挑战且容易出错。

在 Go 1.24 版本中，我们引入了全新的实验性测试包 `testing/synctest` 来支持并发代码测试。本文将阐述该实验功能的研发背景，演示synctest包的具体使用方法，并探讨其未来发展前景。

需要注意的是，Go 1.24 中的 `testing/synctest` 目前处于实验阶段，不受 Go 语言兼容性承诺的保障。该功能默认不开启，如需使用，必须在编译时通过设置环境变量 `GOEXPERIMENT=synctest` 来激活。

## 测试并发代码是困难的

首先，让我们从一个简单示例开始分析。

`context.AfterFunc` 函数的作用是在上下文取消后，安排特定函数在独立的 goroutine 中执行。以下是针对 AfterFunc 的测试：

```go
func TestAfterFunc(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())

    calledCh := make(chan struct{}) // AfterFunc调用时关闭
    context.AfterFunc(ctx, func() {
        close(calledCh)
    })

    // TODO: 验证AfterFunc尚未被调用
    cancel()
    // TODO: 验证AfterFunc已被调用
}
```

这个测试需要验证两个关键条件：

1. 上下文取消前函数未被调用
2. 上下文取消后函数被调用

在并发系统中验证否定条件（未被调用）尤其困难。虽然我们可以轻松检测函数当前是否被调用，但如何确保它将来也不会被调用？

常见的解决方案是引入等待超时机制。我们在测试中添加以下辅助函数：

```go
// funcCalled检测函数是否被调用
funcCalled := func() bool {
    select {
    case <-calledCh:
        return true
    case <-time.After(10 * time.Millisecond):
        return false
    }
}

if funcCalled() {
    t.Fatalf("上下文取消前AfterFunc已被调用")
}
cancel()
if !funcCalled() {
    t.Fatalf("上下文取消后AfterFunc未被调用")
}
```

这种实现存在两个明显问题：

1. 执行效率低：虽然单次10毫秒延迟看似短暂，但在大量测试用例累积时会显著拖慢整体测试速度
2. 结果不可靠：
    1. 在性能强劲的机器上，10毫秒已经是较长的等待时间
    2. 在共享且高负载的CI环境中，经常会出现数秒级的延迟波动

我们面临两难选择：

1. 若要提高稳定性，就需要延长等待时间，导致测试更慢
2. 若要加快测试速度，就必须缩短等待时间，但会降低可靠性

无论如何折衷，都无法同时实现快速和可靠的测试效果。

## 引入 testing/synctest 包

`testing/synctest` 包解决了这个问题。它让我们可以重写这个测试，使其简单、快速且可靠，而无需修改被测试的代码。

该包仅包含两个函数：`Run` 和 `Wait`。

`Run` 会在一个新的 goroutine 中调用函数。这个 goroutine 和它启动的任何 goroutine 都存在于一个我们称为 **bubble** 的隔离环境中。`Wait` 函数则会等待当前 goroutine 所在 **bubble** 中的所有 goroutine 都进入阻塞状态（等待同一个 **bubble** 内的其他 goroutine）。

让我们使用 `testing/synctest` 包重写上面的测试：

```go
func TestAfterFunc(t *testing.T) {
    synctest.Run(func() {
        ctx, cancel := context.WithCancel(context.Background())

        funcCalled := false
        context.AfterFunc(ctx, func() {
            funcCalled = true
        })

        synctest.Wait()
        if funcCalled {
            t.Fatalf("AfterFunc function called before context is canceled")
        }

        cancel()

        synctest.Wait()
        if !funcCalled {
            t.Fatalf("AfterFunc function not called after context is canceled")
        }
    })
}
```

这与我们最初的测试几乎完全相同，但我们将测试封装在 `synctest.Run` 调用中，并在断言函数是否被调用之前调用了 `synctest.Wait`。

`Wait` 函数会等待调用者 **bubble** 内的所有 goroutine 进入阻塞状态。当它返回时，我们就可以确定 `context` 包要么已经调用了该函数，要么在我们采取进一步行动之前不会调用它。

现在这个测试既快速又可靠。

测试也变得更简单：我们用布尔值替换了原来的 `calledCh` 通道。之前我们需要使用通道来避免测试 goroutine 和 `AfterFunc` goroutine 之间的数据竞争，但现在 `Wait` 函数提供了这种同步机制。

`race` 检测器也能够理解 `Wait` 调用，这个测试在使用 `-race` 参数运行时能够通过。如果我们移除第二个 `Wait` 调用，`race` 检测器会正确地报告测试中存在数据竞争。

## 和时间相关的测试

并发代码常常需要处理时间问题。

测试涉及时间的代码可能会很困难。正如前文所见，在测试中使用真实时间会导致测试缓慢且不稳定。有人可能会使用模拟时间，这样则需要避免直接调用 `time` 包中的函数，并确保被测代码能兼容可选的模拟时钟。

`testing/synctest` 包让测试涉及时间的代码变得更加简单。

在 `Run` 启动的作用域 **bubble** 内，goroutine 会使用一个模拟时钟。在该作用域内，`time` 包的所有函数都会操作这个模拟时钟。只有当所有 goroutine 都进入阻塞状态时，模拟时间才会向前推进。

为了演示这一点，我们来为 `context.WithTimeout` 函数编写一个测试。`WithTimeout` 会创建一个子 `context`，并在给定的超时时间后自动失效。

```go
func TestWithTimeout(t *testing.T) {
    synctest.Run(func() {
        const timeout = 5 * time.Second
        ctx, cancel := context.WithTimeout(context.Background(), timeout)
        defer cancel()

        // 先等待至接近超时时刻。
        time.Sleep(timeout - time.Nanosecond)
        synctest.Wait()
        if err := ctx.Err(); err != nil {
            t.Fatalf("before timeout, ctx.Err() = %v; want nil", err)
        }

        // 再等待剩余时间直至超时生效。
        time.Sleep(time.Nanosecond)
        synctest.Wait()
        if err := ctx.Err(); err != context.DeadlineExceeded {
            t.Fatalf("after timeout, ctx.Err() = %v; want DeadlineExceeded", err)
        }
    })
}
```

我们编写这个测试的方式，就像在使用真实时间一样。唯一的区别在于：

1. 我们将测试函数包裹在 `synctest.Run` 中。
2. 在每次调用 `time.Sleep` 后，都会调用 `synctest.Wait` 来等待 `context` 包的计时器完成运行。

## 阻塞和 bubble

`testing/synctest` 的核心概念是 **bubble** 进入持久阻塞状态。当 **bubble** 内的所有 goroutine 均被阻塞，且只能由同一 **bubble** 内的其他 goroutine 解除阻塞时，即触发此状态。

持久阻塞的判定规则:

1. 存在等待调用：若有未完成的 `Wait` 调用，则立即返回
2. 时间推进机制：若无等待调用，模拟时钟将跳转到下一个可能唤醒 goroutine 的时间点
3. 死锁检测：若上述条件均不满足，则判定为死锁并触发 `Run` 函数 panic

注意：若 goroutine 的阻塞可能被 **bubble** 外部事件解除，则不符合持久阻塞定义

可触发持久阻塞的操作:

1. 对 nil 通道的发送/接收操作
2. 对同 **bubble** 内创建的通道的阻塞式发送/接收
3. 所有 case 分支均满足持久阻塞条件的 select 语句
4. `time.Sleep`
5. `sync.Cond.Wait`
6. `sync.WaitGroup.Wait`

## Mutexes

对 `sync.Mutex` 的操作不属于持久阻塞行为。

在实际开发中，函数获取全局互斥锁的情况十分常见。例如，`reflect` 包中的许多函数都会使用受互斥锁保护的全局缓存。如果在 synctest 测试 **bubble** 内的某个 goroutine 因尝试获取被外部 goroutine 持有的互斥锁而阻塞，这种情况并不构成持久阻塞——虽然当前处于阻塞状态，但其解除阻塞依赖于测试 **bubble** 之外的 goroutine。

考虑到互斥锁通常不会长时间持有，直接将其排除在 `testing/synctest` 的考量范围之外。

## Channels

在 **bubble** 内创建的通道与外部通道具有不同的行为特性：

1. 通道阻塞规则。只有当通道属于当前 **bubble** 时，通道操作才可能触发持久阻塞。对非 **bubble** 通道的操作不会被视为持久阻塞条件。

2. 跨 **bubble** 访问限制。从外部 **bubble** 操作内部通道将直接引发 panic（运行时恐慌）。这一机制确保了 goroutine 只会在与同 **bubble** 内的其他 goroutine 通信时才可能进入持久阻塞状态。


## I/O

外部 I/O 操作（例如从网络连接读取数据）不属于持久阻塞行为。

网络读取操作可能会被来自测试 **bubble** 外部的写入操作解除阻塞，甚至可能来自其他进程。即使某个网络连接的唯一写入方也位于同一个测试 **bubble** 内，运行时系统仍然无法区分以下两种情况：

1. 连接正在等待更多数据到达
2. 内核已经接收到数据但正在传递过程中

使用 synctest 测试网络服务器或客户端时，通常需要提供模拟网络实现。例如，[net.Pipe](https://go.dev/pkg/net#Pipe) 函数可以创建一对使用内存网络连接的 `net.Conn`，这些连接可用于 synctest 测试。

## Bubble 生命周期

`Run` 函数会在一个新的 **bubble** 中启动 goroutine。该函数会在以下两种情况下返回：

1. 当 **bubble** 内的所有 goroutine 都退出时
2. 如果 **bubble** 进入持久阻塞状态且无法通过时间推进来解除阻塞，则会触发 panic

由于 Run 函数要求 **bubble** 的所有 goroutine 都必须退出后才能返回，这意味着测试代码必须特别注意：

1. 在测试完成前清理所有后台 goroutine
2. 确保没有 goroutine 被意外泄漏

## 测试含有网络的操作的代码

让我们看另一个示例，这次使用 `testing/synctest` 包来测试网络程序。在这个示例中，我们将测试 `net/http` 包对 `100 Continue` 响应的处理。

发送请求的 HTTP 客户端可以包含 "Expect: 100-continue" 头，告诉服务器客户端有额外的数据要发送。然后，服务器可能会响应 `100 Continue` 信息性响应以请求剩余的请求内容，或者用其他状态码告诉客户端不需要内容。例如，上传大文件的客户端可以使用此功能在发送文件前确认服务器是否愿意接收该文件。

我们的测试将验证：当发送 "Expect: 100-continue" 头时，HTTP 客户端不会在服务器请求前发送请求内容，并且在收到 `100 Continue` 响应后会发送内容。

通常，测试通信的客户端和服务器可以使用环回网络连接。然而，在使用 `testing/synctest` 时，我们通常希望使用模拟网络连接，以便检测所有 goroutine 何时在网络操作上阻塞。我们将通过创建一个使用 [net.Pipe](https://go.dev/pkg/net#Pipe) 创建的内存网络连接的 `http.Transport`(HTTP 客户端)来开始这个测试。

```go
func Test(t *testing.T) {
    synctest.Run(func() {
        srvConn, cliConn := net.Pipe()
        defer srvConn.Close()
        defer cliConn.Close()
        tr := &http.Transport{
            DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
                return cliConn, nil
            },
            // 设置非零超时时间会启用"Expect: 100-continue"处理机制
            // 由于后续测试中没有使用 sleep 操作
            // 即使在运行缓慢的机器上测试耗时较长
            // 也不会触发这个超时条件
            ExpectContinueTimeout: 5 * time.Second,
        }
```

我们通过这个 `transport` 发送一个设置了"Expect: 100-continue"请求头的请求。该请求在一个新的 goroutine 中发送，因为它要到测试结束时才会完成。

```go
        body := "request body"
        go func() {
            req, _ := http.NewRequest("PUT", "http://test.tld/", strings.NewReader(body))
            req.Header.Set("Expect", "100-continue")
            resp, err := tr.RoundTrip(req)
            if err != nil {
                t.Errorf("RoundTrip: unexpected error %v", err)
            } else {
                resp.Body.Close()
            }
        }()
```

我们读取客户端发送的请求头信息。


```go
        req, err := http.ReadRequest(bufio.NewReader(srvConn))
        if err != nil {
            t.Fatalf("ReadRequest: %v", err)
        }
```

现在进入测试的核心部分：我们需要确认客户端此时尚未发送请求体内容。

我们启动一个新的 goroutine，将发送给服务器的请求体内容复制到 `strings.Builder` 中，等待 **bubble** 中的所有 goroutine 进入阻塞状态，然后验证此时尚未从请求体中读取到任何数据。

如果忘记调用 `synctest.Wait`，竞态检测器会正确报告存在数据竞争，但加入了 `Wait` 调用后就能保证测试的安全性。

```go
        var gotBody strings.Builder
        go io.Copy(&gotBody, req.Body)
        synctest.Wait()
        if got := gotBody.String(); got != "" {
            t.Fatalf("before sending 100 Continue, unexpectedly read body: %q", got)
        }
```

我们向客户端写入"100 Continue"响应，并验证此时客户端已开始发送请求体数据。

```go
        srvConn.Write([]byte("HTTP/1.1 100 Continue\r\n\r\n"))
        synctest.Wait()
        if got := gotBody.String(); got != body {
            t.Fatalf("after sending 100 Continue, read body %q, want %q", got, body)
        }
```

最后，我们发送"200 OK"响应来完成此次请求处理。

在本测试过程中，我们启动了多个 goroutine。`synctest.Run` 调用将等待所有 goroutine 退出后才会返回。

```go
        srvConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
    })
}
```

该测试可轻松扩展以验证其他行为，例如：

1. 当服务器未要求时，确保请求体不会被发送。
2. 当服务器未在超时时间内响应时，确保请求体会被正常发送。