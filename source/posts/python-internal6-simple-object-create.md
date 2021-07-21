title: Python 2.7 源码 - 简单对象创建的字节码分析
date: 2018-01-13 16:20:02
categories: 语言
tags: 
- Python2.7源码
series: Python 源码之旅

---

## 前言

Python 源码编译后，有常量表，符号表。一个作用域运行时会对应一个运行时栈。

大部分字节码就是基于常量表、符号表和运行时栈，运算后得到所需结果。

本篇就来分析简单对象创建的字节码。以下面这段代码为分析样本：

```python
i = 1
s = 'python'
d = {}
l = []
```

对生成的 pyc 文件解析，可得如下的结构，其中包括字节码反编译的结果：

```xml

magic 03f30d0a
moddate 836a595a (Sat Jan 13 10:10:11 2018)
<code>
   <argcount> 0 </argcount>
   <nlocals> 0</nlocals>
   <stacksize> 1</stacksize>
   <flags> 0040</flags>
   <codeobject> 6400005a00006401005a01006900005a02006700005a030064020053</codeobject>
   <dis>
  1           0 LOAD_CONST               0 (1)
              3 STORE_NAME               0 (i)

  2           6 LOAD_CONST               1 ('python')
              9 STORE_NAME               1 (s)

  3          12 BUILD_MAP                0
             15 STORE_NAME               2 (d)

  4          18 BUILD_LIST               0
             21 STORE_NAME               3 (l)
             24 LOAD_CONST               2 (None)
             27 RETURN_VALUE
   </dis>
   <names> ('i', 's', 'd', 'l')</names>
   <varnames> ()</varnames>
   <freevars> ()</freevars>
   <cellvars> ()</cellvars>
   <filename> '.\\test.py'</filename>
   <name> '<module>'</name>
   <firstlineno> 1</firstlineno>
   <consts>
      1
      'python'
      None
   </consts>
   <lnotab> 060106010601</lnotab>
</code>
```

我们清楚的看到 `consts` 常量表，`names` 符号表，这些表中的元素都是有明确顺序的。

## 整数赋值

第一条语句 `i = 1`。对应的字节码为：

```
0 LOAD_CONST               0 (1)
3 STORE_NAME               0 (i)
```

`LOAD_CONST` 对应的 C 语言源码为：

```c
TARGET(LOAD_CONST)
{
    x = GETITEM(consts, oparg); // 从常量表 oparg 位置处取出对象
    Py_INCREF(x);
    PUSH(x); // 压入堆栈
    FAST_DISPATCH();
}
```

该字节码带参，这里参数为 0。表示从常量表第 0 个位置取出整数，并将该数压入运行时栈：

```
+-------+----------+
| stack | f_locals |
+-------+----------+
| 1     |          |
|       |          |
|       |          |
+-------+----------+
```

左侧为运行时栈，右侧为当前作用域内的局部变量。

`STORE_NAME` 所对应的 C 语言源码为：

```c
TARGET(STORE_NAME)
{
    w = GETITEM(names, oparg); // 从符号表 oparg 位置处取出符号名
    v = POP(); // 弹出运行时栈的栈顶元素
    if ((x = f->f_locals) != NULL) {
        if (PyDict_CheckExact(x))
            err = PyDict_SetItem(x, w, v); // 将符号名作为键，栈顶元素作为值，放入字典中
        else
            err = PyObject_SetItem(x, w, v);
        Py_DECREF(v);
        if (err == 0) DISPATCH();
        break;
    }
    t = PyObject_Repr(w);
    if (t == NULL)
        break;
    PyErr_Format(PyExc_SystemError,
                    "no locals found when storing %s",
                    PyString_AS_STRING(t));
    Py_DECREF(t);
    break;
}
```

该字节码带参，参数为 0。表示从符号表第 0 个位置处取出符号名，即 `i`。然后弹出运行时栈的栈顶元素，并将符号名作为键，栈顶元素作为值，放入字典中 `f_locals`：

```
+-------+------------+
| stack | f_locals   |
+-------+------------+
|       | i, <int 1> |
|       |            |
|       |            |
+-------+------------+
```

## 字符串赋值

语句 `s = 'python'` 所对应的字节码为：

```
6 LOAD_CONST               1 ('python')
9 STORE_NAME               1 (s)
```

和整数赋值的字节码完全相同，只是参数不同。这里不再做重复分析，赋值后，运行时栈变为：

```
+-------+-------------------+
| stack | f_locals          |
+-------+-------------------+
|       | i, <int 1>        |
|       | s, <str 'python'> |
|       |                   |
+-------+-------------------+
```

## 字典赋值

语句 `d = {}` 对应的字节码为：

```
12 BUILD_MAP                0
15 STORE_NAME               2 (d)
```

`BUILD_MAP` 所对应的 C 语言源码为：

```c
// ceval.c
TARGET(BUILD_MAP)
{
    x = _PyDict_NewPresized((Py_ssize_t)oparg);
    PUSH(x);
    if (x != NULL) DISPATCH();
    break;
}

// dictobject.c
PyObject *
_PyDict_NewPresized(Py_ssize_t minused)
{
    PyObject *op = PyDict_New();

    if (minused>5 && op != NULL && dictresize((PyDictObject *)op, minused) == -1) {
        Py_DECREF(op);
        return NULL;
    }
    return op;
}
```

该字节码带参，参数为 0。而深入 `_PyDict_NewPresized` 可以看到，若参数小于 5，实际上创建的是默认大小的字典。创建完毕后，会将该字典对象压入运行时栈。

```
+--------+-------------------+
| stack  | f_locals          |
+--------+-------------------+
| <dict> | i, <int 1>        |
|        | s, <str 'python'> |
|        |                   |
+--------+-------------------+
```

最后 `STORE_NAME` 将该对象与符号 `d` 绑定：

```
+-------+-------------------+
| stack | f_locals          |
+-------+-------------------+
|       | i, <int 1>        |
|       | s, <str 'python'> |
|       | d, <dict>         |
+-------+-------------------+
```

## 列表赋值

语句 `l = []` 对应的字节码为：

```
18 BUILD_LIST               0
21 STORE_NAME               3 (l)
```

`BUILD_LIST` 对应的 C 语言源码为：

```c
TARGET(BUILD_LIST)
{
    x =  PyList_New(oparg); // 创建空列表
    if (x != NULL) {
        for (; --oparg >= 0;) {
            w = POP(); // 从栈中弹出元素
            PyList_SET_ITEM(x, oparg, w); // 将弹出的元素放入列表中
        }
        PUSH(x); // 将列表对象放入栈中
        DISPATCH();
    }
    break;
}
```

该字节码首先创建一个列表，列表依据参数值预先分配空间。这里不对列表做深入分析，只指出，这里的空间大小不是存放元素所占用的空间，而是 `PyObject *` 指针。

列表建完后，便会不停从运行时栈中弹出元素，然后将元素放入列表中。这里是空列表，所以 `BUILD_LIST` 运行时，栈为空，该字节码的参数也为 0。

我们换一个非空列表来看一下：

```python
l = [1, 2, 3]
```

编译后

```
  1           0 LOAD_CONST               0 (1)
              3 LOAD_CONST               1 (2)
              6 LOAD_CONST               2 (3)
              9 BUILD_LIST               3
             12 STORE_NAME               0 (l)
             15 LOAD_CONST               3 (None)
             18 RETURN_VALUE
```

可以看到，在 `BUILD_LIST` 之前会将三个对象压入运行时栈中。

回到本文最初的 Python 程序，4 条语句运行完后， `f_locals` 为：

```
+-------+-------------------+
| stack | f_locals          |
+-------+-------------------+
|       | i, <int 1>        |
|       | s, <str 'python'> |
|       | d, <dict>         |
|       | l, <list>         |
+-------+-------------------+
```

## 结束

在最后，我们还看到两行字节码：

```
24 LOAD_CONST               2 (None)
27 RETURN_VALUE
```

它们好像与我们的四条赋值语句没有任何关系。原来，Python 在执行了一段 CodeBlock 后，一定要返回一些值，既然如此，那就随便返回一个 `None` 好了。