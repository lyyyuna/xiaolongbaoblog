title: 实现基于 HTTPS 代理的中间人攻击
date: 2018-03-16 19:44:50
categories: 网络
tags:
- http proxy
---

## 前言

在给产品做 Web 安全测试时，经常会使用代理工具来收集 URL 及相关请求参数。

在我之前的文章介绍了 [使用 Python 实现一个简单的 HTTP 代理](http://www.lyyyuna.com/2016/01/16/http-proxy-get1/)。但这留下一个问题，如何处理 HTTPS 流量？

## HTTP 隧道代理原理

RFC 为这类代理给出了规范，[Tunneling TCP based protocols through Web proxy servers](https://tools.ietf.org/html/draft-luotonen-web-proxy-tunneling-01)。简单来讲就是通过 Web 代理服务器用隧道方式传输基于 TCP 的协议。HTTP 协议正文部分为客户端发送的原始 TCP 流量，代理发送给远端服务器后，将接收到的 TCP 流量原封不动返回给浏览器。

下面这张图片来自于《HTTP 权威指南》，展示了 HTTP 隧道代理的原理。
![HTTP 隧道](/img/blog/201803/connect.png)

浏览器首先发起 CONNECT 请求：

    CONNECT example.com:443 HTTP/1.1

代理收到这样的请求后，依据 host 地址与服务器建立 TCP 连接，并响应给浏览器这样一个 HTTP 报文：

    HTTP/1.1 200 Connection Established

该报文不需要正文。浏览器一旦收到这个响应报文，就可认为与服务器的 TCP 连接已打通，后续可直接透传。

## HTTPS 流量中间人攻击

我们很容易想到，HTTPS 代理本质上就是隧道透传，代理服务器只是透传 TCP 流量，与 GET/POST 代理有本质区别。隧道透传是安全的，代理没有私钥来解密 TLS 流量。

这带来一个问题，现在 HTTPS 越来越普遍，测试时不会特意关掉 TLS，做安全测试也就拿不到 URL 及请求参数。那怎么做呢？

首先是来看正常的隧道代理示意图：

![TLS 示意图 1](/img/blog/201803/tls1.png)

在如图红色的透传流量中，插入我们的**中间人**：

1. 用一个 TLS 服务器伪装成远端的真正的服务器，接下浏览器的 TLS 流量，解析成明文。
2. 用明文作为原始数据，模拟 TLS 客户端向远端服务器转发。

示意图如下：

![TLS 示意图 2](/img/blog/201803/tls2.png)

由于中间人拿到了明文，也就能够继续收集 URL 及相关请求参数。

### 证书问题

大家知道，HTTP 是需要证书的。浏览器会验证服务器发来的证书是否合法。证书若是由合法的 CA 签发，则称为合法的证书。现代浏览器在安装时都会附带全世界所有合法的 CA 证书。由 CA 证书可验证远端服务器的证书是否是合法 CA 签发的。

在 TLS 示意图 2 中，浏览器会验证假 TLS 服务器的证书：

1. 第一验证是否是合法 CA 签发。
2. 第二验证该证书 CN 属性是否是所请求的域名。即若浏览器打开 `www.example.com`，则返回的证书 CN 属性必须是 `www.example.com`。

对于第一点，合法 CA 是不可能为我们签证书的，否则就是重大安全事件了。我们只能自制 CA，并将自制 CA 导入浏览器信任链。

对于第二点，需要自制 CA 实时为域名 `www.example.com` 签一个假的证书。

## Go 实现

不同于之前 [Python 实现的 HTTP 代理](http://www.lyyyuna.com/2016/01/16/http-proxy-get1/)，这次的 HTTPS 中间人代理用 Go 实现。源码见 [https://github.com/lyyyuna/mitm](https://github.com/lyyyuna/mitm)

首先是启动一个 http server。

```go
// mitmproxy.go
func Gomitmproxy(conf *config.Cfg, ch chan bool) {
	tlsConfig := config.NewTLSConfig("gomitmproxy-ca-pk.pem", "gomitmproxy-ca-cert.pem", "", "")
	handler := InitConfig(conf, tlsConfig)
	server := &http.Server{
		Addr:         ":" + *conf.Port,
		ReadTimeout:  1 * time.Hour,
		WriteTimeout: 1 * time.Hour,
		Handler:      handler,
    }
............
	go func() {
		server.ListenAndServe()
		ch <- true
	}()

	return
}
```

`handler` 是一个实现了 `ServeHTTP` 接口的 `Handler`。

```go
func (handler *HandlerWrapper) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method == "CONNECT" {
		handler.https = true
		handler.InterceptHTTPS(resp, req)
	} else {
		handler.https = false
		handler.DumpHTTPAndHTTPS(resp, req)
	}
}
```

根据请求不同分为两大类。普通 GET/POST 请求，由于是明文，可直接进行抓包。而 CONNECT 请求，则走 `InterceptHTTPS`。我们默认走 CONNECT 隧道的都是 HTTPS 流量，其他 TCP 应用层协议则不予考虑。

```go
func (handler *HandlerWrapper) InterceptHTTPS(resp http.ResponseWriter, req *http.Request) {
	addr := req.Host
	host := strings.Split(addr, ":")[0]

    // step 1, 为每个域名签发证书
	cert, err := handler.FakeCertForName(host)
	if err != nil {
		logger.Fatalln("Could not get mitm cert for name: %s\nerror: %s", host, err)
		respBadGateway(resp)
		return
	}

    // step 2，拿到原始 TCP 连接
	connIn, _, err := resp.(http.Hijacker).Hijack()
	if err != nil {
		logger.Fatalln("Unable to access underlying connection from client: %s", err)
		respBadGateway(resp)
		return
	}

	tlsConfig := copyTlsConfig(handler.tlsConfig.ServerTLSConfig)
    tlsConfig.Certificates = []tls.Certificate{*cert}
    // step 3，将 TCP 连接转化为 TLS 连接
	tlsConnIn := tls.Server(connIn, tlsConfig)
	listener := &mitmListener{tlsConnIn}
	httpshandler := http.HandlerFunc(func(resp2 http.ResponseWriter, req2 *http.Request) {
		req2.URL.Scheme = "https"
		req2.URL.Host = req2.Host
		handler.DumpHTTPAndHTTPS(resp2, req2)
	})

	go func() {
        // step 4，启动一个伪装的 TLS 服务器
		err = http.Serve(listener, httpshandler)
		if err != nil && err != io.EOF {
			logger.Printf("Error serving mitm'ed connection: %s", err)
		}
	}()

	connIn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
}
```

1. 第一步通过 `FakeCertForName` 为浏览器请求的域名签发证书。签发所使用的 CA 为 `gomitmproxy-ca-pk.pem`, `gomitmproxy-ca-cert.pem`。
2. 第二步通过 `http.Hijacker` 拿到原始的 TCP 连接。
3. 第三步通过 `tlsConnIn := tls.Server(connIn, tlsConfig)`，将 TCP 连接转换为 TLS 连接。该 TLS 连接已配置有 CA 签发的证书。所谓的 TLS 连接，即 Go 应用程序可直接在该连接上拿到原始明文。
4. 第四步通过 `http.Serve(listener, httpshandler)` 响应这个 TLS 连接。响应的回调函数所看到的都是明文，即可进行 HTTP 抓包。

## 结语

上述过程即为 Burp Suite, ZAP 和 fiddler 等进行 HTTPS 抓包的原理。

我自制 HTTPS 中间人代理，主要是想结合 Sqlmap 做一个自动化 SQL 注入系统。由于目前所在 QA 团队并不具备 SQL 注入测试的经验，最大化的自动化所有过程就成了我的目标。目前还有 csrf token 未解决，主要是 csrf 实现千差万别，没有通用解决方法。。。