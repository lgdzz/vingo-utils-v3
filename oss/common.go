// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/5/13
// 描述：
// *****************************************************************************

package oss

import (
	"encoding/base64"
	"net/http"

	"github.com/lgdzz/vingo-utils-v3/request"
	"github.com/lgdzz/vingo-utils-v3/vingo"
)

func GetImageBase64(addr string, timeout ...int) (string, string) {
	return GetBase64(addr, timeout...)
}

func GetBase64(addr string, timeout ...int) (string, string) {
	t := vingo.Of(60)
	if len(timeout) > 0 {
		t = vingo.Of(timeout[0])
	}
	data := request.Get(addr, request.Option{Timeout: t})
	mimeType := http.DetectContentType(data)
	base64Str := base64.StdEncoding.EncodeToString(data)
	return "data:" + mimeType + ";base64," + base64Str, mimeType
}
