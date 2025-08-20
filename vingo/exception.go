package vingo

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lgdzz/vingo-utils-exception/exception"
	"runtime/debug"
)

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
			case *exception.DbException:
				context.Response(&ResponseData{Message: t.Message, Status: 200, Error: 1, ErrorType: "数据库错误"})
			case *exception.ConfirmException:
				context.Response(&ResponseData{Message: t.Message, Status: 200, Error: 2, ErrorType: "业务错误"})
			case *exception.BackException:
				context.Response(&ResponseData{Message: t.Message, Status: 200, Error: 3, ErrorType: "业务错误"})
			case *exception.AuthException:
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
