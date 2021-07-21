title: 用 Python 实现一个 ORM
date: 2018-04-28 11:33:31
categories: 语言
tags: 
- Python
summary: 如何用元类控制类的生成？
---

## 前言

本文实现一个非常简单的 ORM 初稿：

1. 完成 Python 类（模型）与数据库表的映射
2. 完成类实例与表每行记录的映射
3. 完成实例操作与增删改查的 SQL 语句的映射

这个初稿不涉及数据库的真正操作，只是在 `user.save()` 的时候打印类似 `insert into user ...` 的 SQL 语句。本文所有代码基于 Python2。

## ORM 为什么需要元类

假设有如下的类：

```python
class User():
    __table__ = 'User_table'
    student_id = IntegerField('studentid', primaryKey=True)
```

回想 Django 的 ORM，每个模型都继承了一个 `Model` 类，我们也如法炮制。而所谓类与表的映射，就是在 Python 虚拟机启动后，自动寻找类属性，并将 `__table__` 转化为表名， `student_id` 转化为列名。这种需求类似于运行时自省，而普通类的 `__new__` `__init__` 都是实例化类时被调用，在这两个方法上做文章没有用处。

这时候就该用元类 `metaclass` 了。

在 [Python2.7 源码 - 整数对象](http://www.lyyyuna.com/2017/12/24/python-internal2-integer-object/) 中已经有过介绍，元类 `metaclass` 是类的类。除了整数这种内置类型，用户自定义类型也有元类的概念。

* 内置类定义在 C 源码中，故虚拟机运行后，就立即存在。
* 而用 `class` 语法定义的类，则需要根据元类 `metaclass` 来创建。
* 内置类也有元类，最终两者在虚拟机中拥有相同的结构。

元类 `metaclass` 实例化的结果就是我们的普通类，由虚拟机启动时自动执行。在元类实例化的过程中，便可以扫描类定义属性，实现类与表的映射。自定义类默认继承自 `object`，获得的元类为 `type`。

Python2.x 中，用以下语法

```python
class C():
    __metaclass__ = Meta
```

可以将类 `C` 对应的元类替换为 `Meta`。这么一看，只要设计自己的元类，并在模型中添加进去就可以了：

```python
class User():
    __metaclass__ = Meta
    __table__ = 'User_table'
    student_id = IntegerField('studentid', primaryKey=True)
```

但这么做，会在产品业务代码中暴露太多复杂的概念。我们可以设计一个公共的父类，并修改此父类的元类，这样所有继承的子类都能获得新的元类：

```python
class ModelType(type):
    def __new__(cls, name, bases, attrs):
        return type.__new__(cls, name, bases, attrs)

class Model():
    __metaclass__ = ModelType

# Application
class User(Model):
    __table__ = 'User_table'

class Teacher(Model):
    __table__ = 'Teacher_table'
```

现在当 `User` `Teacher` 类在虚拟机中创建时，其行为就由 `ModelType` 控制。

`ModelType` 需实现 `__new__` 函数，接受 `name, bases, attrs` 三个参数，`attrs` 是类属性构成的字典。该 `__new__` 函数最终需返回 `type.__new__(cls, name, bases, attrs)` 的结果，即类在虚拟机中的表示。

其中，`attrs` 字典由 Python 虚拟机创建类时填入。

## 扫描表名

由上可知，可在元类的 `__new__` 函数中获取 `__table__`：

```python
class ModelType(type):
    def __new__(cls, name, bases, attrs):
        tablename = attrs.get('__table__', None) or name
        print 'Get table name', tablename
        return type.__new__(cls, name, bases, attrs)

class Model():
    __metaclass__ = ModelType

class User(Model):
    __table__ = 'User_table'
```

输出

```
Get table name Model
Get table name User_table
```

为什么会输出两行？因为父类 `Model` 也会被元类 `ModelType` 扫描，所以应将父类自身排除出扫描：

```python
class ModelMeta(type):
    def __new__(cls, name, bases, attrs):
        if name=='Model':
            return type.__new__(cls, name, bases, attrs)

        tablename = attrs.get('__table__', None) or name
        print 'Get table name', tablename
```

这次输出没有 `Model` ：

```
Get table name User_table
```

## 扫描字段

字段的扫描是类似的，我们先完成字段类型的定义，让所有字段继承自 `Field` 类。

```python
class Field(object):
    def __init__(self, name, sqlType, primaryKey, default):
        self.name = name
        self.sqlType = sqlType
        self.primaryKey = primaryKey
        self.default = default

    def __str__(self):
        return '<%s, %s:%s>' % (self.__class__, self.sqlType, self.name)


class StringField(Field):
    def __init__(self, name=None, sqlType='character varying(100)', primaryKey=False, default=''):
        super(StringField, self).__init__(name, sqlType, primaryKey, default)


class IntegerField(Field):
    def __init__(self, name=None, sqlType='integer', primaryKey=False, default=0):
        super(IntegerField, self).__init__(name, sqlType, primaryKey, default)
```

由于所有表字段类有一个相同的父类，所以可以通过 `isinstance(v, Field)` 识别一个类属性是否属于表字段。以下代码在类属性 `attrs` 中遍历，找出字段并存储在 `mappings` 字典和 `fields` 列表中。

```python
mappings = {}
fields = []
primary = None
for k, v in attrs.iteritems():
    if isinstance(v, Field):
        print 'Found one field', k
        mappings[k] = v
        if v.primaryKey == True:
            if primary == None:
                primary = k
            else:
                raise RuntimeError("Duplicate primary key: %s", k)
        else:
            fields.append(k)
if primary == None:
    raise RuntimeError("No primary key given.")
```

将上述代码应用于 `User`：

```python
class User(Model):
    __table__ = 'User_table'
    student_id = IntegerField('studentid', primaryKey=True)
    name = StringField('username')
    age = IntegerField('age')
```

输出：

```
Get table name User_table
Found one field age
Found one field name
Found one field student_id
```

## 实例属性

在类中需要移除所有的字段属性 `attrs.pop(k)`。

```python
for k in mappings.keys():
    attrs.pop(k)
```

因为业务代码不需要类属性，每个实例需要访问的是实例属性。比如

```python
u2 = User(student_id=3, name='blue', age=123)
print(u2.name)
```

这里，让父类 `Model` 继承 `dict` 类，并添加 `__getattr__` 和 `__setattr__` 方法：

```python
class Model(dict):
    __metaclass__ = ModelMeta
    __table__ = 'Should not show'

    def __init__(self, **kw):
        super(Model, self).__init__(**kw)

    def __getattr__(self, key):
        try:
            return self[key]
        except:
            raise AttributeError(r"'Dict' object has no attribute '%s'" % key)

    def __setattr__(self, key, value):
        self[key] = value
```

访问 `u2.name` 便等价于 `u2[name]`，而 `User` 间接继承自字典，`User(student_id=3, name='blue', age=123)`初始化后，便能访问字典元素`u2[name]`。

## 与 SQL 语句的映射

我们可以在元类 `ModelMeta` 中预先定义一些 SQL 语句的模板，并存储在类属性 `attrs` 中：

```python
    attrs['__select__'] = "select '%s', %s from '%s'" % (primary, ','.join(escaped_fields), tablename)
    attrs['__insert__'] = "insert into '%s' (%s, '%s') values (%s)" % (tablename, ','.join(escaped_fields), primary, create_args_string(len(escaped_fields)+1))
    attrs['__update__'] = "update '%s' set %s where '%s' =?" % (tablename, ','.join(map(lambda x: "'%s'=?" % (mappings.get(x).name), fields)), primary)
    attrs['__delete__'] = "delete from '%s' where '%s' = ?" % (tablename, primary)
```

接着在 `Model` 中添加 `select` 和 `save` 方法。

```python
    @classmethod
    def select(cls, id):
        print "%s where '%s' = %s;" % (cls.__select__, cls.__primarykey__, id)

    def getValue(self, k):
        value = getattr(self, k, None)
        if value is None:
            field = self.__mappings__[k]
            if field.default is not None:
                value = field.default
                setattr(self, k, value)
        return value

    def save(self):
        args = map(self.getValue, self.__fields__)
        args.append(self.getValue(self.__primarykey__))
        print self.__insert__, args
```

这两个方法对应于 SQL 中的 `SELECT` 和 `INSERT` 语句。逻辑也比较简单，从之前存储在 `mappings` 和 `fields` 的字段名，再结合实例属性，拼接出 SQL 语句。

```python
print 'Test select sql command:'
User.select(id=1)
print

print 'Test insert sql command:'
u2 = User(student_id=3, name='blue', age=123)
u2.save()
```

输出：

```sql
Test select sql command:
select 'student_id', 'age','name' from 'User_table' where 'student_id' = 1;

Test insert sql command:
insert into 'User_table' ('age','name', 'student_id') values (?, ?, ?) [123, 'blue', 3]
```

## 小结

至此，一个简单的 ORM 原型就完成了。详细代码可见：[https://github.com/lyyyuna/script_collection/tree/master/orm_draft](https://github.com/lyyyuna/script_collection/tree/master/orm_draft)