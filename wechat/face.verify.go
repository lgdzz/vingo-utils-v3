// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/6/22
// 描述：人脸核身2.0
// *****************************************************************************

package wechat

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/lgdzz/vingo-utils-v3/request"
	"github.com/lgdzz/vingo-utils-v3/vingo"
)

type FaceVerify struct{}

type FaceGetVerifyIdResponse struct {
	Errcode  int    `json:"errcode"`
	Errmsg   string `json:"errmsg"`
	VerifyId string `json:"verify_id"`
}

type CertInfo struct {
	Name   string `json:"name"`
	IdCard string `json:"idCard"`
}

type FaceQueryObject struct {
	QueryData map[string]any `json:"queryData"`
	CertInfo  CertInfo       `json:"certInfo"`
}

// FaceGetVerifyId 获取用户人脸核身会话唯一标识
func (s *MiniProgram) FaceGetVerifyId(name string, idCard string, openid string) string {
	accessToken, err := s.GetAuth().GetAccessToken()
	if err != nil {
		panic(err)
	}
	seqNo := vingo.OrderNo(32, nil)
	response, _ := request.PostJSON(fmt.Sprintf("https://api.weixin.qq.com/cityservice/face/identify/getverifyid?access_token=%v", accessToken), map[string]any{
		"out_seq_no": seqNo,
		"cert_info": map[string]any{
			"cert_type": "IDENTITY_CARD",
			"cert_name": name,
			"cert_no":   idCard,
		},
		"openid": openid,
	}, request.Option{})

	var result FaceGetVerifyIdResponse
	vingo.StringToJson(string(response), &result)
	if result.Errcode != 0 {
		panic(result.Errmsg)
	}

	certTypeBase64 := base64.StdEncoding.EncodeToString([]byte("IDENTITY_CARD"))
	certNameBase64 := base64.StdEncoding.EncodeToString([]byte(name))
	certNoBase64 := base64.StdEncoding.EncodeToString([]byte(idCard))
	raw := fmt.Sprintf(
		"cert_type=%s&cert_name=%s&cert_no=%s",
		certTypeBase64,
		certNameBase64,
		certNoBase64,
	)
	hash := sha256.Sum256([]byte(raw))
	certHash := hex.EncodeToString(hash[:])

	s.RedisApi.Set(result.VerifyId, FaceQueryObject{
		QueryData: map[string]any{
			"verify_id":  result.VerifyId,
			"out_seq_no": seqNo,
			"cert_hash":  certHash,
			"openid":     openid,
		},
		CertInfo: CertInfo{
			Name:   name,
			IdCard: idCard,
		},
	}, time.Hour)

	return result.VerifyId
}

type FaceGetVerifyRetResponse struct {
	Errcode   int    `json:"errcode"`
	Errmsg    string `json:"errmsg"`
	VerifyRet int    `json:"verify_ret"`
}

// FaceQueryVerifyInfo 查询用户人脸核身真实验证结果
func (s *MiniProgram) FaceQueryVerifyInfo(verifyId string, success func(info CertInfo)) {

	var query FaceQueryObject
	if !s.RedisApi.Get(verifyId, &query) {
		panic("人脸核身信息失效")
	}

	accessToken, err := s.GetAuth().GetAccessToken()
	if err != nil {
		panic(err)
	}
	response, _ := request.PostJSON(fmt.Sprintf("https://api.weixin.qq.com/cityservice/face/identify/queryverifyinfo?access_token=%v", accessToken), query.QueryData, request.Option{})

	var result FaceGetVerifyRetResponse
	vingo.StringToJson(string(response), &result)
	if result.Errcode != 0 {
		panic(result.Errmsg)
	}

	if result.VerifyRet != 10000 {
		panic(fmt.Sprintf("人脸核身失败，错误码：%v", result.VerifyRet))
	}

	success(query.CertInfo)

	s.RedisApi.Del(verifyId)
}
