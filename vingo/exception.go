package vingo

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"runtime/debug"
)

// AuthException 身份验证异常
type AuthException struct {
	Message string
}

// DbException 数据库事务异常
type DbException struct {
	Message string
}

// ConfirmException 需要点击确认的异常
type ConfirmException struct {
	Message string
}

// BackException 页面后退异常
type BackException struct {
	Message string
}

// ExceptionHandler 异常处理
func ExceptionHandler(c *gin.Context) {
	context := &Context{c}

	defer func() {
		if err := recover(); err != nil {
			if GinDebug {
				debug.PrintStack()
			}
			switch t := err.(type) {
			case string:
				context.Response(&ResponseData{Message: t, Status: 200, Error: 1, ErrorType: "业务错误"})
			case *DbException:
				context.Response(&ResponseData{Message: t.Message, Status: 200, Error: 1, ErrorType: "数据库错误"})
			case *ConfirmException:
				context.Response(&ResponseData{Message: t.Message, Status: 200, Error: 2, ErrorType: "业务错误"})
			case *BackException:
				context.Response(&ResponseData{Message: t.Message, Status: 200, Error: 3, ErrorType: "业务错误"})
			case *AuthException:
				context.Response(&ResponseData{Message: t.Message, Status: 401, Error: 1})
			default:
				context.Response(&ResponseData{Message: t.(error).Error(), Status: 200, Error: 1, ErrorType: "异常错误"})
			}
			c.Abort()
		}
	}()
	c.Next()
}

func ExceptionCatch(s string, emit bool) {
	if err := recover(); err != nil {
		LogError(fmt.Sprintf("%v：%v", s, err))
		// 将异常往外抛
		if emit {
			panic(err)
		}
	}
}
