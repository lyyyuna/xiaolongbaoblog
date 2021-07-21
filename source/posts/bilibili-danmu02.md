title: B站全站直播弹幕收集系统的简单设计
date: 2016-03-19 11:20:36
categories: 网络
tags: 
- bilibili
- asyncio
---

## 前言

虽然标题是全站，但目前只做了等级 top 100 直播间的全天弹幕收集。

[弹幕收集系统](https://github.com/lyyyuna/bilibili_danmu_colloector)基于之前的[Python版B站直播弹幕姬](https://github.com/lyyyuna/bilibili_danmu)修改而来。具体协议分析可以看[上一篇文章](http://www.lyyyuna.com/2016/03/14/bilibili-danmu01/)。

直播弹幕协议是直接基于 TCP 协议，所以如果 B 站对类似我这种行为做反制措施，比较困难。应该有我不知道的技术手段来检测类似我这种恶意行为。

我试过同时连接 100 个房间，和连接单个房间 100 次的实验，都没有问题。>150 会被关闭链接。

## 直播间的选取

现在[弹幕收集系统](https://github.com/lyyyuna/bilibili_danmu_colloector)在选取直播间上比较简单，直接选取了等级 top100。

以后会修改这部分，改成定时去 http://live.bilibili.com/all 查看新开播的直播间，并动态添加任务。

## 异步任务和弹幕存储

收集系统仍旧使用了 asyncio 异步协程框架，对于每一个直播间都使用如下方法来加进 loop 中。

    danmuji = bilibiliClient(url, self.lock, self.commentq, self.numq)
    task1 = asyncio.ensure_future(danmuji.connectServer())
    task2 = asyncio.ensure_future(danmuji.HeartbeatLoop())

其实若将心跳任务 HeartbeatLoop 放入 connectorServer 中去启动，代码看起来更优雅一些。但这么做是因为我需要维护一个任务列表，后面会有描述。

在弹幕存储上我花了些时间选择。
数据库存储是一个同步 IO 的过程，Insert 的时候会阻塞弹幕收集的任务。虽然有 aiomysql 这种异步接口，但配置数据库太麻烦，我的设想是这个小系统能够方便地部署。

最终我选择使用自带的 sqlite3。但 sqlite3 无法做并行操作，故开了一个线程单独进行数据库存储。在另一个线程中，100 * 2 个任务搜集所有的弹幕、人数信息，并塞进队列 commentq, numq 中。存储线程每隔 10s 唤醒一次，将队列中的数据写进 sqlite3 中，并清空队列。

在多线程和异步的配合下，网络流量没有被阻塞。

## 可能的连接失败场景处理

[弹幕协议](http://www.lyyyuna.com/2016/03/14/bilibili-danmu01/)是直接基于 TCP，位与位直接关联性较强，一旦解析错误，很容易就抛 Exception（个人感觉，虽然 TCP 是可靠传输，但B站服务器自身发生错误也是有可能的）。所以有必要设计一个自动重连机制。

在 asyncio 文档中提到，

    Done means either that a result / exception are available, or that the future was cancelled.
    
函数正常返回、抛出异常或者是被 cancel，都会退出当前任务。可以使用 done() 来判断。

每一个直播间对应两个任务，解析任务是最容易挂的，但并不会影响心跳任务，所以必须找出并将对应心跳任务结束。

在创建任务的时候使用字典记录每个房间的两个任务，

    self.tasks[url] = [task1, task2]

在运行过程中，每隔 10s 做一次检查，

    for url in self.tasks:
        item = self.tasks[url]
        task1 = item[0]
        task2 = item[1]
        if task1.done() == True or task2.done() == True:
            if task1.done() == False:
                task1.cancel()
            if task2.done() == False:
                task2.cancel()
            danmuji = bilibiliClient(url, self.lock, self.commentq, self.numq)
            task11 = asyncio.ensure_future(danmuji.connectServer())
            task22 = asyncio.ensure_future(danmuji.HeartbeatLoop())
            self.tasks[url] = [task11, task22]
            
实际我只见过一次任务失败的场景，是因为主播房间被封了，导致无法进入直播间。

## 结论

* B站人数是按照连接弹幕服务器的链接数量统计的。通过操纵链接量，可以**瞬间增加任意人数观看**，有商机？
* 运行的这几天中，发现即使大部分房间不在直播，也能有 >5 的人数，包括凌晨。我只能猜测也有和我一样的人在 24h 收集弹幕。
* top100 平均一天 40M 弹幕数据。
* 收集的弹幕能做什么？还没想好，可能可以拿来做用户行为分析 -_^