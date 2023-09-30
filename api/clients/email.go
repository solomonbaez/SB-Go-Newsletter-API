package clients

import (
	"github.com/gin-gonic/gin"
	"github.com/go-gomail/gomail"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/configs"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

type EmailClient interface {
	SendEmail(c *gin.Context, email *Message) error
}

type SMTPClient struct {
	SmtpServer   string // export for testing
	smtpPort     int
	smtpUsername string
	smtpPassword string
	sender       models.SubscriberEmail
}

type Message struct {
	Recipient models.SubscriberEmail
	Subject   string
	Text      string
	Html      string
}

func NewSMTPClient(file string) (*SMTPClient, error) {
	cfg, e := configs.ConfigureEmailClient(file)
	if e != nil {
		return nil, e
	}

	// validate cfg sender email
	s := cfg.Sender
	sender, e := models.ParseEmail(s)
	if e != nil {
		return nil, e
	}

	client := &SMTPClient{
		cfg.Server,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		sender,
	}

	return client, nil
}

func (client *SMTPClient) SendEmail(c *gin.Context, message *Message) error {
	requestID := c.GetString("requestID")

	m := gomail.NewMessage()
	m.SetHeader("From", client.sender.String())
	m.SetHeader("To", message.Recipient.String())
	m.SetHeader("Subject", message.Subject)
	m.SetBody("text/plain", message.Text)
	m.AddAlternative("text/html", message.Html)

	log.Info().
		Str("requestID", requestID).
		Str("sender", client.sender.String()).
		Str("recipient", message.Recipient.String()).
		Msg("Attempting to send a confirmation email...")

	dialer := gomail.NewDialer(client.SmtpServer, client.smtpPort, client.smtpUsername, client.smtpPassword)
	if e := dialer.DialAndSend(m); e != nil {
		log.Error().
			Err(e).
			Msg("Failed to send confirmation email")

		return e
	}

	log.Info().
		Str("requestID", requestID).
		Str("sender", client.sender.String()).
		Str("recipient", message.Recipient.String()).
		Msg("Email sent")

	return nil
}
