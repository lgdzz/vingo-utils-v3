// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/9/5
// 描述：比率类型 (数据库存储小数，前端显示百分数)
// *****************************************************************************

package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

type Ratio float64

// Value 存储到数据库时，保持为小数（如 0.2）
func (r Ratio) Value() (driver.Value, error) {
	return float64(r), nil
}

// Scan 从数据库读取，支持 float64/int/string/[]byte
func (r *Ratio) Scan(value any) error {
	switch v := value.(type) {
	case float64:
		*r = Ratio(v)
	case int: // 转成 float64
		*r = Ratio(float64(v))
	case []byte:
		f, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			panic(fmt.Sprintf("Ratio.Scan parse error: %v", err))
		}
		*r = Ratio(f)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			panic(fmt.Sprintf("Ratio.Scan parse error: %v", err))
		}
		*r = Ratio(f)
	default:
		panic(fmt.Sprintf("Ratio.Scan unsupported type: %T", value))
	}
	return nil
}

// MarshalJSON 默认返回百分比（如 20）
func (r Ratio) MarshalJSON() ([]byte, error) {
	// 保留两位小数
	p := math.Round(r.Percent()*100) / 100

	// 转成数字（不是字符串）
	result, err := json.Marshal(p)
	if err != nil {
		panic(fmt.Sprintf("Ratio.MarshalJSON error: %v", err))
	}
	return result, nil
}

// UnmarshalJSON 反序列化，支持百分比或小数
func (r *Ratio) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || len(data) == 0 {
		*r = 0
		return nil
	}

	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		panic(fmt.Sprintf("Ratio.UnmarshalJSON unmarshal error: %v", err))
	}
	switch v := temp.(type) {
	case float64:
		// 判断是小数还是百分比（规则：>1 当百分比）
		if v > 1 {
			*r = Ratio(v / 100)
		} else {
			*r = Ratio(v)
		}
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			panic(fmt.Sprintf("Ratio.UnmarshalJSON parse error: %v", err))
		}
		if f > 1 {
			*r = Ratio(f / 100)
		} else {
			*r = Ratio(f)
		}
	default:
		panic(fmt.Sprintf("Ratio.UnmarshalJSON unsupported type: %T", temp))
	}
	return nil
}

// Decimal 返回小数（如 0.2）
func (r Ratio) Decimal() float64 {
	return float64(r)
}

// Percent 返回百分数（如 20）
func (r Ratio) Percent() float64 {
	return float64(r) * 100
}

// Mul 用比率乘以一个数值，结果保留两位小数
func (r Ratio) Mul(value float64) float64 {
	res := float64(r) * value
	return math.Round(res*100) / 100 // 保留两位小数
}
