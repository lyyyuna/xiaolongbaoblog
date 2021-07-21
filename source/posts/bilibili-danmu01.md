title: B站直播弹幕协议详解
date: 2016-03-14 21:20:36
categories: 网络
tags: bilibili
---

## 前言

B站直播弹幕是在 flash 里做的，所以浏览器的开发工具抓不到包。去年我根据 Wireshark 抓的 TCP 包硬写了一个[直播终端版](https://github.com/lyyyuna/script_collection/blob/master/bilibli_danmu/11.py)，不过当时对里面很多二进制位不知所云。

上个星期我发现B站官方有专门为 UP 主准备的直播弹幕姬，且是用 C# 写的，于是就逆向了。。

这里是最新的 [B 站直播弹幕姬 Python 版](https://github.com/lyyyuna/bilibili_danmu)。

本篇将对直播协议做一个完整介绍。

* B站协议和端口是会变的（至少改变过一次），故不能保证本篇所述有向后兼容性。
* 有兴趣去逆向B站直播弹幕姬的朋友，不要使用 ILSpy debug 版本，因为有 C# 5, async/await，debug 版本看起来会非常凌乱。

## 总体过程

直播协议首先需要获取 RoomId，获取服务器地址，然后与 B 站服务器建立 TCP 链接并保持，之后就处于监听状态，B 站会连续地将弹幕消息推送过来。

    Get RoomId (HTTP) --> Get server address (HTTP) --> Open TCP & wait --> parse json
                                                               |                |
                                                               \ -------------- /
                                                                                                           
其中获取 RoomId 有些费解，一般我们认为一个 UP 主的房间 url 若为 http://live.bilibili.com/44515，则其 RoomId 就应该为 44515，但其实部分 UP 主的 url 和其 RoomId 并不对应。比如神奇陆夫人，其 url 显示为 115，但实际上 RoomId 为 1016。真正的 RoomId 需要去 http://live.bilibili.com/115 的 html 中寻找。

    <script>
        var ROOMID = 1016;
        var DANMU_RND = 1457957537;
        var NEED_VIDEO = 1;
        var ROOMURL = 115;
    </script>
    
这些特别的 UP 主经测试 url 都在 100 左右，猜测他们可能是 B 站老用户，或者他们向 B 站申请过特殊 url。

### 房间连接过程

获取 RoomId 之后，紧接一个 HTTP 请求 'http://live.bilibili.com/api/player?id=cid:RoomId'，获取 xml 数据,

    <uid>0</uid>
    <uname></uname>
    <login>false</login>
    <isadmin>false</isadmin>
    <time>1457954601</time>
    <rank></rank>
    <level></level>
    <chatid>1016</chatid>
    <server>livecmt-1.bilibili.com</server>
    <user_sheid_keyword></user_sheid_keyword>
    <sheid_user></sheid_user>
    <block_time>0</block_time>
    <block_type>0</block_type>
    <state>LIVE</state>

其中 livecmt-1.bilibili.com 为直播弹幕地址。

然后向 livecmt-1.bilibili.com:788 开一个 TCP 链接，之后所有的弹幕和心跳包都会发生在此链接上。

进入直播间需要发送如下的数据包：

    00000000  00 00 00 35 00 10 00 01  00 00 00 07 00 00 00 01 ...5.... ........
    00000010  7b 22 72 6f 6f 6d 69 64  22 3a 31 30 31 36 2c 22 {"roomid ":1016,"
    00000020  75 69 64 22 3a 31 35 35  39 37 33 36 38 35 37 32 uid":155 97368572
    00000030  38 31 36 30 7d                                   8160}

0x35 是一次数据包的长度，0x00100001 不详，0x07 代表请求进入直播间，0x00000001 不详。
后面跟了一串 json 数据，uid 为客户端随机生成，算法如下：

    (int)(100000000000000.0 + 200000000000000.0*random.random())
    
如果进入房间成功，则会返回

    00 00 00 10 00 10 00 01  00 00 00 08 00 00 00 01
    
目前看来并没有其他特殊含义。

### 消息种类

如果接收到的数据包为

    00000010  00 00 00 14 00 10 00 01  00 00 00 03 00 00 00 01 ........ ........
    00000020  00 00 3c 49                                      ..<I

其中 0x14 为一次数据包的长度，0x0010001 不详，0x03 代表这是一个在线人数数据包，0x00000001 不详，0x3c49 = 15433 为在线人数。根据逆向的源码显示，0x03,0x02,0x01 都代表在线人数数据包，不过我只抓到了 0x03 这一种。


如果接收到的数据包为

    00000825  00 00 00 d4 00 10 00 00  00 00 00 05 00 00 00 00 ........ ........
    00000835  7b 22 69 6e 66 6f 22 3a  5b 5b 30 2c 31 2c 32 35 {"info": [[0,1,25
    00000845  2c 31 36 37 37 37 32 31  35 2c 31 34 35 37 39 35 ,1677721 5,145795
    00000855  38 33 37 34 2c 22 31 34  35 37 39 35 35 36 35 35 8374,"14 57955655
    00000865  22 2c 30 2c 22 61 66 61  39 66 37 32 64 22 2c 30 ",0,"afa 9f72d",0
    00000875  5d 2c 22 e7 81 ab e6 8a  8a e8 80 90 e4 b9 85 ef ],"..... ........
    00000885  bc 88 30 32 ef bc 89 22  2c 5b 36 31 34 39 34 33 ..02..." ,[614943
    00000895  2c 22 e8 8d 92 e5 b7 9d  e5 90 b9 e6 b0 b4 22 2c ,"...... ......",
    000008A5  30 2c 30 2c 30 5d 2c 5b  39 2c 22 e7 b2 be e8 8b 0,0,0],[ 9,".....
    000008B5  b1 22 2c 22 e7 a5 9e e5  a5 87 e9 99 86 e5 a4 ab .",".... ........
    000008C5  e4 ba ba 22 2c 31 31 35  2c 31 32 32 32 35 32 35 ...",115 ,1222525
    000008D5  35 5d 2c 5b 32 31 2c 32  35 38 32 37 5d 2c 5b 5d 5],[21,2 5827],[]
    000008E5  5d 2c 22 63 6d 64 22 3a  22 44 41 4e 4d 55 5f 4d ],"cmd": "DANMU_M
    000008F5  53 47 22 7d                                      SG"}
    
这是一个弹幕数据包，0xd4 为数据包的长度，0x00100000 不详，0x05 代表这是弹幕，0x00000000 不详。之后就是 json 格式的弹幕消息。

### 弹幕种类

跟据 json['cmd'] 的值，可分为：

* LIVE 直播中
* PREPARING 准备中
* DANMU_MSG 弹幕消息
* SEND_GIFT 赠送礼物消息
* WELCOME 欢迎某人进入直播间 

对于 DANMU_MSG，

* json['info'][1] 为弹幕消息主体，为 utf-8 编码
* json['info'][2][1] 为发送者昵称，为 utf-8 编码
* json['info'][2][2] == '1' 是否是管理员
* json['info'][2][3] == '1' 是否是 VIP

对于 SEND_GIFT，

* json['data']['giftName'] 为礼物名称
* json['data']['uname'] 为发送者昵称
* json['data']['rcost'] 不详
* json['data']['num'] 为礼物数目

对于 WELCOME，

* json['data']['uname'] 为新进入直播间的用户昵称

### 心跳包

心跳包和进入直播间所发送的包仅有些许不同

    00 00 00 10 00 10 00 01  00 00 00 02 00 00 00 01
    
把 0x07 变为 0x02 即可。心跳间隔为 30s。


## 结论

以下是我登陆 C菌 直播间抓取的弹幕消息。

![效果图](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201603/bilibili.png)