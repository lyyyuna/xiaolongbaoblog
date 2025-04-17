title: k6 指标的使用
date: 2025-04-09 10:56:33
series: Grafana k6 登堂入室
summary: 指标 -> metrics

---

## 指标

指标用于衡量系统在测试条件下的表现。默认情况下，k6 会自动收集内置指标。除此之外还可以创建自定义指标。

指标主要分四类：

* **计数器指标**（Counters）用于累加数值。
* **仪表指标**（Gauges）跟踪最小值、最大值和最新值。
* **速率指标**（Rates）记录非零值的发生频率。
* **趋势指标**（Trends）计算多组值的统计信息（如平均值、众数或百分位数）。

若要让测试因未满足特定条件而失败，可以通过基于指标条件编写阈值（Threshold）实现（具体表达式形式取决于指标类型）。要筛选指标，可使用标签（Tags）和分组（groups）。你还可以将指标以多种汇总或精细格式导出。

## 内置指标

每个 k6 测试都会生成内置和自定义指标。每个支持的协议也都有其特定的指标。

### 标准内置指标

| **指标名称**	 | **类型** | **描述** |
|-------|-------|-------|
| checks  |   Rate |  成功检查的比率。|
| data_received  |   Counter	|  接收到的数据量。此示例说明如何跟踪单个 URL 的数据。  |
| data_sent  |   Counter	|  发送的数据量。跟踪单个 URL 的数据以监控其发送数据情况。  |
| dropped_iterations  |   Counter	|  由于虚拟用户（VU）不足（针对 arrival-rate 调度器）或时间超限（针对基于迭代的调度器的 maxDuration 过期）而未启动的迭代次数。  |
| iteration_duration  |   Trend		|  完成一次完整迭代所需的时间（包括 setup 和 teardown 阶段）。若需计算特定场景下迭代函数本身的持续时间，可参考此方法。  |
| iterations  |   Counter	|  虚拟用户执行 JS 脚本（默认函数）的总次数。  |
| vus  |   Gauge	|  当前活跃虚拟用户数。  |
| vus_max  |   Gauge		|  最大可能虚拟用户数（虚拟用户资源是预先分配的，以避免在增加负载时影响性能）。  |

### HTTP 专用内置指标


| **指标名称**               | **类型**       | **描述**                                                                                     |
|-------|-------|-------|
| http_req_blocked           | Trend          | 在发起请求前因等待空闲的 TCP 连接槽而被阻塞的时间（单位：浮点数）。                           |
| http_req_connecting        | Trend          | 建立与远程主机 TCP 连接所消耗的时间（单位：浮点数）。                                         |
| http_req_duration          | Trend          | 请求的总耗时。等于 `http_req_sending + http_req_waiting + http_req_receiving`，即远程服务器处理请求并响应的时间（不含 DNS 解析和初始连接时间）。 |
| http_req_failed            | Rate           | 根据 `setResponseCallback` 定义的失败请求比率。                                               |
| http_req_receiving         | Trend          | 从远程主机接收响应数据所消耗的时间（单位：浮点数）。                                       |
| http_req_sending           | Trend          | 向远程主机发送请求数据所消耗的时间（单位：浮点数）。                                       |
| http_req_tls_handshaking   | Trend          | 与远程主机进行 TLS 握手所消耗的时间（单位：浮点数）。                                       |
| http_req_waiting           | Trend          | 等待远程主机响应的时间（即“首字节时间”，Time To First Byte，TTFB）（单位：浮点数）。      |
| http_reqs                  | Counter        | k6 生成的总 HTTP 请求次数。                                                                    |

对于所有 `http_req_*` 指标，其时间戳会在请求结束时记录。换句话说，当 k6 接收到响应体末尾或请求超时时，就会生成该时间戳。

### 内置 WebSocket 指标

| **指标名称**               | **类型**       | **描述**                                                                                     |
|-------|-------|-------|
| ws_connecting              | Trend          | 建立 WebSocket 连接请求的总耗时（单位：浮点数）。                                             |
| ws_msgs_received           | Counter        | 接收到的消息总数。                                                                          |
| ws_msgs_sent               | Counter        | 发送的消息总数。                                                                            |
| ws_ping                    | Trend          | 从发送 ping 请求到接收 pong 响应的时间间隔（单位：浮点数）。                                   |
| ws_session_duration        | Trend          | WebSocket 会话的持续时间。从连接开始到虚拟用户（VU）执行结束的时间间隔（单位：浮点数）。     |
| ws_sessions                | Counter        | 启动的 WebSocket 会话总数。                                                                   |

### 内置 gRPC 指标

| **指标名称**               | **类型**       | **描述**                                                                                     |
|-------|-------|-------|
| grpc_req_duration          | Trend          | 从发送请求到接收远程主机响应的总耗时（单位：浮点数）。                                     |
| grpc_streams               | Counter        | 启动的 gRPC 流总数。                                                                          |
| grpc_streams_msgs_received | Counter        | 通过流接收到的消息总数。                                                                    |
| grpc_streams_msgs_sent     | Counter        | 通过流发送的消息总数。                                                                      |

## 自定义指标

除了内置指标外，你还可以创建自定义指标。每种指标类型都有一个用于创建自定义指标的构造函数，该构造函数会生成声明类型的指标对象，每种类型都提供 add 方法来实现指标测量。

### 创建自定义指标

自定义指标必须在初始化阶段创建。这一限制既控制了内存占用，又能确保 k6 验证所有阈值时对应的指标均已正确定义。下面是详细步骤：

```js
import http from 'k6/http';
// 1. 导入 k6/metrics 模块。也可以通过命名导入的方式指定要创建的指标类型
import { Trend } from 'k6/metrics';

// 2. 创建自定义趋势指标 myTrend，在结果输出中对应的指标名称显示为 waiting_time
const myTrend = new Trend('waiting_time');

export default function () {
  const r = http.get('https://quickpizza.grafana.com/');
  // 3. 使用 add 方法，将数据加入指标中
  myTrend.add(r.timings.waiting);
  console.log(myTrend.name); // waiting_time
}
```

### 查看自定义指标结果

自定义指标会出现在测试结果输出中，既包含测试结束时的总结，也包含详细的数据点。每种指标类型都有特定的聚合方法。你还可以为自定义指标的任意值添加标签，这些标签可用于筛选测试结果。

以下是前面脚本的输出可能在测试结束摘要中的显示方式。由于该指标是趋势型指标，k6 会根据数值数量及其总和计算各类趋势数据。

```
$ k6 run script.js

  ...
  INFO[0001] waiting_time                                  source=console

  ...
  iteration_duration.............: avg=1.15s    min=1.15s    med=1.15s    max=1.15s    p(90)=1.15s    p(95)=1.15s
  iterations.....................: 1     0.864973/s
  waiting_time...................: avg=265.245396 min=265.245396 med=265.245396 max=265.245396 p(90)=265.245396 p(95)=265.245396
``