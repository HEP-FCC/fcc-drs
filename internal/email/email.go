package email

import (
	"net/smtp"
	"os"
	"strings"
)

type Config struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

func ConfigFromEnv() Config {
	return Config{
		Host: os.Getenv("SMTP_HOST"),
		Port: os.Getenv("SMTP_PORT"),
		User: os.Getenv("SMTP_USER"),
		Pass: os.Getenv("SMTP_PASS"),
		From: os.Getenv("SMTP_FROM"),
	}
}

func (c Config) Enabled() bool {
	return c.Host != "" && c.From != ""
}

func (c Config) Send(to, subject, body string) error {
	if !c.Enabled() {
		return nil
	}
	port := c.Port
	if port == "" {
		port = "587"
	}
	addr := c.Host + ":" + port

	msg := strings.Join([]string{
		"From: FCC Dataset Request System <" + sanitizeHeader(c.From) + ">",
		"To: " + sanitizeHeader(to),
		"Subject: " + sanitizeHeader(subject),
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if c.User != "" {
		auth = smtp.PlainAuth("", c.User, c.Pass, c.Host)
	}

	return smtp.SendMail(addr, auth, c.From, []string{to}, []byte(msg))
}

// sanitizeHeader strips CR/LF from a value bound for a raw email header,
// preventing header injection (e.g. a request title smuggling a Bcc: line).
func sanitizeHeader(s string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}
