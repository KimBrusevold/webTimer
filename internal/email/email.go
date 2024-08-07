package email

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
)

type EmailMessage struct {
	to          []string
	from        string
	subject     string
	body        string
	contentType string
}

type EmailClient struct {
	HostAddr   string
	SenderAddr string
	Password   string
}

func (c EmailClient) SendEmail(e *EmailMessage) error {
	auth := smtp.PlainAuth("", c.SenderAddr, c.Password, c.HostAddr)

	body := e.BuildBody()
	err := smtp.SendMail(c.HostAddr+":587", auth, c.SenderAddr, e.to, body)

	return err
}

func (e *EmailMessage) BuildBody() []byte {
	from := fmt.Sprintf("From: %s\r\n", e.from)
	to := "To: "

	for i := 0; i < len(e.to); i++ {
		to = to + e.to[i]
		if i != len(e.to)-1 {
			to = to + ", "
		}
	}

	to = to + "\r\n"
	subjectBytes := []byte(e.subject)
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(subjectBytes)))
	base64.StdEncoding.Encode(dst, subjectBytes)
	log.Print(string(dst))
	subject := "Subject: =?utf-8?B?" + string(dst) + "?=" + "\r\n"
	contentType := "Content-Type: " + e.contentType + " charset=utf-8" + "\r\n"
	return []byte(from + to + subject + contentType + "\r\n" + e.body)
}

func NewEmailMessage(from string) *EmailMessage {
	return &EmailMessage{
		from: from,
	}
}

func (e *EmailMessage) AddRecipients(eAddr ...string) *EmailMessage {
	e.to = append(e.to, eAddr...) // Use append here, so that it can be called multiple times
	return e
}

func (e *EmailMessage) SetSubject(subject string) *EmailMessage {
	e.subject = subject
	return e
}

func (e *EmailMessage) AddStringContent(body string) *EmailMessage {
	e.body = body
	e.contentType = "text/plain"
	return e
}
