// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：
// *****************************************************************************

package moment

import "strings"

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
