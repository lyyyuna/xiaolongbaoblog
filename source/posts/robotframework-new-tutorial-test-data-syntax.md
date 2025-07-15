title: Robot Framework 新版教程 - 测试数据语法
date: 2025-07-11 14:50:08
series: Robot Framework 7.x 教程

---

本文介绍了 Robot Framework 的测试数据语法，后续章节将详细讲解如何实际创建测试用例、测试套件等内容。虽然本节主要使用"测试"这一术语，但这些规则同样适用于创建任务。

## 文件和目录

构建测试用例的层次结构安排如下：

* 测试用例创建在套件文件中。
* 一个测试用例文件会自动创建一个包含该文件内所有测试用例的测试套件。
* 包含测试用例文件的目录会形成一个更高层级的测试套件。此类套件目录的子测试套件即是由其中的测试用例文件生成的套件。
* 测试套件目录也可以包含其他测试套件目录，并且这种层级结构可以根据需要无限嵌套。
* 测试套件目录可以包含特殊的初始化文件，用于配置所创建的测试套件。

除此之外，还包括：

* 测试库（Test libraries） - 包含最底层的关键字
* 资源文件（Resource files） - 存放变量及高层级用户关键字
* 变量文件（Variable files） - 提供比资源文件更灵活的变量定义方式

测试用例文件、测试套件初始化文件和资源文件均采用 Robot Framework 测试数据语法编写。而测试库和变量文件则使用"真正的"编程语言（通常为 Python）创建。

## 测试数据 section

Robot Framework 测试数据通过以下不同 section（通常也称为表格）进行定义：

* **Settings**，导入测试库、资源文件和变量文件，为测试套件和测试用例定义元数据。
* **Variables**，定义可在测试数据其他位置使用的变量。
* **Test Cases**，从可用的关键字创建测试用例。
* **Tasks**，从可用的关键字创建任务。单个文件只能包含测试或任务中的一种。
* **Keywords**，基于现有底层关键字创建用户关键字。
* **Comments**，附加注释或数据（Robot Framework 将忽略此内容）。

不同 section 通过其 header 行进行识别。推荐的 header 格式是 `** Settings **`，但 header 不区分大小写，前后空格可选，且开头的星号数量可以变化（只要至少有一个星号）。例如，`*settings` 也会被识别为 section header。

Robot Framework 也支持单数形式的 header（如 *** Setting ***），但从 Robot Framework 6.0 开始该支持已被弃用。自 Robot Framework 7.0 起会显示弃用警告，最终将完全不再支持单数形式的 header。

header 行除了实际的 section header 外，还可以包含其他内容。额外内容必须使用数据格式相关的分隔符（通常是两个或更多空格）与 section header 隔开。解析时会忽略这些额外内容，但可用于文档说明。这在采用数据驱动风格创建测试用例时特别有用。

第一个 section 之前的所有数据都会被忽略。

## 支持的文件格式

Robot Framework 最常见的数据创建方式是采用空格分隔格式，其中数据单元（如关键字及其参数）之间用两个或更多空格分隔。另一种方式是使用管道分隔格式，其分隔符为带空格的管道符（ | ）。

测试套件文件通常使用 `.robot` 扩展名，但实际可解析的文件类型可配置。资源文件也可使用 `.robot` 扩展名，但推荐使用专用的 `.resource` 扩展名，未来可能强制要求使用该扩展名。包含非 ASCII 字符的文件必须使用 UTF-8 编码保存。

Robot Framework 也支持 reStructuredText 文件，其中标准的 Robot Framework 数据被嵌入到代码块中。默认情况下，只有扩展名为 `.robot.rst` 的文件会被解析。若希望使用 `.rst` 或 `.rest` 扩展名，需单独配置。

Robot Framework 数据还可采用 JSON 格式创建，该格式主要面向工具开发者而非普通用户。默认情况下，只有使用自定义 `.rbt` 扩展名的 JSON 文件会被解析。

早期 Robot Framework 版本还支持 HTML 和 TSV 格式的数据。若数据与空格分隔格式兼容，TSV 格式仍可使用，但对 HTML 格式的支持已完全移除。若遇到此类数据文件，需将其转换为纯文本格式才能在 Robot Framework 3.2 或更新版本中使用。最简单的方法是使用 Tidy 工具，但必须使用 Robot Framework 3.1 附带的版本，因为更新版本已完全不支持 HTML 格式。

### 空格分隔的格式

当 Robot Framework 解析数据时，首先会将数据拆分成行，然后将每行拆分为多个标记（如关键字和参数）。在使用空格分隔格式时，标记之间的分隔符为两个及以上空格，或者一个及以上制表符。除普通的 ASCII 空格外，任何被视为空格的 Unicode 字符（如不间断空格）均可作为分隔符。只要保证至少有两个空格，用作分隔符的空格数量可以变化——这样在配置等场景中就能通过合理对齐数据来提升可读性。

```markdown
*** Settings ***
Documentation     Example using the space separated format.
Library           OperatingSystem

*** Variables ***
${MESSAGE}        Hello, world!

*** Test Cases ***
My Test
    [Documentation]    Example test.
    Log    ${MESSAGE}
    My Keyword    ${CURDIR}

Another Test
    Should Be Equal    ${MESSAGE}    Hello, world!

*** Keywords ***
My Keyword
    [Arguments]    ${path}
    Directory Should Exist    ${path}
```

由于制表符和连续空格均被视为分隔符，若需在关键字参数或实际数据中使用这些字符时，必须进行转义处理。可通过特殊转义语法实现，例如：

* 使用 `\t` 表示制表符
* 使用 `\xA0` 表示不间断空格
* 使用内置变量 `${SPACE}` 和 `${EMPTY}`

### 管道分隔的格式

空格分隔格式的最大痛点在于：关键字与参数的视觉区分较为困难。当关键字需要接收大量参数，或参数本身包含空格时，这一问题尤为突出。此时采用管道分隔格式往往更优，因为管道符能提供更明显的视觉分隔效果。

同一文件可同时包含空格分隔行和管道分隔行。管道分隔行通过行首强制的管道符识别（行尾管道符可选），除行首行尾外，每个管道符两侧必须至少保留一个空格或制表符。虽然不强制要求管道符纵向对齐，但保持对齐能显著提升代码可读性。

```markdown
| *** Settings ***   |
| Documentation      | Example using the pipe separated format.
| Library            | OperatingSystem

| *** Variables ***  |
| ${MESSAGE}         | Hello, world!

| *** Test Cases *** |                 |               |
| My Test            | [Documentation] | Example test. |
|                    | Log             | ${MESSAGE}    |
|                    | My Keyword      | ${CURDIR}     |
| Another Test       | Should Be Equal | ${MESSAGE}    | Hello, world!

| *** Keywords ***   |                        |         |
| My Keyword         | [Arguments]            | ${path} |
|                    | Directory Should Exist | ${path} |
```

在使用管道分隔格式时，参数内部的连续空格或制表符无需转义处理。同样，空列通常也无需转义——除非位于行末。但需特别注意：若实际测试数据中包含被空格包围的管道符（ | ），则必须使用反斜杠进行转义。


```markdown
| *** Test Cases *** |                 |                 |                      |
| Escaping Pipe      | ${file count} = | Execute Command | ls -1 *.txt \| wc -l |
|                    | Should Be Equal | ${file count}   | 42                   |
```

### reStructuredText 格式

reStructuredText（reST）是一种易读的纯文本标记语法，常用于 Python 项目的文档编写，包括 Python 官方文档及本用户指南。reST 文档通常被编译为 HTML 格式，同时也支持其他输出格式。将 reST 与 Robot Framework 结合使用，既能编写格式丰富的文档，又能以简洁的文本格式保存测试数据，便于通过简易文本编辑器、差异比对工具和版本控制系统进行操作。

在 Robot Framework 与 reStructuredText 文件结合使用时，常规的 Robot Framework 数据会被嵌入到所谓的代码块(code blocks)中。标准的 reST 代码块使用 `code` 指令进行标记，但 Robot Framework 同时支持 [Sphinx](http://sphinx-doc.org/) 工具所使用的 `code-block` 或 `sourcecode` 指令格式。

```markdown
reStructuredText example
------------------------

code 块之外的文字会被忽略。

.. code:: robotframework

   *** Settings ***
   Documentation    Example using the reStructuredText format.
   Library          OperatingSystem

   *** Variables ***
   ${MESSAGE}       Hello, world!

   *** Test Cases ***
   My Test
       [Documentation]    Example test.
       Log    ${MESSAGE}
       My Keyword    ${CURDIR}

   Another Test
       Should Be Equal    ${MESSAGE}    Hello, world!

此段文本同样位于代码块之外，将被忽略。不包含 Robot Framework 数据的代码块也会被忽略。

.. code:: robotframework

   # Both space and pipe separated formats are supported.

   | *** Keywords ***  |                        |         |
   | My Keyword        | [Arguments]            | ${path} |
   |                   | Directory Should Exist | ${path} |

.. code:: python

   # This code block is ignored.
   def example():
       print('Hello, world!')
```

Robot Framework 支持使用以下扩展名的 reStructuredText 文件：

* `.robot.rst`
* `.rst`
* `.rest`

为避免解析无关的 reStructuredText 文件，默认情况下执行目录时仅会解析 `.robot.rst` 扩展名的文件。如需解析其他扩展名的文件，可通过以下任一命令行选项启用：

* `--parseinclude`
* `--extension`

当 Robot Framework 解析 reStructuredText 文件时，会忽略低于 `SEVERE` 级别的错误，以避免因非标准指令等标记产生的干扰信息。虽然这可能掩盖部分真实错误，但这些错误在使用标准 reStructuredText 工具处理文件时仍可被发现。

### JSON 格式

Robot Framework 同时支持 JSON 格式的数据文件。该格式主要面向工具开发者而非普通用户，且不建议手动编辑。其主要应用场景包括：

* 跨进程/机器数据传输。测试套件可在某台机器转换为 JSON 格式，并在其他机器重建。
* 高效存储测试套件。将常规 Robot Framework 数据（包括嵌套套件）保存为单个 JSON 文件，可显著提升解析速度。
* 外部工具测试生成。为生成测试/任务的外部工具提供替代数据格式。

#### 把套件转换为 JSON 格式

可通过 [TestSuite.to_json](https://robot-framework.readthedocs.io/en/master/autodoc/robot.running.html#robot.running.model.TestSuite.to_json) 方法将测试套件结构序列化为 JSON 格式。该方法在不带参数调用时，会以字符串形式返回 JSON 数据；同时也支持接收文件路径或已打开的文件对象作为写入目标，并可通过配置选项调整 JSON 格式输出：

```py
from robot.running import TestSuite


# 基于文件系统中的数据创建测试套件
suite = TestSuite.from_file_system('/path/to/data')

# 以字符串形式获取 JSON 数据
data = suite.to_json()

# 使用自定义缩进格式将 JSON 数据保存至文件
suite.to_json('data.rbt', indent=2)
```

若你更倾向使用 Python 原生数据结构，并自行转换为 JSON 或其他格式，可改用 [TestSuite.to_dict](https://robot-framework.readthedocs.io/en/master/autodoc/robot.running.html#robot.running.model.TestSuite.to_dict) 方法实现。

#### 从 JSON 格式转换成套件

可通过 [TestSuite.from_json](https://robot-framework.readthedocs.io/en/master/autodoc/robot.running.html#robot.running.model.TestSuite.from_json) 方法基于 JSON 数据构建测试套件，该方法同时支持以下两种输入形式：

* JSON 格式字符串
* JSON 文件路径

```py
from robot.running import TestSuite


# 从 JSON 文件创建测试套件
suite = TestSuite.from_json('data.rbt')

# 从 JSON 字符串创建测试套件
suite = TestSuite.from_json('{"name": "Suite", "tests": [{"name": "Test"}]}')

# 执行测试套件（注意：需单独生成日志和报告）
suite.run(output='example.xml')
```

若已有 Python 字典格式的数据，可改用 [TestSuite.from_dict](https://robot-framework.readthedocs.io/en/master/autodoc/robot.running.html#robot.running.model.TestSuite.from_dict) 方法。无论采用何种方式重建测试套件，其结果仅存在于内存中，不会在文件系统重建原始数据文件。

如上述示例所示，可通过 [TestSuite.run](https://robot-framework.readthedocs.io/en/master/autodoc/robot.running.html#robot.running.model.TestSuite.run) 方法执行创建的测试套件。但直接执行 JSON 文件可能更为简便，具体方法将在下节说明。

#### 执行 JSON 文件

通过 `robot` 命令执行测试或任务时，系统会自动解析采用自定义 `.rbt` 扩展名的 JSON 文件，包括以下两种场景：

* 直接运行单个 JSON 文件（如 `robot tests.rbt`）
* 运行包含 `.rbt` 文件的目录

若需使用标准 `.json` 扩展名，则需额外配置待解析文件。

#### 调整套件源

通过 `TestSuite.to_json` 和 `TestSuite.to_dict` 获取的数据中，套件源路径(suite source)采用绝对路径格式。若后续在不同机器上重建套件，源路径可能与目标机器的目录结构不匹配。为解决此问题，可先使用 [TestSuite.adjust_source](https://robot-framework.readthedocs.io/en/master/autodoc/robot.model.html#robot.model.testsuite.TestSuite.adjust_source) 方法将套件源转为相对路径再获取数据，待套件重建后补充正确的根目录路径：

```py
from robot.running import TestSuite


# 创建测试套件 -> 调整源路径 -> 转换为JSON格式
suite = TestSuite.from_file_system('/path/to/data')  # 从文件系统加载原始套件
suite.adjust_source(relative_to='/path/to')          # 将源路径转换为相对于指定目录的路径
suite.to_json('data.rbt')                            # 序列化为JSON文件

# 在其他环境重建套件并相应调整源路径
suite = TestSuite.from_json('data.rbt')             # 从JSON文件重建套件
suite.adjust_source(root='/new/path/to')            # 根据新环境设置根目录路径
```


#### JSON 结构

在生成的 JSON 数据中，测试套件文件中的以下元素将与测试用例/任务一同被包含：导入项、变量、关键字。具体 JSON 结构规范详见 `running.json` schema文件。

## 解析测试数据的规则

### 忽略数据

Robot Framework 解析测试数据文件时会自动忽略以下内容：

* 首个测试数据 section 之前的所有内容
* Comments section 内的全部数据
* 所有空白行
* 使用管道分隔格式时，行末的空单元格
* 未用于转义的单反斜线符号（\）
* 当井号（#）出现在单元格开头时，该单元格后续所有字符（支持在测试数据中添加行内注释）

Robot Framework 忽略的数据将不会出现在任何结果报告中，且大多数配套工具也会忽略这些数据。如需在输出中显示信息，请将其放入测试用例或套件的文档或其他元数据中，或使用 BuiltIn 关键字 Log 或 Comment 进行记录。

### 转义

Robot Framework 测试数据中的转义字符是反斜杠（\），此外内置变量 `${EMPTY}` 和 `${SPACE}` 也常用于转义操作。不同转义机制的具体用法将在后续章节详细讨论。

#### 转义特殊字符

反斜杠字符（\）可用于转义特殊字符，使其保留字面值。

| 转义字符 | 功能说明                          | 示例代码                  |
|-------|-------|-------|
| `\$`     | 避免被识别为标量变量起始符          | `${notvar}`             |
| `\@`     | 避免被识别为列表变量起始符          | `@{notvar}`             |
| `\&`     | 避免被识别为字典变量起始符          | `&{notvar}`             |
| `\%`     | 避免被识别为环境变量起始符          | `%{notvar}`             |
| `\#`     | 避免被识别为注释起始符             | `# not comment`         |
| `\=`     | 避免被识别为命名参数语法            | `not=named`             |
| `\|`     | 在管道分隔格式中不作为分隔符         | `ls -1 *.txt \| wc -l`  |
| `\\`     | 作为普通字符使用（不进行转义）       | `c:\\temp`, `\\${var}`  |


#### 构建转义序列

反斜杠字符（\）还可用于创建特殊转义序列，这些序列会被解析为测试数据中难以直接输入的字符。

| 转义序列       | 说明                      | 示例                      |
|-------|-------|-------|
| `\n`          | 换行符                    | `first line\n2nd line`    |
| `\r`          | 回车符                    | `text\rmore text`         |
| `\t`          | 制表符                    | `text\tmore text`         |
| `\xhh`        | 十六进制值字符（2位）       | 空字符：`\x00`<br>ä：`\xE4` |
| `\uhhhh`      | 十六进制值字符（4位）       | 雪人符号：`\u2603`        |
| `\Uhhhhhhhh`  | 十六进制值字符（8位）       | 爱情酒店符号：`\U0001f3e9` |

#### 处理空值

在使用空格分隔格式时，用作分隔符的空格数量可以变化，因此除非进行转义处理，否则无法识别空值。空单元格可通过反斜杠字符或内置变量 `${EMPTY}` 进行转义。通常推荐使用后者，因为更易于理解。



```markdown
*** Test Cases ***
Using backslash
    Do Something    first arg    \
    Do Something    \            second arg

Using ${EMPTY}
    Do Something    first arg    ${EMPTY}
    Do Something    ${EMPTY}     second arg
```

在使用竖线分隔格式时，空值仅当位于行末时才需要进行转义：


```markdown
| *** Test Cases *** |              |           |            |
| Using backslash    | Do Something | first arg | \          |
|                    | Do Something |           | second arg |
|                    |              |           |            |
| Using ${EMPTY}     | Do Something | first arg | ${EMPTY}   |
|                    | Do Something |           | second arg |
```

#### 处理空格

在关键字参数或其他情况下，空格（尤其是连续空格）会引发问题，主要原因有二：

* 若采用空格分隔格式，两个及以上连续空格会被视作分隔符
* 若采用竖线分隔格式，首尾空格会被自动忽略

此时必须对空格进行转义处理。与空值转义方式类似，既可使用反斜杠字符（\），也可调用内置变量 `${SPACE}` 实现转义。

| 反斜杠转义方式          | ${SPACE}变量转义方式           | 技术说明                                                                 |
|-------|-------|-------|
| `\ leading space`      | `${SPACE}leading space`       | 前导空格转义                                                           |
| `trailing space \`     | `trailing space${SPACE}`      | **反斜杠必须跟在空格后**                                               |
| `\ \`                 | `${SPACE}`                    | 需在两侧使用反斜杠                                                     |
| `consecutive \ \ spaces` | `consecutive${SPACE * 3}spaces` | 使用扩展变量语法                                  |

如上述示例所示，使用 `${SPACE}` 变量通常能让测试数据更易于理解。当需要多个空格时，结合扩展变量语法使用尤为方便。

### 将数据分成多行

当数据过长无法完整显示在一行时，可通过省略号（`...`）进行分行处理。省略号可采用与起始行相同的缩进格式，且其后必须跟随常规的测试数据分隔符。

在多数情况下，分行数据与未分行数据具有完全相同的语义。但套件文档、测试用例文档、关键字文档以及套件元数据除外 —— 这些内容的分行值会自动以换行符连接，便于创建多行文本值。

此 `...` 语法同样适用于变量表中的变量分行。当长标量变量（如 `${STRING}`）被分割至多行时，最终值将通过行间拼接获得，默认以空格作为分隔符。如需更改分隔方式，可在值前添加 `SEPARATOR=<sep>` 声明。

以下两个示例展示了包含完全相同数据的不分行与分行处理方案：

```markdown
*** Settings ***
Documentation      Here we have documentation for this suite.\nDocumentation is often quite long.\n\nIt can also contain multiple paragraphs.
Test Tags          test tag 1    test tag 2    test tag 3    test tag 4    test tag 5

*** Variables ***
${STRING}          This is a long string. It has multiple sentences. It does not have newlines.
${MULTILINE}       This is a long multiline string.\nThis is the second line.\nThis is the third and the last line.
@{LIST}            this     list     is    quite    long     and    items in it can also be long
&{DICT}            first=This value is pretty long.    second=This value is even longer. It has two sentences.

*** Test Cases ***
Example
    [Tags]    you    probably    do    not    have    this    many    tags    in    real    life
    Do X    first argument    second argument    third argument    fourth argument    fifth argument    sixth argument
    ${var} =    Get X    first argument passed to this keyword is pretty long    second argument passed to this keyword is long too
```

```markdown
*** Settings ***
Documentation      Here we have documentation for this suite.
...                Documentation is often quite long.
...
...                It can also contain multiple paragraphs.
Test Tags          test tag 1    test tag 2    test tag 3
...                test tag 4    test tag 5

*** Variables ***
${STRING}          This is a long string.
...                It has multiple sentences.
...                It does not have newlines.
${MULTILINE}       SEPARATOR=\n
...                This is a long multiline string.
...                This is the second line.
...                This is the third and the last line.
@{LIST}            this     list     is      quite    long     and
...                items in it can also be long
&{DICT}            first=This value is pretty long.
...                second=This value is even longer. It has two sentences.

*** Test Cases ***
Example
    [Tags]    you    probably    do    not    have    this    many
    ...       tags    in    real    life
    Do X    first argument    second argument    third argument
    ...    fourth argument    fifth argument    sixth argument
    ${var} =    Get X
    ...    first argument passed to this keyword is pretty long
    ...    second argument passed to this keyword is long too
```

## 本地化

Robot Framework 的本地化工作始于 6.0 版本，该版本支持翻译以下内容：

* section header
* 设置项
* 行为驱动开发(BDD)中使用的 Given/When/Then 前缀
* 布尔参数自动转换使用的 true 和 false 字符串

未来计划进一步扩展本地化支持范围，例如扩展到日志和报告模块，并可能包含控制结构。

### 启用语言支持

#### 使用命令行开启

激活语言支持的主要方式是通过命令行使用 `--language` 选项进行指定。对于内置语言，既可使用语言名称（如 `Finnish`）也可使用语言代码（如 `fi`）。名称和代码均不区分大小写及空格，连字符（`-`）也会被忽略。如需启用多语言支持，需多次使用 `--language` 选项：

```bash
robot --language Finnish testit.robot
robot --language pt --language ptbr testes.robot
```

该 `--language` 选项同样适用于激活自定义语言文件。此时参数值可以是文件路径，若文件位于模块搜索路径中，也可直接使用模块名：

```bash
robot --language Custom.py tests.robot
robot --language MyLang tests.robot
```

出于向后兼容性考虑，以及支持部分翻译场景，系统会始终自动激活英语支持。未来版本可能会提供禁用该特性的选项。

#### 文件预配置

还可以直接在数据文件中通过在任何节标题前添加一行 `Language: <value>`（不区分大小写）来启用语言。冒号后的值与 `--language` 选项的解析方式相同：

```bash
Language: Finnish

*** Asetukset ***
Dokumentaatio        Finnish language example.
```

若需启用多语言，可重复 `Language:` 行。此类配置行不可置于注释中，因此类似 `# Language: Finnish` 的写法无效。

由于技术限制，单文件语言配置会影响后续文件的解析及整个执行过程。此行为未来可能变更，请勿依赖该特性。若使用单文件配置，请确保所有文件统一使用该方式，或通过` --language` 选项全局启用语言。

### 内置语言支持

* Arabic (ar)
* Bulgarian (bg)
* Bosnian (bs)
* Czech (cs)
* German (de)
* Spanish (es)
* Finnish (fi)
* French (fr)
* Hindi (hi)
* Italian (it)
* Japanese (ja)
* Korean (ko)
* Dutch (nl)
* Polish (pl)
* Portuguese (pt)
* Brazilian Portuguese (pt-BR)
* Romanian (ro)
* Russian (ru)
* Swedish (sv)
* Thai (th)
* Turkish (tr)
* Ukrainian (uk)
* Vietnamese (vi)
* Chinese Simplified (zh-CN)
* Chinese Traditional (zh-TW)

### 自定义语言文件

若所需语言未内置，或需为特定需求创建完全自定义的语言，你可以轻松创建自定义语言文件。语言文件是 Python 文件，内含一个或多个语言定义——启用该文件时将加载所有定义。语言定义通过继承 `robot.api.Language` 基类并按需覆盖类属性来实现：

```py
from robot.api import Language

class Example(Language):
    test_cases_header = 'Validations'      
    tags_setting = 'Labels'                
    given_prefixes = ['Assuming']         
    true_strings = ['OK', '\N{THUMBS UP SIGN}']  
```

假设上述代码保存在 `example.py` 文件中，启用语言文件时只需提供该文件路径或模块名 `example` 即可。

此示例仅实现了部分可翻译项（因英语会默认启用）。多数属性需指定为字符串，但BDD前缀和真/假值字符串允许使用列表形式定义多个值。更多示例可参考 Robot Framework 内置的 [languages](https://github.com/robotframework/robotframework/blob/master/src/robot/conf/languages.py) 模块——该模块包含 `Language` 类及所有内置语言定义。

## 风格

Robot Framework 的语法构建了一种简易编程语言，与其他编程语言类似，代码风格的规范性至关重要。虽然 Robot Framework 的语法设计具有高度灵活性，但仍存在一些普遍推荐的规范：

* 缩进：采用 4 个空格。
* 间隔：关键字与参数、设置项与其值之间保留 4 个空格。某些场景（如 Settings/Variables 中的值对齐或数据驱动风格）可适当增加空格。
* 变量命名：全局变量使用大写字母（如 `${EXAMPLE}`），局部变量使用小写字母（如 `${example}`）
* 一致性：确保单文件内风格统一，推荐整个项目保持相同规范

目前尚未形成强约束规范的场景是关键字的命名格式。Robot Framework 官方文档通常采用首字母大写形式（如 `Example Keyword`），该风格在测试数据中也较为常见。但对于较长的句子型关键字（例如 `Log into system as an admin`），这种格式可能不够理想。

建议使用 Robot Framework 的团队制定专属编码规范。社区开发的[《Robot Framework 风格指南》](https://docs.robotframework.org/docs/style_guide)可作为优质基准模板，支持按需调整。此外，可通过 [Robocop](https://robocop.readthedocs.io/) 静态检查工具和 [Robotidy](https://robotidy.readthedocs.io/) 代码格式化工具强制实施这些规范。