package openapi_test

import (
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/api"
	"github.com/vaefremov/pnglic/openapi"
	"github.com/vaefremov/pnglic/server"
)

func TestLicenseFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("db", api.MustNewPool(api.DSN))
	c.Params = []gin.Param{gin.Param{Key: "clientId", Value: "55"}, gin.Param{Key: "timeOfIssue", Value: "2018-04-26T14:24:54Z"}}
	openapi.HistoryLicenseFileImpl(c)
	if w.Code != 200 {
		t.Error("Return code not OK", w.Code)
	}
	// fmt.Println(w.Body)
	// t.Error(nil)
}

const mockLicenseFile = `<?xml version="1.0"?>
<!DOCTYPE license_server>

<license_server port="1234" id="4CDCEE4C">

<package
        id="CT_INTERFACE" 
        version="1.00" 
        start="26.04.2018" 
        end="18.06.2018" 
        count="1" 
        dupgroup="DISP" 
        code="CF4449610C33DAC3A9C737CD4D93FFE1" >
    <feature id="MAPCENTER" />
</package>
</license_server>
`

func TestExtractKeyIDFromXML(t *testing.T) {
	keyId, err := openapi.ExtractKeyIDFromXML(mockLicenseFile)
	if err != nil {
		t.Error(err)
	}
	expectedKeyId := "4CDCEE4C"
	if keyId != expectedKeyId {
		t.Errorf("Expected id: %s got: %s", expectedKeyId, keyId)
	}
}

func TestMakeLicenseFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("db", api.MustInMemoryTestPool())
	c.Set("conf", server.NewConfig(""))
	c.Params = []gin.Param{gin.Param{Key: "clientId", Value: "1"}, gin.Param{Key: "keyId", Value: "123abc"}}
	openapi.MakeLicenseFileImpl(c)
	if w.Code != http.StatusOK {
		t.Error("Wrong status")
	}
}
