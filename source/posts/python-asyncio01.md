title: asyncio 的 coroutine 与 Future
date: 2016-03-20 16:20:02
categories: 语言
tags: 
- Python
- asyncio
---

## coroutine 与 Future 的关系

看起来两者是一样的，因为都可以用以下的语法来异步获取结果，

    result = await future
    result = await coroutine
    
实际上，coroutine 是生成器函数，它既可以从外部接受参数，也可以产生结果。使用 coroutine 的好处是，我们可以暂停一个函数，然后稍后恢复执行。比如在涉及到网路操作的情况下，能够停下函数直到响应到来。在停下的这段时间内，我们可以切换到其他任务继续执行。

而 Future 更像是 Javascript 中的 Promise 对象。它是一个占位符，其值会在将来被计算出来。在上述的例子中，当我们在等待网络 IO 函数完成时，函数会给我们一个容器，Promise 会在完成时填充该容器。填充完毕后，我们可以用回调函数来获取实际结果。

Task 对象是 Future 的子类，它将 coroutine 和 Future 联系在一起，将 coroutine 封装成一个 Future 对象。

一般会看到两种任务启动方法，

    tasks = asyncio.gather(
        asyncio.ensure_future(func1()),
        asyncio.ensure_future(func2())
    )
    loop.run_until_complete(tasks)

和

    tasks = [
        asyncio.ensure_future(func1()),
        asyncio.ensure_future(func2())
        ]
    loop.run_until_complete(asyncio.wait(tasks))

ensure_future 可以将 coroutine 封装成 Task。asyncio.gather 将一些 Future 和 coroutine 封装成一个 Future。asyncio.wait 则本身就是 coroutine。

run_until_complete 既可以接收 Future 对象，也可以是 coroutine 对象，

    BaseEventLoop.run_until_complete(future)

    Run until the Future is done.
    If the argument is a coroutine object, it is wrapped by ensure_future().
    Return the Future’s result, or raise its exception.


## Task 任务的正确退出方式

在 asyncio 的任务循环中，如果使用 CTRL-C 退出的话，即使捕获了异常，Event Loop 中的任务会报错，出现如下的错误，

    Task was destroyed but it is pending!
    task: <Task pending coro=<kill_me() done, defined at test.py:5> wait_for=<Future pending cb=[Task._wakeup()]>>

根据官方文档，Task 对象只有在以下几种情况，会认为是退出，

    a result / exception are available, or that the future was cancelled
    
Task 对象的 cancel 和其父类 Future 略有不同。当调用 Task.cancel() 后，对应 coroutine 会在事件循环的下一轮中抛出 CancelledError 异常。使用 Future.cancelled() 并不能立即返回 True（用来表示任务结束），只有在上述异常被处理任务结束后才算是 cancelled。

故结束任务可以用

    for task in asyncio.Task.all_tasks():
        task.cancel()
        
这种方法将所有任务找出并 cancel。

但 CTRL-C 也会将事件循环停止，所以有必要重启事件循环，

```python
    try:
        loop.run_until_complete(tasks)
    except KeyboardInterrupt as e:
        for task in asyncio.Task.all_tasks():
            task.cancel()
        loop.run_forever() # restart loop
    finally:
        loop.close()
```

在每个 Task 中捕获异常是必要的，如果不确定，可以使用

    asyncio.gather(..., return_exceptions=True)

将异常转换为正常的结果返回。

    
