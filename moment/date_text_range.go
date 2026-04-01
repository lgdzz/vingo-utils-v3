// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：
// *****************************************************************************

package moment

import (
	"strings"

	"github.com/lgdzz/vingo-utils-v3/vingo"
)

type DateTextRange string

// ToStruct 日期时间范围字符串转结构
func (s DateTextRange) ToStruct() DateRange {
	arr := strings.Split(string(s), ",")
	if len(arr) != 2 {
		panic("invalid date range format")
	}
	start := DateText(arr[0])
	end := DateText(arr[1])
	return DateRange{Start: *start.ToLocalTime(), End: *end.ToLocalTime()}
}

func (s DateTextRange) ToBetween() (*DateText, *DateText) {
	arr := strings.Split(string(s), ",")
	if len(arr) != 2 {
		panic("invalid date range format")
	}
	var start *DateText
	if arr[0] != "" {
		start = vingo.Of(DateText(arr[0]))
	}
	var end *DateText
	if arr[1] != "" {
		end = vingo.Of(DateText(arr[1]))
	}
	return start, end
}
