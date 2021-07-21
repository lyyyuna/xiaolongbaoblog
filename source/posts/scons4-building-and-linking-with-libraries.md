title: SCons 用户指南第四章 - 编译和链接库
date: 2015-12-25 20:11:03
categories: 杂
tags: scons
---

在大型软件工程中，经常将部分软件合并成一个或多个库文件。SCons 在创建和使用库文件方面非常简单易用。

## 4.1 编译库 

你可以代替 Program 使用 Library 来生成你自己的库文件：

    Library('foo', ['f1.c', 'f2.c', 'f3.c'])

SCons 会自动根据你的系统使用合适的库文件前缀和后缀。所以在 POSIX 或 Linux 系统上，上面的例子编译过程如下：

    % scons -Q
    cc -o f1.o -c f1.c
    cc -o f2.o -c f2.c
    cc -o f3.o -c f3.c
    ar rc libfoo.a f1.o f2.o f3.o
    ranlib libfoo.a

在 Windows 系统上，上述例子的编译过程如下：

    C:\>scons -Q
    cl /Fof1.obj /c f1.c /nologo
    cl /Fof2.obj /c f2.c /nologo
    cl /Fof3.obj /c f3.c /nologo
    lib /nologo /OUT:foo.lib f1.obj f2.obj f3.obj

目标库文件的命名规则类似于直接生成程序：如果你没有显示指定目标库文件名称，SCons 会从第一个指定的源文件名字中推导出，并且 SCons 会加上合适的文件前缀和后缀。

### 4.1.1 从源码或二进制文件编译库

上一个例子战士了如何从一系列源码中编译出库文件，然而你也可以在 Library 调用中放置二进制文件，它会自动分析出这些是二进制文件。事实上，你也可以在源队列中混合使用源代码文件和二进制文件：

    Library('foo', ['f1.c', 'f2.o', 'f3.c', 'f4.o'])
    
SCons 能够意识到在生成最终库文件前需要编译哪些源文件：

    % scons -Q
    cc -o f1.o -c f1.c
    cc -o f3.o -c f3.c
    ar rc libfoo.a f1.o f2.o f3.o f4.o
    ranlib libfoo.a
    
当然，在这个例子中，必须先编译出这些二进制文件。请查看第五章，如何显示编译自己的二进制文件并包含在自己的库文件中。

### 4.1.2 编译静态库：使用 StaticLibrary 方法

Library 方法能够编译一个传统的静态库文件。如果你想明确地控制库文件的类型，你可以使用 StaticLibrary 方法来代替 Library：

    StaticLibrary('foo', ['f1.c', 'f2.c', 'f3.c'])

在功能上它们之间没有区别。

### 4.1.3 编译动态库（DLL）：使用 SharedLibrary 方法

如果你像编译在 POSIX 系统或者 Windows 系统上编译动态库文件，你应该使用 SharedLibrary 方法：

    SharedLibrary('foo', ['f1.c', 'f2.c', 'f3.c'])

在 POSIX 系统上输出为：

    % scons -Q
    cc -o f1.os -c f1.c
    cc -o f2.os -c f2.c
    cc -o f3.os -c f3.c
    cc -o libfoo.so -shared f1.os f2.os f3.os
    
在 Windows 系统上输出为：

    C:\>scons -Q
    cl /Fof1.obj /c f1.c /nologo
    cl /Fof2.obj /c f2.c /nologo
    cl /Fof3.obj /c f3.c /nologo
    link /nologo /dll /out:foo.dll /implib:foo.lib f1.obj f2.obj f3.obj
    RegServerFunc(target, source, env)
    embedManifestDllCheck(target, source, env)
    
你可以注意到 SCons 能够正确处理输出文件的过程，在 POSIX 系统上假设 -shared 选项，在 Windows 系统上加上 /dll 选项。

## 4.2 链接库

通常，你之所以编译库文件是为了在多个程序中链接它。为了链接库，你只需要在 $LIBS 构造变量中指定库，在 $LIBPATH 构造变量中指定库文件所在的目录：

    Library('foo', ['f1.c', 'f2.c', 'f3.c'])
    Program('prog.c', LIBS=['foo', 'bar'], LIBPATH='.')
    
很显然可以注意到，你不需要指定库的前缀（比如 lib）或者后缀（比如 .a 或 .lib）。SCons 能够为当前的系统使用正确的前缀和后缀。

在类 POSIX 系统中，以上例子的编译过程为：

    % scons -Q
    cc -o f1.o -c f1.c
    cc -o f2.o -c f2.c
    cc -o f3.o -c f3.c
    ar rc libfoo.a f1.o f2.o f3.o
    ranlib libfoo.a
    cc -o prog.o -c prog.c
    cc -o prog prog.o -L. -lfoo -lbar
    
在 Windows 系统上，以上例子的编译过程为：

    C:\>scons -Q
    cl /Fof1.obj /c f1.c /nologo
    cl /Fof2.obj /c f2.c /nologo
    cl /Fof3.obj /c f3.c /nologo
    lib /nologo /OUT:foo.lib f1.obj f2.obj f3.obj
    cl /Foprog.obj /c prog.c /nologo
    link /nologo /OUT:prog.exe /LIBPATH:. foo.lib bar.lib prog.obj
    embedManifestExeCheck(target, source, env)
    
像之前一样，可以注意到 SCons 能够正确地构造命令行，以确保在各个系统上能够链接正确的库。

并且，如果你只有一个库需要链接，你可以用单个字符串，来代替 Python 的列表，即：

    Program('prog.c', LIBS='foo', LIBPATH='.')

这和以下是等价的：

    Program('prog.c', LIBS=['foo'], LIBPATH='.')
    
这和 SCons 在处理单个源文件中既可以使用字符串或者列表，是类似的。

## 4.3 寻找库：$LIBPATH 构造变量

链接器默认只在系统指定的路径中寻找库文件。当你指定了 $LIBPATH 构造变量，SCons 知道如何寻找库文件。$LIBPATH 由一系列目录路径构成，如下：

    Program('prog.c', LIBS = 'm', LIBPATH = ['/usr/lib', '/usr/local/lib'])
    
推荐使用 Python 列表，因为这具有可移植性。你也可以将所有目录名字放在一个字符串中。在 POSIX 系统中，使用冒号来分割目录路径：

    LIBPATH = '/usr/lib:/usr/local/lib'
    
或者在 Windows 系统上使用分号来分割：

    LIBPATH = 'C:\\lib;D:\\lib'
    
（注意到，在 Windows 系统上必须对反斜杠进行转义。）

当链接器执行时，SCons 会生成合适的标志位，这样链接器就将去寻找 SCons 指定的路径。所以在 POSIX 系统上，上述例子的编译过程如下：

    % scons -Q
    cc -o prog.o -c prog.c
    cc -o prog prog.o -L/usr/lib -L/usr/local/lib -lm
    
或者在 Windows 系统上，上述例子的编译过程如下：

    C:\>scons -Q
    cl /Foprog.obj /c prog.c /nologo
    link /nologo /OUT:prog.exe /LIBPATH:\usr\lib /LIBPATH:\usr\local\lib m.lib prog.obj
    embedManifestExeCheck(target, source, env)
    
注意到 SCons 能够正确地根据系统生成合适的命令行参数。