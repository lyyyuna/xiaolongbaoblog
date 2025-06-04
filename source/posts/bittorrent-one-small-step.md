title: BitTorrent 一点进展
date: 2025-05-25 10:56:33
series: BitTorrent 协议实现小记

---

过去三四个月又是零零碎碎的在实现这个协议，目前有了[小的进展](https://github.com/lyyyuna/torrent-cli/releases/tag/v0.0.1)。下面来介绍一下。

实现过程不免参考一些开源代码，大家可在上一篇文章中找到些出处，这里不一一列举，需要指出的是，参考并不是简单的 copy，因为很多开源的代码架构风格我并不喜欢，基本都做了重写。

## bencode 编解码

这次尽量实现所有细节，所以 bencode 的编解码没有用现成库。好在 bencode 本身简单，源码不到 150 行：[https://github.com/lyyyuna/torrent-cli/blob/v0.0.1/zhongzi/bencode.py](https://github.com/lyyyuna/torrent-cli/blob/v0.0.1/zhongzi/bencode.py)。

Bencoding 支持四种数据类型：
1. 字符串
    * 格式：<长度>:<内容>
    * 例如：`4:spam` 表示字符串 `"spam"`。
2. 整数
    * 格式：i<数字>e
    * 例如：`i42e` 表示整数 `42`，`i-3e` 表示 `-3`。
3. 列表
    * 格式：l<元素1><元素2>...e
    * 例如：`l4:spami42ee` 表示列表 `["spam", 42]`。
4. 字典
    * 格式：d<键1><值1><键2><值2>...e
    * 键必须是字符串，且按字典序排列。
    * 例如：`d3:foo3:bar5:helloi52ee` 表示字典 `{"foo": "bar", "hello": 52}`。

## torrent 文件

torrent 文件是一个 bencode 编码的对象。其中关键的信息如下(这里换用 json 来表示)：

单文件：

```json
{
    "announce" : "http://torrent.ubuntu.com:6969/announce",
    "info" : {
        "name" : "ubuntu-14.04.6-server-amd64.iso",
        "length" : 662700032,
        "piece length" : 524288,
        "pieces" : xxxxx
    }
}
```

* `announce` 是 http tracker 的地址，客户端可以通过 tracker 拿到 peers 信息。
* `length` 是整个文件的长度。

bittorrent 把被做种的文件分成了多个 piece，除了最后一个 piece，每个 piece 的长度是固定的 `piece length`。

* `pieces` 是一串二进制，可按照 20 个字节切割成一个二进制字符串列表。按次序，每 20 个字节是对应序号的 piece 的 sha1。

每个种子还有个重要属性是 `info_hash`，没有记录在 torrent 文件中，而是直接对 torrent 文件内容计算 sha1。

多文件：

```json
{
    "announce" : "http://torrent.ubuntu.com:6969/announce",
    "info" : {
        "name" : "ubuntu-14.04.6-server-amd64.iso",
        "piece length" : 524288,
        "files" : [
            {
                "path" : "/a.mp4",
                "lenth" : 12334,
            },
            {
                "path" : "/g.mp4",
                "lenth" : 23423,                
            }
        ],
        "pieces" : xxxxx
    }
}
```

多文件没有 `length` 属性，取而代之的是 `files` 列表，记录每个文件的名称和大小。`files` 列表的顺序是重要的，所有文件按次序拼接在 `pieces` 中。

源码见[https://github.com/lyyyuna/torrent-cli/blob/v0.0.1/zhongzi/torrent.py](https://github.com/lyyyuna/torrent-cli/blob/v0.0.1/zhongzi/torrent.py)。

### piece 信息

从 torrent 文件角度看，文件的最小单元是 piece，但 [BEP-3](https://bittorrent.org/beps/bep_0003.html) 指出，下载的最小单元其实是 block，block 一般为 `2**14` 大小。为了后续下载方便，我实现的时候也为 piece 构建了虚拟的 block 列表：

```py
class Piece:
    def __init__(self, index: int, length: int, offset: int, checksum: bytes):
        self.index = index
        self.length = length
        self.offset = offset
        self.checksum = checksum

        self.blocks: List[Block] = []

class Block:
    def __init__(self, index: int, offset: int, length: int):
        self.index = index
        self.offset = offset
        self.length = length
```

## 是否用 tracker？

很多网上的 bittorrent 教程/开源项目，都用 torrent 中的 `tracker` 来获取 `peers`。

但我实际测试下来，这种方式获取的 `peers` 质量太差，不是数量太少，就是对端所持有的 `piece` 不全，导致根本无法下载。大伙可以看看这个挑战 [Build your own BitTorrent](https://app.codecrafters.io/courses/bittorrent/overview) 中的发帖，很多人卡在这一步。

源码见[https://github.com/lyyyuna/torrent-cli/blob/v0.0.1/zhongzi/tracker.py](https://github.com/lyyyuna/torrent-cli/blob/v0.0.1/zhongzi/tracker.py)。实际客户端并未使用这个模块。

## 用 DHT 吧

[BEP-5](https://bittorrent.org/beps/bep_0005.html) 描述的 DHT 协议是另一种获取和交换 `peers` 信息的方式，它主要有两大部分：

* 路由表 `routing table`
* krpc 协议

### 基于 Kademlia 的 routing table 路由表

`routing table` 是每个节点用来维护网络中其他节点信息的数据结构，其核心作用是**高效定位目标节点或数据**。不同的 DHT 实现（如 Kademlia、Chord、Pastry）有各自的路由表设计，但核心逻辑相似：

1. 快速查找：帮助节点确定如何将查询请求（如查找数据或节点）逐步转发到目标节点。
2. 网络拓扑维护：动态记录其他节点的联系信息（如 IP、端口、Node ID），保证网络的连通性。
3. 减少跳数：通过分层或分桶的优化设计，提高查询性能。

[BEP-5](https://bittorrent.org/beps/bep_0005.html) 官方用的是 Kademlia 算法。Python 有一个完整的 Kademlia 实现 [https://github.com/bmuller/kademlia](https://github.com/bmuller/kademlia)，它完全等价地实现了论文[https://pdos.csail.mit.edu/~petar/papers/maymounkov-kademlia-lncs.pdf](https://pdos.csail.mit.edu/~petar/papers/maymounkov-kademlia-lncs.pdf)中提到的算法。其中作者用 [https://github.com/bmuller/rpcudp](https://github.com/bmuller/rpcudp) 作为底层通信协议，如果读者想移植到自己的 bittorrent 客户端中，需要做些修改。

接着介绍 Kademlia，其路由表是一个**二叉树分桶结构**，按节点 ID 的**异或距离**（XOR Distance）组织：

1. Node ID：每个节点有一个唯一标识（如 160 位哈希值）。
2. 异或距离：两个节点 ID 的异或结果（a⊕b）表示它们的逻辑距离。
3. K-Buckets：
    * 路由表被划分为多个桶（Bucket），每个桶负责一定范围的异或距离。
    * 每个桶最多维护 k 个节点（如 k=20），按最近活跃顺序排列（LRU 策略）。
    * 例如，对于 160 位 ID，路由表可能有 160 个桶，第 i 个桶存储距离在 `[i**2, i**(2+1))` 范围内的节点。

`routing table` 涉及以下操作：

1. 插入节点：
    * 计算新节点与自己的异或距离，找到对应的 K-Bucket。
    * 如果桶未满，直接插入；如果已满，剔除最久未响应的节点（防止攻击）。
2. 查找节点：
    * 从路由表中选择与目标 ID 异或距离最近的 k 个节点，向它们发起查询。
3. 动态更新：
    * 节点定期 Ping 桶中的节点，移除失效节点。
    * 新遇到的节点会被加入合适的桶中（保持网络新鲜度）。

### 简化的 routing table

如果觉得 [Kademlia](https://github.com/bmuller/kademlia) 实现非常复杂，你可以参考 [https://github.com/bashkirtsevich-llc/aiobtdht/tree/master/src/aiobtdht/routing_table](https://github.com/bashkirtsevich-llc/aiobtdht/tree/master/src/aiobtdht/routing_table) 和我的 [https://github.com/lyyyuna/torrent-cli/blob/main/zhongzi/dht/routing_table.py](https://github.com/lyyyuna/torrent-cli/blob/main/zhongzi/dht/routing_table.py) 实现。

之所以 [BEP-5](https://bittorrent.org/beps/bep_0005.html) 要搞一个复杂 Kademlia，是因为它能在超大网络节点数下，仍能保持 `O(logN)` 的查询效率。而我们目前所要实现的只是客户端：

1. 不像服务端那样需要长时间维护超大网络节点数信息。一次下载，有几百个节点信息就足够了。
2. 查询 `peers` 不频繁。一次下载，只需要有十来个 `peers`，当有效 `peers` 数量减少，再做一次查询即可。

这个策略下，`routing table` 甚至可以更简单，直接用一个 LRU list 存储即可，查找最近节点时，遍历 list 做异或距离计算。如果 list 只存储 1000 个以内的节点信息，遍历带来的时间复杂度是完全可以接受的。

P.S. 这只是我实现的客户端的策略。这个策略能保证在家庭网络中，下载 ubuntu.iso 达到 `3-4 MB/s` 的速度。

### krpc 协议

krpc 是基于 udp 的轻量级 RPC 协议，它主要有 4 种 RPC 方法：

| **方法**	 | **用途** | **参数** | **返回值** |
|-------|-------|-------|-------|
| `ping`  |   检查节点是否在线	 |  `{"id": "<sender_node_id>"}`  |  `{"id": "<responder_node_id>"}`  |
| `find_node`  |   查找离 `target` 最近的节点		|  `{"id": "<sender_id>", "target": "<target_node_id>"}`  |  `{"id": "<responder_id>", "nodes": "<compact_node_info>"}`  |
| `get_peers`  |   查找某个 `info_hash` 的 `peers`	|  `{"id": "<sender_id>", "info_hash": "<torrent_hash>"}`  |  `{"id": "<responder_id>", "values": ["<peer1>", ...]}` 或 `{"nodes": "<compact_node_info>"}`  |
| `announce_peer`  |   声明自己拥有某个 `info_hash`	|  `{"id": "<sender_id>", "info_hash": "<torrent_hash>", "port": <port>, ...}`  |  `{"id": "<responder_id>"}`  |

客户端主要使用 `find_node` 和 `get_peers`：

1. 用 `find_node` 尽可能填充 `routing table`。
2. 用 `get_peers` 获取 `info_hash` 对应的 `peers`。 

思路和我之前实现的[DHT 公网嗅探器实现（DHT 爬虫）](https://www.lyyyuna.com/2016/05/14/dht-sniffer/)类似，需要：

```py
self._bootstrap_nodes = [
    ("67.215.246.10", 6881),  # router.bittorrent.com
    ("87.98.162.88", 6881),  # dht.transmissionbt.com
    ("82.221.103.244", 6881)  # router.utorrent.com
]
```

作为初始节点，通过它们来启动第一次 `find_node`。

### 基于 Python asyncio 的 udp rpc 服务如何设计

Python 标准库提供了 `asyncio.DatagramProtocol` 来实现 asyncio udp rpc 服务，主结构如下：

```py
class KRPCProtocol(asyncio.DatagramProtocol):
    # 父类调用
    def connection_made(self, transport):
        self.transport = transport

    # 父类调用
    def datagram_received(self, data, addr):
        pass

    # 使用者调用
    async def call_rpc(self, msg, addr):
        self.transport.sendto(msg, addr)

# 使用
# 初始化
loop = asyncio.get_event_loop()
_, protocol = await loop.create_datagram_endpoint(
    lambda: KRPCProtocol(self._ids),
    local_addr=self.bind
)

# 发送请求，并等待数据
await protocol.call_rpc()
```

1. [BEP-5](https://bittorrent.org/beps/bep_0005.html) 指出，krpc 的每一对请求与响应之间通过 `tid` 来串联。
2. `call_rpc` 是由使用者调用，发送 udp rpc 请求后阻塞，直到对应的 udp rpc 响应返回。
3. `datagram_received` 是父类 `asyncio.DatagramProtocol` 在收到 udp 消息后触发的回调。

我们可以通过 `asyncio.Future` 来实现上述的异步阻塞功能：

1. 维护一个 futures 字典，key 是已发送的 `tid`。
2. `self.transport.sendto` 之后创建一个 Future 对象，存入 futures 字典。
3. 对应的 `call_rpc` 阻塞于 Future 对象上。
4. 收到 udp 裸消息后，解析其中 `tid`，查询 futures 字典，取出 Future 对象标记 done。
5. 对应的 `call_rpc` 唤醒。

代码大致为：

```py
class KRPCProtocol(asyncio.DatagramProtocol):
    def __init__(self, node_id: bytes=None):
        self.futures: Dict[bytes, asyncio.Future] = {}
        self.transaction_id = 1

    def datagram_received(self, data, addr):
        msg = decode(data)
        tid: bytes = msg[b't']

        future = self.futures.pop(tid)
        future.set_result(msg)

    async call_rpc(self, msg, addr):
        future = asyncio.Future()
        tid = self._get_transaction_id()       
        self.futures[tid] = future

        self.transport.sendto(msg, addr)

        return await asyncio.wait_for(future, timeout=timeout)
```

## Peer 交互过程

官方的 [BEP-3](https://bittorrent.org/beps/bep_0003.html) 关于消息介绍的很简单，大家可以参考 [https://wiki.theory.org/BitTorrentSpecification#Messages](https://wiki.theory.org/BitTorrentSpecification#Messages) 去了解详细的消息用途。

### 消息的解析

消息遵循固定格式的前缀

```
| len(4 bytes) | id(1 byte) |
```

收到 tcp 流后，每次读取前 5 个字节可判断消息的类型，和消息的边界（长度），解析起来比较简单：

```py
async def parse_one_message(reader: asyncio.StreamReader):
    length_bytes = await reader.readexactly(4)
    length = struct.unpack('>I', length_bytes)[0]
    
    id_bytes = await reader.readexactly(1)
    id = struct.unpack('>b', id_bytes)[0]

    data = await reader.readexactly(length - 1)

    match id:
        case PeerMessage.Choke.value:
            pass
        case PeerMessage.Unchoke.value:
            pass
```

### 连接过程

首先客户端向 peer 建立 tcp 连接后，会进入如下的握手过程：

```
              Handshake
    client --------------> peer    发送 Handshake 请求

              Handshake
    client <-------------- peer    接收 Handshake 响应，比较其中的 info_hash 是否一致

              BitField
    client <-------------- peer    peer 可能发送 BitField，表明自己拥有哪些 piece

             Interested
    client --------------> peer    发送 Interested，表明自己想要下载

              Unchoke
    client <-------------- peer    peer 解除客户端阻塞状态
```

`BitField` 消息是一个二进制长串，其 bit 的个数等于 piece 的个数，而 bit 的次序和文件 piece 的次序一致。如果某 bit 为 1，那表明 peer 拥有对应次序的 piece。如果 peer 在连接过程，逐步获取到了某些新 piece，它也可以通过 `Have` 消息告知本客户端。

* `BitField`，一次性告知文件的 piece 拥有情况。
* `Have`，告知单个 piece 的拥有情况。

`Interested`，`Not Interested`，`Choke`，`Unchoke` 用于控制网络传输，确保 P2P 网络上传/下载的公平，本次实现只把它们作为状态转换的标志位。

`Request` 与 `Piece` 消息成对出现，`index+offset` 相关联，代表了下载请求与响应。用类似上文 udp rpc 中 `asyncio.Future` 对象，即可将这对请求与响应转换成同步方法：

```py
self.futures : Dict[str, asyncio.Future] = {}

async run(self):
    async for msg in self.read():
        match msg:
            case Piece():
                index = msg.index
                offset = msg.begin
                key = f'{index}-{offset}'

                future = self.futures.pop(key)
                if future and not future.done():
                    future.set_result(msg.block)

async def get_piece(self, piece_index: int, offset: int, length: int=2**14) -> bytes:
    data = message.Request(piece_index, offset, length).encode()
    self.writer.write(data)
    await self.writer.drain()

    future = asyncio.Future()
    self.futures[f'{piece_index}-{offset}'] = future

    return await asyncio.wait_for(future)
```

最后的 `Keep Alive` 消息也非常重要，协议规定如果 2 分钟内 peer 没有收到新消息就会断开连接。实际测试发现，有的 peer 下载非常慢，一对 `Request` 与 `Piece` 消息经常超过 2 分钟，容易触发断连，这时候就需要客户端周期性的发送 `Keep Alive` 消息以保活。

## 客户端主流程

讲到这里，一个最小化的 bittorrent 客户端所涉及的关键技术均已介绍，剩下的就是如何组织并设计出一个高并发的下载模型。尽管本文用的是 [Python asyncio](https://github.com/lyyyuna/torrent-cli/blob/v0.0.1/zhongzi/client.py)，你应该很容易切换到 Go 来实现。

主程序由多个协程构成：

1. peers 收集协程：
    * 使用 DHT 协议获取新的 peers，当 peers 数量满足后暂停，按需再次启动。
2. 各 peer 内信息维护协程：
    * 接收 `Bitfield` 和 `Have` 消息，标记各个 peer 拥有哪些 piece。 
3. 下载协程：
    * 解析 torrent 文件，将待下载 piece 塞入 `piece_download_queue` 队列。
    * 启动一组下载协程池，读取 `piece_download_queue` 队列，拿到待下载 piece 后，选择一个有效的 peer 进行下载。
    * 将下载到的 piece 塞入 `piece_saver_queue` 队列。
4. 保存协程：
    * 读取 `piece_saver_queue` 队列，将接收到的 piece 按位置存入文件中。

## 后记

诚如开头所说，现在只是一个小的进展，代码也有很多不完善的地方。

后续会加入多文件、流控、磁力链接的支持，而像上传、NAT 打洞、LSD 等其他 [BEP](https://www.bittorrent.org/beps/bep_0000.html) 协议会慢慢研究。期待最后能实现一个完整的 bittorrent 客户端/服务端。