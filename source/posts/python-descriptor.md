title: Python 中的 descriptor
date: 2017-09-14 16:20:02
categories: 语言
tags: 
- Python
---

## 定义

通常，一个 descriptor 是具有“绑定行为”的对象属性。所绑定行为可通过 descriptor 协议被自定义的 `__get__()`, `__set__()` 和 `__delete__()` 方法重写。如果一个对象的上述三个方法任意一个被重写，则就可被称为 descriptor。

属性的默认操作是从对象字典中获取、设置和删除一个属性。例如，a.x 有一个查找链，先 `a.__dict__['x']`，若没有则 `type(a).__dict__['x']`，若没有增往上查找父类直到元类。如果查找链中，对象被定义了 descriptor 方法，Python 就会覆盖默认行为。

Descriptor 是一个强大的工具，虽然开发者不常接触到它，但它其实就是类、属性、函数、方法、静态方法、类方法以及 `super()` 背后的运行机制。

## Descriptor 协议

三个方法原型如下所示：

```python
descr.__get__(self, obj, type=None) --> value

descr.__set__(self, obj, value) --> None

descr.__delete__(self, obj) --> None
```

数据 descriptor 是同时具有 `__get__()` 和 `__set__()` 方法的对象，若只有 `__get__()` 方法，则为非数据 descriptor。如果实例字典中有和数据 descriptor 同名的入口，则数据 descriptor 优先级更高。相反，非数据 descriptor 优先级低。

让 `__set__()` 方法抛出异常，就能创建一个只读数据 descriptor。

## 调用 descriptor

descriptor 可以直接通过方法名调用。例如，`d.__get__(obj)`。

而通过访问对象属性，自动调用 descriptor 才是更通用的做法。例如，如果 `d` 定义了方法 `__get__()`，则 `obj.d` 会调用 `d.__get__(obj)`。

对于对象，`b.x` 会被转换成 `type(b).__dict__['x'].__get__(b, type(b))`。而对于类（是的，类也可以调用），`B.x` 会被转换成 `B.__dict__['x'].__get__(None, B)`。

## Descriptor 例子

```python
class RevealAccess(object):
    """A data descriptor that sets and returns values
       normally and prints a message logging their access.
    """

    def __init__(self, initval=None, name='var'):
        self.val = initval
        self.name = name

    def __get__(self, obj, objtype):
        print('Retrieving', self.name)
        return self.val

    def __set__(self, obj, val):
        print('Updating', self.name)
        self.val = val

>>> class MyClass(object):
...     x = RevealAccess(10, 'var "x"')
...     y = 5
...
>>> m = MyClass()
>>> m.x
Retrieving var "x"
10
>>> m.x = 20
Updating var "x"
>>> m.x
Retrieving var "x"
20
>>> m.y
5
```


