package mailnotify

import (
	"fmt"
	"net/mail"
	"net/smtp"

	"github.com/scorredoira/email"
	"github.com/vaefremov/cyr2volapiuk"
)

type MailNotifyer interface {
	SendFile(clientName string, keyId string, fileBody []byte) error
	AddTo(addr string) MailNotifyer
}

type MailServiceImpl struct {
	From         string
	To           []string
	Serv         string
	MailServPort string
	A            smtp.Auth
	Send         func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

type unencryptedAuth struct {
	smtp.Auth
}

func (a *unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	server.TLS = true
	return a.Auth.Start(server)
}

func (a *unencryptedAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	return a.Auth.Next(fromServer, more)
}

// New constructs a new notifier to send license files ny e-mail.
func New(serv string, port int, username string, password string) MailNotifyer {
	a := smtp.PlainAuth("", username, password, serv)
	newM := MailServiceImpl{
		Serv: serv,
		A:    a, From: username, MailServPort: serv + fmt.Sprintf(":%d", port),
		To:   []string{},
		Send: smtp.SendMail,
	}
	return &newM
}

// SendFile sends the license file comprized of fileBoby. A short text and a subject
// is added to the letter, the subect is constructed using client's name and key ID.
func (m MailServiceImpl) SendFile(clientName string, keyID string, fileBody []byte) error {
	subj := "License file key " + keyID + " for " + clientName
	msg := email.NewMessage(subj, "Pls find the license file in the attachment.")
	msg.From = mail.Address{Name: "Pangea License Generator", Address: m.From}
	msg.To = m.To
	// msg.AddCc(mail.Address{Name: "Vladimir A. Efremov", Address: "budwe1ser@yandex.ru"})
	fileName := MakeLicenseFileName(clientName, keyID)
	msg.AttachBuffer(fileName, fileBody, false)
	s := &unencryptedAuth{m.A}
	err := m.Send(m.MailServPort, s, m.From, m.To, msg.Bytes())
	return err
}

// AddTo adds address to the list of recipients
func (m *MailServiceImpl) AddTo(addr string) MailNotifyer {
	m.To = append(m.To, addr)
	return m
}

// MakeLicenseFileName makes a valid license file name basing on
// the client name and the key ID
func MakeLicenseFileName(clientName string, keyID string) string {
	nm := cyr2volapiuk.FileName(clientName)
	return fmt.Sprintf("license_%s_%s.xml", keyID, nm)
}
