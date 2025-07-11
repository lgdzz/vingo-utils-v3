// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：
// *****************************************************************************

package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Strings String[T]：数据库中使用逗号字符串，JSON 中是数组
type Strings[T any] []T

func (s Strings[T]) Value() (driver.Value, error) {
	strs := make([]string, 0, len(s))
	for _, item := range s {
		strs = append(strs, toString(item))
	}
	return strings.Join(strs, ","), nil
}

func (s *Strings[T]) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		panic(fmt.Sprintf("String.Scan unsupported type: %T", v))
	}

	if strings.TrimSpace(str) == "" {
		*s = Strings[T]{}
		return nil
	}

	parts := strings.Split(str, ",")
	result := make([]T, 0, len(parts))
	for _, part := range parts {
		result = append(result, parsePrimitive[T](strings.TrimSpace(part)))
	}
	*s = result
	return nil
}

func (s Strings[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal([]T(s))
}

func (s *Strings[T]) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, (*[]T)(s)); err != nil {
		panic(fmt.Sprintf("String.UnmarshalJSON error: %v", err))
	}
	return nil
}

func parsePrimitive[T any](s string) T {
	var zero T
	kind := reflect.TypeOf(zero).Kind()

	switch kind {
	case reflect.String:
		return any(s).(T)
	case reflect.Int:
		n, err := strconv.Atoi(s)
		if err != nil {
			panic(fmt.Sprintf("parsePrimitive int error: %v", err))
		}
		return any(n).(T)
	case reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("parsePrimitive int64 error: %v", err))
		}
		return any(n).(T)
	case reflect.Uint:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("parsePrimitive uint error: %v", err))
		}
		return any(uint(n)).(T)
	case reflect.Float64:
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(fmt.Sprintf("parsePrimitive float64 error: %v", err))
		}
		return any(n).(T)
	default:
		panic(fmt.Sprintf("parsePrimitive unsupported type: %T", zero))
	}
}

func toString[T any](v T) string {
	switch val := any(v).(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", val)
	}
}
func (s Strings[T]) Slice() []T {
	return s
}
