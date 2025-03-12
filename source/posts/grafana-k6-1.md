title: Grafana k6 入门示例
date: 2025-03-12 10:56:33
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

可以在 [GitHub Releases](https://github.com/grafana/k6/releases) 页面中看到预编译好的二进制文件，将对应系统的 k6 放入 `PATH` 目录中即可使用。

### 源码安装

源码安装其实是我推荐的方式，我们后续扩展 k6 时依赖于此。后续文章将详细介绍，这里按下不表。

## 写下你的第一个测试脚本

k6 是一款性能测试工具。它能帮助开发者模拟真实的用户行为，并测试系统在这些行为下的表现。通过使用 k6 编写测试脚本，可以在问题（如响应缓慢或系统故障）在生产环境中发生之前，提前发现潜在问题。

大家的测试目标各不相同：性能、可靠性或可扩展性。基于此，脚本需要不同的配置，例如模拟大量用户或长时间运行。

为了确保 k6 的通用性和灵活性，所有的测试脚本都遵循统一的结构设计。这种一致的结构不仅使脚本易于理解和维护，还能让开发者快速上手，轻松编写出高效的性能测试脚本，它有四个主要模块：

1. 默认函数：这是测试逻辑的核心部分。它定义了测试的内容以及测试在执行期间的行为。该函数需要在脚本中作为默认函数导出。
2. 导入模块：可以导入额外的 k6 模块或 JavaScript 库（jslibs）来扩展脚本的功能，例如发送 HTTP 请求或模拟浏览器交互。
3. 选项（这个模块本身也是可选的）：允许配置测试的执行方式，例如定义虚拟用户的数量、测试持续时间或设置性能阈值。
4. 生命周期操作（可选）：这些代码可以作为预定义函数或在特定的代码范围内执行，在测试执行的不同阶段运行（例如测试开始前从文件中解析数据或从 Amazon S3 下载文件）。

现在有这样一个测试场景：向一个 URL 发送 10 次 `GET` HTTP 请求，并在每次请求之间等待 1 秒钟。

首先创建一个名为 `script.js` 的 JavaScript 文件。

为了发送 HTTP 请求，导入 `k6/http` 模块的 `http` 函数，为了模拟每个请求之间的延迟，导入 `k6` 模块的 `sleep` 函数：

```js
// script.js
import http from 'k6/http';
import { sleep } from 'k6';
```

我们不会在脚本中写一个 for 循环发送 10 次，来实现发送 10 次 `GET` HTTP 请求。而应该定义一个 `options` 代码块，将迭代次数设置为 10，让 k6 来自动控制循环次数：

```js
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  // 设置测试的迭代次数
  iterations: 10,
};
```

默认导出的函数就是测试脚本的入口点。本测试向一个 URL 发起 GET 请求，并在请求之间引入 1 秒的延迟：

```js
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  iterations: 10,
};

export default function () {
  http.get('https://www.baidu.com');
  sleep(1);
}
```

## 运行 k6

用以下命令运行上一节的脚本 `script.js`，：

```bash
$ k6 run script.js


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 

     execution: local
        script: script.js
        output: -

     scenarios: (100.00%) 1 scenario, 1 max VUs, 10m30s max duration (incl. graceful stop):
              * default: 10 iterations shared among 1 VUs (maxDuration: 10m0s, gracefulStop: 30s)


     data_received..................: 20 kB  1.8 kB/s
     data_sent......................: 1.5 kB 138 B/s
     http_req_blocked...............: avg=16.08ms  min=10.45µs med=13.3µs   max=160.72ms p(90)=16.09ms  p(95)=88.4ms  
     http_req_connecting............: avg=4.42ms   min=0s      med=0s       max=44.22ms  p(90)=4.42ms   p(95)=24.32ms 
     http_req_duration..............: avg=51.54ms  min=43.7ms  med=48.77ms  max=77.04ms  p(90)=57.23ms  p(95)=67.14ms 
       { expected_response:true }...: avg=51.54ms  min=43.7ms  med=48.77ms  max=77.04ms  p(90)=57.23ms  p(95)=67.14ms 
     http_req_failed................: 0.00%  0 out of 10
     http_req_receiving.............: avg=179.21µs min=82.12µs med=170.89µs max=252.88µs p(90)=227.39µs p(95)=240.13µs
     http_req_sending...............: avg=55.41µs  min=28.53µs med=40.17µs  max=185.76µs p(90)=83.54µs  p(95)=134.65µs
     http_req_tls_handshaking.......: avg=10.22ms  min=0s      med=0s       max=102.23ms p(90)=10.22ms  p(95)=56.22ms 
     http_req_waiting...............: avg=51.3ms   min=43.4ms  med=48.56ms  max=76.85ms  p(90)=57.06ms  p(95)=66.95ms 
     http_reqs......................: 10     0.935633/s
     iteration_duration.............: avg=1.06s    min=1.04s   med=1.04s    max=1.21s    p(90)=1.09s    p(95)=1.15s   
     iterations.....................: 10     0.935633/s
     vus............................: 1      min=1       max=1
     vus_max........................: 1      min=1       max=1


running (00m10.7s), 0/1 VUs, 10 complete and 0 interrupted iterations
default ✓ [======================================] 1 VUs  00m10.7s/10m0s  10/10 shared iters
```

我们会得到一个聚合过的测试报告，包含各种 http 参数的：

1. 中位值和平均值
2. 最小值和最大值
3. P90、P95 值

除了这些统计指标外，我们还注意到，这次测试持续了 10秒，和脚本中的“10 个迭代，每次持续 1秒”能对应上。JMeter 有虚拟用户数（即用户并发数）的概念，在 k6 中等价的是 `VUs`，设置 VUs 就是设置并发数量：

```bash
$ k6 run script.js --vus 10

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 

     execution: local
        script: script.js
        output: -

     scenarios: (100.00%) 1 scenario, 10 max VUs, 10m30s max duration (incl. graceful stop):
              * default: 10 iterations shared among 10 VUs (maxDuration: 10m0s, gracefulStop: 30s)

...

running (00m01.6s), 00/10 VUs, 10 complete and 0 interrupted iterations
default ✓ [======================================] 10 VUs  00m01.6s/10m0s  10/10 shared iters
```

这次测试报告显示，并发为 10，但只花了 1s 就结束了，这意味着所有并发会共同消耗迭代次数。

每次在启动命令中指定 VUs 略有不便，可以将其固化在脚本的选项中：

```js
import http from 'k6/http';
import { sleep } from 'k6';
export const options = {
  vus: 10,
  iterations: 10,
};
export default function () {
  http.get('http://test.k6.io');
  sleep(1);
}
```

除了用迭代次数间接控制性能测试的时长，还可以直接指定测试时长：

```js
export const options = {
  vus: 10,
  durations: '10s',
};
```

你可以通过配置 `options.stages` 属性来实现虚拟用户（VUs）数量的动态增减（ramping）。这种功能非常适合模拟真实场景中的用户行为，例如逐步增加负载以测试系统的性能极限，或者在测试结束后逐步减少负载。

以下是一个典型的 ramping 配置示例，模拟负载逐步增加、保持峰值、然后逐步减少的过程：

```js
import http from 'k6/http';

export const options = {
  stages: [
    // 第一阶段：30 秒内逐步增加到 50 个 VUs
    { duration: '30s', target: 50 },
    // 第二阶段：保持 50 个 VUs 持续运行 1 分钟
    { duration: '1m', target: 50 },
    // 第三阶段：30 秒内逐步减少到 0 个 VUs
    { duration: '30s', target: 0 },
  ],
};

export default function () {
  http.get('https://test.k6.io');
}
``` 

运行结果为：
1. 第一阶段：在 30 秒内，虚拟用户数从 0 逐步增加到 50。
2. 第二阶段：在 1 分钟内，保持 50 个虚拟用户持续运行。
3. 第三阶段：在 30 秒内，虚拟用户数从 50 逐步减少到 0。

## 结语

在本文中，我们简要浏览了 k6 的基本功能，并进行了初步尝试。可以发现，k6 使用起来非常简单，对于具备编程能力的开发者来说尤为友好。在接下来的文章中，我们将深入探讨 k6 的各个细节，帮助大家更好地掌握这一强大工具。
