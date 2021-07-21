title: 用 asyncio 封装文件读写
date: 2016-03-27 17:36:25
categories: 杂
tags: 
- Python
- asyncio
---



## 前言

和网络 IO 一样，文件读写同样是一个费事的操作。

默认情况下，Python 使用的是系统的阻塞读写。这意味着在 asyncio 中如果调用了 
    
    f = file('xx')
    f.read()
    
会阻塞事件循环。

本篇简述如何用 asyncio.Future 对象来封装文件的异步读写。

代码在 [GitHub](https://github.com/lyyyuna/script_collection/blob/master/aysncfile/asyncfile.py)。目前仅支持 Linux。

## 阻塞和非阻塞

首先需要将文件的读写改为非阻塞的形式。在非阻塞情况下，每次调用 read 都会立即返回，如果返回值为空，则意味着文件操作还未完成，反之则是读取的文件内容。

阻塞和非阻塞的切换与操作系统有关，所以本篇暂时只写了 Linux 版本。如果有过 Unix 系统编程经验，会发现 Python 的操作是类似的。

    flag = fcntl.fcntl(self.fd, fcntl.F_GETFL) 
    if fcntl.fcntl(self.fd, fcntl.F_SETFL, flag | os.O_NONBLOCK) != 0: 
        raise OSError() 

## Future 对象

Future 对象类似 Javascript 中的 Promise 对象。它是一个占位符，其值会在将来被计算出来。我们可以使用

    result = await future
    
在 future 得到值之后返回。而使用

    future.set_result(xxx)
    
就可以设置 future 的值，也意味着 future 可以被返回了。await 操作符会自动调用 future.result() 来得到值。

## loop.call_soon

通过 loop.call_soon 方法可以将一个函数插入到事件循环中。

至此，我们的异步文件读写思路也就出来了。通过 loop.call_soon 调用非阻塞读写文件的函数。若一次文件读写没有完成，则计算剩余所学读写的字节数，并再次插入事件循环直至读写完毕。

可以发现其就是把传统 Unix 编程里，非阻塞文件读写的 while 循环换成了 asyncio 的事件循环。

下面是这一过程的示意代码。

```python
    def read_step(self, future, n, total):
        res = self.fd.read(n)
        if res is None:
            self.loop.call_soon(self.read_step, future, n, total)
            return
        if not res: # EOF
            future.set_result(bytes(self.rbuffer))
            return
        self.rbuffer.extend(res)
        self.loop.call_soon(self.read_step, future, self.BLOCK_SIZE, total)

    def read(self, n=-1):
        future = asyncio.Future(loop=self.loop)

        self.rbuffer.clear()
        self.loop.call_soon(self.read_step, future, min(self.BLOCK_SIZE, n), n)

        return future
```

