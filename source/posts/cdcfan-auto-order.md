title: 一个简单的公司自动订餐脚本实现
date: 2016-01-04 21:20:36
categories: 杂
tags: Python script
---

## 前言

因为食堂难吃，公司推出了集体外卖订餐服务，员工需要每天去指定网页提交订餐。

由于经常忘了订餐，便诞生了下面这个脚本。

该系统功能是工作日每天定时订餐，并发送邮件告知用户订餐结果。

## 登陆过程分析

打开 http://cdcfan/ 便可见到登陆页面，登陆时只需输入自己英文名即可。

![cdcfan 登陆页面](/img/blog/201601/cdcfan_logon.jpg)

这里点击“进入”会发送 GET 请求，然后获得员工的各种信息，比如 ID，部门编号等。

![cdcfan 登陆 GET](/img/blog/201601/cdcfan_logon_get.jpg)

登陆结果如下

![cdcfan 登陆结果](/img/blog/201601/cdcfan_ehuaile.jpg)

当用户点击“下单”，这时候才会发送 POST 参数，请求参数主要为之前 GET 请求获得的部门号和员工 id。

![cdcfan 下单请求](/img/blog/201601/cdcfan_order-new.jpg)

请求会返回一个 json 数据。代表你是否成功订餐，还是已经订过餐。比如已经订过后，exceed_count 为 1，其余为 0.

![cdcfan 下单结果](/img/blog/201601/cdcfan_order_result.jpg)

而页面则显示

![cdcfan 吃胖了](/img/blog/201601/cdcfan_chipangle.jpg)


回过头来看登陆过程，如图 2，每一步 GET 请求，都会附上一个隐含参数 '_=1451880513702'，参数值会逐次递增 1。这个值有什么用我还没研究出来，用 Burp Suite 代理修改了也依然能返回正确结果。不过实际上不需要去分析这个值是由服务器还是客户端产生，因为最关键的下单 POST 请求与这个毫无关系。

于是一切就变得简单了，实测这个系统其实并不依赖于会话机制，员工输入的姓名只是用来获得对应的员工 id 和部门号。那意味着，实现自动化订餐时，**并不需要模拟登陆过程**，只要向 /api/order-new POST 自己的 id 和 部门号，就能完成订餐。

## 自动化订餐

首先定义数据库表，

    create table cdcuser
    {
        id integer primary key,
        psid varchar,
        depcode varchar,
        email varchar    
    }

然后只需要遍历取出的每个记录，分别发送 POST 请求即可。

    for user in r:
        psid = user[2]
        depcode = user[3]
        email = user[4]

        post_data = {'order':'e-1', 'psid':psid, 'depcode':depcode}
        post_data_urlencode = urllib.urlencode(post_data)

        req_url = 'http://cdcfan/api/order-new'
        req = urllib2.Request(url = req_url, data = post_data_urlencode)

        res_data = urllib2.urlopen(req)
        res_json = res_data.read()
        print res_json

## 定时任务和发送邮件

得益于公司内部 MTA 的转发策咯，任意邮件域名不作检查，除恶意邮件外，其他直接被转发。这意味着可以自己搭建 MTA 发送邮件。

MTA 可以直接使用 Postfix，在 Debian 8 上默认配置即可工作。Python 有非常方便的 smtplib 模块用于连接 MTA 并发送邮件

    server = smtplib.SMTP('localhost')
    server.set_debuglevel(1)
    server.sendmail(fromaddr, toaddrs, msg)
    server.quit()

不过我偷懒，直接解析 json 结果后，调用系统的 mail 命令

    res = json.loads(res_json)
    if (res['succeed_count'] == 0):
        cmd = "echo 'ding can shi bai, nin ke neng yi jing ding guo le. DO NOT REPLY' | mail -s cdcfanauto " + email
        # print cmd
        os.system(cmd)
    else:
        cmd = "echo 'ding can cheng gong. DO NOT REPLY' | mail -s cdcfanauto " + email
        os.system(cmd)

最后就是在每周工作日定时调用，这里直接使用 crontab 的定时任务

    10 10 * * 1-5 python /root/cdcfan/cdc.py >/dev/null 2>&1

### 问题

实际测试时遇到了邮件不能发送的问题，原因我没有仔细研究。

（大概是 crontab 在定时时会向系统管理员发送邮件，而由于上述的 python /root/cdcfan/cdc.py >/dev/null 2>&1 其实是 crontab 的子进程？？？，查看了 Postfix 的 maillog 之后发现，这两个进程之间的发送邮件会相互干扰。）

（比如 Postfix 的默认配置下，邮件域名为 @debian.com。所以默认管理员邮箱会向这个邮件域名发送。但这个域名是不存在，作为发送方，可以伪造，因为协议并没有强制要求。但作为接收方，伪造的邮件域名必然会不可达，最终被退回。）

（不过按理说不应该啊，不知道为什么管理员邮件发送失败会造成我调用 mail 失败。我又按照网上说的把 crontab 的发送邮件关闭，现象依旧，maillog 显示邮件被“自己的 MTA”退回。）



## http 服务器

由于之前所述的我不能解释的问题，我只能另辟蹊径。让 crontab 定时发送 GET 请求到本地的一个 http 服务器，本地 http 服务器收到请求后再执行上述的订餐任务。本地 http 服务器和 crontab 完全是两个不相关的进程，不存在相互干扰的问题。

crontab 修改为

    10 10 * * 1-5 curl http://127.0.0.1:5000/cdc/ >/dev/null 2>&1

服务器可以用 Flask 简单搭一个

    from flask import Flask
    import cdc

    app = Flask(__name__)

    @app.route('/cdc/')
    def submit():
        cdc.submitcdcfan()
        return "haha"

    if __name__ == '__main__':
        app.run()

## 最终结果

成功收到订餐成功的提示邮件

![cdcfan 邮箱结果](/img/blog/201601/cdcfan_success.jpg)

## 彩蛋

大概因为是安全公司的内部系统，所以在页面源码里看到了如下的注释

![js 源码](/img/blog/201601/cdcfan_hack.jpg)


