// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/10/9
// 描述：
// *****************************************************************************

package vingo

import (
	"encoding/json"
	"sync"
)

type TreeV2[T float64 | string] struct {
	Rows     any
	IdName   string
	PidName  string
	Iteratee func(map[string]any) map[string]any

	mu          sync.RWMutex
	tree        []map[string]any
	initialized bool
}

// Build 生成树结构数据
func (s *TreeV2[T]) Build() []map[string]any {
	if s.IdName == "" {
		s.IdName = "id"
	}
	if s.PidName == "" {
		s.PidName = "pid"
	}

	// 保留现有输入兼容方式。
	b, err := json.Marshal(s.Rows)
	if err != nil {
		panic(err.Error())
	}

	var rows []map[string]any
	if err = json.Unmarshal(b, &rows); err != nil {
		panic(err.Error())
	}

	// 1. 建立父 ID -> 子节点列表的索引。
	childrenByPid := make(map[T][]map[string]any, len(rows))
	rootIds := make([]T, 0, len(rows))
	seenPid := make(map[T]struct{}, len(rows))

	for _, row := range rows {
		pid := row[s.PidName].(T)

		childrenByPid[pid] = append(childrenByPid[pid], row)

		// 保留原有 rootIds 的生成规则。
		if _, exists := seenPid[pid]; !exists {
			seenPid[pid] = struct{}{}
			rootIds = append(rootIds, pid)
		}
	}

	// 2. 递归构建树，避免重复处理父 ID。
	visited := make(map[T]struct{}, len(rows))

	var build func(id T) []map[string]any
	build = func(id T) []map[string]any {
		children := childrenByPid[id]
		if len(children) == 0 {
			return nil
		}

		// 防止异常循环引用。
		if _, exists := visited[id]; exists {
			return nil
		}
		visited[id] = struct{}{}

		result := make([]map[string]any, 0, len(children))

		for _, row := range children {
			// 与原逻辑一致：先执行自定义转换，再取 ID 递归。
			if s.Iteratee != nil {
				row = s.Iteratee(row)
			}

			nodeId := row[s.IdName].(T)
			childNodes := build(nodeId)

			row["hasChild"] = len(childNodes) > 0
			row["children"] = childNodes
			row["childCount"] = len(childNodes)

			// 直接累加子树总数。
			totalCount := 1
			for _, child := range childNodes {
				totalCount += int(child["totalCount"].(float64))
			}
			row["totalCount"] = float64(totalCount)

			result = append(result, row)
		}

		return result
	}

	// 3. 按原有 rootIds 顺序生成结果。
	result := make([]map[string]any, 0, len(rows))
	for _, rootId := range rootIds {
		if _, exists := visited[rootId]; exists {
			continue
		}
		result = append(result, build(rootId)...)
	}

	return result
}
