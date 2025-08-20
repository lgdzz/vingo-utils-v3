// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：布尔类型
// *****************************************************************************

package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Bool bool

func (b Bool) Value() (driver.Value, error) {
	// 统一写入为布尔类型，MySQL 会处理成 tinyint(1)，PG 为 boolean
	return bool(b), nil
}

func (b *Bool) Scan(value any) error {
	switch v := value.(type) {
	case bool:
		*b = Bool(v)
	case int64:
		*b = v != 0
	case []byte:
		s := string(v)
		*b = s == "1" || s == "true" || s == "TRUE"
	case string:
		*b = v == "1" || v == "true" || v == "TRUE"
	default:
		panic(fmt.Sprintf("Bool.Scan unsupported type: %T", value))
	}
	return nil
}

func (b Bool) MarshalJSON() ([]byte, error) {
	result, err := json.Marshal(bool(b))
	if err != nil {
		panic(fmt.Sprintf("Bool.MarshalJSON error: %v", err))
	}
	return result, nil
}

func (b *Bool) UnmarshalJSON(data []byte) error {
	// 如果为空或 null，直接设为 false
	if string(data) == "null" || len(data) == 0 {
		*b = False
		return nil
	}

	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		panic(fmt.Sprintf("Bool.UnmarshalJSON unmarshal error: %v", err))
	}
	switch v := temp.(type) {
	case bool:
		*b = Bool(v)
	case float64:
		*b = v != 0
	case string:
		*b = v == "1" || v == "true" || v == "TRUE"
	default:
		*b = False
	}
	return nil
}

const (
	True  Bool = true
	False Bool = false
)
