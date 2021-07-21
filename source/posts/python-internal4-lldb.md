title: Python 3.x 源码 - 编译调试之旅
date: 2018-01-01 16:20:02
categories: 语言
tags: 
- Python3.x源码
---

## 建立环境

1. 从 GitHub 上下载源码

```
$ git clone https://github.com/python/cpython
$ cd cpython
```

2. 编译之前打开 `--with-pydebug` 选项

```
$ ./configure --with-pydebug
$ make
```

编译完成之后，会在当前目录看到一个二进制文件 `python`。如果你使用的是 macOS，文件名为 `python.exe`，此文后续在命令行中请用 `python.exe` 代替 `python`。

## 用 GDB 调试

我们将使用 `gdb` 来追踪 `python` 的行为。

本小结也是一个 `gdb` 入门教程。

### GDB 快捷键

* r (run)，运行程序
* b (break)，设置断点
* s (step)，单步运行
* c (continue)，继续运行程序，在断点处会停止运行
* l (list)，列出当前程序的源代码
* ctrl+x，打开 tui 模式
* ctrl+p，往上
* ctrl+n，往下

### GDB 介绍

在编译目录敲入命令 `gdb python`：

```
$ gdb python
GNU gdb (GDB) 7.12
Copyright (C) 2016 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.  Type "show copying"
and "show warranty" for details.
This GDB was configured as "x86_64-pc-linux-gnu".
Type "show configuration" for configuration details.
For bug reporting instructions, please see:
<http://www.gnu.org/software/gdb/bugs/>.
Find the GDB manual and other documentation resources online at:
<http://www.gnu.org/software/gdb/documentation/>.
For help, type "help".
Type "apropos word" to search for commands related to "word"...
Reading symbols from python...done.
(gdb) 
```

在最后一行我们看到 `Reading symbols from python...done.`，说明现在我们可以通过 `gdb` 来调试 python 了。

现在程序还没有运行，调试器在程序最前端停止下来。

每一个 C 程序都从 `main()` 函数开始运行，所以我们在 `main()` 上打一个断点：

```
(gdb) b main
Breakpoint 1 at 0x41d1d6: file ./Programs/python.c, line 20.
(gdb)
```

`gdb` 会在 `Programs/python.c, line 20` 处打上断点。从这条信息可以看出，`Python` 的入口点为 `Programs/python.c:20`。另外，如果你事先已知晓源码，可以直接：

```
(gdb) b Programs/python.c:20
Breakpoint 3 at 0x41d1d6: file ./Programs/python.c, line 20.
(gdb)
```

运行 `python`：

```
(gdb) r
Starting program: /home/grd/Python/cpython/python 
[Thread debugging using libthread_db enabled]
Using host libthread_db library "/usr/lib/libthread_db.so.1".

Breakpoint 1, main (argc=1, argv=0x7fffffffdff8) at ./Programs/python.c:20
20	{
(gdb)
```

我们停在了之前设置的断点处，使用 `list` 列出源代码：

```
(gdb) l
15	}
16	#else
17	
18	int
19	main(int argc, char **argv)
20	{
21	    wchar_t **argv_copy;
22	    /* We need a second copy, as Python might modify the first one. */
23	    wchar_t **argv_copy2;
24	    int i, res;
(gdb) 
```

或者使用 `ctrl+x` 调出 `tui`：

```
   ┌──./Programs/python.c─────────────────────────────────────────────────────────┐
   │15      }                                                                     │
   │16      #else                                                                 │
   │17                                                                            │
   │18      int                                                                   │
   │19      main(int argc, char **argv)                                           │
B+>│20      {                                                                     │
   │21          wchar_t **argv_copy;                                              │
   │22          /* We need a second copy, as Python might modify the first one. */│
   │23          wchar_t **argv_copy2;                                             │
   │24          int i, res;                                                       │
   │25          char *oldloc;                                                     │
   │26                                                                            │
   └──────────────────────────────────────────────────────────────────────────────┘
multi-thre Thread 0x7ffff7f9de In: main                         L20   PC: 0x41d1d6 
(gdb)
```

在 `tui` 模式，也可以看到我们停在了源代码的第 20 行。

继续运行 `continue`，将进入 `python` 的交互式解释器环境。

```
(gdb) c
   │29                                                                                                                                                                                                                                      │
   │30          argv_copy = (wchar_t **)PyMem_RawMalloc(sizeof(wchar_t*) * (argc+1));                                                                                                                                                       │
   │31          argv_copy2 = (wchar_t **)PyMem_RawMalloc(sizeof(wchar_t*) * (argc+1));                                                                                                                                                      │
   │32          if (!argv_copy || !argv_copy2) {                                                                                                                                                                                            │
   └────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
multi-thre Thread 0x7ffff7f9de In: main                                                                                                                                                                                   L20   PC: 0x41d1d6 
[Thread debugging using libthread_db enabled]
Using host libthread_db library "/usr/lib/libthread_db.so.1".

Breakpoint 1, main (argc=1, argv=0x7fffffffdff8) at ./Programs/python.c:20
(gdb) c
Continuing.
Python 3.7.0a0 (default, Feb 22 2017, 22:10:22) 
[GCC 6.3.1 20170109] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>>
```

## 调试语法分析器

### 调试语法分析器

创建一个简单的 python 脚本 `test.py`：

```python
a = 100
```

打开 gdb，设置参数，运行 python：

```
$ gdb python
(gdb) set args test.py
// 或者
$ gdb --args python test.py
```

并在 `main` 函数上打断点：

```
(gdb) b main
Breakpoint 1 at 0x41d1d6: file ./Programs/python.c, line 20.
(gdb) r
Starting program: /home/grd/Python/cpython/python 
[Thread debugging using libthread_db enabled]
Using host libthread_db library "/usr/lib/libthread_db.so.1".

Breakpoint 1, main (argc=1, argv=0x7fffffffdff8) at ./Programs/python.c:20
20	{Starting program: /home/grd/Python/cpython/python 
[Thread debugging using libthread_db enabled]
Using host libthread_db library "/usr/lib/libthread_db.so.1".

Breakpoint 1, main (argc=1, argv=0x7fffffffdff8) at ./Programs/python.c:20
20	{
(gdb) 
```

#### Program/python.c

如果你仔细查看 `Program/python.c` 的源码，会发现 `main()` 做了很多事情，核心是调用 `Py_Main(argc, argv_copy)` 函数：

```c
    argv_copy2[argc] = argv_copy[argc] = NULL;

    setlocale(LC_ALL, oldloc);
    PyMem_RawFree(oldloc);

    res = Py_Main(argc, argv_copy);

    /* Force again malloc() allocator to release memory blocks allocated
       before Py_Main() */
    (void)_PyMem_SetupAllocators("malloc");
```

#### Modules/main.c

简单来说，`Py_Main` 会做如下：

1. 初始化哈希随机：391
2. 重置 Warning 选项：394，395
3. 通过 `_PyOS_GetOpt()` 解析命令行选项：397
4. 通过 `Py_Initialize()` 初始化 python：693
5. 导入 `readline` 模块：723
6. 依次执行：
    1. `run_command()`
    2. `RunModule()`
    3. `RunInteractiveHook()`
    4. `RunMainFromImporter()`
    5. `run_file()`
    6. `PyRun_AnyFileFlags()`

将断点打在 `run_file`，第 793 行，可以用 `p filename` 查看当前文件名：

```
(gdb) b Modules/main.c:793
(gdb) c
(gdb) p filename
$1 = 0x931050 L"test.py
(gdb) s
```

`run_file` 只是一个装饰器，该装饰器会调用 `Python/pythonrun.c` 中的 `PyRun_InteractiveLoopFlags()` 或 `PyRun_PyRun_SimpleFileExFlags`。从名字上就可以看出一个会进入交互式环境，另一个就是带参数的 python 调用。这里，我们带上了参数 `test.py`，所以会运行 `PyRun_PyRun_SimpleFileExFlags`。

#### Python/pythonrun.c

`PyRun_SimpleFileExFlags()` 首先会用 `maybe_pyc_file()` 检查所传文件是否是 `.pyc` 格式。

在我们的例子中，由于是 `.py` 文件，所以接着会调用 `PyRun_FileExFlags()`。最后调用 `PyParser_ASTFromFileObject()` 来建立抽象语法树（AST）。

抽象语法树需调用 `Parser/parsetok.c` 中的 `PyParser_ParseFileObject()` 创建节点，再用 `PyAST_FromNodeObject()` 函数从节点构建 AST 树。

#### Parser/parsetok.c

`PyParser_ParseFileObject()` 会从 `PyTokenizer_FromFile()` 中获取所有的 toekn，将这些 token 传入 `parsetok()` 创建节点。

最有趣的部分是其中包含一个无限循环：

```c
type = PyTokenizer_Get(tok, &a, &b);
```

该函数是 `tok_get()` 的装饰器，它会返回预定义于 `token.h` 中的 token 类型：

```c
// include/token.h
#define ENDMARKER	0
#define NAME		1
#define NUMBER		2
#define STRING		3
#define NEWLINE		4
#define INDENT		5
#define DEDENT		6
#define LPAR		7
#define RPAR		8
#define LSQB		9
...
#define RARROW          51
#define ELLIPSIS        52
/* Don't forget to update the table _PyParser_TokenNames in tokenizer.c! */
#define OP		53
#define AWAIT		54
#define ASYNC		55
#define ERRORTOKEN	56
#define N_TOKENS	57
```

在 for 循环的第一轮，我们在 gdb 中打印上述代码中的 `type`：

```
(gdb) p type
$1 = 1
```

根据头文件中的宏，值 1 对应 `NAME`，说明该 token 为一个变量名。

在第 236 行，`str[lbbben] = '\0'` 会储存 token 所对应的字符串，即 `a`：

```
(gdb) p str
$2 = 0x7ffff6eb5a18 "a"
```

看起来很有道理，因为我们的源码为 `a = 100`，第一个 token 字符串应对应于 `a`，类型为 `NAME`。

解析器接下来会调用 `PyParser_AddToken()`，这会将 token 加入语法树中。

### 语法生成

语法的文本表示在 `Grammar/Grammar` 中，这是用 `yacc` 写的，我建议直接忽略。而语法的数字表示在 `Python/graminit.c` 中，其中包含了 DFA 数组。

修改 `test.py` 的内容为：

```python
class foo:
    pass
```

打开 gdb，在 `PyParser_AddToken()` 上打断点：

```
$ gdb python
(gdb) b PyParser_AddToken
(gdb) r test.py
[Thread debugging using libthread_db enabled]
Using host libthread_db library "/usr/lib/libthread_db.so.1".

Breakpoint 1, PyParser_AddToken (ps=ps@entry=0x9cf4f0, type=type@entry=1, str=str@entry=0x7ffff6eb5a18 "class", lineno=1, col_offset=0, expected_ret=expected_ret@entry=0x7fffffffdc34) at Parser/parser.c:229
229	{
```

为了看到 DFA 的状态变化，在 `dfa *d = ps->p_stack.s_top->s_dfa` 下一行打上断点：

```
(gdb) b 244
Breakpoint 2 at 0x5b5933: file Parser/parser.c, line 244.
(gdb) c
Breakpoint 2, PyParser_AddToken (ps=ps@entry=0x9cf4f0, type=type@entry=1, str=str@entry=0x7ffff6eb5a18 "class", lineno=1, col_offset=0, expected_ret=expected_ret@entry=0x7fffffffdc34) at Parser/parser.c:244
244	        state *s = &d->d_state[ps->p_stack.s_top->s_state];
```

然后打印出 `d` 的值：

```
(gdb) p *d
$6 = {d_type = 269, d_name = 0x606340 "stmt", d_initial = 0, d_nstates = 2, d_state = 0x8a6a40 <states_13>, d_first = 0x60658c ""}
(gdb) set print pretty
(gdb) p *d
$8 = {
  d_type = 269,
  d_name = 0x606340 "stmt",
  d_initial = 0,
  d_nstates = 2,
  d_state = 0x8a6a40 <states_13>,
  d_first = 0x60658c ""
}
...
...
(gdb) p d->d_name
'file_input'
(gdb)
```

对比 `d_name` 的值，发现它出现在 `Grammar/Grammar` 中：

```
# Grammar for Python

single_input: NEWLINE | simple_stmt | compound_stmt NEWLINE
file_input: (NEWLINE | stmt)* ENDMARKER
eval_input: testlist NEWLINE* ENDMARKER
```

运行多次，我们会发现 `d_name` 按如下的顺序变化：`file_input`, `stmt`, `compound_stmt`, `classdef`, `classdef`, `classdef`, `classdef`, `suite`, `suite`, `suite`, `stmt`, `simple_stmt`, `small_stmt`, `pass_stmt`, `simple_stmt`, `suite`, `file_input`, `file_input`。

### 回顾

查看一下到目前为止的调用栈：

* `main()`
* `Py_Main()`
* `run_file()`
* `PyRun_AnyFileExFlags()`
* `PyRun_SimpleFileExFlags()`
* `PyRun_FileExFlags()`
* `PyParser_ASTFromFileObject()`
    * `PyParser_ParseFileObject()`
        * `parsetok()`
        * `PyParser_AddToken()`
    * `PyAST_FromNodeObject()`

## 调试抽象语法树（AST）生成器

我们从 `PyParser_ParseFileObject()` 构建了语法树，下一步是生成 AST。

在此之前，需要介绍一些宏。这些宏定义于 `Include/node.h` 中，用于从节点结构中查询数据：

* `CHILD(node *, int)`，返回第 n 个节点（从 0 开始）
* `RCHILD(node *, int)`，从右往左返回第 n 个节点，使用负数
* `NCH(node *)`，返回节点总数
* `STR(node *)`，返回节点的字符串表示，比如冒号 token，会返回 `:`
* `TYPE(node *)`，返回节点的类型，类型定义于 `Include/graminit.h`
* `REQ(node *, TYPE)`，判断节点类型是否是 `TYPE`
* `LINENO(node *)`，获取解析规则源码所在的行数，规则定义在 `Python/ast.c`

在 `Python/ast.c` 中的 `PyAST_FromNodeObject()` 将语法树转换为 AST。

```c
for (i = 0; i < NCH(n) - 1; i++) {
    ch = CHILD(n, i);
    if (TYPE(ch) == NEWLINE)
        continue;
    REQ(ch, stmt);
    num = num_stmts(ch);
    if (num == 1) {
        s = ast_for_stmt(&c, ch);
        if (!s)
            goto out;
        asdl_seq_SET(stmts, k++, s);
    }
    else {
        ch = CHILD(ch, 0);
        REQ(ch, simple_stmt);
        for (j = 0; j < num; j++) {
            s = ast_for_stmt(&c, CHILD(ch, j * 2));
            if (!s)
                goto out;
            asdl_seq_SET(stmts, k++, s);
        }
    }
}
res = Module(stmts, arena);
```

`ast_for_stmt()` 是 `ast_for_xx` 的装饰器，其中 `xx` 是对应函数处理的语法规则。

## 调试符号表生成器

回到 `Python/pythonrun.c` 中的 `PyRun_FileExFlags()`。它会接着将结果 `mod` 传入 `run_mod()` 函数中。它完成了重要的两步：第一，生成代码对象（`PyAST_CompileObject()`），第二，进入解析循环（`PyEval_EvalCode()`）。

`PyAST_CompileObject()` 位于 `Python/compile.c`。它有两个重要的函数：

1. `PySumtable_BuildObject()`
2. `compiler_mod()`

`Python/symtable.c` 中的 `PySumtable_BuildObject()` 用于生成符号表。

符号表的结构定义在 `Include/symtble.h` 中：

```c
struct _symtable_entry;

struct symtable {
    PyObject *st_filename;          /* name of file being compiled,
                                       decoded from the filesystem encoding */
    struct _symtable_entry *st_cur; /* current symbol table entry */
    struct _symtable_entry *st_top; /* symbol table entry for module */
    PyObject *st_blocks;            /* dict: map AST node addresses
                                     *       to symbol table entries */
    PyObject *st_stack;             /* list: stack of namespace info */
    PyObject *st_global;            /* borrowed ref to st_top->ste_symbols */
    int st_nblocks;                 /* number of blocks used. kept for
                                       consistency with the corresponding
                                       compiler structure */
    PyObject *st_private;           /* name of current class or NULL */
    PyFutureFeatures *st_future;    /* module's future features that affect
                                       the symbol table */
    int recursion_depth;            /* current recursion depth */
    int recursion_limit;            /* recursion limit */
};

typedef struct _symtable_entry {
    PyObject_HEAD
    PyObject *ste_id;        /* int: key in ste_table->st_blocks */
    PyObject *ste_symbols;   /* dict: variable names to flags */
    PyObject *ste_name;      /* string: name of current block */
    PyObject *ste_varnames;  /* list of function parameters */
    PyObject *ste_children;  /* list of child blocks */
    PyObject *ste_directives;/* locations of global and nonlocal statements */
    _Py_block_ty ste_type;   /* module, class, or function */
    int ste_nested;      /* true if block is nested */
    unsigned ste_free : 1;        /* true if block has free variables */
    unsigned ste_child_free : 1;  /* true if a child block has free vars,
                                     including free refs to globals */
    unsigned ste_generator : 1;   /* true if namespace is a generator */
    unsigned ste_coroutine : 1;   /* true if namespace is a coroutine */
    unsigned ste_varargs : 1;     /* true if block has varargs */
    unsigned ste_varkeywords : 1; /* true if block has varkeywords */
    unsigned ste_returns_value : 1;  /* true if namespace uses return with
                                        an argument */
    unsigned ste_needs_class_closure : 1; /* for class scopes, true if a
                                            closure over __class__
                                             should be created */
    int ste_lineno;          /* first line of block */
    int ste_col_offset;      /* offset of first line of block */
    int ste_opt_lineno;      /* lineno of last exec or import * */
    int ste_opt_col_offset;  /* offset of last exec or import * */
    int ste_tmpname;         /* counter for listcomp temp vars */
    struct symtable *ste_table;
} PySTEntryObject;
```

可以看出，符号表其实是一个字典结构，每一项是一个符号对应关系。

在第 281 行的 for 循环打上断点（`for (i = 0; i < asdl_seq_LEN(seq); i++)`），会来到 `symtable_visit_stmt()` 函数，该函数生成符号表的每一项。接着打断点：

```
(gdb) b symtable_visit_stmt
```

就能观察到类似 `xx_kind` 的表达式，例如 `Name_kind` 会调用 `symtable_add_def()` 将一个符号定义加入到符号表中。

## 调试编译器和字节码生成器

回到函数 `PyAST_CompileObject()` 中，下一步是 `compiler_mod()`，将抽象语法树转换为上下文无关语法。

在此处打断点（`b compiler_mod`）。swtich 分支会把我们带进 `Module_kind`，里面会调用 `compiler_body()` 函数，接着单步调试，就会发现一个 for 循环：

```c
for (; i < asdl_seq_LEN(stmts); i++)
    VISIT(c, stmt, (stmt)ty)asdl_seq_GET(stmts, i));
```

这里，我们在抽象语义描述语言（ASDL）中遍历，调用宏 `VISIT`，接着调用 `compiler_visit_expr(c, nodeZ)`。

以下宏会产生字节码：

* `ADDOP()`，增加一个指定的字节码
* `ADDOP_I()`，增加的字节码是带参数的
* `ADDOP_O(struct compiler *c, int op, PyObject * type, PyObject *obj)`，根据指定 `PyObject` 在序列中的位置，增加一个字节码，但是不考虑 name mangling。常用于全局、常量或参数的变量名寻找，因为这种变量名的作用域是未知的。
* `ADDOP_NAME()`，和 `ADDOP_O` 类似，但是会考虑 name mangling。用于属性加载和导入。
* `ADDOP_JABS()`，创建一个绝对跳转
* `ADDOP_JREL()`，创建一个相对跳转

为了验证是否生成了正确的字节码，可以在 `test.py` 上运行：

```
$ python -m dis test.py
  1           0 LOAD_CONST               0 (100)
              2 STORE_NAME               0 (a)
              4 LOAD_CONST               1 (None)
              6 RETURN_VALUE
```

## 调试解析器循环

一旦字节码生成，下一步是由解析器运行程序。回到 `Python/pythonrun.c` 文件中，我们接着会调用函数 `PyEval_EvalCode()`，这是对 `PyEval_EvalCodeEx()/_PyEval_EvalCodeWithName()` 的装饰器函数。

> 和 Python2.7 不一样，`PyEval_EvalCodeEx` 不会建立函数栈，这一步被移入 `_PyEval_EvalCodeWithName`。

栈对象的结构定义于 `Include/frameobject.h`：

```c
typedef struct _frame {
    PyObject_VAR_HEAD
    struct _frame *f_back;      /* previous frame, or NULL */
    PyCodeObject *f_code;       /* code segment */
    PyObject *f_builtins;       /* builtin symbol table (PyDictObject) */
    PyObject *f_globals;        /* global symbol table (PyDictObject) */
    PyObject *f_locals;         /* local symbol table (any mapping) */
    PyObject **f_valuestack;    /* points after the last local */
    /* Next free slot in f_valuestack.  Frame creation sets to f_valuestack.
       Frame evaluation usually NULLs it, but a frame that yields sets it
       to the current stack top. */
    PyObject **f_stacktop;
    PyObject *f_trace;          /* Trace function */

    /* In a generator, we need to be able to swap between the exception
       state inside the generator and the exception state of the calling
       frame (which shouldn't be impacted when the generator "yields"
       from an except handler).
       These three fields exist exactly for that, and are unused for
       non-generator frames. See the save_exc_state and swap_exc_state
       functions in ceval.c for details of their use. */
    PyObject *f_exc_type, *f_exc_value, *f_exc_traceback;
    /* Borrowed reference to a generator, or NULL */
    PyObject *f_gen;

    int f_lasti;                /* Last instruction if called */
    /* Call PyFrame_GetLineNumber() instead of reading this field
       directly.  As of 2.3 f_lineno is only valid when tracing is
       active (i.e. when f_trace is set).  At other times we use
       PyCode_Addr2Line to calculate the line from the current
       bytecode index. */
    int f_lineno;               /* Current line number */
    int f_iblock;               /* index in f_blockstack */
    char f_executing;           /* whether the frame is still executing */
    PyTryBlock f_blockstack[CO_MAXBLOCKS]; /* for try and loop blocks */
    PyObject *f_localsplus[1];  /* locals+stack, dynamically sized */
} PyFrameObject;
```

在 `_PyEval_EvalCodeWithName()` 中，会用 `_PyFrame_New_NoTrack()` 创建一个栈对象，这个栈是对 C 程序函数栈的模拟，在最后，会调用 `PyEval_EvalFrameEx()`。

`PyEval_EvalFrameEx()` 然后会在 `PyThreadState` 上调用 `eval_frame()/_PyEval_EvalFrameDefault()` 函数。这个函数也会被 Python 虚拟机调用。

跟踪进入 `_PyEval_EvalFrameDefault()`，我们可以观察到第 1054 行有一个无限循环，在不断产生字节码。

## 调试 Python 对象

`PyObject` 是通用 Python 对象，定义于 `Include/object.h` 中。

### 简介

每一种 `PyObject` 都有着相似的跟踪步骤：

* 用 gdb 打开 python
* 在对象创建函数上打断点
* 用交互式命令环境，创建我们想要的对象
* 在断点处，开始一步步跟踪代码

例如，我们想单步调试 `PyBoolObject`：

```
$ gdb python
(gdb) b bool_newbb
Breakpoint 1 at 0x44812f: file Objects/boolobject.c, line 44.
(gdb) r
[GCC 6.3.1 20170109] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> a = bool(1)

Breakpoint 1, bool_new (type=0x87a700 <PyBool_Type>, args=(1,), kwds=0x0) at Objects/boolobject.c:44
44	{
```

### PyObject

通用 Python 对象定义为

```c
typedef struct _object {
    _PyObject_HEAD_EXTRA
    Py_ssize_t ob_refcnt;
    struct _typeobject *ob_type;
} PyObjecdt;
```

在预处理器展开宏 `_PyObject_HEAD_EXTRA` 后，它会变成一个双向列表：

```c
typedef struct _object {
    struct _object *_ob_next;
    struct _object *_ob_prev;
    Py_ssize_t ob_refcnt;
    struct _typeobject *ob_type;
} PyObjecdt;
```

该对象包含两个重要元素：引用计数和类型对象。

### PyVarObject

Python 也有变长对象，定义为：

```c
typedef struct {
    PyObject ob_base;
    Py_ssize_t ob_size; /* Number of items in variable part */
} PyVarObject;
```

几乎和 `PyObject` 一样，但多了一项用于表示对象的长度信息。

### PyTypeObject

`PyTypeObject` 是 Python 对象的类型表示。在 Python 中可以如下表达式获取任何对象的类型信息：

```python
>>> t = type(1)
>>> dir(t)
['__abs__', '__add__', '__and__', '__bool__', '__ceil__', '__class__', '__delattr__', '__dir__', '__divmod__', '__doc__', '__eq__', '__float__', '__floor__', '__floordiv__', '__format__', '__ge__', '__getattribute__', '__getnewargs__', '__gt__', '__hash__', '__index__', '__init__', '__init_subclass__', '__int__', '__invert__', '__le__', '__lshift__', '__lt__', '__mod__', '__mul__', '__ne__', '__neg__', '__new__', '__or__', '__pos__', '__pow__', '__radd__', '__rand__', '__rdivmod__', '__reduce__', '__reduce_ex__', '__repr__', '__rfloordiv__', '__rlshift__', '__rmod__', '__rmul__', '__ror__', '__round__', '__rpow__', '__rrshift__', '__rshift__', '__rsub__', '__rtruediv__', '__rxor__', '__setattr__', '__sizeof__', '__str__', '__sub__', '__subclasshook__', '__truediv__', '__trunc__', '__xor__', 'bit_length', 'conjugate', 'denominator', 'from_bytes', 'imag', 'numerator', 'real', 'to_bytes']
```

这些方法都定义在 `PyTypeObject` 中：

```c
#ifdef Py_LIMITED_API
typedef struct _typeobject PyTypeObject; /* opaque */
#else
typedef struct _typeobject {
    PyObject_VAR_HEAD
    const char *tp_name; /* For printing, in format "<module>.<name>" */
    Py_ssize_t tp_basicsize, tp_itemsize; /* For allocation */

    /* Methods to implement standard operations */

    destructor tp_dealloc;
    printfunc tp_print;
    getattrfunc tp_getattr;
    setattrfunc tp_setattr;
    PyAsyncMethods *tp_as_async; /* formerly known as tp_compare (Python 2)
                                    or tp_reserved (Python 3) */
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
    unsigned long tp_flags;

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

    destructor tp_finalize;

#ifdef COUNT_ALLOCS
    /* these must be last and never explicitly initialized */
    Py_ssize_t tp_allocs;
    Py_ssize_t tp_frees;
    Py_ssize_t tp_maxalloc;
    struct _typeobject *tp_prev;
    struct _typeobject *tp_next;
#endif
} PyTypeObject;
#endif
```

所有的整数对象现在都在 `Objects/longobject.c` 中实现，定义为 `PyLong_Type` 类型。`PyLong_Type` 就是一个 `PyTypeObject` 对象。

```c
PyTypeObject PyLong_Type = {
    PyVarObject_HEAD_INIT(&PyType_Type, 0)
    "int",                                      /* tp_name */
    offsetof(PyLongObject, ob_digit),           /* tp_basicsize */
    sizeof(digit),                              /* tp_itemsize */
    long_dealloc,                               /* tp_dealloc */
    0,                                          /* tp_print */
    0,                                          /* tp_getattr */
    0,                                          /* tp_setattr */
    0,                                          /* tp_reserved */
    long_to_decimal_string,                     /* tp_repr */
    &long_as_number,                            /* tp_as_number */
    0,                                          /* tp_as_sequence */
    0,                                          /* tp_as_mapping */
    (hashfunc)long_hash,                        /* tp_hash */
    0,                                          /* tp_call */
    long_to_decimal_string,                     /* tp_str */
    PyObject_GenericGetAttr,                    /* tp_getattro */
    0,                                          /* tp_setattro */
    0,                                          /* tp_as_buffer */
    Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE |
        Py_TPFLAGS_LONG_SUBCLASS,               /* tp_flags */
    long_doc,                                   /* tp_doc */
    0,                                          /* tp_traverse */
    0,                                          /* tp_clear */
    long_richcompare,                           /* tp_richcompare */
    0,                                          /* tp_weaklistoffset */
    0,                                          /* tp_iter */
    0,                                          /* tp_iternext */
    long_methods,                               /* tp_methods */
    0,                                          /* tp_members */
    long_getset,                                /* tp_getset */
    0,                                          /* tp_base */
    0,                                          /* tp_dict */
    0,                                          /* tp_descr_get */
    0,                                          /* tp_descr_set */
    0,                                          /* tp_dictoffset */
    0,                                          /* tp_init */
    0,                                          /* tp_alloc */
    long_new,                                   /* tp_new */
    PyObject_Del,                               /* tp_free */
};
```

### PyLongObject

定义于 `Include/longobject.h`：

```c
typedef struct _longobject PyLongObject; /* Revealed in longintrepr.h */
```

### PyBoolObject

`PyBoolObject` 在 Python 中存储布尔类型，定义于 `Include/boolobject.h` 中。

### PyFloatObject

在 `Include/floatobject.h` 中：

```c
typedef struct {
    PyObject_HEAD
    double ob_fval;
} PyFloatObject;
```

### PyListObject

在 `Include/listobject.h` 中：

```c
typedef struct {
    PyObject_VAR_HEAD
    /* Vector of pointers to list elements.  list[0] is ob_item[0], etc. */
    PyObject **ob_item;

    /* ob_item contains space for 'allocated' elements.  The number
     * currently in use is ob_size.
     * Invariants:
     *     0 <= ob_size <= allocated
     *     len(list) == ob_size
     *     ob_item == NULL implies ob_size == allocated == 0
     * list.sort() temporarily sets allocated to -1 to detect mutations.
     *
     * Items must normally not be NULL, except during construction when
     * the list is not yet visible outside the function that builds it.
     */
    Py_ssize_t allocated;
} PyListObject;
```