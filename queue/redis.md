### 初始化
```go
    // 在主进程中执行初始化
    queue.InitRedisQueue(queue.RedisQueueConfig{})
```

### 消费消息
```go

type Methods struct {}

func (s *Methods) Test(name string) {
    fmt.Println("执行Test方法")
    fmt.Println(name)
}

// 使用方法
queue.Redis.StartMonitor("test", &Methods{}) // 实时队列协程
queue.Redis.StartMonitorDelay("test") // 延迟队列协程

// 可以开启多个不同的消费主题队列协程

```

### 生产消息
```go
// 立即执行
queue.Redis.Push("test", queue.MessagePackage{
    Method: "Test",
    Params: map[string]any{"name": "张三"},
})
// 5秒后执行
queue.Redis.PushDelay("test", queue.MessagePackage{
    Method: "Test",
    Params: map[string]any{"name": "张三"},
}, 5)
```