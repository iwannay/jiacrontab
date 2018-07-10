package conf

import (
	"log"
	"net/mail"
)

var (
	MailService *Mailer
)

type Mailer struct {
	QueueLength       int    `json:"queue_length"`
	SubjectPrefix     string `json:"Subject_Prefix"`
	Host              string `json:"host"`
	From              string `json:"from"`
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
	if !sec.Key("enabled").MustBool() {
		return
	}

	MailService = &Mailer{
		QueueLength:    sec.Key("send_buffer_len").MustInt(100),
		SubjectPrefix:  sec.Key("subject_prefix").MustString("[" + "jiacrontab" + "] "),
		Host:           sec.Key("host").String(),
		User:           sec.Key("user").String(),
		Passwd:         sec.Key("passwd").String(),
		DisableHelo:    sec.Key("disable_helo").MustBool(),
		HeloHostname:   sec.Key("helo_hostname").String(),
		SkipVerify:     sec.Key("skip_verify").MustBool(),
		UseCertificate: sec.Key("use_certificate").MustBool(),
		CertFile:       sec.Key("cert_file").String(),
		KeyFile:        sec.Key("key_file").String(),
		UsePlainText:   sec.Key("use_plain_text").MustBool(),
	}

	MailService.From = sec.Key("from").MustString(MailService.User)

	if len(MailService.From) > 0 {
		parsed, err := mail.ParseAddress(MailService.From)
		if err != nil {
			log.Fatal("Invalid mailer.FROM (%s): %v", MailService.From, err)
		}
		MailService.FromEmail = parsed.Address
	}
}
