// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：
// *****************************************************************************

package moment

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeTextToTime 时间字符串转time.Time
func TimeTextToTime(timeText string, layouts ...string) time.Time {
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, timeText, time.Local); err == nil {
			return t
		}
	}
	panic("invalid date format")
}

// YesterdayStartTime 获取昨日开始时间
func YesterdayStartTime() time.Time {
	yesterday := time.Now().AddDate(0, 0, -1)
	return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
}

// YesterdayEndTime 获取昨日结束时间
func YesterdayEndTime() time.Time {
	yesterday := time.Now().AddDate(0, 0, -1)
	return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, yesterday.Location())
}

// BeforeNDates 返回最近 n 天的日期字符串（格式为 yyyy-MM-dd），默认到今天为止。
// 可选传入 startDay，指定结束日期，返回从 (startDay - n + 1) 到 startDay 的日期列表。
// Example BeforeNDates(7) 最近7日
// Example BeforeNDates(30) 最近30日
func BeforeNDates(n int, endDay ...time.Time) []string {
	if n <= 0 {
		return []string{}
	}

	end := time.Now()
	if len(endDay) > 0 {
		end = endDay[0]
	}

	start := end.AddDate(0, 0, -n+1)

	dates := make([]string, 0, n)
	for i := 0; i < n; i++ {
		day := start.AddDate(0, 0, i)
		dates = append(dates, day.Format(DateFormat))
	}
	return dates
}

// AfterNDates 返回未来 n 天的日期字符串（格式为 yyyy-MM-dd），默认从今天开始。
// 可选传入 startDay，指定结束日期，返回从 (startDay - n + 1) 到 startDay 的日期列表。
// Example AfterNDates(7) 未来7日
// Example AfterNDates(30) 未来30日
func AfterNDates(n int, startDay ...time.Time) []string {
	if n <= 0 {
		return []string{}
	}

	start := time.Now()
	if len(startDay) > 0 {
		start = startDay[0]
	}

	dates := make([]string, 0, n)
	for i := 0; i < n; i++ {
		day := start.AddDate(0, 0, i)
		dates = append(dates, day.Format(DateFormat))
	}
	return dates
}

// GetMonthRange 返回指定月份或当前月份的起止时间范围。
// 不传参默认获取当前月份；传入格式为 "2006-01" 的月份字符串将解析为指定月份。
// 若解析失败，将 panic。
func GetMonthRange(monthStr ...string) DateRange {
	var t time.Time
	var err error

	if len(monthStr) > 0 {
		t, err = time.ParseInLocation("2006-01", monthStr[0], time.Local)
		if err != nil {
			panic(fmt.Sprintf("GetMonthRange 解析失败: %v", err))
		}
	} else {
		t = time.Now()
	}

	year, month, _ := t.Date()
	loc := t.Location()
	start := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	end := time.Date(year, month+1, 0, 23, 59, 59, 0, loc)

	return DateRange{
		Start: *ToLocalTime(start),
		End:   *ToLocalTime(end),
	}
}

// GetQuarterRange 返回当前季度或指定季度的起止时间范围。
// 支持格式："2025-Q1"、"2025-Q2"、"2025-Q3"、"2025-Q4"
// 不传参数时默认返回当前季度；若格式错误，将 panic。
func GetQuarterRange(quarterStr ...string) DateRange {
	var year int
	var quarter int
	loc := time.Local

	if len(quarterStr) > 0 {
		// 解析格式 "2025-Q1"
		parts := strings.Split(quarterStr[0], "-Q")
		if len(parts) != 2 {
			panic(fmt.Sprintf("GetQuarterRange 格式错误，应为 yyyy-Qn，例如 2025-Q2"))
		}

		y, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(fmt.Sprintf("GetQuarterRange 年份解析失败: %v", err))
		}

		q, err := strconv.Atoi(parts[1])
		if err != nil || q < 1 || q > 4 {
			panic("GetQuarterRange 季度必须为 Q1~Q4")
		}

		year = y
		quarter = q
	} else {
		now := time.Now()
		year = now.Year()
		quarter = int((now.Month()-1)/3 + 1)
		loc = now.Location()
	}

	startMonth := time.Month((quarter-1)*3 + 1)
	start := time.Date(year, startMonth, 1, 0, 0, 0, 0, loc)
	end := time.Date(year, startMonth+3, 0, 23, 59, 59, 0, loc) // 当前季度最后一天的 23:59:59

	return DateRange{
		Start: *ToLocalTime(start),
		End:   *ToLocalTime(end),
	}
}

// GetYearRange 返回当前年份或指定年份的起止时间范围。
// 不传参数时默认返回当前年份；
// 传入格式为 "2025" 的年份字符串；若格式非法则 panic。
func GetYearRange(yearStr ...string) DateRange {
	var year int

	if len(yearStr) > 0 {
		y, err := strconv.Atoi(yearStr[0])
		if err != nil || y <= 0 {
			panic(fmt.Sprintf("GetYearRange 年份格式错误: %v", err))
		}
		year = y
	} else {
		year = time.Now().Year()
	}

	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(year, 12, 31, 23, 59, 59, 0, time.Local)

	return DateRange{
		Start: *ToLocalTime(start),
		End:   *ToLocalTime(end),
	}
}

// GetYesterdayRange 获取昨日时间范围
func GetYesterdayRange() DateRange {
	start := YesterdayStartTime()
	end := start.Add(time.Duration(DaySec-1) * time.Second)
	return DateRange{
		Start: *ToLocalTime(start),
		End:   *ToLocalTime(end),
	}
}

// GetTodayRange 获取今日时间范围
func GetTodayRange() DateRange {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(time.Duration(DaySec-1) * time.Second)
	return DateRange{
		Start: *ToLocalTime(start),
		End:   *ToLocalTime(end),
	}
}

// GetDateRange 获取指定日期时间范围
func GetDateRange(date string) DateRange {
	start, err := time.ParseInLocation(DateFormat, date, time.Local)
	if err != nil {
		panic(fmt.Sprintf("GetDateRange parse error: %v", err))
	}
	end := start.Add(time.Duration(DaySec-1) * time.Second)
	return DateRange{
		Start: *ToLocalTime(start),
		End:   *ToLocalTime(end),
	}
}

// 支持的时间范围类型: DateTextRange DateRange
func parseRange(ts any) (time.Time, time.Time) {
	switch v := ts.(type) {
	case DateTextRange:
		s := v.ToStruct()
		return time.Time(s.Start), time.Time(s.End)
	case DateRange:
		return time.Time(v.Start), time.Time(v.End)
	default:
		panic(fmt.Sprintf("不支持的时间范围类型: %T", ts))
	}
}

// GenerateDates 生成指定日期范围内的日期数组，格式为 yyyy-MM-dd
// 支持的时间范围类型: DateTextRange DateRange
func GenerateDates(ts any) []string {
	start, end := parseRange(ts)
	var result []string
	for t := start; !t.After(end); t = t.AddDate(0, 0, 1) {
		result = append(result, t.Format(DateFormat))
	}
	return result
}

// GenerateMonths 生成指定日期范围内的月份数组，格式为 yyyy-MM
// 支持的时间范围类型: DateTextRange DateRange
func GenerateMonths(ts any) []string {
	start, end := parseRange(ts)

	// 归整到月初
	start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())

	var result []string
	for t := start; !t.After(end); t = t.AddDate(0, 1, 0) {
		result = append(result, t.Format(YearMonthFormat))
	}
	return result
}

// GetDayFirstMoment 获取指定时间所在日期的初始一刻
func GetDayFirstMoment(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// GetDayLastMoment 获取指定时间所在日期的最后一刻
func GetDayLastMoment(t time.Time) time.Time {
	return GetDayFirstMoment(t).AddDate(0, 0, 1).Add(-time.Second)
}

// GetMonthFirstMoment 获取指定时间所在月份的初始一刻
func GetMonthFirstMoment(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// GetMonthLastMoment 获取指定时间所在月份的最后一刻
func GetMonthLastMoment(t time.Time) time.Time {
	return GetMonthFirstMoment(t).AddDate(0, 1, 0).Add(-time.Second)
}

// GetYearFirstMoment 获取指定时间所在年份的初始一刻
func GetYearFirstMoment(t time.Time) time.Time {
	return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
}

// GetYearLastMoment 获取指定时间所在年份的最后一刻
func GetYearLastMoment(t time.Time) time.Time {
	return GetYearFirstMoment(t).AddDate(1, 0, 0).Add(-time.Second)
}
