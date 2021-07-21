title: 使用 Python 实现一个简单的 HTTP 代理 - GET
date: 2016-01-16 19:44:50
categories: 网络
tags:
- http proxy
- Python script
---


## 前言

HTTP 代理是位于服务器与客户端之间的中间实体，在各个端点之间来回传送 HTTP 报文。

按照用途分类，HTTP proxy 可以分为

* 内容过滤。
* 科学上网。
* Web 缓存。维护服务器常用文档的一个副本，增加客户端的访问速度。
* 反向代理。反向代理可以接收发给服务器的真实请求，然后按需交给真实的服务器。类似于路由的功能。
* ...

按照代理对客户端的可见性又可以分为

* 透明代理
* 非透明代理

本文要实现一个简单的 HTTP 非透明代理，暂时只支持 GET 请求的转发，且不追求性能和错误处理。

## HTTP 非透明代理下 GET 请求的不同

除了部分请求头和 URL 之外，在非透明代理下，浏览器发给服务器和发给代理的报文是完全一样的。

1. 首先浏览器会向非透明代理发送完整的绝对路径 URL。而在普通情况下只会发送相对路径，不需要主机名。
2. 浏览器会用 Proxy-Connection 头部代替 Connection 头部。

### 为何要用绝对路径

早期的 HTTP 设计中，客户端只会与单个服务器进行通信，所以一旦 TCP 连接建立起来以后，只需要相对路径。

但代理就有问题，客户端首先和代理建立 TCP 连接，但由于传递的请求头中使用相对路径，代理就不知道使用什么 IP 和 端口来向远端的服务器建立 TCP 连接。所以，对于早期的 HTTP 1.0，强制客户端发给代理时使用完整路径，如

    GET http://www.douban.com/ HTTP /1.0
    
较新的 HTTP 1.1 则规定了必须包含 Host 头部。所以对于 HTTP 1.1 的代理来说，完整 URL 不是必须的。但由于网络上还有大量旧版代理，Host 头部代理或许根本不识别，所以现在浏览器在使用代理时，还是会使用完整 URL。

以下是用 nc 监听 8888 端口，火狐配置 nc 8888 为代理时，所发送 GET 请求

    root@yan:~# nc -lvp 8888
    listening on [any] 8888 ...
    Warning: forward host lookup failed for promote.cache-dns.local: Unknown host
    connect to [192.168.27.128] from promote.cache-dns.local [192.168.27.1] 29798
    GET http://www.douban.com/ HTTP/1.1
    Host: www.douban.com
    User-Agent: Mozilla/5.0 (Windows NT 10.0; WOW64; rv:42.0) Gecko/20100101 Firefox/42.0
    Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
    Accept-Language: zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3
    Accept-Encoding: gzip, deflate
    Cookie: bid="UwoBYPAPsOE"; ll="118159"; __utma=30149280.478845192.1452777657.1452777657.1452777657.1; __utmz=30149280.1452777657.1.1.utmcsr=(direct)|utmccn=(direct)|utmcmd=(none); _pk_id.100001.8cb4=d15426da1038582e.1452777681.1.1452777681.1452777681.
    Connection: keep-alive
    Cache-Control: max-age=0

可以看出使用的是完整路径。

### 为什么要用 Proxy-Connection 头部

Connection 头部是为了减少建立 TCP 连接的次数，复用连接产生的。默认 HTTP 1.1 是 Keepalive，但 1.0 的代理则不识别此头部。对于不认识的头部，代理会直接转发，以保持向后兼容性。

假如 Connection: Keep-alive 发给了代理，代理不识别转发给了服务器，而恰巧服务器识别此头部，便会出现严重问题。服务器和浏览器都保持连接，而代理则中断了连接。

为解决这个问题，出现了一个新的头部 Proxy-Connection。如果 1.1 的代理，代理会改写为 Connection 头部。如果 1.0 的代理，那么会直接转发此头部，服务器发现 Proxy-Connection 后，就会采用非长连接的方式。

Proxy-Connection 头部我实际实验的时候，并不是每个浏览器都会发送。同样是火狐，家里电脑直接发送了 Connection，而单位电脑则是 Proxy-Connection。

## 实现

我们要做的有

* 替换完整路径为相对路径
* 去掉 Proxy-Connection
* 将 Connection 头部改为 close （为了简单起见）

我采用 BaseHTTPServer 和 BaseHTTPRequestHandler 来处理浏览器发送的 GET 请求，原始套接字 socket 来转发请求给远端服务器，既没有多线程也没有 IO 复用。采用 urllib 库来解析完整路径 URL，并取出相对路径。

代码如下

```python
    from BaseHTTPServer import BaseHTTPRequestHandler, HTTPServer
    import socket
    import urllib

    class MyHandler(BaseHTTPRequestHandler):

        def do_GET(self):
            uri = self.path
            # print uri
            proto, rest = urllib.splittype(uri)
            host, rest = urllib.splithost(rest)
            # print host
            path = rest        
            host, port = urllib.splitnport(host)
            if port < 0:
                port = 80
            # print host
            host_ip = socket.gethostbyname(host)
            # print port

            del self.headers['Proxy-Connection']
            self.headers['Connection'] = 'close'

            send_data = 'GET ' + path + ' ' + self.protocol_version + '\r\n'
            head = ''
            for key, val in self.headers.items():
                head = head + "%s: %s\r\n" % (key, val)
            send_data = send_data + head + '\r\n'
            # print send_data
            so = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            so.connect((host_ip, port))
            so.sendall(send_data)

            # 因为采用非长连接，所以会关闭连接， recv 会退出
            data = ''
            while True:
                tmp = so.recv(4096)
                if not tmp:
                    break
                data = data + tmp

            # socprint data
            so.close()

            self.wfile.write(data)


        # do_CONNECT = do_GET

    def main():
        try:
            server = HTTPServer(('', 8888), MyHandler)
            print 'Welcome to the machine...'
            server.serve_forever()
        except KeyboardInterrupt:
            print '^C received, shutting down server'
            server.socket.close()

    if __name__ == '__main__':
        main()
```