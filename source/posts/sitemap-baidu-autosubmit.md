title: 向百度收录自动推送链接
date: 2015-12-29 20:20:34
categories: 杂
tags: Python script
---

## 前言

相比较三个搜索引擎，必应对我站点的收录居然是最快的，而百度的抓取确实很慢，至今未收录。研究了下百度的站长工具，里面提到可以通过发送 POST http://data.zz.baidu.com/urls?site=www.lyyyuna.com&token=XXXXXXXXXXXXX&type=original 请求，提交希望百度抓取的链接。不过百度没有给 Python 的例子，于是我准备自己写一个。

## 自动推送设计思路

整个流程比较简单，毕竟只有一个 POST 请求。公司的电脑常年开机，上面跑了虚拟机，可以作为自动推送机。

由于每天允许最大推送链接数只有 500 条，所以我计划每天定时推送一次，且每次只推送新产生的链接。

定时任务可以直接使用 crontab 完成。

对于如何找到新链接，一开始我想的是是用爬虫爬一遍，虽然我的网站是全静态的，但写爬虫真的麻烦。后来我想到，sitemap.xml 会随 hexo g 自动产生，我只需要去 sitemap.xml 找新链接即可。在我本地的自动推送机上，使用 Python 自带的 sqlite3 数据库保存已推送过的链接。

## 自动推送链接的实现

### crontab 定时触发脚本

crontab 设置成每天 1:10 AM 推送。如何设置可以谷歌或参照自带帮助说明。

    10 1 * * * python /root/sitemap_auto_push/1.py >/dev/null 2>&1

### sqlite3 数据库

更简单了，只需要一个表。

    create table urls
    {
        id integer primary key autoincrement,
        url varchar(200)    
    }

### XML 解析

sitemap.xml 是 XML 标记的，可以直接使用 Python 的 xml 模块来解析。我个人主页的 sitemap.xml 树结构如下

    <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    
    <url>
        <loc>http://www.lyyyuna.com/2015/12/28/robotframework-quickstartguide/</loc>
        <lastmod>2015-12-28T15:06:00.354Z</lastmod>
    </url>
    <url>
        <loc>http://www.lyyyuna.com/2015/12/25/scons4-building-and-linking-with-libraries/</loc>
        <lastmod>2015-12-25T13:30:14.968Z</lastmod>
    </url>    
    ...
    ...
    ...
    </urlset>

所以简单遍历一下就可以取出所有链接，再与数据库中的内容比较即可

    loc_string = '{http://www.sitemaps.org/schemas/sitemap/0.9}loc'
    for url in root:
        for attr in url:
            if attr.tag == loc_string:
                if not check_if_exist(attr.text):
                
### POST 请求

每行一个链接 POST 即可，没有键-值，不需要 urlencode。

    baidu_zhanzhang_url = 'http://data.zz.baidu.com/urls?site=www.lyyyuna.com&token=XXXXXXXXXXXXX&type=original'
    # values = 'http://www.lyyyuna.com/2015/12/28/robotframework-quickstartguide/'

    req = urllib2.Request(baidu_zhanzhang_url, values)
    response = urllib2.urlopen(req)
    the_page = response.read()      

### 完整源代码

    # -*- coding: utf-8 -*-

    import urllib
    import urllib2

    def check_if_exist(url):
        for cached_url in all_urls:
            ascii_cached_url = cached_url[1].encode('ascii')
            # print ascii_cached_url
            if url == ascii_cached_url:
                return True
        return False
        
    sitemap_url = 'http://www.lyyyuna.com/sitemap.xml'
    get = urllib2.urlopen(sitemap_url)
    xml_string = get.read()

    import sqlite3
    import xml.etree.ElementTree as ET
    root = ET.fromstring(xml_string)

    cx = sqlite3.connect('sitemap.db3')
    cur = cx.cursor()
    cur.execute('select * from urls')
    all_urls = cur.fetchall()
    # print all_urls[0][1]

    values = ''
    loc_string = '{http://www.sitemaps.org/schemas/sitemap/0.9}loc'
    for url in root:
        for attr in url:
            if attr.tag == loc_string:
                if not check_if_exist(attr.text):
                    # print attr.text
                    cur.execute("insert into urls (url) values ('" + attr.text + "')")
                    values = values + attr.text + '\n'

    cx.commit()

    print values

    baidu_zhanzhang_url = 'http://data.zz.baidu.com/urls?site=www.lyyyuna.com&token=XXXXXXXXXXXXX&type=original'
    # values = 'http://www.lyyyuna.com/2015/12/28/robotframework-quickstartguide/'

    req = urllib2.Request(baidu_zhanzhang_url, values)
    response = urllib2.urlopen(req)
    the_page = response.read()

    print the_page

若看到

    {"remain":496,"success":2}
    
你就成功了。


          