title: PostgreSQL 中的函数稳定性
date: 2018-07-06 16:36:38
categories: 数据库
tags: PostgreSQL
---

## 定义

PostgreSQL 中函数有三个稳定性状态可选：

1. immutable，函数不可以修改数据库的数据,并且在任何情况下调用，只要输入参数一致，返回结果都一致。
2. stable，函数不可以修改数据库的数据，同一个QUERY中，如果需要返回该函数的结果，那么将合并多次运算为一次这个函数。
3. volatile，函数可以修改数据库的数据，输入同样的参数可以返回不同的结果，同一个QUERY中，如果需要返回该函数的结果，那么每一行都会运算一遍这个函数。

函数的稳定性会影响执行计划。在索引比较的时候，被比较的值只会运算一次，所以 volatile 不能被执行计划选择作为索引的比较条件。

## 例子

### 查看函数的稳定性

```
ddei=# select proname, provolatile from pg_proc where proname in ('now', 'clock_timestamp');
     proname     | provolatile 
-----------------+-------------
 now             | s
 clock_timestamp | v
(2 rows)
```

其中 clock_timestamp 是 voatile, now 是 stable。

### 测试插入语句

创建一个测试表

```
ddei=# create table test(id int, time1 timestamp, time2 timestamp);
CREATE TABLE
ddei=# insert into test select generate_series(1,1000),clock_timestamp(), now();
INSERT 0 1000
```

插入语句，对于 stable 函数 `now()` 应该只执行一次：

```
ddei=# select count(*),count(distinct time1),count(distinct time2) from test;
 count | count | count 
-------+-------+-------
  1000 |  1000 |     1
(1 row)
```

### 测试对索引的影响

在测试表上创建索引，并查看执行计划：

```
ddei=# create index test_idx on test(time1);
CREATE INDEX
ddei=# 
ddei=# explain select * from test where time1>now();
                              QUERY PLAN                              
----------------------------------------------------------------------
 Index Scan using test_idx on test  (cost=0.00..4.27 rows=1 width=20)
   Index Cond: (time1 > now())
(2 rows)

ddei=# explain select * from test where time1>clock_timestamp();
                       QUERY PLAN                       
--------------------------------------------------------
 Seq Scan on test  (cost=0.00..22.00 rows=333 width=20)
   Filter: (time1 > clock_timestamp())
(2 rows)
```

对于 volatile 的函数 clock_timestamp 在 where 条件中，不走索引。而 stable 函数 now 在 where 条件中，会走索引。

## 修改函数稳定性

使用以下语句可修改函数稳定性：

```
ddei=# alter function clock_timestamp() strict stable;
ALTER FUNCTION
```

再次测试 clock_timestamp 的索引情况：
```
ddei=# explain select * from test where time1>clock_timestamp();
                              QUERY PLAN                              
----------------------------------------------------------------------
 Index Scan using test_idx on test  (cost=0.00..4.27 rows=1 width=20)
   Index Cond: (time1 > clock_timestamp())
(2 rows)
```

这次 clock_timestamp 在 where 条件中走了索引。

不过不要随意修改系统自带函数的稳定性。