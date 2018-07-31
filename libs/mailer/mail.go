package mailer

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

var (
	MailConfig *Mailer
	mailQueue  chan *Message
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
	HookMode          bool
}

type Message struct {
	*gomail.Message
	Info        string
	confirmChan chan struct{}
}

func NewMessage(to []string, subject, htmlBody string) *Message {
	return NewMessageFrom(to, MailConfig.From, subject, htmlBody)
}

func NewMessageFrom(to []string, from, subject, htmlBody string) *Message {
	log.Printf("QueueLength (%d) NewMessage (htmlBody) \n%s\n", len(mailQueue), htmlBody)
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to...)
	msg.SetHeader("Subject", subject)
	msg.SetDateHeader("Date", time.Now())
	contentType := "text/html"

	msg.SetBody(contentType, htmlBody)
	return &Message{
		Message:     msg,
		confirmChan: make(chan struct{}),
	}
}

type Sender struct {
}

func (s *Sender) Send(from string, to []string, msg io.WriterTo) error {
	host, port, err := net.SplitHostPort(MailConfig.Host)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: MailConfig.SkipVerify,
		ServerName:         host,
	}
	if MailConfig.UseCertificate {
		cert, err := tls.LoadX509KeyPair(MailConfig.CertFile, MailConfig.KeyFile)
		if err != nil {
			return err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 3*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	isSecureConn := false
	if strings.HasSuffix(port, "465") {
		conn = tls.Client(conn, tlsConfig)
		isSecureConn = true
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("NewClient: %v", err)
	}

	if MailConfig.DisableHelo {
		hostname := MailConfig.HeloHostname
		if len(hostname) == 0 {
			hostname, err = os.Hostname()
			if err != nil {
				return err
			}
		}

		if err = client.Hello(hostname); err != nil {
			return fmt.Errorf("Hello:%v", err)
		}
	}

	hasStartTLS, _ := client.Extension("STARTTLS")
	if !isSecureConn && hasStartTLS {
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("StartTLS:%v", err)
		}
	}

	canAuth, options := client.Extension("AUTH")

	if canAuth && len(MailConfig.User) > 0 {
		var auth smtp.Auth
		if strings.Contains(options, "CRAM-MD5") {
			auth = smtp.CRAMMD5Auth(MailConfig.User, MailConfig.Passwd)
		} else if strings.Contains(options, "PLAIN") {
			auth = smtp.PlainAuth("", MailConfig.User, MailConfig.Passwd, host)
		} else if strings.Contains(options, "LOGIN") {
			// Patch for AUTH LOGIN
			auth = LoginAuth(MailConfig.User, MailConfig.Passwd)
		}
		if auth != nil {
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("Auth: %v", err)
			}
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("Mail: %v", err)
	}
	for _, rec := range to {
		if err = client.Rcpt(rec); err != nil {
			return fmt.Errorf("Rcpt: %v", err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("Data: %v", err)
	} else if _, err = msg.WriteTo(w); err != nil {
		return fmt.Errorf("WriteTo: %v", err)
	} else if err = w.Close(); err != nil {
		return fmt.Errorf("Close: %v", err)
	}
	return client.Quit()
}

func processMailQueue() {
	sender := &Sender{}
	for {
		select {
		case msg := <-mailQueue:
			log.Printf("New e-mail sending request %s: %s\n", msg.GetHeader("To"), msg.Info)
			if err := gomail.Send(sender, msg.Message); err != nil {
				log.Printf("Fail to send emails %s: %s - %v\n", msg.GetHeader("To"), msg.Info, err)
			} else {
				log.Printf("E-mails sent %s: %s\n", msg.GetHeader("To"), msg.Info)
			}
			msg.confirmChan <- struct{}{}
		}
	}
}

func InitMailer(m *Mailer) {
	MailConfig = m
	if MailConfig == nil || mailQueue != nil {
		return
	}

	mailQueue = make(chan *Message, MailConfig.QueueLength)
	go processMailQueue()
}

func Send(msg *Message) {
	mailQueue <- msg
	if MailConfig.HookMode {
		<-msg.confirmChan
		return
	}
	go func() {
		<-msg.confirmChan
	}()
}

func SendMail(to []string, subject, content string) error {
	if MailConfig == nil {
		return errors.New("update mail config must restart service")
	}
	msg := NewMessage(to, subject, content)
	Send(msg)
	return nil
}
