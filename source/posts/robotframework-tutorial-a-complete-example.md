title: Robot Framework 教程 - 一个完整的例子（译）
date: 2016-04-09 14:50:08
categories: 测试
tags: robot framework
series: Robot Framework 教程

---


## 前言

用 [Robot Framework](http://code.google.com/p/robotframework/) 时有太多的选择：

* 使用 Python, Jython 还是 Java？
* 测试用例使用哪种输入格式（HTML, Text, BDD）？
* 要使用 [Robot IDE(RIDE)](http://www.lyyyuna.com/2016/01/07/robotframework-tutorial-overview/) 吗？
* 如何在本地和持续集成环境中运行相同的测试？
* 如何运行所有的测试 (scripting, ANT, Maven)？

那什么是最好的选择呢？我见过的世面太多了。当然，在 Eclipse 中用 Maven 做 Robot 测试非常酷。BDD 相比较 HTML 更适合敏捷开发。

所有这些有着相同的共性：简单！这不仅意味着设置和运行简单，还意味着更容易排错。在不同技术背景的团队合作间，这尤其重要。

接下来我们用一个完整的例子展示 Robot Framework 的使用方法。


## 测试准备

开始测试工程前，首先要想好被测系统需要哪些测试库：

* 是测 web 应用？那你可能需要 SeleniumLibrary 或者 Selenium2Library。
* 是测数据库？Python 和 Java 都有相应的数据库测试库。
* 是测试 SSH/SFTP？那你可能需要 SSHLibrary。

这个列表可以继续列下去，直到没有可用的测试库为止。这时候你就需要自己写啦（需要单独写一篇文章来阐述）。

为什么如此重要？测试库的选择直接影响到了是使用 Python, Jython 还是 Java 版的 Robot Framework。某些测试库只有 Java 的实现，如果要用纯 Python 来调用此库，则要求其实现 **Remote Library 接口**。因此，在测之前，需要好好想想。

> 本文的代码在 [GitHub](https://github.com/lyyyuna/Robot-Framework-Sample-Project)
 
我们假设被测系统是一个利用 MySQL 数据库做存储的 web 应用（非常普遍）。浏览器使用 Python 的 SeleniumLibrary，数据库使用 Java 版本的 DatabaseLibrary，并用 **Remote Library 接口**。

## 测试框架

下图是测试框架的概览：

![测试框架概览](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201604/Sample_Overview1.png)

Selenium 需要 Selenium Server。这可以是 Robot Framework 所在的同一台机器，也可以是另一台可通过 TCP/IP 连接的服务器。Database Library Server 同理。虽然 DatabaseLibrary 可以本地使用，但那就意味着必须使用 Jython 来测试了。当然也可以在同一台机器上运行多个服务器。在一些正式的测试环境中，Robot Framework 和 CI (持续集成) 服务器经常部署在一起。然后，Selenium Server 通常跑在 Windows 服务器上，因为需要尽量模拟用户的使用场景。DatabaseLibrary Server 也可以部署在 CI 服务器上。

## 测试实现和管理

最后让我们来实现该测试。不是每一个细节都会 cover，具体可以看 [GitHub](https://github.com/lyyyuna/Robot-Framework-Sample-Project)。

但在此之前，让我们再多做一些常规性考虑。比如用哪种格式来组织测试用例，是否使用 RIDE。而 RIDE 的使用又会直接影响到测试用例的格式。团队成员的技术背景，以及不同团队合作潜在的维护成本，对上述选择都有影响。

> Tips: 如果你已经使用 Excel 来管理则是用例，你可以直接复制粘贴进 RIDE。

要我在本文的例子中选择，我会选择 HTML 格式和 RIDE，理由如下：

* RIDE 相比较于最初版本已经有了十足的进步，支持关键字自动补全，实现 Test Suites 和 Resource Files 也十分便利。
* 使用 RIDE 不用特意考虑 BDD 风格。但其中有一些我不喜欢的语法元素。而且，非技术团队成员编写和维护测试用例比较困难，因为现在机器还不能完全看懂人类语言。并且我认为，如果 BDD 是唯一或者最重要的需求，其他那些只支持 BDD 的测试框架才会有优势。
* HTML 格式有着简单粗暴的优点。你可以直接在浏览器中可视化这些测试用例，尤其是那些熟悉 Excel 的非技术团队成员，看到这些会感到非常亲切。
* HTML 格式也有着缺点，在版本管理时，HTML 会带来各种各样的问题。


在实现测试时最重要的就是能够同时在本地和正式测试环境（CI 服务器）中运行。幸运的是，Robot Framework 支持向关键字传入参数，这样便能轻松切换环境：

* 参数为 web 应用的 URL
* Selenium 服务器的 IP 地址与端口
* Database Library 所使用的 JDBC 连接字符串

这些参数可以存储在变量文件中。这些变量文件可以在命令行中传入 Robot Framework。由于在本地测试和 CI 服务器中有不同启动脚本，这样便能实现不同环境的快速切换。

这意味着最好以如下的目录树来组织你的测试工程。

![测试工程目录树](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201604/Sample_DirectoryStructure2.png)

定义一个通用的目录树结构有助于工程的复用。上述的目录结构对我来说工作的不错，我很早就开始用啦。

> Tips: 所有路径都应该用相对路径。

首先，我们将所有文件放于顶层文件 robot 中。然后将测试的实现和执行分开存放。在实现这边，testsuites 和 resource 分开存放。当然在一些大型工程中，还需要额外的子目录来更好的组织测试用例。更重要的是，最好用相对路径来引用这些文件。使用相对路径能够更好的在不同系统间移植，项目成员间通过版本管理系统也能更好地共享工程。

执行分支这边必须处理不同运行目标环境的问题，比如本地开发环境和正式的 CI 环境。若还有其他的部署环境需要在此目录中实现。scripts 目录用于保存执行用的脚本（robot 本身，Selenium Server, Database Library Remote Server），setting 目录放置特殊的变量文件。请注意，这些脚本写完之后就不应该频繁改动，对于配置文件亦是如此，除非执行环境有变化。

最后是 lib 文件夹，这取决于项目是否需要自己编写库文件。

## 执行测试

当执行测试时，我坚持使用 shell 脚本。易于理解，历史悠久且不出问题，在 CI 环境中使用方便。当然，我们很可能需要两套不同的启动脚本，因为本地测试通常在 Windows 电脑上，而正式的 CI 环境是一些 bash 或者 csh 脚本。但需注意，这些写了“写了一遍就忘记”的脚本，其实并不复杂。

在最初，我们需要三个脚本：

1. robot 测试的启动脚本
2. Selenium Server 的启动脚本
3. Database Library Remote Server 的启动脚本

我们也可以把三个脚本合并成一个，但为什么不这么做呢，因为其实后两个服务器只需启动一次，只有测试才需要重复执行。

## 整合

首先我们需要在开发机器上安装 Robot Framework 和库。我们假设平台是 Windows，在 Unix 上安装也不会太复杂。

> Robot Framework 同时支持 2.x 和 3.x。

安装如下的工具包：

* Python 2.7
* Robot Framwork
* wxPython
* RIDE
* Selenium2Library
* Database Library Server

按顺序安装，然后配置 PATH 目录为  “C:\Python27;C:\Python27\Scripts”。现在你可以用

    pybot
   
来运行 Robot Framework，用

    ride
    
来启动 RIDE。示意图如下，

![测试工程目录树](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201604/RF_Ride.png)

Selenium Library 通常会包含一个对应的 Selenium Server JAR 包。为了独立使用不同目录下的 Selenium Server（比如其他小组成员安装的），你需要指定一个新的环境变量 RF_HOME，该变量指向 Python 的安装目录。该变量用于 selenium 服务器的本地启动脚本。

对于本地的 MySQL 数据库，其配置[在此](https://github.com/ThomasJaspers/robotframework-dblibrary/tree/master/sample)。然后安装 MySQL，创建测试 schema 和相应的用户：

    C:\xampp\mysql\bin>mysql -u root -p
    mysql> create database databaselibrarydemo;
    mysql> create user ‘dblib’@’localhost’ identified by ‘dblib’;
    mysql> grant all privileges on databaselibrarydemo.* to ‘dblib’;

这里是工程的源码 [GitHub](https://github.com/lyyyuna/Robot-Framework-Sample-Project)。

在 robot/execution/local/scripts 是执行测试前所有需要运行的脚本。测试的实现在 robot/implementation/testsuites 目录中。测试用例可以直接用 RIDE 打开 implementation 目录，然后直接查看和修改。

为了运行测试，必须先启动 Selenium 服务器和 DBLibrary 服务器。然后运行 Testsuite。Windows 的批处理脚本在 robot\execution\local\scripts 目录中。因为都使用相对路径，一切应该按计划顺利运行。这里虽然在被测服务器上部署文件，但本地可以很容易地适配。

## 结论和感想

我们已经看到，Robot Framework 提供了众多功能和可能，即使同一件事也能用不同方法来完成。所以，在正式开始测试前做些基本分析是很有意义的。

![在 RIDE 中编辑 Testsuites 和 Resource 文件](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201604/impression_ride1.png)

使用 RIDE 使得实现测试功能更简单，尤其是那些非技术团队。简单意味着好维护（不只是 Robot Framework 测试哦 ;-)）。

![Selenium Server 启动和运行](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201604/impression_seleniumserver.png)

顺便说一下，我还没有明确指出过，Robot Framework 的 **报表** 和 logging 非常棒，在 troubleshooting 时非常有用。

![Robot Framework 的 log 文件，其中含有浏览器屏幕截图](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201604/impression_logfile.png)

Robot Framework 在各种不同的测试库中提供大量的测试功能。一旦决定哪个测试库最好用时，大大加速了写测试的过程，提高了生产力。

![Database Library Server 运行](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201604/impression_dbserver.png)

尤其是在许多不同工程工作时，一个通用的工程结构和工具非常有用。在一些公共的资源文件中也需要实现一些产品相关的关键字。

希望本文有助于你开始使用 Robot Framework，并有效地组织你的测试工程结构。当然，本例还有许多增强的地方，希望这是一个良好的起点。

## Robot Framework 教程目录

[原文链接](https://blog.codecentric.de/en/2012/04/robot-framework-tutorial-a-complete-example/)

1. [Robot Framework 教程 - 概览（译）](http://www.lyyyuna.com/2016/01/07/robotframework-tutorial-overview/)
2. [Robot Framework 教程 - 一个完整的例子（译）](http://www.lyyyuna.com/2016/04/09/robotframework-tutorial-a-complete-example/)
3. [Robot Framework 教程 - 集成开发环境 RIDE 概览 (译)](http://www.lyyyuna.com/2016/04/30/robotframework-ide-ride-overview/)
4. [Robot Framework 教程 - 如何组织一个可伸缩可维护的验收测试套件（译）](http://www.lyyyuna.com/2016/05/15/robotframework-tutorial-how-to-structure-a-scalable-and-maintainable-acceptance-test-suite/)
5. [Robot Framework 教程 - 循环，条件判断，字符串和列表（译）](http://www.lyyyuna.com/2016/05/28/robotframework-tutorial-loops-conditional-execution-and-more/)