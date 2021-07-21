title: Robot Framework 教程 - 循环，条件判断，字符串和列表（译）
date: 2016-05-28 09:05:09
categories: 测试
tags: robot framework
series: Robot Framework 教程

---

目前为止，[Robot Framework](http://robotframework.org/)教程一直关注于高阶抽象的概念，所以这次有必要介绍一下框架本身的基础概念。所有这些特性都直接来自于标准库，而本文的例子再安装额外的库。

## 有关循环的关键字

让我们从循环开始。[Robot Framework](http://robotframework.org/)为循环提供了多种方案。

* 在一系列元素中循环
* 根据数字的范围来循环
* 重复执行某一个关键字多次

最后一个和真正的循环是有差别的，意味着你得将所有操作封装到一个关键字中。并且，在执行结束之前无法退出循环。

让我们来看一些例子：

**测试**：

    *** Settings ***
    Library           String

    *** Test Cases ***
    For-Loop-In-Range
        : FOR    ${INDEX}    IN RANGE    1    3
        \    Log    ${INDEX}
        \    ${RANDOM_STRING}=    Generate Random String    ${INDEX}
        \    Log    ${RANDOM_STRING}

    For-Loop-Elements
        @{ITEMS}    Create List    Star Trek    Star Wars    Perry Rhodan
        :FOR    ${ELEMENT}    IN    @{ITEMS}
        \    Log    ${ELEMENT}
        \    ${ELEMENT}    Replace String    ${ELEMENT}    ${SPACE}    ${EMPTY}
        \    Log    ${ELEMENT}

    For-Loop-Exiting
        @{ITEMS}    Create List    Good Element 1    Break On Me    Good Element 2
        :FOR    ${ELEMENT}    IN    @{ITEMS}
        \    Log    ${ELEMENT}
        \    Run Keyword If    '${ELEMENT}' == 'Break On Me'    Exit For Loop
        \    Log    Do more actions here ...

    Repeat-Action
        Repeat Keyword    2    Log    Repeating this ...
        
 **输出**：
 
    Starting test: StandardLoopDemo.For-Loop-In-Range
    20130426 11:24:14.389 :  INFO : 1
    20130426 11:24:14.390 :  INFO : ${RANDOM_STRING} = B
    20130426 11:24:14.390 :  INFO : B
    20130426 11:24:14.391 :  INFO : 2
    20130426 11:24:14.392 :  INFO : ${RANDOM_STRING} = ih
    20130426 11:24:14.392 :  INFO : ih
    Ending test:   StandardLoopDemo.For-Loop-In-Range

    Starting test: StandardLoopDemo.For-Loop-Elements
    20130426 11:24:14.394 :  INFO : @{ITEMS} = [ Star Trek | Star Wars | Perry Rhodan ]
    20130426 11:24:14.395 :  INFO : Star Trek
    20130426 11:24:14.396 :  INFO : ${ELEMENT} = StarTrek
    20130426 11:24:14.396 :  INFO : StarTrek
    20130426 11:24:14.397 :  INFO : Star Wars
    20130426 11:24:14.398 :  INFO : ${ELEMENT} = StarWars
    20130426 11:24:14.398 :  INFO : StarWars
    20130426 11:24:14.399 :  INFO : Perry Rhodan
    20130426 11:24:14.400 :  INFO : ${ELEMENT} = PerryRhodan
    20130426 11:24:14.400 :  INFO : PerryRhodan
    Ending test:   StandardLoopDemo.For-Loop-Elements

    Starting test: StandardLoopDemo.For-Loop-Exiting
    20130426 11:24:14.402 :  INFO : @{ITEMS} = [ Good Element 1 | Break On Me | Good Element 2 ]
    20130426 11:24:14.402 :  INFO : Good Element 1
    20130426 11:24:14.403 :  INFO : Do more actions here ...
    20130426 11:24:14.404 :  INFO : Break On Me
    Ending test:   StandardLoopDemo.For-Loop-Exiting

    Starting test: StandardLoopDemo.Repeat-Action
    20130426 11:24:14.408 :  INFO : Repeating keyword, round 1/2
    20130426 11:24:14.408 :  INFO : Repeating this ...
    20130426 11:24:14.408 :  INFO : Repeating keyword, round 2/2
    20130426 11:24:14.409 :  INFO : Repeating this ...
    Ending test:   StandardLoopDemo.Repeat-Action   

语法非常直接，不需要过多解释。唯一需要注意的是，循环体内的关键字必须用 '\' 来进行转义。


## 有关条件判断的关键字

在测试代码中使用条件判断会带来不少争议。不用担心，唯一应该记住的是，测试实现应该尽可能简单明了，不要混杂过多条件逻辑。

在下面的例子中将使用以下相关的关键字。

* **Run Keyword** - 这个关键字将其他关键字作为一个变量传入。这意味着，测试能够动态地改变执行时所使用的关键字，比如执行其他函数返回的关键字。
* **Run Keyword If** - 在测试复杂结构时非常有用，比如被测的 web 页面在输入不同时会有不同的选项。但是在测试中混有过多的程序结构会使 troubleshooting 变得困难。
* **Run Keyword And Ignore Error** - 哈，我还没有找到对应的实际例子。
* **Run Keyword If Test Failed** - 如果测试失败了可以用这个打一些 log，或者打一个快照。在 troubleshooting 时有用。

**测试**：

    *** Test Cases ***
    Run-Keyword
        ${MY_KEYWORD}=    Set Variable    Log
        Run Keyword    ${MY_KEYWORD}    Test

    Run-Keyword-If
        ${TYPE}=    Set Variable    V1
        Run Keyword If    '${TYPE}' == 'V1'    Log     Testing Variant 1
        Run Keyword If    '${TYPE}' == 'V2'    Log    Testing Variant 2
        Run Keyword If    '${TYPE}' == 'V3'    Log    Testing Variant 3

    Run-Keyword-Ignore-Error
        @{CAPTAINS}    Create List    Picard    Kirk    Archer
        Run Keyword And Ignore Error    Should Be Empty    ${CAPTAINS}
        Log    Reached this point despite of error

**输出**：

    Starting test: Robot Blog.StandardConditionDemo.Run-Keyword
    20130426 13:34:50.840 :  INFO : ${MY_KEYWORD} = Log
    20130426 13:34:50.841 :  INFO : Test
    Ending test:   Robot Blog.StandardConditionDemo.Run-Keyword

    Starting test: Robot Blog.StandardConditionDemo.Run-Keyword-If
    20130426 13:34:50.843 :  INFO : ${TYPE} = V1
    20130426 13:34:50.844 :  INFO : Testing Variant 1
    Ending test:   Robot Blog.StandardConditionDemo.Run-Keyword-If

    Starting test: Robot Blog.StandardConditionDemo.Run-Keyword-Ignore-Error
    20130426 13:34:50.847 :  INFO : @{CAPTAINS} = [ Picard | Kirk | Archer ]
    20130426 13:34:50.848 :  INFO : Length is 3
    20130426 13:34:50.849 :  FAIL : '[u'Picard', u'Kirk', u'Archer']' should be empty
    20130426 13:34:50.850 :  INFO : Reached this point despite of error
    Ending test:   Robot Blog.StandardConditionDemo.Run-Keyword-Ignore-Error


## 字符串和列表

可以看到，[Robot Framework](http://robotframework.org/)框架包含了完整的可编程结构。而一些高级语言特有的字符串和列表也能通过 Collection Library 和 String Library 来实现。

**测试**：

    *** Settings ***
    Library           String
    Library           Collections

    *** Test Cases ***
    StringsAndLists
        ${SOME_VALUE}=    Set Variable    "Test Value"
        Log    ${SOME_VALUE}
        @{WORDS}=    Split String    ${SOME_VALUE}    ${SPACE}
        ${FIRST}=    Get From List    ${WORDS}    0
        Log    ${FIRST}

**输出**：

    Starting test: Robot Blog.StandardStringsAndListsDemo.StringsAndLists
    20130506 21:21:05.880 :  INFO : ${SOME_VALUE} = "Test Value"
    20130506 21:21:05.881 :  INFO : "Test Value"
    20130506 21:21:05.882 :  INFO : @{WORDS} = [ "Test | Value" ]
    20130506 21:21:05.882 :  INFO : ${FIRST} = "Test
    20130506 21:21:05.883 :  INFO : "Test
    Ending test:   Robot Blog.StandardStringsAndListsDemo.StringsAndLists

## Robot Framework 教程目录

[原文链接](https://blog.codecentric.de/en/2013/05/robot-framework-tutorial-loops-conditional-execution-and-more/)

1. [Robot Framework 教程 - 概览（译）](http://www.lyyyuna.com/2016/01/07/robotframework-tutorial-overview/)
2. [Robot Framework 教程 - 一个完整的例子（译）](http://www.lyyyuna.com/2016/04/09/robotframework-tutorial-a-complete-example/)
3. [Robot Framework 教程 - 集成开发环境 RIDE 概览 (译)](http://www.lyyyuna.com/2016/04/30/robotframework-ide-ride-overview/)
4. [Robot Framework 教程 - 如何组织一个可伸缩可维护的验收测试套件（译）](http://www.lyyyuna.com/2016/05/15/robotframework-tutorial-how-to-structure-a-scalable-and-maintainable-acceptance-test-suite/)
5. [Robot Framework 教程 - 循环，条件判断，字符串和列表（译）](http://www.lyyyuna.com/2016/05/28/robotframework-tutorial-loops-conditional-execution-and-more/)