package mailclient

import (
	mailv1 "github.com/swayrider/protos/mail/v1"
)

type Mail struct {
	From     string
	To       []string
	CC       []string
	BCC      []string
	Subject  string
	HtmlBody string
	TextBody string
}

func NewMail(
	from string,
	to []string,
	cc []string,
	bcc []string,
	subject string,
	htmlBody string,
	textBody string,
) *Mail {
	return &Mail{
		From:     from,
		To:       to,
		CC:       cc,
		BCC:      bcc,
		Subject:  subject,
		HtmlBody: htmlBody,
		TextBody: textBody,
	}
}

func (m Mail) Request() *mailv1.SendRequest {
	return &mailv1.SendRequest{
		From:     m.From,
		To:       m.To,
		Cc:       m.CC,
		Bcc:      m.BCC,
		Subject:  m.Subject,
		HtmlBody: m.HtmlBody,
		TextBody: m.TextBody,
	}
}

type TemplateMail struct {
	From         string
	To           []string
	CC           []string
	BCC          []string
	Subject      string
	HtmlTemplate string
	TextTemplate string
	TemplateData map[string]string
}

func NewTemplateMail(
	from string,
	to []string,
	cc []string,
	bcc []string,
	subject string,
	htmlTemplate string,
	textTemplate string,
	templateData map[string]string,
) *TemplateMail {
	return &TemplateMail{
		From:         from,
		To:           to,
		CC:           cc,
		BCC:          bcc,
		Subject:      subject,
		HtmlTemplate: htmlTemplate,
		TextTemplate: textTemplate,
		TemplateData: templateData,
	}
}

func (m TemplateMail) Request() *mailv1.SendTemplateRequest {
	return &mailv1.SendTemplateRequest{
		From:         m.From,
		To:           m.To,
		Cc:           m.CC,
		Bcc:          m.BCC,
		Subject:      m.Subject,
		HtmlTemplate: m.HtmlTemplate,
		TextTemplate: m.TextTemplate,
		Data:         m.TemplateData,
	}
}
