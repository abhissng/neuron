package email

import (
	"fmt"

	"github.com/abhissng/neuron/adapters/log"
	gomail "gopkg.in/mail.v2"
)

type EmailClient interface {
	Send(data *EmailData) error
}

type GomailClient struct {
	opts ClientOptions
}

// NewGomailClient creates a new gomail client
// @param opts: The options for the gomail client
// @return: The gomail client
// @return: The error if any
// must provide logger
func NewGomailClient(opts ...Option) (EmailClient, error) {
	// default options
	o := &ClientOptions{
		Type: "SMTP",
		Host: "localhost",
		Port: 25,
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.log == nil {
		return nil, fmt.Errorf("logger  is required")
	}
	return &GomailClient{opts: *o}, nil
}

func (c *GomailClient) Send(data *EmailData) error {
	// Template replacements
	var subject, html, text string
	if data.TemplateData != nil {
		subject = applyTemplate(data.Subject, data.TemplateData)
		html = applyTemplate(data.HTMLBody, data.TemplateData)
		text = applyTemplate(data.TextBody, data.TemplateData)
	} else {
		subject = data.Subject
		html = data.HTMLBody
		text = data.TextBody
	}

	m := gomail.NewMessage()
	m.SetHeader("From", data.From)
	if len(data.To) > 0 {
		m.SetHeader("To", data.To...)
	}
	if len(data.CC) > 0 {
		m.SetHeader("Cc", data.CC...)
	}
	if len(data.BCC) > 0 {
		m.SetHeader("Bcc", data.BCC...)
	}
	m.SetHeader("Subject", subject)

	// Bodies
	if text != "" {
		m.SetBody("text/plain", text)
	}
	if html != "" {
		// add HTML as alternative
		// if text was set, alternative, else it's just body
		m.AddAlternative("text/html", html)
	}

	// Custom headers (if any)
	for k, v := range data.Headers {
		m.SetHeader(k, v)
	}

	// Attachments
	if err := attachFiles(m, data.Attachments); err != nil {
		return fmt.Errorf("attachment error: %w", err)
	}

	// Dialer
	d := gomail.NewDialer(c.opts.Host, c.opts.Port, c.opts.Username, c.opts.Password)
	if c.opts.TLSConfig != nil {
		d.TLSConfig = c.opts.TLSConfig
	}

	// Send
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email via gomail: %w", err)
	}
	c.opts.log.Info("email sent successfully", log.Any("to", data.To), log.Any("cc", data.CC), log.Any("bcc", data.BCC), log.Any("subject", subject))
	return nil
}
