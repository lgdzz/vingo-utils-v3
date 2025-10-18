package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/lgdzz/vingo-utils-v3/vingo"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Option struct {
	Headers        *map[string]string
	Timeout        *int
	FileFieldName  *string
	FileOtherField *map[string]string
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

func Get(url string, opt Option) []byte {
	opt = NewOption(&opt)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	setHeaders(req, opt.Headers)
	return doRequest(req, *opt.Timeout)
}

func PostJSON(url string, body interface{}, opt Option) []byte {
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
	return doRequest(req, *opt.Timeout)
}

func PostFormData(url string, form map[string]string, opt Option) []byte {
	opt = NewOption(&opt)

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	for k, v := range form {
		_ = writer.WriteField(k, v)
	}
	_ = writer.Close()

	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	setHeaders(req, opt.Headers)

	return doRequest(req, *opt.Timeout)
}

func PostFormURLEncoded(urlPath string, form map[string]string, opt Option) []byte {
	opt = NewOption(&opt)

	data := url.Values{}
	for k, v := range form {
		data.Set(k, v)
	}
	req, err := http.NewRequest("POST", urlPath, strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	setHeaders(req, opt.Headers)

	return doRequest(req, *opt.Timeout)
}

func PostFile(url string, opt Option, filePath string) []byte {
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

	req, err := http.NewRequest("POST", url, body)
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

func DownloadFile(fileUrl string, saveDir string, randomName bool) string {
	parsedUrl, err := url.Parse(fileUrl)
	if err != nil {
		panic(err)
	}

	fileName := ""
	if randomName {
		fileName = vingo.GetUUID() + path.Ext(parsedUrl.Path)
	} else {
		fileName = path.Base(parsedUrl.Path)
	}

	err = os.MkdirAll(saveDir, 0777)
	if err != nil {
		panic(err)
	}

	savePath := filepath.Join(saveDir, fileName)

	out, err := os.Create(savePath)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(fileUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("server returned non-200: %v", resp.StatusCode))
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}
	return savePath
}

// -------------------- 内部工具方法 --------------------

func setHeaders(req *http.Request, headers *map[string]string) {
	if headers != nil {
		for k, v := range *headers {
			req.Header.Set(k, v)
		}
	}
}

func doRequest(req *http.Request, timeout int) []byte {
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
	return data
}
