title: c++ 函数重载是如何实现的
date: 2016-12-22 20:35:53
categories: 语言
tags: 
- cpp
---


函数重载是 c++ 的编译时多态的一部分，也就是说，该行为在编译完成后即是确定的。事实上，这是编译器和链接器之间玩的小花招。链接器通过符号（symbol）定位各个函数，所谓符号可以简单理解为一个字符串。

编译器会给每个函数名一个符号，在 c 语言中，符号名只和函数名有关。

来一个 c 语言程序的例子，使用 Visual Studio 编译

```cpp
void add(int a, int b)
{}

int main()
{
	add(1, 2);
}
```

用 Visual Studio 自带的工具 dumpbin 查看 .obj 文件的符号表

    017 00000000 SECT4  notype ()    External     | _add

我们换一个函数声明

```cpp
void add(double a, double b)
{}

int main()
{
	add(1, 2);
}
```

再用 dumpbin 查看 .obj 文件的符号表

    017 00000000 SECT4  notype ()    External     | _add

还是同样的符号。所以， c 语言编译器不支持函数重载，函数名相同的话，链接器永远只能看到一个名字。

那么，c++ 呢？

```cpp
#include <iostream>
#include <string>

using namespace std;

void add(int a, int b)
{}
void add(double a, double b)
{}
void add(string a, string b)
{}

int main()
{
	add(1, 2);
	add(1.0, 2.0);
	add(string("1"), string("2"));

	return 0;
}
```

再用 dumpbin 查看 .obj 文件的符号表

    2F3 00000000 SECT87 notype ()    External     | ?add@@YAXHH@Z (void __cdecl add(int,int))
    2F4 00000000 SECT89 notype ()    External     | ?add@@YAXNN@Z (void __cdecl add(double,double))
    2F5 00000000 SECT8B notype ()    External     | ?add@@YAXV?$basic_string@DU?$char_traits@D@std@@V?$allocator@D@2@@std@@0@Z (void __cdecl add(class std::basic_string<char,struct std::char_traits<char>,class std::allocator<char> >,class std::basic_string<char,struct std::char_traits<char>,class std::allocator<char> >))

可以看到，每个符号都不一样啦。这时候的函数声明不仅和函数名有关，也和参数类型有关，但和返回类型无关。符号能唯一确定，编译器自然也能顺利实现重载。

顺便可以发现，同一个函数声明在 c 和 c++ 中是完全不一样的。这也是为什么 c 和 c++ 之间动静态库不能直接互相调用的原因。为此 cpp 使用了 extern "C" 语法，强制使用 c++ 编译器使用 c 语言的符号命名方法。

我们实验一下

```cpp
#include <iostream>
#include <string>

using namespace std;

extern "C"{
	void add(int a, int b)
	{}
}
void add(double a, double b)
{}
void add(string a, string b)
{}

int main()
{
	add(1, 2);
	add(1.0, 2.0);
	add(string("1"), string("2"));

	return 0;
}
```

查看 .obj 文件的符号表

    2F3 00000000 SECTD3 notype ()    External     | _add
    2F4 00000000 SECT87 notype ()    External     | ?add@@YAXNN@Z (void __cdecl add(double,double))
    2F5 00000000 SECT89 notype ()    External     | ?add@@YAXV?$basic_string@DU?$char_traits@D@std@@V?$allocator@D@2@@std@@0@Z (void __cdecl add(class std::basic_string<char,struct std::char_traits<char>,class std::allocator<char> >,class std::basic_string<char,struct std::char_traits<char>,class std::allocator<char> >))

可以看到，第一个函数的符号和 c 语言一致了。