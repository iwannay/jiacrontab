package conf

import (
	"log"
	"net/mail"
)

var (
	MailService *Mailer
)

type Mailer struct {
	QueueLength       int
	SubjectPrefix     string
	Host              string
	From              string
	FromEmail         string
	User, Passwd      string
	DisableHelo       bool
	HeloHostname      string
	SkipVerify        bool
	UseCertificate    bool
	CertFile, KeyFile string
	UsePlainText      bool
}

func newMailService() {
	sec := cf.Section("mailer")
	if !sec.Key("ENABLED").MustBool() {
		return
	}

	MailService = &Mailer{
		QueueLength:    sec.Key("SEND_BUFFER_LEN").MustInt(100),
		SubjectPrefix:  sec.Key("SUBJECT_PREFIX").MustString("[" + "jiacrontab" + "] "),
		Host:           sec.Key("HOST").String(),
		User:           sec.Key("USER").String(),
		Passwd:         sec.Key("PASSWD").String(),
		DisableHelo:    sec.Key("DISABLE_HELO").MustBool(),
		HeloHostname:   sec.Key("HELO_HOSTNAME").String(),
		SkipVerify:     sec.Key("SKIP_VERIFY").MustBool(),
		UseCertificate: sec.Key("USE_CERTIFICATE").MustBool(),
		CertFile:       sec.Key("CERT_FILE").String(),
		KeyFile:        sec.Key("KEY_FILE").String(),
		UsePlainText:   sec.Key("USE_PLAIN_TEXT").MustBool(),
	}

	MailService.From = sec.Key("FROM").MustString(MailService.User)

	if len(MailService.From) > 0 {
		parsed, err := mail.ParseAddress(MailService.From)
		if err != nil {
			log.Fatal("Invalid mailer.FROM (%s): %v", MailService.From, err)
		}
		MailService.FromEmail = parsed.Address
	}
}
