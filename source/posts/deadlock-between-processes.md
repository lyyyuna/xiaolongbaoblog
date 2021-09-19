title: exec.Command 中一个有趣的多进程死锁例子
date: 2021-08-05 16:48:25
summary: 什么情况会在不经意间造成多进程死锁？
---

## 前言

前几天，公司开发的一个工具在某个工程中总是卡死，进入容器中再次运行，又能顺利运行，感觉挺有意思，于是便 debug 了一下。

这个工具使用 Golang 编写，暂未开源，所以本文只展示部分代码。

## 问题描述

这个工具大致流程是

1. 使用 `exec.Command` 调用 `go list -json all` 命令
2. 读取 `stdout` / `stderr`，再使用 `json.NewDecoder` 解析

代码大致为

```go
cmd := exec.Command("go", "list", "-json", "all")
stdouIn, _ := cmd.StdoutPipe()
stderrIn, _ := cmd.StderrPipe()

var stderrBuf bytes.Buffer

cmd.Start()

dec := json.NewDecoder(stdouIn)
dg = &DepGraph{}
for {
    var di DepInfo
    dec.Decode(&di)
}

_, errStderr := io.Copy(&stderrBuf, stderrIn)
if errStderr != nil {
    panic(err)
}

cmd.Wait()

fmt.Println("done")
```

问题：

1. 程序未 `panic` 异常退出
2. 程序未输出 `done`，即未正常退出，卡死
3. 命令后手动执行 `go list -json all` 可正常快速退出

好吧-_-，粗看程序一点问题没有，但就是卡了～

## 调试记录

### 怀疑 1，go list 有 bug？

使用 dlv 挂载 go list 进程 `dlv attach [pid]`。

查看所有 goroutine 的状态：

```
(dlv) grs
  Goroutine 836 - User: /usr/local/go/src/runtime/sema.go:61 internal/poll.runtime_Semacquire (0x447782) [semacquire 450776h20m59.570725376s]
  Goroutine 837 - User: /usr/local/go/src/runtime/sema.go:61 internal/poll.runtime_Semacquire (0x447782) [semacquire 450776h21m0.466524491s]
  Goroutine 838 - User: /usr/local/go/src/runtime/proc.go:310 sync.runtime_notifyListWait (0x4489c8) [sync.Cond.Wait 450776h20m57.704548061s]
  Goroutine 839 - User: /usr/local/go/src/syscall/asm_linux_amd64.s:24 syscall.Syscall (0x4b2efb) (thread 65) [GC assist marking 450776h21m0.664929027s]
  Goroutine 840 - User: /usr/local/go/src/runtime/sema.go:61 internal/poll.runtime_Semacquire (0x447782) [semacquire 450776h20m59.617378098s]
```

`839 goroutine` 停在了 `syscall.Syscall` 系统调用上，看起来比较可疑，详细查看该 goroutine 的调用栈：

```
(dlv) gr 839
Switched from 0 to 839 (thread 65)
(dlv) bt
...
4 0x00000000004d1267 in os.(*File).write
at /usr/local/go/src/os/file_unix.go:280
5 0x00000000004d1267 in os.(*File).Write
at /usr/local/go/src/os/file.go:153
6 0x00000000004dc125 in fmt.Fprintf
at /usr/local/go/src/fmt/print.go:205
7 0x00000000008bae71 in cmd/go/internal/modfetch.DownloadZip.func1
at /usr/local/go/src/cmd/go/internal/modfetch/fetch.go:176
8 0x0000000000791043 in cmd/go/internal/par.(*Cache).Do
at /usr/local/go/src/cmd/go/internal/par/work.go:128
9 0x00000000008ae283 in cmd/go/internal/modfetch.DownloadZip
at /usr/local/go/src/cmd/go/internal/modfetch/fetch.go:163
10 0x00000000008ad8be in cmd/go/internal/modfetch.download
...
```

查看对应 Golang 版本的 `go/src/cmd/go/internal/modfetch/fetch.go:176` [源码](https://github.com/golang/go/blob/d571a77846dfee8efd076223a882915cd6cb52f4/src/cmd/go/internal/modfetch/fetch.go#L176)：

```
if cfg.CmdName != "mod download" {
	fmt.Fprintf(os.Stderr, "go: downloading %s %s\n", mod.Path, mod.Version)
}
```

大致浏览一下上下文，可以知道目前 `go list -json all` 正在下载 `go mod` 依赖，并把下载结果打印到 `os.Stderr`。

而现在输出正卡死在 `os.Stderr` 上。这和我们的直觉相违，`stdout` / `stderr` 这两个应该不会阻塞的呀。

### 怀疑 2，卡死在工具上？

按照怀疑 1 的调试步骤，查看工具本身的调用栈：

```
(dlv) gr 1
Switched from 0 to 1 (thread 25)
(dlv) bt
8 0x00000000004f8aeb in encoding/json.(*Decoder).refill
at /usr/local/go/src/encoding/json/stream.go:165
9 0x00000000004f887f in encoding/json.(*Decoder).readValue
at /usr/local/go/src/encoding/json/stream.go:140
10 0x00000000004f843c in encoding/json.(*Decoder).Decode
at /usr/local/go/src/encoding/json/stream.go:63
11 0x000000000067555e in github.com/ma6174/go_dep_search/depgraph.LoadDeps
at /Users/xxxxx/go/pkg/mod/github.com/ma6174/go_dep_search@v0.0.0-20200721060312-bfd635bcc992/depgraph/depgraph.go:195
12 0x00000000006a73ef in main.listAndSearch
```

确认，卡死在 `dec.Decode(&di)` 上，即卡死在 `di io.Reader` 上。

### 阶段结论

1. `go list -json all` 卡死在输出 `stderr`
2. 工具卡死在读 `stdin`

那很明显了，工具进程和 `go list` 进程出现了某种死锁。

## 分析

首先，Golang 中 `exec.Command` 会创建一个子进程，在 Linux 系统上，`cmd.StderrPipe` 会调用 pipe(管道) 来作为父子进程间通信的方式：

```go
StderrPipe() -> os.Pipe() -> syscall.Pipe2(p[0:], syscall.O_CLOEXEC)
```

然后，`go list` 会先使用 `go mod download` 下载依赖的包，当一个包下载完成，便会输出一行 `go: downloading xxx` 到 `stderr` 上。在所有包下载完毕后，才会开始包依赖分析，最后才将分析结果输出到 `stdout`。

回看工具的代码，可以看到父子进程间其实是如下的关系：

```
父进程(工具) -> 读取 stdout pipe  -> 读取 stderr pipe 
子进程(go list) -> 写到 stderr pipe  -> 写到 stdout pipe 
```

两者读写的 pipe 交错了起来，`go list` 写完 `stderr pipe` 才会写 `stdout pipe`，而工具读完 `stdout pipe` 才会读 `stderr pipe`。

在 Linux 上，pipe 大小是有限的，普通情况下，上面的逻辑并没有暴露问题。但如果下载的依赖包特别多，导致 `stderr pipe` 被塞满，`go list` 便会阻塞在写 `stderr pipe` 上，后续的 `stdout pipe` 也就走不到了，而父进程一直在等待 `stdout pipe` 输出的数据。。形成一个死锁。

## 解决方案

解决方案很简单，直接使用标准库提供的 `func (c *Cmd) Output() ([]byte, error)` 或 `func (c *Cmd) CombinedOutput() ([]byte, error)` 来获取子进程输出。

如非必要，不要直接操作 pipe。

## 联想

pipe 的死锁问题让我联想到另一个经典的 socket 编程死锁案例。socket 编程也是一种多进程/跨机多进程通信方式。

创建一个 TCP 链接，一端往另一端发送数据，会死锁吗？答案是会的。看下面的例子：

```go
// 接收端
func main() {
	ln, err := net.Listen("tcp", ":12345")
	if err != nil {
		panic(err)
	}

	for {
		_, err := ln.Accept()
		if err != nil {
			panic(err)
		}
	}
}

// 发送端
func main() {
	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		panic(err)
	}
	i := 0
	for {
		_, err := conn.Write([]byte("1"))
		if err != nil {
			panic(err)
		}
		i++
		fmt.Println(i)
	}
}
```

在我的电脑上，发送端输出如下：

```
1
2
...
2616001
2616002
2616003
```

打印到 `2616003` 便停止输出卡住了，阻塞在了 `conn.Write([]byte("1"))`。表象和上文 pipe 的例子非常相似。

分析一下发送端程序，当调用 `conn.Write([]byte("1"))` 后，函数成功返回，从程序员角度看，似乎接收端这时候应该是已经收到了 `1`。但其实接收端调用 `ln.Accept()` 之后，并没有去读取连接上的数据。收到的数据去哪了？使用 `netstat` 命令看一下：

```
$ netstat -anp
Proto Recv-Q Send-Q  Local Address           Foreign Address         State       PID/Program name
tcp        0 2548608 127.0.0.1:46196         127.0.0.1:12345         ESTABLISHED 70072/main
tcp6   67395       0 127.0.0.1:12345         127.0.0.1:46196         ESTABLISHED 69896/main
```

可以看到，数据都堆积在了接收队列 `Recv-Q` 和发送队列 `Send-Q`，而且正好 `2548608+67395=2616003`。

P.S. TCP 协议的可靠性，只是保证双方的协议栈能可靠收发数据，应用程序的可靠性需要应用层协议来保证。
