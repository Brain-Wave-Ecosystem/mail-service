package mailer

import (
	"github.com/NawafSwe/gomailer"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
	Options  []gomailer.Options
}
