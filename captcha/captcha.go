// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/18
// 描述：
// *****************************************************************************

package captcha

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/mojocn/base64Captcha"
	"time"
)

type Captcha struct {
	Store *RedisStore
}

func NewCaptcha(client *redis.Client) Captcha {
	return Captcha{
		Store: &RedisStore{
			Client: client,
			Prefix: "captcha:",
			Expire: time.Minute * 5,
		},
	}
}

func (s *Captcha) Generate() map[string]any {
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	c := base64Captcha.NewCaptcha(driver, s.Store)
	id, b64s, answer, err := c.Generate()
	if err != nil {
		panic("验证码生成失败")
	}
	fmt.Println(id)
	fmt.Println(answer)
	return map[string]any{
		"id":   id,
		"b64s": b64s,
	}
}

func (s *Captcha) Verify(id string, answer string) bool {
	return s.Store.Verify(id, answer, true)
}
