package email

type EmailClient struct {
	Sender SenderEmail
}

type SenderEmail string
