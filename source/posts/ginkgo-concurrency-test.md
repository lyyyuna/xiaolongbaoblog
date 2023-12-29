title: Ginkgo 并发测试教程
date: 2023-12-29 21:04:00
series: Ginkgo 使用笔记
summary: 如何正确编写并发测试用例？

---

## 前言

`Ginkgo CLI` 加上 `--nodes xxx` 或 `--procs yyy` 参数后，就能让原本顺序执行的测试用例变成并发执行。

但好奇的你，可能会有如下问题：

1. 集测并发的顺序是固定的还是随机的？
2. 集测之间的共享变量会有并发安全问题吗？
3. 有一些需要独占资源的测试用例，如何在并发中控制它们的顺序？
4. xxx

下面，我们就来一一解答。

P.S. 本文基于 `Ginkgo v2.13.2`。

## 并发模型

`Ginkgo` 是 `Go` 语言编写的集测框架，还兼容 `go test`，不免想当然认为，`Ginkgo` 是基于 goroutine 实现的多并发。但是恰恰相反，经过[Ginkgo 测试框架实现解析](/2022/05/12/inside-the-ginkgo/)的分析，`Ginkgo` 其实是多进程模型，每个进程会使用**相同的随机数种子**打乱用例，得到次序一致的随机序列。

这样，`Ginkgo` 就实现了：

1. 每次运行用例顺序随机。
2. 借助进程得到并发隔离。

那么，单个进程内共享变量是安全的吗？看下面这个例子：

```go
Context("test", Id("27838"), func() {
    var global int

    It("a", func() {
        global++
        fmt.Println("a", global)
    })

    It("b", func() {
        global++
        fmt.Println("b", global)
    })

    It("c", func() {
        global++
        fmt.Println("c", global)
    })
})
```

假设并发 1，实际结果为

```
a 1
b 2
c 3
```

假设并发 2，实际结果为

```
a 1
b 1
c 2

或

a 1
b 2
c 1
```

假设并发 3，实际结果为

```
a 1
b 1
c 1
```

两个用例如果被分配在同一个进程中，去访问同一个变量，就会互相干扰，应使用 `BeforeEach` 对变量显式初始化：

```go
Context("test", Id("27838"), func() {
    var global int

    BeforeEach(func() {
        global = 0
    })

    It("a", func() {
        global++
    ....
    })
})
```

`BeforeEach` 会互相干扰吗？**不会**，同一个进程中同一个时刻只会有一个测试用例在运行。

## 并发集测的初始化

多进程不像 goroutine 那样容易实现并发同步，这给那些只需要做一遍的集测初始化步骤带来了挑战（比如，在集测开始之前，往数据库中导入测试数据）。

为此，`Ginkgo` 提供了以下的解决方案：

```go
// 集测并发开始前执行
func SynchronizedBeforeSuite(
    process1 func() []byte,
    allProcesses func([]byte),
)

// 集测并发结束后执行
func SynchronizedAfterSuite(
    allProcesses func(),
    process1 func(),
)
```

`SynchronizedBeforeSuite` 有两个参数：

1. 其中 `process1` 只在第 1 个进程内运行，`allProcesses` 会在所有进程内运行。
2. `process1` 运行结束后才会运行 `allProcesses`。
3. `process1` 的返回值会作为参数传递给 `allProcesses`。

举个例子，有一个云存储上传测试场景，并发为 5，需要在测试前创建存储空间，空间名用 `bucketid` 来标识，在 `Ginkgo` 中可以这么做：

```go
// 全局变量存储空间名
var bucketId string

func SynchronizedBeforeSuite(
    // process1
    func() []byte {
        // 创建存储空间，得到空间名
        id := CreateBucket()
        // 向后传递空间名
        return []byte(id)
    },
    // allProcesses
    func(rId []byte) {
        // 得到空间名
        bucketId = string(rId)
    },
)
```

1. `process1` 负责创建该存储空间，将空间名以字节形式传递给其他并发进程。
2. 其他并发进程将收到的字节转化为字符串，再赋值给全局变量 `bucketId`。
3. 集测并发开始后，全局变量 `bucketId` 就拿到了初始化后的空间名。

一定要**注意**，这里 `process1` 和 `allProcesses` 虽然写在一起，但其实是在**不同阶段**和**不同进程**内执行的，不能认为它们能够共享全局变量。同时，初始化进程和并发进程之间只能传递字节数据，不能直接传递变量对象（受限于进程间通信），如果要传递复杂的对象，可以使用 `json` 做序列化反序列化。下面的例子中，`info` 变量在每个并发进程内都拿到了初始化后的空间名和用户名：

```go
type testinfo struct {
    bucketId string `json:"bucket_id"`
    userId string `json:"user_id"`
}

// 全局变量
var info testinfo

func SynchronizedBeforeSuite(
    // process1
    func() []byte {
        info := SetupTestInfo()
        // 序列化
        data, _  := json.Marshal(&info)
        return data
    },
    // allProcesses
    func(data []byte) {
        // 反序列化
        json.Unmarshal(data, &info)
    },
)
```

`SynchronizedAfterSuite` 的工作原理类似，不再赘述。

从我个人角度来看，`Ginkgo` 这种机制并不直观，幸好集测初始化一般是由团队中的测试架构者来设计，能够避免初学者误用。

## 控制并发中的顺序

### 排它执行

有一些测试，运行的时候不能有其他测试的干扰，例如：

1. 某个接口的性能测试。
2. 测试独占资源。

`Ginkgo` 提供了 `Serial` 装饰器：

```go
Describe("空间", Serial, func() {
    It("创建空间性能测试", func() {
        ...
    })

    It("测试空间 id 是否单调递增", func() {
        ...
    })

    It("修改空间为私有", func() {
        ...
    })
})
```

`Ginkgo` 会等待所有并发用例结束后，在第一个进程内运行 `Serial` 用例。

### 顺序执行

假设如下场景需要按顺序执行：

1. 创建一个存储空间。
2. 先测试上传文件。
3. 然后测试下载文件，下载的文件可以使用步骤 2 中成功上传的文件。

用例 3 必须等待用例 2 执行完毕。
（你可能会觉得上述的三个用例可以合并为单个用例中的三个步骤，这里只是为了演示，面对复杂的多场景测试，最好还是拆分成多个用例）

`Ginkgo` 提供了 `Ordered` 装饰器：

```go
Describe("空间", Ordered, func() {
    It("创建空间", func() {
        ...
    })

    It("上传文件", func() {
        ...
    })

    It("下载文件", func() {
        ...
    })
})
```

`Ginkgo` 会保证同一个 `Ordered` 容器内的测试用例按顺序执行。和 `Serial` 不一样，`Ordered` 容器内外的用例会并发执行，并且可以放在任意一个并发进程内，不会局限在第一个进程。

#### 一次性初始化

原本的初始化 `BeforeEach` 节点在 `Ordered` 容器内仍然有效，含义不变：

```go
Describe("空间", func() {
    BeforeEach(func() {
        fmt.Println("a")
    })

    Context("", Ordered, func() {
        BeforeEach(func() {
            fmt.Println("b")
        })

        It("创建空间", func() {
            // 执行 BeforeEach a
            // 执行 BeforeEach b
        })

        It("上传文件", func() {
            // 执行 BeforeEach a
            // 执行 BeforeEach b
        })

        It("下载文件", func() {
            // 执行 BeforeEach a
            // 执行 BeforeEach b
        })
    })

    It("列举文件", func () {
        // 只执行 BeforeEach a
    })
})
```

如果 `Ordered` 内的用例，只想执行一次初始化该怎么做呢？对于 `Ordered` 容器**外**初始化，`Ginkgo` 提供了 `OncePerOrdered` 装饰器：

```go
Describe("空间", func() {
    BeforeEach(OncePerOrdered, func() {
        fmt.Println("a")
    })

    Context("", Ordered, func() {
        // 只执行一次 BeforeEach a
        BeforeEach(func() {
            fmt.Println("b")
        })

        It("创建空间", func() {
            // 执行 BeforeEach b
        })

        It("上传文件", func() {
            // 执行 BeforeEach b
        })

        It("下载文件", func() {
            // 执行 BeforeEach b
        })
    })

    It("列举文件", func () {
        // 只执行 BeforeEach a
    })
})
```

对于 `Ordered` 容器**内**初始化，`Ginkgo` 提供了 `BeforeAll` 和 `AfterAll` 节点来替换 `BeforeEach` 和 `AfterEach` 节点：

```go
Describe("空间", func() {
    // a
    BeforeEach(OncePerOrdered, func() {
        fmt.Println("a")
    })

    Context("", Ordered, func() {
        // 只执行一次 BeforeEach a
        // 只执行一次 BeforeEach b
        BeforeAll(func() {
            fmt.Println("b")
        })

        It("创建空间", func() {
        })

        It("上传文件", func() {
        })

        It("下载文件", func() {
        })
    })

    It("列举文件", func () {
        // 只执行 BeforeEach a
    })
})
```

这样，`Ordered` 容器内的三个创建空间、上传文件、下载文件用例只会初始化一次 a 和 b，而 `Ordered` 容器外的列举文件用例，仍然会执行初始化 a。

#### 错误机制

默认情况下 `Ordered` 容器内的用例只要出错，剩下的用例便会跳过，直接执行 `AfterAll` 节点做清理工作，在测试报告中，这些未执行的用例也会显示成 `Skip`。

如果 `Ordered` 容器内的用例并没有依赖关系，只是单纯组合在一起，那这种默认行为就不合适了，你可以使用 `ContinueOnFailure` 装饰器修改成继续执行剩余用例。

#### 重试机制

`Ginkgo run --flake-attempts [int]` 命令可将集测设置成重试模式。对于 `Ordered` 容器内的用例，重试不包括 `BeforeAll` 和 `AfterAll` 内的步骤。

