title: Python 2.7 源码 - 开始
date: 2017-12-19 16:20:02
categories: 语言
tags: 
- Python2.7源码
---

## 源码编译

Python 官网可以下载到[源码](https://www.python.org/downloads/source/)。

Linux 上编译需要先安装额外模块，例如 Ubuntu

```
sudo apt-get build-dep python
sudo apt-get install libreadline-dev libsqlite3-dev libbz2-dev libssl-dev libreadline6-dev libsqlite3-dev liblzma-dev libbz2-dev tk8.5-dev blt-dev libgdbm-dev libssl-dev libncurses5-dev zlib1g-dev libncurses5-dev
```

Windows 上只需打开 PC/VS9.0/pcbuild.sln，选择最少的模块 python, pythoncore 编译，如果这两个模块编译时报错，只需按照错误提示，钩上所需模块即可。

Mac 上只需要

```
./configure
make
```

本系列文章在 Mac 上完成，但环境本身只影响编译，并不影响分析过程，读者选择方便阅读源码的环境即可。

## 第一个实验

Mac 上编译完会生成一个 `python.exe` 的可执行文件。

让我们尝试在 Python 源码中修改整数输出的部分，在每一个 int 打印时，输出 `hello, world`。

修改 /Objects/intobject.c，添加 `fprintf(fp, "hello, world\n");`

```c
static int
int_print(PyIntObject *v, FILE *fp, int flags)
     /* flags -- not used but required by interface */
{
    long int_val = v->ob_ival;
    Py_BEGIN_ALLOW_THREADS
    // 
    fprintf(fp, "hello, world\n");
    fprintf(fp, "%ld", int_val);
    Py_END_ALLOW_THREADS
    return 0;
}
```

以下是实验结果：

```
>>> a=1
>>> a
hello, world
1
>>> print(1)
hello, world
1
>>> print('eee')
eee
```

改动只影响了整数打印的部分。