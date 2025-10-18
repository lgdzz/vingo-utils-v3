// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：
// *****************************************************************************

package moment

import "time"

type DateRange struct {
	Start LocalTime
	End   LocalTime
}

// BetweenText 日期时间范围字符串
func (s DateRange) BetweenText() (start, end string) {
	return s.Start.String(), s.End.String()
}

// BetweenTime 日期时间范围时间
func (s DateRange) BetweenTime() (start, end time.Time) {
	return s.Start.Time(), s.End.Time()
}
