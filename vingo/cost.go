// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/6/12
// 描述：耗时监听
// *****************************************************************************

package vingo

import (
	"fmt"
	"log"
	"time"
)

var CostDebug bool

type Cost struct {
	start time.Time
	last  time.Time
}

func NewCost() *Cost {
	if !CostDebug {
		return &Cost{}
	}
	now := time.Now()
	return &Cost{
		start: now,
		last:  now,
	}
}

const (
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	reset  = "\033[0m"
)

func formatStep(d time.Duration, s string) string {
	switch {
	case d < 100*time.Millisecond:
		return green + s + reset
	case d >= time.Second:
		return red + s + reset
	default:
		return yellow + s + reset
	}
}

func (c *Cost) Mark(name string) {
	if !CostDebug {
		return
	}

	now := time.Now()

	step := now.Sub(c.last)
	total := now.Sub(c.start)

	stepStr := fmt.Sprintf("%8.3fms", float64(step.Microseconds())/1000)
	totalStr := fmt.Sprintf("%8.3fms", float64(total.Microseconds())/1000)

	log.Printf(
		"[耗时] %-25s step=%s total=%s",
		name,
		formatStep(step, stepStr),
		totalStr,
	)

	c.last = now
}
