title: 使用 GenServer 的客户端-服务器通信
date: 2025-06-30 10:56:33
series: 用 Elixir 实现一个 KV 存储服务

---

由于 agent 本身就是进程，每个 bucket 都有一个进程标识符(PID)，但这些 bucket 并没有名称，可以通过 atoms 在 Elixir 中注册进程：

```bash
iex> Agent.start_link(fn -> %{} end, name: :shopping)
{:ok, #PID<0.43.0>}
iex> KV.Bucket.put(:shopping, "milk", 1)
:ok
iex> KV.Bucket.get(:shopping, "milk")
1
```

然而，使用 atoms 来命名动态进程是个糟糕的主意！如果采用 atoms 命名方式，我们就需要将 bucket 名称（通常来自外部客户端输入）转换为 atoms，而**绝对不应该将用户输入直接转换为 atoms**。这是因为 atoms 不会被垃圾回收机制处理——一旦 atoms 被创建就永远存在。若根据用户输入动态生成 atoms，意味着用户可以注入大量不同的名称，最终耗尽系统内存！

实际上，在内存耗尽之前，你更有可能先达到 Erlang 虚拟机对 atoms 数量的上限限制，无论如何这都会导致你的系统崩溃。

与其滥用内置的命名机制，我们将创建自己的进程注册表，将 bucket 名称与对应的 bucket 进程关联起来。

注册表必须保证始终处于最新状态。例如，当某个 bucket 进程因故障崩溃时，注册表必须感知这一变更，避免提供过期的进程信息。在 Elixir 中，这种机制称为"注册表需要监控（monitor）每个 bucket 进程"。由于注册表需要能够接收并处理系统发送的即时消息，仅靠 `Agent` API 已无法满足需求。

我们将使用 `GenServer` 来创建一个能够监控 bucket 进程的注册表进程。`GenServer` 为 Elixir 和 OTP 提供了工业级强度的服务器构建功能。

## GenServer 回调

GenServer 是一种进程，它会在特定条件下调用一组有限的函数。当我们使用 Agent 时，客户端代码和服务器端代码通常是并列编写的，如下所示：

```elixir
def put(bucket, key, value) do
  Agent.update(bucket, &Map.put(&1, key, value))
end
```

让我们来稍微拆解一下这段代码：

```elixir
def put(bucket, key, value) do
  # 这里是客户端代码
  Agent.update(bucket, fn state ->
    # 这里是服务端代码
    Map.put(state, key, value)
  end)
  # 回到客户端代码
end
```

在上面这段代码中，我们有一个进程（称为"客户端"）向一个 Agent 进程（称为"服务端"）发送请求。该请求包含一个匿名函数，这个函数必须在服务端执行。

而在 GenServer 的实现中，上述代码会被拆分成两个独立的函数，大致如下：

```elixir
def put(bucket, key, value) do
  # 发送给服务器 :put 指令
  GenServer.call(bucket, {:put, key, value})
end

# 服务器回调

def handle_call({:put, key, value}, _from, state) do
  {:reply, :ok, Map.put(state, key, value)}
end
```

GenServer 的代码确实需要更多的规范步骤，但正如我们将要看到的，这也带来了一些优势。

目前，我们暂时只为 bucket 注册逻辑编写服务器回调函数，稍后再提供完整的 API 接口。

新建一个文件 `lib/kv/registry.ex`，内容如下：

```elixir
defmodule KV.Registry do
  use GenServer

  ## Missing Client API - will add this later

  ## Defining GenServer Callbacks

  @impl true
  def init(:ok) do
    {:ok, %{}}
  end

  @impl true
  def handle_call({:lookup, name}, _from, names) do
    {:reply, Map.fetch(names, name), names}
  end

  @impl true
  def handle_cast({:create, name}, names) do
    if Map.has_key?(names, name) do
      {:noreply, names}
    else
      {:ok, bucket} = KV.Bucket.start_link([])
      {:noreply, Map.put(names, name, bucket)}
    end
  end
end
```

在 GenServer 中，您可以发送两种类型的请求：call（同步调用）和 cast（异步调用）。call 是同步请求，服务器必须返回响应，在服务器计算响应期间客户端会保持等待。cast 是异步请求，服务器不会返回响应，因此客户端也不会等待响应。这两种请求都是以消息形式发送给服务器的，并且会按顺序依次处理。

在上述实现中，我们通过模式匹配来区分不同消息：

* 处理 `:create` 消息时采用 cast（异步调用）
* 处理 `:lookup` 消息时采用 call（同步调用）

要触发上述回调函数，我们需要通过相应的 GenServer 函数来操作。让我们启动一个注册表进程，创建一个具名 bucket，然后进行查询：

```bash
iex> {:ok, registry} = GenServer.start_link(KV.Registry, :ok)
{:ok, #PID<0.136.0>}
iex> GenServer.cast(registry, {:create, "shopping"})
:ok
iex> {:ok, bucket} = GenServer.call(registry, {:lookup, "shopping"})
{:ok, #PID<0.174.0>}
```

我们的 `KV.Registry` 进程依次接收到了两条消息：首先是 cast 方式的 `{:create, "shopping"}` 请求，随后是 call 方式的 `{:lookup, "shopping"}` 请求。`GenServer.cast `在消息发送给注册表后会立即返回，而 `GenServer.call` 则会等待来自 `KV.Registry.handle_call` 回调函数提供的响应结果。

你可能已经注意到，我们在每个回调函数前都添加了 `@impl true` 注解。`@impl true` 会告知编译器：我们接下来要定义的函数是一个回调实现。这样，如果我们不小心写错了函数名或参数数量（比如错误地定义了一个 `handle_call/2` 函数），编译器就会警告我们并不存在需要实现的 `handle_call/2` 回调，同时还会列出 GenServer 模块所有已知的回调函数清单。

虽然目前实现得不错，但我们还需要为用户提供一个 API 接口层，这样就能隐藏具体的实现细节。

## 客户端 API

GenServer 的实现包含两个部分：客户端 API 和服务器回调。你可以选择将这两部分合并到单个模块中，也可以将它们分别放在客户端模块和服务器模块中。客户端是指调用客户端函数的任何进程，而服务器始终是进程标识符(PID)或进程名称，我们会将其作为参数显式传递给客户端 API。在本文中，我们将使用单个模块来同时包含服务器回调和客户端 API。

编辑 `lib/kv/registry.ex` 文件，补全客户端 API 的空缺部分：

```elixir
  ## Client API

  @doc """
  启动注册中心
  """
  def start_link(opts) do
    GenServer.start_link(__MODULE__, :ok, opts)
  end

  @doc """
  查找存储在 server 中与 name 对应的 bucket 进程 pid

  如果 bucket 存在则返回 {:ok, pid}，否则返回 :error
  """
  def lookup(server, name) do
    GenServer.call(server, {:lookup, name})
  end

  @doc """
  确保在 server 中存在与指定 name 关联的 bucket
  """
  def create(server, name) do
    GenServer.cast(server, {:create, name})
  end
```

第一个函数是 `start_link/1`，它会启动一个新的 GenServer 进程并传入一个选项列表。`start_link/1` 内部调用了 `GenServer.start_link/3` 函数，后者接收三个参数：

1. 实现服务器回调的模块（此处使用 `__MODULE__` 宏表示当前模块）
2. 初始化参数（此处使用原子 `:ok` 作为参数）
3. 选项列表（可用于配置服务器名称等参数）。目前，我们将 `start_link/1` 接收到的选项列表直接透传给 `GenServer.start_link/3`

接下来的两个函数 `lookup/2` 和 `create/2` 负责向服务器发送请求。这里我们分别使用了 `{:lookup, name}` 和 `{:create, name}` 这样的元组结构。通常会将请求动作指定为元组的第一个元素，而将相关参数放在后续元素中，这样就能在第一个参数槽位传递多个"参数"。需要注意的是，这些请求必须与 `handle_call/3` 或 `handle_cast/2` 的第一个参数模式相匹配。

客户端 API 部分就介绍到这里。在服务器端，我们可以实现多种回调函数来确保服务器的初始化、终止和请求处理。这些回调都是可选的，目前我们只实现了所需的部分。下面做个简要回顾：

首先是 `init/1` 回调函数，它接收传给 `GenServer.start_link/3` 的第二个参数，并返回 `{:ok, state}`（其中 state 是一个新建的映射）。从这里我们就能看出 `GenServer` 的 API 设计如何清晰地区分了客户端和服务端逻辑：`start_link/3` 在客户端执行，而 `init/1` 则是对应在服务端运行的回调。

对于 `call/2` 请求，我们通过实现 `handle_call/3` 回调来处理，该回调接收三个参数：

1. 请求内容（`request`）
2. 请求来源进程（`_from`，使用下划线表示未使用变量）
3. 当前服务器状态（`names`）

该回调需要返回 `{:reply, reply, new_state}` 格式的元组，其中：

1. 元组首元素 `:reply` 表示服务器需要向客户端返回响应
2. 第二个元素 `reply` 是将发送给客户端的实际响应数据
3. 第三个元素 `new_state` 是更新后的服务器状态

针对 `cast/2` 请求，我们通过实现 `handle_cast/2` 回调函数来处理，该函数接收两个参数：

1. 请求内容（`request`）
2. 当前服务器状态（`names`）

回调函数需返回 `{:noreply, new_state}` 格式的元组。值得注意的是，在实际生产应用中，我们通常会使用同步调用（call）而非异步请求（cast）来实现 `:create` 操作。这里之所以采用 cast 方式，是为了演示如何实现异步回调处理。

`handle_call/3` 和 `handle_cast/2` 回调函数还可以返回其他格式的元组响应。此外，我们还可以实现像 `terminate/2` 和 `code_change/3` 这样的回调函数。

现在，让我们编写一些测试来验证 GenServer 是否按预期工作。

## 测试 GenServer

测试 GenServer 与测试 Agent 并无太大差异。我们会在测试的 setup 回调中启动服务器进程，并在所有测试用例中复用该进程。创建 `test/kv/registry_test.exs` 文件，内容如下：

```elixir
defmodule KV.RegistryTest do
  use ExUnit.Case, async: true

  setup do
    registry = start_supervised!(KV.Registry)
    %{registry: registry}
  end

  test "spawns buckets", %{registry: registry} do
    assert KV.Registry.lookup(registry, "shopping") == :error

    KV.Registry.create(registry, "shopping")
    assert {:ok, bucket} = KV.Registry.lookup(registry, "shopping")

    KV.Bucket.put(bucket, "milk", 1)
    assert KV.Bucket.get(bucket, "milk") == 1
  end
end
```


测试用例的执行流程如下：

1. 首先验证注册表中不存在任何 bucket
2. 创建一个具名 bucket
3. 查询该 bucket
4. 最终验证其确为可用的 bucket 实例

我们为 `KV.Registry` 编写的 `setup` 代码块与之前 `KV.Bucket` 的 `setup` 有一个重要区别：不再手动调用 `KV.Registry.start_link/1` 来启动注册表，而是改用 `ExUnit.Callbacks.start_supervised!/2` 函数并传入 `KV.Registry` 模块来启动。

`start_supervised!` 函数通过 `use ExUnit.Case` 被注入到我们的测试模块中。它的工作原理是通过调用 `KV.Registry.start_link/1` 来启动注册表进程。使用 `start_supervised!` 的优势在于：

1. **进程生命周期管理**：ExUnit 会确保注册表进程在下个测试开始前被关闭
2. **测试隔离保障**：防止依赖共享资源的测试用例之间产生状态干扰
3. **自动清理机制**：即使测试失败也会执行进程终止操作

在测试中启动进程时，我们应当始终优先使用 `start_supervised!`。


## 监控的必要性

截至目前，我们实现的所有功能其实都可以通过 Agent 来完成。而在本节中，我们将展示 GenServer 相比 Agent 的独特优势——它能实现许多 Agent 无法完成的功能。

让我们从一个测试用例开始，该用例描述了当 bucket 进程停止或崩溃时期望注册表表现的行为：

```elixir
test "removes buckets on exit", %{registry: registry} do
  KV.Registry.create(registry, "shopping")
  {:ok, bucket} = KV.Registry.lookup(registry, "shopping")
  Agent.stop(bucket)
  assert KV.Registry.lookup(registry, "shopping") == :error
end
```

上述测试会在最后一个断言处失败，因为即使 bucket 进程停止后，其名称仍会保留在注册表中。

要修复这个问题，我们需要让注册表监控其创建的每个 bucket 进程。建立监控后，每当 bucket 进程退出时，注册表都会收到通知，从而使我们能够清理注册表。

首先让我们通过 `iex -S mix` 启动新控制台来实践监控机制：

```bash 
iex> {:ok, pid} = KV.Bucket.start_link([])
{:ok, #PID<0.66.0>}
iex> Process.monitor(pid)
#Reference<0.0.0.551>
iex> Agent.stop(pid)
:ok
iex> flush()
{:DOWN, #Reference<0.0.0.551>, :process, #PID<0.66.0>, :normal}
```

请注意，`Process.monitor(pid)` 会返回一个唯一的监控引用（reference），这个引用允许我们将后续收到的消息与对应的监控请求相匹配。当我们停止 Agent 进程后，可以通过 `flush/0` 查看所有消息，此时会观察到一条 `:DOWN` 消息 —— 其中包含与监控引用完全匹配的应用，通知我们 bucket 进程已以 `:normal` 原因退出。

让我们重新实现服务器回调函数来修复这个错误并使测试通过。具体需要以下修改步骤：

1. 状态结构调整：
    * 将 GenServer 状态改为包含两个映射表：
        * `name_to_pid`：存储 名称 -> 进程ID 的映射
        * `ref_to_name`：存储 监控引用 -> 名称 的映射
2. 监控机制实现：
    * 在 `handle_cast/2` 回调中为每个新创建的 bucket 进程设置监控
    * 新增 `handle_info/2` 回调处理监控消息：
        * 匹配 `:DOWN` 消息
        * 清理已终止进程的注册信息

下面是修改后的服务器回调完整实现：

```elixir
## Server callbacks

@impl true
def init(:ok) do
  names = %{}
  refs = %{}
  {:ok, {names, refs}}
end

@impl true
def handle_call({:lookup, name}, _from, state) do
  {names, _} = state
  {:reply, Map.fetch(names, name), state}
end

@impl true
def handle_cast({:create, name}, {names, refs}) do
  if Map.has_key?(names, name) do
    {:noreply, {names, refs}}
  else
    {:ok, bucket} = KV.Bucket.start_link([])
    ref = Process.monitor(bucket)
    refs = Map.put(refs, ref, name)
    names = Map.put(names, name, bucket)
    {:noreply, {names, refs}}
  end
end

@impl true
def handle_info({:DOWN, ref, :process, _pid, _reason}, {names, refs}) do
  {name, refs} = Map.pop(refs, ref)
  names = Map.delete(names, name)
  {:noreply, {names, refs}}
end

@impl true
def handle_info(msg, state) do
  require Logger
  Logger.debug("Unexpected message in KV.Registry: #{inspect(msg)}")
  {:noreply, state}
end
```

可以看到，我们在完全不修改客户端 API 的情况下，就完成了服务端实现的重大调整——这正是明确分离客户端与服务端架构带来的核心优势之一。

特别需要注意的是，与其他回调不同，我们为 `handle_info/2` 定义了一个"全捕获"（catch-all）处理分支，该分支会记录并丢弃所有未知消息。要理解这样设计的原因，让我们继续下一章节的探讨。

## call, cast 还是 info？

截至目前，我们已经使用了三种回调函数：`handle_call/3`、`handle_cast/2` 和 `handle_info/2`。以下是各回调函数的使用场景指南：

1. 同步请求处理 - `handle_call/3`
    * 必须用于同步请求
    * 应作为默认选择，因为等待服务端响应本身就是一种有效的背压(back-pressure)机制
    * 示例：本应在本章 `create/2` 函数中使用（当前为教学目的使用了 cast）
2. 异步请求处理 - `handle_cast/2`
    * 必须用于不需要回复的异步请求
    * 不保证服务端必定收到消息，故应谨慎使用
    * 典型场景：无需确认的广播类操作
3. 系统消息处理 - `handle_info/2`
    * 必须处理所有非 GenServer 标准调用消息
    * 包括：
        * 原生 `send/2` 发送的消息
        * 系统监控消息（如 `:DOWN` 通知）
        * 其他意外消息（需实现兜底处理）

由于所有消息（包括通过 `send/2` 发送的消息）都会交由 `handle_info/2` 处理，因此服务端很可能会收到预期之外的消息。如果不定义"全捕获"（catch-all）处理分支，这些消息就会因无法匹配任何模式而导致注册表进程崩溃。不过对于 `handle_call/3` 和 `handle_cast/2` 则无需此类担忧——因为 call/cast 操作只能通过 `GenServer` API 发起，未知消息通常意味着开发者的编码错误。
