package openapi

import (
	"encoding/xml"
	"fmt"
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

func MakeLicenseFileImpl(c *gin.Context) {
	db := c.MustGet("db").(*api.DbConn)
	clientID, err := strconv.Atoi(c.Param("clientId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 3, Message: err.Error()})
		return
	}
	keyID := c.Param("keyId")
	keyOk, err := db.IsKeyBelongsToOrg(keyID, clientID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 3, Message: err.Error()})
		return

	}
	if !keyOk {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 3, Message: fmt.Sprintf("Key %s does not belong to client id %d", keyID, clientID)})
		return
	}
	dbLicset, err := db.LicensesSetByKeyId(keyID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	resXML, err := makeXMLFromTemplate(keyID, db, dbLicset)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	resXML, err = signLicenseFile(resXML)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(resXML))
}

const featureTemplate = `<%s
	id="%s"
	version="%.2f"
	start="%s"
	end="%s"
	count="%d"
	dupgroup="%s"
	code="00000000000000000000000000000000" >
`

func makeXMLFromTemplate(keyID string, db *api.DbConn, licSet []api.LicenseSetItem) (res string, err error) {
	bodyXML := fmt.Sprintf(`<?xml version="1.0"?><!DOCTYPE license_server>
	
<license_server port="1234" id="%s">
	`, keyID)
	for _, f := range licSet {
		isPkg, err := db.IsPackage(f.Feature)
		if err != nil {
			return "", err
		}
		if isPkg {
			bodyXML += fmt.Sprintf(featureTemplate, "package", f.Feature, f.Version,
				f.Start.Format("02.01.2006"), f.End.Format("02.01.2006"),
				f.Count, f.DupGroup)
			if features, err := db.PackageContent(f.Feature); err == nil {
				for _, ff := range features {
					bodyXML += fmt.Sprintf("    <feature id=\"%s\" />\n", ff.Feature)
				}
			} else {
				return "", err
			}
			bodyXML += "</package>\n"
		} else {
			bodyXML += fmt.Sprintf(featureTemplate, "feature", f.Feature, f.Version,
				f.Start.Format("02.01.2006"), f.End.Format("02.01.2006"),
				f.Count, f.DupGroup)
			bodyXML += "</feature>\n"
		}
	}
	bodyXML += "</license_server>\n"
	return bodyXML, nil
}

// TODO: just a stub! replace with real code
func signLicenseFile(bodyXML string) (res string, err error) {
	return bodyXML, nil
}
