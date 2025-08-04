package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lgdzz/vingo-utils-v3/vingo"
	"net/http"
	"time"
)

func BaseMiddle(hook *Hook) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		// 允许的方法
		switch c.Request.Method {
		case http.MethodGet, http.MethodPost:

			startTime := time.Now()
			c.Set("requestStart", startTime)
			c.Set("requestUUID", vingo.GetUUID())

			if hook.BaseMiddle != nil {
				hook.BaseMiddle(c)
			}

		default:
			// 禁用 TRACE/TRACK 以及其他未显式允许的方法
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Forbidden method: " + method,
			})
			return
		}
	}
}
