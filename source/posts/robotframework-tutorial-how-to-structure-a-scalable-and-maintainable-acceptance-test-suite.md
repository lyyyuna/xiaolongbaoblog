title: Robot Framework 教程 - 如何组织一个可伸缩可维护的验收测试套件（译）
date: 2016-05-15 10:56:18
categories: 测试
tags: robot framework
---


当你在一个 sprint 中开始写自动化验收测试时，你没必要重新测试之前每个 sprint 的结果。但经过几轮自动化测试之后，整个测试看起来不再像一个精心设计过的测试套件，而是乱七八糟。这些你一定经历过。这篇文章将展示一些最佳的模式和经验，让你写出可伸缩可维护的测试结构。

我们只考虑测试框架本身，忽略和执行有关的问题，比如日志系统、并发和测试硬件。由于我们一直使用 [Robot Framework](http://robotframework.org/) 来实现自动化，所以本文的解决方法会有一些局限性。但其他测试框架的用户也可以参考，比如 FitNesse, Cucumber, Concordion, etc。

好，一个单独的，精心设计的验收测试并不会存在太长时间，但如何写一个可维护的测试套件？这里我用测试套件这个术语，意味着并不是单单指测试用例，还包括了库、启动脚本和框架等。

由于经常改变项目需求和源码实现，测试套件也需要跟着改变。如何让测试套件尽可能地适应各种变化？显然，需要在其中分离出可变和稳定的部分。

稳定的部分是指测试框架本身和附加的库。测试用例也会尽可能作为不变的部分，除非需求改变没理由要改变它们。当然，这里允许添加新的测试用例。这里不经要问了，既然框架和测试用例都为稳定的部分，那什么是可变的部分呢？下图展示了一个可伸缩可维护的测试套件的结构。下面会对图中每一层作详细的解释。简单的说，如果系统的上下两头都需要稳定，那么可变的部分只有中间那层。


![测试套件结构](https://raw.githubusercontent.com/lyyyuna/blog_img/master/blog/201605/SMAT_structure.png)

不同的颜色代表了 Robot Framework 中不同的文件：

1. **红色**：测试用例和套件。
2. **绿色**：资源文件。资源文件包括关键字。一个关键字可以是一个方法，或者是其他方法的组合。
3. **蓝色**：框架自身的内容。

在图的左边，你可以看到对每一层稳定性和可变性的度量。在最顶层和最底层，稳定性最高，而在最中间那一层，可变性又变的最高。这意味着中间那层的元素经常改变，而其他相对恒定。

## 每一层

### 测试用例

我们推荐的是 “Given/When/Then” 方式来写用例。这意味着只有当需求变化时，才需要动测试用例。

### 测试套件

测试用例应该按功能组织为测试套件。开头，我们按照用户需求对测试用例分类。在回顾时，发现当相同的用户需求增多时，得花越来越多的时间搜索特定的用例。所以，对用户需求和分类的用例打上**标签**，将有利于搜索。

### 导入

下一层是“导入”。这一层包含了自动化验收测试所需的关键字，能够建立测试用例和资源文件之间的对应关系。既然用例不应经常改变，那么关键字就需要持续重构了。由于测试套件之间包含一些类似的测试用例，它们理应具有相同的导入，这样只需为所有的测试套件写一份资源文件。

### 聚合对象

聚合对象这层改动最为频繁。比如在测试 web UI 时，会建立页面对象的模式。从一个web页抽象的概念，在这一层如何构建经典的软件设计与工程：结构的灵活性和可维护的代码？

### 库适配器

我们增加了库适配器层。每一个库都应该由资源文件来导入，这样能保证系统中只有一个库的实例。而且有时候库需要用不同的参数来初始化。并且，我们随时会在适配器层增加关键字来扩展功能，而对上层的测试用例保持黑盒状态。

### 平台

这一层是指 RObot Framework 框架本身和标准库。

## 结论

一方面，我们构建测试套件的方法产生了深远的影响。在另一方面，我们继续学习新的东西，我相信新的经验和教训将进一步影响测试用例和套件的结构和设计。希望其他项目能从中受惠。


## Robot Framework 教程目录

[原文链接](https://blog.codecentric.de/en/2010/07/how-to-structure-a-scalable-and-maintainable-acceptance-test-suite/)

1. [Robot Framework 教程 - 概览（译）](http://www.lyyyuna.com/2016/01/07/robotframework-tutorial-overview/)
2. [Robot Framework 教程 - 一个完整的例子（译）](http://www.lyyyuna.com/2016/04/09/robotframework-tutorial-a-complete-example/)
3. [Robot Framework 教程 - 集成开发环境 RIDE 概览 (译)](http://www.lyyyuna.com/2016/04/30/robotframework-ide-ride-overview/)
4. [Robot Framework 教程 - 如何组织一个可伸缩可维护的验收测试套件（译）](http://www.lyyyuna.com/2016/05/15/robotframework-tutorial-how-to-structure-a-scalable-and-maintainable-acceptance-test-suite/)
5. [Robot Framework 教程 - 循环，条件判断，字符串和列表（译）](http://www.lyyyuna.com/2016/05/28/robotframework-tutorial-loops-conditional-execution-and-more/)