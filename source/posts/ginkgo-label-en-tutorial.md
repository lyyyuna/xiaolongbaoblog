title: How to use Ginkgo label in your test
date: 2025-04-07 10:04:00
series: Ginkgo 使用笔记
summary: A Ginkgo label tutorial

---

## Abstract

Many testing frameworks provide capabilities to group test cases, such as `pytest`'s `mark`, `Robot Framework`'s `tag`, and `TestNG`'s `groups`.

In `Ginkgo v1`, there was no such functionality, so we had to embed `labels` into string-based titles and use `--focus` to filter test cases. However, string-based implementations have drawbacks like potential conflicts, duplication, and lack of type safety, `Ginkgo v1` is inadequate for large-scale test suites.

In `Ginkgo v2`, the official solution `Label` is introduced, let's explore it together.

P.S. This article is based on `Ginkgo v2.13.2`.

## Basic Usage

### Adding Labels to Test Cases

Labels are defined using the `Label` decorator and can be applied to any node type:

```go
Describe("Upload", Label("integration", "storage"), func() {
    It("Form upload", Label("network", "slow", "library storage"), func() {
        // Final labels [integration, storage, network, slow, library storage]
    })

    Context("Chunked upload", Label("network", "library storage"), func() {
        It("1 chunk", Label("slow") func() {
            // Final labels [integration, storage, network, slow, library storage]
        })

        It("2 chunks", Label("quick", "storage") func() {
            // Final labels [integration, storage, network, quick, library storage]
        })
    })

    DescribeTable("s3 protocol", Label("quick"), func(count int) {
        
    },
        Entry("PUT upload", Label("local"), 17), // Final labels [integration, storage, quick, local]
        Entry("COPY upload", 20), // Final labels [integration, storage, quick]
    )
})
```

Key points:

1. Labels are just strings.
2. Child nodes inherit parent node labels (e.g., `It` inherits from `Context/Describe`).
3. Labels are automatically deduplicated.
4. `Entry` can also have labels, they will not be treated as parameters.

### Filtering

`Ginkgo v1` already supports filtering via `--focus-file`, `--skip-file` (by filename/line number) and `--focus`, `--skip` (by title regex). However, test cases written for testing environments versus production environments may be scattered across various directories, files, and different test structures. Additionally, regular expressions are not intuitive and can lead to cumbersome command-line syntax.

 The new label-based filtering provides a more intuitive syntax, `Ginkgo v2` use `--label-filter=QUERY` with these rules:

* `label1 && label2`, both labels match
* `label1 || label2`, either label matches
* `!label1` excludes cases with this label
* `,` acts like `||`
* `()` can group expressions (e.g., `label1 && (label2 || label3)`)
* case-sensitive
* whitespace around labels is ignored.

For instance, if we are testing a cloud storage product, we have 4 test cases:

1. test case 1: product, local, cn-east-1, slow
2. test case 2: local, cn-east-1, ap-southeast-1
3. test case 3: local, ap-southeast-1
4. test case 4: product, slow

where

1. product indicates the test case can run in the production environment, while local indicates it can run in the testing environment.
2. cn-east-1 means the test case can run in the East China region, and ap-southeast-1 indicates it can run in the Southeast Asia region.
3. slow signifies the test case takes a long time to execute.

If using the following filter expressions:

1. `product`: Selects all test cases that can run in the production environment. This would execute Case 1 and Case 4.
2. `!local`: Selects test cases that cannot run in the testing environment. This would only execute Case 4.
3. `product && cn-east-1`: Selects test cases that run in the production environment in the East China region. This would execute only Case 1.
4. `cn-east-1 || ap-southeast-1`: Selects test cases that can run in either the East China or Southeast Asia regions. This would execute Cases 1, 2, and 3.
5. `!slow`: Excludes test cases that take a long time to execute. This would execute Cases 2 and 3.

It becomes evident that the new filtering syntax is more intuitive, allowing users to quickly construct the desired test case combinations.

### Combined Usage

`Ginkgo v2` retains previous filtering mechanisms:

1. If a test case is marked as `Pending`, it will never execute, regardless of other conditions.
2. If a test case calls the `Skip()` function, it will be skipped even if it matches the filter criteria.
3. If a test case is marked with `Focus`, only that test case will run (all others are excluded).
4. If multiple filters are used in the command line (`--label-filter`, `--focus-file/--skip-file`, or `--focus/--skip`), a test case must satisfy all conditions simultaneously to execute.

### Test Reports

Labels in Ginkgo's built-in `JUnit Report` are not standalone attributes but are appended to the test case title. For example:


```
Kodo e2e Suite.[It] Test s3 Chunked upload [module=bucket, KODO-18044, unstable, id=c522c, id=be25c]
```

The section [xxx, yyy, zzz] following the title lists all labels associated with the test case.

## Advanced Usage

### Composing Labels

`Ginkgo` provides the `Label()` function to define the `Labels` type, with the following relationship:

```go
func Label(labels ...string) Labels {
	return Labels(labels)
}

type Labels = internal.Labels
```

While the default `Label()` function works for basic use cases, it lacks flexibility for complex scenarios. For example, when testing a cloud storage product with four regions:

1. cn-east-1 (East China)
2. cn-north-1 (North China)
3. cn-northwest-1 (Northwest China)
4. ap-southeast-2 (Southeast Asia)

You might define labels to indicate which regions a test case can run in:

```go
var ZCnEast1 = Label("cn-east-1")

var ZCnNorth1 = Label("cn-north-1")

var ZCnNorthWest1 = Label("cn-northwest-1")

var ZApSouthEast2 = Label("ap-southeast-2")
```

However, manually adding all 4 labels to each test case can become cumbersome. To address this, you can define a composite label like ZAll to represent **it can be run in all regions**:

```go
var ZAll = Label("cn-east-1", "cn-north-1", "cn-northwest-1", "ap-southeast-2")
```

To avoid string-based errors and enhance flexibility, you can create helper functions to combine or modify labels:

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

// all region
var ZAll = combine(ZCnEast1, ZCnNorth1, ZCnNorthWest1, ZApSouthEast2)
```

Similarly, if a test case cannot run in the `ap-southeast-2` (Southeast Asia region), you can define a `remove(Labels, ...Labels) Labels` function to exclude specific labels:


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

// can only run in China
var ZChina = remove(ZAll, ZApSouthEast2)
```

### Automatic Filtering

#### Configuration File-Based Filtering

Many applications use configuration files to define environment-specific settings (e.g., production, staging, or local environments). You can automatically apply label filters based on these configurations.

Modify the Ginkgo entry function to inject labels dynamically:

```go
func TestE2E(t *testing.T) {  
    RegisterFailHandler(Fail)  

    // Retrieve current configuration  
    suiteConfig, reportConfig := GinkgoConfiguration()  

    // Add "product" label filter (combined with existing filters via logical AND)  
    suiteConfig.LabelFilter = fmt.Sprintf("(%v) && (%v)",  
        suiteConfig.LabelFilter,  
        "product",  
    )  

    // Run tests with updated configuration  
    RunSpecs(t, "e2e Suite", suiteConfig, reportConfig)  
}  
```

Explanation:

1. The `product` label is automatically added to `--label-filter`, ensuring only production-ready test cases run by default.
2. Existing command-line filters (e.g., `--label-filter=region=us-east-1`) are preserved via the logical `AND` relationship.

#### Manually Filering

Some test cases cannot be fully automated and require manual execution (e.g., UI validation or human-in-the-loop steps). In `Ginkgo v1`, these cases had to be manually commented/uncommented by `Skip()`. With labels, you can define a `manual` label, and exclude it in Ginkgo entry point:

```go
var Manual = Label("manual")  

func TestE2E(t *testing.T) {  
    RegisterFailHandler(Fail)  

    suiteConfig, reportConfig := GinkgoConfiguration()  

    // Automatically exclude manual tests unless explicitly included  
    if !strings.Contains(suiteConfig.LabelFilter, "manual") {  
        suiteConfig.LabelFilter += " && (!manual)"  
    }  

    RunSpecs(t, "e2e Suite", suiteConfig, reportConfig)  
}  
```

Behavior:

1. **Default**: All tests with the `manual` label are skipped (due to `!manual`).
2. **Override**: Run manual tests by adding `--label-filter=manual` to the command line.
