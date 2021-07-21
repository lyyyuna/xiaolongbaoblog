title: SCons 用户指南第一章-编译和安装 SCons
date: 2015-10-17 15:35:56
categories: 杂
tags: 
- scons
---

本章会告诉你在你系统上安装 SCons 的基本步骤，或者是在没有预编译包的情况下自己编译 SCons (有些人会倾向于自己编译的灵活性)。在此之前，本章也会涉及到一些安装 Python 的步骤，因为这是 SCons 的依赖。幸运的是，安装 SCons 和 Python 非常简单，而且有些系统已经自带 Python 环境。

## 1.1 安装 Python

由于 SCons 是由 Python 写成的，所以必须先安装 Python。在安装 Python 之前，首先得确定你是否已经安装过 Python 了。你只需要在你系统令行输入 python -v 或者是 python --version 。

    $ python -V
    Python 2.5.1

在 Windows 系统上，输出结果类似：

    C:\>python -V
    Python 2.5.1

如果系统没有安装 Python，你就会看到类似的错误信息，比如 "command no found" (UNIX/Linux) 或者是 "'python' is not recognized as an internal or external command, operable program or batch file" (Windows)。在这种情况下，你就需要先安装 Python。

官方的下载和安装 Python 的地址如下：
[Python 官网](http://www.python.org/download/)

SCons 支持 2.7 以来的所有 2.x 版本的 Python，3.x 版本的还不支持。我们建议你安装最新的 2.x 版本 Python。最新的 Python 可以显著地提高 SCons 的性能。

## 1.2 从预编译包中安装 SCons

在很多系统上，SCons 都有现成的预编译包可直接安装，包括 Windows 和 Linux。本小节你不需要完全阅读，你只需要阅读你对应的系统。

### 1.2.1 在 Red Hat (和基于RPM) 的 Linux 系统上安装 SCons

SCons 有 RPM (Red Hat Package Manager) 格式的预编译报，可以安装于 Red Hat Linux, Fedora 或者其他使用 RPM 的 Linux 发行版。你的发行版可能已经有一个预编译 SCons RPM，比如 SUSE, Mandrake 和 Fedora。你可以在你发行版下载服务器上搜索 SCons RPM，或者是一些 RPM 搜索站：http://www.rpmfind.net/, http://rpm.pbone.net/。

如果你的发行版支持 yum 安装，你可以直接运行以下命令安装 SCons：

    # yum install scons

如果你的Linux发行版没有包含一个特定的 SCons RPM 文件，你可以下载SCons项目提供的通用的RPM来安装。这会安装SCons脚本到 /usr/bin 目录，安装 SCons 库模块到 /usr/lib/scons。

下载合适的.rpm文件，从命令行安装：

    #rpm -Uvh scons-2.1.0-1.noarch.rpm

或者，你可使用图形包管理器，查询你的包管理器应用文档，找到如何安装下载的 RPM 的特殊指令。

### 1.2.2 在 Debian Linux 系统上安装 SCons

Debian Linux 使用另一个包管理器，而且安装 SCons 也非常方便。

如果你的系统联网，则可以运行以下命令来获取最新的 Debian 包：

    # apt-get install scons

### 1.2.3 在 Windows 系统里安装 SCons

SCons 的 Windows 安装包使得安装极其简单。只需要在 [下载页面](http://www.scons.org/download.php) 下载 scons-2.4.0.win32.exe 文件，然后你需要做的就是打开后不停的下一步。

## 1.3 在任何系统上编译和安装 SCons

如果你的系统上没有预编译的报，那么仍然能够通过 Python 原生的 distutils 包轻易地编译和安装 SCons。

首先第一步是去[下载页面](http://www.scons.org/download.html)下载 scons-2.4.0.tar.gz 或者 scons-2.4.0.zip。

然后解压压缩包，在 Linux/UNIX 上使用 tar，在 Windows 上使用 WinZip。解压之后会在你的本地目录创建一个 scons-2.4.0 临时目录。然后切换你的工作目录到临时目录中执行下列命令：

    # cd scons-2.4.0
    # python setup.py install

这会编译 SCons，然后将 scons 脚本安装在执行 setup.py 脚本的目录中 (/usr/local/bin 或者 c:\Python25\Scripts)，然后将 SCons 的编译构建引擎放置在 python 的库目录中 (/usr/local/lib/scons 或者 C:\Python25\scons)。因为这些都是系统目录，所以你可能需要 root (Linux/UNIX) 或者 Administrator (Windows) 权限。

### 1.3.1 编译并同时安装多个版本的 SCons

SCons 的 setup.py 脚本有一些扩展选项，支持在多个地方安装多个版本的 SCons。举例来说，当你在决定使用哪个版本的 SCons 时，这能让你轻易的下载并实验不同版本的 SCons。

在安装时可以通过  --version-lib 选项来制定安装版本的位置：

    # python setup.py install --version-lib

这会将 SCons 的引擎安装在 /usr/lib/scons-2.4.0 或者是 C:\Python25\scons-2.4.0 目录下。

如果你第一次安装 SCons 时指定了 --version-lib 选项，你之后每一次安装新版本时就无需指定。 SCons 的 setup.py 脚本会自动检测版本的特殊位置名，并且假设你每个版本都安装在不同的位置。当然，你也可以指定 --standalone-lib 选项来消除这个假设。

### 1.3.2 安装 SCons 于其他位置

你可以指定 --prefix= 选项将 SCons 安装在非默认位置，例如：

    # python setup.py install --prefix=/opt/scons

这会将 scons 脚本安装在 /opt/scons/bin 和将编译构建引擎安装在 /opt/scons/lib/scons 中。

现在你也可以同时指定 --prefix= 和 --version-lib 选项。setup.py 脚本会根据特定 prefix 将引擎安装在特定版本的目录。这样会将编译引擎安装在 /opt/scons/lib/scons-2.4.0。

### 1.3.3 在无管理员权限的情况下编译和安装 SCons

如果你没有权限将 SCons 安装到系统目录中，那你可以使用 --prefix= 选项安装到你指定的目录中。例如，你可以将 SCons 安装到相对于用户的 $HOME 中，将 scons 脚本安装在 $HOME/bin，并将引擎安装在 $HOME/lib/scons:

    $ python setup.py install --prefix=$HOME

你当然可以安装你选择的任何地方，并使用 --version-lib 来指定特定版本的目录。

这可以在你已有 SCons 的情况下实验最新的 SCons 版本。当然，你必须在 PATH 环境变量中将最新版本的 SCons 目录放置在旧版本 SCons 目录之前。