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

为了让 k6 能够解析并执行测试，每个 k6 脚本都遵循相同的结构，围绕几个核心组件展开：

