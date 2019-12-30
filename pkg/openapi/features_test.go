package openapi_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vaefremov/pnglic/pkg/dao"
	"github.com/vaefremov/pnglic/pkg/openapi"
)

func TestProlongLicensedFeaturesForKeyImpl(t *testing.T) {
	db := dao.MustInMemoryTestPool()
	c, w := newTestContext(db)
	// Get the initital features
	c.Params = []gin.Param{gin.Param{Key: "keyId", Value: "123abc"}}
	initFeatures := []openapi.LicensedFeature{}
	openapi.LicensedFeaturesForKey(c)
	if err := json.Unmarshal(w.Body.Bytes(), &initFeatures); err != nil {
		t.Error(err)
	}
	fmt.Println(initFeatures)

	c, w = newTestContext(db)
	buf := new(bytes.Buffer)
	c.Request, _ = http.NewRequest("POST", "/v1/prolongLicensedFeaturesForKey/123abc?till=2018-04-30", buf)
	c.Params = []gin.Param{gin.Param{Key: "keyId", Value: "123abc"}}
	openapi.ProlongLicensedFeaturesForKeyImpl(c)
	if w.Code != 202 {
		t.Error("Return code not OK", w.Code, w.Body)
		fmt.Println(w.Body)
	}

	c, w = newTestContext(db)
	// Get the initital features
	c.Params = []gin.Param{gin.Param{Key: "keyId", Value: "123abc"}}
	openapi.LicensedFeaturesForKey(c)
	if err := json.Unmarshal(w.Body.Bytes(), &initFeatures); err != nil {
		t.Error("After setting till", err)
	}
	fmt.Println(initFeatures)
	expectedEnd := "2018-04-30"
	// expectedEnd := time.Now().AddDate(0, 10, 0).Format("2006-01-02")
	if initFeatures[0].End != expectedEnd {
		t.Errorf("Final End of license date not as expected: %s != %s\n", initFeatures[0].End, expectedEnd)
	}

	c, w = newTestContext(db)
	buf = new(bytes.Buffer)
	c.Request, _ = http.NewRequest("POST", "/v1/prolongLicensedFeaturesForKey/123abc?byMonths=10", buf)
	c.Params = []gin.Param{gin.Param{Key: "keyId", Value: "123abc"}}
	openapi.ProlongLicensedFeaturesForKeyImpl(c)
	if w.Code != 202 {
		t.Error("Return code not OK", w.Code, w.Body)
	}

	// Check the final dates dates
	c, w = newTestContext(db)
	// Get the initital features
	c.Params = []gin.Param{gin.Param{Key: "keyId", Value: "123abc"}}

	openapi.LicensedFeaturesForKey(c)
	if err := json.Unmarshal(w.Body.Bytes(), &initFeatures); err != nil {
		t.Error("After setting by months", err)
	}
	fmt.Println(initFeatures)
	expectedEnd = time.Now().AddDate(0, 10, 0).Format("2006-01-02")
	if initFeatures[0].End != expectedEnd {
		t.Errorf("Final End of license date not as expected: %s != %s\n", initFeatures[0].End, expectedEnd)
	}

	// t.Error("nil")
}

func TestChangeLicensesCountImpl(t *testing.T) {
	db := dao.MustInMemoryTestPool()
	c, w := newTestContext(db)
	initFeatures := []openapi.LicensedFeature{}
	// Check the counts
	c, w = newTestContext(db)
	c.Params = []gin.Param{gin.Param{Key: "keyId", Value: "123abc"}}
	buf := new(bytes.Buffer)
	var expCount int32 = 20
	c.Request, _ = http.NewRequest("POST", fmt.Sprintf("/v1/prolongLicensedFeaturesForKey/123abc?setCount=%d", expCount), buf)
	openapi.ChangeLicensesCount(c)
	if w.Code != 202 {
		t.Error("Return code not OK", w.Code, w.Body)
	}

	c, w = newTestContext(db)
	c.Params = []gin.Param{gin.Param{Key: "keyId", Value: "123abc"}}
	openapi.LicensedFeaturesForKey(c)
	if err := json.Unmarshal(w.Body.Bytes(), &initFeatures); err != nil {
		t.Error("After setting count", err)
	}
	fmt.Println("After setting count:", initFeatures)
	assert.Equal(t, expCount, initFeatures[0].CountedFeature.Count)
}

func newTestContext(db *dao.DbConn) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("db", db)
	return c, w
}
