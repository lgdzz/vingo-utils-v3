package wechat

import (
	vReids "github.com/lgdzz/vingo-utils-v3/redis"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/miniprogram"
	"github.com/silenceper/wechat/v2/miniprogram/config"
)

type MiniProgramConfig struct {
	AppID     string      `json:"appId"`
	AppSecret string      `json:"appSecret"`
	RedisApi  *vReids.Api // redis操作对象
}

// MiniProgram 微信小程序
func MiniProgram(miniProgramConfig *MiniProgramConfig) *miniprogram.MiniProgram {
	wc := wechat.NewWechat()
	cfg := &config.Config{
		AppID:     miniProgramConfig.AppID,
		AppSecret: miniProgramConfig.AppSecret,
		Cache:     &Cache{RedisApi: miniProgramConfig.RedisApi},
	}
	return wc.GetMiniProgram(cfg)
}
