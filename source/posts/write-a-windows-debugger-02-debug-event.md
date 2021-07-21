title: Writing Your Own Windows Debugger - Debug Event
date: 2017-05-01 21:36:25
categories: 系统
tags: 
- Windows Debugger
---


We have introduced the debug loop last time, in this post, I will talk about various debug events.

### RIP_EVENT

I find very few documents about this event, only mentioned with words like _system error_ or _internal error_. So I decide to print a error message and skip it. As my project is not fully tested, I have never 
encountered such a situation.

### OUTPUT_DEBUG_STRING_EVENT

When the debuggee calls the _OutpuDebugString_ function, it will raise this debug event. The following structure describes the detail of this event:

```c++
typedef struct _OUTPUT_DEBUG_STRING_INFO {
  LPSTR lpDebugStringData;
  WORD  fUnicode;
  WORD  nDebugStringLength;
} OUTPUT_DEBUG_STRING_INFO, *LPOUTPUT_DEBUG_STRING_INFO;
```

* lpDebugStringData, The debugging string in the calling process's address space.
* fUnicode, The format of the debugging string. If this member is zero, the debugging string is ANSI; if it is nonzero, the string is Unicode.
* nDebugStringLength, The size of the debugging string, in characters. The length includes the string's terminating null character.

With ReadProcessMemory function, the debugger can obtain the value of the string:

```c++
void OnOutputDebugString(const OUTPUT_DEBUG_STRING_INFO* pInfo) 
{
    BYTE* pBuffer = (BYTE*)malloc(pInfo->nDebugStringLength);

    SIZE_T bytesRead;

    ReadProcessMemory(
        g_hProcess,
        pInfo->lpDebugStringData,
        pBuffer, 
        pInfo->nDebugStringLength,
        &bytesRead);

    int requireLen = MultiByteToWideChar(
        CP_ACP,
        MB_PRECOMPOSED,
        (LPCSTR)pBuffer,
        pInfo->nDebugStringLength,
        NULL,
        0);

    TCHAR* pWideStr = (TCHAR*)malloc(requireLen * sizeof(TCHAR));

    MultiByteToWideChar(
        CP_ACP,
        MB_PRECOMPOSED,
        (LPCSTR)pBuffer,
        pInfo->nDebugStringLength,
        pWideStr,
        requireLen);

    std::wcout << TEXT("Debuggee debug string: ") << pWideStr <<  std::endl;

    free(pWideStr);
    free(pBuffer);
}
```

### LOAD_DLL_DEBUG_EVENT

After the debuggee loads a dll, this debug event will be triggered. The following structure describes the detail of this event:

```c++
typedef struct _LOAD_DLL_DEBUG_INFO {
  HANDLE hFile;
  LPVOID lpBaseOfDll;
  DWORD  dwDebugInfoFileOffset;
  DWORD  nDebugInfoSize;
  LPVOID lpImageName;
  WORD   fUnicode;
} LOAD_DLL_DEBUG_INFO, *LPLOAD_DLL_DEBUG_INFO;
```

You may want to use the member _lpImageName_ to retrieve the dll file name, however, it doesn't work. According the explaination on MSDN, this member is pointer to the file name of the associated _hFile_, it  may, in turn, either be NULL or point to the actual filename. Even it is not NULL, ReadProcessMemory may also return a NULL. As a result, this membor is not reliable.

It seems that there is no direct Windows API to get the filename from the file handle. Someone has tried [this way](http://blog.csdn.net/bodybo/archive/2006/08/28/1131346.aspx).

### UNLOAD_DLL_DEBUG_EVENT

When a dll module is unloaded, this event will be triggered, nothing needs handled, just skip it.

### CREATE_PROCESS_DEBUG_EVENT

After the process is created, this is the first debug event. The following structure describes the detail of this event:

```c++
typedef struct _CREATE_PROCESS_DEBUG_INFO {
  HANDLE                 hFile;
  HANDLE                 hProcess;
  HANDLE                 hThread;
  LPVOID                 lpBaseOfImage;
  DWORD                  dwDebugInfoFileOffset;
  DWORD                  nDebugInfoSize;
  LPVOID                 lpThreadLocalBase;
  LPTHREAD_START_ROUTINE lpStartAddress;
  LPVOID                 lpImageName;
  WORD                   fUnicode;
} CREATE_PROCESS_DEBUG_INFO, *LPCREATE_PROCESS_DEBUG_INFO;
```

We can use this structure to get the symbols of the debuggee program.

### EXIT_PROCESS_DEBUG_EVENT

When debuggee process exits, this event will be triggered. The following structure describe the detail of the event:

```c++
typedef struct _EXIT_PROCESS_DEBUG_INFO {
  DWORD dwExitCode;
} EXIT_PROCESS_DEBUG_INFO, *LPEXIT_PROCESS_DEBUG_INFO;
```

What we can do is to print the exit code.

### CREATE_THREAD_DEBUG_EVENT

It is similar to the process create debug event.

### EXIT_THREAD_DEBUG_EVENT

It is similar to the process exit debug event.

### EXCEPTION_DEBUG_EVENT

It is the most important event of our debugger, I will cover it in the next post.