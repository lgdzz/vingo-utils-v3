// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/3/30
// 描述：
// *****************************************************************************

package captcha

import (
	"fmt"
	"log"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/lgdzz/vingo-utils-v3/redis"
	"github.com/lgdzz/vingo-utils-v3/vingo"
	"github.com/wenlng/go-captcha-assets/bindata/chars"
	"github.com/wenlng/go-captcha-assets/resources/fonts/fzshengsksjw"
	"github.com/wenlng/go-captcha-assets/resources/imagesv2"
	"github.com/wenlng/go-captcha/v2/click"
)

type ClickCaptcha struct {
	Click click.Captcha
	Redis *redis.Api
}

func (s *ClickCaptcha) InitClick() {
	builder := click.NewBuilder()

	// fonts
	fonts, err := fzshengsksjw.GetFont()
	if err != nil {
		log.Fatalln(err)
	}

	// background images
	imgs, err := imagesv2.GetImages()
	if err != nil {
		log.Fatalln(err)
	}

	builder.SetResources(
		click.WithChars(chars.GetChineseChars()),
		click.WithFonts([]*truetype.Font{fonts}),
		click.WithBackgrounds(imgs),
	)

	s.Click = builder.Make()
}

type ClickInput struct {
	Key  string      `json:"key"`
	Code []click.Dot `json:"code"`
}

func (s *ClickCaptcha) Generate(c *vingo.Context) {
	captData, err := s.Click.Generate()
	if err != nil {
		panic(err)
	}

	dotData := captData.GetData()
	if dotData == nil {
		panic(">>>>> generate err")
	}

	var key = fmt.Sprintf("captcha:%s", vingo.GetUUID())
	s.Redis.Set(key, dotData, time.Minute*5)

	var mBase64, tBase64 string
	mBase64, err = captData.GetMasterImage().ToBase64()
	if err != nil {
		panic(err)
	}
	tBase64, err = captData.GetThumbImage().ToBase64()
	if err != nil {
		panic(err)
	}

	c.ResponseSuccess(map[string]any{
		"key":     key,
		"mBase64": mBase64,
		"tBase64": tBase64,
	})
}

func (s *ClickCaptcha) Verify(input ClickInput) bool {
	var dots map[int]*click.Dot
	if !s.Redis.Get(input.Key, &dots) {
		panic("验证码已失效")
	}
	for _, dot := range input.Code {
		if d, ok := dots[dot.Index-1]; ok {
			if !click.Validate(dot.X, dot.Y, d.X, d.Y, d.Width, d.Height, 3) {
				panic("验证码错误")
			}
		} else {
			panic("验证码错误")
		}
	}
	return true
}
