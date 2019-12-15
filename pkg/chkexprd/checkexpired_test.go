package chkexprd_test

import (
	"fmt"
	"net/smtp"
	"testing"
	"time"

	"github.com/vaefremov/pnglic/api"
	"github.com/vaefremov/pnglic/pkg/chkexprd"
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
	chkexprd.ReportFeaturesWillExpire(tmpReport, expTerm, m)
	t.Error("Printing...")
	fmt.Println(string(n.body))
}

func TestFindFeaturesWillExpire(t *testing.T) {
	db := api.MustInMemoryTestPool()
	tillDate1 := time.Now().AddDate(0, 0, 1)
	tillDate2 := time.Now().AddDate(0, 0, 2)
	keyID := "123abc"
	newLicset := []api.LicenseSetItem{}
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
	t.Error(res)
}
