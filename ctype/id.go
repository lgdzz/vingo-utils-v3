package ctype

import (
	"encoding/json"
	"fmt"
)

type Id struct {
	Value any
}

func (f *Id) UnmarshalJSON(b []byte) error {
	// 尝试解析为 int
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		f.Value = i
		return nil
	}

	// 尝试解析为 string
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		f.Value = s
		return nil
	}

	return fmt.Errorf("invalid id format")
}
