title: Ginkgo Label 标签的使用教程
date: 2023-12-24 10:04:00
series: Ginkgo 使用笔记
summary: 如何更有效的组织用例？

---

## 前言

很多测试框架都提供了对测试用例分组的能力，比如 `pytest` 中的 `mark`，`Robot Framework` 中的 `tag`，`TestNG` 中的 `groups` 等等。

在 `Ginkgo v1` 中，并没有类似的功能，在七牛，我们不得不将**标签**嵌入在字符串标题中，配合 `--focus` 来过滤用例。但基于字符串的实现并不完美，很容易出现冲突、重复，更不能提供类型安全，导致在海量集测用例面前，`Ginkgo v1` 用起来总感觉力不从心。

到了 `Ginkgo v2`，官方终于给出了 `Label` 这个解决方案，下面，我们来看看它都有哪些使用方法。

P.S. 本文基于 `Ginkgo v2.13.2`。

## 基础使用方法

### 给用例添加标签

标签用 `Label` 装饰器定义，并且可以定义在任意一种节点之上，例如：

```go
Describe("上传", Label("integration", "storage"), func() {
    It("表单上传", Label("network", "slow", "library storage"), func() {
        // 最终获得的标签 [integration, storage, network, slow, library storage]
    })

    Context("分片上传", Label("network", "library storage"), func() {
        It("1 个分片", Label("slow") func() {
            // 最终获得的标签 [integration, storage, network, slow, library storage]
        })

        It("2 个分片", Label("quick", "storage") func() {
            // 最终获得的标签 [integration, storage, network, quick, library storage]
        })
    })

    DescribeTable("s3 协议", Label("quick"), func(count int) {
        
    },
        Entry("put 上传", Label("local"), 17), // 最终获得的标签 [integration, storage, quick, local]
        Entry("拷贝上传", 20), // 最终获得的标签 [integration, storage, quick]
    )
})
```

总结一下：
1. 标签本质是一个字符串。
2. 子节点会继承父节点定义的标签，即 It 会继承 `Context` 和 `Describe` 上的标签。
3. 标签会自动去重，子节点不用担心标签重复。
4. `Entry` 上也可以定义标签，不会被当作参数。

### 过滤

`Ginkgo` 已经有很多过滤用例的方法：首先是 `--focus-file` 和 `--skip-file`，可以根据文件名和行号来过滤用例；然后是 `--focus` 和 `--skip`，可以根据用例的标题来过滤用例。这两个都支持正则表达式。但用例分类并不一定会集中于特定的目录中，比如为测试环境和线上环境各自编写的用例，这些用例会分散在各个目录、各个文件、各种用例中。而且，正则表达式并不直观，如果基于标题写正则表达式，恐怕这个命令行会非常**糟糕**。

而标签真正的威力，就在于其更易用的过滤语法。

`Ginkgo` 使用 `--label-filter=QUERY` 可以传入基于标签的查询语句，其规则规则一目了然：

* `标签1 && 标签2` ，用例同时含有两个标签，即符合条件。
* `标签1 || 标签2`，用例只需拥有一个标签，即符合条件。
* `!标签1`，有标签1的用例不符合条件。
* `,` 的逻辑和 `||` 相同。
* `()` 可以用来组合表达式。例如 `标签1 && (标签2 || 标签3)`
* 标签是大小写敏感的。
* 标签前后的空白会自动去除。 

举例来说，我们正在测试一个云存储产品，我们有以下 4 个测试用例，分别是：

1. 用例1: product, local, cn-east-1, slow
2. 用例2: local, cn-east-1, ap-southeast-1
3. 用例3: local, ap-southeast-1
4. 用例4: product, slow

其中

1. product 代表用例能在线上跑，local 代表用例能在测试环境跑。
2. cn-east-1 代表用例能在华东区域跑，ap-southeast-1 代表用例能在东南亚区域跑。
3. slow 代表用例运行时间较长。

如果使用以下过滤语句：

1. `product`，即挑选所有能在线上跑的用例，那用例1和用例4会执行。
2. `!local`，即挑选所有不能在测试环境跑的用例，那只有用例4会执行。
3. `product && cn-east-1`，即挑选线上华东区域的所有用例，那只有用例1会执行。
4. `cn-east-1 || ap-south-east-1`，即挑选能同时在华东和东南亚区域运行的用例，那用例1、2、3会执行。
5. `!slow`，即排除时间过长的用例，那用例2、3会执行。

可以发现，新的过滤语法更为直观，使用者能快速组合出想要的用例。（如果上述过滤语法仍不满足要求，可以用 `/REGEXP/` 来使用正则表达式）

### 组合使用

`Ginkgo v2` 并没有删去 `Ginkgo v1` 已有的过滤功能，而是可以组合使用：

1. 如果用例被标记为 `Pending`，那无论如何都不会运行。
2. 如果用例中调用了 `Skip()` 函数，即使命中了过滤语句，仍然会被忽略。
3. 如果用例被标记为 `Focus`，那只会运行该用例。
4. 如果命令行同时有 `--label-filter`, `--focus-file/--skip-file`, `--focus/--skip`，那最终用例必须同时符合这些条件。

### 测试报告

标签在官方自带的 `JUnit Report` 报告中，不是一个单独的属性，而是附在标题中：

```
Kodo e2e Suite.[It] 测试 s3 分片上传 [module=bucket, KODO-18044, unstable, id=c522c, id=be25c]
```

后面 `[xxx, yyy, zzz]` 便是该用例所有的标签。

## 高级使用

### 组合标签

`Ginkgo` 提供了 `Label()` 函数来定义标签类型 `Labels`，它们的关系如下：

```go
func Label(labels ...string) Labels {
	return Labels(labels)
}

type Labels = internal.Labels
```

默认的 `Label()` 函数在面对多标签时并不灵活，举个例子，假设我们测试云存储产品，它有 4 个存储区域：

1. `cn-east-1` 华东
2. `cn-north-1` 华北
3. `cn-northwest-1` 西北
4. `ap-southeast-2` 东南亚

我们可能会定义以下四个标签来标注用例是否可以在对应区域运行：

```go
var ZCnEast1 = Label("cn-east-1")

var ZCnNorth1 = Label("cn-north-1")

var ZCnNorthWest1 = Label("cn-northwest-1")

var ZApSouthEast2 = Label("ap-southeast-2")
```

但若每个用例都要标注**4个**标签，写起来比较繁琐，你可能会定义一个**全区域**标签来表示该用例可以在任意区域运行：

```go
var ZAll = Label("cn-east-1", "cn-north-1", "cn-northwest-1", "ap-southeast-2")
```

直接使用字符串会有**类型安全**问题，所以可以定义一个辅助函数来组合已有标签。

因为标签本质是 `Labels` 类型，只要定义一个 `combine([]Labels) Labels` 的函数即可：

```go
func combine(labels ...Labels) Labels {
	mapl := make(map[string]bool)

	for _, ls := range labels {
		for _, l := range ls {
			mapl[l] = true
		}
	}

	out := make([]string, 0)
	for k := range mapl {
		out = append(out, k)
	}

	return Label(out...)
}

// 全球运行
var ZAll = combine(ZCnEast1, ZCnNorth1, ZCnNorthWest1, ZApSouthEast2)
```

同理，如果一个用例只是在 `ap-southeast-2` 东南亚无法运行，可以定义一个 `remove(Labels, ...Labels) Labels` 的函数：

```go
func remove(all Labels, remove ...Labels) Labels {
	labelNeedsDel := make(map[string]bool)

	for _, labels2 := range remove {
		for _, l := range labels2 {
			labelNeedsDel[l] = true
		}
	}

	out := make([]string, 0)
	for _, l := range all {
		if _, ok := labelNeedsDel[l]; !ok {
			out = append(out, l)
		}
	}

	return Label(out...)
}

// 只能在国内运行
var ZChina = skip(ZAll, ZApSouthEast2)
```

### 自动过滤

#### 根据配置文件过滤

配置文件天然含有过滤信息，例如线上环境配置、灰度环境配置、本地环境配置。那 `Ginkgo` 能不能自动添加过滤语句呢？当然可以。

回到 `Ginkgo` 的启动函数：

```go
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "e2e Suite")
}
```

`RunSpecs` 可以接受额外的参数：`SuiteConfig` 和 `ReporterConfig`，其中 `SuiteConfig` 可以定制 `Ginkgo CLI` 启动参数：

```go
var EnvProduct = Label("product")

suiteConfig, reportConfig := GinkgoConfiguration()
// 与已有的逻辑是与的关系
suiteConfig.LabelFilter = fmt.Sprintf("(%v) && (%v)", suiteConfig.LabelFilter, "product")

RunSpecs(t, "e2e Suite", suiteConfig, reportConfig)
```

这样，`product` 标签就被添加到了 `--label-filter` 中，和已有的过滤逻辑是**与**的关系。

#### 手动用例

有可能存在这样一些用例，它们不能完全自动化，验收的时候需要手动执行，人工验证。这些用例在 `Ginkgo v1` 中，只能被标注为 `Skip`，每次执行的时候需要将 `Skip` 装饰器去掉再运行，十分麻烦。

而借助灵活的标签过滤语法，就能在不修改集测代码的情况下运行它，例如：

```go
var Manual = Label("manual")

// 如果已经有 manual 就不再加 !manual 标签
if !strings.Contains(suiteConfig.LabelFilter, "manual") {
    suiteConfig.LabelFilter += " && (!manual)" // 强制排除 manual 标签
}
```

1. 平时执行集测回归时，会自动添加 `!manual`，忽略所有的手动用例。
2. 而一旦命令行添加 `manual` 后，便会过滤出对应的手动用例。
