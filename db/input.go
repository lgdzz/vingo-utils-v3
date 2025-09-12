// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：常用输入参数
// *****************************************************************************

package db

import (
	"github.com/lgdzz/vingo-utils-v3/vingo"
	"strconv"
	"strings"
)

// TextSlice 文本切片字符串形态，如：a,b,c
type TextSlice string

// ToSlice 转换为切片
func (s TextSlice) ToSlice() []any {
	str := string(s)
	if str == "" {
		return []any{}
	}

	parts := strings.Split(str, ",")
	result := make([]any, 0)

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// 1. 先尝试转 int
		if i, err := strconv.ParseInt(part, 10, 64); err == nil {
			result = append(result, float64(i))
			continue
		}

		// 2. 尝试转 float64
		if f, err := strconv.ParseFloat(part, 64); err == nil {
			result = append(result, f)
			continue
		}

		// 3. 尝试转 bool
		if f, err := strconv.ParseBool(part); err == nil {
			result = append(result, f)
			continue
		}

		// 3. 否则默认当字符串处理
		result = append(result, part)
	}

	return result
}

// ToIntSlice 转换为int切片
func (s TextSlice) ToIntSlice() []int {
	result := make([]int, 0)
	str := string(s)
	if str == "" {
		return result
	}
	parts := strings.Split(str, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if i, err := strconv.ParseInt(part, 10, 64); err == nil {
			result = append(result, int(i))
			continue
		}
	}
	return result
}

// ToStringSlice 转换为string切片
func (s TextSlice) ToStringSlice() []string {
	str := string(s)
	if str == "" {
		return []string{}
	}
	return strings.Split(str, ",")
}

// IsEmpty 是否为空
func (s TextSlice) IsEmpty() bool {
	return string(s) == ""
}

// Id id参数
type Id struct {
	Id int `form:"id" json:"id"`
}

func (s Id) Int() int {
	return s.Id
}

func (s Id) String() string {
	return vingo.ToString(s.Id)
}

type UUID struct {
	UUID string `form:"uuid"`
}

type Keyword struct {
	Keyword string `form:"keyword" json:"keyword"`
}

// Location 定位坐标
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// DetailInput 详细记录参数
type DetailInput struct {
	Id
	Fetch TextSlice `form:"fetch"`
}

// UpdateFieldInput 更新字段参数
type UpdateFieldInput struct {
	Id
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type Between[T any] string

// Between 将逗号分隔的字符串转换为起始和结束值
func (s *Between[T]) Between() (T, T) {
	arr := strings.Split(string(*s), ",")
	if len(arr) != 2 {
		panic("范围字符串格式错误，正确格式：1,100")
	}
	return convertToT[T](strings.TrimSpace(arr[0])), convertToT[T](strings.TrimSpace(arr[1]))
}

// convertToT 将字符串转换为指定类型 T
func convertToT[T any](str string) T {
	var zero T
	switch any(zero).(type) {
	case float64:
		num, err := strconv.ParseFloat(str, 64)
		if err != nil {
			panic("转换 float64 类型失败: " + err.Error())
		}
		return any(num).(T)
	case int:
		num, err := strconv.Atoi(str)
		if err != nil {
			panic("转换 int 类型失败: " + err.Error())
		}
		return any(num).(T)
	case uint:
		num, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			panic("转换 uint 类型失败: " + err.Error())
		}
		return any(uint(num)).(T)
	case int64:
		num, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			panic("转换 int64 类型失败: " + err.Error())
		}
		return any(num).(T)
	default:
		panic("不支持的类型")
	}
}
