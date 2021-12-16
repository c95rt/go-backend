package helpers

import (
	"bytes"
	"html/template"
	"io"

	"gopkg.in/gomail.v2"
)

type EmailData struct {
	EmailTo      string
	NameTo       string
	EmailFrom    string
	NameFrom     string
	Subject      string
	TemplatePath string
	FileName     string
	FileContent  []byte
	AwsSMTP      *gomail.Dialer
}

func (ed *EmailData) SendEmail(data interface{}) error {
	htmlSubmit := ed.TemplatePath
	var err error
	t, err := template.ParseFiles(htmlSubmit)
	if err != nil {
		return err
	}
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return err
	}

	result := tpl.String()
	m := gomail.NewMessage()

	if ed.FileContent != nil {
		m.Attach(ed.FileName, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(ed.FileContent)
			if err != nil {
				return err
			}
			return nil
		}))
	}

	m.SetHeader("From", m.FormatAddress(ed.EmailFrom, ed.NameFrom))
	m.SetHeader("To", m.FormatAddress(ed.EmailTo, ed.NameTo))
	m.SetHeader("Subject", ed.Subject)
	m.SetBody("text/html", result)
	if err := ed.AwsSMTP.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
