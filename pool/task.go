// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/9
// 描述：协程池
// *****************************************************************************

package pool

import (
	"context"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/lgdzz/vingo-utils-v3/vingo"
	"sync"
)

// TaskFunc 支持返回结果和错误的任务函数签名
type TaskFunc func(ctx context.Context) Result

type Result struct {
	Index  int     `json:"index"`
	Data   any     `json:"data"`
	Result any     `json:"result"`
	Error  *string `json:"error"`
}

type GoroutinePool struct {
	maxWorkers int
	tasks      chan TaskFunc
	results    []Result
	mu         sync.Mutex
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewGoroutinePool 创建协程池，支持 context
func NewGoroutinePool(ctx context.Context, maxWorkers int) *GoroutinePool {
	c, cancel := context.WithCancel(ctx)
	return &GoroutinePool{
		maxWorkers: maxWorkers,
		tasks:      make(chan TaskFunc),
		results:    make([]Result, 0),
		ctx:        c,
		cancel:     cancel,
	}
}

// Run 启动 worker
func (s *GoroutinePool) Run() {
	for i := 0; i < s.maxWorkers; i++ {
		go func() {
			for {
				select {
				case <-s.ctx.Done():
					return
				case task, ok := <-s.tasks:
					if !ok {
						return
					}
					res := task(s.ctx)

					s.mu.Lock()
					s.results = append(s.results, res)
					s.mu.Unlock()
					s.wg.Done()
				}
			}
		}()
	}
}

// Submit 提交任务
func (s *GoroutinePool) Submit(task TaskFunc) {
	s.wg.Add(1)
	s.tasks <- task
}

// Cancel 取消所有任务
func (s *GoroutinePool) Cancel() {
	s.cancel()
}

// CloseAndWait 等待所有任务完成并关闭通道，返回所有结果
func (s *GoroutinePool) CloseAndWait() []Result {
	s.wg.Wait()
	close(s.tasks)

	if len(s.results) > 0 && s.results[0].Index > -1 {
		_ = slice.SortByField(s.results, "Index", "asc")
	}

	return s.results
}

func BusinessHandle[T any](data T, index int, handle func(object T) any) (res Result) {
	defer func() {
		if err := recover(); err != nil {
			vingo.LogError(fmt.Sprintf("BusinessHandle：%v", err))
			res = Result{
				Index:  index,
				Data:   data,
				Result: "fail",
				Error:  vingo.Of(fmt.Sprintf("%v", err)),
			}
		}
	}()

	resData := handle(data)
	if resData != nil {
		res = Result{
			Index:  index,
			Data:   resData,
			Result: "success",
		}
	} else {
		res = Result{
			Index:  index,
			Data:   data,
			Result: "success",
		}
	}
	return
}

// FastPoolWithContext 快速创建带上下文的协程池
func FastPoolWithContext(ctx context.Context, maxWorkers int, handle func(p *GoroutinePool)) (out []Result) {
	p := NewGoroutinePool(ctx, maxWorkers)
	p.Run()

	handle(p)
	out = p.CloseAndWait()
	return
}

// FastPool 快速创建协程池
func FastPool(maxWorkers int, handle func(p *GoroutinePool)) (out []Result) {
	p := NewGoroutinePool(context.Background(), maxWorkers)
	p.Run()

	handle(p)
	out = p.CloseAndWait()
	return
}

// FastSubmit 快速提交协程任务
// 如果是for循环协程任务可用index排序，默认-1，不排序
func (s *GoroutinePool) FastSubmit(handle func() any, index ...int) {
	s.Submit(func(ctx context.Context) Result {
		var i = -1
		if len(index) > 0 {
			i = index[0]
		}
		return BusinessHandle(nil, i, func(object any) any {
			return handle()
		})
	})
}
