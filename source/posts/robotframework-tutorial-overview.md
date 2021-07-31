title: Robot Framework 教程 - 概览（译）
date: 2016-01-07 20:33:01
categories: 测试
tags: robot framework
series: Robot Framework 教程
---

## 前言

[Robot Framework](http://code.google.com/p/robotframework/) 是一个通用的自动化测试框架。这是本系列的第一篇文章，将会给出一个全面的概述。

请注意，第一篇文章几乎不会包含任何"真正实现方面的东西"，而是讲述一些高级别、抽象的概念，来为以后的文章打下坚实的基础。

## 什么是自动化测试框架

如果你已经有自动化测试的经验（例如，使用过一些自动化测试工具），你可以直接跳过本小节。

现在，我想先问一个问题：什么是自动化测试？以及它为何不同于手动执行的测试？Michael Bolton 写的[这篇文章](http://www.developsense.com/blog/2009/08/testing-vs-checking/)对这两个问题给出了一个非常好的答案：

> 核对是机器干的事，测试则需要智慧。

然而请注意，当我们下面讨论自动化测试时，会同时使用测试和核对两个术语。

让我们来看个具体的例子，比如说一个保险公司的评级引擎。这个引擎会根据输入的某些参数（数字）来计算值。对于这样一个算法已知的系统而言，显然大量（自动化）核对是非常合适的。但检验这个算法正确与否，则需要一些思考。

假设我们有一个基于数据库表的接口，和一个批处理程序：从一个表中取出数据，计算后与另一个表中的数据核对结果。

首先，我们需要测试脚本语言（取决于个人喜好，可以是 Shell, Perl, Java 等等）。此外我们还要准备一些基本的测试功能。然后访问数据库表，执行一个又一个脚本，测试结果最好以某种报告的形式返回。一旦脚本运行完毕，我们便可以开始核对检查工作。基本上，我们认为，一个典型的自动化测试框架应该提供上述的功能。

![最简测试框架](/img/blog/201601/GenericFrameworkView.png)

上图描述了一个非常基本的自动化测试框架。该框架有一个可执行测试的核心系统，可以输出一些报告，并提供接口来插入特定的测试功能。这个插入接口实现会非常简单。

这就带来一个基本的问题：当我用这些测试框架时，该用什么编程语言来实现我的测试功能？稍后我们会详细回答这个问题，但现在我们可以说，**Robot Framework** 测试框架允许使用很多不同的语言。

在了解 **Robot Framework** 的具体体系结构之前，我们先讨论下 **Robot Framework** 的核心术语，即**关键字驱动测试**。

## 什么是关键字驱动测试

每当我试图解释什么是关键字的时候，我总会把它称为函数或者方法，其能够用于测试被测系统的一个方面。

真正强大的是，一个关键字可以由其关键字来定义。这就是为什么通常说：

* **高级别关键字**：这些关键字用来测试系统业务逻辑。
* **低级别关键字**：为了控制高级别关键字的实现在合适的大小，会将其分割成几个低级别关键字。
* **技术性关键字**：这些关键字提供技术实现细节。

下面这张图用一个关键字例子描述了这三者的关系。

![嵌套的关键字定义](/img/blog/201601/Keywords.png)

通常技术性关键字可以由任何编程语言来实现（好吧，不是真的）。其他的**关键字**则是由已存在的关键字组合而成。即使本文关注的是抽象的概念，我们还是来看一个具体的关键字定义：

![GoogleS earch 关键字](/img/blog/201601/KeywordGoogleSearch.png)

这个例子表明， Google Search 这个关键字可以由 **Selenium Library** 库中的关键字来创建。好消息是已经有大量预定义的关键字，它们的集合称为**测试库**。

好，让我们开始。。。

## 说好的概览

最后开始我们的主题，**Robot Framework** 概览。安装 **Robot Framework** 时，一些标准测试库会随核心框架一起安装。

除了[标准测试库](http://code.google.com/p/robotframework/wiki/TestLibraries)外，还有很多[额外的外部测试库](http://code.google.com/p/robotframework/wiki/TestLibraries#External_test_libraries)。它们通常是社区由不同的目的贡献的。在写特定测试用例的时候，你完全可以混用不同测试库的所有关键字。这意味着，在测试一个 web 应用时，可以用 **Selenium Library** 来与 web 前端交互，用 **Database Library** 来检测数据库中数据的正确性。理想情况下，完全不需要编程，只需组合库中的关键字来构成高级别关键字。

![Robot Framework 概览](/img/blog/201601/Overview_3.png)

Robot Framework 除了核心功能和测试库外，还提供了一个 IDE (RIDE, Robot Integrated Development Environment)，用户可以在此编写和组织自己的测试用例和关键字。请注意，这个 RIDE 不是用来写技术性关键字的。技术性关键字取决于你的开发环境，比如 Eclipse 来开发 Java 写的关键字。

上图中尚未包含**资源文件**。我们测试用例的集合称作**测试套件**，听起来很有道理。现在为**测试套件**添加新关键字也是可行的。但最好在外部的**资源文件**中定义关键字。

现在，我们在使用 Robot Framwwork 框架时，有三个重要的概念：

* **测试套件**：这是测试用例（机器的核对工作）的容器。通常每个项目至少有一个*测试套件*。在大型工程中，需要根据功能来划分不同的*测试套件*。
* **资源文件**：为了让测试设计者的角度看，几乎总会定义高级别关键字。反过来说，通常会有自己的资源文件。特别是对产品开发或者一些长期项目而言，肯定能益于**关键字**，而且还能被其他项目组使用。
* **测试库**：通常不需要编写自己的**技术性关键字**，除非你在使用特定技术细节，你才需要自己实现一个新的测试库，不过这并不费时。

需要强调的是，**测试库**中的关键字和**资源文件**中组合成的关键字，在使用时没有区别。

## 自定义测试功能该用什么语言

Robot Framework 自身和其核心库都是由 [Python](http://www.python.org/) 实现的。因此，如果熟悉 Python （或者打算开始熟悉 :)），用其写自己的关键字是个好的选择。我一直认为 Python 是个很酷的语言，但如果 Robot Framework 局限于 Python，它不会这么成功。这就是为什么会有 [Jython](http://www.jython.org/)。有了 Jython 就可以在 Java 的虚拟机上运行 Python 代码。这使得我们能够用 Java 来编写测试库，甚至是任何编译成 Java Byte Code 的语言。

> 在 .NET 系中，IronPython 和 Jython 是类似的。

这引出了以下的安装堆栈：

![Robot Framework 安装堆栈](/img/blog/201601/InstallationStacks.png)

Robot Framework 历史上是上图最左边的安装堆栈（在没有 RIDE 的情况下）。早期 Jython 的安装和支持欠佳。然而，现在 Java 已经支持的很好了，只有很少的缺陷。

现在还有 JAR 安装方式的 Robot Framework，Python 的测试库和 Jython 都被打包成一个大的 JAR 文件。这有很大的优势，你可以将 JAR 放入版本控制中，或者是放入本地的 Maven 仓库中。这样就能保证团队成员都使用相同版本的 Robot Framework，且能够实时更新。不过也有缺点，这种情况下，RIDE 无法显示 JAR 文件关键字的帮助信息，所以 RIDE 还需要单独安装。

好，现在回到我们的主题：用什么语言来实现自己的测试功能。

## Remote Libraries

目前为止，无论是本地还是服务器，Robot Framework 都是安装在同一台机器。然后我们也看到可以用 Python, Jython 和纯 Java 语言来开发测试库。

用 [Remote Libraries](http://code.google.com/p/robotframework/wiki/RemoteLibrary)，可以在其他机器中，用支持 [XML-RPC protocol](http://code.google.com/p/robotframework/wiki/RemoteLibrary) 的任何语言来编写测试库。

![Remote Library 用法](/img/blog/201601/RemoteLibrary.png)

当在测试用例和资源文件中导入 Remote Library，它们用起来和普通的库没有区别。还有一个优点是，你也可以从 Remote Library 中获取帮助文件。如果对其实现感兴趣，可以看看 Database Library 的源码。需要指出的是，Remote Library 功能本身就是某些测试库的附加功能。

Remote Library 作为一个远程的服务器，而 Robot Framework 作为一个客户端。当然这两个库完全能在本地使用。

这是绝对不能被低估的一个非常强大的功能。

## 持续集成

把 Robot Framework 集成到持续集成服务器中非常直接，因为框架本身是用脚本语言写成的。Java 版本当然也可以使用 Maven 集成。

## Robot Framework 教程目录

[原文链接](https://blog.codecentric.de/en/2012/03/robot-framework-tutorial-overview/)

1. [Robot Framework 教程 - 概览（译）](http://www.lyyyuna.com/2016/01/07/robotframework-tutorial-overview/)
2. [Robot Framework 教程 - 一个完整的例子（译）](http://www.lyyyuna.com/2016/04/09/robotframework-tutorial-a-complete-example/)
3. [Robot Framework 教程 - 集成开发环境 RIDE 概览 (译)](http://www.lyyyuna.com/2016/04/30/robotframework-ide-ride-overview/)
4. [Robot Framework 教程 - 如何组织一个可伸缩可维护的验收测试套件（译）](http://www.lyyyuna.com/2016/05/15/robotframework-tutorial-how-to-structure-a-scalable-and-maintainable-acceptance-test-suite/)
5. [Robot Framework 教程 - 循环，条件判断，字符串和列表（译）](http://www.lyyyuna.com/2016/05/28/robotframework-tutorial-loops-conditional-execution-and-more/)