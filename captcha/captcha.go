// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/18
// 描述：
// *****************************************************************************

package captcha

import (
	"fmt"
	"github.com/lgdzz/vingo-utils-v3/redis"
	"github.com/mojocn/base64Captcha"
	"log"
	"time"
)

type Captcha struct {
	Enable bool
	Debug  bool
	Length int
	Driver string
	Store  *RedisStore
}

func NewCaptcha(api *redis.Api, enable bool) Captcha {
	return Captcha{
		Enable: enable,
		Length: 5,
		Driver: "digit",
		Store: &RedisStore{
			Client: api.Client,
			Prefix: "captcha:",
			Expire: time.Minute * 5,
		},
	}
}

func (s *Captcha) Generate() *map[string]any {
	if !s.Enable {
		return nil
	}

	var driver base64Captcha.Driver
	switch s.Driver {
	case "digit":
		driver = base64Captcha.NewDriverDigit(80, 240, s.Length, 0.7, 80)
	default:
		panic("unsupported driver")
	}

	c := base64Captcha.NewCaptcha(driver, s.Store)
	id, b64s, answer, err := c.Generate()
	if err != nil {
		panic("验证码生成失败")
	}
	if s.Debug {
		log.Println(fmt.Sprintf("验证码：%v", answer))
	}
	return &map[string]any{
		"id":   id,
		"b64s": b64s,
	}
}

func (s *Captcha) Verify(id string, answer string) {
	if s.Enable && !s.Store.Verify(id, answer, true) {
		panic("验证码不正确")
	}
}
