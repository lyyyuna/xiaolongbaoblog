title: Robot Framework 快速入门指南
date: 2015-12-28 20:20:34
categories: 测试
tags: robot framework
---

## 前言

### 关于本指南

《Robot Framework 快速入门指南》介绍了 [Robot Framework](http://robotframework.org/) 的一些最重要的特性。你不仅可以打开并浏览这些例子，而且你也可以把本指南当成一个 [可执行的演示程序](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#executing-this-guide)。所有这些特性在《[Robot Framework 用户指南](http://robotframework.org/robotframework/#user-guide)》中有详细的介绍。

### Robot Framework 概览

[Robot Framework](http://robotframework.org/) 是一个通用的开源自动化测试框架，常被用作验收测试 (acceptance testing) 和验收测试驱动开发 (acceptance test-driven development, ATDD)。它有着易用的表格测试数据语法，并且采用了关键字驱动的测试方法。你可以使用 Python 或者 Java 编写的测试库莱扩展框架的测试能力，而且用户可以使用和测试用例相同的语法，来创建新的高级关键字。

Robot Framework 独立于操作系统和应用。核心框架采用 Python 编写，同样能够在 Jython (JVM) 和 IronPython (.NET) 上运行。该框架有着丰富的生态系统，包含多种多样独立开发的通用测试库和工具。

有关 Robot Framework 极其生态系统更多的信息，你可以浏览 [http://robotframework.org/](http://robotframework.org/)。在那，你可以看到更多丰富的文档，演示程序，测试库和其他的工具列表，等等。

### 演示程序

在本指南中的示例应用程序是一个经典的登录示例的变体：这是用 Python 编写的基于命令行的身份验证服务器。该应用程序允许用户做三件事：

* 用有效的密码创建一个新账户；
* 使用有效的账户和密码登陆；
* 用已有的账户改变密码。

应用程序本身是文件 [sut/login.py](https://github.com/robotframework/QuickStartGuide/blob/master/sut/login.py) 中，可以直接执行命令 python sut/login.py。如果试图用一个不存在的用户帐户，或者无效的密码登陆，都会显示如下的错误消息：

    > python sut/login.py login nobody P4ssw0rd
    Access Denied
    
当创建一个有效帐户和密码，并登录成功后： 

    > python sut/login.py create fred P4ssw0rd
    SUCCESS

    > python sut/login.py login fred P4ssw0rd
    Logged In
    
当使用无效凭据更改密码，显示的错误消息和之前相同。新密码会进行有效性验证，如果无效，则给出如下错误消息：   

    > python sut/login.py change-password fred wrong NewP4ss
    Changing password failed: Access Denied

    > python sut/login.py change-password fred P4ssw0rd short
    Changing password failed: Password must be 7-12 characters long

    > python sut/login.py change-password fred P4ssw0rd NewP4ss
    SUCCESS 
    
该应用程序使用一个简单的数据库文件来追踪用户状态。该文件位于操作系统相关的临时目录。

## 执行演示程序

这些说明会解释如何自己运行本指南的演示程序，如果你不感兴趣，你仍然可以查看 [在线结果](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#viewing-results)。

### 安装

推荐使用 [Python](http://python.org/) 的 [pip](http://pip-installer.org/) 工具来安装 Robot Framework。你可以直接运行：

    pip install robotframework
    
你可以查看 [Robot Framework 安装指南](https://github.com/robotframework/robotframework/blob/master/INSTALL.rst)，了解更多的安装方法，和有关安装的更多一般信息。

这个示例是使用 [reStructuredText](http://docutils.sourceforge.net/rst.html) 标记语言书写，采用框架的代码块中的测试数据。想要用这种格式执行本测试需要安装 [docutils](Robot Framework test data in code blocks) 模块：

    pip install docutils
    
请注意，目前官方还不支持 Python3。请查阅 [安装指南](https://github.com/robotframework/robotframework/blob/master/INSTALL.rst) 了解非官方的 Python3 移植，极其最近的支持程度。

### 执行

安装完以后，你还需要获取演示实例。最方便的方法就是下载一个 [发布版本](https://github.com/robotframework/QuickStartGuide/releases) 或者获取 [最新版本](https://github.com/robotframework/QuickStartGuide/archive/master.zip) 后在任意位置解压，你也可以直接克隆这个 [仓库](https://github.com/robotframework/QuickStartGuide)。

当你安装完，并获取一切必要条件之后，你可以用 pybot 命令运行本演示：

    pybot QuickStart.rst
    
你还可以配置各种命令行选项来执行：

    pybot --log custom_log.html --name Custom_Name QuickStart.rst
    
运行 pybot --help 来获取可用的选项列表。

### 查看结果

运行演示程序将会生成以下三个结果文件。这些文件被在线链接到可用的预执行文件，但运行演示程序时会在本地创建它们。

* [report.html](http://robotframework.org/QuickStartGuide/report.html) 测试报告
* [log.html](http://robotframework.org/QuickStartGuide/log.html) 详细的测试执行 log
* [output.xml](http://robotframework.org/QuickStartGuide/output.xml) XML 格式的机读报告

## 测试用例

### 工作流测试

Robot Framework 测试用例是使用简单的表格语法创建。例如，下面的表有两个测试：

* 用户能够创建账户并登录
* 用户不能使用错误的密码登录


    *** Test Cases ***
    User can create an account and log in
        Create Valid User    fred    P4ssw0rd
        Attempt to Login with Credentials    fred    P4ssw0rd
        Status Should Be    Logged In

    User cannot log in with bad password
        Create Valid User    betty    P4ssw0rd
        Attempt to Login with Credentials    betty    wrong
        Status Should Be    Access Denied
        
请注意，这些测试读起来就像是用英语书写的手动测试步骤，而并不像是自动化测试。这是因为 Robot Framework 采用关键字驱动的测试方法，使得编写测试时能够使用自然语言，来描述过程的步骤和预期结果。测试用例是由关键字和其可能的参数构成。

### 高级别测试

测试用例也可以使用更抽象的关键字创建，这些例子没有任何位置参数。这允许使用更自由的文字，甚至方便那些不懂技术的客户或其他项目利益相关者互相交流。在 [验证测试驱动开发](http://en.wikipedia.org/wiki/Acceptance_test-driven_development) 或其他变体中这种方法尤其重要。

Robot Framework 并不强制要求测试用例的格式。一个常见的风格是使用 given-when-then 风格，即 [行为驱动开发](http://en.wikipedia.org/wiki/Behavior_driven_development) (behavior-driven development, BDD)。

    *** Test Cases ***
    User can change password
        Given a user has a valid account
        When she changes her password
        Then she can log in with the new password
        And she cannot use the old password anymore

### 数据驱动测试

我们经常会碰到这种情况，测试用例相似仅仅是输入或输出数据不同。在这种情况下_数据驱动测试_允许不同的测试数据，而无需重复工作流。在 Robot Framework 中使用 [Template] 设置，可以将用例转换为数据驱动的测试，而用例中定义的数据作为模板关键字运行：

    *** Test Cases ***
    Invalid password
        [Template]    Creating user with invalid password should fail
        abCD5            ${PWD INVALID LENGTH}
        abCD567890123    ${PWD INVALID LENGTH}
        123DEFG          ${PWD INVALID CONTENT}
        abcd56789        ${PWD INVALID CONTENT}
        AbCdEfGh         ${PWD INVALID CONTENT}
        abCD56+          ${PWD INVALID CONTENT}

除了在单独的测试中使用 [Template] 设置，也可以在之后介绍的 [启动和卸载](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#setups-and-teardowns) 的设置表格中使用 Test Template 选项。在本例中，模板能够避免为了过长或过短密码，再去创建其他无效的用例。如果不使用模板，你就不得不为每一个输入输出创建一个用例，而使用模板只要一个用例。

请注意，上述例子的错误消息是使用 [变量](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#variables) 来指定。

## 关键字

测试用例的关键字有两种来源。[库关键字](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#library-keywords) 来自导入的测试库，所谓的 [用户关键字](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#user-keywords) 可由和创建测试用例相同的表格语法来书写。

### 库关键字

所有在测试库中定义的低级关键字都是由标准编程语言实现的，通常是 Python 或者 Java。Robot Framework 有着丰富的 [测试库](http://robotframework.org/#test-libraries)，它们被分成_标准库_，_外部库_和_自定义库_。标准库分布于核心框架和一些通用库中，比如 OperatingSystem, Screenshot 和 BuiltIn 中。 这些库比较特别，当安装完框架后便是可用的。而如用于网络测试的 [Selenium2Library](https://github.com/rtomac/robotframework-selenium2library/#readme)，必须单独安装。如果这些库还不够用， [创建自定义库](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#creating-test-libraries) 是非常简单的。

为了使用一个测试库提供的关键字，必须先导入该库。本指南中的测试需要 OperatingSystem 库 (如 Remove File) 和自定义库 LoginLibrary 库 (如 Attempt to login with credentials)。它们都在如下的设置表格中导入：

    *** Settings ***
    Library           OperatingSystem
    Library           lib/LoginLibrary.py

### 用户关键字

Robot Framework 最强大的特性是允许用户使用已有关键字来创建新的高级关键字。这种语法被称作_用户自定义关键字_，或者简称为_用户关键字_。关键字语法和创建测试用例是类似的。上述例子中所有的高级关键字都在如下的关键字表格中创建：

    *** Keywords ***
    Clear login database
        Remove file    ${DATABASE FILE}

    Create valid user
        [Arguments]    ${username}    ${password}
        Create user    ${username}    ${password}
        Status should be    SUCCESS

    Creating user with invalid password should fail
        [Arguments]    ${password}    ${error}
        Create user    example    ${password}
        Status should be    Creating user failed: ${error}

    Login
        [Arguments]    ${username}    ${password}
        Attempt to login with credentials    ${username}    ${password}
        Status should be    Logged In

    # Keywords below used by higher level tests. Notice how given/when/then/and
    # prefixes can be dropped. And this is a commend.

    A user has a valid account
        Create valid user    ${USERNAME}    ${PASSWORD}

    She changes her password
        Change password    ${USERNAME}    ${PASSWORD}    ${NEW PASSWORD}
        Status should be    SUCCESS

    She can log in with the new password
        Login    ${USERNAME}    ${NEW PASSWORD}

    She cannot use the old password anymore
        Attempt to login with credentials    ${USERNAME}    ${PASSWORD}
        Status should be    Access Denied

用户自定义关键字可以包含其他用户关键字，或者是库关键字。正如你在本例中看到的那样，用户自定义关键字可以包含参数。它们还可以返回值，甚至是包括 FOR 循环。现在，重要的是，用户自定义关键字使得测试作者能过复用之前相同的步骤序列。用户自定义关键字也使得测试作者保持测试用例具有高可读性，并在不同场景下使用合适的抽象级别。

## 变量

### 定义变量

变量是 Robot Framework 的一个组成部分。通常在测试中使用的任何数据，如有更改，最好定义为变量。变量的定义语法非常简单，如下变量表所示：

    *** Variables ***
    ${USERNAME}               janedoe
    ${PASSWORD}               J4n3D0e
    ${NEW PASSWORD}           e0D3n4J
    ${DATABASE FILE}          ${TEMPDIR}${/}robotframework-quickstart-db.txt
    ${PWD INVALID LENGTH}     Password must be 7-12 characters long
    ${PWD INVALID CONTENT}    Password must be a combination of lowercase and uppercase letters and numbers

变量还可以由命令行给出，这对不同环境中执行测试用例是非常有用的。例如，可以如下执行演示程序：

    pybot --variable USERNAME:johndoe --variable PASSWORD:J0hnD0e QuickStart.rst

除用户定义的变量外，有一些是始终可用的内置变量。这些变量包括 ${TEMPDIR} 和 ${/}，这在上述示例中被使用。

### 使用变量

变量可以在测试数据的大多数地方使用。他们最常被用作关键字的参数，像下面的测试用例演示的那样。返回值分配给变量和以后使用。例如，如下的 Database Should Contain [用户关键字](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#user-keywords)，将数据库内容设置在 ${database} 变量中，然后使用 [BuiltIn](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#user-keywords) 关键字 Should Contain 来验证内容的正确性。库和用户关键字都能返回值。

    *** Test Cases ***
    User status is stored in database
        [Tags]    variables    database
        Create Valid User    ${USERNAME}    ${PASSWORD}
        Database Should Contain    ${USERNAME}    ${PASSWORD}    Inactive
        Login    ${USERNAME}    ${PASSWORD}
        Database Should Contain    ${USERNAME}    ${PASSWORD}    Active

    *** Keywords ***
    Database Should Contain
        [Arguments]    ${username}    ${password}    ${status}
        ${database} =     Get File    ${DATABASE FILE}
        Should Contain    ${database}    ${username}\t${password}\t${status}\n

## 组织测试用例

### 测试套件

测试用例的集合被称为 Robot Framework 中的测试套件。每个输入文件，该文件包含测试用例构成一个测试套件。当 [执行本指南](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#executing-this-guide)，你可以看到控制台输出的 QuickStart。这个名字就是从文件名中生成的，在报告和日志中也可以看到。


也可以通过层次结构来组织测试用例，将测试用例文件放入目录，再将这些目录放置到其他目录中。所有这些目录会自动创建高级测试套件，目录名即为测试套件的名字。由于测试套件只是文件和目录，它们可以很方便地放置在任何版本控制系统中。

### 安装和卸载

如果要在每个测试用例想在之前或之后要执行的某些关键字，可以在 Test Setup 和 Test Teardown 中设置。同样的，如果要在测试套件之前或之后执行某些关键字，你只需要在 Suite Setup 和 Suite Teardown 中设置。 单个测试也可以通过测试用例表格的 [Setup] 和 [Teardown] 使用自定义安装或卸载。这和之前 [数据驱动测试](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#data-driven-tests) 使用 [Template] 的方法是相同的。

在本演示中我们要确保在开始执行之前，每个测试之后，数据库被清除：

    *** Settings ***
    Suite Setup       Clear Login Database
    Test Teardown     Clear Login Database

### 使用标签

Robot Framework 允许设置测试用例标签，赋予其元数据。标签可以 Force Tags  和  Default Tags 为文件中所有测试用例强制设置，如下表中的所有测试用例。还可以通过 [Tags] 为单个测试用例设置，比如 [之前](https://github.com/robotframework/QuickStartGuide/blob/master/QuickStart.rst#using-variables) 的 User status is stored in database 测试。

    *** Settings ***
    Force Tags        quickstart
    Default Tags      example    smoke

当你在执行完后查看报告，你可以看到每个测试用例有指定的标签，也可以看到针对每个标签生成的统计信息。标签也可以用于许多其他用途，其中最重要的是有选择地执行测试。你可以试试，比如下列命令：

    pybot --include smoke QuickStart.rst
    pybot --exclude database QuickStart.rst

## 创建测试库

Robot Framework 提供了一个简单的 API 来使用 Python 或 Java 创建测试库，一些远程库接口还允许使用其他编程语言。《[Robot Framework 用户指南](http://robotframework.org/robotframework/#user-guide)》包含有关库 API 的详细说明。

举例来说，我们来看一下本例中的 LoginLibrary 测试库。库位于 [lib/LoginLibrary.py](https://github.com/robotframework/QuickStartGuide/blob/master/lib/LoginLibrary.py)，源代码如下。通过源代码可以看到，关键字 Create User 如何映射到实际的执行方法 create_user。

    import os.path
    import subprocess
    import sys


    class LoginLibrary(object):

        def __init__(self):
            self._sut_path = os.path.join(os.path.dirname(__file__),
                                        '..', 'sut', 'login.py')
            self._status = ''

        def create_user(self, username, password):
            self._run_command('create', username, password)

        def change_password(self, username, old_pwd, new_pwd):
            self._run_command('change-password', username, old_pwd, new_pwd)

        def attempt_to_login_with_credentials(self, username, password):
            self._run_command('login', username, password)

        def status_should_be(self, expected_status):
            if expected_status != self._status:
                raise AssertionError("Expected status to be '%s' but was '%s'."
                                    % (expected_status, self._status))

        def _run_command(self, command, *args):
            command = [sys.executable, self._sut_path, command] + list(args)
            process = subprocess.Popen(command, stdout=subprocess.PIPE,
                                    stderr=subprocess.STDOUT)
            self._status = process.communicate()[0].strip()

