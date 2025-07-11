// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/11
// 描述：
// *****************************************************************************

package redis

import (
	"encoding/json"
	"errors"
	"github.com/go-redis/redis"
	"time"
)

type Api struct {
	Client *redis.Client
	Config Config
}

func (s *Api) buildKey(key string) string {
	return s.Config.Prefix + key
}

func (s *Api) Get(key string, value any) (exist bool) {
	text, err := s.Client.Get(s.buildKey(key)).Result()
	if errors.Is(err, redis.Nil) {
		return
	} else if err != nil {
		panic(err)
	} else {
		err = json.Unmarshal([]byte(text), value)
		if err != nil {
			panic(err.Error())
		}
		exist = true
		return
	}
}

func (s *Api) Set(key string, value any, expiration time.Duration) string {
	v, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	result, err := s.Client.Set(s.buildKey(key), v, expiration).Result()
	if err != nil {
		panic(err)
	}
	return result
}

func (s *Api) HSet(key string, field string, value any) bool {
	v, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	result, err := s.Client.HSet(s.buildKey(key), field, v).Result()
	if err != nil {
		panic(err)
	}
	return result
}

func (s *Api) HGet(key string, field string, value any) (exist bool) {
	text, err := s.Client.HGet(s.buildKey(key), field).Result()
	if errors.Is(err, redis.Nil) {
		return
	} else if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(text), value)
	if err != nil {
		panic(err.Error())
	}
	exist = true
	return
}

func (s *Api) Del(key ...string) int64 {
	var keys = make([]string, 0)
	for _, item := range key {
		keys = append(keys, s.buildKey(item))
	}
	result, err := s.Client.Del(keys...).Result()
	if err != nil {
		panic(err)
	}
	return result
}
