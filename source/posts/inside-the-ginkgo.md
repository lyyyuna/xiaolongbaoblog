title: Ginkgo 测试框架实现解析
date: 2022-05-12 12:19:02
summary: 你是否好奇 Ginkgo 是如何实现 BDD 风格的测试框架呢？
series: Ginkgo 使用笔记

---

## 前言

先看一段典型的 Ginkgo 测试代码：

```go
var _ = Describe("Book", func() {
    Context("my", func() {
        It("1", func() {
        })

        It("2", func() {
        })

        It("3", func() {
        })
    })
})
```

习惯了 pytest, Junit, TestNG, go 自带单元测试等等测试框架后，会觉得 Ginkgo 的测试用例 `It` 怎么这么奇怪？是嵌套在 `func(){}` 中的函数？难道上面的三个用例是顺序执行的三个函数调用？

类似这样的疑问还有很多，本文会选择一些我感兴趣的问题，并试图解答：

1. Ginkgo 的测试用例是如何组织起来的？`Context`/`Describe` 中的其他语句如何与 `It` 区分开来？
2. Ginkgo 写的测试用例为什么能可以编译成二进制分发？为什么也能用 `go test` 执行？
3. Ginkgo 的并发是多进程还是多个 goroutine？

本文基于 [Ginkgo v2.1.4](https://github.com/onsi/ginkgo/tree/v2.1.4) 源码进行分析。

## 测试用例

### 测试入口

Ginkgo 测试的入口一般长这样：

```go
func TestBooks(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Books Suite")
}
```

该函数 `TestBooks` 其实是 go test 标准库框架的下的一个测试用例，这也是为什么 `go test` 也能识别 Ginkgo 测试用例的原因。Ginkgo 的测试用例和 `go test` 的测试用例并不等价，所有 Ginkgo 代码只相当于 `go test` 中的一个用例。既然大框架上属于 `go test`，那么 Ginkgo 也要遵循标准库测试的规则： 测试入口文件必须是 `_test.go` 结尾，否则 Ginkgo 不能识别。

### 用例容器

在 Ginkgo 中，如本文开头的例子，测试用例 `It` 不能单独存在，必须从属于 `Describe`/`Context` 这样的容器。这两类容器类型分别定义为：

```go
NodeTypeContainer NodeType = 1 << iota
NodeTypeIt
```

这些用例的容器都有着类似的实现：

```go
func Describe(text string, args ...interface{}) bool {
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeContainer, text, args...))
}

func NewNode(deprecationTracker *types.DeprecationTracker, nodeType types.NodeType, text string, args ...interface{}) (Node, []error) {
...
	for _, arg := range args {
		switch t := reflect.TypeOf(arg); {
...
		case t.Kind() == reflect.Func:
...
			node.Body = arg.(func())
...
        }
    }
}
```

当我们用 `var _ = Describe()` 匿名变量定义顶层容器后， `Describe` 会创建一个节点，并将用例所在的 `func(){ xxx }` 赋值给节点的 `node.Body` 成员。

而这个匿名变量是顶层的全局变量，根据 Go 语言的内存模型规则，必须在 `main` 函数或者 `TestBooks` 测试函数运行之前初始化完毕。这也就意味着运行 `RunSpecs()` 之前，Ginkgo 便通过 `Describe` 收集到当前测试套件内所有的顶层容器。

接着，go 测试函数运行 `RunSpecs()`：

```go
func RunSpecs(t GinkgoTestingT, description string, args ...interface{}) bool {
...
	err := global.Suite.BuildTree()
...
}

func (suite *Suite) BuildTree() error {
	suite.phase = PhaseBuildTree
	for _, topLevelContainer := range suite.topLevelContainers {
		err := suite.PushNode(topLevelContainer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (suite *Suite) PushNode(node Node) error {
...
	if node.NodeType == types.NodeTypeContainer {
...
		if suite.phase == PhaseBuildTree {
			parentTree := suite.tree
			suite.tree = &TreeNode{Node: node}
			parentTree.AppendChild(suite.tree)
			
            node.Body()

			suite.tree = parentTree
			return err
		}
	} else {
		suite.tree.AppendChild(&TreeNode{Node: node})
		return nil
	}
...
}
```

以上是关键代码，`PushNode()` 函数在匿名变量初始化、用例解析阶段多次调用，递归调用后逻辑揉在一起就有点绕，这里总结如下：

1. 在 `PhaseBuildTree` 阶段，将所有顶层容器统一在一棵多叉树下。
2. 接着运行 `node.Body()`，即包含了各种子容器 `Context`。
3. 子容器的定义又会调用 `PushNode() --> node.Body()`。
4. 2-3 步骤会循环调用，将所有容器节点加入步骤 1 创建的多叉树内。
5. 如果当前节点是 `It`，只会将节点加入树的叶子节点，并不会运行 `node.Body()`。

最终树如下：

```
        Describe
       /       \
  Context    Context
  /    \        |
It     It       It
```

这里要注意，包含用例逻辑的 `It` 并没有被执行，只有各种父级容器运行了 `func(){}`。

### 执行用例

当树构造完毕后，`RunSpecs` 便会开始遍历树：

```go
func RunSpecs(t GinkgoTestingT, description string, args ...interface{}) bool {
...
	err := global.Suite.BuildTree()
...
	passed, hasFocusedTests := global.Suite.Run(...)
...
}

func (suite *Suite) Run(...) (bool, bool) {
...
	success := suite.runSpecs(description, suiteLabels, suitePath, hasProgrammaticFocus, specs)
...
}

func (suite *Suite) runSpecs(description string, suiteLabels Labels, suitePath string, hasProgrammaticFocus bool, specs Specs) bool {
...
	newGroup(suite).run(specs.AtIndices(groupedSpecIndices[groupedSpecIdx]))
...
}

func (g *group) run(specs Specs) {
	g.specs = specs

	for _, spec := range g.specs {
        ...
        g.attemptSpec(attempt == maxAttempts-1, spec)
        ...
    }
    ...
}

func (g *group) attemptSpec(isFinalAttempt bool, spec Spec) {
...
	for _, node := range nodes {
...
		g.suite.currentSpecReport.State, g.suite.currentSpecReport.Failure = g.suite.runNode(node, interruptStatus.Channel, spec.Nodes.BestTextFor(node))
...
	}
...
}

func (suite *Suite) runNode(node Node, interruptChannel chan interface{}, text string) (types.SpecState, types.Failure) {
...
	go func() {
		node.Body()
	}()
...
}
```

而叶子节点 `It` 用例本身的内容，只有到了 `runNode()` 阶段才会真正运行。

回答本文的第一个问题：

1. 全局匿名函数 `var _ = Describe("", func(){})` 定义了顶层的测试用例容器。
2. `RunSpecs` 会运行容器 `func(){}` 内的语句，将所有用例统一在一棵树上。
3. 接着 `RunSpecs` 会运行 `It` 内的语句，执行真正的测试用例。

熟悉了用例的解析原理后，你是否能回答该示例中 0，1，2，3 分别是什么时机打印？分别打印几次？

```go
var _ = Describe("Book", func() {
    Context("my", func() {
        fmt.Println(0)

        It("1", func() {
            fmt.Println(1)
        })

        It("2", func() {
            fmt.Println(2)
        })

        It("3", func() {
            fmt.Println(3)
        })
    })
})
```

## ginkgo run

### 编译

虽然 Ginkgo 可以直接编译成二进制，但平时用的最多的还是 `ginkgo run` 命令来运行用例。

`ginkgo run` 的入口在 `ginkgo/run/run_command.go`：

```go
func BuildRunCommand() command.Command {
...
	return command.Command{
...
		Command: func(args []string, additionalArgs []string) {
...
			runner.RunSpecs(args, additionalArgs)
		},
	}
}

func (r *SpecRunner) RunSpecs(args []string, additionalArgs []string) {
...
	opc := internal.NewOrderedParallelCompiler(r.cliConfig.ComputedNumCompilers())
	opc.StartCompiling(suites, r.goFlagsConfig)
...
}
```

从 `opc.StartCompiling(suites, r.goFlagsConfig)` 名字来看，run 命令会先对测试用例编译，是 Ginkgo 自己把编译器的活也做了？我们接着看 `StartCompiling` 的实现：

```go
func (opc *OrderedParallelCompiler) StartCompiling(suites TestSuites, goFlagsConfig types.GoFlagsConfig) {
...
	for compiler := 0; compiler < opc.numCompilers; compiler++ {
		go func() {
...
			suite = CompileSuite(suite, goFlagsConfig)
...
		}()
	}
...
}

func CompileSuite(suite TestSuite, goFlagsConfig types.GoFlagsConfig) TestSuite {
...
	path, err := filepath.Abs(filepath.Join(suite.Path, suite.PackageName+".test"))
...
	args, err := types.GenerateGoTestCompileArgs(goFlagsConfig, path, "./")
...
	cmd := exec.Command("go", args...)
	cmd.Dir = suite.Path
	output, err := cmd.CombinedOutput()
...
}

func GenerateGoTestCompileArgs(...) ([]string, error) {
...
	args := []string{"test", "-c", "-o", destination, packageToBuild}
...
}
```

看来 Ginkgo 没有另起炉灶，正是用官方的 `go test -c` 来编译成二进制的，这也正好回答了本文的第二个疑问。

### 运行用例

```go
func (r *SpecRunner) RunSpecs(args []string, additionalArgs []string) {
...
	suites[suiteIdx] = internal.RunCompiledSuite(...)
...
}

func RunCompiledSuite(...) TestSuite {
...
	suite = runSerial(...)
...
}

func runSerial(...) TestSuite {
...
	cmd, buf := buildAndStartCommand(suite, args, true)
...
}

func buildAndStartCommand(suite TestSuite, args []string, pipeToStdout bool) (*exec.Cmd, *bytes.Buffer) {
...
	cmd := exec.Command(suite.PathToCompiledTest, args...)
	cmd.Dir = suite.Path
...
}
```

和预想的一致，上一小节编译得到的 `xxx.test` 二进制在运行环节会直接调用。

这里要注意 `cmd.Dir = suite.Path` 这一行，说明测试用例的**当前目录**，并不是运行 `ginkgo run` 所在的目录，而是测试用例所在的目录，我们有时候会在用例中读写文件，明白 Ginkgo **当前目录**的规则，才能让你无论在哪都能正常运行。

### 清理

既然每次每次 run 都会编译，那为什么我在本地从来没看到过呢？那必然是有一套清理机制：

```go
func (r *SpecRunner) RunSpecs(args []string, additionalArgs []string) {
...
	internal.Cleanup(r.goFlagsConfig, suites...)
...
}

func Cleanup(goFlagsConfig types.GoFlagsConfig, suites ...TestSuite) {
...
	for _, suite := range suites {
		if !suite.Precompiled {
			os.Remove(suite.PathToCompiledTest)
		}
	}
}
```

## 并发模型

### 并发入口

标准库的 `go test` 采用多协程来并发执行多个用例。Ginkgo 虽然在 `go test` 的大框架下，但本身只作为其中的**一个**用例，无法在原框架下并发自身的多个 Ginkgo 的用例。

Ginkgo 的[文档](https://onsi.github.io/ginkgo/#mental-model-how-ginkgo-runs-parallel-specs)中提到：

> Ginkgo ensures specs running in parallel are fully isolated from one another. It does this by running the specs in different processes.

其并发是多进程模型，由于多进程的特点，不同进程间的用例是隔离的。之前我们了解到，Ginkgo 会将用例编译成可执行二进制再运行，那派生出来的多个子进程又是什么？我们看一下运行用例函数 `RunCompiledSuite` 在并发场景下是什么流程：

```go
func RunCompiledSuite(...) TestSuite {
...
	if suite.IsGinkgo && cliConfig.ComputedProcs() > 1 {
		suite = runParallel(...)
	} else if suite.IsGinkgo {
		suite = runSerial(...)
	} 
...
}

func runParallel(...) TestSuite {
...
	server, err := parallel_support.NewServer(numProcs, reporters.NewDefaultReporter(reporterConfig, formatter.ColorableStdOut))
...
	for proc := 1; proc <= numProcs; proc++ {
...
		args, err := types.GenerateGinkgoTestRunArgs(procGinkgoConfig, reporterConfig, procGoFlagsConfig)
		cmd, buf := buildAndStartCommand(suite, args, false)
...
	}
...
}
```

在并发场景下：

1. 首先启动一个服务器监听随机端口，有可能是 RPC 服务器，也有可能是 HTTP 服务器，功能相同，只是协议不同。
2. 然后按并发数多次调用 `buildAndStartCommand`（上面分析过），即启动第一阶段编译好的二进制。
3. `Ginkgo run` CLI 承担了服务器功能，编译后的二进制完成了客户端的工作并执行用例。

那每个并发进程是如何与其他进程区别的呢？我们手动编译测试用例，并查看其 help 文档：

```bash
# ./test.test -h
Controlling Test Parallelism
These are set by the Ginkgo CLI, do not set them manually via go test.
Use ginkgo -p or ginkgo -procs=N instead.
  --ginkgo.parallel.process [int] (default: 1)
    This worker process's (one-indexed) process number.  For running specs in
    parallel.
  --ginkgo.parallel.total [int] (default: 1)
    The total number of worker processes.  For running specs in parallel.
  --ginkgo.parallel.host [string] (default: set by Ginkgo CLI)
    The address for the server that will synchronize the processes.
```

Ginkgo 会为二进制插入三个内置 flag，并且还特地强调了 **do not set them manually via go test**，仅内部使用，不要为它们赋值。它们分别有以下用途：

1. ginkgo.parallel.process，当前是第几个并发进程？相当于唯一 ID。
2. ginkgo.parallel.total，所有进程数。
3. ginkgo.parallel.host，服务器地址。

并发场景下启动的 `buildAndStartCommand` 会为每个并发进程设置好上面三个命令行参数，各自也就有了各自的 ID，并且知道如何与父进程通信（通过 HTTP 或 RPC 服务器）。

### 并发通信

服务器与并发客户端的通信接口如下：

```go
type Client interface {
	Connect() bool
	Close() error

	PostSuiteWillBegin(report types.Report) error
	PostDidRun(report types.SpecReport) error
	PostSuiteDidEnd(report types.Report) error
	PostSynchronizedBeforeSuiteCompleted(state types.SpecState, data []byte) error
	BlockUntilSynchronizedBeforeSuiteData() (types.SpecState, []byte, error)
	BlockUntilNonprimaryProcsHaveFinished() error
	BlockUntilAggregatedNonprimaryProcsReport() (types.Report, error)
	FetchNextCounter() (int, error)
	PostAbort() error
	ShouldAbort() bool
	Write(p []byte) (int, error)
}
```

其方法名一目了然，大概能猜测出中心服务器负责任务的派发和最终结果的收集，这里不再一一分析。

### 随机

Ginkgo 会打乱测试用例书写的顺序，以随机的方式来执行。在服务端代码（即 `Ginkgo run`）中，我们没有找到解析用例并随机打乱的流程，难道这部分工作也在客户端中？我们回到本文最开始的 `runSpecs` 中寻找答案：

```go
func (suite *Suite) runSpecs(description string, suiteLabels Labels, suitePath string, hasProgrammaticFocus bool, specs Specs) bool {
...
	groupedSpecIndices, serialGroupedSpecIndices := OrderSpecs(specs, suite.config)

	if suite.isRunningInParallel() {
		nextIndex = suite.client.FetchNextCounter
	}

	for {
		groupedSpecIdx, err := nextIndex()
		if err != nil {
			suite.report.SpecialSuiteFailureReasons = append(suite.report.SpecialSuiteFailureReasons, fmt.Sprintf("Failed to iterate over specs:\n%s", err.Error()))
			suite.report.SuiteSucceeded = false
			break
		}

		if groupedSpecIdx >= len(groupedSpecIndices) {
			if suite.config.ParallelProcess == 1 && len(serialGroupedSpecIndices) > 0 {
				groupedSpecIndices, serialGroupedSpecIndices, nextIndex = serialGroupedSpecIndices, GroupedSpecIndices{}, MakeIncrementingIndexCounter()
				suite.client.BlockUntilNonprimaryProcsHaveFinished()
				continue
			}
			break
		}

		newGroup(suite).run(specs.AtIndices(groupedSpecIndices[groupedSpecIdx]))
	}
...
}
```

可以看到：
1. 用例每次会通过 `FetchNextCounter` RPC/HTTP 方法获取下一个索引，索引由服务器保证全局唯一。
2. 每个并发进程根据索引选出要执行的用例，执行其 `node.Body()` 方法。
3. 索引可能代表一个用例，也可能代表一个用例容器内的一组用例，取决于配置。

确实如我们猜测的那样，随机打乱是在客户端做的，那每个客户端都做随机，不会重合么？

不用担心：

```go
import "math/rand"

func OrderSpecs(specs Specs, suiteConfig types.SuiteConfig) (GroupedSpecIndices, GroupedSpecIndices) {
...
	r := rand.New(rand.NewSource(suiteConfig.RandomSeed))
...
	permutation := r.Perm(len(shufflableGroupingIDs))
...
}
```

这里用的是 `math/rand` 伪随机库，只要种子 `suiteConfig.RandomSeed` 相同，各个并发客户端得到的随机序列是完全一样的。

## 结语

至此，本文开始提的三个问题已经获得了答案。有了本文的指引，读者可以接着去探索：

1. `BeforeEach`, `AfterEach` 是如何解析并运行的？
2. DescribeTable 和 Entry 是怎么实现的？为什么 Entry 内参数是在初始化阶段就执行完毕了？
3. report 是如何收集的？并发场景下呢？
4. 等等
