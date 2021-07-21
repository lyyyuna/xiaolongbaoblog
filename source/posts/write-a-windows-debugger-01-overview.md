title: Writing Your Own Windows Debugger - Overview
date: 2017-04-27 17:36:25
categories: 系统
tags: 
- Windows Debugger
---


## Introduction

Debuggers are the apple of the hacker's eye. We benefit a lot from the debugger, but few of us know the principle of it.

In the book _Gray Hat Python_ , the author has constructed a simple debugger. However, it is too simple, it is only a machine language level debugger, and can only set basic breakpoints and show CPU register information. We also want to know how to 

* Show source code 
* Set breakpoint based on lines, not memory address
* Set Step In, Step Out, Step Over
* Show stack trace
* Show global and local variables

In this Chinese blog [Zplutor's](http://www.cnblogs.com/zplutor/archive/2011/03/04/1971279.html), I find a excellent series which has covered most above topics. I decide to write a English blog about it, and I will turn his code into a C++ version.

Before getting started, let's make some limitations:

* It is only a user mode debugger.
* It is only a Windows debugger. Although the principle is quite same, but Windows has offered lots of convenient APIs. The implementation will be different on Linux.
* It is only a terminal-based debugger.
* Different from _Gray Hat Python_ , the debugger will be implemented by C++.
* The debuggee program is single thread.

The modified debugger can be found [here](https://github.com/lyyyuna/anotherDebugger). It is only tested under Windows 10 + Visual Studio 2013.

## To Start the Debuggee Program

The so-called user mode debugger is to debug the program in user mode. Windows has provided a series of open API for debugging, and they can be devided into three categories:

* API for starting the debuggee program
* API for handling debug event during debug loop
* API for inspecing and modifying debuggee program

The first thing to do before debugging a program is to start it. On Windows, we use *CreateProcess* to start to program:

```c++
STARTUPINFO startupinfo = { 0 };
startupinfo.cb = sizeof(startupinfo);
PROCESS_INFORMATION processinfo = { 0 };
unsigned int creationflags = DEBUG_ONLY_THIS_PROCESS | CREATE_NEW_CONSOLE;

if (CreateProcess(
    "L:\\git_up\\anotherDebugger\\anotherDebugger\\Debug\\test.exe",
    //path,
    NULL,
    NULL,
    NULL,
    FALSE,
    creationflags,
    NULL,
    NULL,
    &startupinfo,
    &processinfo) == FALSE)
{
    std::cout << "CreateProcess failed: " << GetLastError() << std::endl;
    return;
}
```

* DEBUG_ONLY_THIS_PROCESS means the subprocess of the debuggee will not be debugged. If you need subprocess, use DEBUG_PROCESS.
* CREATE_NEW_CONSOLE means the debuggee's and debugger's output will be separated in two consoles.
* If the debugger process exits, the debuggee will also exit.

## Debugger loop

The debugger loop is a bit like Windows GUI message loop, some operations and exceptions will stop the debuggee and send event to the debugger. We always use 

```c++
DEBUG_EVENT debugEvent;
WaitForDebugEvent(&debugEvent, INFINITE)
```

to capture the debug event.

There are 9 debug event in total:

* CREATE_PROCESS_DEBUG_EVENT. Reports a create-process debugging event. 
* CREATE_THREAD_DEBUG_EVENT. Reports a create-thread debugging event.
* EXCEPTION_DEBUG_EVENT. Reports an exception debugging event.
* EXIT_PROCESS_DEBUG_EVENT. Reports an exit-process debugging event.
* EXIT_THREAD_DEBUG_EVENT. Reports an exit-thread debugging event.
* LOAD_DLL_DEBUG_EVENT. Reports a load-dynamic-link-library (DLL) debugging event.
* OUTPUT_DEBUG_STRING_EVENT. Reports an output-debugging-string debugging event.
* RIP_EVENT. Reports a RIP-debugging event (system debugging error).
* UNLOAD_DLL_DEBUG_EVENT. Reports an unload-DLL debugging event. 

If the debug event has been handled correctly, then

```c++
ContinueDebugEvent(debuggeeprocessID, debuggeethreadID, DBG_CONTINUE);
```

to continue the debuggee process. Let's combine the above to construct the debug loop:

```c++
while (WaitForDebugEvent(&debugEvent, INFINITE) == TRUE)
{
    debuggeeprocessID = debugEvent.dwProcessId;
    debuggeethreadID = debugEvent.dwThreadId;
    if (dispatchDebugEvent(debugEvent) == TRUE)
    {
        ContinueDebugEvent(debuggeeprocessID, debuggeethreadID, FLAG.continueStatus);
    }
    else {
        break;
    }
}

bool dispatchDebugEvent(const DEBUG_EVENT & debugEvent)
{
    switch (debugEvent.dwDebugEventCode)
    {
    case CREATE_PROCESS_DEBUG_EVENT:
        // TBD
        break;

    case CREATE_THREAD_DEBUG_EVENT:
        // TBD
        break;

    case EXCEPTION_DEBUG_EVENT:
        // TBD
        break;

    case EXIT_PROCESS_DEBUG_EVENT:
        // TBD
        break;

    case EXIT_THREAD_DEBUG_EVENT:
        // TBD
        break;

    case LOAD_DLL_DEBUG_EVENT:
        // TBD
        break;

    case UNLOAD_DLL_DEBUG_EVENT:
        // TBD
        break;

    case OUTPUT_DEBUG_STRING_EVENT:
        // TBD
        break;

    case RIP_EVENT:
        // TBD
        break;

    default:
        cout << "Unknown debug event." << endl;
        return false;
        break;
    }
}
```

In the next part of the series, I intend to give a brief introduction about the 9 debug events.