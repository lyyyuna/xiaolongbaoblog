title: DHT 公网嗅探器实现（DHT 爬虫）
date: 2016-05-14 09:02:03
categories: 网络
tags: 
- DHT
summary: 做一个自己的种子爬虫。
series: BitTorrent 协议实现小记
---


## 前言

这里实现的 DHT 嗅探器会在公网捕获种子的 infohash，源码见 [GitHub](https://github.com/lyyyuna/DHT_sniffer)。

DHT 协议介绍见 [DHT 协议 - 译](http://www.lyyyuna.com/2016/03/26/dht01/)。

其实大部分代码来自了 [simDHT](https://github.com/Fuck-You-GFW/simDHT/blob/master/simDHT.py)，我只是改成了 gevent 版本。

为什么我没有用 Python 3 的 asyncio 呢？最大的原因是我没有在官方 API 中找到方便使用 UDP 协议的接口。asyncio 底层有着类似 Twisted 的事件驱动编程，对于 TCP 协议，官方又封装了一层 Stream，可以用 await 类似同步的方式异步编程。但不知道为啥就是没有 UDP 的封装，见 [谷歌讨论组](https://groups.google.com/forum/#!topic/python-tulip/xYgQRXkb83g)，Guido van Rossum 他自己觉得不需要。然而事件编程我不习惯，在这个项目里要用违反直觉的方式封装，所以还是放弃了 asyncio。

## gevent 中的 UDP

使用 gevent，对于原程序改动很小。使用 DatagramServer 作为 UDP server。

```python
from gevent.server import DatagramServer

class DHTServer(DatagramServer): 
     def __init__(self, max_node_qsize, bind_ip): 
         s = ':' + str(bind_ip) 
         self.bind_ip = bind_ip 
         DatagramServer.__init__(self, s) 
```


## 协议相关

### 如何捕获 infohash

虽是嗅探，但 DHT 的流量不会无缘故的跑过来，必须把自己伪装成一个客户端才能捕获到 DHT 流量。

以下设我们伪造的客户端为 X。

1. 首先向其他 node 发送 find_node 请求，所发送的 node 可以随机生成，我们的目的只是为了让对方 node 能在其路由表中记录下我们伪造的 X。
2. 当其他 node 想要下载 torrent 时，便会向其路由表中最近的 node 依次发送 get_peers/announce_peer 请求。这样其必会向我们伪造的 X 发送 get_peers/announce_peer 请求，即包含真实的 infohash。

这样，一个真实的 infohash 就到手了。总结就是，不断和其他 node 交朋友，然后等着他们发送 infohash 过来。

既然是伪造，意味着不需要实现完整的 DHT 协议。只需要 find_node 和 get_peers/announce_peer 请求即可。

### 路由表

有必要实现完整的路由表吗？协议中的路由表需要维护一个较复杂的数据结构，监控一个 node 的健康程度，若不活跃则需将其删除。作为 DHT 嗅探器这是多余的，因为原数据结构是要保证路由表中是健康的节点，以提高下载速度。而我们是为了认识更多的节点，向 node 发送一次 find_node 请求之后，即可删除该数据。

队列，更符合伪造的客户端对 node 管理的需求。

    self.nodes = deque(maxlen=max_node_qsize)
    
其他 node 除了会发送 get_peers 请求来获取 torrent 之外，也会发送 find_node 来获取节点信息。这意味着，其他 node 的 find_node 请求便是我们更新 node 信息的来源。

```python
    def process_find_node_response(self, msg, address):
        # print 'find node' + str(msg)
        nodes = decode_nodes(msg["r"]["nodes"])
        for node in nodes:
            (nid, ip, port) = node
            if len(nid) != 20: continue
            if ip == self.bind_ip: continue
            if port < 1 or port > 65535: continue
            n = KNode(nid, ip, port)
            self.nodes.append(n)
```

我们伪造的 X 从队列中取出一个 node，然后发送 find_node，

```python
    def auto_send_find_node(self):

        wait = 1.0 / self.max_node_qsize / 5.0
        while True:
            try:
                node = self.nodes.popleft()
                self.send_find_node((node.ip, node.port), node.nid)
            except IndexError:
                pass
            sleep(wait) 
```
            
实际测试表明，运行时队列一直是满的，所以不用担心 node 不够用。

### 解析 infohash

get_peers 和 announce_peer 请求都含有 infohash。虽然 announce_peer 请求质量更高，但数量少。

```python
    def on_get_peers_request(self, msg, address):
        try:
            infohash = msg["a"]["info_hash"]
            tid = msg["t"]
            nid = msg["a"]["id"]
            token = infohash[:TOKEN_LENGTH]
            print 'get peers: ' + infohash.encode("hex"),  address[0], address[1]
            msg = {
                "t": tid,
                "y": "r",
                "r": {
                    "id": get_neighbor(infohash, self.nid),
                    "nodes": "",
                    "token": token
                }
            }
            self.send_krpc(msg, address)
        except KeyError:
            pass

    def on_announce_peer_request(self, msg, address):
        try:
            print 'announce peer'
            infohash = msg["a"]["info_hash"]
            token = msg["a"]["token"]
            nid = msg["a"]["id"]
            tid = msg["t"]

            if infohash[:TOKEN_LENGTH] == token:
                if msg["a"].has_key("implied_port") and msg["a"]["implied_port"] != 0:
                    port = address[1]
                else:
                    port = msg["a"]["port"]
                    if port < 1 or port > 65535: return
                print 'announce peer: ' + infohash.encode("hex"),  address[0], port
        except Exception:
            pass
        finally:
            self.ok(msg, address)
```

对着 [DHT 协议 - 译](http://www.lyyyuna.com/2016/03/26/dht01/) 很容易看懂。

### 嗅探器启动

看到这里，你会发现嗅探器启动时，队列是空的。所以必须先放几个已知的 node。

```python
BOOTSTRAP_NODES = (
    ("router.bittorrent.com", 6881),
    ("dht.transmissionbt.com", 6881),
    ("router.utorrent.com", 6881)
)

    def join_DHT(self):
        for address in BOOTSTRAP_NODES:
            self.send_find_node(address)
```

启动之后，队列就会被其他 node 发送的 find_node 所填满。

## 结果

公网捕获，一小时 10000 个左右。

![效果图](/img/blog/201605/psb.jpg)