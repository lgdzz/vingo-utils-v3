---
AIGC:
  ContentProducer: '001191110102MAD55U9H0F10002'
  ContentPropagator: '001191110102MAD55U9H0F10002'
  Label: '1'
  ProduceID: '5fa21998-ed69-4dbe-abbc-9f71487a2a22'
  PropagateID: '5fa21998-ed69-4dbe-abbc-9f71487a2a22'
  ReservedCode1: '2d132f1e-377d-494c-8938-8cb034f22c6d'
  ReservedCode2: '2d132f1e-377d-494c-8938-8cb034f22c6d'
---

# vingo-utils-v3 API 速查

按包分节的常用导出函数速查。签名摘自源码，以库内实际代码为准；使用前如有疑问直接读对应包源码确认。

**目录**
- [vingo（地基包）](#vingo)
- [db（数据库）](#db)
- [redis](#redis)
- [router（Web 启动）](#router)
- [logs（结构化日志）](#logs)
- [ctype（自定义类型）](#ctype)
- [moment（时间）](#moment)
- [cache](#cache)
- [pool（协程池）](#pool)
- [queue（Redis 队列）](#queue)
- [kafka](#kafka)
- [oss（对象存储）](#oss)
- [request（HTTP 客户端）](#request)
- [jwt](#jwt)
- [captcha（验证码）](#captcha)
- [cryptor（加密）](#cryptor)
- [file](#file)
- [email](#email)
- [model（通用业务模型）](#model)
- [region（行政区划）](#region)
- [wechat（微信小程序）](#wechat)
- [onlyoffice](#onlyoffice)
- [cli（命令行）](#cli)
- [extend/config（下游配置模板）](#extendconfig)

## vingo

地基包：Context 封装、统一响应、异常处理、工具函数、树结构。

```go
// 上下文与入参
type Context struct{ *gin.Context }
func GetRequestBody[T any](c *Context, valid ...bool) T   // 泛型绑定 JSON body，valid=true 时跑 validator 校验
func GetRequestQuery[T any](c *Context) T                 // 泛型绑定 query

// 统一响应（唯一出口）
func (c *Context) Response(d *ResponseData)               // {uuid, error, message, s, data, timestamp}
func (c *Context) ResponseBody(data any)
func (c *Context) ResponseSuccess(data ...any)
func (c *Context) ResponseBase64(data ...any)
func (c *Context) ResponseExchangeEncode(data any)        // 数据经 cryptor.ExchangeEncode 加密

// 路由注册（自动包 Context、自动写 API 操作日志）
func RoutesGet/RoutesPost/RoutesPut/RoutesPatch/RoutesDelete(g *gin.RouterGroup, path string, handler func(*Context))

// 用户身份（20+ 对 Get/Set）
func (c *Context) SetUserId(v int) / GetUserId() int      // 另有 SetAccId/GetOrgId/GetOrgPid/GetDeptIds/GetRoleIds/GetDataScope 等
func (c *Context) VerifyRoleTags(tags ...string) bool     // 另有 VerifyOrgTypes / VerifyRoleIds

// 数据权限四级分发
type DataScope struct{ MaxLevel, OrgLevel, DeptLevel, AccLevel func() }
func (s *DataScope) Handle(c *Context)

// 异常处理（顶层）
func ExceptionHandler(c *gin.Context)                     // 全局 recover 中间件，按 panic 类型转统一响应
func ExceptionCatch(s string, emit bool)                  // 协程内 recover + 日志
func ShieldRobots(r *gin.Engine)                          // robots.txt 屏蔽爬虫
func AllowCrossDomain(r *gin.Engine)                      // 跨域
func ApiAddress(port int)

// 工具函数
func Of[T any](v T) *T                                    // 取指针（高频）
func SY[T any](condition bool, trueValue, falseValue T) T // 三元
func GoSafe(fn func())                                    // 带 recover 的协程
func Convert[T any](input any) T                          // JSON 方式结构互转
func JsonToString(data any) string / JsonToStringRaw(v any) string  // Raw 不转义 HTML
func StringToJson(data string, output any)
func ToInt/ToInt64/ToUint/ToFloat/ToBool/ToString/ToBase64(value any)
func ToMap[T any, K comparable, V any](array []T, iteratee func(T) (K, V)) map[K]V
func GetUUID() string
func RandomString(length int) string / RandomNumber(length int) string
func OrderNo(length int, check func(string) bool) string            // 时间+随机数单号，check 查重回调
func OrderNoPrefix(prefix string, length int, check func(string) bool) string
func DiffSlice[T comparable](oldSlice, newSlice []T) []T
func DiffByFunc[T any, K comparable](oldList, newList []T, keyFunc func(T) K) []K
func FormatBytes(size int64, precision int) string
func Pinyin(text string) string / PinyinInitial(text string) string
func SafeDivision(left, right float64) float64
func ComputeGrowRate(now, prev float64) string
func NextVersion(v string) string
func Print(content any)

// 树结构
func FastTree[T float64 | string](rows any) []map[string]any          // 一行生成树（自动 hasChild/childCount/totalCount）
func FindNodeFromTree(tree []map[string]any, target any, keys ...string) map[string]any

// 耗时监听
func NewCost() *Cost
func (c *Cost) Mark(name string)                          // 受 vingo.CostDebug 控制，绿/黄/红三色分步耗时

// 其他
func Img2Pdf(imgBase64 []string) string                   // A4 一图一页，返回 base64 PDF
var Valid *validator.Validate                             // init() 自动创建
var GinDebug bool                                         // 异常时是否打印堆栈
var CT_JSON/CT_PNG/CT_DOCX ...                            // 60+ ContentType 常量（content.type.go）
```

## db

GORM 封装：多驱动、泛型查询/分页、Diff 变更追踪、字典与模型生成。子包 `db/book`（HTML 字典）、`db/model`（模型模板）、`db/pathutil`（树表 path 维护）。

```go
// 初始化（工厂入口）
type Config struct{ Host, Port, Dbname, Schema, Username, Password, Charset string;
    ConnectTimeout, MaxIdleConns, MaxOpenConns int; Driver string;   // "mysql"(默认)/"pgsql"/"sqlite"
    Secret string; Debug bool; InitAfter func(tx *gorm.DB) }
func NewDatabase(config Config) *Api                       // 同时设置 ctype.Secret、注册异常/Diff 回调插件

type Api struct{ *gorm.DB; *Common; Adapter; Config Config; ChangeLog func(tx *gorm.DB, option ChangeLogOption) }

// 查询（查不到直接 panic）
func Find[T any](db *gorm.DB, condition ...any) T
func FindWithDiff[T any](db *gorm.DB, condition ...any) T  // 返回带 DiffBox 旧值
func FindById[T any](db *gorm.DB, id int) T
func QueryList[T any](db *gorm.DB, pq PageQuery, option *QueryListOption[T]) any
// PageQuery{Limit, Order, OrderRaw, Keyword, LikeColumn, LikeValue, LikeWhitelist}
// Page 为 nil 不分页；支持 Iteratee 映射、协程池映射、树结构；导出超 1000 行走 Rows 流式

// 分页
func NewPage[T any](option QueryOption[T]) PageResult     // ExportSizeThreshold=1000

// 事务
func (s *Api) FastCommit(handler func(tx *gorm.DB))       // 快捷事务 + 自动回滚
func (s *Api) AutoCommit(tx *gorm.DB, callback ...func())

// Diff 变更追踪（模型需带 Diff db.DiffBox[T] 字段，gorm:"-"）
type DiffBox[T any] struct{ Old, New *T; Result *map[string]DiffItem }
func (s *DiffBox[T]) Compare()
func (s *DiffBox[T]) IsChange(column string) bool
func (s *DiffBox[T]) IsModify(column string, callback func())   // 变更才执行回调
func (s *DiffBox[T]) IsChangeOr/IsChangeAnd(...)
func (s *DiffBox[T]) ResultContent() string               // 中文摘要"将X的值[a]变更为[b]"
func (s *DiffBox[T]) ResultJson() *ChangeItems
func (s *DiffBox[T]) HasChange() bool
func (s *DiffBox[T]) DiffLog(tx *gorm.DB, result string)

// 查询构造器（挂在 *Api 上，链式）
func (s *Api) QueryWhere(model any, query string, value ...any) *gorm.DB
func (s *Api) QueryWhereIn/QueryWhereNotIn/QueryWhereLike/QueryWhereLikeRight(...)
func (s *Api) QueryWhereDate/QueryWhereBetween/QueryWherePath/QueryWhereExists(...)
func (s *Api) Exists(model any, condition ...any) bool
func (s *Api) NotExistsErr/NotExistsErrMsg(...)
func (s *Api) CheckHasChild(model any, id int) bool
func (s *Api) Operator(ctx any) *gorm.DB / OperatorWithTx(...)   // 注入操作人
func (s *Api) Diff() *gorm.DB

// 方言适配器（MysqlAdapter / PgsqlAdapter 双实现）
type Adapter interface {
    GetDatabases() []DatabaseInfo
    GetTables() []TableInfo
    GetColumns(tableName string) []Column
    GetTableDDL(...) ...
    Book() string                                          // HTML 数据字典
    ModelFiles(tableNames ...string) (bool, error)         // 生成 model 文件
    ...
}

// 输入类型
type TextSlice string        // "a,b,c" 形态，ToSlice/ToIntSlice/ToStringSlice/IsEmpty
type Between[T any] string
type Id/UUID/Keyword/DetailInput/UpdateFieldInput struct

// 子包 db/pathutil（树表 path/len 维护）
func SetPathWithCreate[T any](model *T, parent *T, option *Option)
func SetPathWithUpdate[T any](model *T, option Option)
```

初始化约定：`db.NewDatabase(Config{...})` → `*Api`；用密文字段（ctype.Ciphertext/Password）必须传 `Secret`。

## redis

```go
func NewRedis(config Config) *Api    // Config{Host, Port, Select, Password, PoolSize, MinIdleConns, Prefix}
func (s *Api) Get(key string, value any) (exist bool)   // miss 不报错返回 false，自动 JSON 反序列化
func (s *Api) Set(key string, value any, expiration time.Duration) string
func (s *Api) HSet(key, field string, value any) / HGet(key, field string, value any)
func (s *Api) Del(key ...string) int64
// 所有 key 自动加 Config.Prefix 前缀
```

## router

一行启动完整 Gin 服务。

```go
type Hook struct {
    Option        HookOption      // Name/Port/Copyright/Debug/Database *db.Api/Redis *redis.Config/startTime
    RegisterRouter func(r *gin.Engine)
    BaseMiddle    func(c *gin.Context)
    LoadWeb       []WebItem       // go-bindata-assetfs 静态资源
    AllowMethods  map[string]struct{}
}
func InitRouter(hook *Hook)       // 应用入口：异常中间件+BaseMiddle+404 统一响应+控制台+SSL 路由+启动横幅+r.Run
func BaseMiddle(hook *Hook) gin.HandlerFunc   // UUID/开始时间注入、方法白名单、请求日志（logs.Request）
func Console(c *gin.Context, option HookOption)
func FormatDuration(start, end time.Time) string
// SSL：/ssl.input、/ssl.deploy 路由（写 pem + 执行重启命令）
```

## logs

gookit/slog + rotatefile 新版日志，与 router 集成。

```go
func Init(filePath ...string)          // sync.Once，默认 runtime/logs/app.log，每日轮转、100M 上限、保留 15 天
func Info/Debug/Warn/Error/Fatal/Request/Response(args ...any)
func SetConsoleEnabled(enabled bool)   // 运行时开关控制台输出
func SetBuffSize(size int) / SetMaxSize(size uint64) / SetBackupTime(day uint) / Close()
```

## ctype

20+ 带语义的自定义字段类型，全部实现 driver.Valuer + json.Marshaler/Unmarshaler。

```go
type Ciphertext string                 // AES-GCM 透明加解密，依赖包级 var Secret（db.NewDatabase 设置）
type Password Ciphertext               // md5+salt+强度校验+Totp
func NewPassword(raw string, level int, isTemp bool) Password
func (s *Password) Match(raw string) bool / IsExpired(day int) bool
func (s *Password) EnableTotp(issuer, accountName string) / MatchTotp(code string, skew ...uint) bool
type Phone string                      // ^1[3-9]\d{9}$ 校验，Mask() 脱敏，Carrier() 运营商
type IdCard string                     // Analysis()/Age()/Birthday()/IsValid()
type Money float64                     // Round()/Format() 千分位/ToChinese() 大写金额
type Moneys string                     // "1,100" 范围
type Bool bool / Int int / Float float64 / String string / Text string / IP string
type Strings[T any] []T                // 入库 "1,2,3"，出库 JSON 数组
type Jsons[T any] []T                  // JSON 数组字段
type Path[T PathInterface] string      // "1,2,3" 路径，Slice/Len/Contains/First/Last/Parents
type Ratio float64                     // 比例，入库 ×100
type LoginLog                          // Modify(ip)/IsActive()，var ActiveLoginDay = 7
```

## moment

```go
type LocalTime time.Time               // 全项目标准时间字段，JSON 格式 2006-01-02 15:04:05
func ToLocalTime[T time.Time | ~string](value T) *LocalTime
func NowLocalTime()/TodayLocalTime()/YesterdayLocalTime() LocalTime
type DateRange struct{...} / DateText / DateTextRange
func GetTodayRange/GetYesterdayRange/GetMonthRange/GetQuarterRange/GetYearRange(...) DateRange
func BeforeNDates(n int, endDay ...time.Time) []string / AfterNDates(...)
func GenerateDates/GenerateMonths(ts any) []string
func DiffDaysFromNow(t time.Time) int
// 常量：YearFormat/YearMonthFormat/DateFormat/DateTimeFormat、DaySec=86400
```

## cache

```go
func Fast[T any](redisApi *redis.Api, key string, expired time.Duration, handle func() T) T  // read-through
func FastRefresh[T any](redisApi *redis.Api, key string, expired time.Duration, handle func() T) T
func ExpiredToday() time.Duration       // 至当日 23:59:59
func ExpiredWeekEnd() time.Duration     // 至本周末 23:59:59
func ExpiredMomentEnd() time.Duration   // 至月末 23:59:59
```

## pool

```go
func NewGoroutinePool(ctx context.Context, maxWorkers int) *GoroutinePool
func (s *GoroutinePool) Run() / Submit(task TaskFunc) / Cancel() / CloseAndWait() []Result
func (s *GoroutinePool) FastSubmit(handle func() any, index ...int)
func (s *GoroutinePool) FastSubmitLock(handle func() any, setData func(any), index ...int)  // 带锁回写
func FastPool(maxWorkers int, handle func(p *GoroutinePool)) []Result
func FastPoolWithContext(ctx context.Context, maxWorkers int, handle func(p *GoroutinePool)) []Result
func BusinessHandle[T any](data T, index int, handle func(object T) any) Result  // panic 自动捕获转 Result
type Result struct{ Index int; Data, Result any; Error *string }  // 按 Index 排序返回
```

## queue

Redis 消息队列，包级全局单例。

```go
var Redis Queue
func InitRedisQueue(config Config)      // 只需执行 1 次（Config 大量 *int/*bool 指针默认值）
func (s *Queue) Push(topic string, value any) bool
func (s *Queue) PushDelay(topic string, value any, delayed int64) bool   // ZSet 延迟队列
func (s *Queue) StartMonitor(topic string, methods any)                  // 监听+守卫协程，异常自动重启
func PushTopicDefault(method string, params ...any) bool                 // 反射按方法名派发
func PushTopicDefaultDelayed(method string, delayed int64, params ...any) bool
```

## kafka

```go
func NewProducer(config *Config) *Producer     // SASL/PLAIN+TLS
func (p *Producer) Send(msg any) error / Close()
func NewConsumer(config *Config, groupId string, startOffset int64) *Consumer
func (s *Consumer) Start(handler func(msg []byte) error)   // 自动重连，10s 重试
type Config struct{ Broker, Topic, Username, Password string }
```

## oss

对象存储适配器（策略模式），Driver：`"minio"|"qiniu"|"aliyun"`（常量 MinioOss/QiniuOss/AliyunOss）。

```go
func NewOSS(config Config) *Api
type Adapter interface {
    ObjectUrl(objectName string) string / ObjectName(objectUrl string) string
    UploadSign(objectName string) any
    Delete(objectName string) error
    UploadBase64(objectName, contentType, fileBase64 string)
    GetImageBase64/GetBase64/GetBase64NotPrefix(...) (string, string)
    Client() any
}
// Config 含 Domain/ThumbDomain/Private/Transport
```

## request

```go
func Get(url string, opt Option) ([]byte, *http.Response)
func PostJSON(url string, body interface{}, opt Option) ([]byte, *http.Response)
func PostFormData/PostFormURLEncoded/PostFile(...)
func PostJSONStream(url string, body any, opt Option, receive func(...byte))
func DownloadFile(option DownloadOption) string    // 支持进度回调 Percent func(percent float64)
func InitProxyRewrite(hosts map[string]string)     // 代理改写 Transport
type Option struct{ Headers *map[string]string; Timeout *int; FileFieldName *string;
    FileOtherField *map[string]string; Ctx *context.Context }   // 指针字段，NewOption 填默认值
```

## jwt

```go
func NewJwt[T any](secret string, redisApi *redis.Api) *Api[T]   // 泛型：业务数据塞 claims
func (s *Api[T]) Issued(body Body[T]) Response    // >=24h 顺延至 23:59:59；CheckTK 单点登录凭证存 Redis
func (s *Api[T]) Check(token string) Body[T]      // 失败 panic(&exception.AuthException{...}) → 401
```

## captcha

```go
func NewCaptcha(api *redis.Api, enable bool) Captcha
func (s *Captcha) Generate() *map[string]any     // {id, b64s}
func (s *Captcha) Verify(id, answer string)      // 不通过 panic
type ClickCaptcha                                // go-captcha 点击验证码：Generate(c *vingo.Context)、Verify(input ClickInput) bool
// RedisStore：base64Captcha 存储实现，前缀 captcha:，5 分钟过期
```

## cryptor

```go
func TextEncode(text string, secret []byte) string / TextDecode(...)   // AES-GCM
func ExchangeEncode(text string, times ...int) string / ExchangeDecode(...)  // 自研多层 Base64+字符对交换混淆
func Md5(str string) string / Md5File(filePath string) string
func TextBase64Encode/Decode(...)
```

## file

少数返回 error 的包；特色：重名自动改名（uniquePath），冲突不覆盖。

```go
func Mkdir(path string, mode os.FileMode) (string, error)
func CreateText/Copy(src, dst string) (string, error) / Move / Rename
func Delete(path string) error
func ReadText/WriteText(...)
func Zip(src []string, dst string, method ...string) (string, error) / Unzip(...)
func Size(...)
```

## email

```go
func SendMail(host string, port int, username, password, from, fromName string, to []string, subject, body string) error
// SMTPS(465) 直发；常量 EMAIL_CHECK/EMAIL_BIND/PASSWORD_FORGET（邮件验证码场景标识）
```

## model

通用业务模型（均带 Diff db.DiffBox[T] + TableName()）。

```go
type Rule struct{...}        // 权限规则树：path/len 钩子自动维护，MergeApis()
type RuleOptions            // 页面/接口/外链三类
type UpdateLog struct{...}  // 变更日志表，配合 db.Api.ChangeLog
type Dict struct{...}       // 字典表
```

## region

```go
func NewRegion(nodes []Node) Region
// GetNameByCode/GetNamesByCode/GetCodeByName/GetChildrenByCode/GetSonCodes/GetSonNames
// CreateWithAutoCode(parentCode, name)   自动生成区划码（2/4/6/9 位分层）
// Create/Update/Delete
type Node struct{ NodeBase; Children []Node }
```

## wechat

```go
func NewMiniProgram(miniProgramConfig *MiniProgramConfig) *MiniProgram   // silenceper/wechat v2 + Redis 缓存 access_token
func (s *MiniProgram) FaceGetVerifyId(name, idCard, openid string) string          // 人脸核身 2.0
func (s *MiniProgram) FaceQueryVerifyInfo(verifyId string, success func(info CertInfo))
type Cache struct{ RedisApi *vReids.Api }   // 实现 wechat SDK 的 cache 接口
```

## onlyoffice

```go
func NewApi(config Config) *Api    // Config 内嵌 OSS *oss.Api
func (s *Api) CreateDocx(objectName string)              // 创建空 docx 上传 OSS
func (s *Api) UpdateDocx(c *vingo.Context)               // OnlyOffice 回调处理，status=6 拉回文件覆盖原 object
func CreateEmptyDocxBase64() (string, error)
```

## cli

```go
func InitCli(options Options)    // Options{Enable bool; DatabaseApi *db.Api; Register func()}
// flag：-m table1,table2（生成模型）、-build-dev/-build-prod l|w|m|l_arm|m_arm（交叉编译）、-v3 <ver>（go get 更新自身）、-h
var Version = "dev"              // 编译期注入，格式 V2006.01.02_15.04.05
func BuildProject(value string, version string)
```

## extend/config

下游应用配置模板（供下游项目复制参考，不是库自身的运行配置）。

```go
type Config struct{ System System; Database db.Config; Redis redis.Config }   // yaml
func InitConfig()    // //go:embed dev.yml/prod.yml 按 ldflags 注入的 version 选环境，工作目录 config.yml 可覆盖；设 TZ=Asia/Shanghai
var ApiConfig *Config
// System 含 service/super/auth（lock、strength 密码强度）/right/cli/secret/sync
```

## 已知隐患（修改相关代码时注意）

- `db/common.go` 的 `QueryWhereLike/QueryWherePath` 等直接 fmt.Sprintf 拼值到 SQL，值未参数化（`PageOrder.HandleColumn` 对列名做了白名单，但 LIKE 类值未处理）——扩展查询构造器时优先参数化
- `vingo/context.go` 中有大量注释掉的旧请求日志代码（功能已迁移到 `router/middle.go` + logs 包），不要误恢复
- test/ 目录下多个文件同为 `package main`，直接 `go build ./...` 不受影响，但 `go run` 时需指定单文件

> AI生成