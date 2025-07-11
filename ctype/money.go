package ctype

import (
	"fmt"
	"math"
	"strings"
)

// Money 金额类型
type Money float64

var cnNums = [...]string{"零", "壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖"}
var cnIntUnits = [...]string{"", "拾", "佰", "仟"}
var cnIntRadice = [...]string{"", "万", "亿", "兆"}
var cnDecUnits = [...]string{"角", "分"}

// Round 保留两位小数
func (m Money) Round() Money {
	return Money(math.Round(float64(m)*100) / 100)
}

// Format 格式化显示，如：1,234.56
func (m Money) Format() string {
	val := m.Round()
	s := fmt.Sprintf("%.2f", val)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	decPart := parts[1]

	// 插入千位逗号
	var sb strings.Builder
	n := len(intPart)
	for i, ch := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			sb.WriteByte(',')
		}
		sb.WriteRune(ch)
	}
	return sb.String() + "." + decPart
}

// FormatWithSymbol 带货币符号，如 ￥1,234.56
func (m Money) FormatWithSymbol(symbol string) string {
	return symbol + m.Format()
}

// IsZero 判断金额是否为零
func (m Money) IsZero() bool {
	return float64(m) == 0
}

// ToChinese 金额转大写中文
func (m Money) ToChinese() string {
	val := m.Round()
	if val == 0 {
		return "零元整"
	}

	s := fmt.Sprintf("%.2f", val)
	parts := strings.Split(s, ".")
	integerPart := parts[0]
	decimalPart := parts[1]

	var result strings.Builder
	zeroCount := 0
	intLen := len(integerPart)

	// 整数部分
	for i := 0; i < intLen; i++ {
		p := intLen - i - 1
		num := integerPart[i] - '0'
		unitPos := p % 4
		radicePos := p / 4

		if num == 0 {
			zeroCount++
		} else {
			if zeroCount > 0 {
				result.WriteString("零")
			}
			result.WriteString(cnNums[num])
			result.WriteString(cnIntUnits[unitPos])
			zeroCount = 0
		}

		if unitPos == 0 && zeroCount < 4 {
			result.WriteString(cnIntRadice[radicePos])
		}
	}

	result.WriteString("元")

	// 小数部分
	jiao := decimalPart[0] - '0'
	fen := decimalPart[1] - '0'
	if jiao == 0 && fen == 0 {
		result.WriteString("整")
	} else {
		if jiao > 0 {
			result.WriteString(cnNums[jiao])
			result.WriteString(cnDecUnits[0])
		}
		if fen > 0 {
			result.WriteString(cnNums[fen])
			result.WriteString(cnDecUnits[1])
		}
	}

	return result.String()
}
