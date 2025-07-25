title: k6 检查的使用
date: 2025-04-10 10:56:33
series: Grafana k6 登堂入室
summary: 检查 -> Checks

---

## 检查

`检查`用于验证测试用例中布尔条件是否成立。测试人员通过检查验证系统是否返回预期内容。例如，某个检查可以确认 POST 请求的响应状态码是否为 201，或者响应体的大小是否符合预期。

`检查`和许多测试框架中的`断言`（assert）类似，但在 k6 中，失败的检查不会导致测试中断。相反，k6 会在测试持续运行过程中记录检查的失败率。

### 检查 HTTP 响应码

检查功能非常适合针对HTTP请求和响应制定断言规范。例如，以下代码片段可确保HTTP响应码为 200：

```js
import { check } from 'k6';
import http from 'k6/http';

export default function () {
  const res = http.get('http://test.k6.io/');
  check(res, {
    'is status 200': (r) => r.status === 200,
  });
}
```

### 检查 HTTP 响应体

有时，即使HTTP 200响应中也可能包含错误信息。在这种情况下，建议添加一个检查来验证响应体，例如：

```js
import { check } from 'k6';
import http from 'k6/http';

export default function () {
  const res = http.get('http://test.k6.io/');
  check(res, {
    'verify homepage text': (r) =>
      r.body.includes('Collection of simple web-pages suitable for load testing'),
  });
}
```

### 检查 HTTP 响应大小

要验证响应体的大小，可以使用如下检查：


```js
import { check } from 'k6';
import http from 'k6/http';

export default function () {
  const res = http.get('http://test.k6.io/');
  check(res, {
    'body size is 11,105 bytes': (r) => r.body.length == 11105,
  });
}
```

### 查看检查的通过率

当脚本包含检查时，汇总报告会显示测试的检查中有多少通过：

```
$ k6 run script.js

  ...
    ✓ is status 200

  ...
  checks.........................: 100.00% ✓ 1        ✗ 0
  data_received..................: 11 kB   12 kB/s
```

### 添加多个检查

你还可以在单个 `check()` 语句中添加多个检查条件：

```js
import { check } from 'k6';
import http from 'k6/http';

export default function () {
  const res = http.get('http://test.k6.io/');
  check(res, {
    'is status 200': (r) => r.status === 200,
    'body size is 11,105 bytes': (r) => r.body.length == 11105,
  });
}
```

当执行此测试时，输出结果会大致如下：

```js
$ k6 run checks.js

  ...
    ✓ is status 200
    ✓ body size is 11,105 bytes

  ...
  checks.........................: 100.00% ✓ 2        ✗ 0
  data_received..................: 11 kB   20 kB/s
```

当检查失败时，脚本将继续成功执行，并不会返回‘失败’的退出状态。若需根据检查结果让整个测试失败，必须将检查与**阈值**结合使用。这一功能在特定场景下特别有用，例如将 k6 集成到持续集成（CI）管道中，或在安排性能测试时接收告警。
