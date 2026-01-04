// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/1/4
// 描述：
// *****************************************************************************

package email

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net/smtp"
	"strings"
)

// SendMail SendMailSMTPS 使用 SMTPS(SSL/TLS) 发送邮件
func SendMail(
	host string, // SMTP 服务器，如: mail.example.com
	port int, // SMTPS 端口，通常 465
	username string, // 邮箱账号
	password string, // 邮箱密码 / 授权码
	from string, // 发件人
	fromName string,
	to []string, // 收件人列表
	subject string, // 主题
	body string, // 正文（纯文本）
) error {

	addr := fmt.Sprintf("%s:%d", host, port)

	// 1. TLS 配置
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false, // ✅ 生产环境不要设 true
	}

	// 2. 建立 TLS 连接
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial error: %w", err)
	}
	defer conn.Close()

	// 3. 创建 SMTP client
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client error: %w", err)
	}
	defer client.Quit()

	// 4. 认证
	auth := smtp.PlainAuth("", username, password, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth error: %w", err)
	}

	// 5. 设置发件人
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("set from error: %w", err)
	}

	// 6. 设置收件人
	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("set rcpt error: %w", err)
		}
	}

	// 7. 写入邮件内容
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("data error: %w", err)
	}

	message := buildMessage(from, fromName, to, subject, body)

	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}

func encodeRFC2047(str string) string {
	return mime.QEncoding.Encode("UTF-8", str)
}

// 构建邮件内容（纯文本）
func buildMessage(
	fromAddr string, // noreply@example.com
	fromName string, // 系统通知
	to []string,
	subject, body string,
) string {

	from := fmt.Sprintf("%s <%s>", encodeRFC2047(fromName), fromAddr)

	headers := map[string]string{
		"From":         from,
		"To":           strings.Join(to, ", "),
		"Subject":      encodeRFC2047(subject),
		"MIME-Version": "1.0",
		"Content-Type": `text/html; charset="UTF-8"`,
	}

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	return msg.String()
}
