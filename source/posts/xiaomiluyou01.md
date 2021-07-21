title: 小米路由器 mini 刷机记录
date: 2015-12-19 08:47:52
categories: 系统
tags: openwrt
---

## 前言

以前并没有玩过路由器，但是单片机嵌入式玩过不少，很多概念都是相通的。

刷新路由器的固件，其实也就是烧写 flash 的过程，有些系统也可能将程序存储在 EEPROM 中。很多网上的路由器资料中都会提到什么 TTL, JTAG, uboot 烧写等等，他们有什么区别呢。

### flash 种类

我知道的 flash 分为两种，一种是类似 norflash，自带数据总线和地址总线，代码可以直接在 flash 运行，不需要像 PC 那样将程序从硬盘 copy 到内存中。在单片机中，大部分都是这种结构。

第二种是类似 nandflash，只有数据总线，cpu 无法寻址，代码必须先从 flash 拷贝到 ram 中，才能真正开始执行代码。

### 代码的启动方式

首先明确一点，cpu 必须通过数据总线和地址总线寻址到代码，才能启动。对于 norflash 这种，非常方便，只要将 flash 地址配置到 cpu 上电后的第一条指令所在的范围即可。在早期的嵌入式设计中，经常在系统中同时配有 norflash 和 nandflash，norflash 比较贵且容量较小，所以里边一般只放置搬运代码。cpu 从 norflash启动后，执行里面的搬运代码将 nandflash 中真正的大段产品代码搬运到 ram 中。

而 nandflash 启动就需要处理器支持。现在的 cpu 会在内部自带上述的搬运代码，比如三星的 ARM11 S3C6410，假如配置成 nandflash 启动（通过某些引脚的高低电平），会自动将 nandflash 前 4K 代码 copy 到 ram 中。

我知道有较新的 cpu 支持从 sd 卡启动，这其实和 nandflash 启动原理类似。比如树莓派。

### flash 烧写方式

最暴力原始的烧写方式，就是自制烧录器。比如一些 flash 的接口是 SPI，通过单片机将 PC 的串口数据转成 SPI 数据写进 flash。

但路由器板子上 flash 已经焊死，很多人并不会焊接，这时就可以借助 JTAG 接口。我知道的 JTAG 是一种调试处理器的接口，一些编程器实现了该接口，比如 j-link。通过 j-link 可以直接操纵 cpu 对 flash 进行烧写。但实际上 j-link 烧写支持的 flash 种类是有限的，据我所知对 nandflash 支持一般（或者没有支持？）。而且正版的 j-link 非常昂贵（> 1000$）。

除了以上两种烧录方式，其他都得借助 BootLoader。

### BootLoader

首先在逻辑上对 flash 内容分区，其中 cpu 上电后执行的第一分区放入 BootLoader，将读到的 TTL，网口，USB 等等数据烧录到另一分区中，实现自我烧写，烧写完成后跳入另一分区中执行产品代码。有了 BootLoader，只要不破坏这一段分区内容，就能一直烧录。

经典的 BootLoader 有 uboot。显然 BootLoader 的出现极大的方便了路由器的开发，节省了成本。一个好的 BootLoader 完全可以同时支持 TTL，网口，USB，sd卡。

## 小米路由器刷固件过程

小米路由器的 BootLoader 支持 usb 刷固件。系统启动后，可以从管理员页面刷新固件。

因为不清楚小米官方 BootLoader 会不会对固件有限制，我打算将 BootLoader 和 固件 全部刷新。

### 获取 root 权限和 ssh

首先要根据 [官方指南](https://d.miwifi.com/rom/ssh) 在网页管理员页面刷成开发板固件，然后绑定小米账号获取 ssh 开发工具，再从 usb 烧录 ssh 开发工具。

ssh 登陆后，首先查看系统的分区。

    root@XiaoQiang:~# cat /proc/mtd

    dev:    size        erasesize   name
    mtd0:   01000000    00010000    "ALL"
    mtd1:   00030000    00010000    "Bootloader"
    mtd2:   00010000    00010000    "Config"
    mtd3:   00010000    00010000    "Factory"
    mtd4:   00c80000    00010000    "OS1"
    mtd5:   00b19a3b    00010000    "rootfs"
    mtd6:   00200000    00010000    "OS2"
    mtd7:   00100000    00010000    "overlay"
    mtd8:   00010000    00010000    "crash"
    mtd9:   00010000    00010000    "reserved"
    mtd10:  00010000    00010000    "Bdata"

可以将一些分区备份至 U 盘中。比如 Bdata 分区应该有 sn 信息，如果你对保修比较在意。

    root@XiaoQiang:~# dd if=/dev/mtd1 of=/extdisks/sda1/xiaomi-bootloader.bin

### 刷 BootLoader 

我选择 hackpascal 开发的 [Breed](http://www.right.com.cn/forum/thread-161906-1-1.html) 作为我的 BootLoader，它支持 LAN 口，更方便。

    root@XiaoQiang:~# wget -O /tmp/breed.bin http://breed.hackpascal.net/latest/breed-mt7620-xiaomi-mini.bin
    root@XiaoQiang:~# mtd -r write /tmp/breed.bin Bootloader

按照 [Breed](http://www.right.com.cn/forum/thread-161906-1-1.html) 的说明运行 BreedEnter.exe 后，重启路由器。这时路由器和 PC 通过网线直连。按照 hackpascal 所说，Breed 是带 dhcp 的，实际测试没有成功，我手动配置了电脑的 ip 为 192.168.1.2，打开网页 http://192.168.1.1。

### 刷固件 

有了 Breed 之后，刷固件就简单多了。下载 [PandoraBox_xiaomi_20150608](http://downloads.openwrt.org.cn/PandoraBox/Xiaomi-Mini-R1CM/stable/PandoraBox-ralink-mt7620-xiaomi-mini-squashfs-sysupgrade-r1024-20150608.bin)。按照 http://192.168.1.1 的指示上传即可。对于小米的固件_可能_需要_配置一下固件的启动方式_。

你也可以从后台

    root@XiaoQiang:~# mtd -r write /tmp/PandoraBox-ralink-mt7620-xiaomi-mini-squashfs-sysupgrade-r1024-20150608.bin OS1

但有了 Breed 减少了你手抖刷错变砖的几率。

重启完后，路由器就是 PandoraBox 系统了。分区的名称也变了。
      _______________________________________________________________ 
     |    ____                 _                 ____               |
     |   |  _ \ __ _ _ __   __| | ___  _ __ __ _| __ )  _____  __   |
     |   | |_) / _` | '_ \ / _` |/ _ \| '__/ _` |  _ \ / _ \ \/ /   |
     |   |  __/ (_| | | | | (_| | (_) | | | (_| | |_) | (_) >  <    |
     |   |_|   \__,_|_| |_|\__,_|\___/|_|  \__,_|____/ \___/_/\_\   |
     |                                                              |
     |                  PandoraBox SDK Platform                     |
     |                  The Core of SmartRouter                     |
     |       Copyright 2013-2015 D-Team Technology Co.,Ltd.SZ       |
     |                http://www.pandorabox.org.cn                  |
     |______________________________________________________________|
      Base on OpenWrt BARRIER BREAKER (14.09, r1024)
    [root@PandoraBox:/root]#cat /proc/mtd
    dev:    size   erasesize  name
    mtd0: 00030000 00010000 "u-boot"
    mtd1: 00010000 00010000 "u-boot-env"
    mtd2: 00010000 00010000 "Factory"
    mtd3: 01000000 00010000 "fullflash"
    mtd4: 00f80000 00010000 "firmware"
    mtd5: 001230fb 00010000 "kernel"
    mtd6: 00e3cf05 00010000 "rootfs"
    mtd7: 00860000 00010000 "rootfs_data"
    mtd8: 00020000 00010000 "panic_oops"
    mtd9: 00010000 00010000 "culiang-crash"
    mtd10: 00010000 00010000 "culiang-reserved"
    mtd11: 00010000 00010000 "culiang-Bdata"
