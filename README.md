# vingo-utils-v3

`vingo-utils-v3` 是一个基于 Go 语言开发的工具库，提供了数据库操作、Redis 缓存、消息队列、协程池、数据处理等多种实用功能，旨在帮助开发者提高开发效率。


## 主要功能模块

### 数据库操作
- **支持多种数据库**：提供 MySQL 和 PostgreSQL 的连接池创建功能，支持自定义配置。
- **数据库字典生成**：可以生成数据库的 HTML 格式数据字典，方便开发和维护。
- **事务管理**：提供快捷的事务处理方法。

### Redis 操作
- **基础操作**：支持 `Get`、`Del` 等基础操作。
- **缓存序列化**：自动处理数据的 JSON 序列化和反序列化。

### 消息队列
- **延迟任务**：支持推送延迟任务，定时执行。
- **消息监听**：提供消息监听和消费功能，支持异常重试。

### 协程池
- **快速创建**：提供快速创建协程池的方法，支持带上下文的协程池。
- **任务提交**：支持快速提交协程任务，可对任务结果进行排序。

### 数据处理
- **自定义数据类型**：提供 `Money`、`IdCard`、`Ciphertext` 等自定义数据类型，方便数据处理。
- **金额转换**：支持金额转大写中文和格式化显示。
- **身份证验证**：可以验证身份证号码的有效性。

## 安装
```bash
go get github.com/lgdzz/vingo-utils-v3
