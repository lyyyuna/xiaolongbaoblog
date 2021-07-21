title: Python 2.7 源码 - 字符串对象
date: 2017-12-28 16:20:02
categories: 语言
tags: 
- Python2.7源码
series: Python 源码之旅

---

## Python 的字符串类型和对象

有了之前[整数对象](http://www.lyyyuna.com/2017/12/24/python-internal2-integer-object/)的铺垫，研究字符串类型及其对象，当然是先看其对应的类型结构体和对象结构体。

```c
// stringobject.c
PyTypeObject PyString_Type = {
    PyVarObject_HEAD_INIT(&PyType_Type, 0)
    "str",
    PyStringObject_SIZE,
    sizeof(char),
    string_dealloc,                             /* tp_dealloc */
    (printfunc)string_print,                    /* tp_print */
    0,                                          /* tp_getattr */
    ...
    ...
    &PyBaseString_Type,                         /* tp_base */
    0,                                          /* tp_dict */
    0,                                          /* tp_descr_get */
    0,                                          /* tp_descr_set */
    0,                                          /* tp_dictoffset */
    0,                                          /* tp_init */
    0,                                          /* tp_alloc */
    string_new,                                 /* tp_new */
    PyObject_Del,                               /* tp_free */
};

// stringobject.c
typedef struct {
    PyObject_VAR_HEAD
    long ob_shash;
    int ob_sstate;
    char ob_sval[1];
} PyStringObject;
```

`ob_shash` 是该字符串的哈希值，由于 Python 的字典实现大量使用了哈希值，且字典的健多为 `PyStringObject`，预先计算哈希值并保存下来，可以加速字典的运算。

`ob_sstate` 和字符串对象的 intern 机制有关。

`ob_sval` 为什么是长度为 1 的数组？这种定义方法其实符合 C99 标准。[Flexible array member](https://en.wikipedia.org/wiki/Flexible_array_member)规定，若要支持柔性数组，可在结构体末尾放置一个不指定长度的数组。而大多数编译器都支持长度为 1 的定义方法，所以就写成 1 了。如果你单独定义 `char buf[]`，那必然是会报错的。

## 创建一个 PyStringObject

最底层的生成字符串的函数为 `PyString_FromString`。

```c
// stringobject.c
PyObject *
PyString_FromString(const char *str)
{
    ...
    size = strlen(str);
    ...

    /* Inline PyObject_NewVar */
    op = (PyStringObject *)PyObject_MALLOC(PyStringObject_SIZE + size);
    if (op == NULL)
        return PyErr_NoMemory();
    (void)PyObject_INIT_VAR(op, &PyString_Type, size);
    op->ob_shash = -1;
    op->ob_sstate = SSTATE_NOT_INTERNED;
    Py_MEMCPY(op->ob_sval, str, size+1);  // 将原始C字串的值搬运过来

    ...
    ...
    return (PyObject *) op;
}
```

函数是根据原始的 C 语言字符串生成对应的 `PyStringObject`。原始字符串被复制到 `ob_sval` 中。

## intern 机制

和整数对象一样，`PyStringObject` 需要优化才堪实用，于是 Python 的设计者便开发了 intern 机制。

所谓 intern，即如果两个字符串对象的原始字符串相同，那么其 `ob_sval` 共享同一份内存。若程序中出现了 100 次 `hello, world`，那么在内存中只会保存一份。

intern 机制的核心在于字典 `interned`。该字典为 Python 的内建数据结构，可以简单等价于 C++ 的 `map<T,R>`。该字典的健值都为字符串本身 `pystring:pystring`，所有需 intern 的字符串会缓存到该 `interned` 字典中，当在程序中再遇到相同的字符串 `pystring`，便可通过字典在 `O(1)` 时间内检索出。

```c
// stringobject.c
PyObject *
PyString_FromString(const char *str)
{
    ...
    ...
    /* share short strings */
    if (size == 0) {
        PyObject *t = (PyObject *)op;
        PyString_InternInPlace(&t);
        op = (PyStringObject *)t;
        nullstring = op;
        Py_INCREF(op);
    } else if (size == 1) {
        PyObject *t = (PyObject *)op;
        PyString_InternInPlace(&t);
        op = (PyStringObject *)t;
        characters[*str & UCHAR_MAX] = op;
        Py_INCREF(op);
    }
    return (PyObject *) op;
}
```

在 `PyString_FromString` 函数最后，会强制将长度为 0 和 1 的字符串 intern，而这一操作的核心为 `PyString_InterInPlace` 函数。

```c
// stringobject.c
void
PyString_InternInPlace(PyObject **p)
{
    ...
    ...
    if (interned == NULL) {
        interned = PyDict_New();
        if (interned == NULL) {
            PyErr_Clear(); /* Don't leave an exception */
            return;
        }
    }
    t = PyDict_GetItem(interned, (PyObject *)s);
    if (t) {
        Py_INCREF(t);
        Py_SETREF(*p, t);
        return;
    }

    if (PyDict_SetItem(interned, (PyObject *)s, (PyObject *)s) < 0) {
        PyErr_Clear();
        return;
    }

    Py_REFCNT(s) -= 2;
    PyString_CHECK_INTERNED(s) = SSTATE_INTERNED_MORTAL;
}

// object.h
#define Py_SETREF(op, op2)                      \
    do {                                        \
        PyObject *_py_tmp = (PyObject *)(op);   \
        (op) = (op2);                           \
        Py_DECREF(_py_tmp);                     \
    } while (0)
```

函数的开始会尝试新建 `interned` 字典。然后是尝试在 `interned` 字典中找该字符串 `PyDict_GetItem`。

* 若找到，就需要增加该健值对上的引用计数，并减去 `PyStringObject` 对象的引用计数。`PyStringObject` 对象减为 0 后会被回收内存。为啥原对象要被回收？因为后续程序只会通过 `interned` 字典引用字符串，原对象留着没啥用处了。
* 若没找到，会尝试在字典中新建健值对 `PyDict_SetItem`。新建的健值对需要减去 2 个引用计数。我们的 `interned` 字典健值都是原字符串，该 `PyStringObject` 无论如何都至少会有两个引用。健值仅仅是作为 Python 虚拟机内部使用，不应影响所运行程序的内存回收，故需减 2。

## 单字符字符串的进一步优化

在 `PyString_FromString` 函数中，还看到了 `characters`：

```c
    else if (size == 1) {
        PyObject *t = (PyObject *)op;
        PyString_InternInPlace(&t);
        op = (PyStringObject *)t;
        characters[*str & UCHAR_MAX] = op;
        Py_INCREF(op);
    }
```

单字节的字符串被缓存到了 `characters` 数组中。在创建字符串函数时，直接从数组中取出单字节字符串：

```c
PyObject *
PyString_FromString(const char *str)
{
    register size_t size;

    ...
    ...
    if (size == 1 && (op = characters[*str & UCHAR_MAX]) != NULL) {
#ifdef COUNT_ALLOCS
        one_strings++;
#endif
        Py_INCREF(op);
        return (PyObject *)op;
    }
    ...
    ...
```

数组比哈希字典效率更高。

## 字符串拼接所做的优化

字符串虽然是变长对象，但并不是可变对象，创建之后，`ob_sval` 数组的长度无法再改变。在拼接两个字符串 s1, s2 时，必须重新生成一个 `PyStringObject` 对象来放置 `s1->ob_sval + s2->sval`。如果要连接 N 个 `PyStringObject` 对象，那么就必须进行 N-1 次的内存申请及内存搬运的工作。毫无疑问，这将严重影响 Python 的执行效率。

所以官方推荐的做法是使用 `join` 函数，该函数一次性分配好所有内存，然后统一搬运。

```python
s = "-"
seq = ("a", "b", "c")
print s.join( seq )
```

## 实验

何种字符串会 intern？不同的 Python 版本似乎采取了不同的策略，以我 Mac 上 Python 2.7.10 为例：

```python
>>> 'foo' is 'foo'
True
>>> 'foo!' is 'foo!'
True
>>> 'a'*20 is 'a'*20
True
>>> 'a'*21 is 'a'*21
False
```

