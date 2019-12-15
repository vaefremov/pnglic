package chkexprd

import (
	"fmt"
	"log"
	"time"

	"github.com/vaefremov/pnglic/api"
	"github.com/vaefremov/pnglic/pkg/mailnotify"
	"github.com/vaefremov/pnglic/server"
)

func RunExpiryNotifications(conf *server.Config) {
	db := api.MustNewPool(conf.DSN)
	notifyer := mailnotify.New(conf.MailServer, conf.MailPort, conf.MailUser, conf.MailPass).AddTo(conf.AdminMail).AddTo(conf.BackMail)
	ticker := time.NewTicker(24 * time.Hour)
	expTerm1 := time.Duration(time.Duration(24) * time.Duration(conf.DaysToExpire1) * time.Hour)

	for {
		features, err := FindFeaturesWillExpire(db, expTerm1)
		if err != nil {
			panic(err)
		}
		if len(features) > 0 {
			err = ReportFeaturesWillExpire(features, expTerm1, notifyer, conf)
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

func ReportFeaturesWillExpire(features map[string]ExpFeaturesReportElt, expTerm time.Duration, nt mailnotify.MailNotifyer, conf *server.Config) error {
	log.Println("Reporting features that will expire within ", expTerm.String())
	message, _ := MakeMessageFromTemplate(features, expTerm, conf)
	err := nt.SendMessage("Warning: some licenses will expire in "+fmt.Sprint(expTerm.Hours()/24)+" days", message)
	return err
}
