package mail

// For future use
// Adapted from https://hackernoon.com/golang-sendmail-sending-mail-through-net-smtp-package-5cadbe2670e0

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type Mailer interface {
	SendMail(*Mail) error
}

type Recipient struct {
	Name string
	Mail string
}

func (r Recipient) String() string {
	return fmt.Sprintf("%s <%s>", r.Name, r.Mail)
}

type Mail struct {
	Sender  string
	To      []Recipient
	Cc      []Recipient
	Bcc     []Recipient
	Subject string
	Body    string
}

type SmtpServer struct {
	Host      string
	Port      int
	Password  string
	User      string
	TlsConfig *tls.Config
}

type TestMailer struct {
	Mails []*Mail
}

func (m *TestMailer) SendMail(mail *Mail) error {
	m.Mails = append(m.Mails, mail)
	return nil
}

func (mail *Mail) BuildMessage() string {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "From: %s\r\n", mail.Sender)

	to_strings := make([]string, len(mail.To))
	for i, recipient := range mail.To {
		to_strings[i] = recipient.String()
	}
	cc_strings := make([]string, len(mail.To))
	for i, recipient := range mail.To {
		cc_strings[i] = recipient.String()
	}
	if len(mail.To) > 0 {
		fmt.Fprintf(sb, "To: %s\r\n", strings.Join(to_strings, ";"))
	}
	if len(mail.Cc) > 0 {
		fmt.Fprintf(sb, "Cc: %s\r\n", strings.Join(cc_strings, ";"))
	}

	fmt.Fprintf(sb, "Subject: %s\r\n", mail.Subject)
	fmt.Fprint(sb, "\r\n")
	fmt.Fprint(sb, mail.Body)
	fmt.Fprint(sb, "\r\n.\r\n")

	return sb.String()
}

func ConfirmationEmail(name, email, confirmationURL string) *Mail {
	return &Mail{
		Sender:  "noreply@isamuni.org",
		To:      []Recipient{Recipient{name, email}},
		Subject: "Confirm email",
		Body:    fmt.Sprintf("Hello %s, you're receiving this email because someone (hopefully you) registered it on isamuni.org\r\nIf it was you, please use on %s to confirm your mail.\r\nOtherwise simply ignore this mail.", name, confirmationURL),
	}
}

func (s *SmtpServer) SendMail(mail *Mail) error {

	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port), s.TlsConfig)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return err
	}

	// step 1: Use Auth
	if err = client.Auth(auth); err != nil {
		return err
	}

	// step 2: add all from and to
	if err = client.Mail(mail.Sender); err != nil {
		return err
	}

	receivers := append(mail.To, mail.Cc...)
	receivers = append(receivers, mail.Bcc...)
	for _, k := range receivers {
		log.Println("sending to: ", k)
		if err = client.Rcpt(k.Mail); err != nil {
			return err
		}
	}

	// Data
	w, err := client.Data()
	if err != nil {
		return err
	}

	mailbody := mail.BuildMessage()
	_, err = w.Write([]byte(mailbody))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	err = client.Quit()
	if err != nil {
		return err
	}

	return nil
}

/*
mail := Mail{}
mail.Sender = "abc@gmail.com"
mail.To = []string{"def@yahoo.com", "xyz@outlook.com"}
mail.Cc = []string{"mnp@gmail.com"}
mail.Bcc = []string{"a69@outlook.com"}
mail.Subject = "I am Harry Potter!!"
mail.Body = "Harry Potter and threat to Israel\n\nGood editing!!"

messageBody := mail.BuildMessage()

smtpServer := SmtpServer{Host: "smtp.gmail.com", Port: "465", Password: "password"}
smtpServer.TlsConfig = &tls.Config{
	InsecureSkipVerify: true,
	ServerName:         smtpServer.Host,
}
err := smtpServer.sendMail(mail)
*/
