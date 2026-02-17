package email

type Message struct {
	To       []string
	CC       []string
	BCC      []string
	Subject  string
	TextBody string
	HTMLBody string
	Headers  map[string]string
}
