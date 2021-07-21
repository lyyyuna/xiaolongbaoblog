title: Python 访问外围作用域中的变量
date: 2016-09-10 21:33:31
categories: 语言
tags: 
- Python
---


在表达式中引用变量时，Python 会按照如下的顺序遍历各个作用域，寻找该变量：

1. 当前函数作用域
2. 任何外围作用域（比如包含当前函数的其他函数）
3. global 作用域，即代码所在的模块的作用域

如果上述作用域内都找不到变量，就会报 NameError 异常。

但是对变量赋值时，规则会有所不同。

1. 如果当前作用域变量已存在，那么其值会被替换。
2. 如果不存在，则会视为在当前作用域定义新变量，而不是向外围作用域中寻找。

如下函数

```python
def function():
    flag = True
    def helper():
        flag = False
    helper()
    print flag

function()
```

由于 helper 中变量是赋值，这里 flag 输出仍为 True。习惯了 c 语言之类静态类型语言，这种设计起初会感到困惑，但其可以有效地防止局部变量污染函数外的环境。

需求总是多样的，一定有程序员想在赋值时访问外围作用域。如果是 Python2，他可以这么做

```python
def function():
    flag = [True]
    def helper():
        flag[0] = False
    helper()
    print flag

function()
```

先用 flag[0] 是读操作，产生一次变量引用，寻找到外围作用域中 flag，这时候再赋值 flag[0] = False 便不会新定义变量了。

如果是 Python3，则可以使用 nonlocal 关键字。

```python
def function():
    flag = True
    def helper():
        nonlocal flag
        flag = False
    helper()
    print flag

function()
```

