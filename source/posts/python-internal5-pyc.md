title: Python 2.7 源码 - pyc 的结构
date: 2018-01-06 16:20:02
categories: 语言
tags: 
- Python2.7源码
---

## 前言

Python 是一门解释性语言，源码执行需要经过：编译-字节码-虚拟机的步骤。本文就介绍一下 `.py` 文件编译后的 `.pyc` 文件结构。直接运行的代码不会生成 `.pyc`，而 Python 的 import 机制会触发 `.pyc` 文件的生成。

## magic number 和修改时间

我们在导入模块的源码中，能找到 `.pyc` 文件的蛛丝马迹：

```c
// Python/import.c
static void
write_compiled_module(PyCodeObject *co, char *cpathname, struct stat *srcstat, time_t mtime)
{
    ...
    PyMarshal_WriteLongToFile(pyc_magic, fp, Py_MARSHAL_VERSION);
    /* First write a 0 for mtime */
    PyMarshal_WriteLongToFile(0L, fp, Py_MARSHAL_VERSION);
    PyMarshal_WriteObjectToFile((PyObject *)co, fp, Py_MARSHAL_VERSION);
    ...
}
```

可以看到，`.pyc` 文件包含三个主要内容：magic number，修改时间，和一个 `PyCodeObject` 对象。

magic number 和 Python 版本有关，magic number 不同能够防止低版本误执行高版本编译后的字节码。修改时间能让编译器决定是否重新编译源文件。

我们来读取 `.pyc` 的前八字节来验证一下我们的分析，`func0.py` 是测试脚本：

```python
## func0.py
    f = open(fname, 'rb')
    magic = f.read(4)
    moddate = f.read(4)
    modtime = time.asctime(time.localtime(struct.unpack('I', moddate)[0]))
    print 'magic number: %s' % (magic.encode('hex'))
    print 'moddate %s (%s)' % (moddate.encode('hex'), modtime)
```

被测试源文件：

```python
# test.py
def add(a):
    a=1
    print a
```

使用以下强制编译成 `.pyc`：

```
python -m py_compile test.py
```

测试结果：

```
$ python func0.py test.pyc
magic number: 03f30d0a
moddate 5e9a4b5a (Tue Jan  2 22:42:38 2018)
```

## PyCodeObject 对象

`PyCodeObject` 是一段代码块编译的直接结果。换句话说，一个作用域，对应一个代码，最终对应一个编译后的 `PyCodeObject`。
首先看一下 `PyCodeObject` 的结构：

```c
typedef struct {
    PyObject_HEAD
    int co_argcount;		/* #arguments, except *args */
    int co_nlocals;		/* #local variables */
    int co_stacksize;		/* #entries needed for evaluation stack */
    int co_flags;		/* CO_..., see below */
    PyObject *co_code;		/* instruction opcodes */
    PyObject *co_consts;	/* list (constants used) */
    PyObject *co_names;		/* list of strings (names used) */
    PyObject *co_varnames;	/* tuple of strings (local variable names) */
    PyObject *co_freevars;	/* tuple of strings (free variable names) */
    PyObject *co_cellvars;      /* tuple of strings (cell variable names) */
    /* The rest doesn't count for hash/cmp */
    PyObject *co_filename;	/* string (where it was loaded from) */
    PyObject *co_name;		/* string (name, for reference) */
    int co_firstlineno;		/* first source line number */
    PyObject *co_lnotab;	/* string (encoding addr<->lineno mapping) See
				   Objects/lnotab_notes.txt for details. */
    void *co_zombieframe;     /* for optimization only (see frameobject.c) */
    PyObject *co_weakreflist;   /* to support weakrefs to code objects */
} PyCodeObject;
```

一段程序不可能只有一个作用域，嵌套的子作用域存在于 `co_consts` 之中。

我们写个简单的脚本，展现这种嵌套的结构：

```python
# func0.py
...
def show_code(code, indent=''):
    old_indent = indent  
    print "%s<code>" % indent  
    indent += '   '  
    show_hex("bytecode", code.co_code, indent=indent)  
    print "%s<filename> %r</filename>" % (indent, code.co_filename)  
    print "%s<consts>" % indent  
    for const in code.co_consts:  
        if type(const) == types.CodeType:  
            show_code(const, indent+'   ')  
        else:  
            print "   %s%r" % (indent, const)  
    print "%s</consts>" % indent  
    print "%s</code>" % old_indent  
```

被测试的源文件：
```python
# test.py
def add(a):
    a=1
    print a
```

测试步骤：

```xml
exs $ python func0.py test.pyc
<code>
   <bytecode> 6400008400005a000064010053</bytecode>
   <filename> 'test.py'</filename>
   <consts>
      <code>
         <bytecode> 6401007d00007c0000474864000053</bytecode>
         <filename> 'test.py'</filename>
         <consts>
            None
         </consts>
      </code>
      None
   </consts>
</code>
```

可以看到，函数 `add` 作为全局作用域的中一个子作用域，在编译结果中，是以常量形式存在于全局作用域的 `PyCodeObject` 中。

## 查看字节码

Python 提供了 `dis` 模块，其中的 `disassemble()` 函数可以反编译 `PyCodeObject` 对象，以可读的形式展现出来。
我们修改 `func0.py`，将字节码对应的指令打印出来，增加下述代码：

```python
    ...
    show_hex("bytecode", code.co_code, indent=indent)  
    print "%s<dis>" % indent  
    dis.disassemble(code)  
    print "%s</dis>" % indent  
    ...
```

测试结果：

```
$ python func0.py test.pyc
<code>
   <bytecode> 6400008400005a000064010053</bytecode>
   <dis>
  3           0 LOAD_CONST               0 (<code object add at 0x109aa1c30, file "test.py", line 3>)
              3 MAKE_FUNCTION            0
              6 STORE_NAME               0 (add)
              9 LOAD_CONST               1 (None)
             12 RETURN_VALUE
   </dis>
   <filename> 'test.py'</filename>
   <consts>
      <code>
         <bytecode> 6401007d00007c0000474864000053</bytecode>
         <dis>
  4           0 LOAD_CONST               1 (1)
              3 STORE_FAST               0 (a)

  5           6 LOAD_FAST                0 (a)
              9 PRINT_ITEM
             10 PRINT_NEWLINE
             11 LOAD_CONST               0 (None)
             14 RETURN_VALUE
         </dis>
         <filename> 'test.py'</filename>
         <consts>
            None
            1
         </consts>
      </code>
      None
   </consts>
</code>
```