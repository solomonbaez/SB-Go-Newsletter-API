package clients

import (
	"fmt"

	"github.com/go-gomail/gomail"
	"github.com/solomonbaez/hyacinth/api/configs"
	"github.com/solomonbaez/hyacinth/api/models"
)

type EmailClient interface {
	SendEmail(email *models.Newsletter) error
}

type SMTPClient struct {
	SmtpServer   string // export for testing
	SmtpPort     int
	smtpUsername string
	smtpPassword string
	Sender       *models.SubscriberEmail
}

func NewSMTPClient(cfgFile *string) (client *SMTPClient, err error) {
	var file string
	if *cfgFile != "" {
		if *cfgFile != "test" {
			file = fmt.Sprintf("./api/configs/%v.yaml", *cfgFile)
		} else {
			file = "../api/configs/dev.yaml"
		}
	}

	cfg, e := configs.ConfigureEmailClient(file)
	if e != nil {
		err = fmt.Errorf("failed to configure email client: %w", e)
		return
	}

	sender, e := models.ParseEmail(cfg.Sender)
	if e != nil {
		err = fmt.Errorf("failed to parse email: %w", e)
		return
	}

	client = &SMTPClient{
		SmtpServer:   cfg.Server,
		SmtpPort:     cfg.Port,
		smtpUsername: cfg.Username,
		smtpPassword: cfg.Password,
		Sender:       &sender,
	}

	return
}

func (client *SMTPClient) SendEmail(newsletter *models.Newsletter) (err error) {
	if e := models.ParseNewsletter(newsletter); e != nil {
		err = fmt.Errorf("invalid newsletter: %w", e)
		return
	} else if e = models.ParseNewsletter(newsletter.Content); e != nil {
		err = fmt.Errorf("invalid newsletter content: %w", e)
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", client.Sender.String())
	m.SetHeader("To", newsletter.Recipient.String())
	m.SetHeader("Subject", newsletter.Content.Title)
	m.SetBody("text/plain", newsletter.Content.Text)
	m.AddAlternative("text/html", newsletter.Content.Html)

	dialer := gomail.NewDialer(client.SmtpServer, client.SmtpPort, client.smtpUsername, client.smtpPassword)
	if e := dialer.DialAndSend(m); e != nil {
		err = fmt.Errorf("failed to send email: %w", e)
		return
	}

	return
}
