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
	v := float64(r)
	if math.IsInf(v, 0) || math.IsNaN(v) {
		v = 0
	}
	return v, nil
}

// Scan 从数据库读取，支持 float64/int/string/[]byte
func (r *Ratio) Scan(value any) error {
	switch v := value.(type) {
	case float64:
		r.safeSet(v)
	case int:
		r.safeSet(float64(v))
	case []byte:
		f, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			panic(fmt.Sprintf("Ratio.Scan parse error: %v", err))
		}
		r.safeSet(f)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			panic(fmt.Sprintf("Ratio.Scan parse error: %v", err))
		}
		r.safeSet(f)
	default:
		panic(fmt.Sprintf("Ratio.Scan unsupported type: %T", value))
	}
	return nil
}

func (r *Ratio) safeSet(f float64) {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		*r = 0
	} else if f == 0 {
		*r = 0
	} else {
		*r = Ratio(f)
	}
}

// MarshalJSON 默认返回百分比（如 20）
func (r Ratio) MarshalJSON() ([]byte, error) {
	p := r.Percent()

	// 防止无效值 (Inf 或 NaN)
	if math.IsInf(p, 0) || math.IsNaN(p) {
		p = 0
	}

	// 保留两位小数
	p = math.Round(p*100) / 100

	return json.Marshal(p)
}

// UnmarshalJSON 反序列化，支持百分比或小数
func (r *Ratio) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || len(data) == 0 {
		*r = 0
		return nil
	}
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("Ratio.UnmarshalJSON unmarshal error: %w", err)
	}
	switch v := temp.(type) {
	case float64:
		if v > 1 {
			r.safeSet(v / 100)
		} else {
			r.safeSet(v)
		}
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("Ratio.UnmarshalJSON parse error: %w", err)
		}
		if f > 1 {
			r.safeSet(f / 100)
		} else {
			r.safeSet(f)
		}
	default:
		return fmt.Errorf("Ratio.UnmarshalJSON unsupported type: %T", temp)
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
	if math.IsInf(res, 0) || math.IsNaN(res) {
		return 0
	}
	return math.Round(res*100) / 100
}
