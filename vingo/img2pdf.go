// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/3/11
// 描述：图片转PDF文件
// *****************************************************************************

package vingo

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// Img2Pdf 图片base64转PDF（A4纸，一张图一页）
func Img2Pdf(imgBase64 []string) string {
	if len(imgBase64) == 0 {
		panic("未检测到图片数据")
	}

	const pageW = 210.0
	const pageH = 297.0

	pdf := gofpdf.New("P", "mm", "A4", "")

	for i, b64 := range imgBase64 {

		b64 = strings.TrimSpace(b64)
		if b64 == "" {
			continue
		}

		// 去data前缀
		if idx := strings.Index(b64, ","); idx != -1 && strings.HasPrefix(b64, "data:") {
			b64 = b64[idx+1:]
		}

		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			panic(err)
		}

		img, format, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			panic(err)
		}

		bounds := img.Bounds()
		imgW := bounds.Dx()
		imgH := bounds.Dy()

		if imgW == 0 || imgH == 0 {
			continue
		}

		scale := pageW / float64(imgW)
		if s := pageH / float64(imgH); s < scale {
			scale = s
		}

		w := float64(imgW) * scale
		h := float64(imgH) * scale

		x := (pageW - w) / 2
		y := (pageH - h) / 2

		pdf.AddPage()

		imgName := fmt.Sprintf("img_%d", i)

		var reader *bytes.Reader
		var imgType string

		switch strings.ToLower(format) {

		case "png":
			imgType = "PNG"
			reader = bytes.NewReader(data)

		case "jpeg", "jpg":
			imgType = "JPG"
			reader = bytes.NewReader(data)

		case "gif":
			imgType = "GIF"
			reader = bytes.NewReader(data)

		default:

			// 转JPEG
			var buf bytes.Buffer
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
			if err != nil {
				panic(err)
			}

			imgType = "JPG"
			reader = bytes.NewReader(buf.Bytes())
		}

		opt := gofpdf.ImageOptions{
			ImageType: imgType,
			ReadDpi:   false,
		}

		pdf.RegisterImageOptionsReader(imgName, opt, reader)

		if pdf.Error() != nil {
			panic(pdf.Error())
		}

		pdf.ImageOptions(imgName, x, y, w, h, false, opt, 0, "")
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
