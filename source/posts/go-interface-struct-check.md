title: Golang 实现 interface 与 struct 强关联
date: 2021-09-19 12:19:02
summary: Golang 的 interface 和 struct 是非耦合的，interface 变动后，struct 如何感知呢？
---

在 [从 Rust 看 Go](/2021/07/21/from-rust-to-go/) 中，我提到“当修改/增加接口内方法签名时，波及的实现类很难一下找出，只有当这些在使用接口时才会被发现”。

确实，Golang 在语法层面没有这样的机制来做接口和实现的强关联，但是，可以借助编译器来达到相同的效果。

看下面这段的代码：

```go
type Foo interface {
	FuncA()
}

type Bar struct{}

func (b Bar) FuncA() {}
```

代码中，Bar 类型实现了 Foo 接口。这时候还没有用 Bar 类型开发代码，所以一旦 Foo 接口变动，Foo 和 Bar 之间的联系就断了。

为此，我们可以添加一行：

```go
type Foo interface {
	FuncA()
}

type Bar struct{}

func (b Bar) FuncA() {}

var _ Foo = Bar{}
```

`var _ Foo = Bar{}` 是一个类赋值给接口的语句，它在 Foo 和 Bar 之间建立了强耦合，编译器若发现 Bar 没有实现 Foo 的某些方法便会报错。而 `_` 在 Golang 属于一个丢弃值，所以最终编译所得的二进制不会占用额外的空间。

非常巧妙～