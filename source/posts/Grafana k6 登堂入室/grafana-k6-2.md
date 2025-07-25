title: 在 k6 中发送 HTTP 请求
date: 2025-04-08 10:56:33
series: Grafana k6 登堂入室

---

## 发送 HTTP 请求

`k6/http` 模块可处理各种类型的 HTTP 请求和方法。比如最常见的 GET 请求可以通过下面的方式发送：

```js
import http from 'k6/http';

export default function () {
  http.get('http://test.k6.io');
}
```

比如发送 POST 请求，带上请求体，并加上特定的 HTTP header：

```js
import http from 'k6/http';

export default function () {
  const url = 'http://test.k6.io/login';
  const payload = JSON.stringify({
    email: 'aaa',
    password: 'bbb',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'abced',
    },
  };

  http.post(url, payload, params);
}
```

## 跟随重定向

默认情况下，k6 会在停止并返回最后一个响应之前自动跟随一定数量的重定向。你可以通过以下方式自定义设置：

* 使用 `maxRedirects` 选项全局设置重定向次数上限。
* 使用 `Params.redirects` 属性为特定请求设置重定向次数，该属性会覆盖 `maxRedirects` 选项。

## HTTP 请求的标签

k6 会自动给 HTTP 请求添加标签。你可以使用这些标签来过滤测试结果，并对分析内容进行分类整理：

1. `expected_response`，默认情况下，状态响应码在 200～399 之间的话，为 `true`。默认行为可以通过 `setResponseCallback` 来改变。
2. `group`，请求可以分组，这里为组名。默认组为空。
3. `name`，默认为请求的 URL。
4. `method`，请求的类型。
5. `scenario`，当请求属于某个场景时，这里为场景的名字。默认值为 `default`。
6. `status`，响应状态码。
7. `url`，默认为请求的 URL。

以下 JSON 示例展示了测试结果数据点的日志记录方式。可以看到，标签是通过 `tags` 来组织数据的。


```json
{
  "type": "Point",
  "metric": "http_req_duration",
  "data": {
    "time": "2017-06-02T23:10:29.52444541+02:00",
    "value": 586.831127,
    "tags": {
      "expected_response": "true",
      "group": "",
      "method": "GET",
      "name": "http://test.k6.io",
      "scenario": "default",
      "status": "200",
      "url": "http://test.k6.io"
    }
  }
}
```

读者可能对上面提到的 `group` 和 `scenario` 不太熟悉，这里介绍一下。

group 可以将相关的测试步骤分组，方便在测试结果中按组分析性能数据（如响应时间内、错误率等）。如按下面的方式分组：

```js
import http from 'k6/http';

export default function() {
  // 登录相关的操作分组
  const loginGroup = group('用户登录', () => {
    http.get('https://api.example.com/login');
  });

  // 数据查询操作分组
  group('数据查询', () => {
    http.get('https://api.example.com/data');
  });
}
```

scenario 则是定义虚拟用户（VU）的行为模式、负载策略和资源分配，模拟不同用户角色或负载场景。场景就定义在[Grafana k6 入门示例](https://www.lyyyuna.com/2025/03/12/grafana-k6-1/)中介绍的配置项中：

```js
export let options = {
  scenarios: {
    // 场景1：模拟 100 个恒定并发用户执行登录操作
    login_scenario: {
      executor: 'constant-vus',
      vus: 100,
      duration: '30s',
      exec: 'login',
    },

    // 场景2：模拟注册用户行为，逐步增加并发
    register_scenario: {
      executor: 'ramping-vus',
      startVUs: 10,
      stages: [
        { target: 50, duration: '20s' },
        { target: 100, duration: '30s' },
      ],
      exec: 'register',
    },
  },
};

// 定义场景执行的函数
export function login() {
  http.get('https://api.example.com/login');
}

export function register() {
  http.post('https://api.example.com/register', { username: 'test' });
}
```

### 合并标签

上文提到，`name` 标签默认是完整的请求 URL。以下代码会访问 100 个不同的 URL：

```js
import http from 'k6/http';

export default function() {
  // 生成并访问100个不同路径的URL
  for (let i = 1; i <= 100; i++) {
    http.get(`https://api.example.com/data/${i}`);
  }
}
```

测试结果会产生一大批唯一的 URL。每次请求的 name 标签都会记录完整的URL（如 /data/1、/data/2 等），这会导致：

1. **数据分散**：指标会被分散到100个不同的标签组中，难以聚合分析整体性能。
2. **存储压力**：大量唯一标签会增加指标存储的负担。

这个问题可以通过自定义标签覆盖默认行为，例如将 name 赋值为固定值：

```js
import http from 'k6/http';

export default function () {
  for (let id = 1; id <= 100; id++) {
    http.get(`http://example.com/posts/${id}`, {
      tags: { name: 'PostsItemURL' },
    });
  }
}
```