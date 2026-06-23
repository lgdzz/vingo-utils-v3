// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/6/22
// 描述：SSL证书部署
// *****************************************************************************

package router

import (
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type SslInput struct {
	Domain                  string `json:"domain"`
	Domains                 string `json:"domains"`
	Certificate             string `json:"certificate"`
	PrivateKey              string `json:"private_key"`
	ServerCertificate       string `json:"server_certificate"`
	IntermediateCertificate string `json:"intermediate_certificate"`
	RestartCmd              string `json:"restart_cmd"`
}

func saveFile(name string, content string) {
	err := os.WriteFile(name, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
}

func ssl(r *gin.Engine) {
	r.GET("/ssl.input", func(c *gin.Context) {
		c.JSON(200, SslInput{})
	})

	r.POST("/ssl.deploy", func(c *gin.Context) {
		var input SslInput
		if err := c.ShouldBindJSON(&input); err != nil {
			panic(err.Error())
		}

		saveFile("ssl/fullchain.pem", input.Certificate)
		saveFile("ssl/privkey.pem", input.PrivateKey)

		if input.RestartCmd != "" {
			cmdText := strings.Split(strings.TrimSpace(input.RestartCmd), " ")
			cmd := exec.Command(cmdText[0], cmdText[1:]...)
			_, err := cmd.CombinedOutput()
			if err != nil {
				panic(err)
			}
		}

		c.JSON(200, gin.H{})
	})
}
