package router

import (
	"fmt"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gin-gonic/gin"
	"github.com/lgdzz/vingo-utils-v3/db"
	"github.com/lgdzz/vingo-utils-v3/moment"
	"github.com/lgdzz/vingo-utils-v3/redis"
	"github.com/lgdzz/vingo-utils-v3/vingo"
	"net/http"
	"time"
)

type Hook struct {
	Option         HookOption
	RegisterRouter func(r *gin.Engine)
	BaseMiddle     func(c *gin.Context)
	LoadWeb        []WebItem // 通过此配置将前端项目打包到项目中
}

type HookOption struct {
	Name      string
	Port      int
	Copyright string
	Debug     bool
	Database  *db.Api
	Redis     *redis.Config
	startTime time.Time // 启动时间
}

type WebItem struct {
	Route string
	FS    *assetfs.AssetFS // go-bindata-assetfs -pkg {admin} -o router/{admin}/bindata.go dist/...
}

// InitRouter 初始化路由
func InitRouter(hook *Hook) {
	var option = hook.Option
	if option.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	vingo.GinDebug = option.Debug

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	_ = r.SetTrustedProxies(nil)

	// 加载web前端
	for _, item := range hook.LoadWeb {
		currentItem := item
		r.GET(currentItem.Route+"/*filepath", func(c *gin.Context) {
			http.FileServer(currentItem.FS).ServeHTTP(c.Writer, c.Request)
		})
	}

	// 屏蔽搜索引擎爬虫
	vingo.ShieldRobots(r)

	// 404捕捉
	r.NoRoute(func(c *gin.Context) {
		context := vingo.Context{Context: c}
		context.Response(&vingo.ResponseData{Message: "404:Not Found", Error: 1})
	})

	// 注册异常处理、基础中间件
	r.Use(vingo.ExceptionHandler, BaseMiddle(hook))

	// 注册基础路由
	r.GET("/", func(c *gin.Context) {
		Console(c, option)
	})

	// 注册路由
	if hook.RegisterRouter == nil {
		fmt.Println("请实现hook.RegisterRouter")
		return
	}
	hook.RegisterRouter(r)

	fmt.Println("+------------------------------------------------------------+")
	fmt.Println(fmt.Sprintf("+ 项目名称：%v", option.Name))
	fmt.Println(fmt.Sprintf("+ 服务端口：%d", option.Port))
	fmt.Println(fmt.Sprintf("+ 调试模式：%v", option.Debug))
	if option.Database != nil {
		fmt.Println(fmt.Sprintf("+ %v：%v:%v db:%v", option.Database.Config.Driver, option.Database.Config.Host, option.Database.Config.Port, option.Database.Config.Dbname))
	}
	if option.Redis != nil {
		fmt.Println(fmt.Sprintf("+ redis：%v:%v db:%v prefix:%v", option.Redis.Host, option.Redis.Port, option.Redis.Select, option.Redis.Prefix))
	}
	if option.Copyright == "" {
		option.Copyright = fmt.Sprintf("Copyright © %d lgdz", time.Now().Year())
	}

	option.startTime = time.Now()

	vingo.ApiAddress(int(option.Port))
	fmt.Println(fmt.Sprintf("+ 启动时间：%v", option.startTime.Format(moment.DateTimeFormat)))
	fmt.Println(fmt.Sprintf("+ 技术支持：%v", option.Copyright))
	fmt.Println("+------------------------------------------------------------+")

	// 开启服务
	_ = r.Run(fmt.Sprintf(":%d", option.Port))
}

func Console(c *gin.Context, option HookOption) {
	c.JSON(200, gin.H{
		"启动时间": option.startTime.Format(moment.DateTimeFormat),
		"已运行":  FormatDuration(option.startTime, time.Now()),
	})
}

func FormatDuration(start, end time.Time) string {
	duration := end.Sub(start)

	// 取绝对值
	if duration < 0 {
		duration = -duration
	}

	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	return fmt.Sprintf("%d天 %d时 %d分 %d秒", days, hours, minutes, seconds)
}
