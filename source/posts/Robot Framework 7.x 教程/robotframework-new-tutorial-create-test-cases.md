title: Robot Framework 新版教程 - 创建测试用例
date: 2025-12-04 14:50:08
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

之前的示例已展示了关键字如何接收不同参数，本节将更深入地探讨这一重要功能。关于如何实际实现带不同参数的用户关键字和库关键字，将在后续独立章节中详细讨论。

关键字可接收零到多个参数，部分参数可能具有默认值。关键字具体支持哪些参数取决于其实现方式，通常最佳查询渠道是查阅关键字的说明文档。

### 位置参数

大多数关键字都有固定数量的必须提供的参数。在关键字文档中，这通过逗号分隔的参数名称表示，例如 `first`, `second`, `third`。此时参数名称的实际含义并不重要（除非需要解释参数作用），但关键是要确保参数数量与文档严格一致。参数过少或过多都会触发错误。

以下测试案例使用了 [OperatingSystem](https://robotframework.org/robotframework/latest/libraries/OperatingSystem.html) 库中的 `Create Directory` 和 `Copy File` 关键字。其参数分别标注为 `path` 与 `source`, `destination`，表示前者需 1 个参数，后者需 2 个参数。最后调用的 [BuiltIn](https://robotframework.org/robotframework/latest/libraries/BuiltIn.html) 库关键字 `No Operation` 则无需任何参数。

```markdown
*** Test Cases ***
Example
    Create Directory    ${TEMPDIR}/stuff
    Copy File    ${CURDIR}/file.txt    ${TEMPDIR}/stuff
    No Operation
```

### 带默认值的参数

参数通常具有可选默认值。在文档中，默认值通过等号与参数名称连接表示，格式为 `name=default value`。所有参数均可设置默认值，但注意：一旦某个参数设定了默认值，其后续的所有参数也必须具有默认值（即不能出现必选参数位于可选参数之后的情况）。

以下示例通过 `Create File` 关键字演示默认值的使用，该关键字参数定义为 `path`, `content=`, `encoding=UTF-8`。若调用时不提供任何参数或超过三个参数，则会导致错误。

```markdown
*** Test Cases ***
Example
    Create File    ${TEMPDIR}/empty.txt
    Create File    ${TEMPDIR}/utf-8.txt         Hyvä esimerkki
    Create File    ${TEMPDIR}/iso-8859-1.txt    Hyvä esimerkki    ISO-8859-1
```

### 可变数量的参数

关键字也可以接受任意数量的参数。这些所谓的可变参数可以与强制参数和带默认值的参数结合使用，但必须始终位于这些参数之后。在文档中，这类参数名称前会带有一个星号，例如 `*varargs`。

例如，[OperatingSystem](https://robotframework.org/robotframework/latest/libraries/OperatingSystem.html) 库中的 `Remove Files` 和 `Join Paths` 关键字分别具有参数 `*paths` 与 `base`、`*parts`。前者可以接受任意数量的参数，而后者至少需要一个参数。

```markdown
*** Test Cases ***
Example
    Remove Files    ${TEMPDIR}/f1.txt    ${TEMPDIR}/f2.txt    ${TEMPDIR}/f3.txt
    @{paths} =    Join Paths    ${TEMPDIR}    f1.txt    f2.txt    f3.txt    f4.txt
```

### 命名参数

命名参数语法使得使用带有默认值的参数更加灵活，并且能够明确标注特定参数值的含义。从技术上讲，命名参数的工作原理与 Python 中的关键字参数完全相同。

#### 基本语法

可以通过在参数值前添加"参数名="的方式为关键字指定参数，例如 `arg=value`。当多个参数具有默认值时，这种方式尤其有用，因为可以仅指定部分参数名称，而让其他参数继续使用默认值。例如，某个关键字接受参数 `arg1=a`, `arg2=b`, `arg3=c`，若调用时只传入一个参数 `arg3=override`，则 `arg1` 和 `arg2` 将保持默认值，而 `arg3` 将获得 `override` 值。如果这听起来有些复杂，希望下面关于命名参数的示例能帮助理解。

命名参数语法对大小写和空格都很敏感。前者意味着如果存在参数 `arg`，则必须使用 `arg=value` 的形式，使用 `Arg=value` 或 `ARG=value` 均无效。后者意味着等号前不允许有空格，而等号后的空格将被视为给定值的一部分。

在使用命名参数语法调用用户关键字时，参数名称必须省略 `${}` 修饰符。例如，具有参数 `${arg1}=first`, `${arg2}=second` 的用户关键字必须使用 `arg2=override` 的形式调用。

在命名参数后使用常规的位置参数（例如 `| Keyword | arg=value | positional |`）是无效的。而命名参数之间的相对顺序则无关紧要。

#### 带变量的命名参数

在命名参数的参数名和参数值中都可以使用变量。如果参数值是单个标量变量，该变量将以原样传递给关键字。这使得在使用命名参数语法时，不仅可以使用字符串，还可以使用任何对象作为值。例如，使用 `arg=${object}` 的形式调用关键字时，变量 `${object}` 将直接传递给关键字，而不会转换为字符串。

如果在命名参数的参数名中使用变量，变量会在与参数名匹配之前被解析。

命名参数语法要求等号必须直接写在关键字调用中。这意味着仅靠变量本身永远不会触发命名参数语法，即使其值形如 `foo=bar` 也不例外。这一点在将关键字封装到其他关键字中时尤其需要注意。例如，如果某个关键字接受可变数量的参数（如 `@{args}`），并使用相同的 `@{args}` 语法将所有参数传递给另一个关键字，那么调用端可能使用的 `named=arg` 语法将无法被识别。下面的示例说明了这种情况。

```markdown
*** Test Cases ***
Example
    Run Program    shell=True    # 此处的 shell=True 不会作为命名参数传递给 Run Process

*** Keywords ***
Run Program
    [Arguments]    @{args}
    Run Process    program.py    @{args}    # 无法从 @{args} 中识别出命名参数
```

#### 转义命名参数语法

仅当等号前的部分与关键字的某个参数名匹配时，才会触发命名参数语法。在某些情况下，可能存在一个值为字面量 `foo=quux` 的位置参数，同时另一个不相关的参数恰好名为 `foo`。此时，参数 `foo` 可能会错误地获得值 `quux`，或者更常见的是触发语法错误。

在这类极少出现的意外匹配场景中，可以使用反斜杠字符对语法进行转义，例如写作 `foo\=quux`。这样参数将直接获得字面值 `foo=quux`。需要注意的是，如果不存在名为 `foo` 的参数，则无需转义，但显式使用转义符能够使代码意图更明确，这通常是个值得提倡的做法。

#### 命名参数的适用场景

如前所述，命名参数语法适用于关键字调用。除此之外，该语法同样适用于库导入场景。

用户关键字和大多数测试库均支持命名参数，唯一例外的是显式使用仅限位置参数的 Python 关键字。

#### 命名参数示例

以下示例展示了如何在库关键字、用户关键字以及导入 Telnet 测试库时使用命名参数语法。

```markdown
*** Settings ***
Library    Telnet    prompt=$    default_log_level=DEBUG

*** Test Cases ***
Example
    Open connection    10.0.0.42    port=${PORT}    alias=example
    List files    options=-lh
    List files    path=/tmp    options=-l

*** Keywords ***
List files
    [Arguments]    ${path}=.    ${options}=
    Execute command    ls ${options} ${path}
```

### 自由命名参数

Robot Framework 支持自由命名参数（通常也称为自由关键字参数或 kwargs），其实现方式与 Python 中的 **kwargs 类似。这意味着关键字可以接收所有使用命名参数语法（`name=value`）且与关键字签名中定义的任何参数都不匹配的参数。

支持普通命名参数的关键字类型同样支持自由命名参数。不同关键字类型定义接收自由命名参数的方式有所不同：基于 Python 的关键字直接使用 `**kwargs`，而用户关键字则使用 `&{kwargs}`。

自由命名参数对变量的支持方式与命名参数类似。具体来说，变量既可用于参数名也可用于参数值，但转义符必须直接可见。例如，只要使用的变量存在，`foo=${bar}` 和 `${foo}=${bar}` 都是有效的。一个额外的限制是：自由参数名必须始终为字符串类型。

#### 示例

作为使用自由命名参数的首个示例，我们来看 [Process](https://robotframework.org/robotframework/latest/libraries/Process.html) 库中的 `Run Process` 关键字。其参数签名为 `command, *arguments, **configuration`，这意味着它接收要执行的命令（`command`）、可变数量的命令参数（`*arguments`），以及最终以自由命名参数形式传入的可选配置参数（`**configuration`）。以下示例同时展示了变量在自由关键字参数中的使用方式与命名参数语法完全一致。

```markdown
*** Test Cases ***
Free Named Arguments
    Run Process    program.py    arg1    arg2    cwd=/home/user
    Run Process    program.py    argument    shell=True    env=${ENVIRON}
```

更多关于在自定义测试库中使用自由命名参数语法的信息，请参阅《创建测试库》章节中的自由关键字参数（**kwargs）部分。

作为第二个示例，我们为上文中的 `program.py` 创建一个封装用户关键字。该封装关键字 `Run Program` 接收所有位置参数和命名参数，并将它们与要执行的命令名一起传递给 `Run Process`。

```markdown
*** Test Cases ***
Free Named Arguments
    Run Program    arg1    arg2    cwd=/home/user
    Run Program    argument    shell=True    env=${ENVIRON}

*** Keywords ***
Run Program
    [Arguments]    @{args}    &{config}
    Run Process    program.py    @{args}    &{config}
```

### 仅限命名参数

自 Robot Framework 3.1 起，关键字可以支持必须通过命名参数语法指定的参数。例如，若某关键字接受一个仅限命名参数 `example`，则调用时必须使用 `example=value` 的形式，直接使用 `value` 将无效。此语法设计灵感来源于 Python 3 中的仅限关键字参数。

在大多数情况下，仅限命名参数的工作机制与普通命名参数一致。主要区别在于：基于 Python 2 静态库 API 实现的测试库不支持此语法。

以下是通过用户关键字使用仅限命名参数的示例，这是对前文自由命名参数示例中 `Run Program` 关键字的改造版本，现仅支持配置 `shell` 参数：

```markdown
*** Test Cases ***
Named-only Arguments
    Run Program    arg1    arg2              # 'shell' 取默认值 False
    Run Program    argument    shell=True    # 'shell' 设置为 True

*** Keywords ***
Run Program
    [Arguments]    @{args}    ${shell}=False
    Run Process    program.py    @{args}    shell=${shell}
```

### 关键字名称中嵌入参数

另一种完全不同的参数指定方式是将参数直接嵌入到关键字名称中。该语法同时适用于测试库关键字和用户关键字。



## 处理失败

### 当测试用例失败时

如果测试用例使用的任何关键字失败，则该测试用例被视为失败。通常，这意味着该测试用例的执行会停止，并可能执行测试清理工作，然后从下一个测试用例继续执行。如果不希望停止测试执行，也可以使用特殊的可继续运行的失败类型。

### 错误信息

分配给失败测试用例的错误信息直接来自失败的关键字。通常，错误信息由关键字本身生成，但某些关键字允许配置错误信息。

在某些情况下，例如使用了可继续运行的失败类型时，一个测试用例可能会失败多次。此时，最终的错误信息将通过组合各个独立错误来生成。为便于阅读报告，过长的错误信息会自动从中间部分截断，但完整的错误信息始终可以在日志文件中作为失败关键字的记录进行查看。

默认情况下，错误信息为纯文本格式，但它们也可以包含 HTML 格式。这需要通过以标记字符串 `*HTML*` 开头来启用。此标记将在报告和日志中显示的最终错误信息里被移除。在自定义信息中使用 HTML 的示例如下。

```markdown
*** Test Cases ***
Normal Error
    Fail    This is a rather boring example...

HTML Error
    ${number} =    Get Number
    Should Be Equal    ${number}    42    *HTML* Number is not my <b>MAGIC</b> number.
```

## 测试用例名称与文档

测试用例的名称直接取自测试用例部分：即测试用例列中准确输入的内容。同一测试套件中的测试用例应具有唯一的名称。与此相关的是，您还可以在测试内部使用自动变量 `${TEST_NAME}` 来引用测试名称。该变量在测试执行期间（包括所有用户关键字、测试初始化及测试清理）均可用。

从 Robot Framework 3.2 开始，测试用例名称中可能存在的变量会被解析，从而使最终名称包含变量的值。如果变量不存在，其名称将保持不变。

```markdown

*** Variables ***
${MAX AMOUNT}      ${5000000}

*** Test Cases ***
Amount cannot be larger than ${MAX AMOUNT}
    # ...
```

`[Documentation]` 设置项允许为测试用例设置自由格式的文档。该文本将显示在命令行输出以及生成的日志和报告中。如果文档内容较长，可以拆分为多行。可以使用简单的 HTML 格式化，并且可以通过变量使文档内容动态变化。可能存在的不存在的变量将保持不变。

```markdown
*** Test Cases ***
Simple
    [Documentation]    Simple and short documentation.
    No Operation

Multiple lines
    [Documentation]    First row of the documentation.
    ...
    ...                Documentation continues here. These rows form
    ...                a paragraph when shown in HTML outputs.
    No Operation

Formatting
    [Documentation]
    ...    This list has:
    ...    - *bold*
    ...    - _italics_
    ...    - link: http://robotframework.org
    No Operation

Variables
    [Documentation]    Executed at ${HOST} by ${USER}
    No Operation
```

为测试用例设定清晰且具有描述性的名称至关重要，通常情况下，这样做就不再需要额外的文档。如果测试用例的逻辑需要文档说明，这通常意味着测试用例中的关键字需要更好的命名或进行增强，而不是添加额外的文档。最后，像上述最后一个示例中的环境和用户信息这类元数据，通常更适合使用标签来指定。

## 为测试用例添加标签

在 Robot Framework 中使用标签是一种简单而强大的机制，用于对测试用例及用户关键字进行分类。标签为自由文本，除了下面讨论的保留标签外，Robot Framework 本身对它们没有特殊含义。标签至少可用于以下目的：

* 它们显示在测试报告、日志中，当然也显示在测试数据中，因此它们为测试用例提供了元数据。
* 系统会根据标签自动收集测试用例的统计数据（总计、通过、失败和跳过）。
* 它们可用于包含、排除以及跳过测试用例。

为测试用例指定标签有多种方式，如下所述：

* **Settings 节中的 `[Test Tags]` 设置项**：具有此设置的测试用例文件中的所有测试始终获得指定的标签。如果在套件初始化文件中使用此设置，则所有子套件中的测试都会获得这些标签。
* **每个测试用例中的 `[Tags]` 设置项**：除了使用 `[Test Tags]` 设置指定的标签外，测试还会获得这些标签。`[Tags]` 设置还允许通过使用 `-tag` 语法来移除由 `[Test Tags]` 设置的标签。
* **--settag 命令行选项**：所有测试除了从其他地方获得的标签外，还会获得通过此选项设置的标签。
* **Set Tags、Remove Tags、Fail 和 Pass Execution 关键字**：这些 BuiltIn 库中的关键字可用于在测试执行期间动态地操作标签。

例子：

```markdown
*** Settings ***
Test Tags       requirement: 42    smoke

*** Variables ***
${HOST}         10.0.1.42

*** Test Cases ***
No own tags
    [Documentation]    Test has tags 'requirement: 42' and 'smoke'.
    No Operation

Own tags
    [Documentation]    Test has tags 'requirement: 42', 'smoke' and 'not ready'.
    [Tags]    not ready
    No Operation

Own tags with variable
    [Documentation]    Test has tags 'requirement: 42', 'smoke' and 'host: 10.0.1.42'.
    [Tags]    host: ${HOST}
    No Operation

Remove common tag
    [Documentation]    Test has only tag 'requirement: 42'.
    [Tags]    -smoke
    No Operation

Remove common tag using a pattern
    [Documentation]    Test has only tag 'smoke'.
    [Tags]    -requirement: *
    No Operation

Set Tags and Remove Tags keywords
    [Documentation]    This test has tags 'smoke', 'example' and 'another'.
    Set Tags    example    another
    Remove Tags    requirement: *
```

如示例所示，标签可以通过变量创建，但在其他情况下，它们会保留在数据中使用的确切名称。当标签进行比较时（例如，用于收集统计数据、选择要执行的测试或去除重复项），比较操作对大小写、空格和下划线不敏感。

如上文示例所示，使用 `-tag` 语法移除标签支持简单的模式，例如 `-requirement:*`。以连字符开头的标签，除非在 `[Tags]` 设置中使用，否则没有特殊含义。如果需要使用 `[Tags]` 设置一个以连字符开头的标签，可以使用转义格式，如 `\-tag`。

目前，`-tag` 语法仅可用于通过 `[Tags]` 设置项来移除标签，但计划在 Robot Framework 8.0 中，在 `Test Tags` 设置中也支持此功能（[#5250](https://github.com/robotframework/robotframework/issues/5250)）。在 Robot Framework 7.2 中，已弃用 `Test Tags` 设置中值为以连字符开头的字面量的标签（[#5252](https://github.com/robotframework/robotframework/issues/5252)）。如果需要此类值的标签，可以使用 `\-tag` 这样的转义格式。


### 弃用 Force Tags 和 Default Tags

在 Robot Framework 6.0 之前，可以通过 Setting 节中的两种不同设置来为测试指定标签：

* **Force Tags**：所有测试都无条件获得这些标签。这与现今的 `Test Tags` 完全相同。
* **Default Tags**：所有测试默认获得这些标签。如果测试已具有 `[Tags]` 设置，则不会获得这些标签。

这两种设置目前仍然有效，但已被视为弃用。未来（很可能在 Robot Framework 8.0 中）会添加可见的弃用警告，并最终移除这些设置。可以使用 Tidy 等工具来简化迁移过程。

更新 `Force Tags` 仅需将其重命名为 `Test Tags`。`Default Tags` 设置将被完全移除，但 Robot Framework 7.0 引入的 `-tag` 功能提供了相同的底层功能。以下示例展示了所需的更改。

旧语法：

```markdown
*** Settings ***
Force Tags      all
Default Tags    default

*** Test Cases ***
Common only
    [Documentation]    Test has tags 'all' and 'default'.
    No Operation

No default
    [Documentation]    Test has only tag 'all'.
    [Tags]
    No Operation

Own and no default
    [Documentation]    Test has tags 'all' and 'own'.
    [Tags]    own
    No Operation
```

新语法：

```markdown
*** Settings ***
Test Tags      all    default

*** Test Cases ***
Common only
    [Documentation]    Test has tags 'all' and 'default'.
    No Operation

No default
    [Documentation]    Test has only tag 'all'.
    [Tags]    -default
    No Operation

Own and no default
    [Documentation]    Test has tags 'all' and 'own'.
    [Tags]    own    -default
    No Operation
```

### 保留标签

用户通常可以自由使用在其上下文中有效的任何标签。然而，存在某些标签对 Robot Framework 本身具有预定义的含义，将它们用于其他目的可能会导致意外结果。Robot Framework 目前及将来拥有的所有特殊标签都带有 `robot:` 前缀。因此，为避免出现问题，用户除非确实需要激活特定功能，否则不应使用任何带有此前缀的标签。当前的保留标签列表如下，但未来可能会添加更多此类标签。

* `robot:continue-on-failure` 和 `robot:recursive-continue-on-failure`，用于启用"失败后继续运行"模式。
* `robot:stop-on-failure` 和 `robot:recursive-stop-on-failure`，用于禁用"失败后继续运行"模式。
* `robot:exit-on-failure`，如果带有此标签的测试失败，则停止整个执行过程。
* `robot:skip-on-failure`，标记测试在失败时将被跳过。
* `robot:skip`，无条件地标记测试为跳过。
* `robot:exclude`，无条件地标记测试为排除。
* `robot:private`，将关键字标记为私有。
* `robot:no-dry-run`，标记关键字不在"试运行"模式下执行。
* `robot:exit`，当执行被正常停止时，自动添加到测试中。
* `robot:flatten`，在执行时启用扁平化关键字功能。

自 RobotFramework 4.1 起，保留标签默认在标签统计中被隐藏。当通过 `--tagstatinclude robot:*` 命令行选项明确包含时，它们才会显示。

## 测试初始化与清理

Robot Framework 具备与其他许多自动化测试框架类似的测试初始化与清理功能。简而言之，测试初始化是在测试用例执行前运行的操作，而测试清理则在测试用例执行后运行。在 Robot Framework 中，初始化与清理操作本身也只是普通的关键字，可以附带参数。

一次初始化或清理始终对应一个关键字。如果需要处理多个独立任务，可以为此创建更高级别的用户关键字。另一种解决方案是使用 BuiltIn 库中的 Run Keywords 关键字来执行多个关键字。

测试清理在两个方面具有特殊性。首先，即使测试用例失败，清理也会执行，因此可用于执行那些无论测试用例状态如何都必须完成的清理活动。其次，即使清理中的某个关键字失败，其后的所有关键字仍会继续执行。这种"失败后继续运行"功能也可用于普通关键字，但在清理操作中默认启用。

在测试用例文件中为测试用例指定初始化或清理的最简单方式，是在 Setting 节中使用 `Test Setup` 和 `Test Teardown` 设置项。单个测试用例也可以拥有自己的初始化或清理，它们通过测试用例节中的 `[Setup]` 或 `[Teardown]` 设置项定义，并会覆盖可能的 `Test Setup` 和 `Test Teardown` 设置。在 `[Setup]` 或 `[Teardown]` 设置项后不指定关键字，意味着没有初始化或清理操作。也可以使用值 `NONE` 来表示测试没有初始化/清理。

```markdown
*** Settings ***
Test Setup       Open Application    App A
Test Teardown    Close Application

*** Test Cases ***
Default values
    [Documentation]    Setup and teardown from setting section
    Do Something

Overridden setup
    [Documentation]    Own setup, teardown from setting section
    [Setup]    Open Application    App B
    Do Something

No teardown
    [Documentation]    Default setup, no teardown at all
    Do Something
    [Teardown]

No teardown 2
    [Documentation]    Setup and teardown can be disabled also with special value NONE
    Do Something
    [Teardown]    NONE

Using variables
    [Documentation]    Setup and teardown specified using variables
    [Setup]    ${SETUP}
    Do Something
    [Teardown]    ${TEARDOWN}
```

作为初始化或清理操作来执行的关键字名称可以是一个变量。这样便于通过从命令行以变量形式提供关键字名称，从而在不同环境中使用不同的初始化或清理例程。

## 测试模板

测试模板能够将常规的关键字驱动型测试用例转化为数据驱动型测试。普通的基于关键字的测试用例，其主体由关键字及其可能的参数构成；而使用模板的测试用例，仅包含模板关键字的参数。我们无需在每个测试中重复使用同一个关键字，也无需在文件的所有测试中重复使用，而可以实现每个测试只用一次，甚至整个文件仅使用一次。

模板关键字既可以接收常规的位置参数和命名参数，也支持将参数内嵌于关键字名称中。与其他设置不同，无法使用变量来定义模板。

### 基本用法

下面的示例测试用例展示了如何将一个接受常规位置参数的关键字用作模板。这两个测试在功能上完全一致。

```markdown
*** Test Cases ***
Normal test case
    Example keyword    first argument    second argument

Templated test case
    [Template]    Example keyword
    first argument    second argument
```

如示例所示，可以使用 `[Template]` 设置来为单个测试用例指定模板。另一种方法是在 Settings 节中使用 `Test Template` 设置，这样该模板将应用于该测试用例文件中的所有测试。`[Template]` 设置会覆盖 Settings 节中可能设置的模板，并且将 `[Template]` 的值设为空则表示该测试没有模板，即使使用了 `Test Template` 也是如此。也可以使用值 `NONE` 来表示测试没有模板。

使用具有默认值的关键字、接受可变数量参数的关键字、使用命名参数以及自由命名参数等功能，在模板中的工作方式与在其他场景中完全相同。在参数中使用变量也照常支持。

### 含多次迭代的模板

如果一个模板化测试用例的主体包含多行数据，则该模板会逐一应用于所有行。这意味着同一个关键字会被执行多次，每次使用一行数据。

```markdown
*** Settings ***
Test Template    Example keyword

*** Test Cases ***
Templated test case
    first round 1     first round 2
    second round 1    second round 2
    third round 1     third round 2
```

模板化测试的特殊之处在于，即使其中一个或多个迭代失败或跳过，所有迭代仍会执行。具有多次迭代的模板化测试的汇总结果如下：

* FAIL：如果任意一次迭代失败。
* PASS：如果没有任何失败，且至少有一次迭代通过。
* SKIP：如果所有迭代都被跳过。

### 使用嵌入式参数的模板

模板支持一种嵌入式参数的变体语法。在模板中，该语法的工作原理是：如果模板关键字的名称中包含变量，这些变量会被视为参数的占位符，并被替换为模板所使用的实际参数。然后，生成的关键字将被直接调用，而不再附带位置参数。这通过一个示例能最清晰地说明：

```markdown
*** Test Cases ***
Normal test case with embedded arguments
    The result of 1 + 1 should be 2
    The result of 1 + 2 should be 3

Template with embedded arguments
    [Template]    The result of ${calculation} should be ${expected}
    1 + 1    2
    1 + 2    3

*** Keywords ***
The result of ${calculation} should be ${expected}
    ${result} =    Calculate    ${calculation}
    Should Be Equal    ${result}     ${expected}
```

当模板使用嵌入式参数时，模板关键字名称中的参数数量必须与使用时传入的参数数量相匹配。不过，参数名称不需要与原始关键字中的参数匹配，并且完全可以使用完全不同的参数：

```markdown
*** Test Cases ***
Different argument names
    [Template]    The result of ${foo} should be ${bar}
    1 + 1    2
    1 + 2    3

Only some arguments
    [Template]    The result of ${calculation} should be 3
    1 + 2
    4 - 1

New arguments
    [Template]    The ${meaning} of ${life} should be 42
    result    21 * 2
```

在模板中使用嵌入式参数的主要好处在于，参数名称被显式地指定了。当使用普通参数时，可以通过为包含参数的数据列命名来实现相同的效果。这将在下一节数据驱动风格的示例中进行说明。

### 结合 FOR 循环的模板

如果模板与 FOR 循环结合使用，则该模板将应用于循环内的所有步骤。在这种情况下，同样会启用“失败后继续”的模式，这意味着即使存在失败，所有步骤仍会对所有循环元素执行。

```markdown
*** Test Cases ***
Template with FOR loop
    [Template]    Example keyword
    FOR    ${item}    IN    @{ITEMS}
        ${item}    2nd arg
    END
    FOR    ${index}    IN RANGE    42
        1st arg    ${index}
    END
```

### 结合 IF/ELSE 结构的模板

IF/ELSE 结构也可以与模板结合使用。这很有用，例如，当与 FOR 循环一起使用时，可以过滤要执行的参数。

```markdown
*** Test Cases ***
Template with FOR and IF
    [Template]    Example keyword
    FOR    ${item}    IN    @{ITEMS}
        IF  ${item} < 5
            ${item}    2nd arg
        END
    END
```

## 不同的测试用例风格

编写测试用例有几种不同的方式。描述某种工作流的测试用例可以采用关键字驱动或行为驱动风格来编写。数据驱动风格则可用于使用不同的输入数据来测试相同的工作流程。

### 关键字驱动风格

工作流测试（例如前面描述的 `Valid Login` 测试）由多个关键字及其可能的参数构成。其典型结构是：首先将系统置于初始状态（`Valid Login` 示例中的 `Open Login Page`），然后对系统执行某些操作（`Input Name`、`Input Password`、`Submit Credentials`），最后验证系统是否按预期响应（`Welcome Page Should Be Open`）。

### 数据驱动风格

编写测试用例的另一种方式是数据驱动方法。在这种风格中，测试用例只使用一个更高级别的关键字（通常创建为用户关键字）来隐藏实际测试流程。当需要使用不同的输入和/或输出数据测试同一场景时，这类测试非常有用。尽管可以在每个测试中重复使用相同的关键字，但测试模板功能允许只指定一次要使用的关键字。

```markdown
*** Settings ***
Test Template    Login with invalid credentials should fail

*** Test Cases ***                USERNAME         PASSWORD
Invalid User Name                 invalid          ${VALID PASSWORD}
Invalid Password                  ${VALID USER}    invalid
Invalid User Name and Password    invalid          invalid
Empty User Name                   ${EMPTY}         ${VALID PASSWORD}
Empty Password                    ${VALID USER}    ${EMPTY}
Empty User Name and Password      ${EMPTY}         ${EMPTY}
```

上面的示例包含了六个独立的测试，分别对应每种无效的用户名/密码组合。而下面的示例则展示了如何将所有组合整合在一个测试中。使用测试模板时，即使测试过程中出现失败，测试内的所有轮次仍会执行，因此这两种风格在功能上并没有本质区别。在上面的示例中，每个组合都有独立的命名，这使测试内容更加清晰易读；但如果有大量此类测试，可能会影响统计结果的整洁度。具体采用哪种风格取决于具体情境和个人偏好。

```markdown
*** Test Cases ***
Invalid Password
    [Template]    Login with invalid credentials should fail
    invalid          ${VALID PASSWORD}
    ${VALID USER}    invalid
    invalid          whatever
    ${EMPTY}         ${VALID PASSWORD}
    ${VALID USER}    ${EMPTY}
    ${EMPTY}         ${EMPTY}
```

### 行为驱动风格

测试用例也可以编写成非技术性项目相关人员也需理解的需求形式。这类可执行的需求通常被称为验收测试驱动开发（ATDD）或实例化规格，是这一开发流程的基石。

编写这些需求/测试的一种方式是采用行为驱动开发（BDD）所推广的 `Given-When-Then` 风格。以此风格编写测试用例时，初始状态通常使用以 `Given` 开头的关键字描述，操作使用以 `When` 开头的关键字描述，预期结果则使用以 `Then` 开头的关键字描述。如果一个步骤包含多个动作，还可以使用以 `And` 或 `But` 开头的关键字。

```markdown
*** Test Cases ***
Valid Login
    Given login page is open
    When valid username and password are inserted
    and credentials are submitted
    Then welcome page should be open
```

#### 忽略 Given/When/Then/And/But 前缀

在创建关键字时，可以省略 `Given、When、Then、And` 和 `But` 这些前缀。例如，上面例子中的 `Given login page is open` 通常在实现时会省略单词 `Given`，使其名称仅为 `Login page is open`。省略前缀使得同一个关键字可以与不同的前缀一起使用。例如，`Welcome page should be open` 可以同时用作 `Then welcome page should be open` 或 `And welcome page should be open`。

#### 将数据嵌入到关键字中

在编写具体示例时，能够将实际数据传递给关键字实现非常有用。这可以通过将参数嵌入到关键字名称中来实现。