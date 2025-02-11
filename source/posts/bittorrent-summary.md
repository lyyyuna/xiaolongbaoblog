title: BitTorrent 协议简单分析
date: 2025-02-11 10:56:33
series: BitTorrent 协议实现小记

---

大约十年之前，我分析并实现了部分的 `DHT` 协议，记录在 [DHT 协议 - 译](https://www.lyyyuna.com/2016/03/26/dht01/) 和 [DHT 公网嗅探器实现（DHT 爬虫）](https://www.lyyyuna.com/2016/05/14/dht-sniffer/) 中。去年，我看到了 [Build your own BitTorrent](https://app.codecrafters.io/courses/bittorrent/overview) 的挑战，就想着也实现一个 torrent 下载器，中间来来回回、断断续续，一直没坚持下来。这次春节我仔细做了实验，比较了几个开源的实现，终于能得出一个初步的结论：要想做一个实用的 torrent 下载器，[Build your own BitTorrent](https://app.codecrafters.io/courses/bittorrent/overview) 介绍的远远不够，以我的能力和空闲时间，可能要持续投入一年。

首先是几个开源项目的分析：

`Go` 实现的 [torrent-client](https://github.com/veggiedefender/torrent-client/)，只实现了 [BEP3](https://bittorrent.org/beps/bep_0003.html)，并且还存在错误的假设 Tracker 都会主动返回 Bitfield 响应。但优点是代码结构还算比较清晰，加上 `Go` 的并发比较方便，并发的逻辑不会太干扰主流程的理解。

`Python` 实现的 [pieces](https://github.com/eliasson/pieces)，也是只实现了 [BEP3](https://bittorrent.org/beps/bep_0003.html)，协议交互上没太大问题。问题主要是源码实现有点绕，在块下载管理器和底层协议上代码混在了一起，没有做好分层。这么实现固然能在当时的 `asyncio` 框架上做到比较高效，但我怀疑之后会非常难以扩展其他 BEP 协议？

`Elixir` 实现的 [torrex](https://github.com/ryotsu/torrex)，和上面两个差不多。

总的来说，这几个项目，都不实用：
1. 只实现了 [BEP3](https://bittorrent.org/beps/bep_0003.html) ，用这个协议能找到的 peers 很有限，下载速度提升不上去。
2. 没有考虑多文件的情况。

以上只是我的抱怨，或许我该去参考参考较完善的开源项目:
1. [rain](https://github.com/cenkalti/rain/tree/master/torrent)
2. [torrent](https://github.com/anacrolix/torrent)
3. [Taipei-Torrent](https://github.com/jackpal/Taipei-Torrent)

不过我会尽量先用 `Python` 实现。