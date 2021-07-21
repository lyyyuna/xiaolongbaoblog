title: shellcode学习-绕过条件判断
date: 2015-08-30 16:36:38
categories: 系统
tags: shellcode
summary: shellcode 管窥。
---

shellcode学习第一个例子。

以下有一段c语言编写的命令行程序，检验用户输入的数字，并判断是否合法。这里用户的输入被放在了函数的缓冲区里，但程序没有对缓冲区长度做检查，留下了漏洞。这里可以利用该漏洞绕过数字检察，使得任意输入都会被判定为正确。
在 _validate_serial_ 中，_do_valid_stuff_ 的地址溢出到函数的返回值上，就可实现。

## 源程序

    #include <stdio.h>
    #include <stdlib.h>
    #include <string.h>
    
    int valid_serial(char * psz)
    {
        size_t len = strlen(psz);
        unsigned total = 0;
        size_t i;
    
        if (len<10)
            return 0;
    
        for (i = 0; i < len; i++)
        {
            if ((psz[i]<'0') || (psz[i]>'z'))
                return 0;
            total += psz[i];
        }
    
        if (total % 853 == 83)
            return 1;
    
        return 0;    
    }
    
    int valildate_serial()
    {
        char serial[24];
    
        fscanf(stdin, "%s", serial);
    
        if (valid_serial(serial))
            return 1;
        else
            return 0;
    }
    
    int do_valid_stuff()
    {
        printf("the serial number is valid!\n");
        exit(0);
    }
    
    int do_invalid_stuff()
    {
        printf("invalid serial number!\nexiting\n");
        exit(1);
    }
    
    int main(int argc, char * argv[])
    {
        if (valildate_serial())
            do_valid_stuff();
        else
            do_invalid_stuff();
    
        return 0;
    }

## 反汇编main

    (gdb) disass main
    Dump of assembler code for function main:
       0x0804861a <+0>:     push   %ebp
       0x0804861b <+1>:     mov    %esp,%ebp
       0x0804861d <+3>:     call   0x804859f <valildate_serial>
       0x08048622 <+8>:     test   %eax,%eax
       0x08048624 <+10>:    je     0x804862d <main+19>
       0x08048626 <+12>:    call   0x80485de <do_valid_stuff>
       0x0804862b <+17>:    jmp    0x8048632 <main+24>
       0x0804862d <+19>:    call   0x80485fc <do_invalid_stuff>
       0x08048632 <+24>:    mov    $0x0,%eax
       0x08048637 <+29>:    pop    %ebp
       0x08048638 <+30>:    ret
    End of assembler dump.
    
可得到 _do_valid_stuff_ 的地址为 0x08048626。_validate_serial_ 的返回地址为 0x08048622。下面就通过溢出修改返回地址。

## 缓冲区溢出

源码中，缓冲区长度为24，理论上只要覆盖24+2处的数据就可以了。我们需要检验一下，在fscanf处打断点，观察堆栈内容。

    Breakpoint 1, valildate_serial () at serial.c:31
    31          fscanf(stdin, "%s", serial);
    (gdb) x/20x $esp
    0xbffff6bc:     0x0804869b      0x00000001      0xbffff794      0xbffff79c
    0xbffff6cc:     0xbffff6e8      0xb7e987f5      0xb7ff0590      0x0804865b
    0xbffff6dc:     0xb7fc7ff4      0xbffff6e8      0x08048622      0xbffff768
    0xbffff6ec:     0xb7e7fe46      0x00000001      0xbffff794      0xbffff79c
    0xbffff6fc:     0xb7fe0860      0xb7ff6821      0xffffffff      0xb7ffeff4
    (gdb)c
    AAAAAAAAAABBBBBBBBBBCCCCCCCC1234
    
    Breakpoint 2, valildate_serial () at serial.c:33
    33          if (valid_serial(serial))
    (gdb) x/20x $esp
    0xbffff6bc:     0xb7fc8440      0x080486d0      0xbffff6c8      0x41414141
    0xbffff6cc:     0x41414141      0x42424141      0x42424242      0x42424242
    0xbffff6dc:     0x43434343      0x43434343      0x34333231      0xbffff700
    0xbffff6ec:     0xb7e7fe46      0x00000001      0xbffff794      0xbffff79c
    0xbffff6fc:     0xb7fe0860      0xb7ff6821      0xffffffff      0xb7ffeff4
    (gdb)

可以看到“1234”对应的ascii码“0x34333231”已经被写入了返回值"0x08048622"原来所在的地方。
接下来把“1234”换成我们需要的返回地址。

我们回到shell中实验一下

    lyyyuna@yan:~/Desktop/shellcode/validate_serial$ printf "AAAAAAAAAABBBBBBBBBBCCCCCCCC\x26\x86\x04\x08" | ./serial
    the serial number is valid!

成功绕过了程序的检验机制。
