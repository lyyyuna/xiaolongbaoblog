title: Robot Framework 新版教程 - 创建测试用例
date: 2025-07-23 14:50:08
series: Robot Framework 7.x 教程

---

本节介绍测试用例的总体语法规则。关于如何通过套件文件及套件目录组织测试用例构成测试套件，将在下一文章中详细讨论。

## 测试用例语法

### 基本语法

测试用例在 test case section 中通过可用关键字构建。关键字可从测试库或资源文件导入，或在测试用例文件自身的 keyword section 中创建。

test case section 的第一列包含测试用例名称。测试用例从该列有内容的行开始，延续至下一个测试用例名称或 section 结尾。若在 section 标题与第一个测试用例之间存在内容，则视为错误。

第二列通常为关键字名称。例外情况是通过关键字返回值设置变量时，第二列（及可能的后续列）包含变量名，而关键字名位于其后。无论哪种情况，关键字名之后的列都包含该关键字的参数。

```markdown
*** Test Cases ***
Valid Login
    Open Login Page
    Input Username    demo
    Input Password    mode
    Submit Credentials
    Welcome Page Should Be Open

Setting Variables
    Do Something    first argument    second argument
    ${value} =    Get Some Value
    Should Be Equal    ${value}    Expected value
```


### 测试用例中的设置部分

测试用例也可以包含专属设置项。设置名称始终位于第二列（通常放置关键字的位置），其参数值则位于后续列。设置名称需用方括号标识，以区别于普通关键字。可用的设置项如下所示，本节后续将详细说明：

* `[Documentation]`，用于指定测试用例的说明文档
* `[Setup]`, `[Teardown]`，指定测试的初始化和清理操作
* `[Tags]`，用于为测试用例添加标签
* `[Template]`，指定模板关键字。测试用例本身仅包含作为该关键字参数使用的数据
* `[Timeout]`，用于设置测试用例超时时间。超时设置将在专门章节详细讨论

示例：

```markdown
*** Test Cases ***
Test With Settings
    [Documentation]    Another dummy test
    [Tags]    dummy    owner-johndoe
    Log    Hello, world!
```

### 与测试用例相关的配置项

Setting section 可包含以下与测试用例相关的设置项。这些设置项主要为前文列出的测试用例专属设置提供默认值：

* `Test Setup`, `Test Teardown`，测试初始化和清理操作的默认设置
* `Test Tags`，测试套件中所有测试用例都将继承的默认标签（会与各测试用例的自定义标签合并）
* `Test Template`，默认使用的模板关键字
* `Test Timeout`，测试用例超时的默认值（超时设置将在专门章节详细讨论）

## 使用参数