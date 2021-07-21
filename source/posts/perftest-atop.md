title: atop 简单使用
date: 2018-07-04 16:36:38
categories: 系统
---

atop 是一个系统性能监控工具，可以在系统级别监控 CPU、内存、硬盘和网络的使用情况。

atop 不仅可以以交互式的方式运行，还可以一一定的频率，将性能数据写入日志中。所以当服务器出现问题之后，便可分析 atop 日志文件来判断是否有进程异常退出、内存和 CPU 方面的异常。

## 字段含义

#### PRC - Process level totals

1. sys, 内核态下运行时间
2. user, 用户态下运行时间
3. #proc, 当前所有的进程数量
4. #trun, 处于 running 状态下线程数量
5. #zombie，僵尸进程的数量
6. #exit，采样周期内退出的进程数量

#### CPU - CPU utilization

展示所有 CPU 的使用情况。在多处理器的系统中，会展示每一个独立内核的使用情况。

1. sys、usr, CPU 被用于处理进程时，进程在内核态、用户态所占CPU的时间比例
2. irq, CPU 被用于处理中断的时间比例
3. idle, CPU 处在完全空闲状态的时间比例
4. wait, CPU 处在“进程等待磁盘IO 导致 CPU 空闲”状态的时间比例

#### CPL - CPU load information

展示 CPU 的负载情况。

1. avg1、avg5和avg15：过去1分钟、5分钟和15分钟内运行队列中的平均进程数量
2. csw，指示上下文交换次数
3. intr，指示中断发生次数

#### MEM - Memory occupation

1. tot，物理内存总量
2. free，空闲内存大小
3. cache，页缓存的内存大小
4. buff，文件系统缓存的内存大小
5. slab，系统内核分配的内存大小
6. dirty，页缓存中脏内存的大小

#### SWP - Swap occupation and overcommit info

1. tot，交换区总量
2. free，示空闲交换空间大小

#### PAG - Paging frequency

1. swin，换入的页内存数目
2. swout， 换出的页内存数目

#### DSK/LVM - Disk utilization/Logical volumn

1. busy，磁盘忙时比例
2. read，读请求数量
3. write，写请求数量
4. KiB/r，每次读的千字节数
5. Kib/w，每次写的千字节数
6. MBr/s，每秒读入兆字节带宽
7. MBw/s，每秒写入兆字节带宽
8. avio，每次传输所需要的毫秒

#### NET - Network utilization (TCP/IP)

第一行是传输层信息，第二行是 IP 层信息，后面几行是各网卡的信息。

## 常用快捷键

1. g, 通用输出
2. m, 展示与内存有关的输出
3. d, 展示与硬盘使用有关的输出
4. c, 展示每个进程是由哪个命令行启动的
5. p, 展示进程相关的活动信息
6. C, 按照 CPU 使用排序
7. M, 按照内存使用排序
8. P, 按下后，即可输入正则表达式来搜索对应进程
9. t, 向前一个采样间隔，在分析 atop 日志时使用
10. T, 向后一个采样间隔，在分析 atop 日志时使用
11. v, 输出更详细的进程信息，包括进程的启动时间，进程号，用户和所在组，当前状态。


## atop日志

每个时间点采样页面组合起来就形成了一个atop日志文件，我们可以使用"atop -r XXX"命令对日志文件进行查看。

通常日志文件位于 `/var/log/`，采样间隔为 10min。