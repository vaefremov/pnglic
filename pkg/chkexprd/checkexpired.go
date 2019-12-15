package chkexprd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vaefremov/pnglic/api"
	"github.com/vaefremov/pnglic/pkg/mailnotify"
	"github.com/vaefremov/pnglic/server"
)

func RunExpiryNotifications(conf *server.Config) {
	db := api.MustNewPool(conf.DSN)
	notifyer := mailnotify.New(conf.MailServer, conf.MailPort, conf.MailUser, conf.MailPass).AddTo(conf.AdminMail)
	ticker := time.NewTicker(24 * time.Hour)
	expTerm1 := time.Duration(time.Duration(24) * time.Duration(conf.DaysToExpire1) * time.Hour)

	for {
		features, err := FindFeaturesWillExpire(db, expTerm1)
		if err != nil {
			panic(err)
		}
		if len(features) > 0 {
			err = ReportFeaturesWillExpire(features, expTerm1, notifyer)
			if err != nil {
				panic(err)
			}
		}
		<-ticker.C
	}
}

type ExpFeaturesReportElt struct {
	ClientName string
	ExpTime    time.Time
	ExpTerm    time.Duration
	Features   []string
}

func FindFeaturesWillExpire(db *api.DbConn, expTerm time.Duration) (map[string]ExpFeaturesReportElt, error) {
	log.Println("Checking if there are features going to expire in ", expTerm)
	res := map[string]ExpFeaturesReportElt{}
	features, err := db.WillEndSoon(expTerm)
	if err != nil {
		return res, err
	}
	for _, f := range features {
		if r, ok := res[f.KeyID]; !ok {
			clFull, err := db.KeyOfWhichOrg(f.KeyID)
			if err != nil {
				return nil, err
			}
			clName := clFull.Name

			res[f.KeyID] = ExpFeaturesReportElt{ClientName: clName, ExpTime: f.ExpTime, ExpTerm: f.ExpTerm, Features: []string{f.Feature}}
		} else {
			// res[f.KeyID].Features = append(res[f.KeyID].Features, f.Feature)
			r.Features = append(r.Features, f.Feature)
			res[f.KeyID] = r
		}
	}
	return res, nil
}

func ReportFeaturesWillExpire(features map[string]ExpFeaturesReportElt, expTerm time.Duration, nt mailnotify.MailNotifyer) error {
	log.Println("Reporting features that will expire within ", expTerm.String())
	bld := strings.Builder{}
	bld.WriteString("Please, check the following keys for features that will expire soon:\r\n")
	for k, v := range features {
		bld.WriteString("Key ID: " + k + "\r\n")
		bld.WriteString("  " + fmt.Sprint(v.ClientName) + "\r\n")
		// bld.WriteString("  " + fmt.Sprint(v.ExpTime.String()) + "\r\n")
		bld.WriteString("  Features: " + fmt.Sprint(v.Features) + "\r\n")
		expDays := v.ExpTerm.Hours() / 24
		bld.WriteString("  Will expire within " + fmt.Sprint(expDays) + " days\r\n")
		bld.WriteString("\r\n")
	}
	bld.WriteString("\r\nHope that helps. Thanks!")
	err := nt.SendMessage("Warning: some features will expire in "+fmt.Sprint(expTerm.Hours()/24)+" days", bld.String())
	return err
}
