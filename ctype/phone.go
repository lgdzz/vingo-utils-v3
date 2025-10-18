// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/26
// 描述：
// *****************************************************************************

package ctype

import (
	"fmt"
	"regexp"
)

// Phone 自定义手机号类型，底层为 string
type Phone string

// 手机号正则校验：1 开头 + 第二位 3-9 + 9 位数字（共11位）
var phoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

// IsValid 判断手机号是否合法（返回 bool）
func (p Phone) IsValid() bool {
	return phoneRegexp.MatchString(string(p))
}

// Validate 判断手机号是否合法（返回 error）
func (p Phone) Validate() error {
	if !p.IsValid() {
		return fmt.Errorf("无效手机号格式: %s", p)
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
	p.Check()
	return string(p[:3]) + "****" + string(p[7:])
}

// MaskMiddle 自定义掩码字符（如 MaskMiddle('#') => 133####8888）
func (p Phone) MaskMiddle(maskChar rune) string {
	p.Check()
	mask := string([]rune{maskChar, maskChar, maskChar, maskChar})
	return string(p[:3]) + mask + string(p[7:])
}

// Prefix 返回手机号前 3 位
func (p Phone) Prefix() string {
	if len(p) < 3 {
		return ""
	}
	return string(p[:3])
}

// Suffix 返回手机号后 4 位
func (p Phone) Suffix() string {
	if len(p) < 11 {
		return ""
	}
	return string(p[7:])
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
	return string(p)
}

// ToString 显式转换为 string
func (p Phone) ToString() string {
	return string(p)
}
