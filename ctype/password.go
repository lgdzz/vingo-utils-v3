// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/26
// 描述：
// *****************************************************************************

package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/random"
	"github.com/lgdzz/vingo-utils-v3/cryptor"
	"github.com/lgdzz/vingo-utils-v3/moment"
	"strings"
	"time"
	"unicode"
)

const (
	MinPasswordLen = 6  // 最小密码长度
	MaxPasswordLen = 25 // 最大密码长度

	PasswordWeak   = 1 // 简单密码
	PasswordMedium = 2 // 中等密码，任意两种字符组合
	PasswordStrong = 3 // 复杂密码，必须包含四种字符组合
)

type Password Ciphertext

type PasswordObj struct {
	PasswordMd5 string            `json:"passwordMd5"`
	Salt        string            `json:"salt"`
	CreatedAt   *moment.LocalTime `json:"createdAt,omitempty"`
	IsTemp      bool              `json:"isTemp"`
}

// ========== GORM 接口 ==========

func (s Password) Value() (driver.Value, error) {
	return Ciphertext(s).Value()
}

func (s *Password) Scan(value any) error {
	return (*Ciphertext)(s).Scan(value)
}

func (s Password) String() string {
	return string(s)
}

// NewPassword 生成密码（加密 + salt + 时间）
func NewPassword(raw string, level int, isTemp bool) Password {
	checkPasswordLevel(raw, level)
	salt := random.RandString(6)
	md5 := cryptor.Md5(cryptor.Md5(raw) + salt)
	obj := &PasswordObj{
		PasswordMd5: md5,
		Salt:        salt,
		CreatedAt:   moment.NowLocalTime(),
		IsTemp:      isTemp,
	}
	return mustMarshalPassword(obj)
}

// checkPasswordLevel 密码强度检查
func checkPasswordLevel(raw string, level int) {
	n := len(raw)

	if n < MinPasswordLen || n > MaxPasswordLen {
		panic(fmt.Sprintf("密码长度需符合 %d-%d 个字符长度要求", MinPasswordLen, MaxPasswordLen))
	}

	hasDigit, hasUpper, hasLower, hasSpecial := analyzeChars(raw)

	switch level {
	case PasswordMedium:
		if countTrue(hasDigit, hasUpper, hasLower, hasSpecial) < 2 {
			panic("密码需满足至少两种字符组合（数字、大写字母、小写字母、特殊符号）")
		}
	case PasswordStrong:
		if !(hasDigit && hasUpper && hasLower && hasSpecial) {
			panic("密码需满足四种字符组合（数字、大写字母、小写字母、特殊符号）")
		}
	}
}

// CreatedAt 获取密码创建时间
func (s Password) CreatedAt() *moment.LocalTime {
	o := s.getObj()
	return o.CreatedAt
}

// Match 判断密码是否等于指定原始密码
func (s Password) Match(raw string) bool {
	o := s.getObj()
	return o.PasswordMd5 == cryptor.Md5(cryptor.Md5(raw)+o.Salt)
}

// IsExpired 检查是否过期
func (s Password) IsExpired(day int) bool {
	o := s.getObj()
	if o.CreatedAt == nil && !o.IsTemp {
		return false
	}
	// 临时密码 或 超过有效期
	return o.IsTemp || time.Since(o.CreatedAt.Time()) > time.Duration(day)*24*time.Hour
}

func mustMarshalPassword(obj *PasswordObj) Password {
	bs, err := json.Marshal(obj)
	if err != nil {
		panic(fmt.Sprintf("密码序列化失败: %v", err))
	}
	return Password(bs)
}

func (s Password) getObj() *PasswordObj {
	str := strings.TrimSpace(string(s))
	if str == "" {
		return &PasswordObj{}
	}
	var obj PasswordObj
	if err := json.Unmarshal([]byte(str), &obj); err != nil {
		panic(fmt.Sprintf("密码格式错误，无法解析: %v", err))
	}
	return &obj
}

// 工具函数：字符分析
func analyzeChars(pass string) (hasDigit, hasUpper, hasLower, hasSpecial bool) {
	for _, ch := range pass {
		switch {
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsPunct(ch), unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}
	return
}

// 工具函数：统计 true 个数
func countTrue(flags ...bool) int {
	count := 0
	for _, f := range flags {
		if f {
			count++
		}
	}
	return count
}
