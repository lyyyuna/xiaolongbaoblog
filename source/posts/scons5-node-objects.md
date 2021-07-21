title: SCons 用户指南第五章 - 节点对象
date: 2016-01-03 07:33:47
categories: 杂
tags: scons
---

SCons 在内部把文件和目录都表示成节点。灵活运用内部对象（非二进制文件），可以让你的 SConscript 脚本文件更容易移植且易读。

## 5.1 编译方法会返回目标文件节点的列表

所有编译方法都会返回节点对象的列表，它们代表了即将被编译的目标文件。这些的节点对象又可以被当作参数传给其他的编译方法。

举例来说，假如我们想使用不同选项来编译由两个二进制文件所构成的程序。那意味着，我们会对每一个源文件指定不同选项，调用 Object 编译方法：

    Object('hello.c', CCFLAGS='-DHELLO')
    Object('goodbye.c', CCFLAGS='-DGOODBYE')

接着，会调用 Program 编译方法来生成最终程序。在调用时，我们会在参数中列出二进制文件的名字：

    Object('hello.c', CCFLAGS='-DHELLO')
    Object('goodbye.c', CCFLAGS='-DGOODBYE')
    Program(['hello.o', 'goodbye.o'])

问题来了，我们在 SConstruct 文件中直接用字符串中硬编码了文件名称，这回使得脚本在不同操作系统间缺乏移植性。比如在 Windows 系统中，第一步生成的二进制文件名称为 hello.obj 和 goodbye.obj，而不是 hello.o 和 goodbye.o。

一个更好的解决方案是将 Object 编译方法返回的目标列表存储在变量中，然后将变量直接用于 Program 方法：

    hello_list = Object('hello.c', CCFLAGS='-DHELLO')
    goodbye_list = Object('goodbye.c', CCFLAGS='-DGOODBYE')
    Program(hello_list + goodbye_list)
    
这样，我们的 SConstruct 文件具备了可移植性，在 Linux 系统上编译输出如下：

    % scons -Q
    cc -o goodbye.o -c -DGOODBYE goodbye.c
    cc -o hello.o -c -DHELLO hello.c
    cc -o hello hello.o goodbye.o
    
在 Windows 上如下：

    C:\>scons -Q
    cl /Fogoodbye.obj /c goodbye.c -DGOODBYE
    cl /Fohello.obj /c hello.c -DHELLO
    link /nologo /OUT:hello.exe hello.obj goodbye.obj
    embedManifestExeCheck(target, source, env)
    
在本指南余下例子里，我们都会使用编译方法返回的节点列表。

## 5.2 显示创建文件和目录列表

值得一提的是，在 SCons 中，表示文件和目录的节点之间有着显著的区别。SCons 分别支持用 File 和 Dir 方法返回一个文件和目录节点：

    hello_c = File('hello.c')
    Program(hello_c)

    classes = Dir('classes')
    Java(classes, 'src')

通常你不需要直接调用 File 和 Dir。那是因为在调用编译方法时会自动将文件和目录的字符串转换。而当你需要显示地查询传递到编译方法的节点属性，或是无歧义的指定目录中的一个文件的时候，你就需要 File 和 Dir 方法了。

有时候，你在不知道是文件还是目录的前提下，需要一个文件系统的入口点。这时候你就可以使用 Entry 方法，它可以返回文件或是目录的节点:

    xyzzy = Entry('xyzzy')
    
返回的 xyzzy 节点一旦传入编译方法或其他需要节点的方法中，就会转换成对应的文件和目录节点。

## 5.3 打印节点文件名

节点最常用的就是打印其对应的文件名称。请记住，编译方法返回的是节点的列表，所以必须从列表中把单个节点逐一取出。例如如下的 SConstruct 文件：

    object_list = Object('hello.c')
    program_list = Program(object_list)
    print "The object file is:", object_list[0]
    print "The program file is:", program_list[0]

就会在 POSIX 系统上打印如下的文件名：

    % scons -Q
    The object file is: hello.o
    The program file is: hello
    cc -o hello.o -c hello.c
    cc -o hello hello.o

而在 Windows 系统上是：

    C:\>scons -Q
    The object file is: hello.obj
    The program file is: hello.exe
    cl /Fohello.obj /c hello.c /nologo
    link /nologo /OUT:hello.exe hello.obj
    embedManifestExeCheck(target, source, env)

请注意，以上的例子中，object_list[0] 取出了列表的一个节点对象，Python 的 print 函数将对象转化为字符串以打印。

## 5.4 将节点文件名作为字符串

之所以上一小节能够直接打印节点名打印出来，是因为节点相应的字符串对应了文件名。假如你不只是要打印文件名，你可以用 Python 内置的 str 函数将文件名取出。比如，你可以用 Python 的 os.path.exists 来判断一个文件是否存在：

    import os.path
    program_list = Program('hello.c')
    program_name = str(program_list[0])
    if not os.path.exists(program_name):
        print program_name, "does not exist!"
 
在 POSIX 系统中执行结果如下：

    % scons -Q
    hello does not exist!
    cc -o hello.o -c hello.c
    cc -o hello hello.o     

## 5.5 GetBuildPath: 从节点或字符串中获取路径

env.GetBuildPath(file_or_list) 可以返回节点的路径，或是字符串所表示的路径。它甚至可以输入节点或字符串列表，输出对应路径的列表。如果传入单个节点，结果和调用 str(node) 相同。这些字符串可以有内嵌的构造变量，能够使用环境变量来展开。这些路径可以是文件也可以是目录，而且不需要存在。

    env=Environment(VAR="value")
    n=File("foo.c")
    print env.GetBuildPath([n, "sub/dir/$VAR"])

将会打印如下的文件名：

    % scons -Q
    ['foo.c', 'sub/dir/value']
    scons: `.' is up to date.

同时，存在着不需要 Environment 变量，就能调用的 GetBuildPath 函数。它使用了 SCons 默认的 Environment 环境，来作字符串参数的替代。
