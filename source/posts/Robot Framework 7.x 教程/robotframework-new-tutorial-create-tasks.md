title: Robot Framework 新版教程 - 创建任务
date: 2025-12-18 14:50:08
series: Robot Framework 7.x 教程

---

除了测试自动化，Robot Framework 还可用于其他自动化场景，例如机器人流程自动化（RPA）。虽然这一功能始终可行，但直到 Robot Framework 3.1 版本才正式扩展了对非测试类自动化任务的支持。在大多数情况下，创建任务的流程与创建测试几乎完全相同，唯一的实质区别在于术语体系。与测试用例类似，任务同样可以通过套件形式进行结构化组织。

## 任务语法

任务的创建基于已有的关键字，这与创建测试用例的方式完全相同，并且任务的语法总体上与测试用例的语法一致。主要的区别在于，任务是在"任务"部分而非"测试用例"部分创建的：

```markdown
*** Tasks ***
Process invoice
    Read information from PDF
    Validate information
    Submit information to backend system
    Validate information is visible in web UI
```

注意：在同一文件中同时包含测试和任务是错误的。

## 任务相关配置设置

任务部分可使用的配置设置与测试用例部分完全相同。在设置部分，可以使用 `Task Setup`、`Task Teardown`、`Task Template` 和 `Task Timeout`，以替代相应的测试变体（如 `Test Setup`、`Test Teardown` 等）。