package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Moneys 表示金额范围，数据库以逗号分隔的字符串保存，JSON 中为数组
type Moneys string

// Between 将逗号分隔的字符串转换为起始和结束 Money
func (s *Moneys) Between() (Money, Money) {
	str := strings.TrimSpace(string(*s))
	if str == "" {
		return Money(0), Money(0)
	}
	arr := strings.Split(str, ",")
	if len(arr) != 2 {
		panic("范围字符串格式错误，正确格式：1,100")
	}
	a := strings.TrimSpace(arr[0])
	b := strings.TrimSpace(arr[1])
	fa, err := strconv.ParseFloat(a, 64)
	if err != nil {
		panic(fmt.Sprintf("parse money error: %v", err))
	}
	fb, err := strconv.ParseFloat(b, 64)
	if err != nil {
		panic(fmt.Sprintf("parse money error: %v", err))
	}
	return Money(fa), Money(fb)
}

// Value 将 Moneys 存入数据库（字符串）
func (s Moneys) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan 从数据库读取
func (s *Moneys) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		panic(fmt.Sprintf("Moneys.Scan unsupported type: %T", v))
	}
	*s = Moneys(str)
	return nil
}

// MarshalJSON 将 Moneys 转为 JSON 数组 [start,end]
func (s Moneys) MarshalJSON() ([]byte, error) {
	if strings.TrimSpace(string(s)) == "" {
		return json.Marshal([]Money{})
	}
	arr := strings.Split(string(s), ",")
	res := make([]Money, 0, len(arr))
	for _, part := range arr {
		p := strings.TrimSpace(part)
		if p == "" {
			res = append(res, Money(0))
			continue
		}
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("Moneys.MarshalJSON parse float error: %v", err)
		}
		res = append(res, Money(f))
	}
	return json.Marshal(res)
}

// UnmarshalJSON 支持两种 JSON 格式：数组（[1.2,3.4]）或字符串（"1.2,3.4"）
func (s *Moneys) UnmarshalJSON(b []byte) error {
	str := strings.TrimSpace(string(b))
	if str == "null" || str == `""` || len(str) == 0 {
		*s = Moneys("")
		return nil
	}
	// 尝试解析为数组数字
	var nums []float64
	if err := json.Unmarshal(b, &nums); err == nil {
		parts := make([]string, 0, len(nums))
		for _, n := range nums {
			parts = append(parts, strconv.FormatFloat(n, 'f', -1, 64))
		}
		*s = Moneys(strings.Join(parts, ","))
		return nil
	}
	// 尝试解析为字符串
	var sstr string
	if err := json.Unmarshal(b, &sstr); err == nil {
		*s = Moneys(sstr)
		return nil
	}
	return fmt.Errorf("Moneys.UnmarshalJSON unsupported format: %s", str)
}

func (s *Moneys) String() string {
	if s == nil {
		return ""
	}
	return string(*s)
}

func (s *Moneys) IsEmpty() bool {
	return s == nil || strings.TrimSpace(string(*s)) == ""
}

// ToSlice 将 Moneys 转为 []Money（如果为空返回空切片）
func (s *Moneys) ToSlice() []Money {
	if s == nil || strings.TrimSpace(string(*s)) == "" {
		return []Money{}
	}
	parts := strings.Split(string(*s), ",")
	res := make([]Money, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			res = append(res, Money(0))
			continue
		}
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			panic(fmt.Sprintf("Moneys.ToSlice parse float error: %v", err))
		}
		res = append(res, Money(f))
	}
	return res
}
