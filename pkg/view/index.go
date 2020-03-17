package view

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/config"
	"github.com/vaefremov/pnglic/pkg/chkexprd"
	"github.com/vaefremov/pnglic/pkg/dao"
	"github.com/vaefremov/pnglic/pkg/mailnotify"
)

// Index is the index handler.
func Index(c *gin.Context) {
	// Dispatch to the right page template, index.html is the default
	params := gin.H{
		"title":   "Pangea Licenses",
		"version": "0.1.4",
	}
	switch c.Param("c") {
	case "/keys.html":
		Keys(c, &params)
	case "/keyfeatures.html":
		KeyFeatures(c, &params)
	case "/licenses.html":
		History(c, &params)
	case "/features.html":
		Features(c, &params)
	case "/templates.html":
		Templates(c, &params)
	case "/packagescontent.html":
		PackagesContent(c, &params)
	case "/singlepackage.html":
		SinglePackageContent(c, &params)
	case "/clients.html":
		Clients(c, &params)
	default:
		StartPage(c, &params)
	}
}

func StartPage(c *gin.Context, params *gin.H) {
	conf := c.MustGet("conf").(*config.Config)
	db := c.MustGet("db").(*dao.DbConn)
	expTerm := time.Duration(time.Duration(24) * time.Duration(conf.DaysToExpire1) * time.Hour)
	expired, err := chkexprd.FindFeaturesWillExpire(db, expTerm)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	expiredNo := len(expired)
	fmt.Println(conf.DaysToExpire1)
	(*params)["expired_no"] = expiredNo
	(*params)["exp_term"] = expTerm
	(*params)["exp_days"] = conf.DaysToExpire1
	(*params)["will_expire"] = expired
	c.HTML(http.StatusOK, "index.html", params)
}

type FeatureOut struct {
	dao.LicenseSetItem
	EltId          string
	IsPackage      bool
	InsideFeatures []string
}

// KeyFeatures output page that implements use cases related to extending term of
// use of licenses
func KeyFeatures(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*dao.DbConn)
	keyID := c.Query("keyId")
	fullPage := (c.Query("fullPage") == "true")
	features, err := db.LicensesSetByKeyId(keyID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	proposedCount := 1
	featuresOut := []FeatureOut{}
	ind := 0
	for _, f := range features {
		isPackage, _ := db.IsPackage(f.Feature)
		packageContentStr := []string{}
		if isPackage {
			if tmpPackageContent, err := db.PackageContent(f.Feature); err == nil {
				for _, ff := range tmpPackageContent {
					packageContentStr = append(packageContentStr, ff.Feature)
				}
			}
		}
		tmp := FeatureOut{LicenseSetItem: f,
			IsPackage:      isPackage,
			InsideFeatures: packageContentStr,
			EltId:          fmt.Sprintf("F_%d", ind),
		}
		if tmp.Count > proposedCount {
			proposedCount = tmp.Count
		}
		featuresOut = append(featuresOut, tmp)
		ind += 1
	}
	client, err := db.KeyOfWhichOrg(keyID)
	conf := c.MustGet("conf").(*config.Config)
	(*params)["features"] = featuresOut
	(*params)["keyId"] = keyID
	(*params)["client"] = client
	(*params)["proposedExtTerm"] = time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	(*params)["proposedCount"] = proposedCount
	(*params)["mailTo"] = conf.AdminMail
	(*params)["licenseFileName"] = mailnotify.MakeLicenseFileName(client.Name, keyID)
	(*params)["fullPage"] = fullPage
	c.HTML(http.StatusOK, "keyfeatures.html", params)
}

type HistoryItemView struct {
	dao.HistoryItem
	TimeOfIssueStr string
}

// History outputs history of issues
func History(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*dao.DbConn)
	clientID := 1
	if tmp, err := strconv.Atoi(c.Query("clientId")); err == nil {
		clientID = tmp
	} else {
		(*params)["history"] = []dao.HistoryItem{}
		c.HTML(http.StatusOK, "licenses.html", params)
		fmt.Println("Unable to get client ID")
		return
	}
	history, err := db.HistoryForClientId(clientID)
	fmt.Println("History lenfth:", len(history))
	sort.Slice(history, func(i, j int) bool { return history[i].IssueTime.After(history[j].IssueTime) })
	historyOut := []HistoryItemView{}
	for _, h := range history {
		historyOut = append(historyOut, HistoryItemView{HistoryItem: h, TimeOfIssueStr: h.IssueTime.Format(time.RFC3339)})
	}
	(*params)["history"] = historyOut
	(*params)["clientId"] = clientID
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.HTML(http.StatusOK, "licenses.html", params)
}

func Features(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*dao.DbConn)
	features, err := db.Features()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	sort.Slice(features, func(i, j int) bool {
		if features[i].IsPackage != features[j].IsPackage {
			return features[i].IsPackage && !features[j].IsPackage
		}
		return features[i].Feature < features[j].Feature
	})
	(*params)["features"] = features
	c.HTML(http.StatusOK, "features.html", params)
}

func Templates(c *gin.Context, params *gin.H) {
	// db := c.MustGet("db").(*dao.DbConn)
	c.HTML(http.StatusOK, "templates.html", params)
}

func PackagesContent(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*dao.DbConn)
	features, err := db.Features()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	featuresOut := []struct {
		Name      string
		IsPackage bool
		Features  []string
	}{}
	for _, f := range features {
		packageContentStr := []string{}
		if f.IsPackage {
			if tmpPackageContent, err := db.PackageContent(f.Feature); err == nil {
				for _, ff := range tmpPackageContent {
					packageContentStr = append(packageContentStr, ff.Feature)
				}
			}
			tmp := struct {
				Name      string
				IsPackage bool
				Features  []string
			}{
				Name:      f.Feature,
				IsPackage: f.IsPackage,
				Features:  packageContentStr,
			}
			featuresOut = append(featuresOut, tmp)
		}
	}
	(*params)["features"] = featuresOut
	c.HTML(http.StatusOK, "packagescontent.html", params)

}

func SinglePackageContent(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*dao.DbConn)
	packageName := c.Query("package")
	if packageName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if isPackage, err := db.IsPackage(packageName); err != nil || !isPackage {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	tmpPackageContent, err := db.PackageContent(packageName)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	tmpFeatures, err := db.Features()
	mapFeatures := map[string]dao.Feature{}
	for _, f := range tmpFeatures {
		if f.Description == "" {
			f.Description = "No description so far"
		}
		mapFeatures[f.Feature] = f
	}
	features := []struct {
		Name        string
		Description string
	}{}
	for _, f := range tmpPackageContent {
		features = append(features, struct {
			Name        string
			Description string
		}{
			Name:        f.Feature,
			Description: mapFeatures[f.Feature].Description,
		})
	}
	(*params)["package"] = packageName
	(*params)["packageDescription"] = mapFeatures[packageName].Description

	(*params)["features"] = features
	(*params)["total"] = len(features)

	c.HTML(http.StatusOK, "singlepackage.html", params)
}
