package email

import (
	"fmt"
	"net/smtp"
)

type Client struct {
	host     string
	port     string
	user     string
	password string
	from     string
}

func NewClient(host, port, user, password, from string) *Client {
	return &Client{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
	}
}

func (c *Client) SendPasswordReset(to, resetURL string) error {
	subject := "Reset your Rosslib password"
	body := fmt.Sprintf(`You requested a password reset for your Rosslib account.

Click the link below to reset your password. This link expires in 1 hour.

%s

If you didn't request this, you can safely ignore this email.`, resetURL)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		c.from, to, subject, body)

	addr := c.host + ":" + c.port
	auth := smtp.PlainAuth("", c.user, c.password, c.host)

	return smtp.SendMail(addr, auth, c.from, []string{to}, []byte(msg))
}

func (c *Client) SendVerification(to, verifyURL string) error {
	subject := "Verify your Rosslib email"
	body := fmt.Sprintf(`Welcome to Rosslib!

Click the link below to verify your email address. This link expires in 24 hours.

%s

If you didn't create this account, you can safely ignore this email.`, verifyURL)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		c.from, to, subject, body)

	addr := c.host + ":" + c.port
	auth := smtp.PlainAuth("", c.user, c.password, c.host)

	return smtp.SendMail(addr, auth, c.from, []string{to}, []byte(msg))
}
