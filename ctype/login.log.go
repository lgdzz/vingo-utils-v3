// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/4/27
// 描述：
// *****************************************************************************

package ctype

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/lgdzz/vingo-utils-v3/moment"
	"github.com/lgdzz/vingo-utils-v3/vingo"
)

var ActiveLoginDay = 7

type LoginRecord struct {
	Time moment.LocalTime `json:"time"`
	IP   IP               `json:"ip"`
}

type LoginLog struct {
	Last    *LoginRecord `json:"last"`
	Current *LoginRecord `json:"current"`
}

func (s *LoginLog) Value() (driver.Value, error) {
	output, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(output), nil
}

func (s *LoginLog) Scan(value any) error {
	if value == nil {
		*s = LoginLog{}
		return nil
	}

	var result LoginLog

	switch v := value.(type) {
	case string:
		vingo.StringToJson(v, &result)
	case []byte:
		vingo.StringToJson(string(v), &result)
	default:
		return nil
	}

	*s = result
	return nil
}

// Modify 更新记录
func (s *LoginLog) Modify(ip string) {
	s.Last = s.Current
	s.Current = &LoginRecord{
		Time: *moment.NowLocalTime(),
		IP:   IP(ip),
	}
}

// IsActive 是否活跃
func (s *LoginLog) IsActive() bool {
	return s != nil && s.Current != nil && s.Current.Time.AddDays(ActiveLoginDay).GtNow()
}
