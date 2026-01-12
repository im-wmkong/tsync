# tsync

tsync 是一个**基于 Go 泛型（Go 1.20+）的类型安全并发工具库**，对标准库 `sync` / `sync/atomic` 中以 `any` 作为类型边界的 API 进行**最小且克制的封装**，在**不改变并发语义、不引入额外抽象**的前提下，将类型错误前移到编译期。

> 设计原则  
> 不发明新的并发模型，只让现有模型更安全、更可读、更难被误用。

---

## 核心特性

- **类型安全**
  - 所有对外 API 均为泛型接口
  - 不向调用方暴露 `any`

- **语义对齐**
  - 行为与 `sync` / `sync/atomic` 官方文档严格一致
  - 不引入额外的并发语义或隐藏行为

- **克制设计**
  - 只解决“类型边界”问题
  - 不尝试替代 channel 或更高层并发抽象

- **工程友好**
  - 零第三方依赖
  - 零值语义与标准库保持一致
  - 全量单元测试，覆盖并发与泛型正确性

---

## 安装

```bash
go get github.com/im-wmkong/tsync
```

环境要求：
- Go 1.20 或以上
- 标准库 `sync` / `sync/atomic`

---

## 为什么需要 tsync

在 Go 标准库中，以下类型为了通用性大量使用 `any`：
- `sync.Map`
- `sync.Pool`
- `sync.Once`
- `atomic.Value`
- `atomic.Pointer`

这类 API 在工程实践中通常会带来：
- 类型错误只能在运行时发现
- 频繁的类型断言与防御性代码
- IDE 自动补全与静态分析能力下降
- Code Review 成本增加

tsync 的目标非常明确：
- **不改变任何并发行为**
- **不增加任何运行时成本**
- **仅通过泛型约束，消除类型断言**

---

## API 概览

### Map

```go
type Map[K comparable, V any] struct
```

提供与 `sync.Map` 对齐的能力，并增强了类型安全性：
- `Load`
- `Store`
- `LoadOrStore`
- `LoadOrInit`
- `Delete`
- `Range`
- `MustLoad`（扩展方法，不存在时 panic）

示例：
```go
var m tsync.Map[string, int]
m.Store("a", 1)
m.Store("b", 2)

// 安全加载，无需类型断言
v, ok := m.Load("a") // v 类型为 int

// 范围遍历，类型安全
m.Range(func(key string, value int) bool {
    fmt.Printf("key: %s, value: %d\n", key, value)
    return true // 继续遍历
})

// 加载或初始化
v, loaded := m.LoadOrInit("c", func() int {
    return expensiveComputation()
})

// 必须加载，不存在时 panic
v := m.MustLoad("a")
```

---

### Pool

```go
type Pool[T any] struct
```

类型安全的对象池，与 `sync.Pool` 语义完全一致：

示例：
```go
import "bytes"

// 创建一个 bytes.Buffer 对象池
p := tsync.NewPool(func() *bytes.Buffer {
    return &bytes.Buffer{}
})

// 获取对象
buf := p.Get()
defer p.Put(buf)

// 使用对象（类型安全，无需断言）
buf.WriteString("hello, tsync")
fmt.Println(buf.String())
```

---

### OnceValue

```go
type OnceValue[T any] struct
```

类型安全的 `sync.Once` 封装，确保初始化函数只执行一次：

示例：
```go
// 创建 OnceValue，传入初始化函数
ov := tsync.NewOnceValue(func() *Config {
    // 这部分代码只会执行一次
    return loadConfigFromFile()
})

// 获取值，线程安全
cfg := ov.Get()
// 再次调用不会重新初始化
cfg = ov.Get()
```

---

### MutexValue

```go
type MutexValue[T any] struct
```

封装了 `sync.Mutex` 的值容器，强制在锁保护范围内修改值：

示例：
```go
// 创建线程安全的计数器
counter := tsync.NewMutexValue(0)

// 在锁保护下安全修改值
counter.Lock(func(v *int) {
    *v++ // 安全修改
})

// 安全读取值
current := counter.Load()
```

---

### RWMutexValue

```go
type RWMutexValue[T any] struct
```

封装了 `sync.RWMutex` 的值容器，支持多读单写：

示例：
```go
// 创建线程安全的共享状态
state := tsync.NewRWMutexValue(MyState{
    Version: 1,
    Data:    "initial",
})

// 读取锁 - 并发安全的只读访问
state.RLock(func(v MyState) {
    fmt.Printf("Version: %d, Data: %s\n", v.Version, v.Data)
})

// 写入锁 - 安全修改
state.Lock(func(v *MyState) {
    v.Version++
    v.Data = "updated"
})
```

---

### WaitGroup

```go
type WaitGroup struct
```

增强版 `sync.WaitGroup`，增加了 panic 恢复和上下文支持：

```go
import "context"

// 创建带 panic 恢复的 WaitGroup
wg := tsync.NewWaitGroup(
    tsync.WithPanicRecovery(func(p any) {
        log.Printf("捕获到 panic: %v", p)
    }),
)

// 启动多个任务
for i := 0; i < 10; i++ {
    wg.Go(func() {
        performTask()
    })
}

// 启动带上下文的任务
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

wg.GoCtx(ctx, func(ctx context.Context) {
    // 任务将在上下文取消时自动结束
    performTaskWithContext(ctx)
})

// 等待所有任务完成
wg.Wait()
```

---

### AtomicPointer

```go
type AtomicPointer[T any] struct
```

类型安全的 `atomic.Pointer` 封装：

示例：
```go
var p tsync.AtomicPointer[string]

// 存储指针
v := "hello"
p.Store(&v)

// 加载指针
if ptr := p.Load(); ptr != nil {
    fmt.Println(*ptr)
}

// CAS 操作
newV := "world"
if p.CompareAndSwap(&v, &newV) {
    fmt.Println("CAS 成功")
}
```

---

### AtomicValue

```go
type AtomicValue[T any] struct
```

类型安全的 `atomic.Value` 封装，确保 Store/Load 类型一致：

示例：
```go
// 创建原子值
av := tsync.NewAtomicValue(0)

// 原子存储
av.Store(10)

// 原子加载
v := av.Load() // v 类型为 int
```

---

## 设计哲学

### tsync 明确不做的事情
- 不替代 channel
- 不实现 future / promise
- 不引入 actor / scheduler
- 不自动规避 data race

### tsync 坚持的原则
1. **语义一致性**：与标准库行为完全一致
2. **最小侵入性**：只在类型边界提供价值
3. **零成本抽象**：不增加任何运行时开销
4. **显式行为**：不引入隐式并发语义

---

## 测试策略

tsync 采用严格的测试策略，确保：

- **并发正确性**：测试所有并发场景下的行为一致性
- **泛型安全性**：验证泛型约束的正确性
- **语义对齐**：与标准库 `sync` / `sync/atomic` 行为对比测试
- **性能一致性**：确保无额外性能开销

运行测试：
```bash
go test ./...
```

---

## License

MIT License

---

## 作者说明

tsync 并不是为了“比 sync 更聪明”，而是为了让并发代码在工程实践中：
- 更安全：编译期类型检查
- 更易读：消除冗余的类型断言
- 更容易维护：IDE 支持与静态分析增强

如果你在使用它时几乎感受不到它的存在，那说明它完成了自己的设计目标。

---

## 贡献

欢迎提交 Issue 和 Pull Request！

贡献指南：
1. 保持与标准库的语义一致性
2. 遵循克制设计原则
3. 提供完整的单元测试
4. 更新相关文档