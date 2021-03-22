package email

import (
	"bytes"
	htmltpl "html/template"
	"io"
	"io/fs"
	"mime"
	"path"
	texttpl "text/template"

	"github.com/m12r/email/internal/bbpool"
)

type MessageOpt func(*Message) error

func Attach(name, contentType string, r io.Reader) MessageOpt {
	return func(m *Message) error {
		return m.Attach(name, contentType, r)
	}
}

func AttachFromFile(fsys fs.FS, name string) MessageOpt {
	return func(m *Message) error {
		return m.AttachFromFile(fsys, name)
	}
}

func To(addrs ...string) MessageOpt {
	return func(m *Message) error {
		m.AddTo(addrs...)
		return nil
	}
}

func Cc(addrs ...string) MessageOpt {
	return func(m *Message) error {
		m.AddCc(addrs...)
		return nil
	}
}

func Bcc(addrs ...string) MessageOpt {
	return func(m *Message) error {
		m.AddBcc(addrs...)
		return nil
	}
}

func SetPlain(r io.Reader) MessageOpt {
	return func(m *Message) error {
		return m.SetPlain(r)
	}
}

func SetPlainFromString(msg string) MessageOpt {
	return func(m *Message) error {
		return m.SetPlainFromString(msg)
	}
}

func SetPlainFromTemplate(tpl *texttpl.Template, tplName string, data interface{}) MessageOpt {
	return func(m *Message) error {
		return m.SetPlainFromTemplate(tpl, tplName, data)
	}
}

func SetHtml(r io.Reader, opts ...HtmlOpt) MessageOpt {
	return func(m *Message) error {
		return m.SetHtml(r, opts...)
	}
}

func SetHtmlFromString(msg string, opts ...HtmlOpt) MessageOpt {
	return func(m *Message) error {
		return m.SetHtmlFromString(msg, opts...)
	}
}

func SetHtmlFromTemplate(tpl *htmltpl.Template, tplName string, data interface{}, opts ...HtmlOpt) MessageOpt {
	return func(m *Message) error {
		return m.SetHtmlFromTemplate(tpl, tplName, data, opts...)
	}
}

type Message struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Plain       []byte
	Html        *Html
	Attachments []*Attachment
}

func NewMessage(from string, subject string, opts ...MessageOpt) (*Message, error) {
	m := &Message{
		From:    from,
		Subject: subject,
	}
	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *Message) AddTo(addrs ...string) {
	for _, addr := range addrs {
		m.To = append(m.To, addr)
	}
}

func (m *Message) AddCc(addrs ...string) {
	for _, addr := range addrs {
		m.Cc = append(m.Cc, addr)
	}
}

func (m *Message) AddBcc(addrs ...string) {
	for _, addr := range addrs {
		m.Bcc = append(m.Bcc, addr)
	}
}

func (m *Message) SetPlain(r io.Reader) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		buf = bbpool.Get()
		defer bbpool.Put(buf)

		if _, err := buf.ReadFrom(r); err != nil {
			return err
		}
	}

	m.Plain = make([]byte, buf.Len())
	copy(m.Plain, buf.Bytes())
	return nil
}

func (m *Message) SetPlainFromString(msg string) error {
	m.Plain = []byte(msg)
	return nil
}

func (m *Message) SetPlainFromTemplate(tpl *texttpl.Template, tplName string, data interface{}) error {
	buf := bbpool.Get()
	defer bbpool.Put(buf)

	if tplName != "" {
		if err := tpl.ExecuteTemplate(buf, tplName, data); err != nil {
			return err
		}
	} else {
		if err := tpl.Execute(buf, data); err != nil {
			return err
		}
	}

	return m.SetPlain(buf)
}

func (m *Message) SetHtml(r io.Reader, opts ...HtmlOpt) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		buf = bbpool.Get()
		defer bbpool.Put(buf)

		if _, err := buf.ReadFrom(r); err != nil {
			return err
		}
	}

	h := &Html{}
	h.Body = make([]byte, buf.Len())
	copy(h.Body, buf.Bytes())

	for _, opt := range opts {
		if err := opt(h); err != nil {
			return err
		}
	}
	m.Html = h
	return nil
}

func (m *Message) SetHtmlFromString(msg string, opts ...HtmlOpt) error {
	h := &Html{
		Body: []byte(msg),
	}

	for _, opt := range opts {
		if err := opt(h); err != nil {
			return err
		}
	}
	m.Html = h
	return nil
}

func (m *Message) SetHtmlFromTemplate(tpl *htmltpl.Template, tplName string, data interface{}, opts ...HtmlOpt) error {
	buf := bbpool.Get()
	defer bbpool.Put(buf)

	if tplName != "" {
		if err := tpl.ExecuteTemplate(buf, tplName, data); err != nil {
			return err
		}
	} else {
		if err := tpl.Execute(buf, data); err != nil {
			return err
		}
	}

	return m.SetHtml(buf, opts...)
}

func (m *Message) Attach(name, contentType string, r io.Reader) error {
	a, err := newAttachment(name, contentType, r)
	if err != nil {
		return err
	}
	m.Attachments = append(m.Attachments, a)
	return nil
}

func (m *Message) AttachFromFile(fsys fs.FS, name string) error {
	f, err := fsys.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	base := path.Base(name)
	contentType := contentTypeFromFileName(name)

	return m.Attach(base, contentType, f)
}

type HtmlOpt func(*Html) error

func Inline(name, contentType string, r io.Reader) HtmlOpt {
	return func(h *Html) error {
		return h.Inline(name, contentType, r)
	}
}

func InlineFromFile(fsys fs.FS, name string) HtmlOpt {
	return func(h *Html) error {
		return h.InlineFromFile(fsys, name)
	}
}

type Html struct {
	Body    []byte
	Inlines []*Attachment
}

func (h *Html) Inline(name, contentType string, r io.Reader) error {
	a, err := newAttachment(name, contentType, r)
	if err != nil {
		return err
	}
	h.Inlines = append(h.Inlines, a)

	return nil
}

func (h *Html) InlineFromFile(fsys fs.FS, name string) error {
	f, err := fsys.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	base := path.Base(name)
	contentType := contentTypeFromFileName(name)

	return h.Inline(base, contentType, f)
}

type Attachment struct {
	Name        string
	ContentType string
	Body        []byte
}

type Sender interface {
	Send(msg *Message) error
}

func contentTypeFromFileName(name string) string {
	ct := "application/octet-stream"
	ext := path.Ext(name)
	if ext == "" {
		return ct
	}
	tmp := mime.TypeByExtension(ext)
	if tmp == "" {
		return ct
	}
	return tmp
}

func newAttachment(name, contentType string, r io.Reader) (*Attachment, error) {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		buf = bbpool.Get()
		defer bbpool.Put(buf)

		if _, err := buf.ReadFrom(r); err != nil {
			return nil, err
		}
	}

	a := &Attachment{
		Name:        name,
		ContentType: contentType,
	}

	a.Body = make([]byte, buf.Len())
	copy(a.Body, buf.Bytes())
	return a, nil
}
