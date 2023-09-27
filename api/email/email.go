package email

import (
	"github.com/gin-gonic/gin"
	"github.com/go-gomail/gomail"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/configs"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

type EmailClient struct {
	smtpServer   string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	sender       models.SubscriberEmail
}

type Email struct {
	Recipient models.SubscriberEmail
	Subject   string
	Html      string
	Text      string
}

func NewEmailClient() (*EmailClient, error) {
	cfg, e := configs.ConfigureEmailClient()
	if e != nil {
		return nil, e
	}

	// validate cfg sender email
	s := cfg.Sender
	sender, e := models.ParseEmail(s)
	if e != nil {
		return nil, e
	}

	client := &EmailClient{
		cfg.Server,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		sender,
	}

	return client, nil
}

func (client *EmailClient) SendEmail(c *gin.Context, email Email) error {
	requestID := c.GetString("requestID")

	message := gomail.NewMessage()
	message.SetHeader("From", client.sender.String())
	message.SetHeader("To", email.Recipient.String())
	message.SetHeader("Subject", email.Subject)
	message.SetBody("text/plain", email.Text)
	message.AddAlternative("text/html", email.Html)

	log.Info().
		Str("requestID", requestID).
		Str("sender", client.sender.String()).
		Str("recipient", email.Recipient.String()).
		Msg("Attempting to send an email")

	dialer := gomail.NewDialer(client.smtpServer, client.smtpPort, client.smtpUsername, client.smtpPassword)
	if e := dialer.DialAndSend(message); e != nil {
		log.Error().
			Err(e).
			Msg("Failed to send email")

		return e
	}

	log.Info().
		Str("requestID", requestID).
		Str("sender", client.sender.String()).
		Str("recipient", email.Recipient.String()).
		Msg("Email sent")

	return nil
}
