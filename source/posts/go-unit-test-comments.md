title: Golang 官方关于单元测试方法的一些建议
date: 2021-01-01 12:20:02
categories: 测试
tags: 
- Go
---

## 前言

本文是翻译+理解+改编自 Go 官方的[TestComments](https://github.com/golang/go/wiki/TestComments)，原文是 Go 源码本身开发时`Code Review`的注意项。

## 断言

在测试时避免使用断言库。那些有类似`xUnit`测试框架使用背景的 Go 开发者喜欢写如下的代码：

```go
assert.isNotNil(t, "obj", obj)
assert.stringEq(t, "obj.Type", obj.Type, "blogPost")
assert.intEq(t, "obj.Comments", obj.Comments, 2)
assert.stringNotEq(t, "obj.Body", obj.Body, "")
```

但有的断言库会过早的终止测试用例（如果断言中调用了`t.Fatalf`或者`panic`），有的会漏掉测试如何通过的关键信息。测试应该是精确的，能够一眼看出哪些部分导致用例失败，哪些部分是正确的。不仅如此，放着 Go 的语法不用，断言库却常常创造自己的一整套语法来做非空判断（`isNotNil`）、字符串比较（`stringEq`）、表达式求值等等。

综上，上面那个例子应该改写为：

```go
if obj == nil || obj.Type != "blogPost" || obj.Comments != 2 || obj.Body == "" {
    t.Errorf("AddPost() = %+v", obj)
}
```

## 使用可读性高的子测试用例名

当使用`t.Run`来创建子测试时，第一个参数是用例的名字。为了确保测试结果在日志上具备高可读性，用例名应该描述要测试的场景，并保证在转义后仍可读（测试用例执行时，会将空格转换为下划线，并转义不可打印的字符）。

也可以在子测试的函数体中使用`t.Log`打印子测试用例名，或者包含在失败信息中，这两个地方用例名都不会被转义。

## 直接比较结构体

如果函数返回的是结构体，不推荐一个一个字段的比较，而是构造出你预期的结果，用下文提到的[cmp 方法](#相等性比较和diff)直接比较。该规则同样适用于数组和字典。

如果结构体之间是某种语义上的相等，或者某些字段不支持比较操作（比如类型为`io.Reader`的字段），那你可以在[cmp.Diff](https://godoc.org/github.com/google/go-cmp/cmp#Diff)和[cmp.Equal](https://godoc.org/github.com/google/go-cmp/cmp#Equal)的参数中传入类型为[cmpopts](https://godoc.org/github.com/google/go-cmp/cmp/cmpopts)的[cmpopts.IgnoreInterfaces](https://godoc.org/github.com/google/go-cmp/cmp/cmpopts#IgnoreInterfaces)来忽略它们。要是还不行，那，就自由发挥吧。

如果函数返回多个结果，逐个比较并打印，不必拼成一个结构体。

## 只比较稳定的结果

如果被测函数依赖的外部包不受控制，导致输出结果不稳定，就该避免在测试中使用该结果。相反，应该去比较那些在语义上稳定的信息。

那些输出格式化/序列化字符串的功能，不该假设其输出的字符串一尘不变。举个实际的例子，`json.Marshal`并不保证输出的字节流永远是相同的，历史上该函数的实现变动过。如果从字符串是否相等的角度去测试`json`库，那测试用的执行结果无法稳定。而鲁棒性的做法应去解析 JSON 字符串，然后比较其中的每个对象。

## 相等性比较和diff

`==`操作符会按照[Go 语言规范](https://golang.org/ref/spec#Comparison_operators)定义的行为执行相等比较。数字、字符串和指针可以执行比较操作，结构体的每个字段如果都是以上三种类型，那结构体也可做比较。其中指针比较特别，只支持相等操作。

使用[cmp](https://godoc.org/github.com/google/go-cmp/cmp)包的[cmp.Equal](https://godoc.org/github.com/google/go-cmp/cmp#Equal)可直接比较两个任意的对象，使用[cmp.Diff](https://godoc.org/github.com/google/go-cmp/cmp#Diff)则会输出这两个对象间的差异，而且可读性非常高。

虽然`cmp`不在标准库中，但它是由 Go 官方维护的，和每一版的 Go 兼容，适用于大部分对象间的比较需求。

老旧的代码中会使用`reflect.DeepEqual`函数来比较复杂的结构体，现在建议用`cmp`包来代替，因为`reflect.DeepEqual`对一些未导出的字段和实现细节的变动非常敏感。

（`cmp`包使用时添加`cmp.Compare(proto.Equal)`选项即可直接用于 protocol buffer 消息的比较。）

## 不仅打印期望值，也要打印实际值

测试结果在打印期望结果之前，应该打印函数的实际结果。通常我们会将测试结果格式化为：`YourFunc(%v) = %v, want %v`。

对于 diff 的输出，期望结果和实际结果谁前谁后不明显，这时需要加入额外信息帮助理解。这两个什么顺序并不重要，重要的是整个工程应该具有一致性。

## 标识函数名

在大部分测试中，失败消息应该包含所在函数名，即使该消息显而易见来自测试函数。

优先使用：

```go
t.Errorf("YourFunc(%v) = %v, want %v", in, got, want)
```

而不是：

```go
t.Errorf("got %v, want %v", got, want)
```

## 标识输入

在大部分测试中，函数输入参数也应该包含在失败消息中。如果输入参数的相关属性不明显（比如，参数较大或晦涩难懂），你应该在测试名中描述本测试的内容，并且将描述信息放入错误消息中。

对于表格驱动型测试，不要将序号作为测试名的一部分。在测试用例失败后，没人希望回到表格中一个个数来找出失败来自哪个用例。

## 失败继续执行

即使测试用例遇到了失败，它也应尽可能地继续执行，以便能在一次运行中打印出所有失败检查点。这样，如果有人要依照测试结果修复代码时，不用一遍遍重复执行用例来找出下一个 bug。

从实际角度出发，优先使用`t.Error`而不是`t.Fatal`。当比较函数的多个输出时，对每一个分别使用`t.Error`。

`t.Fatal`适合在 setup 中使用，因为 setup 一旦失败，其余的步骤便没有再执行的必要。表格驱动的测试中，`t.Fatal`适合在所有子测试开始前使用。表格中的每一测试用例若遇到不可恢复的错误，如何处理要分具体情况：

* 如果你没有使用`t.Run`运行子测试，那应该使用`t.Error`并使用`conitnue`语句直接跳转到下一项用例。
* 如果你使用`t.Run`运行子测试，那`t.Fatal`只会中断当前用例，其余子测试会继续执行。

## 标记测试辅助函数

辅助函数常用于 setup 和 teardown 任务中，比如构造一个测试数据。

在辅助函数中调用[t.Helper](https://godoc.org/testing#T.Helper)后，如果辅助函数中某个判断出错，那在测试日志中的错误提示会忽略该辅助函数的调用栈，标记出错的行会焦点在测试用例中，而非在辅助函数中。有点绕，看个例子便一目了然。

例如未使用`t.Helper`之前：

```go
package main

import "testing"

func testHelper(t *testing.T) {
	t.Helper()
	t.Fatal()
}

func TestHelloWorld(t *testing.T) {
	testHelper(t)
}
```

出错信息为：

```bash
--- FAIL: TestHelloWorld (0.00s)
    main_test.go:6:
FAIL
FAIL	test.test	0.001s
FAIL
```

标记代码后：

```go
package main

import "testing"

func testHelper(t *testing.T) {
	t.Helper()
	t.Fatal()
}

func TestHelloWorld(t *testing.T) {
	testHelper(t)
}
```

出错信息为：

```bash
--- FAIL: TestHelloWorld (0.00s)
    main_test.go:11:
FAIL
FAIL	test.test	0.002s
FAIL
```

可以看到，显示出错的第几行不一样。显然若`testHelper`被多个测试用例调用，后者的测试日志更易排查。

## 打印 diff

如果函数返回的输出比较长，而出错的地方只是其中一小段，那很难一眼看出区别。这对调试不友好，建议直接输出期望和实际结果的 diff 值。

## 表格驱动测试 vs 多个测试函数

当多个测试用例有着相同的测试逻辑，只是输入数据不同时，就应该使用[表格驱动测试](https://github.com/golang/go/wiki/TableDrivenTests)方法。

而当每个测试用例需用不同的方法验证时，表格驱动就显得不合适，因为那样就不得不写一堆控制变量放入表格中，将原本的测试逻辑淹没其中，降低了用例的可读性和表格的可维护性。

实际测试两种方法需结合使用。比如可以写两个表格驱动测试方法，一个测试函数的正常返回结果，另一个测试不同错误消息。

## 测试错误语义

单元测试避免使用字符串比较或者是`reflect.DeepEqual`去检查函数的错误输出。错误消息若随着业务成长需要经常变动，你会不得不经常修改单元测试用例。

而依赖库中的错误消息则相对稳定，拿来做字符串比较是可接受的。

我们应该区分哪些是为了提高排查效率增添的错误消息，哪些只是用于内部编程，而多用`fmt.Errorf`恰恰会打破内部的稳定性，应尽量少用。

许多人并不关心他们的 API 返回具体什么错误消息，这种情况下，单元测试中只做错误非空判断就可以了。