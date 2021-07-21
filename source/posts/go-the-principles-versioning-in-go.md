title: Golang 的版本管理原则
date: 2020-02-22 16:20:02
categories: 语言
tags: 
- Go
summary: Go Modules 与其他语言的包管理工具有什么不同？
---

## 前言

本文翻译+删选+理解自 [The Principles of Versioning in Go](https://research.swtch.com/vgo-principles)

## 为什么需要版本？

让我们先看下传统基于`GOPATH`的`go get`是如何导致版本管理失败的。

假设有一个全新安装的 Go 环境，我们需要写一个程序导入`D`，因此运行`go get D`。记住现在是基于`GOPATH`的`go get`，不是 `go mod`。

```
$ go get D
```

![](/img/posts/version-in-go/vgo-why-1@1.5x.png)

该命令会寻找并下载最新版本的`D 1.0`，并假设现在能成功构建。

几个月后我们需要一个新的库`C`，我们接着运行`go get C`，该库的版本为 1.8。

```
$ go get C
```

![](/img/posts/version-in-go/vgo-why-2@1.5x.png)

`C`导入`D`，但是`go get`发现当前环境内已经下载过`D`库了，所以 Go 会重复使用该库。不幸的是，本地的`D`版本是 1.0，而`C`对`D`有版本依赖，必须是 1.4 以上（有可能 1.4 有一些 bugfix 或者新 feature）。

显而易见这里`C`会构建失败。我们再次运行`go get -u C`。

```
$ go get -u C
```
![](/img/posts/version-in-go/vgo-why-3@2x.png)

不幸的是，（假设）一小时前`D`的作者发布了`D 1.6`，该版本又引入了一个缺陷。因为`go get -u`一直使用最新的依赖，所以使用 1.6 的`C`又构建失败了。

由这个例子可以看出，基于`GOPATH`的`go get`缺乏版本管理，会导致两种问题，要么版本过低，要么版本过高。我们需要一种机制，`C`和`D`的作者能够一起开发和测试。

自从`goinstall/go get`推出之后，Go 程序员就对版本管理有着强烈的诉求，过去几年间，很多第三方的工具被开发出来。然而，这些工具对版本控制的细节有着不同的实现和理解，这会导致不同的库若使用不同的工具，库之间仍然无法协同工作。

## 软件工程中的版本

过去两年间（2019），官方试图在`go`命令中引入`Go moduless`的概念来支持版本管理。`Go moduless`带来新的库导入语法——即语义化导入版本(Semantic import versioning)，而在选择版本时，使用了新的算法——即最小版本选择算法。

你可能会问：为什么不使用其他语言的现成经验？Java 有`Maven`，Node 有`NPM`，Ruby 有`Bundler`，Rust 有`Cargo`，他们解决版本依赖的思路不好么？

你可能还会问：Go 团队在 2018 早些时候引入了一个试验性的工具`Dep`，该工具实现上与`Bundler`和`Cargo`一致，现在为啥又变卦了？

答案是我们从使用`Bundler`/`Cargo`/`Dep`的经验中发现，它们所谓处理依赖的方法，只会使项目越来越复杂，`go modules`决定另辟蹊径。

## 三原则

回到一个很基础的问题：什么是软件工程？软件工程和编程有什么区别？原作者 Russ Cox 使用了这个定义：

> Software engineering is what happens to programming
> when you add time and other programmers.

为了简化软件工程，`Dep`和`Go moduless`在原则上有三个显著的改变，它们是兼容性、可重复性和可合作性。本文余下部分会详细阐述这三个指导思想。

### 原则 #1：兼容性

第一原则是兼容性，或者称之为稳定性，程序中**名字**的意义不能随着时间改变。一年前一个名字的含义和今年、后年应该完全一致。

例如，程序员经常会对标准库`string.Split`的细节困扰。我们期望在`"hello world"`调用后产生两个字符串`"hello"`和`"world`。但是如果函数输入有前、后或着重复的空格，输出结果也会包含空字符串：

```
Example: strings.Split(x, " ")

"hello world"  => {"hello", "world"}
"hello  world" => {"hello", "", "world"}
" hello world" => {"", "hello", "world"}
"hello world " => {"hello", "world", ""}
```

假设我们决定改变这一行为，去除所有空字符串，可以么？

**不**

因为我们已经在旧版`string.Split`的文档和实现上达成一致。有无数的程序依赖于这一行为，改变它会破话兼容性原则。

对于新的行为，正确的做法是给一个新的名字。事实上也是如此，我们没有重新定义`strings.Split`，几年前，标准库引入了`strings.Fields`函数。

```
Example: strings.Fields(x)

"hello world"  => {"hello", "world"}
"hello  world" => {"hello", "world"}
" hello world" => {"hello", "world"}
"hello world " => {"hello", "world"}
```

遵守兼容性原则可以大大简化软件工程。当程序员理解程序时，你无需把时间纳入考量范围内，2015 年使用的`strings.Split`和今年使用的`strings.Split`是一样的。工具也是如此，比如重构工具可以随意地将`strings.Split`在不同包内移动而不用担心函数含义随着时间发生改变。

实际上，Go 1 最重要的特性就是其语言不变性。这一特性在官方文档中得到明确，[golang.org/doc/go1compat](golang.org/doc/go1compat)：

> It is intended that programs written to the Go 1 specification will continue to compile and run correctly, unchanged, over the lifetime of that specification. Go programs that work today should continue to work even as future “point” releases of Go 1 arise (Go 1.1, Go 1.2, etc.).

所有 Go 1.x 版本的程序在后续版本仍能继续编译，并且正确运行，行为保持不变。今天写了一个 Go 程序，未来它仍能正常工作。Go 官方同样也对标准库中的函数作出了承诺。

兼容性和版本有啥管理？当今最火的版本管理方法——[语义化版本](https://semver.org/)是鼓励不兼容的，这意味着你可以通过语义化版本号，更轻易地作出不兼容的改变。

如何理解？

语义化版本有着`vMAJOR.MINOR.PATCH`的形式。如果两个版本有着系统的主版本号，那么后一个版本应该向前兼容前一个版本。如果主版本号不同，那他俩就是不兼容的。该方法鼓励包的作者，如果你想作出不兼容的行为，那改变主版本号吧！

对于 Go 程序来说，光改变主版本号还不够，两个主版本号如果名字一模一样，阅读代码还是会造成困扰。

看起来，情况变得更加糟糕。

假设包`B`期望使用 V1 版本的`string.Split`，而`C`期望使用 V2 版本的`string.Split`。如果`B`和`C`是分别构建的，那 OK。

![](/img/posts/version-in-go/vgo-why-4@2x.png)

但如果有一个包`A`同时导入了包`B`和`C`呢？那该如何选择`string.Split`的版本？

![](/img/posts/version-in-go/vgo-why-5@2x.png)

针对`Go modules`的设计思想，官方意识到兼容性是最基础的原则，是必须支持、鼓励和遵循的。Go 的 FAQ 中写到：

> Packages intended for public use should try to maintain backwards compatibility as they evolve. The Go 1 compatibility guidelines are a good reference here: don’t remove exported names, encourage tagged composite literals, and so on. If different functionality is required, add a new name instead of changing an old one. If a complete break is required, create a new package with a new import path.

大致意思是如果新旧两个包导入路径相同，那它们就应该被当作是兼容的。

这和语义化版本有什么关系呢？兼容性原则要求不同主版本号之间不需要有兼容性上的联系，所以，很自然地要求我们使用不同的导入路径。而`Go modules`中的做法是把主版本号放入导入路径，我们称之为语义化导入版本(Semantic import versioning)。

![](/img/posts/version-in-go/impver@2x.png)

在这个例子中，`my/thing/v2`表示使用版本 2。如果是版本 1，那就是`my/thing`，没有显式在路径指定版本号，所以路径成为了主版本号的一部分，以此类推，版本 3 的导入路径为`my/thing/v3`。

如果`strings`包是我们开发者自己的模块，我们不想增加新的函数`Fields`而是重新定义`Split`，那么可以创建两个模块`strings`(主版本号 1)和`strings/v2`(主版本号 2)，这样可以同时存在两个不同的`Split`。


![](/img/posts/version-in-go/vgo-why-6@2x.png)

依据此路径规则，`A`、`B`和`C`都能构建成功，整个程序都能正常运行。开发者和各种工具都能明白它们是不同的包，就像`crypto/rand`和`math/rand`是不同的一样显而易见。

让我们回到那个不可构建的程序。把`strings`抽象成包`D`，，这时候若不使用*语义化导入版本方法*，这样就遇到了经典的“钻石依赖问题”：`B`和`C`单独都能构建，但放在一起就不行。如果尝试构建程序`A`，那该如何选择版本`D`呢？

![](/img/posts/version-in-go/vgo-why-7@2x.png)

语义化导入版本切断了“钻石依赖”。因为`D`的版本 2.0 有不一样的导入路径，`D/v2`。

![](/img/posts/version-in-go/vgo-why-8@2x.png)

### 原则 #2：可重复性

第二原则是程序构建必须具有可重复性，一个指定版本包的构建结果不应随时间改变。在该原则下，今天我编译代码的结果和其他程序员明年编译的结果是匹配的。**大部分包管理系统并不保证一点**。

在第一小节我们也看到了，基于`GOPATH`的`go get`用的不是最新就是最旧的`D`。你可能认为，“`go get`当然会犯错误：它对版本一无所知”。但其实其他一些包管理工具也会犯同样的错误，这里以`Dep`为例。（`Cargo`和`Bundler`也类似）

`Dep`要求每一个包包含一个`manifest`来存放元数据，记录下对所有依赖的要求。当`Dep`下载了`C`，它读入`C`的元数据，知道了`C`需要`D 1.4`之后的版本。然后`Dep`下载了最新版本的`D`来满足这一限制。

假设在昨天，`D`最新版本是 1.5：

![](/img/posts/version-in-go/vgo-why-9@2x.png)

而今天，`D`更新为了 1.6：

![](/img/posts/version-in-go/vgo-why-10@2x.png)

可以看出，该决策方法是不可重复的，会随时间发生改变。

当然，`Dep`的开发者意识到了这一点，它们引入了第二个元数据文件——lock 文件。如果`C`本身是一个完整的程序，当 Go 调用`package main`的时候，lock 文件会记录下`C`使用依赖的确切版本，而当需要重复构建时，lock 文件内所记录的依赖具有更高的优先级。也就是说，lock 文件同样能保证重复性原则。

但 lock 文件只是针对整体程序而言——`package main`。如果`C`被别的更大程序所使用，lock 文件就无效了，库`C`的构建仍会随着时间的改变而改变。

而`Go modules`的算法非常简单，那就是“最小版本选择算法”——每一个包指定其依赖的最低版本号。比如假设`B 1.3`要求最低`D 1.3`，`C 1.8`要求最低`D 1.4`。`Go modules`不会选择最新的版本，而是选择最小能满足要求的版本，这样，构建的结果是可重复的。

![](/img/posts/version-in-go/vgo-why-12@2x.png)

如果构建的不同部分有不同最低版本要求，`go`命令会使用最近的那个版本。如图所示，`A`构建时发现同时有`D 1.3`和`D 1.4`的依赖，由于 1.4 大于 1.3，所以构建时会选择`D 1.4`。`D 1.5`或者`D 1.6`存在与否并不会影响该决策。

在没有 lock 文件的情况下，该算法依然保证了程序和库的可重复性构建。

### 原则 #3：可合作性

第三原则是可合作性。为了维护 Go 包的生态，我们追求的是一个统一的连贯的系统。相反，我们想避免的是生态分裂，变成一组一组互相之间不可合作的包。

若开发者们不合作，无论我们使用的工具有多么精巧，技巧多么高超，整个 Go 开源生态一定会走向分裂。这里隐含的意思是，为了修复不兼容性，必须要合作，我们不应排斥合作。

还是拿库`C 1.8`举例子，它要求最低版本`D 1.4`。由于可重复性原则，`C 1.8`构建会使用`D 1.4`。如果`C 1.8`是被其他更大的程序所依赖，且该程序要求`D 1.5`，那根据最小版本选择算法，会选择`D 1.5`。这时候构建仍是正确的。

现在问题来了，`D` 的作者发布了 1.6 版本，但该版本有问题，`C 1.8`无法与该版本构建。

![](/img/posts/version-in-go/vgo-why-13@2x.png)

解决的方法是`C`和`D`的作者合作来发布 fix。解决方法多种多样。

`C` 可以推出 1.9 版本，规避掉`D 1.6`中的 bug。

![](/img/posts/version-in-go/vgo-why-15@2x.png)

`D` 也可以推出 1.7 版本，修复其存在的 bug。同时，根据兼容性原则，`C 1.9`可以指定其要求最低`D 1.7`。

![](/img/posts/version-in-go/vgo-why-14@2x.png)

再来复盘一下刚才的故事，最新版本的`C`和`D`突然不能一起工作了，这打破了 Go 包的生态，两库的作者必须合作来修复 bug。这种合作对生态是良性的。而正由于`Go modules`采用的包选择算法/可重复性，那些没有显式指定`D 1.6`的库都不会被影响。这给了`C`和`D`的作者充分的时间来给出最终解决方案。

## 结论

以上是 Go 版本管理的三原则，也是`Go modules`区别于`Dep`，`Bundler`和`Cargo`的根本原因。

* 兼容性，程序中所使用的名字不随时间改变。
* 可重复性，指定版本的包构建结果不随时间改变。
* 可合作性，为了维护 Go 包的生态，互相必须易于合作。

三原则来自于对年复一年软件工程的思考，它们互相巩固，是一个良性的循环：兼容性原则使用的版本选择算法带来了可重复性。而可重复性保证除非开发者显式指定，否则构建不会使用最新的、或是有问题的库，这给了我们时间来修复问题。而这种合作性又能保证兼容性。

`Go 1.13`中，`Go modules`已经可用于生成环境，很多公司，包括 Google 已经接纳了它。`Go 1.14`和`Go 1.15`会带来更多方便开发者的特性，它的最终目标是彻底移除`GOPATH`。

具体`Go modules`的使用方法，可以参考这个系列博客[Using Go Modules](https://blog.golang.org/using-go-modules)。