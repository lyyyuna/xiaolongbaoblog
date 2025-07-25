title: Robot Framework 新版教程 - 入门指南
date: 2025-07-02 14:50:08
series: Robot Framework 7.x 教程

---

## 前言

十年前我翻译过 [Robot Framework 的一些列文章](https://www.lyyyuna.com/series/Robot%20Framework%20%E6%95%99%E7%A8%8B/)，这些年其实我早已不再使用这个框架，但是看到搜索引擎仍有读者访问我翻译的旧文章，因此我决定更新一下。

## 介绍

Robot Framework 是一个基于 Python 的、可扩展的关键字驱动自动化框架，适用于以下领域：

1. 验收测试（Acceptance Testing）
2. 验收测试驱动开发（ATDD）
3. 行为驱动开发（BDD）
4. 机器人流程自动化（RPA）

该框架可在分布式异构环境中使用，支持跨不同技术和接口的自动化需求。

### 为什么使用 Robot Framework？

1. 直观的表格化语法
    * 提供统一、易用的表格语法编写测试用例
2. 关键字复用体系
    * 支持基于现有关键字创建更高层次的可复用关键字
3. 可视化测试报告
    * 生成易读的HTML格式测试报告与日志
4. 跨平台独立性
    * 与平台和被测应用解耦
5. 可扩展库架构
    * 提供简洁的库 API，支持使用 Python 原生开发定制测试库
6. 持续集成支持
    * 提供命令行接口与基于 XML 的输出文件，轻松对接 CI 系统
7. 全栈测试能力，支持测试类型包括：
    * Web应用
    * REST APIs
    * 移动应用
    * 进程监控
    * 远程系统（Telnet/SSH等）
8. 数据驱动测试
    * 支持创建数据驱动的测试用例
9. 多环境变量支持
    * 内置变量机制，特别适合多环境测试
10. 智能测试分类
    * 通过标签(tag)体系实现测试用例的分类与筛选
11. 版本控制集成
    * 测试套件以文件/目录形式存储，可与产品代码同步版本管理
12. 生命周期控制
    * 提供测试用例级和测试套件级的setup/teardown机制
13. 模块化架构
    * 支持为具有多接口的复杂应用创建测试

### 架构

Robot Framework 是一个通用的、与应用和具体技术无关的框架。它具有一个高度模块化的架构，如下图所示：

![Robot Framework 架构](/img/posts/rf-new-tutor/architecture.png)

测试数据采用简单、易编辑的表格格式。Robot Framework 启动时会对数据进行处理，执行测试用例并生成日志和报告。该核心框架本身不涉及任何被测目标信息，与目标的交互均由库文件处理。这些库既可直接调用应用程序接口，也能通过底层测试工具作为驱动进行操作。

### 截图

以下截图展示了测试数据样例及生成的报告和日志。

![测试用例文件](/img/posts/rf-new-tutor/testdata_screenshots.png)

![测试报告和日志](/img/posts/rf-new-tutor/screenshots.png)

## 安装步骤

本指南介绍了在不同操作系统上安装 Robot Framework 及其前置条件的方法。若已安装 Python，则可通过标准包管理器 pip 安装 Robot Framework：

```bash
$ pip install robotframework
```

### 安装 Python

Robot Framework 基于 Python 实现，因此安装前需预先安装 Python 或其替代实现 PyPy。另一项推荐的前置条件是确保已配置 pip 包管理器。

Robot Framework 要求 Python 版本为 3.8 或更高。

#### 在 Linux 上安装 Python

在 Linux 系统上，默认情况下应该已安装合适的 Python 版本并自带 pip 工具。若未安装，则需要查阅所用发行版的官方文档了解安装方法。如果想使用发行版默认提供版本之外的其它 Python 版本，同样需要参考发行版文档进行操作。

要检查当前安装的Python版本，可在终端中运行以下命令：

```bash
$ python --version
Python 3.10.13
```

请注意，如果你的 Linux 发行版同时提供了较旧的 Python 2，直接运行 `python` 命令可能会调用 Python 2。要使用 Python 3，你可以使用 `python3` 命令，或者更精确地指定版本（例如 `python3.8`）。如果你安装了多个 Python 3 版本，并且需要明确指定使用哪一个，同样需要使用这些带版本号的命令：

```bash
$ python3.11 --version
Python 3.11.7
$ python3.12 --version
Python 3.12.1
```

在系统自带的 Python 环境下直接安装 Robot Framework 存在一定风险，可能导致操作系统依赖的 Python 环境出现问题。如今大多数 Linux 发行版默认采用用户级安装（user installs）来避免此类情况，但用户也可以自行选择使用虚拟环境（virtual environments）进行隔离。

#### 在 Windows 上安装 Python

在 Windows 系统上，Python 默认并未预装，但安装过程十分简便。推荐通过 [python.org](http://python.org/) 下载官方 Windows 安装程序进行安装。若需了解其他安装方式（如通过 Microsoft Store 安装），请参阅 [Python 官方文档](https://docs.python.org/3/using/windows.html)。

在 Windows 上安装 Python 时，建议将 Python 添加至 PATH 环境变量，以便通过命令行更便捷地运行 Python 及其相关工具（如 pip 和 Robot Framework）。若使用[官方安装程序](https://docs.python.org/3/using/windows.html#windows-full)，只需在初始对话框勾选"将 Python 3.x 添加到 PATH"选项即可。

要验证 Python 是否安装成功并已添加至 PATH 环境变量，可打开命令提示符并执行以下命令：

```bash
C:\>python --version
Python 3.10.9
```

在 Windows 系统中安装多个 Python 版本时，执行 `python` 命令将默认调用 PATH 环境变量中优先级最高的版本。如需使用其他版本，最简单的方法是使用 [py launcher](https://docs.python.org/3/using/windows.html#launcher)：

```bash
C:\>py --version
Python 3.10.9
C:\>py -3.12 --version
Python 3.12.1
```

#### 在 macOS 上安装 Python

在 macOS 系统中，默认提供的 Python 版本不兼容 Python 3，因此需要单独安装。推荐访问 [python.org](http://python.org/) 下载官方 macOS 安装程序进行安装。若使用 [Homebrew](https://brew.sh/) 等包管理器，也可以通过它来安装 Python。

与其他操作系统相同，可以在 macOS 上使用 `python --version` 命令来验证 Python 是否安装成功。



#### 安装 PyPy

PyPy 是 Python 的另一种实现方案。相较于标准 Python 实现，其主要优势在于运行速度更快且内存占用更低，但实际效果取决于具体使用场景。若执行效率至关重要，至少尝试测试 PyPy 通常是个不错的选择。

PyPy 的安装过程简单直接，你可在 [pypy.org](http://pypy.org/) 获取安装程序及详细指南。要验证 PyPy 是否安装成功，可运行以下命令：

```bash
pypy --version

pypy3 --version
```

#### 配置 PATH

PATH 环境变量定义了系统在哪些目录中查找可执行命令。为方便通过命令行使用 Python、pip 和 Robot Framework，建议将以下两个目录添加至 PATH 中：

1. Python 的安装目录
2. 存放 pip 和 robot 等命令的目录

在 Linux 或 macOS 系统上使用 Python 时，Python 及其相关工具通常会自动配置到 PATH 中。若仍需手动更新 PATH，通常需要编辑系统级或用户级的配置文件。具体需要编辑哪个文件以及如何编辑，取决于操作系统类型，请查阅相应系统的文档获取详细信息。

在 Windows 系统中，确保 PATH 正确配置的最简便方法是在运行安装程序时勾选"将 Python 3.x 添加到 PATH"选项。如需手动修改 PATH，请按以下步骤操作：

1. 通过系统设置找到"环境变量"配置项
    * 系统变量影响所有用户（需管理员权限）
    * 用户变量仅影响当前账户（通常修改此项即可）
2. 选择 PATH 变量（可能显示为 Path）并点击"编辑"
    * 若编辑用户变量且 PATH 不存在，请点击"新建"
3. 将以下两个目录添加到 PATH 中：
    * Python 安装目录
    * 安装目录下的 Scripts 子目录
4. 点击"确定"保存修改
5. 需启动新的命令提示符窗口才能使变更生效

### 用 pip 安装 Robot Framework

这里介绍了使用 Python 标准包管理工具 pip 安装 Robot Framework 的方法。若你使用的是 Conda 等其他包管理器，虽可替代使用，但需查阅其官方文档获取具体安装说明。

通常情况下，安装 Python 时会自动附带安装 pip。如未自动安装，请参考该 Python 发行版的官方文档，了解如何单独安装 pip。

#### 运行 pip 命令

通常可以直接运行 pip 命令使用 pip 工具，但在 Linux 系统上，可能需要改用 pip3 或更具体的版本号命令（如 pip3.8）。执行 pip 命令时，系统会优先调用 PATH 环境变量中第一个匹配的可执行文件。若安装了多个 Python 版本，可通过 `python -m pip` 方式明确指定使用特定版本的 pip。

要确认系统是否已安装 pip，可执行 `pip --version` 来检查。

Linux 上的例子：

```bash
$ pip --version
pip 23.2.1 from ... (python 3.10)
$ python3.12 -m pip --version
pip 23.3.1 from ... (python 3.12)
```

Windows 上的例子：

```bash
C:\> pip --version
pip 23.2.1 from ... (python 3.10)
C:\> py -m 3.12 -m pip --version
pip 23.3.2 from ... (python 3.12)
```

在后续章节中，我们将统一使用 `pip` 命令作为示例。根据你的具体环境，可能需要改用前文所述的其他替代方案（如 `pip3` 或 `python -m pip`）。

#### 安装和卸载 Robot Framework

使用 pip 最简单的方式是让其自动从 Python 软件包索引([PyPI](https://pypi.org/project/robotframework))查找并下载安装包，但也可以手动安装从 PyPI 单独下载的软件包。以下是几种最常用的安装方式，更多详细信息和示例请参阅 [pip 官方文档](https://pip.pypa.io/)：

```bash
# 安装最新版 (不会升级)
pip install robotframework

# 升级到最新稳定版本
pip install --upgrade robotframework

# 升级到最新 pre 版本
pip install --upgrade --pre robotframework

# 安装特定版本
pip install robotframework==7.0

# 从下载的包中安装（无网络）
pip install robotframework-7.0-py3-none-any.whl

# 从 GitHub 最新的代码中安装
pip install https://github.com/robotframework/robotframework/archive/master.zip

# 卸载
pip uninstall robotframework
```

### 从源码安装 Robot Framework


另一种安装方式是获取 Robot Framework 源码并通过 setup.py 脚本进行安装。此方法仅建议在无法使用 pip 的情况下采用。

获取源码有两种途径：

1. 从 PyPI 下载源码发行包并解压
2. 克隆 GitHub 仓库后检出所需的发布标签

获取源码后，执行以下命令即可完成安装：

```bash
python setup.py install
```

`setup.py` 脚本支持多个参数配置，例如：

1. 可指定非默认安装路径（无需管理员权限）
2. 可用于生成不同格式的发行包

运行 `python setup.py --help` 命令查看完整参数说明。

### 验证安装

要验证安装的 Robot Framework 版本是否正确，请执行以下命令：

```bash
$ robot --version
Robot Framework 7.0 (Python 3.10.3 on linux)
```

若执行上述命令时提示"command not found"（命令未找到）或"无法识别"等错误，建议首先检查 PATH 环境变量配置是否正确。

若已在多个 Python 版本下安装 Robot Framework，直接运行 robot 命令将执行 PATH 中优先级最高的版本。如需指定特定版本，可使用 `python -m robot` 命令格式：

```bash
$ python3.12 -m robot --version
Robot Framework 7.0 (Python 3.12.1 on linux)

C:\>py -3.11 -m robot --version
Robot Framework 7.0 (Python 3.11.7 on win32)
```

### 虚拟环境

Python 虚拟环境（[virtual environments](https://packaging.python.org/en/latest/guides/installing-using-pip-and-virtual-environments/#creating-a-virtual-environment)）可将 Python 包安装在特定项目或应用的独立隔离位置，而非全部安装到全局共享位置。其主要应用场景包括：

1. 项目依赖隔离
    * 为不同项目创建独立环境安装所需依赖包
    * 避免项目间因需要同一包的不同版本而产生冲突
2. 系统环境保护
    * 防止全局 Python 环境被污染
    * 在 Linux 系统中尤为重要，因其发行版可能依赖系统 Python 环境
    * 随意修改全局安装可能导致严重系统问题