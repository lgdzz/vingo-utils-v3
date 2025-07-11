package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lgdzz/vingo-utils-v3/vingo"
	"time"
)

func BaseMiddle(hook *Hook) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Set("requestStart", startTime)
		c.Set("requestUUID", vingo.GetUUID())

		if hook.BaseMiddle != nil {
			hook.BaseMiddle(c)
		}
	}
}
