package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lgdzz/vingo-utils-exception/exception"
	"github.com/lgdzz/vingo-utils-v3/cryptor"
	"github.com/lgdzz/vingo-utils-v3/redis"
	"github.com/lgdzz/vingo-utils-v3/vingo"
)

type Api[T any] struct {
	Secret string
	Api    *redis.Api
}

type Ticket struct {
	Key string `json:"key"`
	TK  string `json:"tk"`
}

// Body 签发主体
type Body[T any] struct {
	// 必须
	Id       any `json:"id"`
	Business T   `json:"business"` // 业务数据

	// 单点登录设置为true
	CheckTK bool `json:"checkTk"`

	// 默认有效期24小时
	Hour int `json:"hour"`

	// 单点登录凭证（内部处理，无需传入）
	Ticket *Ticket `json:"ticket"`
}

type Response struct {
	Token  string `json:"token"`
	Expire int64  `json:"expire"`
}

// NewJwt 创建jwt
func NewJwt[T any](secret string, redisApi *redis.Api) *Api[T] {
	return &Api[T]{
		Secret: secret,
		Api:    redisApi,
	}
}

// Issued 签发token
func (s *Api[T]) Issued(body Body[T]) Response {
	if body.Hour == 0 {
		body.Hour = 1
	}
	now := time.Now()
	if body.Hour >= 24 {
		// 如果签发时间超过24小时，则将当前时间延迟到当天23:59:59（避免在白天时间突然失效）
		now = time.Date(
			now.Year(), now.Month(), now.Day(),
			23, 59, 59, 0,
			now.Location(),
		)
	}
	hour := 3600 * int64(body.Hour)
	exp := now.Unix() + hour
	if body.CheckTK {
		body.Ticket = &Ticket{Key: cryptor.Md5(fmt.Sprintf("%v%v", s.Secret, body.Id)), TK: vingo.RandomString(50)}
		s.Api.Set(body.Ticket.Key, body.Ticket.TK, time.Duration(exp-time.Now().Unix()))
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"id": body.Id, "checkTk": body.CheckTK, "ticket": body.Ticket, "business": body.Business, "exp": exp}).SignedString([]byte(s.Secret))
	if err != nil {
		panic(err)
	}
	return Response{
		Token:  token,
		Expire: exp,
	}
}

// Check 检查token是否有效
func (s *Api[T]) Check(token string) Body[T] {
	claims, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(s.Secret), nil
	})
	if err != nil {
		panic(&exception.AuthException{Message: err.Error()})
	}
	body := vingo.Convert[Body[T]](claims.Claims)
	if body.CheckTK {
		var tk string
		if !s.Api.Get(body.Ticket.Key, &tk) || tk != body.Ticket.TK {
			panic(&exception.AuthException{Message: "登录已失效"})
		}
	}
	return body
}
