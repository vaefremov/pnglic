package chkexprd_test

import (
	"fmt"
	"net/smtp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vaefremov/pnglic/config"
	"github.com/vaefremov/pnglic/pkg/chkexprd"
	"github.com/vaefremov/pnglic/pkg/dao"
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

func TestReportWillExpireFeatures(t *testing.T) {
	expTerm := time.Hour * 24
	expTime := time.Now()
	tmpReport := map[string]chkexprd.ExpFeaturesReportElt{
		"key1234": chkexprd.ExpFeaturesReportElt{ClientName: "Org 1", ExpTime: expTime, ExpTerm: expTerm, Features: []string{"F1", "F2"}},
		"key1235": chkexprd.ExpFeaturesReportElt{ClientName: "Org 1", ExpTime: expTime, ExpTerm: expTerm, Features: []string{"P1", "P2"}}}
	notifyer := mailnotify.New("mail.server", 25, "user@pangea.ru", "**pass**").AddTo("some.addressee")
	n := mockNotifyer{}
	m := notifyer.(*mailnotify.MailServiceImpl)
	m.Send = n.Send
	conf := config.Config{Port: 9995, PublicName: "some.host"}
	chkexprd.ReportFeaturesWillExpire(tmpReport, expTerm, m, &conf)
	// t.Error("Printing...")
	assert.Equal(t, 704, len(n.body))
	fmt.Println(string(n.body))
}

func TestFindFeaturesWillExpire(t *testing.T) {
	db := dao.MustInMemoryTestPool()
	tillDate1 := time.Now().AddDate(0, 0, 1)
	tillDate2 := time.Now().AddDate(0, 0, 2)
	keyID := "123abc"
	newLicset := []dao.LicenseSetItem{}
	currentLicSet, err := db.LicensesSetByKeyId(keyID)
	if err != nil {
		t.Error(err)
	}
	for i, f := range currentLicSet {
		switch i {
		case 0:
			f.End = tillDate1
		default:
			f.End = tillDate2
		}
		newLicset = append(newLicset, f)
	}
	db.UpdateLicenseSet(keyID, newLicset)
	res, err := chkexprd.FindFeaturesWillExpire(db, 2*time.Hour*24)
	assert.Equal(t, 1, len(res))
	assert.True(t, (24*time.Hour > tillDate1.Sub(res[keyID].ExpTime)) && (tillDate1.Sub(res[keyID].ExpTime) > 0))
	// t.Error(res)
}
