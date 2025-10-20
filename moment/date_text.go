// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：
// *****************************************************************************

package moment

import (
	"strings"
	"time"
)

// 支持的字符串格式
var layoutFormats = []string{
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
	"2006-01",
	"2006-1-2 15:4:5",
	"2006-1-2 15:4",
	"2006-1-2",
	"2006-1",
	"2006",
}

// DateText 日期格式字符串
type DateText string

func (s DateText) ToString() string {
	return strings.TrimSpace(string(s))
}

// ToTime 日期字符串转 time.Time
func (s DateText) ToTime() time.Time {
	return TimeTextToTime(s.ToString(), layoutFormats...)
}

// ToLocalTime 日期字符串转 LocalTime
func (s DateText) ToLocalTime() *LocalTime {
	return ToLocalTime(s.ToString())
}
