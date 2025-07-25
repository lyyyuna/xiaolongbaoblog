title: k6 测试生命周期
date: 2025-06-18 10:56:33
series: Grafana k6 登堂入室

---

## 概览

在 k6 测试的生命周期中，脚本总是按照以下固定顺序依次执行各个阶段：

1. `init` 上下文中的代码准备脚本、加载文件、导入模块并定义测试生命周期函数。**必须阶段**
2. `setup` 函数执行阶段，设置测试环境并生成测试数据。**可选阶段**
3. VU 代码执行阶段，在 `default` 或 `scenario` 函数中运行，根据测试配置执行指定时长和次数。**必选阶段**
4. `teardown` 函数执行阶段，进行数据后处理并关闭测试环境。**可选阶段**

```js
// 1. init code

export function setup() {
  // 2. setup code
}

export default function (data) {
  // 3. VU code
}

export function teardown(data) {
  // 4. teardown code
}
```

| **测试阶段**	 | **目的** | **示例** | **调用规则** |
|-------|-------|-------|-------|
| 1. init  |   加载本地文件、导入模块、声明生命周期函数 |  打开 JSON 文件、导入模块  |  每个 VU 执行一次  |
| 2. Setup  |   准备测试数据，在 VU 间共享数据 |  调用 API 启动测试环境  |  整个测试执行一次  |
| 3. VU 代码  |   执行测试函数（通常为 `default` 函数） |  发送 HTTPS 请求、验证响应  |  每次迭代执行一次，次数由测试配置决定  |
| 4. Teardown  |   处理 Setup 阶段的结果数据，停止测试环境 |  验证 Setup 结果、发送测试完成 Webhook 通知  |  整个测试执行一次  |

## init 阶段

`init` 阶段是必选的。在测试运行前，k6 需要初始化测试环境。为准备测试，`init` 上下文中的代码会为每个 VU 执行一次。

在 `init` 阶段可能进行的操作包括：

* 导入模块
* 从本地文件系统加载文件
* 为所有测试选项进行配置
* 为 VU、`setup` 和 `teardown` 阶段定义生命周期函数（也包括自定义函数或 `handleSummary()` 函数）

所有未包含在生命周期函数内的代码都属于 `init` 上下文。`init` 上下文中的代码总是最先执行。

```js
// init context: importing modules
import http from 'k6/http';
import { Trend } from 'k6/metrics';

// init context: define k6 options
export const options = {
  vus: 10,
  duration: '30s',
};

// init context: global variables
const customTrend = new Trend('oneCustomMetric');

// init context: define custom function
function myCustomFunction() {
  // ...
}
```

将 `init` 阶段 与 VU 阶段分离，可以避免在 VU 代码中执行无关的计算，从而提升 k6 的性能并确保测试结果更加可靠。不过，`init` 代码有一个限制：它不能发起 HTTP 请求。这一限制是为了确保 `init` 阶段在不同测试中具有可复现性（因为协议请求的响应是动态且不可预测的）。

## VU 阶段

脚本必须至少包含一个场景函数（scenario function），用于定义虚拟用户（VU）的执行逻辑。该函数内部的代码即为 VU 代码。通常情况下，VU 代码位于 `default` 函数中，但也可以定义在场景配置指定的函数内（具体示例请参阅后续章节）。

```js
export default function () {
  // do things here...
}
```

VU 代码会在整个测试期间循环执行。它可以发起 HTTP 请求、输出指标数据，基本上能完成负载测试所需的所有操作——唯独那些属于 `init` 上下文的操作除外。

具体限制包括：

* 无法读取本地文件系统的文件
* 无法导入其他模块

这些功能必须由 `init` 代码来实现，而非 VU 代码。

### 默认函数的生命周期

1. 顺序执行机制。VU 会从头到尾顺序执行 `default()` 函数。当执行到函数末尾时，会自动跳转回起始位置重新执行，形成循环测试。
2. 重启重置机制。每次循环开始时，k6 会对 VU 执行重置操作：
    1. 自动清除所有 cookies
    2. 根据测试配置决定是否断开TCP连接（连接保持行为可通过测试参数配置）

## setup 和 teardown 阶段

与 `default` 函数类似，`setup` 和 `teardown` 也必须是导出函数。但与 `default` 函数不同的是，k6 在整个测试过程中只会调用 `setup` 和 `teardown` 各一次。具体调用时机如下：

1. `setup`：在测试开始时调用，位于 `init` 阶段之后、VU 阶段之前
2. `teardown`：在测试结束时调用，位于 VU 阶段（即 `default` 函数执行）之后

与 `init` 阶段不同，在 `setup` 和 `teardown` 阶段可以调用完整的 k6 API。例如，你可以执行以下操作：

* 发起 HTTP 请求
* 使用所有 k6 模块功能
* 访问完整的运行时指标系统

```js
import http from 'k6/http';

export function setup() {
  const res = http.get('https://quickpizza.grafana.com/api/json');
  return { data: res.json() };
}

export function teardown(data) {
  console.log(JSON.stringify(data));
}

export default function (data) {
  console.log(JSON.stringify(data));
}
```

### 忽略 setup 和 teardown 的执行

你可以通过命令行选项 `--no-setup` 和 `--no-teardown` 来跳过 `setup` 和 `teardown` 阶段的执行。

```bash
k6 run --no-setup --no-teardown ...
```

### 在 default 和 teardown 中使用 setup 中的数据

```js
// 1. init code

export function setup() {
  // 2. setup code
}

export default function (data) {
  // 3. VU code
}

export function teardown(data) {
  // 4. teardown code
}
```

你可能已经注意到，`default()` 和 `teardown()` 函数的签名都接收一个参数（本文中称为 `data`）。

以下示例演示了如何将数据从 `setup` 代码传递到 VU 和 `teardown` 阶段：

```js
export function setup() {
  return { v: 1 };
}

export default function (data) {
  console.log(JSON.stringify(data));
}

export function teardown(data) {
  if (data.v != 1) {
    throw new Error('incorrect data: ' + JSON.stringify(data));
  }
}
```

例如，利用 `setup()` 函数返回的数据，你可以实现以下功能：

* 统一数据分发：为每个 VU（虚拟用户）提供数据的相同副本
* 结果后处理：在 `teardown` 代码中对数据进行最终处理

但需注意以下限制事项：

* 数据类型限制
   - 仅支持传递 JSON 格式的数据（不可传递函数）
   - 无法在 `default()` 中修改数据后传递给 `teardown()`
* 内存消耗警告
   - 当 `setup()` 返回大数据量时会显著增加内存占用
* 数据隔离特性
   - 每个阶段和每个 VU 访问的都是 `setup()` 返回数据的独立副本
   - 修改副本不会影响其他 VU 或阶段的数据

在分布式测试场景中，若要在所有虚拟用户(VU)之间传递可变数据并最终移交至 teardown 阶段，其实现将异常复杂且计算资源消耗巨大。这种设计会直接违背 k6 的核心设计原则：确保同一测试脚本可跨多种执行模式运行。

## 额外的生命周期

k6 还提供了其他几种使用生命周期函数的方式：

1. `handleSummary()`，若需生成自定义测试报告，k6 会在测试完全结束时额外调用此生命周期函数。
2. 场景函数 (Scenario Functions)，除了默认的 `default` 函数外，你也可以在场景函数中运行 VU 代码。

```js
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  scenarios: {
    my_web_test: {
      // the function this scenario will execute
      exec: 'webtest',
      executor: 'constant-vus',
      vus: 50,
      duration: '1m',
    },
  },
};

export function webtest() {
  http.get('https://test.k6.io/contacts.php');
  sleep(Math.random() * 2);
}
```