package moment

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/lgdzz/vingo-utils-v3/vingo"
)

type LocalTime time.Time

// ToLocalTime 转成LocalTime
func ToLocalTime[T time.Time | ~string](value T) *LocalTime {
	switch v := any(value).(type) {
	case time.Time:
		return vingo.Of(LocalTime(v.Local()))
	case string:
		for _, format := range layoutFormats {
			if t, err := time.ParseInLocation(format, v, time.Local); err == nil {
				return vingo.Of(LocalTime(t))
			}
		}
		panic(fmt.Sprintf("NewLocalTime string 解析失败，格式应为: yyyy-MM-dd[ HH:mm[:ss]]，实际: %v", v))
	default:
		panic(fmt.Sprintf("NewLocalTime 不支持的类型: %T", value))
	}
}

// NowLocalTime 当前时间
func NowLocalTime() *LocalTime {
	return vingo.Of(LocalTime(time.Now()))
}

// TodayLocalTime 今天凌晨LocalTime
func TodayLocalTime() *LocalTime {
	now := time.Now()
	return vingo.Of(LocalTime(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())))
}

// YesterdayLocalTime 昨日凌晨LocalTime
func YesterdayLocalTime() *LocalTime {
	now := time.Now().AddDate(0, 0, -1)
	return vingo.Of(LocalTime(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())))
}

func (t LocalTime) MarshalJSON() ([]byte, error) {
	return []byte("\"" + time.Time(t).Format(DateTimeFormat) + "\""), nil
}

func (t *LocalTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "null" || str == "" {
		*t = LocalTime{}
		return nil
	}
	timeParsed, err := time.ParseInLocation(DateTimeFormat, str, time.Local)
	if err != nil {
		panic(fmt.Sprintf("UnmarshalJSON parse error: %v", err))
	}
	*t = LocalTime(timeParsed)
	return nil
}

func (t LocalTime) Value() (driver.Value, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return time.Time(t), nil
}

func (t *LocalTime) Scan(v interface{}) error {
	switch value := v.(type) {
	case time.Time:
		*t = LocalTime(value.Local())
	case string:
		parsed, err := time.ParseInLocation(DateTimeFormat, value, time.Local)
		if err != nil {
			panic(fmt.Sprintf("Scan string parse error: %v", err))
		}
		*t = LocalTime(parsed)
	case []byte:
		parsed, err := time.ParseInLocation(DateTimeFormat, string(value), time.Local)
		if err != nil {
			panic(fmt.Sprintf("Scan []byte parse error: %v", err))
		}
		*t = LocalTime(parsed)
	default:
		panic(fmt.Sprintf("unsupported type %T in LocalTime Scan", v))
	}
	return nil
}

//func (t LocalTime) GormDataType() string {
//	return "datetime" // 或者 "timestamp"
//}

func (t LocalTime) Now() LocalTime {
	return LocalTime(time.Now().Local())
}

func (t *LocalTime) SetNow() {
	*t = LocalTime(time.Now().Local())
}

func (t *LocalTime) To(value time.Time) {
	*t = LocalTime(value.Local())
}

func (t LocalTime) String() string {
	return time.Time(t).Format(DateTimeFormat)
}

func (t LocalTime) Time() time.Time {
	return time.Time(t)
}

func (t LocalTime) Format(layout string) string {
	return t.Time().Format(layout)
}

func (t *LocalTime) ScanFromRow(rows *sql.Rows, columnName string) error {
	var tmp time.Time
	err := rows.Scan(&tmp)
	if err != nil {
		return err
	}
	*t = LocalTime(tmp.Local())
	return nil
}

// 是否是今天
func (t LocalTime) IsToday() bool {
	now := time.Now()
	y1, m1, d1 := now.Date()
	y2, m2, d2 := time.Time(t).Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// 添加天数，负数则减
func (t LocalTime) AddDays(days int) LocalTime {
	return LocalTime(time.Time(t).AddDate(0, 0, days))
}

// 添加天数，可附加时间
func (t LocalTime) AddDaysWithTime(days, hour, min, sec int) LocalTime {
	newTime := time.Time(t).AddDate(0, 0, days)
	y, m, d := newTime.Date()
	return LocalTime(time.Date(y, m, d, hour, min, sec, 0, newTime.Location()))
}

// GtNow t>当前时间
func (t LocalTime) GtNow() bool {
	return time.Time(t).After(time.Now())
}

// LtNow t<当前时间
func (t LocalTime) LtNow() bool {
	return time.Now().After(time.Time(t))
}

// Gt t>o
func (t LocalTime) Gt(o LocalTime) bool {
	return time.Time(t).After(time.Time(o))
}

// Lt t<o
func (t LocalTime) Lt(o LocalTime) bool {
	return time.Time(o).After(time.Time(t))
}
