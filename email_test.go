package email_test

import (
	htmltpl "html/template"
	"os"
	"strings"
	"testing"
	texttpl "text/template"

	"github.com/m12r/email"
	"github.com/m12r/email/smtp"
)

func TestSendEmailTemplate(t *testing.T) {
	sender := setupSender(t)
	from, to := setupFromAndTo(t)

	sysFS := os.DirFS("test_data")

	htmlTpl, err := htmltpl.New("").Parse(htmlTemplate)
	if err != nil {
		t.Fatalf("cannot parse html template: %v", err)
	}

	textTpl, err := texttpl.New("").Parse(textTemplate)
	if err != nil {
		t.Fatalf("cannot parse text template: %v", err)
	}

	data := struct {
		Subject string
		Title   string
		Text    string
	}{
		Subject: "This is a test subject",
		Title:   "This is a test title",
		Text:    "Test eins, zwo, drei, vier!",
	}

	m, err := email.NewMessage(
		from,
		data.Subject,
		email.To(to),
		email.SetHtmlFromTemplate(htmlTpl, "", data, email.InlineFromFile(sysFS, "test.jpg")),
		email.SetPlainFromTemplate(textTpl, "", data),
		email.AttachFromFile(sysFS, "test.pdf"),
	)
	if err != nil {
		t.Fatalf("cannot create message: %v", err)
	}

	if err := sender.Send(m); err != nil {
		t.Fatalf("cannot send message: %v", err)
	}
}

func setupSender(t *testing.T) email.Sender {
	t.Helper()

	smtpAddr, ok := os.LookupEnv("SMTP_ADDR")
	if !ok {
		t.Fatalf("SMTP_ADDR not set")
	}
	username, ok := os.LookupEnv("SMTP_USERNAME")
	if !ok {
		t.Fatalf("SMTP_USERNAME not set")
	}
	password, ok := os.LookupEnv("SMTP_PASSWORD")
	if !ok {
		t.Fatalf("SMTP_PASSWORD not set")
	}
	useStartTLS := true
	useStartTLSStr, ok := os.LookupEnv("SMTP_USE_STARTTLS")
	if ok {
		useStartTLS = strings.EqualFold("true", useStartTLSStr)
	}

	cfg := &smtp.Config{
		ServerAddr:  smtpAddr,
		Username:    username,
		Password:    password,
		UseStartTLS: useStartTLS,
	}

	sender, err := cfg.NewSender()
	if err != nil {
		t.Fatalf("cannot create sender: %v", err)
	}

	return sender
}

func setupFromAndTo(t *testing.T) (string, string) {
	t.Helper()

	from, ok := os.LookupEnv("SMTP_FROM")
	if !ok {
		t.Fatal("SMTP_FROM not set")
	}
	to, ok := os.LookupEnv("SMTP_TO")
	if !ok {
		t.Fatal("SMTP_TO not set")
	}
	return from, to
}

var htmlTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
	<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<title>{{ .Subject }}</title>
</head>
<body style="margin: 0; padding: 0;">
<table border="0" cellpadding="0" cellspacing="0" width="100%">
	<tr>
		<td style="font-size: 24px; font-family: Arial, sans-serif; color: #ff0000; background-color: #ffffee;">
			<b>{{ .Title }}</b>
		</td>
	</tr>
	<tr>
		<td>{{ .Text }}</td>
	</tr>
	<tr>
		<td>This is a <a href="https://m12r.at">link</a>.</td>
	<tr>
</table>
<img src="cid:test.jpg" width="64" height="64" />
</body>
</html>
`

var textTemplate = `{{ .Title }}

{{ .Text }}

Please visit us here: https://m12r.at
`
