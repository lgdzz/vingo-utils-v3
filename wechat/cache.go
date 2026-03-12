package wechat

import (
	"time"

	vReids "github.com/lgdzz/vingo-utils-v3/redis"
)

type Cache struct {
	RedisApi *vReids.Api // redis操作对象
}

// Get 获取一个值
func (s *Cache) Get(key string) interface{} {
	result, err := s.RedisApi.Client.Get(key).Result()
	if err != nil {
		return nil
	}
	return result
}

// Set 设置一个值
func (s *Cache) Set(key string, val interface{}, timeout time.Duration) error {
	return s.RedisApi.Client.Set(key, val, timeout).Err()
}

// IsExist 判断key是否存在
func (s *Cache) IsExist(key string) bool {
	result, _ := s.RedisApi.Client.Exists(key).Result()
	return result > 0
}

// Delete 删除
func (s *Cache) Delete(key string) error {
	return s.RedisApi.Client.Del(key).Err()
}
