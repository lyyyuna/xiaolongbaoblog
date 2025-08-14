title: Go JSON 进化：从 v1 到 v2
date: 2025-08-14 15:36:25
---

Go 1.25 带来了一个实验性的新 JSON 包 `encoding/json/v2`。它提供了改进的 API、更好的性能以及向后兼容的迁移路径。

要尝试新的 JSON 包，你需要：

1. Go 1.25
2. `GOEXPERIMENT=jsonv2` 环境变量

让我们探索主要功能和改进。

## 基本用法

基本的编码和解码操作看起来很熟悉：

```go
package main

import (
    "fmt"
    jsonv2 "encoding/json/v2"
)

type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    // 编码
    p := Person{Name: "Alice", Age: 30}
    data, err := jsonv2.Marshal(p)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(data)) // {"name":"Alice","age":30}

    // 解码
    var p2 Person
    err = jsonv2.Unmarshal(data, &p2)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", p2) // {Name:Alice Age:30}
}
```

## MarshalWrite 和 UnmarshalRead

新包引入了流式接口，允许直接向/从 `io.Writer` 和 `io.Reader` 进行编码/解码：

```go
import (
    "bytes"
    "strings"
    jsonv2 "encoding/json/v2"
)

func main() {
    p := Person{Name: "Bob", Age: 25}
    
    // 写入缓冲区
    var buf bytes.Buffer
    err := jsonv2.MarshalWrite(&buf, p, nil)
    if err != nil {
        panic(err)
    }
    
    // 从 reader 读取
    reader := strings.NewReader(buf.String())
    var p2 Person
    err = jsonv2.UnmarshalRead(reader, &p2, nil)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("%+v\n", p2) // {Name:Bob Age:25}
}
```

## MarshalEncode 和 UnmarshalDecode

对于更复杂的场景，新包提供了 `Encoder` 和 `Decoder` 类型：

```go
import (
    "bytes"
    "strings"
    jsonv2 "encoding/json/v2"
)

func main() {
    // 使用编码器
    var buf bytes.Buffer
    enc := jsonv2.NewEncoder(&buf)
    
    people := []Person{
        {Name: "Alice", Age: 30},
        {Name: "Bob", Age: 25},
    }
    
    for _, p := range people {
        err := enc.Encode(p)
        if err != nil {
            panic(err)
        }
    }
    
    // 使用解码器
    dec := jsonv2.NewDecoder(strings.NewReader(buf.String()))
    
    for {
        var p Person
        err := dec.Decode(&p)
        if err != nil {
            break // EOF 或其他错误
        }
        fmt.Printf("%+v\n", p)
    }
}
```

## 选项

新包引入了灵活的选项系统来自定义编码/解码行为：

```go
import jsonv2 "encoding/json/v2"

func main() {
    p := Person{Name: "Charlie", Age: 35}
    
    // 使用选项进行美观打印
    data, err := jsonv2.Marshal(p, jsonv2.WithIndent("", "  "))
    if err != nil {
        panic(err)
    }
    
    fmt.Println(string(data))
    // {
    //   "name": "Charlie",
    //   "age": 35
    // }
    
    // 严格解码 - 拒绝未知字段
    jsonStr := `{"name":"David","age":40,"unknown":"field"}`
    var p2 Person
    err = jsonv2.Unmarshal([]byte(jsonStr), &p2, jsonv2.WithRejectUnknownMembers(true))
    if err != nil {
        fmt.Printf("错误: %v\n", err) // 错误: 未知字段 "unknown"
    }
}
```

## 标签

v2 包增强了对结构体标签的支持：

```go
type Product struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price,omitempty"`
    Tags  []string `json:"tags,omitzero"`
}

func main() {
    // omitempty: 省略空值
    p1 := Product{ID: 1, Name: "Widget"}
    data1, _ := jsonv2.Marshal(p1)
    fmt.Println(string(data1)) // {"id":1,"name":"Widget"}
    
    // omitzero: 省略零值
    p2 := Product{ID: 2, Name: "Gadget", Tags: []string{}}
    data2, _ := jsonv2.Marshal(p2)
    fmt.Println(string(data2)) // {"id":2,"name":"Gadget"}
}
```

## 自定义编组

新包为自定义编组提供了改进的接口：

```go
import (
    "time"
    jsonv2 "encoding/json/v2"
)

type CustomTime struct {
    time.Time
}

func (ct CustomTime) MarshalJSONV2(enc *jsonv2.Encoder, opts jsonv2.Options) error {
    return enc.WriteString(ct.Format("2006-01-02"))
}

func (ct *CustomTime) UnmarshalJSONV2(dec *jsonv2.Decoder, opts jsonv2.Options) error {
    str, err := dec.ReadString()
    if err != nil {
        return err
    }
    
    t, err := time.Parse("2006-01-02", str)
    if err != nil {
        return err
    }
    
    ct.Time = t
    return nil
}

type Event struct {
    Name string     `json:"name"`
    Date CustomTime `json:"date"`
}

func main() {
    e := Event{
        Name: "Meeting",
        Date: CustomTime{time.Date(2025, 6, 22, 0, 0, 0, 0, time.UTC)},
    }
    
    data, err := jsonv2.Marshal(e)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(string(data)) // {"name":"Meeting","date":"2025-06-22"}
}
```

## 默认行为

v2 包改变了一些默认行为：

- 字符串值中的 HTML 字符默认不再被转义
- 更严格的数字解析
- 改进的错误消息，提供更好的上下文

```go
type Config struct {
    HTML string `json:"html"`
}

func main() {
    c := Config{HTML: "<script>alert('hello')</script>"}
    
    // v1 会转义 < > &
    // v2 默认不转义
    data, _ := jsonv2.Marshal(c)
    fmt.Println(string(data)) // {"html":"<script>alert('hello')</script>"}
    
    // 如果需要转义，使用选项
    dataEscaped, _ := jsonv2.Marshal(c, jsonv2.WithEscapeHTML(true))
    fmt.Println(string(dataEscaped)) // {"html":"\u003cscript\u003ealert('hello')\u003c/script\u003e"}
}
```

## 性能

新包提供了显著的性能改进，特别是在反编组操作方面：

- 反编组速度提升高达 2-3 倍
- 更少的内存分配
- 更好的缓存局部性
- 改进的大型数据集处理

## 迁移

从 v1 迁移到 v2 通常很简单：

1. 将导入从 `"encoding/json"` 更改为 `jsonv2 "encoding/json/v2"`
2. 如果有的话，更新自定义编组接口
3. 测试并调整任何依赖于更改的默认行为的代码

新包设计为大多数现有代码的直接替代品，只需要最少的更改。

## 结论

Go 的新 JSON v2 包代表了 JSON 处理的重大改进，提供了：

- 更好的性能
- 更灵活的 API
- 改进的流式支持
- 增强的自定义选项
- 向后兼容的迁移路径

虽然仍然是实验性的，但 JSON v2 有望成为 Go 生态系统中 JSON 处理的新标准。随着 Go 1.25 发布，现在是探索这些新功能并为迁移做准备的绝佳时机。
