title: k6 阈值的使用
date: 2025-05-12 10:56:33
series: Grafana k6 登堂入室
summary: 阈值 -> Thresholds

---

阈值（Threshold）是为测试指标定义的通过或失败标准。如果被测系统（SUT）的性能未达到阈值设定的条件，**测试将以失败状态结束**。

测试人员通常通过阈值来编码其服务水平目标（SLO）。例如，可以为以下场景创建阈值：

* 错误率限制：不到 1% 的请求返回错误；
* 响应时间分布：95% 的请求响应时间低于 200 毫秒；
* 长尾优化：99% 的请求响应时间低于 400 毫秒；
* 关键接口保障：特定接口始终在 300 毫秒内响应；
* 自定义指标规则：针对自定义指标设定的任何条件。

## HTTP 错误和响应时间的阈值示例

下面的示例脚本定义了两个阈值：一个阈值用于评估 HTTP 错误率（基于 `http_req_failed` 指标）；另一个阈值用于判断 95% 的响应是否在指定时长内完成（基于 `http_req_duration` 指标）。

```js
import http from 'k6/http';

export const options = {
  thresholds: {
    http_req_failed: ['rate<0.01'],   // http 错误率小于 1%
    http_req_duration: ['p(95)<200'], // 95% 的请求响应时间小于 200ms
  },
};

export default function () {
  http.get('https://quickpizza.grafana.com');
}
```

换句话说，当你定义阈值时，需要为`通过`标准指定一个表达式。如果测试结束时该表达式评估为 `false`，k6 会将整个测试视为`失败`。

执行该脚本后，k6 会输出类似以下内容：

```
  █ THRESHOLDS

    http_req_duration
    ✓ 'p(95)<200' p(95)=148.21ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.05%

  █ TOTAL RESULTS

    HTTP
    http_req_duration..............: avg=151.06ms min=151.06ms med=151.06ms max=151.06ms p(90)=151.06ms p(95)=151.06ms
       { expected_response:true }..: avg=151.06ms min=151.06ms med=151.06ms max=151.06ms p(90)=151.06ms p(95)=151.06ms
    http_req_failed................: 0.00%  ✓ 0 ✗ 1
```

此测试满足了两项阈值，故 k6 判定本次测试通过，且 exit code 为 0。

如果任何阈值未通过，指标名称（`http_req_failed`、`http_req_duration`）旁的绿色对勾 ✓ 将显示为红色叉号 ✗，同时 k6 的 exit code 是非 0。

## 阈值语法

按照以下步骤操作使用阈值：

首先在 `options` 对象的 `thresholds` 属性中，使用要设置阈值的指标名称作为键名

```js
export const options = {
  thresholds: {
    /* ... */
  },
};
```

然后定义至少一个阈值表达式。表达式有以下两种形式：

1. 简短格式将所有阈值表达式以字符串形式放入数组中。
2. 详细格式则将每个阈值封装为独立对象，并包含可中止测试的额外属性。

```js
export const options = {
  thresholds: {
    //short format
    METRIC_NAME1: ['THRESHOLD_EXPRESSION', `...`],
    //long format
    METRIC_NAME2: [
      {
        threshold: 'THRESHOLD_EXPRESSION',
        abortOnFail: true, // boolean
        delayAbortEval: '10s', // string
      },
    ], // full format
  },
};
```

P.S. 请注意，`METRIC_NAME1` 和 `THRESHOLD_EXPRESSION` 均为占位符，实际使用时需替换为具体的指标名称和阈值表达式。

该声明用于配置 `metric_name1` 和 `metric_name2` 这两个指标的阈值。脚本将通过评估 `'threshold_expression'` 表达式来判断阈值是否通过。

### 阈值表达式语法

阈值表达式结果为布尔值 `true` 或 `false`。阈值表达式必须是以下的形式：

```js
<aggregation_method> <operator> <value>
```

比如：

* `avg < 200` // 平均耗时必须小于 200 毫秒
* `count >= 500` // 请求次数必须大于等于 500 次
* `p(90) < 300` // 90% 的样本数据必须低于 300

### 按类型划分的聚合方法

k6 根据指标类型进行数据聚合，这些聚合方法将构成阈值表达式的一部分。

| **指标类型**	 | **聚合方法** |
|-------|-------|
| Counter  |  `count` 和 `rate`  |
| Gauge  |   `value`	|
| Rate  |   `rate`	|
| Trend  |  `avg`, `min`, `max`, `med` 和 `p(N)`，其中 `N` 指定百分位阈值（数值范围为 0.0 至 100）。例如 `p(99.99)` 表示第 99.99 百分位数。所有数值均以毫秒为单位。|

这个示例脚本使用了所有不同类型的指标，并为每种指标设置了不同的阈值类型：

```js
import http from 'k6/http';
import { Trend, Rate, Counter, Gauge } from 'k6/metrics';
import { sleep } from 'k6';

export const TrendRTT = new Trend('RTT');
export const RateContentOK = new Rate('ContentOK');
export const GaugeContentSize = new Gauge('ContentSize');
export const CounterErrors = new Counter('Errors');
export const options = {
  thresholds: {
    // Count: 错误内容出现次数不能超过 99 次
    Errors: ['count<100'],
    // Gauge: 返回内容必须小于 4000 字节
    ContentSize: ['value<4000'],
    // Rate: 内容必须正常（OK）的次数不低于 95 次
    ContentOK: ['rate>0.95'],
    // Trend: 百分位数、平均值、中位数及最小值均需保持在指定的毫秒级范围内
    RTT: ['p(99)<300', 'p(70)<250', 'avg<200', 'med<150', 'min<100'],
  },
};

export default function () {
  const res = http.get('https://quickpizza.grafana.com/api/json?name=Bert');
  const contentOK = res.json('name') === 'Bert';

  TrendRTT.add(res.timings.duration);
  RateContentOK.add(contentOK);
  GaugeContentSize.add(res.body.length);
  CounterErrors.add(!contentOK);

  sleep(1);
}
```

请勿通过重复相同对象键名的方式为同一指标设置多个阈值。由于阈值是作为 JavaScript 对象的属性来定义的，因此不能使用相同的属性名来指定多个阈值。

错误示范：

```js
export const options = {
  thresholds: {
    metric_name: ['count<100'],
    metric_name: ['rate<50'],
  },
};
```

后面的配置将被忽略。如需为同一指标设置多个阈值，请改用数组形式指定相同键名对应的值。

## 可直接复制粘贴的阈值配置示例

使用内置指标是快速设置阈值的首选方式。以下提供几个可直接复制的配置示例，可立即投入使用。

### 在指定时间内完成特定百分比的请求

```js
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  thresholds: {
    // 90% 的请求需在 400ms 内完成
    http_req_duration: ['p(90) < 400'],
  },
};

export default function () {
  http.get('https://quickpizza.grafana.com');
  sleep(1);
}
```

### 错误率需低于 1%

```js
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  thresholds: {
    // 在整个测试执行期间，错误率必须始终低于 1%
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  http.get('https://quickpizza.grafana.com');
  sleep(1);
}
```

### 单个指标的多个阈值

你也可以为单个指标设置多个阈值。该阈值针对不同请求百分位设有不同的耗时要求。

```js
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  thresholds: {
    // 90% 的请求必须在 400ms 内完成，95% 的请求需在 800ms 内完成，99.9% 的请求则应在 2s 内完成。
    http_req_duration: ['p(90) < 400', 'p(95) < 800', 'p(99.9) < 2000'],
  },
};

export default function () {
  const res1 = http.get('https://quickpizza.grafana.com');
  sleep(1);
}
```

### 分组耗时阈值

你可以按组设置阈值。此代码中包含针对单个请求和批量请求的分组设置，每个组可配置不同的阈值标准。

```js
import http from 'k6/http';
import { group, sleep } from 'k6';

export const options = {
  thresholds: {
    'group_duration{group:::individualRequests}': ['avg < 400'],
    'group_duration{group:::batchRequests}': ['avg < 200'],
  },
  vus: 1,
  duration: '10s',
};

export default function () {
  group('individualRequests', function () {
    http.get('https://quickpizza.grafana.com/api/json?letter=a');
    http.get('https://quickpizza.grafana.com/api/json?letter=b');
    http.get('https://quickpizza.grafana.com/api/json?letter=c');
  });

  group('batchRequests', function () {
    http.batch([
      ['GET', 'https://quickpizza.grafana.com/api/json?letter=a'],
      ['GET', 'https://quickpizza.grafana.com/api/json?letter=b'],
      ['GET', 'https://quickpizza.grafana.com/api/json?letter=c'],
    ]);
  });

  sleep(1);
}
```

## 为特定标签设置阈值

为单个 URL 或特定标签设置阈值通常非常实用。在 k6 中，带有标签的请求会生成子指标，可直接用于阈值配置：

```js
export const options = {
  thresholds: {
    'metric_name{tag_name:tag_value}': ['threshold_expression'],
  },
};
```

这里有个完整的例子：

```js
import http from 'k6/http';
import { sleep } from 'k6';
import { Rate } from 'k6/metrics';

export const options = {
  thresholds: {
    'http_req_duration{type:API}': ['p(95)<500'], // 仅针对 API 请求的阈值
    'http_req_duration{type:staticContent}': ['p(95)<200'], // 仅针对静态内容的阈值
  },
};

export default function () {
  const res1 = http.get('https://quickpizza.grafana.com/api/headers', {
    tags: { type: 'API' },
  });
  const res2 = http.get('https://quickpizza.grafana.com/api/json', {
    tags: { type: 'API' },
  });

  const responses = http.batch([
    [
      'GET',
      'https://quickpizza.grafana.com/favicon.ico',
      null,
      { tags: { type: 'staticContent' } },
    ],
    ['GET', 'https://quickpizza.grafana.com/admin', null, { tags: { type: 'staticContent' } }],
  ]);

  sleep(1);
}
```

## 当超过阈值时中止测试

若需在超过阈值时立即中止测试，可将 `abortOnFail` 属性设为 `true`。启用该参数后，一旦阈值被突破，测试将立即终止。

有时候，测试可能在初期就触发阈值导致中止，而此时尚未生成足够数据。为避免该情况，可通过 `delayAbortEval` 参数延迟中止判定。如本脚本所示，将 `abortOnFail` 延迟 10 秒生效 —— 即测试仅在持续 10 秒仍无法满足 `p(99) < 10` 阈值时才会中止。

```js
export const options = {
  thresholds: {
    metric_name: [
      {
        threshold: 'p(99) < 10', // string
        abortOnFail: true, // boolean
        delayAbortEval: '10s', // string
        /*...*/
      },
    ],
  },
};
```
各字段定义如下：

| **名称**	 | **类型** | **描述** |
|-------|-------|-------|
| threshold  |   string |  阈值表达式字符串，用于指定需要评估的阈值条件  |
| abortOnFail  |   boolean	|  当测试未完成时若阈值评估为 false，是否中止测试  |
| delayAbortEval  |   string	|  若需延迟阈值评估以收集足够的指标样本，可使用相对时间字符串（如 `10s`、`1m` 等）指定延迟时长  |

例子如下：

```js
import http from 'k6/http';

export const options = {
  vus: 30,
  duration: '2m',
  thresholds: {
    http_req_duration: [{ threshold: 'p(99) < 10', abortOnFail: true }],
  },
};

export default function () {
  http.get('https://quickpizza.grafana.com');
}
```

## 使用检查使负载测试失败

[检查](https://www.lyyyuna.com/2025/04/10/grafana-k6-4/)适用于将断言规则代码化，但与阈值不同，检查项不会影响 k6 的退出状态。

若仅依赖检查项验证系统行为，则无法基于检查结果使整个测试运行失败。通常最佳实践是结合使用检查项与阈值，从而兼得二者优势：

```js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 50,
  duration: '10s',
  thresholds: {
    // 检查的成功率应高于 90%。
    checks: ['rate>0.9'],
  },
};

export default function () {
  const res = http.get('https://quickpizza.grafana.com/api/status/500');

  check(res, {
    'status is 500': (r) => r.status == 500,
  });

  sleep(1);
}
```

在此示例中，阈值基于检查项指标（checks metric）配置，要求检查项成功率必须高于 90%。

此外，你还可以为检查项添加标签，以便针对特定检查项或检查项组设置独立阈值。例如：

```js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 50,
  duration: '10s',
  thresholds: {
    'checks{myTag:hola}': ['rate>0.9'],
  },
};

export default function () {
  let res;

  res = http.get('https://quickpizza.grafana.com/api/status/500');
  check(res, {
    'status is 500': (r) => r.status == 500,
  });

  res = http.get('https://quickpizza.grafana.com/api/status/200');
  check(
    res,
    {
      'status is 200': (r) => r.status == 200,
    },
    { myTag: 'hola' }
  );

  sleep(1);
}
```