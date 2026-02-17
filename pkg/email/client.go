package email

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	"github.com/Alijeyrad/simorq_backend/config"
	"gopkg.in/gomail.v2"
)

type Client struct {
	cfg Config
	d   *gomail.Dialer
}

// NewFromCentral creates a new email client from central config
func NewFromCentral(cfg config.EmailConfig) (*Client, error) {
	return New(FromCentralConfig(cfg))
}

func New(cfg Config) (*Client, error) {
	d := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
	return &Client{cfg: cfg, d: d}, nil
}

func (c *Client) Send(ctx context.Context, m Message) error {
	if !c.cfg.Enabled {
		return ErrDisabled{}
	}

	msg, err := buildMessage(c.cfg.From, m)
	if err != nil {
		return err
	}

	d := c.newDialer()

	done := make(chan error, 1)
	go func() {
		done <- d.DialAndSend(msg)
	}()

	// Respect ctx deadline if it's sooner than our config timeout.
	wait := c.cfg.SMTPTimeout()
	if dl, ok := ctx.Deadline(); ok {
		if d := time.Until(dl); d > 0 && d < wait {
			wait = d
		}
	}

	select {
	case err := <-done:
		if err != nil {
			return ErrSend{Provider: "gomail/smtp", Err: err}
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(wait):
		return context.DeadlineExceeded
	}
}

func (c *Client) newDialer() *gomail.Dialer {
	d := gomail.NewDialer(c.cfg.SMTPHost, c.cfg.SMTPPort, c.cfg.SMTPUsername, c.cfg.SMTPPassword)

	d.SSL = c.cfg.SMTPUseTLS

	if c.cfg.SMTPUseTLS {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return d
}

func buildMessage(from string, m Message) (*gomail.Message, error) {
	msg := gomail.NewMessage()

	// From
	from = strings.TrimSpace(from)
	if from == "" {
		return nil, ErrInvalidMessage{Reason: "from is required"}
	}
	msg.SetHeader("From", from)

	// Recipients
	if len(m.To) > 0 {
		msg.SetHeader("To", cleanAddrs(m.To)...)
	}
	if len(m.CC) > 0 {
		msg.SetHeader("Cc", cleanAddrs(m.CC)...)
	}
	if len(m.BCC) > 0 {
		msg.SetHeader("Bcc", cleanAddrs(m.BCC)...)
	}

	// Subject
	subj := strings.TrimSpace(m.Subject)
	if subj == "" {
		return nil, ErrInvalidMessage{Reason: "subject is required"}
	}
	msg.SetHeader("Subject", subj)

	// Extra headers
	for k, v := range m.Headers {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		msg.SetHeader(k, v)
	}

	// Body
	hasText := strings.TrimSpace(m.TextBody) != ""
	hasHTML := strings.TrimSpace(m.HTMLBody) != ""

	switch {
	case hasText && hasHTML:
		msg.SetBody("text/plain", m.TextBody)
		msg.AddAlternative("text/html", m.HTMLBody)
	case hasHTML:
		msg.SetBody("text/html", m.HTMLBody)
	case hasText:
		msg.SetBody("text/plain", m.TextBody)
	default:
		return nil, ErrInvalidMessage{Reason: "either TextBody or HTMLBody is required"}
	}

	return msg, nil
}

func cleanAddrs(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}
