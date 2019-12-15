package chkexprd_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vaefremov/pnglic/pkg/chkexprd"
	"github.com/vaefremov/pnglic/server"
)

func TestMakeMessageFromTemplate(t *testing.T) {
	expTerm := time.Hour * 24
	expTime := time.Now()
	conf := server.Config{Port: 9995, PublicName: "some.host"}
	tmpReport := map[string]chkexprd.ExpFeaturesReportElt{
		"key1234": chkexprd.ExpFeaturesReportElt{ClientName: "Org 1", ExpTime: expTime, ExpTerm: expTerm, Features: []string{"F1", "F2"}},
		"key1235": chkexprd.ExpFeaturesReportElt{ClientName: "Org 1", ExpTime: expTime, ExpTerm: expTerm, Features: []string{"P1", "P2"}}}
	message, err := chkexprd.MakeMessageFromTemplate(tmpReport, expTerm, &conf)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 447, len(message))
	fmt.Println(message)
}
