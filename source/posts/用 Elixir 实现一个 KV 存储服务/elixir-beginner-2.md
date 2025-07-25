title: 监督树和应用程序
date: 2025-07-25 10:56:33
series: 用 Elixir 实现一个 KV 存储服务

---


在上一章关于 `GenServer` 的内容中，我们实现了 `KV.Registry` 来管理 bucket 。在某个时刻，我们开始监控这些 bucket，这样当 `KV.Bucket` 进程崩溃时，我们就能采取相应的措施。虽然这个改动相对较小，但它引入了一个经常被 Elixir 开发者问到的问题：故障时会发生什么？

在我们添加监控之前，如果一个 bucket 崩溃了，注册表(`registry`)会永远指向一个已经不存在的 bucket 。如果用户试图读取或写入这个已崩溃的 bucket ，操作将会失败。任何尝试用相同名称创建新 bucket 的操作都只会返回已崩溃 bucket 的 PID。换句话说，该 bucket 在注册表中的条目将永远处于错误状态。一旦我们添加了监控，注册表会自动移除已崩溃 bucket 的条目。现在尝试查找已崩溃的 bucket 时，会提示该 bucket 不存在，系统用户可以根据需要成功创建一个新的 bucket 。

在实践中，我们并不期望作为 bucket 的进程会失败。但如果确实发生了故障，无论出于什么原因，我们都可以确信系统将继续按预期工作。

如果你有编程经验，你可能会想知道：我们能否从一开始就保证 bucket 不会崩溃？正如我们将看到的，Elixir 开发者倾向于将这类实践称为"防御性编程"。这是因为一个在线生产系统有数不清的原因可能导致出错。磁盘可能故障，内存可能损坏，出现 bug，网络可能中断一秒等等。如果我们编写软件试图保护或规避所有这些错误，那么我们花在处理故障上的时间将超过编写自己软件的时间！

因此，Elixir 开发者更倾向于**让它崩溃**或**快速失败**。而从故障中恢复的最常见方式之一就是重新启动系统中崩溃的部分。

例如，想象你的电脑、路由器、打印机或任何设备工作不正常。你有多少次通过重启来修复它？一旦我们重启设备，就将设备重置回初始状态，这是经过充分测试并保证能正常工作的状态。在 Elixir 中，我们将同样的方法应用于软件：每当一个进程崩溃时，我们就启动一个新进程来执行与崩溃进程相同的工作。

在 Elixir 中，这是通过**监督者(Supervisor)**来实现的。监督者是一个进程，它监督其他进程并在它们崩溃时重新启动它们。为此，监督者管理任何被监督进程的整个生命周期，包括启动和关闭。

在本文中，我们将学习如何通过监督 `KV.Registry` 进程来将这些概念付诸实践。毕竟，如果注册表出了问题，整个注册表都会丢失，任何 bucket 都无法被找到！为了解决这个问题，我们将定义一个 `KV.Supervisor` 模块，保证我们的 `KV.Registry` 在任何时刻都能正常运行。

在本文的最后，我们还将讨论**应用程序(Applications)**。正如我们将看到的，Mix 一直在将我们所有的代码打包成一个应用程序，我们将学习如何自定义我们的应用程序，以保证每当系统启动时，我们的监督者和注册表都能正常运行。

## 第一个监督者

监督者是一个监督其他进程的进程，我们称之为子进程。监督一个进程包括三个不同的职责。首先是启动子进程。一旦子进程开始运行，监督者可能会重启子进程，要么因为它异常终止，要么因为达到了某种条件。例如，如果任何子进程死亡，监督者可能会重启所有子进程。最后，当系统关闭时，监督者还负责关闭子进程。请参阅 [Supervisor](https://hexdocs.pm/elixir/Supervisor.html) 模块以获得更深入的讨论。

创建一个监督者与创建 `GenServer` 没有多大区别。我们将在 `lib/kv/supervisor.ex` 文件中定义一个名为 `KV.Supervisor` 的模块，它将使用 `Supervisor` 的 `behaviour`：

```elixir
defmodule KV.Supervisor do
  use Supervisor

  def start_link(opts) do
    Supervisor.start_link(__MODULE__, :ok, opts)
  end

  @impl true
  def init(:ok) do
    children = [
      KV.Registry
    ]

    Supervisor.init(children, strategy: :one_for_one)
  end
end
```

到目前为止，我们的监督者只有一个子进程：`KV.Registry`。在我们定义了子进程列表后，我们调用 `Supervisor.init/2`，传入子进程列表和监督策略。

监督策略决定了当其中一个子进程崩溃时会发生什么。`:one_for_one` 意味着如果一个子进程死亡，只有它会被重启。由于我们现在只有一个子进程，这就足够了。`Supervisor` 的 `behaviour` 支持几种策略，我们将在本文中讨论它们。

一旦监督者启动，它将遍历子进程列表，并在每个模块上调用 `child_spec/1` 函数。

`child_spec/1` 函数返回子进程规范，该规范描述了如何启动进程、进程是工作者还是监督者、进程是临时的、瞬态的还是永久的等等。当我们使用 `Agent`、`use GenServer`、`use Supervisor` 等时，`child_spec/1` 函数会自动定义。让我们在终端中用 `iex -S mix` 试试看：

```elixir
iex> KV.Registry.child_spec([])
%{id: KV.Registry, start: {KV.Registry, :start_link, [[]]}}
```

在我们继续本指南的过程中，我们将学习这些细节。如果你想提前了解，可以查看 [Supervisor](https://hexdocs.pm/elixir/Supervisor.html) 文档。

在监督者检索到所有子进程规范后，它会按照定义的顺序，使用子进程规范中 `:start` 键的信息，逐个启动其子进程。对于我们当前的规范，它将调用 `KV.Registry.start_link([])`。

让我们来试用一下这个监督者：

```elixir
iex> {:ok, sup} = KV.Supervisor.start_link([])
{:ok, #PID<0.148.0>}
iex> Supervisor.which_children(sup)
[{KV.Registry, #PID<0.150.0>, :worker, [KV.Registry]}]
```

到目前为止，我们已经启动了监督者并列出了它的子进程。一旦监督者启动，它也启动了所有的子进程。

如果我们故意让监督者启动的注册表崩溃会发生什么？让我们通过在调用时发送一个错误的输入来实现：

```elixir
iex> [{_, registry, _, _}] = Supervisor.which_children(sup)
[{KV.Registry, #PID<0.150.0>, :worker, [KV.Registry]}]
iex> GenServer.call(registry, :bad_input)
08:52:57.311 [error] GenServer #PID<0.150.0> terminating
** (FunctionClauseError) no function clause matching in KV.Registry.handle_call/3
iex> Supervisor.which_children(sup)
[{KV.Registry, #PID<0.157.0>, :worker, [KV.Registry]}]
```

注意，当我们由于错误输入导致注册表崩溃后，监督者如何自动启动了一个新的注册表，使用新的 PID 来替代第一个注册表。

在前面的章节中，我们总是直接启动进程。例如，我们会调用 `KV.Registry.start_link([])`，它会返回 `{:ok, pid}`，这样我们就可以通过 PID 与注册表进行交互。现在进程是由监督者启动的，我们必须直接询问监督者它的子进程是谁，并从返回的子进程列表中获取 PID。在实践中，每次都这样做会非常昂贵。为了解决这个问题，我们经常给进程命名，允许它们在单台机器上从我们代码的任何地方被唯一标识。

让我们学习如何做到这一点。

## 给进程命名

虽然我们的应用程序会有很多 bucket ，但它只会有一个注册表。因此，每当我们启动注册表时，我们想要给它一个唯一的名称，这样我们就可以从任何地方访问它。我们通过向 `KV.Registry.start_link/1` 传递 `:name` 选项来实现这一点。

让我们稍微改变一下子进程的定义（在 `KV.Supervisor.init/1` 中），使用元组列表而不是 atom 列表：

```elixir
  def init(:ok) do
    children = [
      {KV.Registry, name: KV.Registry}
    ]
```

通过这样做，监督者现在将通过调用 `KV.Registry.start_link(name: KV.Registry)` 来启动 `KV.Registry`。

如果你重新查看 `KV.Registry.start_link/1` 的实现，你会记得它只是将选项传递给 `GenServer`：

```elixir
  def start_link(opts) do
    GenServer.start_link(__MODULE__, :ok, opts)
  end
```

这反过来会用给定的名称注册进程。`:name` 选项期望使用 atom 作为本地命名进程（本地命名意味着它在此机器上可用——还有其他选项，我们在这里不讨论）。由于模块标识符是 atom，我们可以用实现该进程的模块名来命名进程，前提是该名称只有一个进程。这在调试和检查系统时很有帮助。

让我们在 `iex -S mix` 中尝试更新后的监督者：

```elixir
iex> KV.Supervisor.start_link([])
{:ok, #PID<0.66.0>}
iex> KV.Registry.create(KV.Registry, "shopping")
:ok
iex> KV.Registry.lookup(KV.Registry, "shopping")
{:ok, #PID<0.70.0>}
```

这次监督者启动了一个命名的注册表，使我们能够创建 bucket ，而无需显式地从监督者那里获取 PID。你也应该知道如何在不查找其 PID 的情况下再次使注册表崩溃：试试看吧。

> 此时，你可能会想：是否也应该给 bucket 进程本地命名？记住 bucket 是根据用户输入动态启动的。由于本地名称必须是 atom，我们必须动态创建 atom，但这不是一个好主意，因为一旦定义了 atom，它永远不会被擦除或垃圾回收。这意味着，如果我们根据用户输入动态创建 atom，最终会耗尽内存（或者更准确地说，虚拟机会因为对 atom 数量设置了硬限制而崩溃）。正是这个限制促使我们创建了自己的注册表（或者使用 Elixir 内置的 Registry 模块的原因）。

我们越来越接近一个完全工作的系统了。监督者会自动启动注册表。但是我们如何在系统启动时自动启动监督者呢？要回答这个问题，让我们来谈谈应用程序。

## 理解应用程序

我们一直以来都在一个应用程序中工作。每当我们更改文件并运行 `mix compile` 时，我们都能在编译输出中看到 `Generated kv app` 消息。

我们可以在 `_build/dev/lib/kv/ebin/kv.app` 找到生成的 `.app` 文件。让我们看看它的内容：

```elixir
{application,kv,
             [{applications,[kernel,stdlib,elixir,logger]},
              {description,"kv"},
              {modules,['Elixir.KV','Elixir.KV.Bucket','Elixir.KV.Registry',
                        'Elixir.KV.Supervisor']},
              {registered,[]},
              {vsn,"0.1.0"}]}.
```

这个文件包含 Erlang 项（使用 Erlang 语法编写）。尽管我们不熟悉 Erlang，但很容易猜到这个文件保存了我们的应用程序定义。它包含我们的应用程序版本、所有由它定义的模块，以及我们依赖的应用程序列表，如 Erlang 的 `kernel`、`elixir` 本身和 `logger`。

简而言之，一个应用程序由 `.app` 文件中定义的所有模块组成，包括 `.app` 文件本身。一个应用程序通常只有两个目录：`ebin`，用于存放 Elixir 的编译产物，如 `.beam` 和 `.app` 文件；以及 `priv`，存放应用程序可能需要的其他产物或资源文件。

尽管 Mix 为我们生成并维护 `.app` 文件，但我们可以通过在 `mix.exs` 项目文件中的 `application/0` 函数内添加新条目来自定义其内容。我们很快就会进行第一次自定义。

## 启动应用程序

我们系统中的每个应用程序都可以启动和停止。启动和停止应用程序的规则也在 `.app` 文件中定义。当我们调用 `iex -S mix` 时，Mix 会编译我们的应用程序然后启动它。

让我们在实践中看看这一点。使用 `iex -S mix` 启动一个控制台并尝试：

```elixir
iex> Application.start(:kv)
{:error, {:already_started, :kv}}
```

哎呀，它已经启动了。Mix 会自动启动当前应用程序及其所有依赖项。对于 `mix test` 和许多其他 Mix 命令也是如此。

但是，我们可以停止我们的 `:kv` 应用程序，以及 `:logger` 应用程序：

```elixir
iex> Application.stop(:kv)
:ok
iex> Application.stop(:logger)
:ok
```

再次启动我们的应用程序：

```elixir
iex> Application.start(:kv)
{:error, {:not_started, :logger}}
```

现在我们得到了一个错误，因为 `:kv` 依赖的一个应用程序（在这个例子中是 `:logger`）没有启动。我们需要按照正确的顺序手动启动每个应用程序，或者调用 `Application.ensure_all_started/1`，如下所示：

```elixir
iex> Application.ensure_all_started(:kv)
{:ok, [:logger, :kv]}
```

在实践中，我们的工具总是为我们启动应用程序，但如果你需要细粒度的控制，也有相应的 API 可用。

## 应用程序回调

每当我们调用 `iex -S mix` 时，Mix 会通过调用 `Application.start(:kv)` 自动启动我们的应用程序。但是我们能自定义应用程序启动时发生的事情吗？事实上，我们可以！为此，我们得定义一个应用程序回调。

首先是告诉我们的应用程序定义（例如，我们的 `.app` 文件）哪个模块将实现应用程序回调。让我们打开 `mix.exs` 并将 `def application` 更改为以下内容：

```elixir
  def application do
    [
      extra_applications: [:logger],
      mod: {KV, []}
    ]
  end
```

`:mod` 选项指定了“应用程序回调模块”，后面跟着在应用程序启动时要传递的参数。应用程序回调模块可以是任何实现了 `Application` 行为的模块。

要实现 `Application` 行为，我们必须使用 `Application` 并定义一个 `start/2` 函数。`start/2` 函数的目标是启动一个监督者，然后由该监督者启动任何子服务或执行我们的应用程序可能需要的任何其他代码。让我们利用这个机会来启动我们在本文之前实现的 `KV.Supervisor`。

由于我们已经指定 `KV` 作为模块回调，让我们修改在 `lib/kv.ex` 中定义的 `KV` 模块来实现 `start/2` 函数：

```elixir
defmodule KV do
  use Application

  @impl true
  def start(_type, _args) do
    # 我们不需要直接使用下面的监督者名称，但在调试或检查系统时，它可能很有用。
    KV.Supervisor.start_link(name: KV.Supervisor)
  end
end
```

当我们使用 `Application` 时，我们可以定义几个函数，类似于我们使用 `Supervisor` 或 `GenServer` 时。这次我们只需要定义一个 `start/2` 函数。`Application` 行为还有一个 `stop/1` 回调，但在实践中很少使用。你可以查看文档获取更多信息。

现在你已经定义了一个启动我们监督者的应用程序回调，我们期望一旦启动 `iex -S mix`，`KV.Registry` 进程就会上线并运行。让我们再试一次：

```elixir
iex> KV.Registry.create(KV.Registry, "shopping")
:ok
iex> KV.Registry.lookup(KV.Registry, "shopping")
{:ok, #PID<0.88.0>}
```

让我们回顾一下正在发生的事情。每当我们调用 `iex -S mix` 时，它会通过调用 `Application.start(:kv)` 自动启动我们的应用程序，然后调用应用程序回调。应用程序回调的任务是启动一个监督树。现在，我们的监督者有一个名为 `KV.Registry` 的子进程，以 `KV.Registry` 名称启动。我们的监督者可能还有其他子进程，其中一些子进程可能是它们自己的监督者，有自己的子进程，从而形成所谓的监督树。

## 项目还是应用程序？

Mix 区分项目和应用程序。根据我们的 `mix.exs` 文件内容，我们可以说我们有一个定义了 `:kv` 应用程序的 Mix 项目。正如我们将在后续章节中看到的，有些项目不定义任何应用程序。

当我们说“项目”时，你应该想到 Mix。Mix 是管理你的项目的工具。它知道如何编译你的项目、测试你的项目等等。它还知道如何编译和启动与你的项目相关的应用程序。

当我们谈论应用程序时，我们谈论的是 OTP。应用程序是运行时整体启动和停止的实体。你可以在 [Application](https://hexdocs.pm/elixir/Application.html) 模块的文档中了解更多关于应用程序以及它们如何与整个系统的启动和关闭相关联的信息。

## 下一步

虽然这是我们第一次实现监督者，但这不是我们第一次使用监督者！在上一篇文章中，当我们在测试中使用 `start_supervised!` 启动注册表时，ExUnit 在框架自身管理的监督者下启动了注册表。通过定义我们自己的监督者，我们为如何在应用程序中初始化、关闭和监督进程提供了更多结构，使我们的生产代码和测试与最佳实践保持一致。

但我们还没有完成。到目前为止，我们只是在监督注册表，但我们的应用程序也在启动 bucket 。由于 bucket 是动态启动的，我们可以使用一种特殊类型的监督者，称为 `DynamicSupervisor`，它针对处理此类场景进行了优化。让我们在下一篇中继续探索。