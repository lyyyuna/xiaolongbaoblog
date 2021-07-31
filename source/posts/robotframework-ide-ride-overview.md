title: Robot Framework 教程 - 集成开发环境 RIDE 概览 (译)
date: 2016-04-30 15:30:49
categories: 测试
tags: robot framework
---

## 正文

Robot Framework IDE (RIDE) 是该框架本身的集成开发环境。[Robot Framework](http://code.google.com/p/robotframework/) 是一个通用的自动化测试框架，[这里有其简单介绍](http://www.lyyyuna.com/2015/12/28/robotframework-quickstartguide/)。

改项目原本托管在 [Google Code](http://code.google.com/p/robotframework-ride/) 上，现托管在 [GitHub](https://github.com/robotframework/RIDE/downloads) 上。 

下载和安装这种事就不用我再重复了。当你打开 RIDE，导入一个测试套件或者是包含几个测试套件的目录时，就会在编辑器的左方展现一个树形结构。可以针对该树形结构，选择每一个测试套件的每一个测试用例。而且，每一个被引用的资源文件都会自动被装载，并显示在树形结构的 External Resources 中。只要测试套件能够被选中，就能修改其全局属性，例如 Suite-Setuo 和 Suite-Teardown。

![RIDE 1](/img/blog/201605/RIDE_1.png)

使用 RIDE 的一个巨大优势是可以从各个方面去配置测试套件。如果什么都是手写的，你很有可能会遗漏某些特性。而且 RIDE 中还自带完整的 Robot Framework 文档，在写关键字的时候很有帮助。当按下编辑按钮时，会弹出新窗口供你编辑。对于一些特殊的语法，比如关键字的参数使用惯导符号分割，不用担心，窗口中都会有文字提示。

![argument](/img/blog/201605/ride_3.png)

一旦你适应了如何编辑这些项。。。实际上也很容易。编辑单个测试用例时，你必须从树形结构中选取。每个用例的通用项 - 比如文档和标签 - 可以在编辑器的上方填入。在编辑器下方，是由关键字组成的用例步骤，用表格的形式给出。关键一般写在第一列，第二列开始是对应的参数。如图所示，都比较直观。

![编辑器](/img/blog/201605/RIDE_7.png)

接下来是文本编辑器部分，你可以从上方的面板中切换。RIDE 并不适用 HTML 格式来存储用例，而是采用纯粹的文本文件格式。这估计是因为 HTML 源文件难编辑的多的原因。RIDE 内部会对文本解析，展现在可视化编辑区。

![源文件](/img/blog/201605/RIDE_2.png)

在文本编辑和可是还编辑之间切换平滑，改动会自动同步。

习惯了像 Visual Studio 这么方便的 IDE 的同学，肯定也希望 RIDE 有自动补全功能，当然它有。你只需要在写关键字时按下 'CTRL-Space' 就可以了。在空白处，RIDE 会显示库中所有的可能项。请注意，下面截图的显示并不完整。

![关键字自动补全](/img/blog/201605/RIDE_33.png)

对于那些在资源文件中定义的关键字，你可以在测试用例中双击，并跳转到定义的地方。相反的，你还可以看到所有使用该关键字的地方。你可以从菜单栏上选择 'Tools -> Search Keywords -> Find Usages' 来找到它们。

![关键字所有的引用](/img/blog/201605/RIDE_4.png)

这个功能在重构测试用例时非常有用，通过单击每一项搜索结果，可以直接在编辑器中跳转到相应位置。

![搜索结果](/img/blog/201605/RIDE_51.png)

最后来看一下第三个面板，**Run**。它可以让用户直接运行打开的测试套件中的用例。运行的脚本可以是 pybot, jybot 或者是自定义脚本。对于小型项目，前两个选项够用了。但是大型项目需要额外的运行参数，和更多的独立的启动配置脚本。

脚本的运行结果可以在编辑器中看到。下面的例子可以看到，所有的测试用例都没通过:-)。

![测试失败](/img/blog/201605/RIDE_6.png)

我个人对 RIDE 的评价是，作为集成开发环境，它刚好够格，因为也没别的更好选择。这个工具提供了很多指南和内部文档，对于非技术人员无疑是有利的。

## Robot Framework 教程目录

[原文链接](https://blog.codecentric.de/en/2012/01/robot-framework-ide-ride-overview/)

1. [Robot Framework 教程 - 概览（译）](http://www.lyyyuna.com/2016/01/07/robotframework-tutorial-overview/)
2. [Robot Framework 教程 - 一个完整的例子（译）](http://www.lyyyuna.com/2016/04/09/robotframework-tutorial-a-complete-example/)
3. [Robot Framework 教程 - 集成开发环境 RIDE 概览 (译)](http://www.lyyyuna.com/2016/04/30/robotframework-ide-ride-overview/)
4. [Robot Framework 教程 - 如何组织一个可伸缩可维护的验收测试套件（译）](http://www.lyyyuna.com/2016/05/15/robotframework-tutorial-how-to-structure-a-scalable-and-maintainable-acceptance-test-suite/)
5. [Robot Framework 教程 - 循环，条件判断，字符串和列表（译）](http://www.lyyyuna.com/2016/05/28/robotframework-tutorial-loops-conditional-execution-and-more/)