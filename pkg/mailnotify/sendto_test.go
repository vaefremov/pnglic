package mailnotify_test

import (
	"bytes"
	"fmt"
	"io"
	"net/mail"
	"net/smtp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vaefremov/pnglic/pkg/mailnotify"
)

type mockNotifyer struct {
	addr string
	auth string
	from string
	to   string
	body []byte
}

func (n *mockNotifyer) Send(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	n.addr = fmt.Sprint(addr)
	n.auth = fmt.Sprint(a)
	n.from = fmt.Sprint(from)
	n.to = fmt.Sprint(to)
	n.body = msg
	return nil
}

func TestNotify(t *testing.T) {
	n := mockNotifyer{}
	// m := mailnotify.MailServiceImpl{Send: n.Send}
	mi := mailnotify.New("mail.server", 25, "user@pangea.ru", "**pass**").AddTo("some.addressee").AddTo("some.addressee2")
	m := mi.(*mailnotify.MailServiceImpl)
	m.Send = n.Send
	err := m.SendFile("VIP Client", "1234abc", []byte("test body"))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "mail.server:25", n.addr)
	assert.Equal(t, "user@pangea.ru", n.from)
	assert.Equal(t, "[some.addressee some.addressee2]", n.to)
}

func TestSendMessage(t *testing.T) {
	n := mockNotifyer{}
	// m := mailnotify.MailServiceImpl{Send: n.Send}
	mi := mailnotify.New("mail.server", 25, "user@pangea.ru", "**pass**").AddTo("some.addressee").AddTo("some.addressee2")
	m := mi.(*mailnotify.MailServiceImpl)
	m.Send = n.Send
	expSubj := "My subject"
	expBody := "Body of the message"
	err := m.SendMessage(expSubj, expBody)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "mail.server:25", n.addr)
	assert.Equal(t, "user@pangea.ru", n.from)
	assert.Equal(t, "[some.addressee some.addressee2]", n.to)
	r := bytes.NewReader(n.body)
	mess, err := mail.ReadMessage(r)
	bld := strings.Builder{}
	io.Copy(&bld, mess.Body)
	res := bld.String()
	assert.Equal(t, expBody+"\r\n", res)
	// t.Error(mess.Header.Get("Subject"))
}
