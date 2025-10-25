// Package email provides email adapter implementations
package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/message"
	"github.com/thanhphuchuynh/dear-future/pkg/effects"
)

// SMTPConfig holds configuration for SMTP email service
type SMTPConfig struct {
	Host          string // SMTP server host (e.g., smtp.gmail.com)
	Port          string // SMTP server port (e.g., 587 for TLS, 465 for SSL)
	Username      string // SMTP username (email address)
	Password      string // SMTP password (app password for Gmail)
	FromEmail     string // Email address to send from
	FromName      string // Display name for sender
	UseTLS        bool   // Use TLS (STARTTLS)
	SkipTLSVerify bool   // Skip TLS certificate verification (for dev only)
}

// SMTPEmailService implements EmailService using SMTP
// Supports Gmail and other SMTP providers
type SMTPEmailService struct {
	config SMTPConfig
}

// NewSMTPEmailService creates a new SMTP email service
func NewSMTPEmailService(config SMTPConfig) (*SMTPEmailService, error) {
	if config.Host == "" || config.Port == "" || config.Username == "" || config.Password == "" {
		return nil, fmt.Errorf("SMTP configuration is incomplete: host, port, username, and password are required")
	}

	if config.FromEmail == "" {
		config.FromEmail = config.Username
	}

	if config.FromName == "" {
		config.FromName = "Dear Future"
	}

	return &SMTPEmailService{
		config: config,
	}, nil
}

// SendMessage sends a scheduled message via email
func (s *SMTPEmailService) SendMessage(ctx context.Context, deliveryInfo message.MessageDeliveryInfo) common.Result[effects.EmailResult] {
	subject := deliveryInfo.Subject
	body := s.buildMessageBody(deliveryInfo)
	recipient := deliveryInfo.RecipientEmail

	err := s.sendEmail(ctx, recipient, subject, body)
	if err != nil {
		return common.Ok(effects.EmailResult{
			MessageID: deliveryInfo.Message.ID().String(),
			Status:    effects.EmailStatusFailed,
			SentAt:    time.Now(),
			Error:     common.Some(err.Error()),
			Recipient: recipient,
			Subject:   subject,
		})
	}

	return common.Ok(effects.EmailResult{
		MessageID: deliveryInfo.Message.ID().String(),
		Status:    effects.EmailStatusSent,
		SentAt:    time.Now(),
		Error:     common.None[string](),
		Recipient: recipient,
		Subject:   subject,
	})
}

// SendVerificationEmail sends an email verification email
func (s *SMTPEmailService) SendVerificationEmail(ctx context.Context, email, verificationToken string) common.Result[effects.EmailResult] {
	subject := "Verify your Dear Future account"
	body := s.buildVerificationEmailBody(email, verificationToken)

	err := s.sendEmail(ctx, email, subject, body)
	if err != nil {
		return common.Ok(effects.EmailResult{
			MessageID: "",
			Status:    effects.EmailStatusFailed,
			SentAt:    time.Now(),
			Error:     common.Some(err.Error()),
			Recipient: email,
			Subject:   subject,
		})
	}

	return common.Ok(effects.EmailResult{
		MessageID: "",
		Status:    effects.EmailStatusSent,
		SentAt:    time.Now(),
		Error:     common.None[string](),
		Recipient: email,
		Subject:   subject,
	})
}

// SendPasswordResetEmail sends a password reset email
func (s *SMTPEmailService) SendPasswordResetEmail(ctx context.Context, email, resetToken string) common.Result[effects.EmailResult] {
	subject := "Reset your Dear Future password"
	body := s.buildPasswordResetEmailBody(email, resetToken)

	err := s.sendEmail(ctx, email, subject, body)
	if err != nil {
		return common.Ok(effects.EmailResult{
			MessageID: "",
			Status:    effects.EmailStatusFailed,
			SentAt:    time.Now(),
			Error:     common.Some(err.Error()),
			Recipient: email,
			Subject:   subject,
		})
	}

	return common.Ok(effects.EmailResult{
		MessageID: "",
		Status:    effects.EmailStatusSent,
		SentAt:    time.Now(),
		Error:     common.None[string](),
		Recipient: email,
		Subject:   subject,
	})
}

// ValidateEmailConfiguration validates the SMTP configuration by attempting to connect
func (s *SMTPEmailService) ValidateEmailConfiguration(ctx context.Context) common.Result[bool] {
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	// Attempt to connect and authenticate
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// For TLS connections
	if s.config.UseTLS {
		tlsConfig := &tls.Config{
			ServerName:         s.config.Host,
			InsecureSkipVerify: s.config.SkipTLSVerify,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return common.Err[bool](fmt.Errorf("failed to connect to SMTP server: %w", err))
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.config.Host)
		if err != nil {
			return common.Err[bool](fmt.Errorf("failed to create SMTP client: %w", err))
		}
		defer client.Quit()

		if err := client.Auth(auth); err != nil {
			return common.Err[bool](fmt.Errorf("SMTP authentication failed: %w", err))
		}

		return common.Ok(true)
	}

	// For non-TLS connections (STARTTLS)
	client, err := smtp.Dial(addr)
	if err != nil {
		return common.Err[bool](fmt.Errorf("failed to connect to SMTP server: %w", err))
	}
	defer client.Quit()

	if err := client.Auth(auth); err != nil {
		return common.Err[bool](fmt.Errorf("SMTP authentication failed: %w", err))
	}

	return common.Ok(true)
}

// sendEmail sends an email via SMTP
func (s *SMTPEmailService) sendEmail(ctx context.Context, to, subject, body string) error {
	from := s.config.FromEmail
	msg := s.buildEmailMessage(from, to, subject, body)

	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	// For Gmail and most modern SMTP servers, use STARTTLS on port 587
	if s.config.UseTLS && s.config.Port == "465" {
		// Use TLS directly (SSL)
		return s.sendEmailTLS(addr, auth, from, []string{to}, msg)
	}

	// Use STARTTLS (port 587)
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

// sendEmailTLS sends email using TLS (for port 465)
func (s *SMTPEmailService) sendEmailTLS(addr string, auth smtp.Auth, from string, to []string, msg string) error {
	tlsConfig := &tls.Config{
		ServerName:         s.config.Host,
		InsecureSkipVerify: s.config.SkipTLSVerify,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS connection failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("MAIL command failed: %w", err)
	}

	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("RCPT command failed: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("writing message failed: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("closing data writer failed: %w", err)
	}

	return nil
}

// buildEmailMessage builds a properly formatted email message
func (s *SMTPEmailService) buildEmailMessage(from, to, subject, body string) string {
	fromHeader := from
	if s.config.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", s.config.FromName, from)
	}

	return fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", fromHeader, to, subject, body)
}

// buildMessageBody builds the HTML body for a scheduled message
func (s *SMTPEmailService) buildMessageBody(deliveryInfo message.MessageDeliveryInfo) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .message { background: white; padding: 20px; border-left: 4px solid #667eea; margin: 20px 0; border-radius: 5px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .scheduled-date { color: #667eea; font-weight: bold; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📬 A Message from Your Past Self</h1>
        <p>Scheduled for: <span class="scheduled-date">%s</span></p>
    </div>
    <div class="content">
        <h2>%s</h2>
        <div class="message">
            <p>%s</p>
        </div>
        <div class="footer">
            <p>This message was sent by <strong>Dear Future</strong></p>
            <p>Your message to tomorrow, delivered today.</p>
        </div>
    </div>
</body>
</html>
`, deliveryInfo.ScheduledTime.Format("January 2, 2006 at 3:04 PM"), deliveryInfo.Subject, deliveryInfo.Body)
}

// buildVerificationEmailBody builds the verification email HTML body
func (s *SMTPEmailService) buildVerificationEmailBody(email, token string) string {
	verificationURL := fmt.Sprintf("https://dearfuture.app/verify?token=%s", token)
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 15px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Welcome to Dear Future!</h1>
    </div>
    <div class="content">
        <p>Hello,</p>
        <p>Thank you for signing up! Please verify your email address to start sending messages to your future self.</p>
        <center>
            <a href="%s" class="button">Verify Email Address</a>
        </center>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #667eea;">%s</p>
        <div class="footer">
            <p>If you didn't create this account, you can safely ignore this email.</p>
        </div>
    </div>
</body>
</html>
`, verificationURL, verificationURL)
}

// buildPasswordResetEmailBody builds the password reset email HTML body
func (s *SMTPEmailService) buildPasswordResetEmailBody(email, token string) string {
	resetURL := fmt.Sprintf("https://dearfuture.app/reset-password?token=%s", token)
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 15px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Password Reset Request</h1>
    </div>
    <div class="content">
        <p>Hello,</p>
        <p>We received a request to reset your password for your Dear Future account.</p>
        <center>
            <a href="%s" class="button">Reset Password</a>
        </center>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #667eea;">%s</p>
        <div class="warning">
            <strong>Important:</strong> This link will expire in 1 hour for security reasons.
        </div>
        <div class="footer">
            <p>If you didn't request a password reset, you can safely ignore this email.</p>
            <p>Your password will not be changed unless you click the link above.</p>
        </div>
    </div>
</body>
</html>
`, resetURL, resetURL)
}
