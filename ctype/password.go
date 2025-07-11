// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/26
// 描述：
// *****************************************************************************

package ctype

import (
	"database/sql/driver"
	"fmt"
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

// ========== 密码强度检查 ==========

func (s Password) CheckLevel(level int) {
	passwordPart, _ := s.Split()
	n := len(passwordPart)

	if n < MinPasswordLen || n > MaxPasswordLen {
		panic(fmt.Sprintf("密码长度需符合 %d-%d 个字符长度要求", MinPasswordLen, MaxPasswordLen))
	}

	hasDigit, hasUpper, hasLower, hasSpecial := analyzeChars(passwordPart)

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

// MarkTime 标记密码创建时间
func (s *Password) MarkTime() {
	if _, t := s.Split(); t == nil {
		*s = Password(fmt.Sprintf("%s,%s", s.String(), time.Now().Format(moment.DateTimeFormat)))
	}
}

// Split 分割密码与时间
func (s Password) Split() (password string, createdAt *moment.LocalTime) {
	parts := strings.SplitN(string(s), ",", 2)
	password = parts[0]
	if len(parts) == 2 {
		createdAt = moment.ToLocalTime(parts[1])
	}
	return
}

// CreatedAt 获取创建时间（可空）
func (s Password) CreatedAt() *moment.LocalTime {
	_, t := s.Split()
	return t
}

// Match 判断密码是否等于指定原始密码（不包含时间部分）
func (s Password) Match(raw string) bool {
	password, _ := s.Split()
	return password == raw
}

// IsExpired 检查密码是否已过期（从创建时间起超过 day）
func (s Password) IsExpired(day int) bool {
	createdAt := s.CreatedAt()
	if createdAt == nil {
		return false // 没有创建时间，视为永不过期
	}
	return time.Since(createdAt.Time()) > time.Duration(day)*24*time.Hour
}

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

func countTrue(flags ...bool) int {
	count := 0
	for _, f := range flags {
		if f {
			count++
		}
	}
	return count
}
