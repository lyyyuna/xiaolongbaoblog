title: Agent 实现简单的状态管理
date: 2025-06-27 10:56:33
series: 用 Elixir 实现一个 KV 存储服务

---

在本章中，我们将学习如何在多个实体之间保存和共享状态。如果你有编程经验，可能会想到全局共享变量，但我们这里要学习的模式截然不同。后续章节将对这里介绍的概念进行拓展。

## 创建工程

让我们通过命令行调用 `mix new` 来创建第一个项目。我们将项目路径作为参数传入（本例中为 `kv`）。默认情况下，应用程序名称和模块名称会从路径中获取。因此我们需要告知 Mix，我们的主模块应该是全大写的 `KV`，而不是默认的 `Kv`：

```bash
$ mix new kv --module KV
```

Mix 会创建一个名为 `kv` 的目录，其中包含以下文件：

```bash
* creating README.md
* creating .formatter.exs
* creating .gitignore
* creating mix.exs
* creating lib
* creating lib/kv.ex
* creating test
* creating test/test_helper.exs
* creating test/kv_test.exs
```

## （可变）状态带来的问题

Elixir 是一门基于不可变数据的语言，默认情况下所有数据都不共享。当我们需要在多个地方读取和修改共享信息时，Elixir 主要提供两种方案：

* 通过进程与消息传递：每个进程维护自己的状态，通过异步消息进行通信
* ETS (Erlang Term Storage)：内存中的键值存储，支持高效并发访问

我们很少手动创建原始进程，而是使用 Elixir 和 OTP 提供的抽象层：

* Agent - 对状态进行简单封装
* GenServer - “通用服务器”进程，封装状态并提供同步/异步调用，支持代码热更新等特性
* Task - 异步计算单元，允许创建进程并在稍后获取计算结果

这里我们将使用 Agent 来创建一个名为 `KV.Bucket` 的模块，该模块负责以允许其他进程读取和修改的方式存储我们的键值对数据。

## Agent 代理入门

`Agent` 代理是对状态的简单封装。如果你只需要一个进程来维护状态，那么 agent 会是一个很好的选择。让我们在项目内启动一个 `iex` 会话：

```bash
$ iex -S mix
```

然后我们来简单体验一下 agent 的使用：

```bash
iex> {:ok, agent} = Agent.start_link(fn -> [] end)
{:ok, #PID<0.57.0>}
iex> Agent.update(agent, fn list -> ["eggs" | list] end)
:ok
iex> Agent.get(agent, fn list -> list end)
["eggs"]
iex> Agent.stop(agent)
:ok
```

我们启动了一个初始状态为空列表的 agent。随后更新了 agent 的状态，将新项添加到列表头部。`Agent.update/3` 的第二个参数是一个函数，它接收 agent 当前状态作为输入并返回期望的新状态。最后，我们获取了整个列表。`Agent.get/3` 的第二个参数同样是一个函数，它接收状态作为输入并返回 `Agent.get/3` 自身将输出的值。当我们完成 agent 操作后，可以调用 `Agent.stop/3` 来终止 agent 进程。

`Agent.update/3` 函数的第二个参数可以接收任何单参数函数（该函数接收一个参数并返回一个值）。

```bash
iex> {:ok, agent} = Agent.start_link(fn -> [] end)
{:ok, #PID<0.338.0>}
iex> Agent.update(agent, fn _list -> 123 end)
:ok
iex> Agent.update(agent, fn content -> %{a: content} end)
:ok
iex> Agent.update(agent, fn content -> [12 | [content]] end)
:ok
iex> Agent.update(agent, fn list -> [:nop | list] end)
:ok
iex> Agent.get(agent, fn content -> content end)
[:nop, 12, %{a: 123}]
```

如你所见，我们可以任意修改 agent 的状态。因此，我们通常不希望在整个代码库的多个不同位置直接访问 Agent 的 API。相反，我们希望将所有与 Agent 相关的功能封装在一个单独的模块中，这里我们将其命名为 `KV.Bucket`。在实现这个模块之前，我们先编写一些测试用例，通过这些测试来明确该模块对外暴露的 API 接口。

创建一个文件 `test/kv/bucket_test.exs`（注意扩展名是 `.exs`），内容如下：

```elixir
defmodule KV.BucketTest do
  use ExUnit.Case, async: true

  test "stores values by key" do
    {:ok, bucket} = KV.Bucket.start_link([])
    assert KV.Bucket.get(bucket, "milk") == nil

    KV.Bucket.put(bucket, "milk", 3)
    assert KV.Bucket.get(bucket, "milk") == 3
  end
end
```

`use ExUnit.Case` 负责为我们的模块设置测试环境，并导入许多与测试相关的功能，例如 `test/2` 宏。

我们的第一个测试通过调用 `start_link/1` 并传入一个空列表作为选项来启动一个新的 `KV.Bucket`。然后对其执行了一些 `get/2` 和 `put/3` 操作，并对结果进行了断言。

需要特别注意的是传入 `ExUnit.Case` 的 `async: true` 选项。这个选项会让测试用例以异步方式运行，与其他标记了 `:async` 的测试用例并行执行，充分利用机器的多核处理能力。这对于加速测试套件执行非常有效。但需要注意：只有当测试用例不依赖或修改任何全局状态时才能设置 `:async` 标志。例如，如果测试需要写入文件系统或访问数据库，就应该保持同步运行（去掉 `:async` 选项），以避免测试间出现竞态条件。

无论是否异步运行，我们新添加的测试显然会失败，因为被测模块中尚未实现任何相关功能：


```
** (UndefinedFunctionError) function KV.Bucket.start_link/1 is undefined (module KV.Bucket is not available)
```

为了修复失败的测试，我们将创建 `lib/kv/bucket.ex` 文件，内容如下：

```elixir
defmodule KV.Bucket do
  use Agent

  @doc """
  Starts a new bucket.
  """
  def start_link(_opts) do
    Agent.start_link(fn -> %{} end)
  end

  @doc """
  Gets a value from the `bucket` by `key`.
  """
  def get(bucket, key) do
    Agent.get(bucket, &Map.get(&1, key))
  end

  @doc """
  Puts the `value` for the given `key` in the `bucket`.
  """
  def put(bucket, key, value) do
    Agent.update(bucket, &Map.put(&1, key, value))
  end
end
```

第一步是调用 `use Agent`。我们将学习的大多数功能（如 `GenServer` 和 `Supervisor`）都遵循这个模式。对于所有这些模块，调用 `use` 都会生成一个带有默认配置的 `child_spec/1` 函数，这对于监督进程非常有用（后面的博客会介绍）。接着，我们定义 `start_link/1` 函数来实际启动 `agent`。按照惯例，`start_link/1` 函数总是接收一个选项列表参数。然后我们调用 `Agent.start_link/1`，它接收一个返回 `agent` 初始状态的匿名函数。我们在 `agent` 内部维护一个 `map` 来存储键值对。通过 Agent API 和捕获操作符 `&` 来实现对 `map` 的读写操作。当调用 `Agent.get/2` 和 `Agent.update/2` 时，`agent` 会通过 `&1` 参数将其状态传递给匿名函数。

现在 `KV.Bucket` 模块已经定义完成，我们的测试应该可以通过了！可以通过运行 `mix test` 亲自验证。

## 使用 ExUnit 回调函数配置测试

在继续为 `KV.Bucket` 添加更多功能之前，我们先来讨论一下 `ExUnit` 的回调机制。所有 `KV.Bucket` 测试都需要启动并运行一个 bucket agent。幸运的是，ExUnit 支持回调函数，可以帮助我们省去这类重复性工作。

让我们用回调来重写测试用例：

```elixir
defmodule KV.BucketTest do
  use ExUnit.Case, async: true

  setup do
    {:ok, bucket} = KV.Bucket.start_link([])
    %{bucket: bucket}
  end

  test "stores values by key", %{bucket: bucket} do
    assert KV.Bucket.get(bucket, "milk") == nil

    KV.Bucket.put(bucket, "milk", 3)
    assert KV.Bucket.get(bucket, "milk") == 3
  end
end
```

我们首先使用 `setup/1` 宏定义了一个设置回调。这个 `setup/1` 宏定义的回调会在每个测试之前运行，并且与测试本身在同一个进程中执行。

需要注意的是，我们需要一种机制将 `bucket PID` 从回调传递给测试。这里我们通过测试上下文(test context)实现这一功能。当回调返回 `%{bucket: bucket}` 时，ExUnit 会将该映射合并到测试上下文中。由于测试上下文本身就是一个映射，我们可以通过模式匹配从中提取出 bucket，从而在测试内部访问这个 bucket 实例：

```elixir
test "stores values by key", %{bucket: bucket} do
  # `bucket` is now the bucket from the setup block
end
```


## 其他 agent 行为


除了可以分别获取值和更新 Agent 状态外，Agent 还允许我们通过 `Agent.get_and_update/2` 函数在一次调用中同时完成获取值和更新状态的操作。让我们实现一个 `KV.Bucket.delete/2` 函数，该函数会从 bucket 中删除指定键并返回其当前值：

```elixir
@doc """
Deletes `key` from `bucket`.

Returns the current value of `key`, if `key` exists.
"""
def delete(bucket, key) do
  Agent.get_and_update(bucket, &Map.pop(&1, key))
end
```

## agents 中的服务端/客户端架构

在进入下一章节之前，让我们先探讨 Agent 中的客户端/服务端二分架构。我们将扩展刚才实现的 `delete/2` 函数来说明这一概念：

```elixir
def delete(bucket, key) do
  Agent.get_and_update(bucket, fn map ->
    Map.pop(map, key)
  end)
end
```

我们传递给 Agent 的函数内部所有操作都在 Agent 进程中执行。在这种情况下，由于 Agent 进程负责接收和响应我们的消息，我们称 Agent 进程为服务端。而函数外部的所有操作都发生在客户端。

这个区分非常重要。如果需要执行耗时操作，你必须仔细考虑是在客户端还是服务端执行更为合适。例如：

```elixir
def delete(bucket, key) do
  Process.sleep(1000) # 让客户端 sleep
  Agent.get_and_update(bucket, fn map ->
    Process.sleep(1000) # 让服务端 sleep
    Map.pop(map, key)
  end)
end
```

当在服务端执行耗时操作时，该特定服务端的所有其他请求都将阻塞等待当前操作完成，这可能导致部分客户端请求超时。

在下一章中，我们将探讨 GenServer，其客户端与服务端的隔离机制将体现得更加明显。
