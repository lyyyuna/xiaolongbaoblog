title: k6 选项的使用
date: 2025-06-10 10:56:33
series: Grafana k6 登堂入室
summary: 选项 -> Options

---

选项（Options）用于配置测试运行行为。比如你可以通过选项定义测试标签、阈值、用户代理以及虚拟用户数和迭代次数。

## 如何使用

k6 提供多种设置选项的方式：

1. 通过命令行标志设置
2. 通过环境变量设置
3. 在脚本的 `options` 对象中设置
4. 通过配置文件设置

具体在何处设置测试选项通常取决于使用场景，还可以在测试运行时动态获取这些选项值。

## 选项的顺序

你可以在多个位置设置选项。当配置出现冲突时，k6 将按照优先级顺序决定哪个起作用。

```
默认值 -> 配置文件 -> 测试脚本设置 -> 环境变量 -> 命令行参数
```

1. 首先采用选项的默认值。
2. 其次使用通过 --config 参数指定的配置文件中的设置。
3. 然后读取脚本中定义的值（如已设置）。
4. 接着采用环境变量中的值（如已设置）。
5. 最终以命令行参数（CLI flag）的设置为最高优先级。

## 在何处设置选项

选项的设置方式有时取决于个人偏好，有时则需根据测试场景选择最合理的位置。

* **脚本内选项：便于版本控制与测试整洁化**

通常建议将选项定义在脚本的 options 对象中：

1. 自动纳入版本控制系统
2. 便于测试脚本复用
3. 支持模块化脚本开发

* **CLI参数：快速动态配置**

命令行参数适合临时测试场景：

1. 可即时覆盖脚本中的配置（遵循优先级规则）
2. 示例：当脚本预设60秒测试时长时，通过 `--duration 30s` 参数即可临时改为30秒测试，其他配置保持不变

* **环境变量：与构建流程集成**

适用于从以下系统获取配置：

1. Docker容器定义
2. CI/CD系统界面
3. 密钥管理系统，典型用例：如 block hostnames 这类选项就非常适合通过环境变量配置

## 示例

### 在脚本中设置选项

```js
import http from 'k6/http';

export const options = {
  hosts: { 'test.k6.io': '1.2.3.4' },
  stages: [
    { duration: '1m', target: 10 },
    { duration: '1m', target: 20 },
    { duration: '1m', target: 0 },
  ],
  thresholds: { http_req_duration: ['avg<100', 'p(95)<200'] },
  noConnectionReuse: true,
  userAgent: 'MyK6UserAgentString/1.0',
};

export default function () {
  http.get('http://test.k6.io/');
}
```

### 在环境变量中设置选项

你也可以通过环境变量和命令行参数来配置上述示例中的选项：


```bash
K6_NO_CONNECTION_REUSE=true K6_USER_AGENT="MyK6UserAgentString/1.0" k6 run script.js

$ k6 run --no-connection-reuse --user-agent "MyK6UserAgentString/1.0" script.js
```

### 从 k6 变量中设置选项

通过 `--env` 命令行标志，你可以直接在 CLI 中定义 k6 变量，然后在脚本文件中使用这些变量来动态设置选项值。

例如，你可以通过以下方式为用户代理(user agent)定义变量：

```bash
$ k6 run script.js --env MY_USER_AGENT="hello"
```

随后，脚本便可根据该变量值动态设置 `userAgent` 选项，实现快速配置。

```js
import http from 'k6/http';

export const options = {
  userAgent: __ENV.MY_USER_AGENT,
};

export default function () {
  http.get('http://test.k6.io/');
}
```

### 通过 --config 设置选项

k6 提供了以下配置文件管理方式：

1. 可直接编辑默认配置文件
2. 或创建新配置文件并通过 CLI 参数指定路径

通过 `--config` 参数指定配置文件路径，例如：

```bash
$ k6 run --config options.json script.js
```

执行此命令时，k6 将按照 `options.json` 文件中的值来设置测试选项。

```json
{
  "hosts": {
    "test.k6.io": "1.2.3.4"
  },
  "stages": [
    {
      "duration": "1m",
      "target": 10
    },
    {
      "duration": "1m",
      "target": 30
    },
    {
      "duration": "1m",
      "target": 0
    }
  ],
  "thresholds": {
    "http_req_duration": ["avg<100", "p(95)<200"]
  },
  "noConnectionReuse": true,
  "userAgent": "MyK6UserAgentString/1.0"
}
```

若需实现配置与逻辑分离，你还可以在脚本文件中使用 `JSON.parse()` 方法：

```js
const testConfig = JSON.parse(open('./config/test.json'));

export const options = testConfig;
```

## 从脚本中获取选项值

k6 的 `k6/execution` API 提供了 `test.options` 对象。通过该对象，你可以在测试运行时访问经过整合处理后的最终选项值。

该功能常用于记录标签值，但实际用途远不止于此。例如，以下脚本演示了如何获取当前测试阶段的配置值：

```js
import exec from 'k6/execution';

export const options = {
  stages: [
    { duration: '5s', target: 100 },
    { duration: '5s', target: 50 },
  ],
};

export default function () {
  console.log(exec.test.options.scenarios.default.stages[0].target); // 100
}
```