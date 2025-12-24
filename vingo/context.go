package vingo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FunctionModule struct {
	Name string
}

type Context struct {
	*gin.Context
}

type KeyValue struct {
	Key   string
	Value string
}

var Valid *validator.Validate

func init() {
	if Valid == nil {
		Valid = validator.New()
	}
}

// AddQuery 动态添加查询条件
func (c *Context) AddQuery(kv ...KeyValue) {
	var query = c.Request.URL.Query()
	for _, item := range kv {
		query.Add(item.Key, item.Value)
	}
	c.Request.URL.RawQuery = query.Encode()
}

// UrlDecode url解码
func (c *Context) UrlDecode() (dStr string) {
	var (
		err  error
		eStr = c.Request.RequestURI
	)
	dStr, err = url.QueryUnescape(eStr)
	if err != nil {
		dStr = eStr
	}
	return
}

// GetRealClientIP 获取客户端真实IP
func (c *Context) GetRealClientIP() string {
	var ip string
	if ip = c.Request.Header.Get("X-Forwarded-For"); ip == "" {
		ip = c.Request.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = c.Request.RemoteAddr
	} else {
		ips := strings.Split(ip, ", ")
		ip = ips[len(ips)-1]
	}
	return ip
}

// Response 请求成功
func (c *Context) Response(d *ResponseData) {
	c.Set("clientIp", c.GetRealClientIP())

	if d.Message == "" {
		d.Message = "Success"
	}
	if d.Status == 0 {
		d.Status = 200
	}
	uuid := c.GetString("requestUUID")
	c.JSON(d.Status, gin.H{
		"uuid":      uuid,
		"error":     d.Error,
		"message":   d.Message,
		"data":      d.Data,
		"timestamp": time.Now().Unix(),
	})

	if !d.NoLog {
		// 记录请求日志
		go func(context *Context, uuid string, d *ResponseData) {
			ExceptionCatch("记录请求日志异常", false)
			startTime := context.GetTime("requestStart")
			endTime := time.Now()
			latency := endTime.Sub(startTime)
			millisecond := float64(latency.Nanoseconds()) / float64(time.Millisecond)
			duration := fmt.Sprintf("%.3fms", millisecond)
			if millisecond > 300 {
				duration += ":慢接口"
			}

			var err string
			if d.Error == 1 {
				err = d.Message
			}

			if context.Request.Method == "GET" {
				LogRequest(duration, fmt.Sprintf("{\"uuid\":\"%v\",\"method\":\"%v\",\"url\":\"%v\",\"err\":\"%v\",\"errType\":\"%v\",\"userAgent\":\"%v\",\"clientIP\":\"%v\",\"user\":\"%v\"}", uuid, context.Request.Method, context.UrlDecode(), err, d.ErrorType, c.GetHeader("User-Agent"), c.GetString("clientIp"), c.GetString("user")))
			} else {
				body := context.GetString("requestBody")
				if body == "" {
					body = "\"\""
				}
				LogRequest(duration, fmt.Sprintf("{\"uuid\":\"%v\",\"method\":\"%v\",\"url\":\"%v\",\"body\":%v,\"err\":\"%v\",\"errType\":\"%v\",\"userAgent\":\"%v\",\"clientIP\":\"%v\",\"user\":\"%v\"}", uuid, context.Request.Method, context.Request.RequestURI, body, err, d.ErrorType, c.GetHeader("User-Agent"), c.GetString("clientIp"), c.GetString("user")))
			}

			if d.ErrorType == "异常错误" {
				stack := string(debug.Stack())
				LogError(stack)
			}
		}(c, uuid, d)
	}
}

// ResponseBody 请求成功，带data数据
func (c *Context) ResponseBody(data any) {
	c.Response(&ResponseData{Data: data})
}

// ResponseSuccess 请求成功，默认
func (c *Context) ResponseSuccess(data ...any) {
	if len(data) == 0 {
		c.Response(&ResponseData{})
	} else {
		c.Response(&ResponseData{Data: data[0]})
	}
}

// ResponseBase64 请求成功，数据进行base64编码
func (c *Context) ResponseBase64(data ...any) {
	if len(data) == 0 {
		c.Response(&ResponseData{})
	} else {
		byteData, err := json.Marshal(data[0])
		if err != nil {
			panic(fmt.Sprintf("Error marshaling data: %v", err))
		}
		c.Response(&ResponseData{Data: base64.StdEncoding.EncodeToString(byteData)})
	}
}

// ShieldRobots 屏蔽搜索引擎爬虫
func ShieldRobots(r *gin.Engine) {
	r.GET("/robots.txt", func(c *gin.Context) {
		c.String(200, `User-agent: *
Disallow: /`)
	})
}

// AllowCrossDomain 设置允许跨域访问的域名或IP地址
func AllowCrossDomain(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
}

// ApiAddress 输出接口地址
func ApiAddress(port int) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, item := range addr {
		if ipNet, ok := item.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				fmt.Println(fmt.Sprintf("+ 接口地址：http://%v:%d", ipNet.IP.String(), port))
			}
		}
	}
}

func (c *Context) SetUserId(value int) {
	c.Set("userId", value)
}
func (c *Context) SetRealname(value string) {
	c.Set("realname", value)
}
func (c *Context) SetAccId(value int) {
	c.Set("accId", value)
}
func (c *Context) SetAccName(value string) {
	c.Set("accName", value)
}
func (c *Context) SetOrgPid(value int) {
	c.Set("orgPid", value)
}
func (c *Context) SetOrgId(value int) {
	c.Set("orgId", value)
}
func (c *Context) SetOrgName(value string) {
	c.Set("orgName", value)
}
func (c *Context) SetOrgTypeIds(value []int) {
	c.Set("orgTypeIds", value)
}
func (c *Context) SetOrgTypeNames(value []string) {
	c.Set("orgTypeNames", value)
}
func (c *Context) SetDeptIds(value []int) {
	c.Set("deptIds", value)
}
func (c *Context) SetDeptNames(value []string) {
	c.Set("deptNames", value)
}
func (c *Context) SetRoleIds(value []int) {
	c.Set("roleIds", value)
}
func (c *Context) SetRoleNames(value []string) {
	c.Set("roleNames", value)
}
func (c *Context) SetRoleTags(value []string) {
	c.Set("roleTags", value)
}
func (c *Context) SetDataScope(value int) {
	c.Set("dataScope", value)
}
func (c *Context) SetIsManageChannel(value bool) {
	c.Set("isManageChannel", value)
}

func (c *Context) GetUserId() int {
	return c.GetInt("userId")
}
func (c *Context) GetRealname() string {
	return c.GetString("realname")
}
func (c *Context) GetAccId() int {
	return c.GetInt("accId")
}
func (c *Context) GetAccName() string {
	return c.GetString("accName")
}
func (c *Context) GetOrgPid() int {
	return c.GetInt("orgPid")
}
func (c *Context) GetOrgId() int {
	return c.GetInt("orgId")
}
func (c *Context) GetOrgName() string {
	return c.GetString("orgName")
}
func (c *Context) GetOrgTypeIds() []int {
	return c.getInts("orgTypeIds")
}
func (c *Context) GetOrgTypeNames() []string {
	return c.getTexts("orgTypeNames")
}
func (c *Context) GetDeptIds() []int {
	return c.getInts("deptIds")
}
func (c *Context) GetDeptNames() []string {
	return c.getTexts("deptNames")
}
func (c *Context) GetRoleIds() []int {
	return c.getInts("roleIds")
}
func (c *Context) GetRoleNames() []string {
	return c.getTexts("roleNames")
}
func (c *Context) GetRoleTags() []string {
	return c.getTexts("roleTags")
}
func (c *Context) GetDataScope() int {
	return c.GetInt("dataScope")
}
func (c *Context) GetIsManageChannel() bool {
	return c.GetBool("isManageChannel")
}

func (c *Context) getTexts(key string) []string {
	value, exists := c.Get(key)
	if exists {
		return value.([]string)
	}
	return []string{}
}
func (c *Context) getInts(key string) []int {
	value, exists := c.Get(key)
	if exists {
		return value.([]int)
	}
	return []int{}
}

func (c *Context) VerifyRoleTags(tags ...string) bool {
	for _, t := range tags {
		if slice.Contain(c.GetRoleTags(), t) {
			return true
		}
	}
	return false
}
func (c *Context) VerifyOrgTypes(types ...int) bool {
	for _, t := range types {
		if slice.Contain(c.GetOrgTypeIds(), t) {
			return true
		}
	}
	return false
}
func (c *Context) VerifyRoleIds(id ...int) bool {
	for _, i := range id {
		if slice.Contain(c.GetRoleIds(), i) {
			return true
		}
	}
	return false
}

type DataScope struct {
	MaxLevel  func()
	OrgLevel  func()
	DeptLevel func()
	AccLevel  func()
}

func (s *DataScope) Handle(c *Context) {
	switch c.GetDataScope() {
	case 4: // 本单位至下属单位（最大维度）
		if s.MaxLevel != nil {
			s.MaxLevel()
		}
	case 3: // 本单位
		if s.OrgLevel != nil {
			s.OrgLevel()
		}
	case 2: // 本部门
		if s.DeptLevel != nil {
			s.DeptLevel()
		}
	case 1: // 本账户
		if s.AccLevel != nil {
			s.AccLevel()
		}
	}
}

// 注册get路由
func RoutesGet(g *gin.RouterGroup, path string, handler func(*Context)) {
	g.GET(path, func(c *gin.Context) {
		context := &Context{Context: c}
		handler(context)
		if ApiOplog.Enable {
			go ApiOplog.Write(context)
		}
	})
}

// RoutesPost 注册post路由
func RoutesPost(g *gin.RouterGroup, path string, handler func(*Context)) {
	g.POST(path, func(c *gin.Context) {
		context := &Context{Context: c}
		handler(context)
		if ApiOplog.Enable {
			go ApiOplog.Write(context)
		}
	})
}

// RoutesPut 注册put路由
func RoutesPut(g *gin.RouterGroup, path string, handler func(*Context)) {
	g.PUT(path, func(c *gin.Context) {
		context := &Context{Context: c}
		handler(context)
		if ApiOplog.Enable {
			go ApiOplog.Write(context)
		}
	})
}

// RoutesPatch 注册patch路由
func RoutesPatch(g *gin.RouterGroup, path string, handler func(*Context)) {
	g.PATCH(path, func(c *gin.Context) {
		context := &Context{Context: c}
		handler(context)
		if ApiOplog.Enable {
			go ApiOplog.Write(context)
		}
	})
}

// RoutesDelete 注册delete路由
func RoutesDelete(g *gin.RouterGroup, path string, handler func(*Context)) {
	g.DELETE(path, func(c *gin.Context) {
		handler(&Context{Context: c})
	})
}

// GetRequestBody 获取请求body
// GetRequestBody[结构体类型](c)
func GetRequestBody[T any](c *Context, valid ...bool) T {
	var body T
	if err := c.ShouldBindJSON(&body); err != nil {
		panic(err.Error())
	}

	if len(valid) > 0 && valid[0] {
		if err := Valid.Struct(body); err != nil {
			// handle validation error
			panic(err)
		}
	}

	if data, err := json.Marshal(body); err != nil {
		panic(err.Error())
	} else {
		c.Set("requestBody", string(data))
	}
	return body
}

// GetRequestQuery 获取请求query
// GetRequestQuery[结构体类型](c)
func GetRequestQuery[T any](c *Context) T {
	var query T
	if err := c.ShouldBindQuery(&query); err != nil {
		panic(err.Error())
	}
	return query
}

type ResponseData struct {
	Status    int    // 状态
	Error     int    // 0-无错误|1-有错误
	ErrorType string // 错误类型
	Message   string // 消息
	Data      any    // 返回数据内容
	NoLog     bool   // true时不记录日志
}

type Oplog struct {
	Enable bool
	Write  func(c *Context)
}

var ApiOplog = Oplog{}
