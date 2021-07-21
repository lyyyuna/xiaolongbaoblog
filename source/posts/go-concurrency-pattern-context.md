title: Golang 并发模式 - Context
date: 2020-05-25 22:19:02
categories: 语言
tags: 
- Go
---

## 前言

本文翻译+删选+理解自 [Go Concurrency Patterns: Context](https://blog.golang.org/context)

在使用 Go 编写的服务器程序中，每个请求都由一个 goroutine 来处理，通常这些请求又会启动额外的 goroutines 来访问后台数据库或者调用 RPC 服务。这些与同一个请求相关的 goroutines，常常需要访问同一个特定的资源，比如用户标识，认证 token 等等。当请求取消或者超时时，所有相关的 goroutines 都应该快速退出，这样系统才能回收不用的资源。

为此，Google 公司开发了[context](https://golang.org/pkg/context)包。该库可以跨越 API 边界，给所有 goroutines 传递请求相关的值、取消信号和超时时间。这篇文章会介绍如何使用[context](https://golang.org/pkg/context)库，并给出一个完整的例子。

## Context

`context`包的核心是`Context`结构体：

```go
// A Context carries a deadline, cancelation signal, and request-scoped values
// across API boundaries. Its methods are safe for simultaneous use by multiple
// goroutines.
type Context interface {
    // Done returns a channel that is closed when this Context is canceled
    // or times out.
    Done() <-chan struct{}

    // Err indicates why this context was canceled, after the Done channel
    // is closed.
    Err() error

    // Deadline returns the time when this Context will be canceled, if any.
    Deadline() (deadline time.Time, ok bool)

    // Value returns the value associated with key or nil if none.
    Value(key interface{}) interface{}
}
```

`Done`方法返回一个 channel，它会给`Context`上所有的函数发送取消信号，当 channel 关闭时，这些函数应该终止剩余流程立即返回。`Err`方法返回的错误指出了`Context`为什么取消。[Pipelines and Cancelation](https://blog.golang.org/pipelines)中讨论了`Done`的具体用法。

由于接收和发送信号的通常不是同一个函数，`Context`并没有提供`Cancel`方法，基于相同的理由，`Done`channel 只负责接收。尤其当父操作开启 goroutines 执行子操作时，子操作肯定不能取消父操作。作为替代，`WithCancel`函数可以用来取消`Context`。

`Context`对 goroutines 来说是并发安全的，你可以将单个`Context`传递给任意数量的 goroutines，然后取消该`Context`给这些 goroutines 同时发送信号。

`Deadline`方法用于判断函数究竟要不要运行，比如截止时间将近时，运行也就没必要了。代码可依此为 I/O 操作设置超时时间。

`Value`方法则为`Context`存储了请求所有的数据，访问这些数据必须是并发安全的。

## Context 派生

使用`Context`包提供的方法可以从已有的`Context`值派生出新值。这些派生出的值逻辑上构成了一棵树：当根`Context`取消，其派生出的子`Context`也会跟着取消。

`Background`是所有`Context`树的根，它永远不会被取消：

```go
// Background returns an empty Context. It is never canceled, has no deadline,
// and has no values. Background is typically used in main, init, and tests,
// and as the top-level Context for incoming requests.
func Background() Context
```

`WithCancel`和`WithTimeout`函数返回的派生`Context`值，可以先于父值取消。当请求的回调函数返回后，与请求相关的`Context`即可被取消。当有多个备份后台程序同时提供服务时，`WithCancel`可用于去除多余的请求。`WithTimeout`则可用于为请求设置超时时间。

```go
// WithCancel returns a copy of parent whose Done channel is closed as soon as
// parent.Done is closed or cancel is called.
func WithCancel(parent Context) (ctx Context, cancel CancelFunc)

// A CancelFunc cancels a Context.
type CancelFunc func()

// WithTimeout returns a copy of parent whose Done channel is closed as soon as
// parent.Done is closed, cancel is called, or timeout elapses. The new
// Context's Deadline is the sooner of now+timeout and the parent's deadline, if
// any. If the timer is still running, the cancel function releases its
// resources.
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
```

`WithValue`则在`Context`上存储了请求相关的值：

```go
// WithValue returns a copy of parent whose Value method returns val for key.
func WithValue(parent Context, key interface{}, val interface{}) Context
```

## 例子：谷歌网页搜索

本例子提供一个 HTTP 服务器，接收类似`/search?q=golang&timeout=1s`的 GET 请求，并把查询参数值"golang"转推到[谷歌网页搜索API](https://developers.google.com/web-search/docs/)。参数`timeout`告诉服务器如果谷歌 API 超时就取消请求。

代码分为三个包：

* [server](https://blog.golang.org/context/server/server.go)提供 main 函数，处理`/search`来的请求。
* [userip](https://blog.golang.org/context/userip/userip.go)提供的函数将用户 IP 地址绑定到一个`Context`值上。 
* [google](https://blog.golang.org/context/google/google.go)提供 Search 函数将请求转发到谷歌。

## server 程序

[server](https://blog.golang.org/context/server/server.go)中，请求回调创建了一个名为`ctx`的`Context`值，当回调退出时，延迟函数`defer cancel()`即执行取消操作。如果请求的 URL 带有 timeout 参数，那超时后`Context`会自动取消：

```go
func handleSearch(w http.ResponseWriter, req *http.Request) {
    // ctx is the Context for this handler. Calling cancel closes the
    // ctx.Done channel, which is the cancellation signal for requests
    // started by this handler.
    var (
        ctx    context.Context
        cancel context.CancelFunc
    )
    timeout, err := time.ParseDuration(req.FormValue("timeout"))
    if err == nil {
        // The request has a timeout, so create a context that is
        // canceled automatically when the timeout expires.
        ctx, cancel = context.WithTimeout(context.Background(), timeout)
    } else {
        ctx, cancel = context.WithCancel(context.Background())
    }
    defer cancel() // Cancel ctx as soon as handleSearch returns.
```

回调函数从请求中获取参数信息，并调用`userip`包获取客户端 IP 地址。由于后台服务中会使用到客户端 IP 地址，故需要将此存储于`ctx`中：

```go
    // Check the search query.
    query := req.FormValue("q")
    if query == "" {
        http.Error(w, "no query", http.StatusBadRequest)
        return
    }

    // Store the user IP in ctx for use by code in other packages.
    userIP, err := userip.FromRequest(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    ctx = userip.NewContext(ctx, userIP)
```

传入`ctx`和`query`参数调用`google.Search`：

```go
    // Run the Google search and print the results.
    start := time.Now()
    results, err := google.Search(ctx, query)
    elapsed := time.Since(start)
```

如果搜索成功，会渲染出结果：

```go
    if err := resultsTemplate.Execute(w, struct {
        Results          google.Results
        Timeout, Elapsed time.Duration
    }{
        Results: results,
        Timeout: timeout,
        Elapsed: elapsed,
    }); err != nil {
        log.Print(err)
        return
    }
```

## userip 包

[userip](https://blog.golang.org/context/userip/userip.go)包提供了解析用户 IP 地址函数的包，同时会将 IP 地址存储于一个`Context`值中。`Context`提供了键值对存储，键值都是`interface{}`类型，键必须可比较，值必须是并发安全。`userip`包屏蔽了实现上的细节，并以强类型方式访问`Context`值。

为了避免键冲突，`userip`包首先定义一个非导出类型`key`，然后用该类型定义的值作为`Context`的键：

```go
// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// userIPkey is the context key for the user IP address.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const userIPKey key = 0
```

`FromRequest`从`http.Request`解析出客户端 IP 地址`userIP`：

```go
func FromRequest(req *http.Request) (net.IP, error) {
    ip, _, err := net.SplitHostPort(req.RemoteAddr)
    if err != nil {
        return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
    }
```

`NewContext`将`userIP`存储于新建的`Context`中：

```go
func NewContext(ctx context.Context, userIP net.IP) context.Context {
    return context.WithValue(ctx, userIPKey, userIP)
}
```

`FromContext`则相反，从`Context`取出 IP 地址：

```go
func FromContext(ctx context.Context) (net.IP, bool) {
    // ctx.Value returns nil if ctx has no value for the key;
    // the net.IP type assertion returns ok=false for nil.
    userIP, ok := ctx.Value(userIPKey).(net.IP)
    return userIP, ok
}
```

## google 包

`google.Search`函数对谷歌 API 发起 HTTP 请求，并解析 JSON 结果。该函数同时接收一个`Context`参数`ctx`，如果请求处理时，`ctx.Done`关闭了，则立即退出。

谷歌 API 会将搜索内容和用户 IP 地址`userIP`作为请求参数：

```go
func Search(ctx context.Context, query string) (Results, error) {
    // Prepare the Google Search API request.
    req, err := http.NewRequest("GET", "https://ajax.googleapis.com/ajax/services/search/web?v=1.0", nil)
    if err != nil {
        return nil, err
    }
    q := req.URL.Query()
    q.Set("q", query)

    // If ctx is carrying the user IP address, forward it to the server.
    // Google APIs use the user IP to distinguish server-initiated requests
    // from end-user requests.
    if userIP, ok := userip.FromContext(ctx); ok {
        q.Set("userip", userIP.String())
    }
    req.URL.RawQuery = q.Encode()
```

`Search`使用了辅助函数`httpDo`来发起和取消 HTTP 请求，辅助函数参数有一个是处理 HTTP 响应的闭包：

```go
    var results Results
    err = httpDo(ctx, req, func(resp *http.Response, err error) error {
        if err != nil {
            return err
        }
        defer resp.Body.Close()

        // Parse the JSON search result.
        // https://developers.google.com/web-search/docs/#fonje
        var data struct {
            ResponseData struct {
                Results []struct {
                    TitleNoFormatting string
                    URL               string
                }
            }
        }
        if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
            return err
        }
        for _, res := range data.ResponseData.Results {
            results = append(results, Result{Title: res.TitleNoFormatting, URL: res.URL})
        }
        return nil
    })
    // httpDo waits for the closure we provided to return, so it's safe to
    // read results here.
    return results, err
```

`httpDo`函数会在新的 goroutine 中处理 HTTP 请求和响应，同时如果`ctx.Done`提前关闭，函数会直接退出。

## 为 Context 作代码适配

许多服务器框架已经存储了请求相关值，我们可以实现`Context`接口的所有方法，来为`Context`参数适配这些现有框架。而框架的使用者在调用代码时则需要多传入一个`Context`。

参考实现：

1. [gorilla.go](https://blog.golang.org/context/gorilla/gorilla.go)适配了 Gorilla 的[github.com/gorilla/context](http://www.gorillatoolkit.org/pkg/context)
2. [tomb.go](https://blog.golang.org/context/tomb/tomb.go)适配了[Tomb](https://godoc.org/gopkg.in/tomb.v2)
