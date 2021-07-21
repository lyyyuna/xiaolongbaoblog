title: Golang 与子测试
date: 2020-12-14 16:20:02
categories: 测试
tags: 
- Go
---

## 前言

表格驱动测试可谓是最受欢迎的测试方法了，它抽取了相似用例的公共步骤，结构清晰，维护简单，比如：

```go
func TestOlder(t *testing.T) {
	cases := []struct {
		age1     int
		age2     int
		expected bool
	}{
        // 第一个测试用例
		{
			age1:     1,
			age2:     2,
			expected: false,
		},
        // 第二个测试用例
		{
			age1:     2,
			age2:     1,
			expected: true,
		},
	}

	for _, c := range cases {
		_, p1 := NewPerson(c.age1)
		_, p2 := NewPerson(c.age2)

		got := p1.older(p2)

		if got != c.expected {
			t.Errorf("Expected %v > %v, got %v", p1.age, p2.age, got)
        }
    } 
}
```

但是这种写法有着一个致命的缺陷，你无法像之前一样选择某个用例执行，即不支持 `go test -run regex` 命令行来选择只执行第一个或第二个测试用例。

`Go 1.7` 中加入了子测试的概念，以解决该问题。

## 什么是 Go 的子测试

子测试在 `testing` 包中由 [Run 方法](https://golang.org/pkg/testing/#T.Run) 提供，它有俩个参数：子测试的名字和子测试函数，其中名字是子测试的标识符。

子测试和其他普通的测试函数一样，是在独立的 goroutine 中运行，测试结果也会计入测试报告，所有子测试运行完毕后，父测试函数才会结束。

## 如何使用`t.Run`

使用`t.Run`重构前言中的测试代码，代码变动了不少：

```go
func TestOlder(t *testing.T) {
	cases := []struct {
		name     string
		age1     int
		age2     int
		expected bool
	}{
		{
			name:     "FirstOlderThanSecond",
			age1:     1,
			age2:     2,
			expected: false,
		},
		{
			name:     "SecondOlderThanFirst",
			age1:     2,
			age2:     1,
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, p1 := NewPerson(c.age1)
			_, p2 := NewPerson(c.age2)

			got := p1.older(p2)

			if got != c.expected {
				t.Errorf("Expected %v > %v, got %v", p1.age, p2.age, got)
			}
		})
	}

}
```

首先我们修改了定义用例的结构体，加入了`string`类型的`name`属性。这样每个用例都有了自己的名字来标示自己。例如，第一个用例由于参数`arg1`大于参数`arg2`，所以被命名称`FirstOlderThanSecond`。

然后在`for`循环中，我们把整个测试逻辑包裹在`t.Run`块中，并把用例名作为第一个参数。

运行该测试，可得：

```bash
$ go test -v -count=1
=== RUN   TestOlder
=== RUN   TestOlder/FirstOlderThanSecond
=== RUN   TestOlder/SecondOlderThanFirst
--- PASS: TestOlder (0.00s)
    --- PASS: TestOlder/FirstOlderThanSecond (0.00s)
    --- PASS: TestOlder/SecondOlderThanFirst (0.00s)
PASS
ok  	person	0.004s
```

从结果中我们发现，`TestOlder`派生出另外两个子测试函数：`TestOlder/FirstOlderThanSecond` 和 `TestOlder/SecondOlderThanFirst`。在这两个子测试结束之前，`TestOlder`都不会结束。

子测试函数的测试结果在终端里是缩进的，且测试用例的名字都以`TestOlder`开头，这些都用来凸显测试用例之间的父子关系。

## `go test`选择子测试运行

在调试特定测试用例或复现某个 bug 时我们常用`go test -run=regex`来指定。子测试`regex`的命名规则和上一节中测试结果一致：`父测试名/子测试名`。

比如可用以下命令执行子测试`FirstOlderThenSecond`：

```bash
$ go test -v -count=1 -run="TestOlder/FirstOlderThanSecond"
=== RUN   TestOlder
=== RUN   TestOlder/FirstOlderThanSecond
--- PASS: TestOlder (0.00s)
    --- PASS: TestOlder/FirstOlderThanSecond (0.00s)
PASS
```

如果要执行某个父测试下的所有子测试，可键入：

```bash
$ go test -v -count=1 -run="TestOlder"
=== RUN   TestOlder
=== RUN   TestOlder/FirstOlderThanSecond
=== RUN   TestOlder/SecondOlderThanFirst
--- PASS: TestOlder (0.00s)
    --- PASS: TestOlder/FirstOlderThanSecond (0.00s)
    --- PASS: TestOlder/SecondOlderThanFirst (0.00s)
PASS
```

## Setup 和 Teardown 和 TestMain

使用过其他测试框架的同学一定不会对`Setup`和`Teardown`陌生，这几乎是测试框架的标配。而 `testing` 包长期以来在这块是缺失的，我们无法为所有的测试用例添加一些公共的初始化和结束步骤。引入`t.Run`之后，我们便可以实现缺失的功能。

请看下面的例子，在子测试开始时，先调用`setupSubtest(t)`做初始化工作，然后使用`defer teardownSubtest(t)`保证在`t.Run`结束前执行清理工作。

```go
func setupSubtest(t *testing.T) {
	t.Logf("[SETUP] Hello 👋!")
}

func teardownSubtest(t *testing.T) {
	t.Logf("[TEARDOWN] Bye, bye 🖖!")
}

func TestOlder(t *testing.T) {
......
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
            // setup
            setupSubtest(t)
            // teardown
			defer teardownSubtest(t)

			_, p1 := NewPerson(c.age1)
			_, p2 := NewPerson(c.age2)

			got := p1.older(p2)

			t.Logf("[TEST] Hello from subtest %s \n", c.name)
			if got != c.expected {
				t.Errorf("Expected %v > %v, got %v", p1.age, p2.age, got)
			}
		})
	}
}
```

运行测试后，可以看到`Setup`和`Teardown`在每个子测试中都会被调用：

```bash
$ go test -v -count=1 -run="TestOlder"
=== RUN   TestOlder
=== RUN   TestOlder/FirstOlderThanSecond
=== RUN   TestOlder/SecondOlderThanFirst
--- PASS: TestOlder (0.00s)
    --- PASS: TestOlder/FirstOlderThanSecond (0.00s)
        person_test.go:33: [SETUP] Hello 👋!
        person_test.go:71: [TEST] Hello from subtest FirstOlderThanSecond
        person_test.go:37: [TEARDOWN] Bye, bye 🖖!
    --- PASS: TestOlder/SecondOlderThanFirst (0.00s)
        person_test.go:33: [SETUP] Hello 👋!
        person_test.go:71: [TEST] Hello from subtest SecondOlderThanFirst
        person_test.go:37: [TEARDOWN] Bye, bye 🖖!
PASS
ok  	person	0.005s
```

进一步的，每个包的测试文件其实都包含一个“隐藏”的`TestMain(m *testing.M)`函数：

```go
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
```

若重写该函数，在`m.Run`上下加入`Setup`和`Teardown`后便得到了全局的初始化和清理函数。

```go
func setupSubtest() {
	fmt.Println("[SETUP] Hello 👋!")
}

func teardownSubtest() {
	fmt.Println("[TEARDOWN] Bye, bye 🖖!")
}

func TestMain(m *testing.M) {
    setupSubtest()
    code := m.Run()
    teardownSubtest(t)
    os.Exit(code)
}
```