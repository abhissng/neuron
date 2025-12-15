package email

import "strings"

// EmailData represents the data for an email message
// This is the data that is passed to the email client to send an email
// @field From: The sender's email address <No Reply <email>>
// @field To: The recipient's email address <email>
// @field CC: The carbon copy recipients' email addresses <email>
// @field BCC: The blind carbon copy recipients' email addresses <email>
// @field Subject: The subject of the email <string>
// @field TextBody: The body of the email in plain text <string>
// @field HTMLBody: The body of the email in HTML <string>
// @field Attachments: The attachments to be sent with the email <string>
// @field TemplateData: The data to be used for templating <map[string]string>
// @field Headers: The custom headers to be added to the email <map[string]string>
type EmailData struct {
	From         string
	To           []string
	CC           []string
	BCC          []string
	Subject      string
	TextBody     string
	HTMLBody     string
	Attachments  []string // supports local paths or URLs
	TemplateData map[string]string
	// Custom headers if needed
	Headers map[string]string
}

func NewEmailData() *EmailData {
	return &EmailData{
		To:           make([]string, 0),
		CC:           make([]string, 0),
		BCC:          make([]string, 0),
		Attachments:  make([]string, 0),
		TemplateData: make(map[string]string),
		Headers:      make(map[string]string),
	}
}

// Apply simple templating: replace {{.key}} with value
func applyTemplate(input string, data map[string]string) string {
	for k, v := range data {
		placeholder := "{{." + k + "}}"
		input = strings.ReplaceAll(input, placeholder, v)
	}
	return input
}
