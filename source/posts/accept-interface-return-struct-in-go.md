title: 传入抽象接口，返回具体类型
date: 2021-12-08 12:19:02
summary: 和 preemptive interface 说再见
---

## Go 接口的误用

在大多数强类型语言中，接口被用作描述一组类型共有的行为。比如：

```go
type Auth interface {
    GetUser() (User, error)
}
type authImpl struct {
    // ...
}
func NewAuth() Auth {
    return &authImpl
}
```

使用接口好处多多，可以让不同模块之间更好的解藕，单元测试 mock 的时候也更加方便。Java 程序员太喜欢用接口了，以至于大部分转型成 Go 的程序员认为使用接口理所当然，似乎只要充满了接口，程序架构就上了个档次。真的是这样吗？

我们来看一个例子，假设有一个生产者接口 `Thinger`：

```go
package producer

type Thinger interface { Thing() bool }

type defaultThinger struct{}
func (t defaultThinger) Thing() bool {}

func NewThinger() Thinger { return defaultThinger{} }
```

`defaultThinger` 是具体的 `Thinger` 生产者，这里初始化函数返回的是一个 `Thinger` 接口。

然后是消费者代码：

```go
package consumer  // consumer.go

func Foo(t Thinger) {
    t.Thing()
}
```

消费者调用时使用的是 `Thinger` 接口，另外定义 mock，方便实现单元测试：

```go
type mockThinger struct{}
func (t mockThinger) Thing() bool {}

func NewMock() Thinger { return mockThinger{} }
```

然而，这里接口的引入并没有让两个模块彻底解藕，这段代码隐含的问题是生产者接口的任何改动，都会传导至所有的消费者。我们仔细探讨一下。

作为消费者的函数 `Foo(t Thinger)`，它对接口的需求只是 `Thing()` 方法。作为生产者 `type Thinger interface` 而言，它可能面对不止一个消费者，所以一旦其他消费者有了新需求，生产者接口必须新增方法。比如：

```go
type Thinger interface { 
    Thing() bool 
    AnotherThing() bool
}
```

这个方法并不是 `Foo` 所需的，但它也不得不为此改动代码，比如 `mockThinger` 类型得实现一个 `AnotherThing()` 才能让单测代码编译，尽管 `AnotherThing()` 和单测毫无关系。

Go 不像 Java 那样在语法层面有机制来确保接口和实现的强关联，可见，误用滥用 Go 的接口并不一定能彻底解藕。

## 传入抽象接口，返回具体类型

之所以上面举例时要特意用“生产者”和“消费者”这两个名词，是因为解藕的关键就在于理解它们。

- 代码中具体功能的提供方为生产者
- 代码中使用功能方为消费者
- 相同的功能可以由多个生产者提供，例如读取输入，可以从网络、磁盘、终端上读取

在 Go 中，接口的真正用途是**明确**消费者是对某个功能的需求，例如 `io.Copy` 接口：

```go
func Copy(dst io.Writer, src io.Reader) (written int64, err error)
```

它希望 `src` 可以调用 `Read` 方法，`dst` 可以调用 `Write` 方法。它可以是一个 `net.TCPConn` TCP 连接，也可以是一个 `os.File` 文件描述符，它们都是具体的生产者。而 `io.Reader` 和 `io.Writer` 接口则是随着消费者 `io.Copy` 一起定义的，所以 `net.TCPConn` 和 `os.File` 类型可以放心的提供其他方法而不破坏生产者和消费者之间的约定。

小结一下：

- 生产者返回具体类型
- 由消费者定义接口

## 示例改进

明确了以上原则之后，回看之前的示例，问题的症结点一是作为生产者不应该自己定义接口把自己框死，初始化函数应改为返回 `defaultThinger` 类型：

```go
package producer

type defaultThinger struct{}
func (t defaultThinger) Thing() bool {}
func (t defaultThinger) AnotherThing() bool {}

func NewThinger() defaultThinger { return defaultThinger{} }
```

然后把接口定义挪入消费者中：

```go
package consumer  // consumer.go

type Thinger interface { Thing() bool }

func Foo(t Thinger) {
    t.Thing()
}

// mock
type mockThinger struct{}
func (t mockThinger) Thing() bool {}

func NewMock() Thinger { return mockThinger{} }
```

生产者可以随时新增 `AnotherThing()` 方法，而消费者 `Foo` 和相关的单测 mock 不受影响。