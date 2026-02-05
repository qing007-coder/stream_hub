package email

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"stream_hub/pkg/errors"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/utils"
)

type Client struct {
	srv      *gomail.Message
	username string
	password string
	host     string
	port     int
}

func NewClient(conf *config.CommonConfig) *Client {
	srv := gomail.NewMessage()

	return &Client{
		srv:      srv,
		username: conf.Email.Username,
		password: conf.Email.Password,
		host:     conf.Email.Host,
		port:     conf.Email.Port,
	}
}

func (s *Client) SendVerificationCode(target string) (string, error) {
	code := utils.RandomNumber(6)

	s.srv.SetHeader("From", fmt.Sprintf("xubo <%s>", s.username))
	s.srv.SetHeader("To", target)
	s.srv.SetHeader("Subject", "邮箱验证(测试)")
	s.srv.SetBody("text/plain", fmt.Sprintf("【注册验证】验证码：%s,有效期15分钟", code))
	d := gomail.NewDialer(s.host, s.port, s.username, s.password)

	if err := d.DialAndSend(s.srv); err != nil {
		return "", errors.EmailSendingFailed
	}

	return code, nil
}
