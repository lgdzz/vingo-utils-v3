package router

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lgdzz/vingo-utils-v3/logs"
	"github.com/lgdzz/vingo-utils-v3/vingo"
)

func BaseMiddle(hook *Hook) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		if _, ok := hook.AllowMethods[method]; ok {
			startTime := time.Now()
			requestUUID := vingo.GetUUID()
			c.Set("requestStart", startTime)
			c.Set("requestUUID", requestUUID)

			defer func() {
				if err := recover(); err != nil {
					c.Set("responseMessage", fmt.Sprint(err))
					requestLog(c)
					panic(err)
				} else {
					requestLog(c)
				}
			}()

			if hook.BaseMiddle != nil {
				hook.BaseMiddle(c)
			}
		} else {
			// 禁止的方法
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Forbidden method: " + method,
			})
			return
		}
	}
}

func requestLog(c *gin.Context) {
	logs.Request(
		"请求访问",
		vingo.JsonToStringRaw(map[string]any{
			"requestUUID":     c.GetString("requestUUID"),
			"method":          c.Request.Method,
			"path":            c.Request.URL.Path,
			"query":           c.Request.URL.RawQuery,
			"status":          c.Writer.Status(),
			"cost":            time.Since(c.GetTime("requestStart")).String(),
			"clientIP":        c.ClientIP(),
			"requestBody":     c.GetString("requestBody"),
			"responseMessage": c.GetString("responseMessage"),
		}),
	)
}
