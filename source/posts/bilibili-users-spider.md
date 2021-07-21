title: 分布式 B 站用户信息爬虫
date: 2017-04-24 21:20:36
categories: 网络
tags: 
- Celery
- 爬虫
---

## 前言

上周末写了一个 B 站用户信息爬虫，怎么爬网页实在是没什么好讲的，所以这篇描述一下简单的分布式爬虫。

知乎和 B 站用户爬虫极其类似。它们都有两种爬取策略：

* 从一个用户开始，爬取关注者和被关注者，将新找到的用户存储下来，然后再从其关注网络中遍历爬取新的用户。这种爬取策略类似广度优先搜素，根据著名的[六度人脉理论](http://baike.baidu.com/link?url=w0Tr_YMnE4BHSLk8MN9QBaAlbAUS18BJrlq85ZuhNDYHcN4pQKXg9KIxJ6fMIcW-rr7pQbT3Ya02hlHfiFZVijScjomLbTfhvwwavVAN3XD4GQCjRACiVhza_tndVf0KUjhj1iYrBgvZ6mTe8UCGw_)，我们可以将大部分用户都找出来。不过该方法会增加程序的复杂度，你得维护好已搜索用户和未搜索用户的列表，快速去重等。
* 大部分网站在建立用户时会使用自增字段，或者有一个字段其值会在特定的数字范围内。对于 B 站，每个人的主页就是很好的例子，[http://space.bilibili.com/8222478/#!/](http://space.bilibili.com/8222478/#!/)。而对于知乎，这个值藏的比较隐秘。后来我发现，当一个用户关注你之后，便会邮件一个关注者链接给你，例如 [http://www.zhihu.com/l/G4gQn](http://www.zhihu.com/l/G4gQn)，最后是一个 16 进制的五位数，大部分用户都是 G，E 等开头的数字。这种链接在知乎正式页面没有发现过，猜测是历史原因保留了这种形式。

自增数字的方式程序更容易实现。

## 换代理的痛点

剩下的问题就是反爬虫，简单的换 header 头部就不说了，因为 B 站和知乎都是根据 IP 来反爬，那么我能想到的就只有：

* 换代理
* 分布式爬虫

从实现方式看，代理也能看作是一个分布式爬虫。根据 HTTP 返回的状态码，可以判断：

* 代理是否正常工作（比如连接 timeout），
* 对方服务器是否判定我们访问过于频繁（比如返回 403）。

基于上述判断，可以实现一些复杂的代理切换策略：

* 定时去免费代理网站下载新的代理节点，
* 负载均衡，将请求平均分配到每一个节点，保证不会访问过于频繁，
* 代理探活，将不活跃的代理清除。

不幸的是，免费好用的代理非常难找，使用 scrapy 配上[我写的几百行的代理中间件](https://github.com/lyyyuna/bilibili_papapa/commit/632a8827c187aa051186089d715616db8ae7fd86)，也很难维持一个正常的下载速率。
用 Python 来爬虫几乎是网上教程的标配了 -_-，导致即使找到一个可用代理，其 IP 也早被列入服务器的黑名单。

## 分布式爬虫

### 基于 HTTP 服务器实现分布式

![基于 HTTP 服务器的分布式爬虫](https://github.com/lyyyuna/blog_img/raw/master/blog/201704/http_structure.png)

* 主从之间通信借助现有的 HTTP 协议。
* 主控 HTTP 服务器生成用户 id，每当有一个新的 GET 请求，自增 id，并返回给客户端。
* 每一个爬虫在这个结构中是一个客户端，当获取到新的 id 后，便去爬取解析网页，将结果以 POST 请求的方式返回给 HTTP 服务器。

优点是：

* HTTP 协议简单，用 Flask 框架 5 分钟即能快速搭建。
* 客户端数量扩展极其方便，且不需要公网 IP。
* 爬虫下载速率可由每个客户端各自控制。

缺点是：

* 由于 HTTP 是无状态协议，服务器端不能跟踪一个发布的 id 是否已被正确处理。有可能一个客户端只 GET 而不 POST。需要维护另外的队列来存放未完成的任务。
* HTTP 服务器发布任务是被动的。

这个链接[https://github.com/lyyyuna/zhihu_user_papapa](https://github.com/lyyyuna/zhihu_user_papapa)里是用这种思路实现的 B 站用户爬虫。

### 基于消息队列实现分布式

![基于消息队列实现分布式爬虫](https://github.com/lyyyuna/blog_img/raw/master/blog/201704/messagequeue_structure.png)

采用消息队列后，Producer 主控侧不再是被动地发布任务，而是主动推送任务。上一小节有的优点消息队列同样拥有，同时

* Producer 也能控制任务发布的速率。
* 利用消息队列的持久化能力，可以在意外退出的情况下，记录未能成功发布的任务和未能成功接收的结果。

这种结构的分布式爬虫，同样需要显示地维护一来一回蓝色红色的数据流，还是稍显复杂。

### 基于 Celery 任务队列实现分布式

![基于 Celery 实现分布式爬虫](https://github.com/lyyyuna/blog_img/raw/master/blog/201704/celery_structure.png)

初看上去，这个结构和上一节没有区别，但是，上图的红色数据流是不需要显示维护的！

举一个简单例子。假设在机器 A，有如下的 worker

```python
from celery import Celery

app = Celery('tasks', broker='pyamqp://guest@localhost//')

@app.task
def add(x, y):
    return x + y
```

那么在另一台机器 B，只要运行

```python
from tasks import add
result = add.delay(4, 4)
```

就能得到结果。同时

* 这个被 @app.task 修饰过的方法是异步的。机器 A 可以通过 result.ready() 来来获知任务是否被执行完。通过 result.get() 得到执行的结果。

基于这个思想，我完成了这个分布式爬虫[https://github.com/lyyyuna/bilibili_papapa](https://github.com/lyyyuna/bilibili_papapa)。只需要 Celery 的 broker 具有公网 IP，然后把程序扔给朋友让他们帮我跑就行。唯一的不便是 Celery worker 在个人电脑上部署不便，而基于 HTTP 的分布式爬虫，我只要 C# 写个 HTTP 客户端生成 exe 就可以了。

