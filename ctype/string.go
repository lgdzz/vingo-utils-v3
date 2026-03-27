package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
	"strings"
)

type String string

// JSON 反序列化（接收时 trim）
func (s *String) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = String(strings.TrimSpace(str))
	return nil
}

// 入库时 trim
func (s String) Value() (driver.Value, error) {
	return strings.TrimSpace(string(s)), nil
}

// 基础方法
func (s String) String() string {
	return strings.TrimSpace(string(s))
}

func (s String) Int() int {
	i, err := strconv.Atoi(strings.TrimSpace(string(s)))
	if err != nil {
		panic(err)
	}
	return i
}

func (s String) Float() float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(string(s)), 64)
	if err != nil {
		panic(err)
	}
	return f
}
