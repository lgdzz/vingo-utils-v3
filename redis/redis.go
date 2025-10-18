// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/11
// 描述：
// *****************************************************************************

package redis

import (
	"fmt"
	"github.com/go-redis/redis"
)

// NewRedis 新建一个redis连接池
func NewRedis(config Config) *Api {
	config.StringValue(&config.Host, "127.0.0.1")
	config.StringValue(&config.Port, "6379")
	config.StringValue(&config.Prefix, "")
	config.IntValue(&config.PoolSize, 20)
	config.IntValue(&config.MinIdleConns, 10)

	var api = Api{
		Config: config,
	}

	api.Client = redis.NewClient(&redis.Options{
		//连接信息
		Network:  "tcp",                                          //网络类型，tcp or unix，默认tcp
		Addr:     fmt.Sprintf("%v:%v", config.Host, config.Port), //主机名+冒号+端口，默认localhost:6379
		Password: config.Password,                                //密码
		DB:       config.Select,                                  // redis数据库index

		//连接池容量及闲置连接数量
		PoolSize:     config.PoolSize,     // 连接池最大socket连接数，应该设置为服务器CPU核心数的两倍
		MinIdleConns: config.MinIdleConns, // 在启动阶段创建指定数量的Idle连接，一般来说，可以将其设置为PoolSize的一半
	})
	// 测试连接是否正常
	_, err := api.Client.Ping().Result()
	if err != nil {
		fmt.Println(fmt.Sprintf("Redis连接异常：%v", err.Error()))
	}

	return &api
}
