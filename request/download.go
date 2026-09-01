// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/9/1
// 描述：
// *****************************************************************************

package request

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/lgdzz/vingo-utils-v3/vingo"
)

type DownloadOption struct {
	Url        string
	SaveDir    string
	RandomName bool
	Timeout    time.Duration
	Percent    func(percent float64)
}

type progressWriter struct {
	writer   io.Writer
	total    int64
	written  int64
	callback func(percent float64)

	lastPercent int
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)

	w.written += int64(n)

	if w.total > 0 && w.callback != nil {
		percent := int(float64(w.written) / float64(w.total) * 100)

		if percent > 100 {
			percent = 100
		}

		// 每变化 1% 更新一次
		if percent != w.lastPercent {
			w.lastPercent = percent
			w.callback(float64(percent))
		}
	}

	return n, err
}

func DownloadFile(option DownloadOption) string {
	parsedUrl, err := url.Parse(option.Url)
	if err != nil {
		panic(err)
	}

	fileName := ""
	if option.RandomName {
		fileName = vingo.GetUUID() + path.Ext(parsedUrl.Path)
	} else {
		fileName = path.Base(parsedUrl.Path)
	}

	err = os.MkdirAll(option.SaveDir, 0777)
	if err != nil {
		panic(err)
	}

	savePath := filepath.Join(option.SaveDir, fileName)

	out, err := os.Create(savePath)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	client := &http.Client{
		Timeout: option.Timeout,
	}

	resp, err := client.Get(option.Url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("server returned non-200: %v", resp.StatusCode))
	}

	if option.Percent != nil {
		option.Percent(0)
	}

	pw := &progressWriter{
		writer:      out,
		total:       resp.ContentLength,
		callback:    option.Percent,
		lastPercent: -1,
	}

	_, err = io.Copy(pw, resp.Body)
	if err != nil {
		panic(err)
	}

	if option.Percent != nil {
		option.Percent(100)
	}

	return savePath
}
