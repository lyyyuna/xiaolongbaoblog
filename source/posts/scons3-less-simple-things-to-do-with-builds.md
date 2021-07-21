title: SCons 用户指南第三章-让 SCons 编译更简单
date: 2015-12-03 21:47:19
categories: 杂
tags: scons
series: SCons 用户指南

---



在本章中，你将看到一些使用 SCons 编译配置的简单例子。这些例子表明，在不同系统和不同语言上使用 SCons 都是非常易用的。

## 3.1 指定目标文件的名称

在前几章的例子中，你已经看到了使用 Program 编译方法生成的目标文件名称和源文件相同。也就是说，如下的命令在 POSIX 系统上将生成可执行文件 hello，而在 Windows 上将生成 hello.exe。

    Program('hello.c')

如果你想要自定义名称，你只需要在调用时在源文件左侧指定就行：

    Program('new_hello', 'hello.c')

（SCons 要求目标文件的名称放在第一个，然后是源文件的名称，因此这个顺序模仿了大部分编程语言中的赋值顺序，包括 Python: "program = source files"）

如下的命令将在 POSIX 系统上生成一个可执行程序 new_hello：

    % scons -Q
    cc -o hello.o -c hello.c
    cc -o new_hello hello.o

如下的命令将在 Windows 系统上生成可执行程序 new_hello.exe：

    C:\>scons -Q
    cl /Fohello.obj /c hello.c /nologo
    link /nologo /OUT:new_hello.exe hello.obj
    embedManifestExeCheck(target, source, env)

## 3.2 编译多个源文件

你已经看到了如何使用 SCons 从单个源文件编译的例子。在实际开发过程中，更多的是编译多个源文件的例子。为此，你只需要将源文件放在 Python list 中即可，如下所示：

    Program(['prog.c', 'file1.c', 'file2.c'])

编译过程如下：

    % scons -Q
    cc -o file1.o -c file1.c
    cc -o file2.o -c file2.c
    cc -o prog.o -c prog.c
    cc -o prog prog.o file1.o file2.o

可以看到，SCons 自动从列表的第一个源文件名推导出可执行程序的名字。由于第一个源文件名为 prog.c，SCons 会将生成文件命名为 prog （或者在 Windows 上是 prog.exe）。如果你想指定不同的名字，你只需要将源文件列表放在第二个参数，在第一个参数放置可执行程序名：

    Program('program', ['prog.c', 'file1.c', 'file2.c'])

在 Linux 上，编译过程如下：

    % scons -Q
    cc -o file1.o -c file1.c
    cc -o file2.o -c file2.c
    cc -o prog.o -c prog.c
    cc -o program prog.o file1.o file2.o

在 Windows 上：

    C:\>scons -Q
    cl /Fofile1.obj /c file1.c /nologo
    cl /Fofile2.obj /c file2.c /nologo
    cl /Foprog.obj /c prog.c /nologo
    link /nologo /OUT:program.exe prog.obj file1.obj file2.obj
    embedManifestExeCheck(target, source, env)

## 3.3 用 Glob 函数生成文件列表

你也可以使用 Glob 函数来匹配出所有符合特定模式的源文件。模式可以是标准 shell 的模式匹配格式，比如 *, ?, [abc], [!abc]。这种方法使得在写多源文件时变得非常容易。

    Program('program', Glob('*.c'))

## 3.4 指定单个文件 Vs. 文件列表

我们已经展示了两种指定程序源文件的方法，第一种是使用文件的列表：

    Program('hello', ['file1.c', 'file2.c'])

第二种是使用单个文件：

    Program('hello', 'hello.c')

你实际上也可以将单个文件名放入一个列表中。这样你的脚本格式能够保持一致：

    Program('hello', ['hello.c'])

SCons 函数能够接受以上任何一种形式。实际上 SCons 将输入都看作是文件列表，但是允许你在输入单个源文件时省略方括号。

### 重要

虽然 SCons 函数不严格区分字符串和列表，但 Python 本身对这两者严格区分。所以在 SCons 允许既可以字符串和列表时：

    # The following two calls both work correctly:
    Program('program1', 'program1.c')
    Program('program2', ['program2.c'])

如果对混有字符串和列表的情况做一些 Python 的操作，可能会导致发生错误和不正确的结果：

    common_sources = ['file1.c', 'file2.c']

    # THE FOLLOWING IS INCORRECT AND GENERATES A PYTHON ERROR
    # BECAUSE IT TRIES TO ADD A STRING TO A LIST:
    Program('program1', common_sources + 'program1.c')

    # The following works correctly, because it's adding two
    # lists together to make another list.
    Program('program2', common_sources + ['program2.c'])

## 3.5 让文件列表更容易阅读

使用 Python 列表的一打缺点就是每个源文件名都必须包含在引号之间。当文件名很长时，就会显得累赘且不易阅读。幸运的是，SCons 和 Python 都提供了多种方式来简化 SConstruct 的阅读。

为了使长文件列表易于处理，SCons 提供了 Split 函数来处理被引号包含且用空格或其他空白字符分隔的文件名，处理之后将转换成文件列表。对前一个例子使用 Split 函数：

    Program('program', Split('main.c file1.c file2.c'))

（你如果熟悉 Python, 你应该已经意识到这和 string 模块中的 split() 函数非常相似。和 split() 成员函数不同的是， Split 函数并不要求输入一定是一个字符串，并将列表中单个字符串包起来。如果对象已经是一个列表，则直接原样返回不变。对于任意值传递到 SCons 函数中，这是一个方便的方法，这有就无需手工检查该变量的类型。）

将 Split 函数放在 Program 调用中仍然显得笨拙。一个易读的方法是将 Split 调用的输出赋值给一个变量，然后将变量传递给 Program 函数：

    src_files = Split('main.c file1.c file2.c')
    Program('program', src_files)

最后需说明的是，Split 函数并不关心文件名之间空格的长度，这样你可以创建跨越多行的文件列表字符串，更容易编辑：

    src_files = Split("""main.c
                         file1.c
                         file2.c""")
    Program('program', src_files)

（这个例子中使用了 Python 的三重引号表示法，使得字符串能跨越多行，这里既可以是单引号也可以是双引号。）

## 3.6 关键字参数

SCons 也允许你使用 Python 的关键字参数来指定输出文件和输入源文件。输出文件是 target，输入文件是 source。Python 语法如下：

    src_files = Split('main.c file1.c file2.c')
    Program(target = 'program', source = src_files)

由于关键字显示指定了参数，所以实际上如果你喜欢，你可以颠倒它们的顺序：

    src_files = Split('main.c file1.c file2.c')
    Program(source = src_files, target = 'program')

至于你是否用关键字参数纯属个人喜好，SCons 函数的执行效果完全相同。

## 3.7 编译多个程序

为了能够在下单个 SConstruct 文件中编译多个程序， 你只需要调用 Program 方法多次即可，每一个输出程序对应一个调用。

    Program('foo.c')
    Program('bar', ['bar1.c', 'bar2.c'])

SCons 的编译输出如下：

    % scons -Q
    cc -o bar1.o -c bar1.c
    cc -o bar2.o -c bar2.c
    cc -o bar bar1.o bar2.o
    cc -o foo.o -c foo.c
    cc -o foo foo.o

注意到 SCons 编译的顺序和 SConstruct 文件中指定的顺序并不相同。实际上 SCons 能够自动识别出单个二进制文件的编译顺序。我们将在下面的“_依赖_”章节中详细阐述。

## 3.8 在多个编译程序中共享源文件

在多个编译程序中共享源文件是常见的复用代码的方法。其中一个方法就是根据常用的源文件生成库，然后让最终生成程序都链接该库。（创建库将在第 4 章，生成和链接库 中详述。）

更直接，可能稍微麻烦的方法是将共同的源文件都放入各自的源文件列表中：

    Program(Split('foo.c common1.c common2.c'))
    Program('bar', Split('bar1.c bar2.c common1.c common2.c'))

即使 common1.c 和 common2.c 生成的二进制文件被链接了两次，SCons 也能够自动识别出它们只需要分别编译一次即可：

    % scons -Q
    cc -o bar1.o -c bar1.c
    cc -o bar2.o -c bar2.c
    cc -o common1.o -c common1.c
    cc -o common2.o -c common2.c
    cc -o bar bar1.o bar2.o common1.o common2.o
    cc -o foo.o -c foo.c
    cc -o foo foo.o common1.o common2.o

如果两个或多个程序共享了大量的源文件，那么重复输入源文件名会带来维护上的问题。你可以使用另外一个 Python 列表来存储共同源文件名，然后使用 Python 的 + 运算符进行链接：

    common = ['common1.c', 'common2.c']
    foo_files = ['foo.c'] + common
    bar_files = ['bar1.c', 'bar2.c'] + common
    Program('foo', foo_files)
    Program('bar', bar_files)
        
这和上一个例子功能完全相同。