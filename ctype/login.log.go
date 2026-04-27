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

type LoginLog struct {
	LastTime *moment.LocalTime `json:"lastTime"`
	LastIP   *IP               `json:"lastIP"`

	CurrentTime *moment.LocalTime `json:"currentTime"`
	CurrentIP   *IP               `json:"currentIP"`
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

func (s *LoginLog) Modify(ip string) {
	s.LastTime = s.CurrentTime
	s.LastIP = s.CurrentIP
	s.CurrentTime = moment.NowLocalTime()
	s.CurrentIP = vingo.Of(IP(ip))
}
