package chkexprd

import (
	"fmt"
	"github.com/vaefremov/pnglic/server"
	"strings"
	"text/template"
	"time"
)

type templData struct {
	ExpTerm         time.Duration
	ExpTermDays     int
	ServerPort      int
	ServerPublicURL string
	Keys            []templDataElt
}

type templDataElt struct {
	KeyID       string
	ExpTermDays int
	ExpTimeStr  string
	ExpFeaturesReportElt
}

func MakeSimpleMessage(features map[string]ExpFeaturesReportElt, expTerm time.Duration, conf *server.Config) (res string, err error) {
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
	return bld.String(), nil
}

func MakeMessageFromTemplate(features map[string]ExpFeaturesReportElt, expTerm time.Duration, conf *server.Config) (res string, err error) {
	bld := strings.Builder{}
	data := templData{ExpTerm: expTerm, ExpTermDays: int(expTerm.Hours() / 24),
		ServerPort: conf.Port, ServerPublicURL: fmt.Sprintf("http://%s:%d", conf.PublicName, conf.Port)}
	for k, v := range features {
		tmp := templDataElt{KeyID: k, ExpFeaturesReportElt: v, ExpTermDays: int(v.ExpTerm.Hours() / 24),
			ExpTimeStr: v.ExpTime.Format("2006-01-02")}
		data.Keys = append(data.Keys, tmp)
	}
	t := template.New("message")
	t, err = t.Parse(messageTemplate)
	if err != nil {
		return "", err
	}
	err = t.Execute(&bld, data)
	return bld.String(), err
}

const messageTemplate = `
Please, check the following keys for features that will expire soon (in {{.ExpTermDays}} day):

{{ range $index, $key := .Keys }}
{{$index}} : {{$key.KeyID}} ({{$.ServerPublicURL}}/v1/view/keyfeatures.html?keyId={{$key.KeyID}}&fullPage=true)
		Client: {{$key.ClientName}}
		{{range $feature := $key.Features}}{{$feature}} {{end}} will expire in {{$key.ExpTermDays}} day(s)
		Expiration date: {{$key.ExpTimeStr}}
{{ end }}

Hope that helps. Thanks!
`
