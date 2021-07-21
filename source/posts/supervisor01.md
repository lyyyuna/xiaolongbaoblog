title: 使用 Supervisor 管理进程
date: 2017-04-08 17:36:25
categories: 系统
---


## 前言

如果需要让某一个进程长期运行，该怎么做？

* 开一个终端，SSH 连上之后不关机。
* Shell 命令加一个 &，把进程扔到后台。
* 写一个 daemon 进程。
* ..

当终端关闭，终端下所有的进程也会被相应的杀死，即使是被扔到后台执行的 job。然而，要把自己的应用程序专门写成 daemon，会增加开发的负担。这时候，一种万能的、对原应用侵入最小的方法，Supervisor，便走进了我们的视线。

Supervisor 可不光具有后台长期执行程序的功能。先举两个实际的例子。

* 我所在组的产品是一个邮件网关，内含七八个扫描引擎，每种引擎都会起数个进程。为了监控和管理这些进程，我们写了很多 Shell 脚本，并用一个看门狗进程来监控进程对应的 pid 文件，一旦进程意外死亡，会被看门狗拉起来。
* 上周末为了写一个 Django + celery + redis 的例子，开了四五个终端，由于是在 virtualenv 下开发的，每次开终端都是一堆重复的 activate 过程。

这些都可以通过 Supervisor，以类似 rc.d 脚本的方式，一劳永逸的解决。

## 安装

Supervisor 是由 Python 写的，安装十分简单。

    pip install supervisor

目前只支持 Python2 (>2.4)。

不过我建议使用包管理器来安装，例如 ubuntu，

    apt install supervisor

这样安装完以后会有一个默认的配置文件生成在

    /etc/supervisor/supervisord.conf


## 配置一个后台进程

Supervisor 会按以下顺序搜索配置文件，

* $CWD/supervisord.conf
* $CWD/etc/supervisord.conf
* /etc/supervisord.conf
* /etc/supervisor/supervisord.conf (since Supervisor 3.3.0)
* ../etc/supervisord.conf (Relative to the executable)
* ../supervisord.conf (Relative to the executable)

配置文件是 Windows 的 INI 格式，我们撇开其他节，直奔主题 [program:x] 

假设有一个循环打印 hello 的程序，使用 virtualenv 中的 Python 环境运行，现在需要其在后台常驻运行。

```python
# /root/test/hello.py
import time
while True:
    print 'Hello, world.'
    time.sleep(2)
```

我们添加一个 [program:x] 小节为

    [program:hellotest]
    command = /root/test/venv/bin/python -u hello.py
    directory = /root/test
    user = root
    stdout_logfile = /root/test/hello.log
    redirect_stderr = true
    autostart = false
    autorestart = true

注意要添加 -u 启动参数，不然 stdout 上的输出会被一直缓存。首先启动 Supervisor 进程本身，安装的时候其本身已经被添加为 Linux 系统的一个 service

    # service supervisor start

然后使用 supervisorctl 工具来启动我们的 hellotest

    # supervisorctl start hellotest
    hellotest: started

查询 hellotest 的运行状态

    # supervisorctl status hellotest
    hellotest                        RUNNING   pid 898, uptime 0:02:01

查看 stdout 上的输出

    # tailf test/hello.log 
    Hello, world.
    Hello, world.
    Hello, world.
    Hello, world.
    Hello, world.
    Hello, world.
    Hello, world.
    Hello, world.
    Hello, world.
    Hello, world.
    Hello, world.

如果我们的参数配置错误，还可以查看 Supervisor 自身的 log

    /var/log/supervisor/supervisor.log

## 配置一组后台进程

配置一组后台进程与之类似，首先我们需要多个 [program:x] 小节

    [program:hellotest]
    command = /root/test/venv/bin/python -u hello.py
    directory = /root/test
    user = root
    stdout_logfile = /root/test/hello.log
    redirect_stderr = true
    autorestart = true
    autostart = false

    [program:hellotest2]
    command = /root/test/venv/bin/python -u hello2.py
    directory = /root/test
    user = root
    stdout_logfile = /root/test/hello.log
    redirect_stderr = true
    autorestart = true
    autostart = false

    [group:hellogroup]
    programs = hellotest, hellotest2

启动一组中所有进程时，命令有些不同

    supervisorctl start hellogroup:*

一旦一个 program 被加入组中，你就不能再用原先的命令启动

    # supervisorctl start hellotest
    hellotest: ERROR (no such process)
    # supervisorctl start hellogroup:hellotest

## 验证

我们可以看一下进程的 pid 号来验证我们的 hello 进程确实是 Supervisor 的子进程

    # ps -ef | grep 1182
    root      1182     1  0 16:07 ?        00:00:00 /usr/bin/python /usr/bin/supervisord -n -c /etc/supervisor/supervisord.conf
    root      1226  1182  0 16:12 ?        00:00:00 /root/test/venv/bin/python -u hello2.py
    root      1227  1182  0 16:12 ?        00:00:00 /root/test/venv/bin/python -u hello.py

再用 kill 命令验证 Supervisor 具有看门狗功能

    # kill -9 1226
    # ps -ef | grep 1182
    root      1182     1  0 16:07 ?        00:00:00 /usr/bin/python /usr/bin/supervisord -n -c /etc/supervisor/supervisord.conf
    root      1227  1182  0 16:12 ?        00:00:00 /root/test/venv/bin/python -u hello.py
    root      1255  1182  0 16:18 ?        00:00:00 /root/test/venv/bin/python -u hello2.py

hello2.py 已经是新的 pid 号。