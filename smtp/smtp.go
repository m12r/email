package smtp

import (
	"crypto/tls"
	"net/smtp"
	"net/textproto"

	"github.com/m12r/email"

	jwe "github.com/jordan-wright/email"
)

type SenderOpt func(e *Sender)

func WithSender(sender string) SenderOpt {
	return func(e *Sender) {
		e.sender = sender
	}
}

func WithTLSConfig(cfg *tls.Config) SenderOpt {
	return func(e *Sender) {
		e.tlsConfig = cfg
	}
}

func WithStartTLS(use bool) SenderOpt {
	return func(e *Sender) {
		e.useStartTLS = use
	}
}

type Sender struct {
	addr        string
	auth        smtp.Auth
	sender      string
	tlsConfig   *tls.Config
	useStartTLS bool
}

func NewSender(addr string, auth smtp.Auth, opts ...SenderOpt) email.Sender {
	e := &Sender{
		addr: addr,
		auth: auth,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Sender) Send(msg *email.Message) error {
	m := buildMessage(msg)
	if e.sender != "" {
		m.Sender = e.sender
	}

	if e.tlsConfig == nil {
		if err := m.Send(e.addr, e.auth); err != nil {
			return err
		}
		return nil
	}
	sendFunc := m.SendWithTLS
	if e.useStartTLS {
		sendFunc = m.SendWithStartTLS
	}
	if err := sendFunc(e.addr, e.auth, e.tlsConfig); err != nil {
		return err
	}
	return nil
}

func buildMessage(msg *email.Message) *jwe.Email {
	m := &jwe.Email{
		From:    msg.From,
		Subject: msg.Subject,
		Headers: make(textproto.MIMEHeader),
	}
	for _, to := range msg.To {
		m.To = append(m.To, to)
	}
	for _, cc := range msg.Cc {
		m.Cc = append(m.Cc, cc)
	}
	for _, bcc := range msg.Bcc {
		m.Bcc = append(m.Bcc, bcc)
	}
	if msg.Plain != nil {
		m.Text = make([]byte, len(msg.Plain))
		copy(m.Text, msg.Plain)
	}
	if msg.Html != nil {
		m.HTML = make([]byte, len(msg.Html.Body))
		copy(m.HTML, msg.Html.Body)
		for _, mi := range msg.Html.Inlines {
			a := &jwe.Attachment{
				Filename:    mi.Name,
				ContentType: mi.ContentType,
				Header:      make(textproto.MIMEHeader),
			}
			a.Content = make([]byte, len(mi.Body))
			copy(a.Content, mi.Body)
			m.Attachments = append(m.Attachments, a)
		}
	}
	for _, ma := range msg.Attachments {
		a := &jwe.Attachment{
			Filename:    ma.Name,
			ContentType: ma.ContentType,
			Header:      make(textproto.MIMEHeader),
		}
		a.Content = make([]byte, len(ma.Body))
		copy(a.Content, ma.Body)
		m.Attachments = append(m.Attachments, a)
	}
	return m
}

func WithPlainAuth(identity, username, password, host string) smtp.Auth {
	return smtp.PlainAuth(identity, username, password, host)
}

func WithCRAMMD5Auth(username, secret string) smtp.Auth {
	return smtp.CRAMMD5Auth(username, secret)
}
