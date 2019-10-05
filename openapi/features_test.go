package openapi_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/api"
	"github.com/vaefremov/pnglic/openapi"
)

func TestProlongLicensedFeaturesForKeyImpl(t *testing.T) {
	db := api.MustInMemoryTestPool()
	c, w := newTestContext(db)
	// Get the initital features
	c.Params = []gin.Param{gin.Param{Key: "keyId", Value: "123abc"}, gin.Param{Key: "till", Value: "2018-04-30"}}

	initFeatures := []openapi.LicensedFeature{}
	openapi.LicensedFeaturesForKey(c)
	if err := json.Unmarshal(w.Body.Bytes(), &initFeatures); err != nil {
		t.Error(err)
	}
	fmt.Println(initFeatures)

	c, w = newTestContext(db)
	c.Params = []gin.Param{gin.Param{Key: "keyId", Value: "123abc"}, gin.Param{Key: "till", Value: "2018-04-30"}}
	openapi.ProlongLicensedFeaturesForKeyImpl(c)
	if w.Code != 202 {
		t.Error("Return code not OK", w.Code, w.Body)
		fmt.Println(w.Body)
	}

	c, w = newTestContext(db)
	c.Params = []gin.Param{gin.Param{Key: "keyId", Value: "123abc"}, gin.Param{Key: "byMonths", Value: "10"}}
	openapi.ProlongLicensedFeaturesForKeyImpl(c)
	if w.Code != 202 {
		t.Error("Return code not OK", w.Code, w.Body)
	}

	// Check the final dates dates
	c, w = newTestContext(db)
	// Get the initital features
	c.Params = []gin.Param{gin.Param{Key: "keyId", Value: "123abc"}}

	openapi.LicensedFeaturesForKey(c)
	json.Unmarshal(w.Body.Bytes(), &initFeatures)
	fmt.Println(initFeatures)
	expectedEnd := "2019-03-02"
	if initFeatures[0].End != expectedEnd {
		t.Errorf("Final End of license date not as expected: %s %s\n", initFeatures[0].End, expectedEnd)
	}
	// t.Error("nil")
}

func newTestContext(db *api.DbConn) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("db", db)
	return c, w
}
