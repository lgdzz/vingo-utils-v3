// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/23
// 描述：
// *****************************************************************************

package ctype

import (
	"strconv"
	"strings"
)

// PathInterface 泛型接口约束
type PathInterface interface {
	int | string
}

type Path[T PathInterface] string

// 获取分隔符
func (s Path[T]) getDelimiter(sep ...string) string {
	if len(sep) > 0 && sep[0] != "" {
		return sep[0]
	}
	return ","
}

// Slice 转换为切片
func (s Path[T]) Slice(sep ...string) []T {
	delimiter := s.getDelimiter(sep...)
	parts := strings.Split(string(s), delimiter)

	result := make([]T, 0, len(parts))
	var zero T

	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		switch any(zero).(type) {
		case int:
			num, err := strconv.Atoi(item)
			if err != nil {
				panic("Path.Slice: invalid int value: " + item)
			}
			result = append(result, T(num))
		case string:
			result = append(result, any(item).(T))
		default:
			panic("Path.Slice: unsupported type in PathInterface")
		}
	}
	return result
}

// Len 获取元素数量
func (s Path[T]) Len(sep ...string) int {
	return len(s.Slice(sep...))
}

// IsEmpty 是否为空
func (s Path[T]) IsEmpty() bool {
	return s.Len() == 0
}

// Contains 是否包含某个值
func (s Path[T]) Contains(target T, sep ...string) bool {
	for _, item := range s.Slice(sep...) {
		if item == target {
			return true
		}
	}
	return false
}

// First 获取第一个元素
func (s Path[T]) First(sep ...string) (T, bool) {
	items := s.Slice(sep...)
	if len(items) == 0 {
		var zero T
		return zero, false
	}
	return items[0], true
}

// Last 获取最后一个元素
func (s Path[T]) Last(sep ...string) (T, bool) {
	items := s.Slice(sep...)
	if len(items) == 0 {
		var zero T
		return zero, false
	}
	return items[len(items)-1], true
}

// Parents 获取除最后一个元素的所有元素
func (s Path[T]) Parents(sep ...string) ([]T, bool) {
	items := s.Slice(sep...)
	if len(items) <= 1 {
		return nil, false
	}
	return items[:len(items)-1], true
}
