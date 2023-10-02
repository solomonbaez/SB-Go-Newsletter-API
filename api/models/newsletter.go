package models

type NewsletterBody struct {
	Title   string
	Content content
}

type content struct {
	text string
	html string
}
