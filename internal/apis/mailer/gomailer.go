package mailer

import (
	"bytes"
	"fmt"
	"github.com/Brain-Wave-Ecosystem/mail-service/internal/models"
	"text/template"

	"github.com/NawafSwe/gomailer"
	"github.com/NawafSwe/gomailer/message"
)

const (
	confirmEmailSubject        = "BW - Verification Code"
	successConfirmEmailSubject = "BW - Account Verified Successfully"
)

var templ *template.Template

type GoMailer struct {
	sender string
	mailer gomailer.SendCloser
}

func NewGoMailer(cfg *Config) (*GoMailer, error) {
	var g GoMailer
	var err error

	mailer := gomailer.NewMailer(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Options...)

	g.mailer, err = mailer.ConnectAndAuthenticate()
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("").ParseFiles("./app/templates/confirm_email.html", "./app/templates/success_confirm_email.html")
	if err != nil {
		return nil, fmt.Errorf("parse html tepmplate error: %v", err)
	}

	templ = tmpl

	g.sender = cfg.Sender

	return &g, nil
}

func (g *GoMailer) Send(msg *message.Message) error {
	return g.mailer.Send(*msg)
}

func (g *GoMailer) SendConfirmMessage(data *models.ConfirmEmailMail) error {
	var body bytes.Buffer
	if err := templ.ExecuteTemplate(&body, "confirm_email.html", data); err != nil {
		return fmt.Errorf("templ execute error: %v", err)
	}

	msg := message.NewMessage()
	msg.From = g.sender
	msg.Recipients = []string{data.Email}
	msg.Subject = confirmEmailSubject
	msg.HTMLBody = body.String()

	return g.mailer.Send(msg)
}

func (g *GoMailer) SendSuccessConfirmMessage(data *models.SuccessConfirmEmailMail) error {
	var body bytes.Buffer
	if err := templ.ExecuteTemplate(&body, "success_confirm_email.html", data); err != nil {
		return fmt.Errorf("templ execute error: %v", err)
	}

	msg := message.NewMessage()
	msg.From = g.sender
	msg.Recipients = []string{data.Email}
	msg.Subject = successConfirmEmailSubject
	msg.HTMLBody = body.String()

	return g.mailer.Send(msg)
}

func (g *GoMailer) Close() error {
	return g.mailer.Close()
}
