// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/1/4
// 描述：
// *****************************************************************************

package email

const (
	EMAIL_BIND      = "email.bind"
	PASSWORD_FORGET = "pwd.forget"
)

type SenderConfig struct {
	Host     string `json:"host"`     // SMTP 地址
	Port     int    `json:"port"`     // SMTP 端口
	Username string `json:"username"` // 发件人账号
	Password string `json:"password"` // 发件人密码
	FromAddr string `json:"fromAddr"` // 发件人邮箱
	FromName string `json:"fromName"` // 发件人名称
	OEM      string `json:"oem"`      // 品牌商
}
