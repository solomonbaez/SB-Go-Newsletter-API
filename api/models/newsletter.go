package models

type Newsletter struct {
	Title   string
	Content content
}

type content struct {
	text string
	html string
}
