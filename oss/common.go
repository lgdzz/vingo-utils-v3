// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/5/13
// 描述：
// *****************************************************************************

package oss

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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

	u, err := url.Parse(addr)
	if err != nil {
		panic(err)
	}
	query := u.Query()
	query.Set("_t", strconv.FormatInt(time.Now().UnixNano(), 10))
	u.RawQuery = query.Encode()

	data, resp := request.Get(u.String(), request.Option{Timeout: t})
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%d", resp.StatusCode), ""
	}
	mimeType := resp.Header.Get("Content-Type")
	base64Str := base64.StdEncoding.EncodeToString(data)
	return "data:" + mimeType + ";base64," + base64Str, mimeType
}

func ExtractObjectName(objectUrl string, bucket string) string {
	u, err := url.Parse(objectUrl)
	if err != nil {
		panic(fmt.Sprintf("文件地址提取objectName错误：%v", err.Error()))
	}

	// 去掉开头 /
	path := strings.TrimPrefix(u.Path, "/")

	// 去掉 bucket（minio需要）
	if bucket != "" {
		prefix := bucket + "/"
		path = strings.TrimPrefix(path, prefix)
	}

	// URL 解码
	res, err := url.PathUnescape(path)
	if err != nil {
		panic(fmt.Sprintf("文件地址提取objectName错误：%v", err.Error()))
	}

	return res
}
