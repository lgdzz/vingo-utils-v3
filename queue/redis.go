package queue

import (
	"fmt"
	"github.com/duke-git/lancet/v2/pointer"
	"github.com/go-redis/redis"
	vReids "github.com/lgdzz/vingo-utils-v3/redis"
	"github.com/lgdzz/vingo-utils-v3/vingo"
	"reflect"
	"time"
)

var Redis Queue

type Config struct {
	Debug             *bool       // 调试模式，为true时日志在控制台输出，否则记录到日志文件，默认为true
	AutoBootTime      *int        // 监控器异常自动重启时间间隔，默认3秒
	SortedSetRestTime *int        // 有序集合中没有消息时休息等待时间，默认2秒
	RetryWaitTime     *int        // 消费失败重试等待时间，默认5秒
	Handle            *Handle     // 消费处理方法调度中心，一般默认即可，特殊要求需实现Handler接口
	RedisApi          *vReids.Api // redis操作对象
}

type Queue struct {
	Config Config
}

// InitRedisQueue 初始化服务（只需要执行1次）
func InitRedisQueue(config Config) {
	if config.RedisApi != nil {
		Redis.Config.RedisApi = config.RedisApi
	}

	if config.Debug != nil {
		Redis.Config.Debug = config.Debug
	} else {
		Redis.Config.Debug = pointer.Of(true)
	}

	if config.AutoBootTime != nil {
		Redis.Config.AutoBootTime = config.AutoBootTime
	} else {
		Redis.Config.AutoBootTime = pointer.Of(3)
	}

	if config.SortedSetRestTime != nil {
		Redis.Config.SortedSetRestTime = config.SortedSetRestTime
	} else {
		Redis.Config.SortedSetRestTime = pointer.Of(2)
	}

	if config.RetryWaitTime != nil {
		Redis.Config.RetryWaitTime = config.RetryWaitTime
	} else {
		Redis.Config.RetryWaitTime = pointer.Of(5)
	}

	if config.Handle != nil {
		Redis.Config.Handle = config.Handle
	} else {
		Redis.Config.Handle = &Handle{}
	}
}

// 将消息转换为字符串类型
func (s *Queue) toString(value any) string {
	var kind = reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Struct:
		return vingo.JsonToString(value)
	case reflect.String:
		return value.(string)
	default:
		panic("RedisQueue.Push未知消息数据类型")
	}
}

func (s *Queue) getTopic(topic string) string {
	return fmt.Sprintf("%v%v.queue", s.Config.RedisApi.Config.Prefix, topic)
}

func (s *Queue) getDelayTopic(topic string) string {
	return fmt.Sprintf("%v%v.queue.delay", s.Config.RedisApi.Config.Prefix, topic)
}

// Push 推送实时任务
// topic-消息队列主题
// value-消息内容，可选类型[struct|string]
func (s *Queue) Push(topic string, value any) bool {
	r, err := s.Config.RedisApi.Client.RPush(s.getTopic(topic), s.toString(value)).Result()
	if err != nil {
		panic(err.Error())
	}
	return r > 0
}

// PushDelay 推送延迟任务
// topic-消息队列主题
// value-消息内容，可选类型[struct|string]
// delayed-延迟时间，单位：秒，如：60秒后执行，则传入60
func (s *Queue) PushDelay(topic string, value any, delayed int64) bool {
	var nowTime = time.Now()
	var score = float64(nowTime.Add(time.Duration(delayed) * time.Second).Unix())
	r, err := s.Config.RedisApi.Client.ZAdd(s.getDelayTopic(topic), redis.Z{Member: s.toString(value), Score: score}).Result()
	if err != nil {
		panic(err.Error())
	}
	return r >= 0
}

// StartMonitor 开始监听队列信息
func (s *Queue) StartMonitor(topic string, methods any) {
	go s.monitorGuard(topic, s.Config.Handle, methods)
	go s.monitorGuardDelay(topic)
}

// 队列监听守卫
func (s *Queue) monitorGuard(topic string, handler Handler, methods any) {
	defer func() {
		if err := recover(); err != nil {
			// 等待3秒后重启监听器
			time.Sleep(time.Second * time.Duration(*s.Config.AutoBootTime))
			if *s.Config.Debug {
				fmt.Println(fmt.Sprintf("[消息队列]监听器异常，进行重启."))
			} else {
				vingo.LogError(fmt.Sprintf("[消息队列]监听器异常，进行重启."))
			}
			s.monitorGuard(topic, handler, methods)
		}
	}()
	s.monitor(topic, handler, methods)
}

// 队列监听
func (s *Queue) monitor(topic string, handler Handler, methods any) {
	topicQueue := s.getTopic(topic)
	for {
		r, err := s.Config.RedisApi.Client.BLPop(0, topicQueue).Result()
		if err != nil {
			panic(err.Error())
		}
		value := r[1]
		func(topic string, value string) {
			defer func() {
				if err := recover(); err != nil {
					// 调试在控制台输出错误日志
					if *s.Config.Debug {
						fmt.Printf("[消息队列]消费失败，Message：%v，Error：%v\n", value, err)
					} else {
						vingo.LogError(fmt.Sprintf("[消息队列]消费失败，Message：%v，Error：%v", value, err))
					}
					// 如果消息处理异常，则将任务推送到延迟队列，在指定时间后再次消费
					s.PushDelay(topic, value, int64(*s.Config.RetryWaitTime))
				}
			}()
			// 执行消息处理
			handler.HandleMessage(&value, methods)
		}(topic, value)
	}
}

// StartMonitorDelay 开始监听队列信息(延迟)
// Deprecated: This function is no longer recommended for use.
// Suggested: 集成到StartMonitor中一起开启
func (s *Queue) StartMonitorDelay(topic string) {
	go s.monitorGuardDelay(topic)
}

// 队列监听守卫(延迟)
func (s *Queue) monitorGuardDelay(topic string) {
	defer func() {
		if err := recover(); err != nil {
			// 等待3秒后重启监听器
			time.Sleep(time.Second * time.Duration(*s.Config.AutoBootTime))
			if *s.Config.Debug {
				fmt.Println(fmt.Sprintf("[消息队列]监听器delay异常，进行重启."))
			} else {
				vingo.LogError(fmt.Sprintf("[消息队列]监听器delay异常，进行重启."))
			}
			s.monitorGuardDelay(topic)
		}
	}()
	s.monitorDelay(topic)
}

// 队列监听(延迟)
func (s *Queue) monitorDelay(topic string) {
	topicDelay := s.getDelayTopic(topic)
	for {
		r, err := s.Config.RedisApi.Client.ZRangeWithScores(topicDelay, 0, 0).Result()
		if err != nil {
			panic(err.Error())
		}
		if len(r) > 0 {
			// 将分数转换为时间戳
			// 获取第一个记录的分数和成员
			score := r[0].Score
			member := r[0].Member.(string)
			// 将分数转换为时间戳
			expiryTime := time.Unix(int64(score), 0)
			// 当前时间
			now := time.Now()
			// 判断是否已经到期
			if expiryTime.After(now) {
				// 计算剩余时间
				remainingTime := expiryTime.Sub(now)
				// 暂停等待剩余时间
				if remainingTime.Seconds() < float64(*s.Config.SortedSetRestTime) {
					// 剩余时间小于休息时间，则按剩余时间暂停
					time.Sleep(remainingTime)
				} else {
					// 否则直接用休息时间暂停
					time.Sleep(time.Second * time.Duration(*s.Config.SortedSetRestTime))
				}
			} else {
				// 删除记录有序集合中的记录
				s.Config.RedisApi.Client.ZRem(topicDelay, member)
				// 将任务加入到实时队列
				s.Push(topic, member)
			}
		} else {
			// 有序集合中没有消息时休息等待
			time.Sleep(time.Second * time.Duration(*s.Config.SortedSetRestTime))
		}
	}
}

type Handler interface {
	HandleMessage(message *string, methods any)
}

type Handle struct{}

// HandleMessage 消费处理方法调度中心
func (s *Handle) HandleMessage(message *string, methods any) {
	// 将消息解析成结构体
	var body MessagePackage
	vingo.StringToJson(*message, &body)
	CallStructFunc(methods, body.Method, body.Params...)
}

type MessagePackage struct {
	Method string
	Params []any
}

func CallStructFunc(obj any, method string, params ...any) {
	t := reflect.TypeOf(obj)
	_func, ok := t.MethodByName(method)
	if !ok {
		fmt.Println(fmt.Sprintf("%v方法不存在", method))
		return
	}

	_param := make([]reflect.Value, 0)
	_param = append(_param, reflect.ValueOf(obj))
	for _, value := range params {
		_param = append(_param, reflect.ValueOf(value))
	}
	_func.Func.Call(_param)
}

func PushTopicDefault(method string, params ...any) bool {
	return Redis.Push("vingo", MessagePackage{
		Method: method,
		Params: params,
	})
}

func PushTopicDefaultDelayed(method string, delayed int64, params ...any) bool {
	return Redis.PushDelay("vingo", MessagePackage{
		Method: method,
		Params: params,
	}, delayed)
}
