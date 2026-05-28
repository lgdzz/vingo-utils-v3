// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/26
// 描述：
// *****************************************************************************

package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

// Phone 自定义手机号类型，底层为 string
type Phone string

// 手机号正则校验：1 开头 + 第二位 3-9 + 9 位数字（共11位）
var phoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

var ErrInvalidPhone = errors.New("手机号不正确")

// trim 去除首尾空白字符
func (p Phone) trim() string {
	return strings.TrimSpace(string(p))
}

// Value 写入数据库时调用
func (p Phone) Value() (driver.Value, error) {
	v := p.trim()

	// 允许空值
	if v == "" {
		return "", nil
	}

	if !Phone(v).IsValid() {
		return nil, ErrInvalidPhone
	}

	return v, nil
}

// UnmarshalJSON gin/json 反序列化时自动 trim
func (p *Phone) UnmarshalJSON(data []byte) error {
	var v string

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	v = strings.TrimSpace(v)

	// JSON阶段直接校验
	if v != "" && !Phone(v).IsValid() {
		return ErrInvalidPhone
	}

	*p = Phone(v)
	return nil
}

// MarshalJSON 序列化时输出 trim 后内容
func (p Phone) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.trim())
}

// IsValid 判断手机号是否合法（返回 bool）
func (p Phone) IsValid() bool {
	return phoneRegexp.MatchString(p.trim())
}

// Validate 判断手机号是否合法（返回 error）
func (p Phone) Validate() error {
	if !p.IsValid() {
		return ErrInvalidPhone
	}
	return nil
}

// Check 快捷校验（不合法直接 panic）
func (p Phone) Check() {
	if err := p.Validate(); err != nil {
		panic(err)
	}
}

// Mask 返回默认掩码手机号（如 133****8888）
func (p Phone) Mask() string {
	s := p.trim()
	Phone(s).Check()
	return s[:3] + "****" + s[7:]
}

// MaskMiddle 自定义掩码字符（如 MaskMiddle('#') => 133####8888）
func (p Phone) MaskMiddle(maskChar rune) string {
	s := p.trim()
	Phone(s).Check()

	mask := strings.Repeat(string(maskChar), 4)
	return s[:3] + mask + s[7:]
}

// Prefix 返回手机号前 3 位
func (p Phone) Prefix() string {
	s := p.trim()
	if len(s) < 3 {
		return ""
	}
	return s[:3]
}

// Suffix 返回手机号后 4 位
func (p Phone) Suffix() string {
	s := p.trim()
	if len(s) < 11 {
		return ""
	}
	return s[7:]
}

// Carrier 简单判断运营商
func (p Phone) Carrier() string {
	switch p.Prefix() {
	case "133", "153", "180", "181", "189", "199":
		return "电信"
	case "130", "131", "132", "155", "156", "185", "186", "166":
		return "联通"
	case "134", "135", "136", "137", "138", "139",
		"150", "151", "152", "157", "158", "159",
		"182", "183", "184", "187", "188", "198":
		return "移动"
	default:
		return "未知"
	}
}

// String 实现 fmt.Stringer 接口
func (p Phone) String() string {
	return p.trim()
}

// ToString 显式转换为 string
func (p Phone) ToString() string {
	return p.trim()
}
