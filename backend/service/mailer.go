package service

import (
	"fmt"
	"log"
	"strconv"

	mail "github.com/wneessen/go-mail"

	"github.com/jaykay/vereinstool/backend/config"
)

type Mailer struct {
	cfg *config.Config
}

func NewMailer(cfg *config.Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) send(to, subject, body string) error {
	msg := mail.NewMsg()
	if err := msg.From(m.cfg.SMTPFrom); err != nil {
		return fmt.Errorf("setting from: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("setting to: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, body)

	port, err := strconv.Atoi(m.cfg.SMTPPort)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %w", err)
	}

	opts := []mail.Option{
		mail.WithPort(port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(m.cfg.SMTPUser),
		mail.WithPassword(m.cfg.SMTPPassword),
	}

	// Use TLS for port 465, STARTTLS for others
	if port == 465 {
		opts = append(opts, mail.WithSSL())
	}

	// Skip auth if no credentials (local dev with MailHog)
	if m.cfg.SMTPUser == "" {
		opts = []mail.Option{
			mail.WithPort(port),
			mail.WithTLSPolicy(mail.NoTLS),
		}
	}

	client, err := mail.NewClient(m.cfg.SMTPHost, opts...)
	if err != nil {
		return fmt.Errorf("creating mail client: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("sending mail: %w", err)
	}

	return nil
}

// SendPasswordReset sends a password reset email.
func (m *Mailer) SendPasswordReset(to, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", m.cfg.AppURL, token)
	body := fmt.Sprintf(`<h2>Passwort zurücksetzen</h2>
<p>Klicke auf den folgenden Link, um dein Passwort zurückzusetzen:</p>
<p><a href="%s">Passwort zurücksetzen</a></p>
<p>Der Link ist 1 Stunde gültig.</p>
<p>Falls du diese E-Mail nicht angefordert hast, kannst du sie ignorieren.</p>`, resetURL)

	if err := m.send(to, "Passwort zurücksetzen – Vereinstool", body); err != nil {
		log.Printf("Failed to send password reset email to %s: %v", to, err)
		return err
	}
	return nil
}

// SendInvitation sends an invitation email with temporary password.
func (m *Mailer) SendInvitation(to, name, tempPassword string) error {
	loginURL := fmt.Sprintf("%s/login", m.cfg.AppURL)
	body := fmt.Sprintf(`<h2>Willkommen beim Vereinstool, %s!</h2>
<p>Du wurdest als Vorstandsmitglied eingeladen.</p>
<p><strong>Deine Zugangsdaten:</strong></p>
<ul>
<li>E-Mail: %s</li>
<li>Temporäres Passwort: <code>%s</code></li>
</ul>
<p><a href="%s">Jetzt anmelden</a></p>
<p>Bitte ändere dein Passwort nach der ersten Anmeldung.</p>`, name, to, tempPassword, loginURL)

	if err := m.send(to, "Einladung zum Vereinstool", body); err != nil {
		log.Printf("Failed to send invitation email to %s: %v", to, err)
		return err
	}
	return nil
}
