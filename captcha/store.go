// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/18
// 描述：
// *****************************************************************************

package captcha

import (
	"github.com/go-redis/redis"
	"time"
)

type RedisStore struct {
	Client *redis.Client
	Prefix string        // 键前缀，例如 "captcha_"
	Expire time.Duration // 有效期，如 time.Minute * 5
}

func (r *RedisStore) Set(id, value string) error {
	return r.Client.Set(r.Prefix+id, value, r.Expire).Err()
}

func (r *RedisStore) Get(id string, clear bool) string {
	key := r.Prefix + id
	val, err := r.Client.Get(key).Result()
	if err != nil {
		return ""
	}
	if clear {
		r.Client.Del(key)
	}
	return val
}

func (r *RedisStore) Verify(id, answer string, clear bool) bool {
	if id == "" {
		panic("验证码Id不能为空")
	} else if answer == "" {
		panic("验证码不能为空")
	}
	val := r.Get(id, clear)
	return val == answer
}
