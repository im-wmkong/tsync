# tsync

[![Go](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

`tsync` 是一个 Go 语言的同步工具库，提供了一系列基于 Go 标准库 `sync` 包的泛型包装器，使同步操作更加类型安全和方便使用。

## 功能特性

- **泛型支持**：所有组件都使用 Go 1.20+ 的泛型特性，提供类型安全的 API
- **扩展标准库**：在标准库 `sync` 的基础上提供更多便利功能
- **简洁易用**：简化同步操作的使用方式，减少样板代码
- **向后兼容**：与标准库 `sync` 包的 API 保持兼容，易于迁移

## 安装

```bash
go get github.com/im-wmkong/tsync
```

## 核心组件

### 1. AtomicValue
泛型包装的原子值，提供类型安全的原子操作。

```go
package main

import "github.com/im-wmkong/tsync"

func main() {
    // 创建一个新的原子值
    av := tsync.NewAtomicValue(42)
    
    // 加载值
    v := av.Load()
    
    // 存储新值
    av.Store(100)
    
    // 交换值
    old := av.Swap(200)
    
    // 比较并交换值
    swapped := av.CompareAndSwap(200, 300)
}
```

### 2. MutexValue
带互斥锁保护的值，提供安全的读写操作。

```go
package main

import "github.com/im-wmkong/tsync"

type Counter struct {
    value int
}

func main() {
    // 创建一个新的带互斥锁保护的值
    mv := tsync.NewMutexValue(Counter{value: 0})
    
    // 原子更新值
    mv.Lock(func(v *Counter) {
        v.value++
    })
    
    // 加载值
    counter := mv.Load()
}
```

### 3. RWMutexValue
带读写锁保护的值，支持并发读取和独占写入。

```go
package main

import "github.com/im-wmkong/tsync"

type Config struct {
    Host string
    Port int
}

func main() {
    // 创建一个新的带读写锁保护的值
    rwmv := tsync.NewRWMutexValue(Config{Host: "localhost", Port: 8080})
    
    // 读取值（共享锁）
    rwmv.RLock(func(v Config) {
        // 读取操作
    })
    
    // 更新值（独占锁）
    rwmv.Lock(func(v *Config) {
        v.Port = 9090
    })
}
```

### 4. Pool
泛型对象池，用于复用对象，减少内存分配。

```go
package main

import "github.com/im-wmkong/tsync"

func main() {
    // 创建一个新的对象池
    pool := tsync.NewPool(func() []byte {
        return make([]byte, 1024)
    })
    
    // 从池中获取对象
    buf := pool.Get()
    
    // 使用对象
    // ...
    
    // 将对象放回池中
    pool.Put(buf)
}
```

### 5. Map
泛型并发 Map，扩展了标准库的 `sync.Map`。

```go
package main

import "github.com/im-wmkong/tsync"

func main() {
    // 创建一个新的并发 Map
    m := &tsync.Map[string, int]{}
    
    // 存储键值对
    m.Store("key1", 42)
    
    // 加载值
    if v, ok := m.Load("key1"); ok {
        // 使用值
    }
    
    // 范围遍历
    m.Range(func(key string, value int) bool {
        // 处理键值对
        return true // 继续遍历
    })
    
    // 加载或初始化
    value, loaded := m.LoadOrInit("key2", func() int {
        return 100
    })
}
```

### 6. OnceValue
泛型的一次性初始化值，确保初始化函数只执行一次。

```go
package main

import "github.com/im-wmkong/tsync"

func main() {
    // 创建一个新的一次性初始化值
    once := tsync.NewOnceValue(func() string {
        // 初始化逻辑，只会执行一次
        return "initialized value"
    })
    
    // 获取值（首次调用会执行初始化函数）
    value := once.Get()
    
    // 再次获取值（直接返回已初始化的值）
    value = once.Get()
}
```

### 7. WaitGroup
增强版的等待组，支持上下文和 panic 恢复。

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/im-wmkong/tsync"
)

func main() {
    // 创建一个新的等待组
    wg := tsync.NewWaitGroup(
        tsync.WithPanicRecovery(func(p any) {
            fmt.Printf("Panic recovered: %v\n", p)
        }),
    )
    
    // 启动多个 goroutine
    for i := 0; i < 10; i++ {
        wg.Go(func() {
            // 执行任务
            time.Sleep(time.Millisecond * 100)
        })
    }
    
    // 启动带上下文的 goroutine
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    wg.GoCtx(ctx, func(ctx context.Context) {
        // 执行任务
        select {
        case <-time.After(time.Millisecond * 200):
        case <-ctx.Done():
            return
        }
    })
    
    // 等待所有任务完成
    wg.Wait()
}
```

### 8. Cond
泛型条件变量，支持上下文和谓词等待。

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/im-wmkong/tsync"
)

func main() {
    // 创建一个新的条件变量
    cond := tsync.NewCond()
    
    ready := false
    
    // 等待条件满足
    go func() {
        cond.WaitUntil(func() bool {
            return ready
        })
        fmt.Println("Condition met!")
    }()
    
    // 带超时的等待
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    go func() {
        err := cond.WaitUntilCtx(ctx, func() bool {
            return ready
        })
        if err != nil {
            fmt.Printf("Wait timed out: %v\n", err)
        }
    }()
    
    // 信号通知
    time.Sleep(time.Millisecond * 500)
    cond.Signal()
    
    // 广播通知
    ready = true
    cond.Broadcast()
}
```

## API 文档

### AtomicValue
- `NewAtomicValue(v T) *AtomicValue[T]` - 创建一个新的原子值
- `Load() T` - 加载当前值
- `Store(v T)` - 存储新值
- `Swap(v T) T` - 交换值并返回旧值
- `CompareAndSwap(old, new T) bool` - 比较并交换值

### MutexValue
- `NewMutexValue(v T) *MutexValue[T]` - 创建一个新的带互斥锁保护的值
- `Lock(fn func(v *T))` - 锁定并更新值
- `Load() T` - 加载当前值

### RWMutexValue
- `NewRWMutexValue(v T) *RWMutexValue[T]` - 创建一个新的带读写锁保护的值
- `RLock(fn func(v T))` - 读锁定并访问值
- `Lock(fn func(v *T))` - 写锁定并更新值

### Pool
- `NewPool(newFn func() T) *Pool[T]` - 创建一个新的对象池
- `Get() T` - 从池中获取对象
- `Put(v T)` - 将对象放回池中

### Map
- `Load(key K) (value V, ok bool)` - 加载键对应的值
- `Store(key K, value V)` - 存储键值对
- `LoadOrStore(key K, value V) (actual V, loaded bool)` - 加载或存储值
- `Delete(key K)` - 删除键值对
- `Range(fn func(key K, value V) bool)` - 范围遍历
- `MustLoad(key K) V` - 加载键对应的值（不存在则 panic）
- `LoadOrInit(key K, init func() V) (value V, loaded bool)` - 加载或初始化值

### OnceValue
- `NewOnceValue(fn func() T) *OnceValue[T]` - 创建一个新的一次性初始化值
- `Get() T` - 获取值（首次调用会执行初始化函数）

### WaitGroup
- `NewWaitGroup(opts ...WaitGroupOption) *WaitGroup` - 创建一个新的等待组
- `WithPanicRecovery(handler PanicHandler) WaitGroupOption` - 配置 panic 恢复处理
- `Go(f func())` - 启动一个 goroutine
- `GoCtx(ctx context.Context, f func(ctx context.Context))` - 启动一个带上下文的 goroutine
- `Wait()` - 等待所有 goroutine 完成

### Cond
- `NewCond() *Cond` - 创建一个新的条件变量
- `WaitUntil(predicate func() bool)` - 等待谓词条件满足
- `WaitUntilCtx(ctx context.Context, predicate func() bool) error` - 带上下文的谓词等待
- `Signal()` - 通知一个等待的 goroutine
- `Broadcast()` - 通知所有等待的 goroutine

## 许可证

本项目采用 MIT 许可证，详情请见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！
