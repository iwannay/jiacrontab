package conf

import (
	"log"
	"net/mail"
)

var (
	MailService *Mailer
)

type Mailer struct {
	Enabled        bool   `json:"enabled"`
	QueueLength    int    `json:"queue_length"`
	SubjectPrefix  string `json:"subject_Prefix"`
	Host           string `json:"host"`
	From           string `json:"from"`
	FromEmail      string `json:"from_email"`
	User           string `json:"user"`
	Passwd         string `json:"passwd"`
	DisableHelo    bool   `json:"disable_helo"`
	HeloHostname   string `json:"helo_hostname"`
	SkipVerify     bool   `json:"skip_verify"`
	UseCertificate bool   `json:"use_certificate"`
	CertFile       string `json:"cert_file"`
	KeyFile        string `json:"key_file"`
	UsePlainText   bool   `json:"use_plain_text"`
}

func LoadMailService() {
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
