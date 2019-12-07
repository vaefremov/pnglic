package mailnotify

import (
	"fmt"
	"net/mail"
	"net/smtp"

	"github.com/scorredoira/email"
)

type MailNotifyer interface {
	SendFile(fileName string, fileBody []byte) error
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

func New(serv string, port int) MailNotifyer {
	a := smtp.PlainAuth("", "git@pangea.ru", "3X6git", serv)
	newM := MailServiceImpl{
		Serv: serv,
		A:    a, From: "Git@pangea.ru", MailServPort: serv + fmt.Sprintf(":%d", port),
		To:   []string{},
		Send: smtp.SendMail,
	}
	return &newM
}

func (m MailServiceImpl) SendFile(fileName string, fileBody []byte) error {
	subj := "License file: " + fileName
	msg := email.NewMessage(subj, "this is the body")
	msg.From = mail.Address{Name: "Pangea License Generator", Address: m.From}
	msg.To = m.To
	msg.AttachBuffer("license.xml", fileBody, false)
	s := &unencryptedAuth{m.A}
	err := m.Send(m.MailServPort, s, m.From, m.To, msg.Bytes())
	return err
}

func (m *MailServiceImpl) AddTo(addr string) MailNotifyer {
	m.To = append(m.To, addr)
	return m
}
