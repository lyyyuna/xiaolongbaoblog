title: 解决 GitHub 禁百度爬虫，并自动 ping 搜索引擎
date: 2016-01-03 22:21:36
categories: 杂
tags: Python script
---

本篇包含两个内容。

## DNSPod 双解析

用百度站长工具测试时，才发现 GitHub 禁了百度爬虫，解决方法也比较简单。

首先更换你的 DNS 解析服务器，推荐使用免费的 DNSPod 解析。它支持多路解析，可以专为百度搜索，解析到另一个服务器。

国内的 GitCafe 并不屏蔽百度，在 GitCafe 建立 Page 的过程是类似的。然后只需要在 hexo 的 _config.yml 上配置两个 repo 即可，如下：

    deploy:
    - type: git
    repository: git@github.com:lyyyuna/lyyyuna.github.io.git
    branch: master
    - type: git
    repository: git@gitcafe.com:lyyyuna/lyyyuna.git
    branch: gitcafe-pages

你需要在你的域名提供商那修改 DNS 服务器为：

    f1g1ns1.dnspod.net
    f1g1ns2.dnspod.net

在 DNSPod 配置面板上：

![DNSPod](/img/blog/201601/dnspod.jpg)

大概需要等待 1-2 天，全球的解析才会恢复正常。而这时候，百度也应该能够正常抓取网页了。

## 自动 ping 搜素引擎

大部分搜索引擎都提供了 ping 服务，当你有新的文章发布时，可以 ping 一下搜索引擎让其来抓取。

这里，在上一篇自动提交 sitemap.xml 功能的基础上，再添加自动 ping 百度和必应的功能。

### ping 必应

必应的接口比较简单，只需要 POST 自己的 sitemap.xml 即可。

    sitemap_url = 'http://www.lyyyuna.com/sitemap.xml'
    bing_ping = 'http://www.bing.com/webmaster/ping.aspx?siteMap=' + sitemap_url

    get = urllib2.urlopen(bing_ping)
    result = get.read()
    print result

### ping 百度

百度的接口稍显复杂，需要 POST XML 结构的数据。

    RPC端点： http://ping.baidu.com/ping/RPC2
    调用方法名： weblogUpdates.extendedPing
    参数： (应按照如下所列的相同顺序传送)
    博客名称
    博客首页地址
    新发文章地址
    博客rss地址 

提交的 XML 格式为

    <?xml version="1.0" encoding="UTF-8"?>
    <methodCall>
        <methodName>weblogUpdates.extendedPing</methodName>
        <params>
            <param>
                <value><string>百度的空间</string></value>
            </param>
            <param>
                <value><string>http://hi.baidu.com/baidu/</string></value>
            </param>
            <param>
                <value><string>http://baidu.com/blog/example.html</string></value>
            </param>
            <param>
                <value><string>http://hi.baidu.com/baidu/rss</string></value>
            </param>
        </params>
    </methodCall>

好在 Python 处理 XML 方便，将新发文章地址作为参数，如下

    root = ET.Element("methodCall")
    methodname = ET.SubElement(root, "methodName").text = 'weblogUpdates.extendedPing'
    params = ET.SubElement(root, "params")

    param = ET.SubElement(params, "param")
    value = ET.SubElement(param, 'value')
    string = ET.SubElement(value, 'string').text = u'lyyyuna 的小花园'

    param = ET.SubElement(params, "param")
    value = ET.SubElement(param, 'value')
    string = ET.SubElement(value, 'string').text = u'http://www.lyyyuna.com/'

    param = ET.SubElement(params, "param")
    value = ET.SubElement(param, 'value')
    string = ET.SubElement(value, 'string').text = url

    param = ET.SubElement(params, "param")
    value = ET.SubElement(param, 'value')
    string = ET.SubElement(value, 'string').text = u'http://www.lyyyuna.com/atom.xml'

    # tree = ET.ElementTree(root)
    xmlstr = ET.tostring(root, encoding='utf8', method='xml')
    # print xmlstr
    # print

    baidu_pingRPC = 'http://ping.baidu.com/ping/RPC2'
    req = urllib2.Request(baidu_pingRPC, xmlstr)
    response = urllib2.urlopen(req)
    the_page = response.read()

    print the_page

对代码进行了整合，可以同时提交 sitemap.xml 并 ping。代码见 [我的GitHub](https://github.com/lyyyuna/script_collection/blob/master/baidu_url_auto_submit/1.py)。