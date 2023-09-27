package email

import (
	"github.com/gin-gonic/gin"
	"github.com/go-gomail/gomail"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

type EmailClient struct {
	SMTPServer   string
	SMTPPort     int
	SMTPUsername string
	SMTPassword  string
	Sender       models.SubscriberEmail
}

type Email struct {
	Recipient models.SubscriberEmail
	Subject   string
	Html      string
	Text      string
}

func NewEmailClient(
	server string,
	port int,
	username string,
	password string,
	sender models.SubscriberEmail,
) *EmailClient {
	return &EmailClient{
		SMTPServer:   server,
		SMTPPort:     port,
		SMTPUsername: username,
		SMTPassword:  password,
		Sender:       sender,
	}
}

func (client EmailClient) SendEmail(c *gin.Context, email Email) error {
	requestID := c.GetString("requestID")

	message := gomail.NewMessage()
	message.SetHeader("From", client.Sender.String())
	message.SetHeader("To", email.Recipient.String())
	message.SetHeader("Subject", email.Subject)
	message.SetBody("text/plain", email.Text)
	message.AddAlternative("text/html", email.Html)

	log.Info().
		Str("requestID", requestID).
		Str("sender", client.Sender.String()).
		Str("recipient", email.Recipient.String()).
		Msg("Attempting to send an email")

	dialer := gomail.NewDialer(client.SMTPServer, client.SMTPPort, client.SMTPUsername, client.SMTPassword)
	if e := dialer.DialAndSend(message); e != nil {
		log.Error().
			Err(e).
			Msg("Failed to send email")

		return e
	}

	log.Info().
		Str("requestID", requestID).
		Str("sender", client.Sender.String()).
		Str("recipient", email.Recipient.String()).
		Msg("Email sent")

	return nil
}
