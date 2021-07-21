title: Python 2.7 源码 - 整数对象
date: 2017-12-24 16:20:02
categories: 语言
tags: 
- Python2.7源码
series: Python 源码之旅

---

## Python 的面向对象

面向对象编程中，对象是数据以及基于这些数据的操作的集合，实际上在计算机中这只是一堆内存逻辑上的集合，无论这段内存是连续的还是分开的。

Python 是由 C 语言写成，描述一段逻辑上结合的内存，直接用结构体 `struct` 就可以了。但是 `struct` 并不是面向对象中类型的概念，对象还需要成员函数。所以还需要另外一个结构体 `struct` 来描述成员函数的集合。

上述特点就导致了在 Python 中，实际的类型也是一个对象，这个类型对象的结构体如下：

```c
typedef struct _typeobject {
    PyObject_VAR_HEAD
    const char *tp_name; /* For printing, in format "<module>.<name>" */
    Py_ssize_t tp_basicsize, tp_itemsize; /* For allocation */

    /* Methods to implement standard operations */

    destructor tp_dealloc;
    printfunc tp_print;
    getattrfunc tp_getattr;
    setattrfunc tp_setattr;
    cmpfunc tp_compare;
    reprfunc tp_repr;

    /* Method suites for standard classes */

    PyNumberMethods *tp_as_number;
    PySequenceMethods *tp_as_sequence;
    PyMappingMethods *tp_as_mapping;

    /* More standard operations (here for binary compatibility) */

    hashfunc tp_hash;
    ternaryfunc tp_call;
    reprfunc tp_str;
    getattrofunc tp_getattro;
    setattrofunc tp_setattro;

    /* Functions to access object as input/output buffer */
    PyBufferProcs *tp_as_buffer;

    /* Flags to define presence of optional/expanded features */
    long tp_flags;

    const char *tp_doc; /* Documentation string */

    /* Assigned meaning in release 2.0 */
    /* call function for all accessible objects */
    traverseproc tp_traverse;

    /* delete references to contained objects */
    inquiry tp_clear;

    /* Assigned meaning in release 2.1 */
    /* rich comparisons */
    richcmpfunc tp_richcompare;

    /* weak reference enabler */
    Py_ssize_t tp_weaklistoffset;

    /* Added in release 2.2 */
    /* Iterators */
    getiterfunc tp_iter;
    iternextfunc tp_iternext;

    /* Attribute descriptor and subclassing stuff */
    struct PyMethodDef *tp_methods;
    struct PyMemberDef *tp_members;
    struct PyGetSetDef *tp_getset;
    struct _typeobject *tp_base;
    PyObject *tp_dict;
    descrgetfunc tp_descr_get;
    descrsetfunc tp_descr_set;
    Py_ssize_t tp_dictoffset;
    initproc tp_init;
    allocfunc tp_alloc;
    newfunc tp_new;
    freefunc tp_free; /* Low-level free-memory routine */
    inquiry tp_is_gc; /* For PyObject_IS_GC */
    PyObject *tp_bases;
    PyObject *tp_mro; /* method resolution order */
    PyObject *tp_cache;
    PyObject *tp_subclasses;
    PyObject *tp_weaklist;
    destructor tp_del;

    /* Type attribute cache version tag. Added in version 2.6 */
    unsigned int tp_version_tag;

#ifdef COUNT_ALLOCS
    /* these must be last and never explicitly initialized */
    Py_ssize_t tp_allocs;
    Py_ssize_t tp_frees;
    Py_ssize_t tp_maxalloc;
    struct _typeobject *tp_prev;
    struct _typeobject *tp_next;
#endif
} PyTypeObject;
```

所以在源码中，Python 最基础的对象表示如下：

```c
typedef struct _object 
{
    Py_ssize_t ob_refcnt;
    struct _typeobject *ob_type;
} PyObject;
```

每一个对象都有一个指针指向自己所属的类型对象，而类型对象则有关于这个对象支持的所有操作的信息。

仔细看 `PyTypeObject` 的头部，`PyObject_VAR_HEAD` 即含有 `ob_type`，难道还有类型的类型这个概念？是的，这个终极的类型就是元类，即 `metaclass`。做个简单的实验。

```python
>>> class A(object):
...     pass
...
>>>
>>>
>>> A.__class__
<type 'type'>
>>> A.__class__.__class__
<type 'type'>
>>> type.__class__
<type 'type'>
```

在 Python 中类型是对象，所以类型对象也有类型，而元类的类型就是自己。

## Python 的整数类型

整数类型没啥可说的，按照 `PyTypeObject` 结构去填充信息即可：

```c
PyTypeObject PyInt_Type = {
    PyVarObject_HEAD_INIT(&PyType_Type, 0)
    "int",
    sizeof(PyIntObject),
    0,
    (destructor)int_dealloc,                    /* tp_dealloc */
    (printfunc)int_print,                       /* tp_print */
    0,                                          /* tp_getattr */
    0,                                          /* tp_setattr */
    (cmpfunc)int_compare,                       /* tp_compare */
    (reprfunc)int_to_decimal_string,            /* tp_repr */
    &int_as_number,                             /* tp_as_number */
    0,                                          /* tp_as_sequence */
    0,                                          /* tp_as_mapping */
    (hashfunc)int_hash,                         /* tp_hash */
    0,                                          /* tp_call */
    (reprfunc)int_to_decimal_string,            /* tp_str */
    PyObject_GenericGetAttr,                    /* tp_getattro */
    0,                                          /* tp_setattro */
    0,                                          /* tp_as_buffer */
    Py_TPFLAGS_DEFAULT | Py_TPFLAGS_CHECKTYPES |
        Py_TPFLAGS_BASETYPE | Py_TPFLAGS_INT_SUBCLASS,          /* tp_flags */
    int_doc,                                    /* tp_doc */
    0,                                          /* tp_traverse */
    0,                                          /* tp_clear */
    0,                                          /* tp_richcompare */
    0,                                          /* tp_weaklistoffset */
    0,                                          /* tp_iter */
    0,                                          /* tp_iternext */
    int_methods,                                /* tp_methods */
    0,                                          /* tp_members */
    int_getset,                                 /* tp_getset */
    0,                                          /* tp_base */
    0,                                          /* tp_dict */
    0,                                          /* tp_descr_get */
    0,                                          /* tp_descr_set */
    0,                                          /* tp_dictoffset */
    0,                                          /* tp_init */
    0,                                          /* tp_alloc */
    int_new,                                    /* tp_new */
};
```

## 整数对象的内存

再看一眼 PyObject 对象，

```c
typedef struct _object 
{
    Py_ssize_t ob_refcnt;
    struct _typeobject *ob_type;
} PyObject;
```

我们看到了 `ob_refcnt` 引用计数对象，可以很容易地联想到 Python 虚拟机是以引用计数为基础构建垃圾回收机制。既然如此，那还有没有必要专门讨论整数对象的内存使用,而直接抽象成引用计数归零后释放内存？

事实上，为了提高虚拟机的性能，整数对象使用了多种技术。

### 大整数创建

在 intobject.c 中定义有

```c
#define N_INTOBJECTS    ((BLOCK_SIZE - BHEAD_SIZE) / sizeof(PyIntObject))

struct _intblock {
    struct _intblock *next;
    PyIntObject objects[N_INTOBJECTS];
};

typedef struct _intblock PyIntBlock;

static PyIntBlock *block_list = NULL;
static PyIntObject *free_list = NULL;
```

`block_list` 是由一个个 `PyIntBlock` 串起来的链表，每一个 `PyIntBlock` 是一个整数数组。`free_list` 是由空闲的 `PyIntObject` 组成的链表，空闲是指这块内存虽然被划分为一个 `PyIntObject`，但并没有被用于表示一个真正的整数，即其所存储的信息是无用的。

整数创建时，`PyObject * PyInt_FromLong(long ival)` 会被调用，

```c
PyObject *
PyInt_FromLong(long ival)
{
    register PyIntObject *v;
    ...
    ...
    if (free_list == NULL) {
        if ((free_list = fill_free_list()) == NULL)
            return NULL;
    }
    /* Inline PyObject_New */
    v = free_list;
    free_list = (PyIntObject *)Py_TYPE(v);
    (void)PyObject_INIT(v, &PyInt_Type);
    v->ob_ival = ival;
    return (PyObject *) v;
}

static PyIntObject *
fill_free_list(void)
{
    PyIntObject *p, *q;
    /* Python's object allocator isn't appropriate for large blocks. */
    p = (PyIntObject *) PyMem_MALLOC(sizeof(PyIntBlock));
    if (p == NULL)
        return (PyIntObject *) PyErr_NoMemory();
    ((PyIntBlock *)p)->next = block_list;
    block_list = (PyIntBlock *)p;
    /* Link the int objects together, from rear to front, then return
       the address of the last int object in the block. */
    p = &((PyIntBlock *)p)->objects[0];
    q = p + N_INTOBJECTS;
    while (--q > p)
        Py_TYPE(q) = (struct _typeobject *)(q-1);
    Py_TYPE(q) = NULL;
    return p + N_INTOBJECTS - 1;
}
```

当创建整数的时候，会先尝试从 `free_list` 中取，如果没有空闲的，就会尝试 `fill_free_list`。这个新的 `PyIntBlock` 中，每一个 `PyIntObject` 都借用 `ob_type` 来连接成链表。

```c
#define Py_TYPE(ob) (((PyObject*)(ob))->ob_type)
```

这里只是借用，初看源码的朋友不要被这里搞混了。因为此时这块内存并没有存放整数，它的成员自然可以借来他用。`free_list` 指向数组的末尾，从后往前链接到数组首部。

### 大整数销毁

当一个整数销毁时，便会进入 `int_dealloc` 函数内。

```c
static void
int_dealloc(PyIntObject *v)
{
    if (PyInt_CheckExact(v)) {
        Py_TYPE(v) = (struct _typeobject *)free_list;
        free_list = v;
    }
    else
        Py_TYPE(v)->tp_free((PyObject *)v);
}
```

这个函数在正常情况下不会走到 `else` 分支，意味着所谓的销毁，只是把这个整数的 `PyIntObject` 重新放回 `free_list` 链表中，并不会释放这块内存。这岂不会造成内存泄漏？只能说，理论上会。整数对象所占用的内存空间，只和这个程序同时拥有的最多的整数数量有关。

上述做法也是为了优化性能，虚拟机不再需要频繁的 `malloc` 和 `free`。

### 小整数

除了普通的整数外，Python 中还存在着一种小整数对象。在之前的 `PyInt_FromLong` 函数中，我们略了一部分，现在我们从另一个角度看。

```c
PyObject *
PyInt_FromLong(long ival)
{
    register PyIntObject *v;

    if (-NSMALLNEGINTS <= ival && ival < NSMALLPOSINTS) {
        v = small_ints[ival + NSMALLNEGINTS];
        Py_INCREF(v);
    ...
    ...

}
```

当想创建的整数在 `-NSMALLNEGINTS ~ NSMALLPOSINTS` 之间时，就会从 `small_ints` 数组中直接取出。这范围内的整数即为小整数。小整数使用广泛，循环的初始值，终结值，步进值等等，都是数值很小的整数。小整数从 Python 虚拟机运行之初就存在，使用它们既不需要 `malloc` 和 `free`，甚至连指针操作 `free_list` 也不要。效率比大整数更高。

而小整数的范围可在编译 Python 时指定，默认为 `-5 ~ 257`。

## 实验

改造一下整数打印函数，反映出内存的变化。

```c
static int values[10];
static int refcounts[10];
/* ARGSUSED */
static int
int_print(PyIntObject *v, FILE *fp, int flags)
     /* flags -- not used but required by interface */
{
    PyIntObject * intObjectPtr;
    PyIntBlock * p = block_list;
    PyIntBlock * last = NULL;
    int count = 0;
    int i = 0;

    long int_val = v->ob_ival;
    Py_BEGIN_ALLOW_THREADS
    fprintf(fp, "%ld", int_val);
    Py_END_ALLOW_THREADS

    while (p!=NULL)
    {
        ++count;
        last = p;
        p = p->next;
    }
    intObjectPtr = last->objects;
    intObjectPtr += N_INTOBJECTS-1;
    printf("\nvalue's address is @%p\n", v);

    for (i = 0; i < 10; i++, --intObjectPtr)
    {
        values[i] = intObjectPtr->ob_ival;
        refcounts[i] = intObjectPtr->ob_refcnt;
    }
    printf("  value : ");
    for (i = 0; i < 8; ++i)
    {
        printf("%d\t", values[i]);
    }
    printf("\n");
    printf("  refcnt : ");
    for (i = 0; i < 8; ++i)
    {
        printf("%d\t", refcounts[i]);
    }
    printf("\n");

    printf(" block_list count : %d\n", count);
    printf(" free_list address : %p\n", free_list);
    return 0;
}
```

运行重新编译后的 python 虚拟机。

### 大整数实验

首先是连续创建两个大整数。

```python
>>> a=1111
>>> a
1111
value's address is @0x100303888
  value : -5    -4      -3      -2      -1      0       1       2
  refcnt : 1    1       1       1       54      389     587     84
 block_list count : 9
 free_list address : 0x1003038a0

>>> b=2222
>>> b
2222
value's address is @0x1003038a0
  value : -5    -4      -3      -2      -1      0       1       2
  refcnt : 1    1       1       1       54      389     587     84
 block_list count : 9
 free_list address : 0x1003038b8
```

第一次的 `free_list`，正好是第二次整数的地址。可以看到小整数都至少有一个引用，有些多于一次是因为 python 虚拟机内部使用的缘故。

当尝试创建一个相同的大整数时。

```python
>>> c=2222
>>> c
2222
value's address is @0x1003038b8
  value : -5    -4      -3      -2      -1      0       1       2
  refcnt : 1    1       1       1       54      389     587     84
 block_list count : 9
 free_list address : 0x1003038d0
```

可以看出，虽然值相同，但并不是同一个内存块。

### 小整数实验

创建两个相同的小整数。

```python
>>> d=1
>>> d
1
value's address is @0x100604ce8
  value : -5    -4      -3      -2      -1      0       1       2
  refcnt : 1    1       1       1       54      389     591     84
 block_list count : 9
 free_list address : 0x1003038d0

>>> c=1
>>> c
1
value's address is @0x100604ce8
  value : -5    -4      -3      -2      -1      0       1       2
  refcnt : 1    1       1       1       54      389     592     84
 block_list count : 9
 free_list address : 0x100303828
```

可以看出，整数 1 只是增加了引用计数，内存块是同一个。