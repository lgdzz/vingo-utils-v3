package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lgdzz/vingo-utils-v3/vingo"
)

func BaseMiddle(hook *Hook) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		if _, ok := hook.AllowMethods[method]; ok {
			startTime := time.Now()
			c.Set("requestStart", startTime)
			c.Set("requestUUID", vingo.GetUUID())

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
