---
name: vingo-utils-dev
description: vingo-utils-v3 私人 Go 工具库开发规范。当任务涉及开发、修改、调试、扩展 vingo-utils-v3 仓库本身（module github.com/lgdzz/vingo-utils-v3），或基于该库编写下游 Go 项目代码、查找库内工具方法、添加新包/新函数时使用。涵盖包结构导览、统一响应契约、panic 式错误处理、模型规范、测试与发布习惯，并强制要求所有文件操作先经用户确认。
name_cn: vingo-utils 开发规范
description_cn: 针对 vingo-utils-v3 私人 Go 工具库的开发规范与常用工具速查，涵盖代码习惯、包结构、模型规范与发布流程，文件操作需用户确认。
create_source: super-agent-skill-creator
AIGC:
  ContentProducer: '001191110102MAD55U9H0F10002'
  ContentPropagator: '001191110102MAD55U9H0F10002'
  Label: '1'
  ProduceID: '87de7149-c4e6-4684-89d4-0f8a118dde44'
  PropagateID: '87de7149-c4e6-4684-89d4-0f8a118dde44'
  ReservedCode1: 'c50d3aea-5e06-44bf-bdde-73572a842138'
  ReservedCode2: 'c50d3aea-5e06-44bf-bdde-73572a842138'
---

# vingo-utils-v3 开发规范

## 铁律：文件操作必须先确认

对本项目（含 skill/ 目录自身）任何文件的**创建、修改、删除**操作，必须先向用户列出具体动作与目标路径，获得明确确认后才能执行。只读操作（读文件、搜索、go build、go vet、gofmt 检查类）可直接进行。

例外：`.temp/` 下的临时文件无需确认。

## 项目概况

- Module：`github.com/lgdzz/vingo-utils-v3`，Go 1.25.0，无 CI、无 `_test.go`，版本以裸 tag 管理（如 v1.1.84）
- 技术栈：Gin + GORM + go-redis(v6) + lancet v2 + gookit/slog + 自研异常库 `github.com/lgdzz/vingo-utils-exception`
- 依赖分层（自底向上）：
  - 地基层：`vingo`（仅依赖 cryptor，被 17 个包依赖）→ 定义 Context/响应契约/异常处理/工具函数
  - 无依赖工具层：`cryptor`、`redis`、`logs`、`kafka`、`region`、`email`、`moment`（moment 依赖 vingo）
  - 功能层：`db`（最重）、`ctype`、`cache`、`pool`、`queue`、`jwt`、`captcha`、`oss`、`request`、`file`、`wechat`、`onlyoffice`、`model`
  - 顶层：`router`（聚合一切，一行启动 Web 服务）、`cli`（命令行/交叉编译）、`extend/config`（下游配置模板）
- 新增代码时严禁制造循环依赖：任何包都不得反向依赖 `router`/`cli`/`extend`

## 硬性开发约定

### 1. 错误处理：panic 化，顶层统一 recover

- 业务错误：`panic("中文错误消息")`
- 数据库错误：`panic(&exception.DbException{Message: ...})`
- 鉴权失败：`panic(&exception.AuthException{Message: ...})`
- 全部由 `vingo.ExceptionHandler` 中间件 recover 后转统一响应（业务错误 status=200 error=1，AuthException status=401）
- 几乎不向调用方返回 error（仅 file/email/kafka 等纯 IO 包返回 error）
- 协程内必须兜底：`vingo.GoSafe(fn)` 或 `pool.BusinessHandle`，或开头 `defer vingo.ExceptionCatch("描述", false)`

### 2. 统一响应契约

所有 HTTP 出口经 `Context.Response(&ResponseData{...})`，结构固定：
`{uuid, error(0/1), message, s, data, timestamp}`。
路由注册用 `vingo.RoutesGet/RoutesPost/RoutesPut/RoutesPatch/RoutesDelete(g, path, handler)`，自动包装 Context 并写 API 操作日志。

### 3. 代码风格

- 注释全中文；新文件头部固定注释块（照抄库内既有文件）：

```go
// 作者: lgdz
// 创建时间: 2026/9/9
// 描述: xxx
```

- 导出函数必有中文单行注释；函数命名大驼峰；接受惯用短名（`Of`、`SY`、`Fast` 系列）
- 模型字段 gorm tag 用驼峰列名（`column:createdAt`）+ json tag 驼峰
- 优先使用已有依赖：切片/转换用 lancet v2，不要引入功能重复的新依赖；新增依赖需先告知用户
- 泛型优先（`Find[T]`、`Convert[T]`、`Strings[T]` 等风格）；组合优于继承（如 `db.Api` 嵌入 `*gorm.DB`）；策略/适配器模式（`db.Adapter`、`oss.Adapter`）
- 工厂函数统一 `NewXxx(config Config) *Api`，Config 结构提供指针字段 + 默认值填充方法

### 4. 模型规范

业务模型必须：
- 带 `Diff db.DiffBox[T]` 字段（`gorm:"-"`）用于变更追踪
- 实现 `TableName() string`
- 时间字段用 `*moment.LocalTime`（标准时间类型，JSON 格式 `2006-01-02 15:04:05`）
- 密码字段用 `ctype.Password`，手机号 `ctype.Phone`，敏感字段 `ctype.Ciphertext`
- 树形表带 `path/len` 字段，由钩子或 `db/pathutil` 维护
- 重要变更操作挂 `db.Api.ChangeLog` 回调写 `update_log` 表（参考 model.UpdateLog）

### 5. 测试与验证习惯

- 不写 `_test.go`；验证方式是 `test/` 目录下带 `main()` 的演示程序（`package main`）
- 注意 test/ 目录多 main 冲突：新演示文件用独立文件名，不与他人共用；如需临时禁用某演示，用注释方式切换，不删除他人文件
- 代码改动后验证：`go build ./...` + `go vet ./...` + `gofmt -l .`（输出应为空），用 `gofmt -w <file>` 格式化

### 6. 发布流程（build.sh）

- `./build.sh` 交互菜单：提交代码（commit 信息固定 `-`）→ 发布版本（输入版本号打裸 tag）→ 删除版本 → 查看最新 tag
- 交叉编译走 `cli.InitCli` 的 `-build-dev/-build-prod`，产物在 `output/`，ldflags 注入版本
- 提交/打 tag/推送属于对外操作，执行前必须用户确认

## 常用工具速查（详见 references/api-reference.md）

- 入参绑定：`vingo.GetRequestBody[T](c)` / `GetRequestQuery[T](c)`
- 查询：`db.Find[T]`、`db.FindById[T]`、`db.QueryList[T]`（分页/树/导出统一入口）、`db.Api.QueryWhereXxx` 链式构造
- 事务：`db.Api.FastCommit(func(tx))`（自动回滚）
- 缓存：`cache.Fast[T](redisApi, key, expired, func() T)`
- 并发：`pool.FastPool(n, func(p))`、`p.FastSubmit`
- 队列：`queue.Redis.Push/PushDelay/StartMonitor`（全局单例，`InitRedisQueue` 一次）
- 时间：`moment.ToLocalTime`、`moment.GetTodayRange` 等
- 类型：`ctype.Strings[T]`、`ctype.Money`、`ctype.Phone`、`ctype.IdCard`
- 完整函数签名与各包初始化约定，读 [references/api-reference.md](references/api-reference.md)

## 工作流程

1. 需求涉及库内工具时，先读 api-reference.md 定位已有实现，**优先复用，避免重复造轮子**
2. 改动前：向用户说明"改哪个文件、加什么内容/删什么内容"，确认后执行
3. 改动后：跑 `go build ./...`、`go vet ./...`、`gofmt -l .` 验证，结果如实汇报
4. 涉及发布（commit/push/tag）时单独确认