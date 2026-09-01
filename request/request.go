package request

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lgdzz/vingo-utils-v3/vingo"
)

type Option struct {
	Headers        *map[string]string
	Timeout        *int
	FileFieldName  *string
	FileOtherField *map[string]string
	Ctx            *context.Context
}

func NewOption(opt *Option) Option {
	def := Option{
		Timeout:       vingo.Of(30),
		FileFieldName: vingo.Of("file"),
	}
	if opt != nil {
		if opt.Headers != nil {
			def.Headers = opt.Headers
		}
		if opt.Timeout != nil {
			def.Timeout = opt.Timeout
		}
		if opt.FileFieldName != nil {
			def.FileFieldName = opt.FileFieldName
		}
		if opt.FileOtherField != nil {
			def.FileOtherField = opt.FileOtherField
		}
	}
	return def
}

func Get(url string, opt Option) ([]byte, *http.Response) {
	opt = NewOption(&opt)

	var req *http.Request
	var err error
	if opt.Ctx != nil {
		req, err = http.NewRequestWithContext(*opt.Ctx, "GET", url, nil)
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}

	if err != nil {
		panic(err)
	}
	setHeaders(req, opt.Headers)
	return doRequest(req, *opt.Timeout)
}

func PostJSON(url string, body interface{}, opt Option) ([]byte, *http.Response) {
	opt = NewOption(&opt)
	var requestBody []byte
	if body != nil {
		requestBody, _ = json.Marshal(body)
	}

	var req *http.Request
	var err error
	if opt.Ctx != nil {
		req, err = http.NewRequestWithContext(*opt.Ctx, "POST", url, bytes.NewBuffer(requestBody))
	} else {
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	}
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	setHeaders(req, opt.Headers)
	return doRequest(req, *opt.Timeout)
}

func PostFormData(url string, form map[string]string, opt Option) ([]byte, *http.Response) {
	opt = NewOption(&opt)

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	for k, v := range form {
		_ = writer.WriteField(k, v)
	}
	_ = writer.Close()

	var req *http.Request
	var err error
	if opt.Ctx != nil {
		req, err = http.NewRequestWithContext(*opt.Ctx, "POST", url, &requestBody)
	} else {
		req, err = http.NewRequest("POST", url, &requestBody)
	}
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	setHeaders(req, opt.Headers)

	return doRequest(req, *opt.Timeout)
}

func PostFormURLEncoded(urlPath string, form map[string]string, opt Option) ([]byte, *http.Response) {
	opt = NewOption(&opt)

	data := url.Values{}
	for k, v := range form {
		data.Set(k, v)
	}
	var req *http.Request
	var err error
	if opt.Ctx != nil {
		req, err = http.NewRequestWithContext(*opt.Ctx, "POST", urlPath, strings.NewReader(data.Encode()))
	} else {
		req, err = http.NewRequest("POST", urlPath, strings.NewReader(data.Encode()))
	}
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	setHeaders(req, opt.Headers)

	return doRequest(req, *opt.Timeout)
}

func PostFile(url string, opt Option, filePath string) ([]byte, *http.Response) {
	opt = NewOption(&opt)

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 附加字段
	if opt.FileOtherField != nil {
		for k, v := range *opt.FileOtherField {
			_ = writer.WriteField(k, v)
		}
	}

	fileName := filepath.Base(filePath)
	part, err := writer.CreateFormFile(*opt.FileFieldName, fileName)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		panic(err)
	}
	_ = writer.Close()

	var req *http.Request
	if opt.Ctx != nil {
		req, err = http.NewRequestWithContext(*opt.Ctx, "POST", url, body)
	} else {
		req, err = http.NewRequest("POST", url, body)
	}
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	setHeaders(req, opt.Headers)

	return doRequest(req, *opt.Timeout)
}

func PostJSONStream(url string, body interface{}, opt Option, receive func(...byte)) {
	opt = NewOption(&opt)

	var requestBody []byte
	if body != nil {
		requestBody, _ = json.Marshal(body)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	setHeaders(req, opt.Headers)

	client := &http.Client{
		Timeout: time.Duration(*opt.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}
		receive(buf[:n]...)
	}
}

// -------------------- 内部工具方法 --------------------

func setHeaders(req *http.Request, headers *map[string]string) {
	if headers != nil {
		for k, v := range *headers {
			req.Header.Set(k, v)
		}
	}
}

func doRequest(req *http.Request, timeout int) ([]byte, *http.Response) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	//contentType := resp.Header.Get("Content-Type")
	return data, resp
}
