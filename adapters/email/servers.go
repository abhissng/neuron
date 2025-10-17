package email

import (
	"fmt"
	"strings"
)

// MailServerConfig holds the complete configuration for both
// SMTP (sending) and IMAP (receiving) services for an email provider.
type MailServerConfig struct {
	SMTPHost string
	SMTPPort int
	IMAPHost string
	IMAPPort int
}

// smtpServers is renamed to mailServerConfigs to reflect the consolidation.
var mailServerConfigs = map[string]MailServerConfig{
	"gmail": {
		SMTPHost: "smtp.gmail.com",
		SMTPPort: 587, // or 465
		IMAPHost: "imap.gmail.com",
		IMAPPort: 993,
	},
	"outlook": {
		SMTPHost: "smtp.office365.com",
		SMTPPort: 587,
		IMAPHost: "outlook.office365.com", // or imap-mail.outlook.com
		IMAPPort: 993,
	},
	"yahoo": {
		SMTPHost: "smtp.mail.yahoo.com",
		SMTPPort: 465,
		IMAPHost: "imap.mail.yahoo.com",
		IMAPPort: 993,
	},
	"zoho": {
		SMTPHost: "smtp.zoho.com",
		SMTPPort: 465,
		IMAPHost: "imap.zoho.com",
		IMAPPort: 993,
	},
	"icloud": {
		SMTPHost: "smtp.mail.me.com",
		SMTPPort: 587,
		IMAPHost: "imap.mail.me.com",
		IMAPPort: 993,
	},
	"aol": {
		SMTPHost: "smtp.aol.com",
		SMTPPort: 587,
		IMAPHost: "imap.aol.com",
		IMAPPort: 993,
	},
	"secureserver": {
		SMTPHost: "smtpout.secureserver.net",
		SMTPPort: 465,
		IMAPHost: "imap.secureserver.net",
		IMAPPort: 993,
	},
}

// AddServer allows adding custom provider configuration for both SMTP and IMAP.
func AddServer(name, smtpHost string, smtpPort int, imapHost string, imapPort int) {
	mailServerConfigs[strings.ToLower(name)] = MailServerConfig{
		SMTPHost: smtpHost,
		SMTPPort: smtpPort,
		IMAPHost: imapHost,
		IMAPPort: imapPort,
	}
}

// GetServer gets the complete mail server config for a provider (SMTP and IMAP).
// The function name GetServer is maintained for backward compatibility with the original intent.
func GetServer(provider string) (MailServerConfig, error) {
	p := strings.ToLower(provider)
	cfg, ok := mailServerConfigs[p]
	if !ok {
		return MailServerConfig{}, fmt.Errorf("unknown email provider: %s", provider)
	}
	return cfg, nil
}
