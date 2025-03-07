title: Grafana k6 入门示例
date: 2025-03-03 10:56:33
series: Grafana k6 登堂入室

---

## 安装

k6 提供了适用于 Linux、Mac 和 Windows 的安装包。此外也可以使用 Docker 容器或独立的二进制文件，甚至是源码安装。

### Linux

Debian/Ubuntu

```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

Fedora/Centos

```bash
sudo dnf install https://dl.k6.io/rpm/repo.rpm
sudo dnf install k6
```

### MacOS

```bash
brew install k6
```

### Windows

```bash
choco install k6
```

### Docker

```bash
docker pull grafana/k6
```

### 二进制安装

可以在 [GitHub Releases](https://github.com/grafana/k6/releases) 页面中看到预编译好的二进制文件，讲对应系统的 k6 放入 `PATH` 目录中即可使用。

### 源码安装

源码安装其实我推荐的方式，也是我们扩展 k6 能力时唯一能使用的方式。后续文章将详细介绍这一种方式。这里按下不表。

## 写下你的第一个测试脚本

k6 是一款可靠性测试工具。它能帮助开发者模拟真实的用户行为，并测试系统在这些行为下的表现。通过使用 k6 编写测试脚本，可以在问题（如响应缓慢或系统故障）在生产环境中发生之前，提前发现潜在问题。

大家的测试目标各不相同：性能、可靠性或可扩展性。基于此，脚本需要不同的配置，例如模拟大量用户或长时间运行测试。

### k6 测试的基本结构

为了让 k6 能够解析并执行测试，每个 k6 脚本都遵循相同的结构，有四个主要模块：

1. 默认函数：这是测试逻辑的核心部分。它定义了测试的内容以及测试在执行期间的行为。该函数需要在脚本中作为默认函数导出。
2. 导入模块：可以导入额外的 k6 模块或 JavaScript 库（jslibs）来扩展脚本的功能，例如发送 HTTP 请求或模拟浏览器交互。上一篇文章提到，k6 的 JavaScript 运行时是 Go 语言开发的，因此，不一定兼容所有的 npm 模块。
3. 选项（这个模块本身也是可选的）：允许配置测试的执行方式，例如定义虚拟用户的数量、测试持续时间或设置性能阈值。
4. 生命周期操作（可选）：由于测试可能需要在执行测试逻辑之前和之后运行某些代码（例如从文件中解析数据或从 Amazon S3 下载文件），生命周期操作允许你编写代码，这些代码可以作为预定义函数或在特定的代码范围内执行，并在测试执行的不同阶段运行。

#### 写下你的第一个测试脚本

假设有这样一个测试场景：向一个 URL 发送 10 次 `GET` HTTP 请求，并在每次请求之间等待 1 秒钟。

首先创建一个名为 `my-first-test.js` 的 JavaScript 文件。

为了发送 HTTP 请求，导入 `k6/http` 模块的 `http` 函数，为了模拟每个请求之间的延迟，导入 `k6` 模块的 `sleep` 函数：

```js
// my-first-test.js
import http from 'k6/http';
import { sleep } from 'k6';
```

我们不会在脚本中写一个 for 循环发送 10 次，来实现发送 10 次 `GET` HTTP 请求。而是定义一个 `options` 代码块，将迭代次数设置为 10，让 k6 来帮我们自动控制循环：

```js
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  // 设置测试的迭代次数
  iterations: 10,
};
```