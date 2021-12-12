title: 如何收集 Go 的实时覆盖率
date: 2021-12-12 12:12:12
summary: 开源一款 Go 实时覆盖率收集工具
---

## 前言

接触过 Go 的同学知道，官方没有提供集成测试覆盖率的收集方案。针对集成测试覆盖率的需求，七牛开源了一款工具 goc ([https://github.com/qiniu/goc](https://github.com/qiniu/goc))，能很好地解决这个问题。

本文分享了 goc 的使用方法和实现的原理。

## Go 单测覆盖率

在深入集测/实时覆盖率之前，先来看看 Go 官方的单测覆盖率是如何收集的。

举个例子，有两个文件，被测业务代码 `kcp.go` 和单测代码 `kcp-test.go`。
```go
// kcp.go
package kcp

func Size(a int) string {
    switch {
    case a < 0:
        return "negative"
    case a == 0:
        return "zero"
    case a < 10:
        return "small"
    case a < 100:
        return "big"
    case a < 1000:
        return "huge"
    }
    return "enormous"
}

// kcp_test.go
package kcp

import "testing"

type Test struct {
    in  int
    out string
}

var tests = []Test{
    {-1, "negative"},
    {5, "small"},
}

func TestSize(t *testing.T) {
    for i, test := range tests {
        size := Size(test.in)
        if size != test.out {
            t.Errorf("#%d: Size(%d)=%s; want %s", i, test.in, size, test.out)
        }
    }
}
```

我们用 `go test` 工具运行单测，打开 `-cover` 选项就可以收集到单测覆盖率，再开启 `-coverprofile` 选项，就可以将覆盖率报告输出到本地：

```
mode: set
kcp/kcp/kcp.go:3.25,4.12 1 1
kcp/kcp/kcp.go:16.5,16.22 1 0
kcp/kcp/kcp.go:5.16,6.26 1 1
kcp/kcp/kcp.go:7.17,8.22 1 0
kcp/kcp/kcp.go:9.17,10.23 1 1
kcp/kcp/kcp.go:11.18,12.21 1 0
kcp/kcp/kcp.go:13.19,14.22 1 0
```

如何解读这个覆盖率报告？Go 不是对每一行去统计，而是把整个代码拆分成了一个个语句块。覆盖率报告的每一行开头表示该语句块所在的文件名，接着是语句块的起始行和起始列，然后是语句块的结束行、结束列，倒数第二列为语句块内的总行数，最后一列为该语句块在这次测试中被执行到的次数。结合所有的语句数量和执行次数便可计算出整个代码的覆盖率了。

但测试不仅仅只有单测，测试同学平时接触最多的反而是手工、集测、系统和自动化测试，对于这些测试手段与被测程序分离、处于不同二进制的情况，需要新的覆盖率收集方法。

## 传统的集测覆盖率收集方法 vs goc

目前业界的覆盖率收集方案如下：

首先在原工程中新增单测代码，添加测试用例 `TestMainStart`，在该用例最后调用被测业务 `main`。

```go
func TestMainStart(t *teing.T){
    van args[]string
    for _, arg := range os.Args{
        if !strings.HasPrefix(arg, "-test"){
            args = append(args,arg)
        }
    }
    os.Angs = args
    main()
}
```

然后使用 `go test -c` 选项将测试用例编译成可执行文件 `cover.test`。

```
$ go test -coverpkg="./..." -c -o cover.test
```

最后用 `-cover.run` 指定运行我们封装过的测试用例 `TestMainStart`。

```
$ ./cover.test -test.run "TestMainStart" -test.coverprofile=cover.out
```

`cover.test` 运行起来后，我们就可以进行各种手工、集测、系统和自动化测试了。由于 `cover.test` 这个二进制本质上是一个 `go test` 的单测代码，所以当程序退出后会自动输出覆盖率报告，我们也就能得到被测业务的集测覆盖率了。

![传统 Go 集测覆盖率收集方法示意图](/img/posts/goc/legacy-collect-way.png)

但该方案有很多缺点：
1. 工程中需要新增测试代码，如果工程中有多个 `main` 函数，就得为每一个 `main` 函数编写测试用例。
2. 改变了程序启动的方式，必须加上 `-test.run` 和 `-test.coverprofile` 两个参数。
3. 被测程序退出后才能得到覆盖率报告，无法获取实时覆盖率。
4. 覆盖率报告只能输出到本地，需要编写脚本集中收集。
5. 分布式微服务架构下，需要对每个服务、多份报告合并才能得到整个系统的覆盖率。

让我们看看 `goc` 如何解决这些缺点：

如果原工程是用 `go build` 命令编译的，那么现在只需替换为 `goc build` 命令编译，即可得到一个支持覆盖率收集的可执行文件。原工程代码无须做任何改动。

![使用 goc 编译示意图](/img/posts/goc/iamge16.gif)

而使用 `goc` 编译出的二进制，不会破坏原程序的启动方式。

`goc` 方案中被测程序不再需要退出就能收集覆盖率。如下面动图所示。左侧是一个被测的 HTTP 服务。右侧是用 `goc` 命令获取的实时覆盖率。上文介绍了报告最后一列代表了语句块执行的次数，当我用 `curl` 命令访问了被测 HTTP 服务后，可以看到部分语句块执行次数从 0 变成了 1，整个测试过程被测业务正常运行。

![goc 获取覆盖率示意图](/img/posts/goc/iamge20.gif)

最后覆盖率报告是统一收集、自动合并。如下动图所示，左边是两个被测的 HTTP 服务，一个监听在 5000 端口，另一个监听在 5001 端口。首先访问其中 5000 端口的服务，可以看到部分语句块执行次数从 0 变到了 1。然后访问 5001 端口的服务，语句块执行次数从 1 变成了 2。`goc` 所展示的覆盖率报告不再是单个服务的，而是把被测系统看作了一个整体。

![goc 获取分布式覆盖率示意图](/img/posts/goc/iamge23.gif)

另外动图中还展示了 `goc` 实时清除/重置覆盖率的能力，测试人员可以结合该能力对某个功能反复测试。

## goc 覆盖率收集原理

接下来我们简单介绍一下 Goc 的原理。

回顾语句覆盖率的定义：

    > 测试时运行被测程序后，程序中被执行到可知性语句的比率。

从这个定义出发我们有一个朴素的想法，能不能在每一行运行完后加入一些统计计算的逻辑去统计覆盖率呢？

一些语言正是这样做的，比如 Python。Python 是带解释器的动态语言，标准库提供了方法 `sys.settrace` 给解释器设置一个回调函数，当解释器每执行一行代码就会调用一次回调函数，传入当前行的行号，最后统计所有执行到的行数，便可得到覆盖率。C++ 没有解释器，所以只能在运行前往代码中插入技术器。但 C++ 认为每一行统计非常低效，对性能损失很大。所以，它把代码分成了很多的基本块，在基本块之间的跳转处插入一个计数器，等待程序结束之后统计这些计数器的结果，便可得到全局的覆盖率。Java 的做法和 C++ 非常类似，但没有在汇编层做插桩，而是在字节码处插桩。Java 虚拟机提供了动态改变字节码的能力，Jacoco 在跳转处插入探针 Probe 做覆盖率收集。

我们回到 Go，看一下 `go test` 单元测试怎么收集覆盖率的。Go 也是把代码分成了很多语句块。比如说这段代码就分成了三块：


![Go 单测语句块](/img/posts/goc/go-unit-test1.png)

但和 C++、Java 在块之间跳转处插桩不同， Go 是在语句块中插入了计数器：

![Go 单测语句块计数器](/img/posts/goc/go-unit-test2.png)

上述计数器的定义如下： 

```go
var GoCover = struct {
    Count     [3]uint32
    Pos       [3 * 3]uint32
    NumStmt   [3]uint16
} {
    Pos: [3 * 3]uint32{
       5, 7, 0xc000d, // [0]
       7, 9, 0x3000c, // [1]
       9, 11, 0x30008, // [2]
    }
    NumStmt: [3]uint16{
        2, // 0
        1, // 1
        1, // 2
    },
}
```

Count 这个字段代表计数器，Pos 就是代表了语句块起始行、起始列、结束行和结束列，NumStmt 代表了语句块内的语句数。

这个方案是源码级的插桩，所以势必会破坏源码。所以 `goc` 采取的策略是首先把整个工程拷贝到临时目录：

![goc 临时目录](/img/posts/goc/goc-temp.png)

接着使用标准库 `ast/parser` 解析出语法树，在每个语句块中插入计数器写入临时目录，最后调用原生的 `go build/install/run` 命令去编译插过桩的代码。

除了计数器，我们还为每个 main 包注入了一个 HTTP API。调用暴露的 HTTP 接口会计算聚合每个计数器，并返回服务当前的覆盖率。插桩服务启动后，首先会访问 `goc server` 这个注册中心，上报自己 ip 地址和 HTTP API 端口。`goc server` 注册中心存储了每个被测服务的信息。

![goc server 注册中心](/img/posts/goc/goc-server.png)

客户端向 `goc server` 注册中心获取覆盖率报告时，中心会拉取所有服务的覆盖率并合并成单份报告，然后返回给客户端。

在实际公测中，有用户反馈在云原生场景下集成 `goc` 非常不便，比如 `docker/k8s` 启动服务必须加 `-p` 参数，将外部的端口映射到内部的端口，才能让注册中心 `goc server` 访问到被插桩的服务。这个问题的本质是因为 `goc server` 注册中心和被插桩服务分属不同网络，它们通过 NAT 实现网络转换。

`goc` 在 v2 版本给出了改进方案，v2 将所有内部通信构建在了 websocket 长连接之上，整个覆盖率收集系统只需保证 `goc server` 能被访问即可，对测试系统的部署侵入降到了最低，更适合云原生业务的测试。

![goc server ws 注册中心](/img/posts/goc/goc-server-ws.png)

## 如何利用 goc 

由于 `goc` 输出的覆盖率报告和 Go 官方的单测覆盖率报告完全一致。一些已有的分析单测覆盖率的系统可以直接用来分析 `goc` 收集的集测覆盖率，比如说 SonarQube、Codecov、Covralls。

借助 `goc` 覆盖率实时收集的能力还可以做精准测试。这里有两个方案：
1. `goc` v1 版中，`cmd: goc profile` 轮询获取全量覆盖率
2. `goc` v2 版中，`cmd: goc watch` 实时推送增量覆盖率，下图展示的是实时推送的增量覆盖率

![goc watch 实时覆盖率](/img/posts/goc/image34.gif)

最后再展示我们实现的一个 VS Code 插件 demo，动图中随着访问不同的 HTTP API 接口，代码相应的位置被高亮了起来，研发和测试同学可以基于该插件实现精准的白盒测试。

![VS Code 实时覆盖率](/img/posts/goc/image35.gif)
