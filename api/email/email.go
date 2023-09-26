package email

import (
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

type EmailClient struct {
	Sender models.SubscriberEmail
}

type Email struct {
	Recipient models.SubscriberEmail
	Subject   string
	Html      string
	Text      string
}

func (client EmailClient) SendEmail(email Email) Email {
	return email
}
