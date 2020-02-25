package openapi

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/config"
	"github.com/vaefremov/pnglic/pkg/dao"
	"github.com/vaefremov/pnglic/pkg/mailnotify"
)

// HistoryLicenseFileImpl - Get license file by client id and timestamp of issue
func HistoryLicenseFileImpl(c *gin.Context) {
	db := c.MustGet("db").(*dao.DbConn)
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

// MakeLicenseFileImpl - Generate license file from the current set of licenses related to key ID and store it in the history
func MakeLicenseFileImpl(c *gin.Context) {
	db := c.MustGet("db").(*dao.DbConn)
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
	resXML, err = signLicenseFile(c, resXML)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	err = db.AddToHistory(clientID, time.Now(), resXML)
	if mailTo := c.Query("mailTo"); mailTo != "" {
		clientName, err := db.ClientNameByID(clientID)
		if err != nil {
			log.Println("Error when sending file ", err)
		}
		conf := c.MustGet("conf").(*config.Config)
		log.Println("Mailing file to ", mailTo)
		notificator := mailnotify.New(conf.MailServer, conf.MailPort, conf.MailUser, conf.MailPass)
		notificator.AddTo(mailTo).AddTo(conf.BackMail)
		if err := notificator.SendFile(clientName, keyID, []byte(resXML)); err != nil {
			log.Println("Error when sending file ", err)
		}
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

func makeXMLFromTemplate(keyID string, db *dao.DbConn, licSet []dao.LicenseSetItem) (res string, err error) {
	// Sort licenses: individual features first
	sort.Slice(licSet, func(i, j int) bool {
		isPackageI, _ := db.IsPackage(licSet[i].Feature)
		isPackageJ, _ := db.IsPackage(licSet[j].Feature)
		if isPackageI == isPackageJ {
			return licSet[i].Feature < licSet[j].Feature
		}
		return !isPackageI
	})

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

func signLicenseFile(c *gin.Context, bodyXML string) (res string, err error) {
	conf := c.MustGet("conf").(*config.Config)
	xmlfilepath, err := tmpXMLFile(bodyXML)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(xmlfilepath)

	cmd := exec.Command(conf.LicfileEncoderLegacy, "-i", xmlfilepath, "-s", conf.SecretsHasp)

	cmdStdOut, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	cmdStdErr, err := cmd.StderrPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		return
	}
	stdout, err := ioutil.ReadAll(cmdStdOut)
	if err != nil {
		return
	}
	stderr, err := ioutil.ReadAll(cmdStdErr)
	if err != nil {
		return
	}
	if len(stderr) > 0 {
		return "", fmt.Errorf("encoding utility reported errors %s", string(stderr))
	}

	return string(stdout), nil
}

func tmpXMLFile(bodyXML string) (path string, err error) {
	tempfile, err := ioutil.TempFile("/tmp", "licfile_*.xml")
	if err != nil {
		return
	}
	if _, err = tempfile.Write([]byte(bodyXML)); err != nil {
		return
	}
	if err = tempfile.Close(); err != nil {
		return
	}
	return tempfile.Name(), nil
}
