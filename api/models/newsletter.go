package models

type Newsletter struct {
	Recipient SubscriberEmail
	Content   *Body
}

type Body struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Html  string `json:"html"`
}
