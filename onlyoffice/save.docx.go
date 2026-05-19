// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/5/15
// 描述：OnlyOffice保存
// *****************************************************************************

package onlyoffice

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lgdzz/vingo-utils-v3/cryptor"
	"github.com/lgdzz/vingo-utils-v3/vingo"
)

type Callback struct {
	ChangesURL    string   `json:"changesurl"`
	ForceSaveType int      `json:"forcesavetype"`
	History       History  `json:"history"`
	Key           string   `json:"key"`
	LastSave      string   `json:"lastsave"`
	Status        int      `json:"status"`
	URL           string   `json:"url"`
	Users         []string `json:"users"`
}

type History struct {
	Changes       []Change `json:"changes"`
	ServerVersion string   `json:"serverVersion"`
}

type Change struct {
	Created string `json:"created"`
	User    User   `json:"user"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateDocx 创建一个新的文件
func (s *Api) CreateDocx(objectName string) {
	docxBase64, err := CreateEmptyDocxBase64()
	if err != nil {
		panic(fmt.Sprintf("CreateEmptyDocxBase64 failed: %v", err))
	}
	s.OSS.UploadBase64(objectName, vingo.CT_DOCX, docxBase64)
}

// UpdateDocx 对已有文件进行修改保存
func (s *Api) UpdateDocx(c *vingo.Context) {
	defer func() {
		if err := recover(); err != nil {
			c.JSON(200, gin.H{"error": 1})
		}
	}()
	input := vingo.GetRequestBody[Callback](c)
	if input.Status == 6 {
		docxBase64, contentType := s.OSS.GetBase64(input.URL)
		s.OSS.UploadBase64(s.OSS.ObjectName(cryptor.TextBase64Decode(input.Key)), contentType, docxBase64)
	}
	c.JSON(200, gin.H{"error": 0})
}
