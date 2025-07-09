title: k6 模块
date: 2025-07-07 10:56:33
series: Grafana k6 登堂入室

---

## 导入模块

在测试脚本中导入模块或模块的部分内容是很常见的。在 k6 中，可以导入以下几种类型的模块：

* 内置模块
* 本地模块
* 远程模块
* 扩展模块

### 内置模块

k6 提供了许多内置模块来实现核心功能。例如，HTTP 客户端模块可用于向被测系统发送请求。完整的内置模块列表请参考 [API 文档](https://grafana.com/docs/k6/latest/javascript-api/)。

```js
import http from 'k6/http';
```

### 本地模块

这些模块存储在本地文件系统中，可通过相对路径或绝对文件系统路径访问。

k6 采用类似**浏览器的模块解析机制**，不支持 Node.js 的模块解析方式。导入的文件名必须完整指定，例如 `./helpers.js`。

```js
// helpers.js
export function someHelper() {
  // ...
}

// my-test.js
import { someHelper } from './helpers.js';

export default function () {
  someHelper();
}
```

### 远程模块

这些模块通过 HTTP(S) 协议从公开来源获取，例如 GitHub、任意 CDN 或任何可公开访问的 Web 服务器。导入的模块会在运行时下载并执行，因此在将代码纳入测试脚本前，务必确保其来源可信。

例如，[jslib](https://grafana.com/docs/k6/latest/javascript-api/jslib/) 是一组 k6 JavaScript 库，可作为远程 HTTPS 模块使用。它们既可以被下载后作为本地模块导入，也能直接作为远程模块引用。

```js
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export default function () {
  randomItem([1, 2, 3]);
}
```

你也可以构建自定义的 JavaScript 库，并通过公共网络托管平台进行分发。例如，[k6-jslib-aws](https://github.com/grafana/k6-jslib-aws) 和 [k6-rollup-example](https://github.com/grafana/k6-rollup-example) 项目就将它们的模块托管为 GitHub 的发布资源。

### 扩展模块

与 k6 API 类似，您可以使用 Go 代码构建自定义模块，并将其作为 JavaScript 模块公开。这类自定义的 Go 转 JS 模块被称为 k6 扩展。

以下示例演示了如何从 [xk6-kubernetes](https://github.com/grafana/xk6-kubernetes) 扩展中导入 k6/x/kubernetes 模块：

```js
import { Kubernetes } from 'k6/x/kubernetes';

const podSpec = {
  apiVersion: 'v1',
  kind: 'Pod',
  metadata: { name: 'busybox', namespace: 'testns' },
  spec: {
    containers: [
      {
        name: 'busybox',
        image: 'busybox',
        command: ['sh', '-c', 'sleep 30'],
      },
    ],
  },
};
export default function () {
  const kubernetes = new Kubernetes();
  kubernetes.create(podSpec);
  const pods = kubernetes.list('Pod', 'testns');
  pods.map(function (pod) {
    console.log(pod.metadata.name);
  });
}
```

k6 扩展（Go 转 JS 模块）是如何工作的？为了提升性能，k6 引擎采用 Go 语言编写，并内置了一个 JavaScript 虚拟机（sobek）来执行 JavaScript 测试代码。这种架构使得你能够用 Go 编写模块，并像普通 JavaScript 模块一样导入使用。

## 分享 JavaScript 模块

如先前所述，用户可通过加载本地模块或远程模块来导入自定义JavaScript库。因此，我们不仅有两种模块导入方式，还支持多种模块分发方案。

### 远程模块

你可以将模块部署在 GitHub 等公共网络服务器或 CDN 上，以便远程调用这些模块。

```js
// As GitHub release assets
import {
  WorkloadConfig,
  sayHello,
} from 'https://github.com/grafana/k6-rollup-example/releases/download/v0.0.2/index.js';

// or hosted in a CDN
import { randomIntBetween, randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
```

当库包含多个文件和模块时，你可能需要将这些模块打包以创建公共版本。

需要特别注意的是，k6 会自动执行远程模块，因此必须确保这些远程模块的源代码可信。某些托管机制可能存在篡改远程模块的风险。为降低此类安全隐患，部分用户倾向于先将模块下载到本地再导入，从而完全掌控源代码。

### 本地模块

在本示例中，前述远程模块已被下载至测试项目的 lib 目录，并按以下方式导入：

```js
import { WorkloadConfig, sayHello } from './libs/test-commons.js';

import { randomIntBetween, randomItem } from './libs/k6-utils.js';
```

另一种分发库的方法是使用 npm 等包管理工具，这种方式支持版本锁定和本地库链接功能，后者在开发阶段特别实用。

k6 既非 Node.js 也非浏览器环境。依赖于 Node.js 特有 API（如 os 和 fs 模块）的包无法在 k6 中运行，同理浏览器专属 API（例如 window 对象）也不受支持。虽然 k6 不支持解析 Node 模块，但可以通过打包工具（如 [k6-rollup-example](https://github.com/grafana/k6-rollup-example) 所示）来加载 npm 依赖项。

