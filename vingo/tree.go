// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/27
// 描述：树结构数据

// example:
// data := []Dept{
//		{Id: 1, Name: "总部", Pid: 0, Path: "1"},
//		{Id: 2, Name: "技术部", Pid: 1, Path: "1,2"},
//		{Id: 3, Name: "产品部", Pid: 1, Path: "1,3"},
//	}
//
//	var tree = Tree[float64]{
//		Rows: data,
//	}
//
//	vingo.Print(tree.Build())

// *****************************************************************************

package vingo

import (
	"encoding/json"
	"github.com/duke-git/lancet/v2/slice"
)

type Tree[T float64 | string] struct {
	Rows     any
	IdName   string
	PidName  string
	Iteratee func(map[string]any) map[string]any
}

// FastTree 快速生成树结构数据
//
//	@param rows 数据
//	@return []map[string]any 树结构数据
func FastTree[T float64 | string](rows any) []map[string]any {
	tree := Tree[T]{
		Rows: rows,
	}
	return tree.Build()
}

// Build 生成树结构数据
func (s *Tree[T]) Build() []map[string]any {
	if s.IdName == "" {
		s.IdName = "id"
	}
	if s.PidName == "" {
		s.PidName = "pid"
	}

	var rows []map[string]any
	b, err := json.Marshal(s.Rows)
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(b, &rows)
	if err != nil {
		panic(err.Error())
	}

	var rootIds = make([]T, 0)
	for _, row := range rows {
		if !slice.Contain(rootIds, row[s.PidName].(T)) {
			rootIds = append(rootIds, row[s.PidName].(T))
		}
	}

	return s.start(&rows, rootIds)
}

func (s *Tree[T]) start(list *[]map[string]any, rootIds []T) []map[string]any {
	result := make([]map[string]any, 0)
	already := make([]T, 0)
	for _, rootId := range rootIds {
		if slice.Contain(already, rootId) {
			continue
		}
		result = append(result, s.item(list, rootId, &already)...)
	}
	return result
}

func (s *Tree[T]) item(list *[]map[string]any, id T, already *[]T) (result []map[string]any) {
	for _, row := range *list {

		if row[s.PidName].(T) != id {
			continue
		}

		*already = append(*already, id)

		if s.Iteratee != nil {
			row = s.Iteratee(row)
		}

		children := s.item(list, row[s.IdName].(T), already)

		childCount := len(children)
		if childCount > 0 {
			row["hasChild"] = true
			row["children"] = children

			row["childCount"] = childCount
			// 递归计算总数
			childTotalCount := 1
			for _, child := range children {
				childTotalCount += int(child["totalCount"].(float64))
			}
			row["totalCount"] = float64(childTotalCount)
		} else {
			row["hasChild"] = false

			row["childCount"] = 0
			row["totalCount"] = 1.0 // 如果没有子节点，只计数自身
		}
		result = append(result, row)
	}

	return
}
