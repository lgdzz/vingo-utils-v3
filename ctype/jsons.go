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
)

type Jsons[T any] []T

func (s Jsons[T]) Value() (driver.Value, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		panic(fmt.Sprintf("Jsons.Value marshal error: %v", err))
	}
	return bytes, nil
}

func (s *Jsons[T]) Scan(value any) error {
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		panic(fmt.Sprintf("Jsons.Scan unsupported type: %T", value))
	}
	if err := json.Unmarshal(b, s); err != nil {
		panic(fmt.Sprintf("Jsons.Scan unmarshal error: %v", err))
	}
	return nil
}

func (s Jsons[T]) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal([]T(s))
	if err != nil {
		panic(fmt.Sprintf("Jsons.MarshalJSON error: %v", err))
	}
	return bytes, nil
}

func (s *Jsons[T]) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, (*[]T)(s)); err != nil {
		panic(fmt.Sprintf("Jsons.UnmarshalJSON error: %v", err))
	}
	return nil
}

func (s Jsons[T]) Slice() []T {
	return s
}
