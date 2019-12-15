package chkexprd

import (
	"fmt"
	"strings"
	"text/template"
	"time"
)

type templData struct {
	ExpTerm     time.Duration
	ExpTermDays int
	Keys        []templDataElt
}

type templDataElt struct {
	KeyID       string
	ExpTermDays int
	ExpTimeStr  string
	ExpFeaturesReportElt
}

func MakeSimpleMessage(features map[string]ExpFeaturesReportElt, expTerm time.Duration) (res string, err error) {
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

func MakeMessageFromTemplate(features map[string]ExpFeaturesReportElt, expTerm time.Duration) (res string, err error) {
	bld := strings.Builder{}
	data := templData{ExpTerm: expTerm, ExpTermDays: int(expTerm.Hours() / 24)}
	for k, v := range features {
		tmp := templDataElt{KeyID: k, ExpFeaturesReportElt: v, ExpTermDays: int(v.ExpTerm.Hours() / 24),
			ExpTimeStr: v.ExpTime.Format("1006-01-02")}
		data.Keys = append(data.Keys, tmp)
	}
	t := template.New("message")
	t, err = t.Parse(messageTemplate)
	if err != nil {
		return "", err
	}
	err = t.Execute(&bld, data)
	return bld.String(), nil
}

const messageTemplate = `
Please, check the following keys for features that will expire soon (in {{.ExpTermDays}} day):

{{ range $index, $key := .Keys }}
{{$index}} : http://localhost:9995/v1/view/keyfeatures.html?keyId={{$key.KeyID}}&fullPage=true
		Client: {{$key.ClientName}}
		{{range $feature := $key.Features}}{{$feature}} {{end}} will expire in {{$key.ExpTermDays}} day
		Expiration date: {{$key.ExpTimeStr}}
{{ end }}

Hope that helps. Thanks!
`
