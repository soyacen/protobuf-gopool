# protobuf-go-pool

protobuf-go-pool 是一个 Protocol Buffer 编译器插件，用于为每个消息类型自动生成对象池（sync.Pool）相关的代码。这有助于减少垃圾回收压力，提高应用程序性能，特别是在高频率创建和销毁 protobuf 消息对象的场景下。

## 功能特性

- 自动为每个 protobuf 消息生成对象池
- 生成 Get、Put 方法用于池操作
- 复用消息对象以减少内存分配
- 与标准 protoc 工具链无缝集成

## 安装

```bash
go install github.com/soyacen/protobuf-gopool/cmd/protoc-gen-gopool@latest
```

## 使用方法

### 1. 基本用法

```bash
protoc --gopool_out=. --go_out=. your_proto_file.proto
```

此命令将为 `your_proto_file.proto` 中定义的每个消息类型生成两个文件：
- `your_proto_file.pb.go` - 标准的 Go protobuf 代码
- `your_proto_file.pb.pool.go` - 包含对象池相关代码

### 2. 示例

假设您有一个 `.proto` 文件定义了 `Person` 消息：

```protobuf
syntax = "proto3";

package example;

option go_package = "./example";

message Person {
  string name = 1;
  int32 age = 2;
}
```

插件将自动生成以下函数：

```go
// PersonPool is a sync.Pool for Person
var PersonPool = &sync.Pool{
  New: func() interface{} {
    return &example.Person{}
  },
}

// GetPerson gets a Person from the pool
func GetPerson() *example.Person {
  return PersonPool.Get().(*example.Person)
}

// PutPerson puts a Person back to the pool
func PutPerson(m *example.Person) {
  // Reset the message before putting it back to the pool
  m.Reset()
  PersonPool.Put(m)
}
```

### 3. 在您的代码中使用

```go
package main

import (
    "your-project/example"  // 导入生成的 protobuf 代码
)

func main() {
    // 从池中获取对象
    person := example.GetPerson()
    
    // 使用对象
    person.Name = "John Doe"
    person.Age = 30
    
    // ... 其他操作 ...
    
    // 将对象放回池中
    example.PutPerson(person)
}
```

## 优势

- **性能提升**: 通过复用对象减少内存分配和垃圾回收压力
- **简单易用**: 自动生成代码，无需手动管理对象池
- **零侵入性**: 不改变原始 protobuf 消息结构
- **线程安全**: 基于 sync.Pool 实现，线程安全

## 注意事项

- 在 Put 回池之前，请确保调用了 Reset() 方法清理对象状态
- 对象池适用于可复用的对象，对于包含非可重置资源的对象需谨慎使用
- 对象池只应在对象生命周期明确结束时使用

## 构建

如果您想从源码构建插件：

```bash
git clone https://github.com/soyacen/protobuf-gopool.git
cd protobuf-gopool
go build -o protoc-gen-gopool ./cmd/protoc-gen-gopool
```

## 许可证

MIT License