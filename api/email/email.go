package email

import (
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

func (client EmailClient) SendEmail(email Email) Email {
	log.Info().
		Str("sender", client.Sender.String()).
		Msg("Attempting to send an email")

	return email
}
