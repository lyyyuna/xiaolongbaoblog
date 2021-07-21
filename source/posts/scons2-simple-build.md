title: SCons 用户指南第二章-简单编译
date: 2015-11-02 07:33:47
categories: 杂
tags: scons
series: SCons 用户指南

---




在本章中，你将会看到使用 SCons 作配置的简单编译的例子。这些例子表现了使用 SCons 作为不同语言不同系统编译构建工具是非常简单的。

## 2.1 编译简单的 C/C++ 程序

以下是著名的 C 语言 "Hello, World!"：

    int
    main()
    {
        printf("Hello, world!\n");
    }

接下来是如何使用 SCons 进行编译。只要在名为 SConstruct 的文件中输入：

    Program('hello.c')

这个最小的配置文件告诉 SCons 两个信息：你想编译成什么（一个可执行的程序），和你的源文件（hello.c）。Program 是一个 builder_method，这是一个 Python 的调用，用来告诉 SCons 你想生成一个可执行的程序。

好了。现在使用 scons 命令来编译程序。在一个 POSIX 兼容的系统上，比如 Linux/UNIX，你将会看到如下的输出：

    % scons
    scons: Reading SConscript files ...
    scons: done reading SConscript files.
    scons: Building targets ...
    cc -o hello.o -c hello.c
    cc -o hello hello.o
    scons: done building targets.

在使用 Microsoft Visual C++ 编译器的 Windows 系统上，你会看到如下的输出：

    C:\>scons
    scons: Reading SConscript files ...
    scons: done reading SConscript files.
    scons: Building targets ...
    cl /Fohello.obj /c hello.c /nologo
    link /nologo /OUT:hello.exe hello.obj
    embedManifestExeCheck(target, source, env)
    scons: done building targets.

首先注意到，你只需要指出源文件的名字，SCons 就会正确推导出 object 和可执行文件的名字。

然后注意到，相同的 SConstruct 文件，不用修改就可以在不同系统生成正确的输出：在 POSIX 系统上生成 hello.o 和 hello，在 Windows 系统上生成 hello.obj 和 hello.exe。这是一个非常简单的例子，展示了 SCons 能够很容易地编写可移植的软件编译系统。

（注意到在本指南中，并不是每个例子都给出了 POSIX 和 Windows 的输出，然而除非特别声明，这些例子在每个系统上都应该是可工作的。）

## 2.2 编译目标文件

Program 编译方法只是 SCons 提供的众多编译方法的一种，他们可用来编译成各种文件。其中另外一个就是 Object 编译方法，能使 SCons 从指定的源文件编译成相应的目标文件：

    Object('hello.c')
      
现在你只要运行 scons 命令编译程序，就能在 POSIX 系统上得到名为 hello.o 的目标文件：

    % scons
    scons: Reading SConscript files ...
    scons: done reading SConscript files.
    scons: Building targets ...
    cc -o hello.o -c hello.c
    scons: done building targets.

在使用 Microsoft Visual C++ 编译器的 Windows 系统上，你会得到：

    C:\>scons
    scons: Reading SConscript files ...
    scons: done reading SConscript files.
    scons: Building targets ...
    cl /Fohello.obj /c hello.c /nologo
    scons: done building targets.

## 2.3 简单 Java 编译

SCons 也使得编译 Java 及其简单。然而不像 Program 和 Object 编译方法，Java编译方法需要你指定生成的 class 文件的目标路径，和你 .java 文件的源路径：

    Java('classes', 'src')

如果 src 目录只包含一个 hello.java 文件，那么使用 scons 命令的输出结果将是（POSIX）：

    % scons
    scons: Reading SConscript files ...
    scons: done reading SConscript files.
    scons: Building targets ...
    javac -d classes -sourcepath src src/hello.java
    scons: done building targets.

我们将在第 26 章讲述更多关于 Java 编译的内容，包括 .jar 和其他类型文件。

## 2.4 在编译之后清理

当使用 SCons 编译完成之后，不需要特别的命令来清理。相反，你只需要 -c 或者 --clean 选项，SCons 就会自动清理相关的编译生成文件。如果在上面的例子之后，我们调用 scons -c 命令，在 POSIX 系统上就会输出：

    % scons
    scons: Reading SConscript files ...
    scons: done reading SConscript files.
    scons: Building targets ...
    cc -o hello.o -c hello.c
    cc -o hello hello.o
    scons: done building targets.
    % scons -c
    scons: Reading SConscript files ...
    scons: done reading SConscript files.
    scons: Cleaning targets ...
    Removed hello.o
    Removed hello
    scons: done cleaning targets.

在 Windows 上就会输出：

    C:\>scons
    scons: Reading SConscript files ...
    scons: done reading SConscript files.
    scons: Building targets ...
    cl /Fohello.obj /c hello.c /nologo
    link /nologo /OUT:hello.exe hello.obj
    embedManifestExeCheck(target, source, env)
    scons: done building targets.
    C:\>scons -c
    scons: Reading SConscript files ...
    scons: done reading SConscript files.
    scons: Cleaning targets ...
    Removed hello.obj
    Removed hello.exe
    scons: done cleaning targets.

注意到 SCons 用 Cleaning targets ... 和 done cleaning targets 来告诉你清理过程。

## 2.5 SConstruct 文件

如果你曾经使用过类似 Make 的编译系统，你可能会想 SConstruct 文件是否等同于 Makefile 文件。是的，SConstruct 文件是 SCons 用来控制编译过程的。

### 2.5.1 SConstruct 文件是 Python 脚本文件

然而 SConstruct 文件和 Makefile 文件一个重要的区别就是：SConstruct 文件实际上就是一个 Python 脚本文件。如果你还不熟悉 Python，不要着急。本用户指南会介绍一下基本的 Python 语法来有效地使用 SCons。而且，Python 非常容易学习。

使用 Python 脚本语言的一方面就是你可以在 SConstruct 文件中使用 Python 的注释，那就是 '#'，该行之后的都会被忽略：

    # Arrange to build the "hello" program.
    Program('hello.c')    # "hello.c" is the source file.

你可以在本指南剩余部分看到，一个真正的脚本语言，可以使用极简单的方法来满足现实世界的复杂需求。

### 2.5.2 SCons 函数是顺序无关的

SConstruct 文件更像 Makefile 文件也是区别于普通 Python 脚本的重要一点就是，SCons 函数的调用顺序并不影响实际编译过程中目标文件的编译顺序。换句话说，当你调用 Program 编译方法（或者其他编译方法）时，你并不是告诉在此立即编译程序。相反，你只是告诉你需要编译的程序，举例来说，如果你想编译 hello.c，完全由 SCons 来决定何时编译程序。（我们将在第六章-依赖中学习 SCons 何时编译和重新编译）。

当 SCons 读入 SConstruct 文件和实际编译程序的时候， SCons 通过在输出的状态信息中反映出调用编译方法如 Program 和实际编译程序过程的区别。这能够反映出 SCons 到底是在执行 Python 语句还是在执行编译命令。

让我们来用一个例子说明这一点。Python 有一个 print 语句能够在将字符串打印在屏幕上。如果 print 语句放在 Program 编译方法之间：

    print "Calling Program('hello.c')"
    Program('hello.c')
    print "Calling Program('goodbye.c')"
    Program('goodbye.c')
    print "Finished calling Program()"

然后当我们调用 SCons 时，我们能够看到打印的 print 语句，指示出这些语句是何时执行的：

    % scons
    scons: Reading SConscript files ...
    Calling Program('hello.c')
    Calling Program('goodbye.c')
    Finished calling Program()
    scons: done reading SConscript files.
    scons: Building targets ...
    cc -o goodbye.o -c goodbye.c
    cc -o goodbye goodbye.o
    cc -o hello.o -c hello.c
    cc -o hello hello.o
    scons: done building targets.

注意到尽管我们先调用了 Program('hello')，但是 goodbye 程序先被编译了，

## 2.6 使 SCons 输出更简洁

你已经看见了 SCons 在实际编译指令周围使用了一些信息来指示出它当前执行的动作：

    C:\>scons
    scons: Reading SConscript files ...
    scons: done reading SConscript files.
    scons: Building targets ...
    cl /Fohello.obj /c hello.c /nologo
    link /nologo /OUT:hello.exe hello.obj
    embedManifestExeCheck(target, source, env)
    scons: done building targets.

这些信息强调了 SCons 工作的顺序：所有的配置文件（大部分是 SConscript 文件）首先被读入和执行，然后是编译目标文件。除了其他好处，这些信息还帮助区分是读入配置错误还是编译过程中发生的错误。

一个显而易见的缺点是这些信息混淆了输出。幸运的是你可以使用 -Q 选项来取消它们：

    C:\>scons -Q
    cl /Fohello.obj /c hello.c /nologo
    link /nologo /OUT:hello.exe hello.obj
    embedManifestExeCheck(target, source, env)

因为本指南想注重于 SCons de 实际工作过程，我们将在剩余的例子中继续使用 -Q 选项来移除这些信息。
