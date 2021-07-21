title: c++ 虚函数是如何实现的
date: 2016-12-21 20:35:53
categories: 语言
tags: 
- cpp
---


## 前言

探索 c++ 对象内部的实现是一件非常有趣的事情。c++ 分为编译时多态和运行时多态。运行时多态依赖于虚函数，大部分人或许听说过虚函数是由虚函数表+虚函数指针实现的，但，真的是这样吗？虽然 c++ 规范有着复杂的语言细节，但底层实现机制却任由编译器厂商想象。（没准某种特殊的处理器电路结构原生支持虚函数，没准这个处理器压根不是冯纽曼型，或者将来厂商发明了比虚函数表更有效率的数据结构。）

本篇文章就来实际检验一下 Visual Studio 2013 编译器在无优化条件下，虚函数的实现。

## 虚函数表

封装把实例的数据和操作结合在了一起，但实例本身只有数据，没有函数，同一个类的函数是共享的。我们通过一个例子来间接证明这一点

```cpp
class Base1
{
public:
	int a;
	void func() { cout << "heel" << endl; }
};

Base1 b1;
cout << sizeof(b1) << endl;
```

打印

    4

如果类中有虚函数，则会在对象中加入一个虚函数指针，该指针指向一个虚函数表，表中是各个虚函数的地址。

    +--------+       +---------+
    | pvtbl  |------>| vfunc1  |
    +--------+       +---------+
    | data1  |       | vfunc2  |
    +--------+       +---------+
    | ...    |       | ...     |

当子类继承父类时，会依次覆盖虚函数表中的各个项，如果子类没有重写某项，那该项就保留。当实例化对象后，虚函数指针就作为一个隐藏数据存在于实例中。如果通过父类指针调用普通成员函数，由于普通函数和类型绑定在一起，所以仍会调用父类成员函数；如果通过父类指针调用虚函数，则会通过对象的虚指针找到虚函数表（即子类的虚函数表），定位虚函数项，实现多态。

原理是不是很简单？c++ 就是通过这种看似原始的方式实现高级抽象。以上是编译器的通用做法，我手上的 Visual Studio 2013 编译器就是这么做的，为了提高性能，VS 保证虚函数指针存在于对象实例中最前面位置（历史上也有编译器不这么做，好像是 Borland 的？）。

## Visual Studio 2013 中的实现

来一个例子（能这么写是因为我已知了 Visual Studio 2013 编译后对象的内存布局）

```cpp
#include <iostream>
using namespace std;

class Base 
{
public:
	typedef void (*func)();
	virtual void func1() { cout << "Base::func1" << endl; }
	virtual void func2() { cout << "Base::func2" << endl; }
	virtual void func3() { cout << "Base::func3" << endl; }
};

class Derived: public Base
{
public:
	virtual void func1() { cout << "Derived::func1" << endl; }
	virtual void func3() { cout << "Derived::func3" << endl; }
};

int main()
{
	Base b, b1;
	int** pvirtualtable1 = (int**)&b;
	cout << "Base object vtbl address: " << pvirtualtable1[0] << endl;
	int** pvirtualtable11 = (int**)&b1;
	cout << "another Base object vtbl address: " << pvirtualtable11[0] << endl;
	cout << "function in virtual table" << endl;
	for (int i = 0; (Base::func)pvirtualtable1[0][i] != NULL; ++i)
	{
		auto p = (Base::func)pvirtualtable1[0][i];
		p();
	}
	cout << endl;

	Derived d;
	int** pvirtualtable2 = (int**)&d;
	cout << "Derived object vtbl address: " << pvirtualtable2[0] << endl;
	cout << "function in virtual table" << endl;
	for (int i = 0; (Base::func)pvirtualtable2[0][i] != NULL; ++i)
	{
		auto p = (Base::func)pvirtualtable2[0][i];
		p();
	}
	cout << endl;
}
```

打印

    Base object pvtbl address: 0029DA58
    another Base object pvtbl address: 0029DA58
    function address in virtual table
    Base::func1
    Base::func2
    Base::func3

    Derived object pvtbl address: 0029DB20
    function address in virtual table
    Derived::func1
    Base::func2
    Derived::func3


可以看到，同一类型不同实例的虚函数表是相同的，继承之后，子类有了自己的虚函数表，表也有相应的更新(Derived::func1, Derived::func3)，表中未重写的项还保留为原值(Base::func2)。
