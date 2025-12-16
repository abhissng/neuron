package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/abhissng/neuron/adapters/aws"
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

// SESClient is an email client that uses AWS SES via AWSManager
type SESClient struct {
	awsManager *aws.AWSManager
	log        *log.Log
}

// NewSESClient creates a new SES email client using an existing AWSManager
func NewSESClient(awsManager *aws.AWSManager, opts ...Option) (EmailClient, error) {
	if awsManager == nil {
		return nil, fmt.Errorf("AWS manager is required")
	}

	o := &ClientOptions{
		Type: "SES",
	}
	for _, opt := range opts {
		opt(o)
	}

	if o.log == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &SESClient{
		awsManager: awsManager,
		log:        o.log,
	}, nil
}

// Send sends an email using AWS SES
func (c *SESClient) Send(data *EmailData) error {
	if data == nil {
		return fmt.Errorf("email data is required")
	}
	fromHeader := formatFromHeader(data.From, data.FromName)

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

	// If there are attachments, use raw email
	if len(data.Attachments) > 0 {
		return c.sendRawEmail(data, subject, html, text)
	}

	// Use simple email for non-attachment emails
	input := &aws.SESEmailInput{
		From:     fromHeader,
		To:       data.To,
		CC:       data.CC,
		BCC:      data.BCC,
		Subject:  subject,
		TextBody: text,
		HTMLBody: html,
	}

	_, err := c.awsManager.SendEmail(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to send email via SES: %w", err)
	}

	c.log.Info("email sent successfully via SES", log.Any("to", data.To), log.Any("cc", data.CC), log.Any("bcc", data.BCC), log.Any("subject", subject))
	return nil
}

// sendRawEmail sends an email with attachments using raw MIME format
func (c *SESClient) sendRawEmail(data *EmailData, subject, html, text string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	fromHeader := formatFromHeader(data.From, data.FromName)

	// Build headers
	headers := make(textproto.MIMEHeader)
	headers.Set("From", fromHeader)
	headers.Set("To", strings.Join(data.To, ", "))
	if len(data.CC) > 0 {
		headers.Set("Cc", strings.Join(data.CC, ", "))
	}
	headers.Set("Subject", subject)
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", writer.Boundary()))

	// Write headers to buffer
	for key, values := range headers {
		for _, value := range values {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}
	buf.WriteString("\r\n")

	// Add text body part
	if text != "" {
		textHeader := make(textproto.MIMEHeader)
		textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
		textHeader.Set("Content-Transfer-Encoding", "quoted-printable")
		part, err := writer.CreatePart(textHeader)
		if err != nil {
			return fmt.Errorf("failed to create text part: %w", err)
		}
		_, _ = part.Write([]byte(text))
	}

	// Add HTML body part
	if html != "" {
		htmlHeader := make(textproto.MIMEHeader)
		htmlHeader.Set("Content-Type", "text/html; charset=UTF-8")
		htmlHeader.Set("Content-Transfer-Encoding", "quoted-printable")
		part, err := writer.CreatePart(htmlHeader)
		if err != nil {
			return fmt.Errorf("failed to create HTML part: %w", err)
		}
		_, _ = part.Write([]byte(html))
	}

	// Add attachments
	for _, att := range data.Attachments {
		attData, filename, contentType, err := fetchAttachment(att)
		if err != nil {
			return fmt.Errorf("failed to fetch attachment %s: %w", att, err)
		}

		attHeader := make(textproto.MIMEHeader)
		attHeader.Set("Content-Type", fmt.Sprintf("%s; name=\"%s\"", contentType, filename))
		attHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		attHeader.Set("Content-Transfer-Encoding", "base64")

		part, err := writer.CreatePart(attHeader)
		if err != nil {
			return fmt.Errorf("failed to create attachment part: %w", err)
		}

		encoded := base64.StdEncoding.EncodeToString(attData)
		_, _ = part.Write([]byte(encoded))
	}

	_ = writer.Close()

	// Collect all recipients
	allRecipients := append([]string{}, data.To...)
	allRecipients = append(allRecipients, data.CC...)
	allRecipients = append(allRecipients, data.BCC...)

	_, err := c.awsManager.SendRawEmail(context.Background(), buf.Bytes(), fromHeader, allRecipients)
	if err != nil {
		return fmt.Errorf("failed to send raw email via SES: %w", err)
	}

	c.log.Info("email with attachments sent successfully via SES", log.Any("to", data.To), log.Any("cc", data.CC), log.Any("bcc", data.BCC), log.Any("subject", subject))
	return nil
}

func (c *GomailClient) Send(data *EmailData) error {
	if data == nil {
		return fmt.Errorf("email data is required")
	}
	fromHeader := formatFromHeader(data.From, data.FromName)
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
	m.SetHeader("From", fromHeader)
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
