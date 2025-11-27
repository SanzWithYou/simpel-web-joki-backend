package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"math"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// Email config
type EmailConfig struct {
	To      string
	Subject string
	HTML    string
	Text    string
	CC      []string
	BCC     []string
}

// SMTP Config dari environment variables
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	FromName string
}

// Ambil konfigurasi SMTP dari environment
func getSMTPConfig() (*SMTPConfig, error) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return nil, fmt.Errorf("SMTP_HOST is not set in environment variables")
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587" // default port
	}

	username := os.Getenv("SMTP_USERNAME")
	if username == "" {
		return nil, fmt.Errorf("SMTP_USERNAME is not set in environment variables")
	}

	password := os.Getenv("SMTP_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD is not set in environment variables")
	}

	from := os.Getenv("SMTP_FROM_EMAIL")
	if from == "" {
		from = username // gunakan username sebagai default
	}

	fromName := os.Getenv("SMTP_FROM_NAME")
	if fromName == "" {
		fromName = "Zem Store"
	}

	// Tambahkan logging untuk konfigurasi SMTP
	log.Printf("📧 SMTP Config - Host: %s, Port: %s, Username: %s, From: %s", host, port, username, from)

	return &SMTPConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		FromName: fromName,
	}, nil
}

// Build email message dengan format MIME
func buildEmailMessage(config EmailConfig, smtpConfig *SMTPConfig) []byte {
	var buf bytes.Buffer

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", smtpConfig.FromName, smtpConfig.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", config.To))

	if len(config.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(config.CC, ", ")))
	}

	if len(config.BCC) > 0 {
		buf.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(config.BCC, ", ")))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", config.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")

	// Jika ada HTML dan Text, gunakan multipart/alternative
	if config.HTML != "" && config.Text != "" {
		boundary := "boundary-" + time.Now().Format("20060102150405")
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		buf.WriteString("\r\n")

		// Plain text part
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(config.Text)
		buf.WriteString("\r\n")

		// HTML part
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(config.HTML)
		buf.WriteString("\r\n")

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if config.HTML != "" {
		// Hanya HTML
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(config.HTML)
	} else if config.Text != "" {
		// Hanya Text
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(config.Text)
	}

	return buf.Bytes()
}

// PERBAIKAN: Fungsi untuk membuat alamat server yang kompatibel dengan IPv6
func formatServerAddress(host, port string) string {
	// Jika host adalah alamat IPv6 (mengandung : dan tidak diapit [])
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") && !strings.HasSuffix(host, "]") {
		return fmt.Sprintf("[%s]:%s", host, port)
	}
	// Untuk IPv4 atau hostname, gunakan format biasa
	return fmt.Sprintf("%s:%s", host, port)
}

// Kirim email dengan SMTP
func sendSMTPEmail(config EmailConfig, smtpConfig *SMTPConfig) error {
	// Build recipients list
	recipients := []string{config.To}
	recipients = append(recipients, config.CC...)
	recipients = append(recipients, config.BCC...)

	// Build message
	message := buildEmailMessage(config, smtpConfig)

	// Setup authentication
	auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)

	// PERBAIKAN: Gunakan fungsi formatServerAddress untuk menangani IPv6
	serverAddr := formatServerAddress(smtpConfig.Host, smtpConfig.Port)

	log.Printf("🔌 Attempting to connect to SMTP server: %s", serverAddr)

	// Gunakan timeout untuk koneksi SMTP dengan cara yang benar
	// Buat dialer dengan timeout
	dialer := &net.Dialer{
		Timeout: 15 * time.Second,
	}

	// Buat koneksi dengan timeout
	conn, err := dialer.Dial("tcp", serverAddr)
	if err != nil {
		log.Printf("❌ Failed to connect to SMTP server: %v", err)
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	log.Printf("✅ Connected to SMTP server successfully")

	// Buat client SMTP dari koneksi
	client, err := smtp.NewClient(conn, smtpConfig.Host)
	if err != nil {
		log.Printf("❌ Failed to create SMTP client: %v", err)
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Start TLS
	tlsConfig := &tls.Config{
		ServerName: smtpConfig.Host,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		log.Printf("❌ Failed to start TLS: %v", err)
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	log.Printf("✅ TLS started successfully")

	// Authenticate
	if err = client.Auth(auth); err != nil {
		log.Printf("❌ Failed to authenticate: %v", err)
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	log.Printf("✅ Authentication successful")

	// Set sender
	if err = client.Mail(smtpConfig.From); err != nil {
		log.Printf("❌ Failed to set sender: %v", err)
		return fmt.Errorf("failed to set sender: %w", err)
	}

	log.Printf("✅ Sender set successfully")

	// Set recipients
	for _, recipient := range recipients {
		if err = client.Rcpt(recipient); err != nil {
			log.Printf("❌ Failed to set recipient %s: %v", recipient, err)
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
		log.Printf("✅ Recipient set successfully: %s", recipient)
	}

	// Send message
	w, err := client.Data()
	if err != nil {
		log.Printf("❌ Failed to send DATA command: %v", err)
		return fmt.Errorf("failed to send DATA command: %w", err)
	}

	_, err = w.Write(message)
	if err != nil {
		log.Printf("❌ Failed to write message: %v", err)
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		log.Printf("❌ Failed to close writer: %v", err)
		return fmt.Errorf("failed to close writer: %w", err)
	}

	log.Printf("✅ Message sent successfully")

	err = client.Quit()
	if err != nil {
		log.Printf("❌ Failed to quit SMTP session: %v", err)
		return fmt.Errorf("failed to quit SMTP session: %w", err)
	}

	log.Printf("✅ SMTP session closed successfully")
	return nil
}

// Kirim email dengan retry mechanism
func SendEmail(config EmailConfig) error {
	// Validasi konten
	if config.HTML == "" && config.Text == "" {
		return fmt.Errorf("either HTML or Text content must be provided")
	}

	// Ambil SMTP config
	smtpConfig, err := getSMTPConfig()
	if err != nil {
		log.Printf("❌ Failed to get SMTP config: %v", err)
		return fmt.Errorf("failed to get SMTP config: %w", err)
	}

	// Kurangi jumlah retry dan waktu tunggu
	maxRetries := 2
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		log.Printf("📧 Attempt %d to send email to %s", i+1, config.To)

		// Try to send email
		err := sendSMTPEmail(config, smtpConfig)
		if err == nil {
			log.Printf("✅ Email successfully sent to %s", config.To)
			return nil
		}

		lastErr = err
		log.Printf("❌ Attempt %d failed to send email: %v", i+1, err)

		// Wait before retrying (exponential backoff)
		if i < maxRetries-1 {
			waitTime := time.Duration(math.Pow(2, float64(i+1))) * time.Second
			log.Printf("⏳ Waiting %v before retrying...", waitTime)
			time.Sleep(waitTime)
		}
	}

	log.Printf("❌ All attempts failed to send email to %s", config.To)
	return fmt.Errorf("failed to send email after %d attempts, last error: %w", maxRetries, lastErr)
}

// Alternatif: Kirim email dengan timeout (wrapper untuk SendEmail)
func SendEmailWithTimeout(config EmailConfig, timeout time.Duration) error {
	// Channel untuk hasil
	done := make(chan error, 1)

	// Jalankan pengiriman email di goroutine
	go func() {
		done <- SendEmail(config)
	}()

	// Tunggu dengan timeout
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("email sending timeout after %v", timeout)
	}
}

// Pisahkan pembuatan template email
func GenerateOrderEmailTemplate(orderID uint, username, jokiType, buktiTransferPath string) string {
	// Ambil URL
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:3000"
	}

	// Ekstrak path
	parts := strings.Split(buktiTransferPath, "/")
	if len(parts) < 5 {
		log.Printf("❌ Invalid URL format: %s", buktiTransferPath)
		return ""
	}
	relativePath := strings.Join(parts[4:], "/")

	// Buat URL bukti
	proofUrl := fmt.Sprintf("%s/api/files/uploads/%s", backendURL, relativePath)

	// Ambil nama app
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "Zem - Store"
	}

	// Template email
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Order Baru Diterima</title>
  <style type="text/css">
    @media only screen and (max-width: 700px) {
      .email-container {
        width: 100%% !important;
        border-radius: 0 !important;
        border-left: none !important;
        border-right: none !important;
      }
      .header {
        padding: 25px 20px !important;
      }
      .header h1 {
        font-size: 22px !important;
      }
      .header p {
        font-size: 13px !important;
      }
      .content {
        padding: 25px 20px !important;
      }
      .details-card {
        padding: 18px !important;
      }
      .details-card h2 {
        font-size: 15px !important;
        margin-bottom: 15px !important;
      }
      .detail-row table {
        width: 100%% !important;
      }
      .detail-row td {
        padding: 10px 0 !important;
        border-bottom: 1px solid #333333 !important;
      }
      .detail-label {
        width: 40%% !important;
        vertical-align: top !important;
      }
      .detail-value {
        width: 60%% !important;
        text-align: right !important;
      }
      .detail-value span {
        word-break: break-word !important;
      }
      .cta-button {
        padding: 12px 24px !important;
        font-size: 14px !important;
        display: inline-block !important;
        width: auto !important;
      }
      .footer {
        padding: 20px 15px !important;
      }
      .footer p {
        font-size: 11px !important;
      }
    }
  </style>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background: #0a0a0a;">

  <!-- Main Container -->
  <table width="100%%" cellpadding="0" cellspacing="0" style="background: #0a0a0a;">
    <tr>
      <td align="center">

        <!-- Email Card -->
        <table class="email-container" width="700" cellpadding="0" cellspacing="0" style="background: #1a1a1a; border-radius: 16px; box-shadow: 0 20px 60px rgba(0,0,0,0.5); overflow: hidden; max-width: 100%%; border: 1px solid #2a2a2a;">

          <!-- Header with Dark Theme -->
          <tr>
            <td class="header" style="background: linear-gradient(135deg, #1a1a1a 0%%, #2d2d2d 100%%); padding: 40px 30px; text-align: center; border-bottom: 1px solid #333333;">
              <h1 style="color: #ffffff; font-size: 28px; font-weight: 700; margin: 0 0 8px 0; letter-spacing: -1px;">Order Baru</h1>
            </td>
          </tr>

          <!-- Content -->
          <tr>
            <td class="content" style="padding: 35px 30px;">

              <!-- Order Details Card -->
              <table width="100%%" cellpadding="0" cellspacing="0" class="details-card" style="background: #222222; border-radius: 12px; padding: 25px; margin-bottom: 25px; border: 1px solid #333333;">
                <tr>
                  <td>
                    <h2 style="color: #ffffff; font-size: 17px; font-weight: 700; margin: 0 0 20px 0; letter-spacing: -0.5px;">Detail Pesanan</h2>

                    <!-- Order ID -->
                    <table width="100%%" cellpadding="0" cellspacing="0" class="detail-row" style="margin-bottom: 18px;">
                      <tr>
                        <td class="detail-label" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="color: #888888; font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px;">Order ID</span>
                        </td>
                        <td class="detail-value" align="right" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="background: #ffffff; color: #000000; padding: 6px 16px; border-radius: 6px; font-size: 14px; font-weight: 700;">#%d</span>
                        </td>
                      </tr>
                    </table>

                    <!-- Username -->
                    <table width="100%%" cellpadding="0" cellspacing="0" class="detail-row" style="margin-bottom: 18px;">
                      <tr>
                        <td class="detail-label" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="color: #888888; font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px;">Username</span>
                        </td>
                        <td class="detail-value" align="right" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="color: #ffffff; font-size: 15px; font-weight: 600;">%s</span>
                        </td>
                      </tr>
                    </table>

                    <!-- Jenis Joki -->
                    <table width="100%%" cellpadding="0" cellspacing="0" class="detail-row" style="margin-bottom: 18px;">
                      <tr>
                        <td class="detail-label" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="color: #888888; font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px;">Jenis Layanan</span>
                        </td>
                        <td class="detail-value" align="right" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="color: #ffffff; font-size: 15px; font-weight: 600;">%s</span>
                        </td>
                      </tr>
                    </table>

                    <!-- Waktu -->
                    <table width="100%%" cellpadding="0" cellspacing="0" class="detail-row">
                      <tr>
                        <td class="detail-label" style="padding: 12px 0;">
                          <span style="color: #888888; font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px;">Waktu Order</span>
                        </td>
                        <td class="detail-value" align="right" style="padding: 12px 0;">
                          <span style="color: #cccccc; font-size: 14px; font-weight: 500;">%s</span>
                        </td>
                      </tr>
                    </table>

                  </td>
                </tr>
              </table>

              <!-- CTA Button -->
              <table width="100%%" cellpadding="0" cellspacing="0">
                <tr>
                  <td align="center" style="padding: 15px 0;">
                    <a href="%s" target="_blank" class="cta-button" style="display: inline-block; background: #ffffff; color: #000000; text-decoration: none; padding: 14px 40px; border-radius: 8px; font-weight: 700; font-size: 15px; box-shadow: 0 4px 20px rgba(255, 255, 255, 0.1);">
                      Lihat Bukti Pembayaran
                    </a>
                  </td>
                </tr>
              </table>

            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td class="footer" style="background: #151515; padding: 25px 30px; text-align: center; border-top: 1px solid #2a2a2a;">
              <p style="margin: 0 0 8px 0; color: #666666; font-size: 12px; line-height: 1.6;">
                Email ini dikirim secara otomatis oleh <strong style="color: #ffffff;">%s</strong>
              </p>
              <p style="margin: 0; color: #555555; font-size: 11px;">
                © 2025 %s. All rights reserved.
              </p>
            </td>
          </tr>

        </table>

      </td>
    </tr>
  </table>

</body>
</html>
`, orderID, username, jokiType, time.Now().Format("02 January 2006, 15:04 WIB"), proofUrl, appName, appName)
}

// Notifikasi order
func SendNewOrderNotificationEmail(orderID uint, username, jokiType, buktiTransferPath string) error {
	// Ambil email admin
	adminEmail := os.Getenv("ADMIN_NOTIFICATION_EMAIL")
	if adminEmail == "" {
		log.Printf("❌ ADMIN_NOTIFICATION_EMAIL is not set in environment variables")
		return fmt.Errorf("ADMIN_NOTIFICATION_EMAIL is not set in environment variables")
	}

	log.Printf("📧 Preparing to send order notification email to %s", adminEmail)

	config := EmailConfig{
		To:      adminEmail,
		Subject: fmt.Sprintf("Order Baru #%d - Segera Proses!", orderID),
		HTML:    GenerateOrderEmailTemplate(orderID, username, jokiType, buktiTransferPath),
	}

	// Gunakan fungsi dengan retry mechanism
	return SendEmail(config)
}

// Email layanan kustom
func SendCustomServiceRequestEmail(requestID uint, name, email, service string) error {
	// Ambil nama app
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "Zem - Store"
	}

	// Template email
	emailContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Permintaan Layanan Kustom</title>
  <style type="text/css">
    @media only screen and (max-width: 700px) {
      .email-container {
        width: 100%% !important;
        border-radius: 0 !important;
        border-left: none !important;
        border-right: none !important;
      }
      .header {
        padding: 25px 20px !important;
      }
      .header h1 {
        font-size: 22px !important;
      }
      .header p {
        font-size: 13px !important;
      }
      .content {
        padding: 25px 20px !important;
      }
      .details-card {
        padding: 18px !important;
      }
      .details-card h2 {
        font-size: 15px !important;
        margin-bottom: 15px !important;
      }
      .detail-row table {
        width: 100%% !important;
      }
      .detail-row td {
        padding: 10px 0 !important;
        border-bottom: 1px solid #333333 !important;
      }
      .detail-label {
        width: 40%% !important;
        vertical-align: top !important;
      }
      .detail-value {
        width: 60%% !important;
        text-align: right !important;
      }
      .detail-value span {
        word-break: break-word !important;
      }
      .footer {
        padding: 20px 15px !important;
      }
      .footer p {
        font-size: 11px !important;
      }
    }
  </style>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background: #0a0a0a;">

  <!-- Main Container -->
  <table width="100%%" cellpadding="0" cellspacing="0" style="background: #0a0a0a;">
    <tr>
      <td align="center">

        <!-- Email Card -->
        <table class="email-container" width="700" cellpadding="0" cellspacing="0" style="background: #1a1a1a; border-radius: 16px; box-shadow: 0 20px 60px rgba(0,0,0,0.5); overflow: hidden; max-width: 100%%; border: 1px solid #2a2a2a;">

          <!-- Header with Dark Theme -->
          <tr>
            <td class="header" style="background: linear-gradient(135deg, #1a1a1a 0%%, #2d2d2d 100%%); padding: 40px 30px; text-align: center; border-bottom: 1px solid #333333;">
              <h1 style="color: #ffffff; font-size: 28px; font-weight: 700; margin: 0 0 8px 0; letter-spacing: -1px;">Permintaan Layanan Kustom</h1>
              <p style="color: #888888; font-size: 15px; margin: 0; font-weight: 400;">Pelanggan meminta layanan baru</p>
            </td>
          </tr>

          <!-- Content -->
          <tr>
            <td class="content" style="padding: 35px 30px;">

              <!-- Request Details Card -->
              <table width="100%%" cellpadding="0" cellspacing="0" class="details-card" style="background: #222222; border-radius: 12px; padding: 25px; margin-bottom: 25px; border: 1px solid #333333;">
                <tr>
                  <td>
                    <h2 style="color: #ffffff; font-size: 17px; font-weight: 700; margin: 0 0 20px 0; letter-spacing: -0.5px;">Detail Permintaan</h2>

                    <!-- Request ID -->
                    <table width="100%%" cellpadding="0" cellspacing="0" class="detail-row" style="margin-bottom: 18px;">
                      <tr>
                        <td class="detail-label" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="color: #888888; font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px;">Request ID</span>
                        </td>
                        <td class="detail-value" align="right" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="background: #ffffff; color: #000000; padding: 6px 16px; border-radius: 6px; font-size: 14px; font-weight: 700;">#%d</span>
                        </td>
                      </tr>
                    </table>

                    <!-- Name -->
                    <table width="100%%" cellpadding="0" cellspacing="0" class="detail-row" style="margin-bottom: 18px;">
                      <tr>
                        <td class="detail-label" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="color: #888888; font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px;">Nama</span>
                        </td>
                        <td class="detail-value" align="right" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="color: #ffffff; font-size: 15px; font-weight: 600;">%s</span>
                        </td>
                      </tr>
                    </table>

                    <!-- Email -->
                    <table width="100%%" cellpadding="0" cellspacing="0" class="detail-row" style="margin-bottom: 18px;">
                      <tr>
                        <td class="detail-label" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="color: #888888; font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px;">Email</span>
                        </td>
                        <td class="detail-value" align="right" style="padding: 12px 0; border-bottom: 1px solid #333333;">
                          <span style="color: #ffffff; font-size: 15px; font-weight: 600;">%s</span>
                        </td>
                      </tr>
                    </table>

                    <!-- Service Description -->
                    <table width="100%%" cellpadding="0" cellspacing="0" class="detail-row">
                      <tr>
                        <td class="detail-label" style="padding: 12px 0;">
                          <span style="color: #888888; font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px;">Deskripsi Layanan</span>
                        </td>
                        <td class="detail-value" align="right" style="padding: 12px 0;">
                          <span style="color: #cccccc; font-size: 14px; font-weight: 500; white-space: pre-wrap;">%s</span>
                        </td>
                      </tr>
                    </table>

                  </td>
                </tr>
              </table>

            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td class="footer" style="background: #151515; padding: 25px 30px; text-align: center; border-top: 1px solid #2a2a2a;">
              <p style="margin: 0 0 8px 0; color: #666666; font-size: 12px; line-height: 1.6;">
                Email ini dikirim secara otomatis oleh <strong style="color: #ffffff;">%s</strong>
              </p>
              <p style="margin: 0; color: #555555; font-size: 11px;">
                © 2025 %s. All rights reserved.
              </p>
            </td>
          </tr>

        </table>

      </td>
    </tr>
  </table>

</body>
</html>
`, requestID, name, email, service, appName, appName)

	// Ambil email admin
	adminEmail := os.Getenv("ADMIN_NOTIFICATION_EMAIL")
	if adminEmail == "" {
		log.Printf("❌ ADMIN_NOTIFICATION_EMAIL is not set in environment variables")
		return fmt.Errorf("ADMIN_NOTIFICATION_EMAIL is not set in environment variables")
	}

	log.Printf("📧 Preparing to send custom service request email to %s", adminEmail)

	config := EmailConfig{
		To:      adminEmail,
		Subject: fmt.Sprintf("Permintaan Layanan Kustom #%d", requestID),
		HTML:    emailContent,
	}

	// Gunakan fungsi dengan retry mechanism
	return SendEmail(config)
}
