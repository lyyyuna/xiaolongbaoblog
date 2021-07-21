title: Robot Framework 模板与数据驱动测试
date: 2016-07-29 19:59:52
categories: 测试
tags: 
- robot framework
---


Robot Framework 是关键字驱动的测试框架，虽然关键字驱动高度抽象了底层实现，减少维护成本，降低了对测试人员编程水平的需求，但在某些类型的测试中，数据驱动导向的测试用例比重多，比如常见的用户输入框就有海量的输入可能性。Robot Framework 提供了测试模板，可以将其转换为数据驱动的测试。

## 基本用法

如果有一个接受参数的关键字，那么它就可以被用作模板。下面的例子展示了这一点。

    *** Settings ***
    Test Setup        Prepare
    Test Template     Compare Two Number ${one} ${three}

    *** Test Cases ***
    Template test case
        1    2
        1    1
        2    3
        2    2

    *** Keywords ***
    Compare Two Number ${one} ${three}
        Should Be Equal    ${one}    ${three}


这里展示了比较两个数是否相等的例子，可以看到只需填入输入数据即可。你也可以用 [Template] 为每个 Test Case 单独指定模板。


## 与循环执行的区别

有人会问，能否运用循环语句来模拟上述行为呢？

首先，由于 Robot Framework 是个测试框架，编程能力被弱化不少，模板语法显得简洁 ^_^，然后，在用例中混入过多的执行控制流也不是推荐的行为（或者绝对地说，用例中就不应该有循环、判断语句）。

其次，模板是处于 continue on failure 模式中，某一项输入 Fail，还会继续执行其他输入。普通 Case 一旦有个语句 Fail，该 Case 就会 tear down。比如上面给的例子，第一行和第二行就会 Fail，实际执行结果如下：

    # log
    Starting test: Testcases.UI Test.Template test case
    20160729 12:08:56.652 :  FAIL : 1 != 2
    20160729 12:08:56.668 :  FAIL : 2 != 3
    Ending test:   Testcases.UI Test.Template test case

    # report
    ========================================================
    Testcases                                                                                                                                                                
    ========================================================
    Testcases.UI Test                                                                                                                                                        
    ========================================================
    Template test case                                                                                                                                               | FAIL |
    Several failures occurred:

    1) 1 != 2

    2) 2 != 3
    --------------------------------------------------------
    Testcases.UI Test                                                                                                                                                | FAIL |
    1 critical test, 0 passed, 1 failed
    1 test total, 0 passed, 1 failed
    ========================================================
    Testcases                                                                                                                                                        | FAIL |
    1 critical test, 0 passed, 1 failed
    1 test total, 0 passed, 1 failed
    ========================================================

单个输入中的 Fail 不会中断执行流。


## 一些不足

这些不足是我自己感受，可能并不准确。

我自己在自动化测试中使用数据驱动测试方法，只是希望减轻手工编写的工作量，对于执行流上的步骤不想简化。我希望每个输入都能完整地走完一遍 Test Setup | Test Execuation | Test Teardown 过程，遗憾的是好像 Robot Framework 做不到。下面是例子，这里增加了一个 Test Setup。

    *** Settings ***
    Test Setup        Prepare
    Test Template     Compare Two Number ${one} ${three}

    *** Test Cases ***
    Template test case
        1    2
        1    1
        2    3
        2    2

    *** Keywords ***
    Compare Two Number ${one} ${three}
        Should Be Equal    ${one}    ${three}

    Prepare
        Log    hello, world

可以看到 hello, world 只打印了一次。

    Starting test: Testcases.UI Test.Template test case
    20160729 12:08:56.652 :  INFO : hello, world
    20160729 12:08:56.652 :  FAIL : 1 != 2
    20160729 12:08:56.668 :  FAIL : 2 != 3
    Ending test:   Testcases.UI Test.Template test case

这就使得框架对测试输入有要求：输入数据不能对后续输入有影响。一旦在执行过程中有 Fail 发生，就无法用 Test Teardown 恢复测试环境（Run Keyword And Continue On Failure 关键字可能可以解决该问题，但这样写逻辑上并不清晰）。

使用 Library 导入的标准库和外部库也有问题，测试框架会为每个 Case 生成一个库的实例（即 Python, Java 类的实例），模板中每一行输入都共享一个实例，若是类中有全局变量，便会在各个输入之间产生干扰。

