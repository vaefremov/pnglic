package openapi

import (
	"encoding/xml"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/api"
)

// LicenseFile - Get license file by client id and timestamp of issue
func LicenseFileImpl(c *gin.Context) {
	db := c.MustGet("db").(*api.DbConn)
	clientID, err := strconv.Atoi(c.Param("clientId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 3, Message: err.Error()})
		return
	}
	wantedTime, err := time.Parse(time.RFC3339, c.Param("timeOfIssue"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 3, Message: err.Error()})
		return
	}
	hist, err := db.HistoryForClientId(clientID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	lastFound := ""
	for _, h := range hist {
		if wantedTime.Equal(h.IssueTime) {
			lastFound = h.ContentXml
		}
	}
	if lastFound == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, Error{Code: 404, Message: "No license file"})
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(lastFound))
}

type licenseServer struct {
	XMLName xml.Name `xml:"license_server"`
	ID      string   `xml:"id,attr"`
}

// ExtractKeyIDFromXML is an auxiliary function to extract hardware key ID from license file
func ExtractKeyIDFromXML(xmlBody string) (keyID string, err error) {
	var ls licenseServer
	err = xml.Unmarshal([]byte(xmlBody), &ls)
	return ls.ID, err
}
