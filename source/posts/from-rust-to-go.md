title: 从 Rust 看 Go
date: 2021-07-21 15:36:25
summary: 从 Rust 的设计哲学中可以借鉴什么？
---

## 前言

Rust 是一门着眼于安全、速度和并发的编程语言。

程序除了业务问题外，最常见的就是安全问题。从前 C++ 存在着内存管理、数据共享时“野指针”等问题。Go 和 Rust 在改善 C++ 的问题上走了两条完全不同的路。

* Go 靠垃圾回收，Java/Python 有垃圾回收+虚拟机
* Rust 则靠的是**编译器**与程序员的种种**约定**规则

在七牛，主力语言是 Go，为什么我们还要学习一门新语言 - 他山之石。新的语言，代表着对事物更新的理解和更好的阐述方式，可以帮助我们更好地编写 Go 代码。

P.S. Rust 语言学习曲线陡峭，不适合初学者，即使简单的代码也需要融汇贯通所有概念。但本文并不要求你有 Rust 背景。

## 变量可变性

首先，变量默认是**不可变的**，意味着如下的代码会报错：

```rust
let x = 1;
x = 2;
```

如果要改变，则需要显式的将变量声明为可变：

```rust
let mut x = 1;
```

这看起来增加了程序员的负担，但在多线程环境下，意味着只读不写，在编译时更易推理出潜在的并发读写问题。

## 变量所有权

这是 Rust 特有的概念。编程中，我们经常需要把一个对象传来传去，

### 理解 = 的意义

现在我们从 Rust 的角度重新看待 `=` 操作符。

Rust 强化了“所有权”的概念：

1. Rust 每一个值都有一个所有者变量与之绑定
2. 是所有者的变量只能有一个
3. 当所有着变量离开作用时，绑定解除，值被丢弃

我们分两种情况：

1. Move 语义
    * 定义变量我们会使用 `=` 符号：`let x = String::from("hello");` 。在 Rust 中这应该理解为：内存中有一个字符串，有变量 x 与之绑定，x 是该字符串的所有者
    * 复制时我们也会用 `=` 符号：`y = x`。在 Rust 中应该理解为：x 放弃对字符串的所有权，转移给 y。也就意味着使用“复制”来形容 `=` 符号，不再合适，这里应称为 `Move`
2. Copy 语义
    * `=` 符号意味着复制，`y = x`，y 得到了一份拷贝，两个变量的所有权互不干扰

变量所有权只能有一个，但 Rust 提供“借用”的方法：不可变借用 `&` 与可变借用 `&mut`：

通过借用，可实现变量的共享访问。（Rust 严格规定：在任意时刻，要么只能有一个可变引用，要么只能有多个不可变引用。）

所有权、借用、可变不可变如何防止潜在错误呢？让我们看四个例子。

例 1：

```rust
struct T(u64);

fn main() {
    let a = T(42);
    let b = bar(a);
    let c = bar(a); // 错误
}

fn bar(x: T) -> u64 {
    x.0 * 2
}
```

这里 bar 函数传参数是 `Move` 语义，第一次转移后，a 已经不再拥有原值的所有权。

例 2：

```rust
struct T(u64);

fn main() {
    let a = T(42);
    let b = bar(&a);
    let c = bar(&a);
}

fn bar(x: &T) -> u64 {
    x.0 * 2
}
```

bar 函数参数是不可变借用，所以可以重复调用。

例 3：

```rust
struct T(u64);

fn main() {
    let a = T(42);
    let b = bar(&a); 
}

fn bar(x: &T) -> u64 {
    x.0 += 1 // 错误
}
```

bar 函数参数是不可变借用，赋值操作改变了值，引起冲突。

例 4：

```rust
struct T(u64);

fn main() {
    let mut a = T(42);
    let b = bar(&mut a);
}

fn bar(x: &mut T) -> u64 {
    x.0 += 1
}
```

bar 函数参数是可变借用，操作合法。

让我们看看 Go 中如下的代码：

```go
func main() {
    arr := []int{1, 2, 3}
    if IsOdd(arr) == true {
        fmt.Println("got")
    }
}

func IsOdd(arr []int) bool {
    ...
}
```

main 函数中我们调用 `IsOdd` 函数来判断数组中是否有奇数，假如 `IsOdd` 是外部库引入，或者是由团队内其他同学提供，我们是否有把握数组不会被误更改？可见，变量传递在 Rust 中如此精细，可有效的防止类似的错误。

### 所有权与并发安全

所有权的强化也促进了并发安全，让我们看看这段 Go 代码：

```go
func main() {
    m := make(map[int]int)

    go func() {
        for {
            time.Sleep(time.Millisecond)
            _, _ = m[1]
        }
    }()

    go func() {
        for {
            time.Sleep(time.Millisecond)
            m[1]++
        }
    }()
}
```

两个 goroutine，一个读一个写，存在并发安全问题，需要加互斥锁，或者使用 `Sync.Map`。那 Rust 如何避免呢？我们看下等价的 Rust 代码：

```rust
let mut m: HashMap<u64, u64> = HashMap::new();

tread::spawn(move || loop {
    thread::sleep(time::Duration::from_millis(1));
    let _ = m.get(&1); // m 所有权被转移
})

tread::spawn(move || {
    let mut i = 0;
    loop {
        thread::sleep(time::Duration::from_millis(1));
        i += 1;
        m.insert(i, 1); // 已被转移，错误
    }
})
```

首先闭包传递要求 `Move` 语义，所以字典 m 的所有权会被移入第一个线程。当第二个线程再使用字典 m 时已无所有权，编译器便会报错，阻止你用错误的方法并发访问字典 m。Rust 另有正确方法来并发读写（使用 Arc 和 Mutex），这里不再介绍。

Rust 与 Go 相比：

* Go 中既可以正确的编写并发代码，也可以错误的编写并发代码，编译器不管
* Rust 中错误的并发方法无法通过编译

## 资源管理

### 传值 vs 传引用

变量专递还存在着经典的“传值 vs 传引用”问题。

比如：

```go
a := 1
b = a
b = 6
fmt.Println(a, b) // 1, 6
```

与

```go
a := []int{1, 2, 3}
b = a
b[1] = 6
fmt.Println(a, b) // {1, 6, 3} {1, 6, 3}
```

Go 初学者分不清区别，老手一不留神也会搞错。

这里问题根源和资源管理的方式有关。变量在内存中一般有两种方式：

1. 栈管理
2. 栈 + 堆管理

### 栈管理

* 函数调用时，会压栈，调用结束返回上一层函数时，会弹栈（处理器支持）
* 对于基础类型/非变长类型的函数内局部变量，可以直接在当前函数栈内分配，这些栈内分配的内存就是资源
* 由于弹栈操作，分配的内存即被回收，无需特殊的逻辑处理

### 栈 + 堆管理

* 对于变长类型（Go 语言的 map/slice，Rust 的 Vec<T> / Map<K, V> 等等）无法在栈内预先分配内存
* 在栈上存放指针，指针本身大小确定，指针指向堆，堆上分配的内存大小可变
* Go 用垃圾回收释放堆内存
* Rust 由上文提到的所有权保证离开作用域时释放

### 对 Go 的思考

再来看“传值 vs 传引用”的问题。无论传值还是传引用，都是对栈管理的值（部分值）的拷贝，即：

* 所谓传值，栈拷贝复制了值的所有部分
* 所谓传引用，栈拷贝只复制了栈上的指针部分，堆的部分没有复制。两个指针指向同一个堆

## 空值与错误处理

ust 有强大的类型系统，支持 enum + 模板类型。它将空值定义为

```rust
struct Option<T> {
    Some(T),
    None,
}
```

将错误定义为

```rust
enum Result<T, E> {
   Ok(T),
   Err(E),
}
```

### 空值

Go 在测试与生产环境中难免遇到空指针异常，比如：

```go
type struct A {
    str *string
    dic map[int]string
}

func (a *A) test() {
    *a.str += "world"
    a.dic[1] = "hello"
}

aa = A{}
aa.test()
```

go 会对变量默认初始化，所以 `aa.str` 得到的是一个未指向任何 string 的空指针，`aa.dic` 是也未指向 map。调用 `aa.test()` 就会发生多种空指针异常：

```
panic: runtime error: invalid memory address or nil pointer dereference

panic: assignment to entry in nil map
```

类似代码在 Rust 里面会怎么样呢？

```rust
struct A {
    ss: Option<Box<String>>,
    dic: HashMap<i64, String>,
}

impl A {
    fn test(&self) {
        match self.ss {
            Some(_) => ...,
            None => ...
        } 
    }
}
```

Rust 要求必须显式初始化，dic 未指向 map 的问题就解了。然后，无论成员变量 ss 初始化为 `Some(T)` 还是 `None`，match 语法会要求程序员对每种情况都编码，从而避免“空指针”。

### 错误处理

Go 和 Rust 都没有使用抛异常，而是返回 err 的方式来处理错误。比如 Go：

```go
result, err := do_something()
if err != nil {
    return nil, err
}
```

Go 采用多返回值方式，程序报错返回错误问题，通过判断 `err != nil` 来决定程序是否继续执行或终止该逻辑。当然，如果接触过 Go 项目时，会发现程序中大量充斥着 `if err != nil` 的代码，判断是手动逻辑，往往我们可能因为疏忽，导致这段逻辑缺失，缺少校验。

Rust 里怎么做呢：

```rust
fn do_something() -> Result<u64> {
    Ok(4)
}

let result = do_something();
match result {
    Ok(_) => {},
    Err(_) => {},
}
```

首先，有 match 语句保证每个枚举值必须得到处理，否则编译器就会报错。进一步的，无论有没有错误返回，上层逻辑只需要面对一个值（即例子中的 result），多个函数可以实现链式调用。

## 面向接口编程

Go 的 interface 和 Rust 的 trait 类似，都是面向接口编程，但有些差别：

1. Go 不需要 struct 显式地指定 interface 实现：它只需要实现接口中定义的所有方法。它们之间是松耦合的关系，靠编译器最终编译时才能串联
2. Rust 需要显式声明 struct 实现某个 trait。而且，Rust 还支持为不是自己定义的类型增加 trait 实现。

在我看来，Go 这种松耦合关系有几个缺点。

首先，通过文档（godoc），很难一眼看出类型是否符合特定接口，比如，[TCPConn 类型](https://golang.org/pkg/net/#TCPConn)，初看文档，完全不知道它是否符合 [Writer 接口](https://golang.org/pkg/io/#Writer) 和 [Reader 接口](https://golang.org/pkg/io/#Reader)，仔细比对方法签名，才能确认。

然后就是当修改/增加接口内方法签名时，波及的实现类很难一下找出，只有当这些在使用接口时才会被发现。

Rust 的 trait 实现强制声明就很好的解决了上述两个痛点。对于文档，类型所有实现的 trait 都一目了然。而当 trait 变动，而类型定义却没有更改时编译器会报错。

## 包管理

Go 的包管理器 `go mod` 起步太晚，Go 1.13 才迈入生产环境，而且其设计理念过于理想化，在主流语言的包管理中独树一帜，现在讨论它优劣还为时过早，可参考笔者做的相关[分享](http://www.lyyyuna.com/2020/02/22/go-the-principles-versioning-in-go/)。

Rust 的包管理 `cargo` 很早便有了，它不仅是包管理工具，更是项目组织管理工具，从项目的建立、构建到测试、运行直至部署，为 Rust 项目的管理提供尽可能完整的手段。

## 总结

我们总结出一些有助于提升 Go 代码安全性的 Tips：

1. 思考变量是否是可变的
2. 思考变量是否是共享的，并加以并发保护
3. 思考变量是值传递还是引用传递，避免副作用
4. 不要遗漏 err 和 nil
5. 涉及接口的变动要慎重

Rust 特性和编程范式极多，本文不可能一一阐述，有兴趣的同学可以移步[官方教程](https://doc.rust-lang.org/book/)。