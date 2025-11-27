package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/resend/resend-go/v3"
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

// Resend Config dari environment variables
type ResendConfig struct {
	APIKey      string
	SenderEmail string
}

// Ambil konfigurasi Resend dari environment
func getResendConfig() (*ResendConfig, error) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("RESEND_API_KEY is not set in environment variables")
	}

	senderEmail := os.Getenv("RESEND_SENDER_EMAIL")
	if senderEmail == "" {
		return nil, fmt.Errorf("RESEND_SENDER_EMAIL is not set in environment variables")
	}

	// Tambahkan logging untuk konfigurasi Resend
	log.Printf("📧 Resend Config - Sender Email: %s", senderEmail)

	return &ResendConfig{
		APIKey:      apiKey,
		SenderEmail: senderEmail,
	}, nil
}

// Fungsi untuk mengirim email menggunakan Resend API dengan library resmi
func sendResendEmail(config EmailConfig, resendConfig *ResendConfig) error {
	log.Printf("📧 Sending email using Resend API")
	log.Printf("📧 Email config - To: %s, Subject: %s", config.To, config.Subject)

	// Buat client Resend
	client := resend.NewClient(resendConfig.APIKey)

	// Siapkan parameter
	params := &resend.SendEmailRequest{
		From:    resendConfig.SenderEmail,
		To:      []string{config.To},
		Subject: config.Subject,
	}

	// Tambahkan HTML atau Text
	if config.HTML != "" {
		params.Html = config.HTML
	}
	if config.Text != "" {
		params.Text = config.Text
	}

	// Tambahkan CC dan BCC jika ada
	if len(config.CC) > 0 {
		params.Cc = config.CC
	}
	if len(config.BCC) > 0 {
		params.Bcc = config.BCC
	}

	// Kirim email
	log.Printf("📧 Sending HTTP request to Resend API")
	sent, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("❌ Failed to send email using Resend: %v", err)
		return fmt.Errorf("failed to send email using Resend: %w", err)
	}

	log.Printf("✅ Email sent successfully using Resend API, ID: %s", sent.Id)
	return nil
}

// Kirim email dengan retry mechanism (hanya Resend)
func SendEmail(config EmailConfig) error {
	// Validasi konten
	if config.HTML == "" && config.Text == "" {
		return fmt.Errorf("either HTML or Text content must be provided")
	}

	// Ambil konfigurasi Resend
	resendConfig, err := getResendConfig()
	if err != nil {
		log.Printf("❌ Failed to get Resend config: %v", err)
		return fmt.Errorf("failed to get Resend config: %w", err)
	}

	log.Printf("📧 Using Resend API to send email")

	// Retry mechanism
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		log.Printf("📧 Attempt %d to send email to %s using Resend", i+1, config.To)

		// Try to send email
		err := sendResendEmail(config, resendConfig)
		if err == nil {
			log.Printf("✅ Email successfully sent to %s using Resend", config.To)
			return nil
		}

		lastErr = err
		log.Printf("❌ Attempt %d failed to send email using Resend: %v", i+1, err)

		// Wait before retrying (exponential backoff)
		if i < maxRetries-1 {
			waitTime := time.Duration(2<<uint(i)) * time.Second // 2, 4, 8 seconds
			log.Printf("⏳ Waiting %v before retrying...", waitTime)
			time.Sleep(waitTime)
		}
	}

	log.Printf("❌ All attempts failed to send email to %s using Resend", config.To)
	return fmt.Errorf("failed to send email after %d attempts using Resend, last error: %w", maxRetries, lastErr)
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
	// Gunakan email tujuan yang ditentukan
	adminEmail := "sanzwicaksono@gmail.com" // Email tujuan yang ditentukan

	// Jika ada environment variable ADMIN_NOTIFICATION_EMAIL, gunakan itu sebagai fallback
	if envEmail := os.Getenv("ADMIN_NOTIFICATION_EMAIL"); envEmail != "" {
		adminEmail = envEmail
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

	// Gunakan email tujuan yang ditentukan
	adminEmail := "sanzwicaksono@gmail.com" // Email tujuan yang ditentukan

	// Jika ada environment variable ADMIN_NOTIFICATION_EMAIL, gunakan itu sebagai fallback
	if envEmail := os.Getenv("ADMIN_NOTIFICATION_EMAIL"); envEmail != "" {
		adminEmail = envEmail
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
