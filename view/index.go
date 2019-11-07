package view

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/api"
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
	c.HTML(http.StatusOK, "index.html", params)
}

type ClientOut struct {
	api.Organization
	Keys []string
}

// Clients output list of clients
func Clients(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
	clients, err := db.Clients()
	clientsOut := []ClientOut{}
	for _, cl := range clients {
		keys := []string{} // ID of keys belonging to an organization
		if tmp, err := db.KeysOfOrg(cl.Id); err == nil {
			for _, k := range tmp {
				keys = append(keys, k.Id)
			}
		} else {
			fmt.Println(cl.Id, err)
		}
		curClient := ClientOut{Organization: cl, Keys: keys}
		clientsOut = append(clientsOut, curClient)
	}
	(*params)["clients"] = clientsOut
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.HTML(http.StatusOK, "clients.html", params)
}

type KeyOut struct {
	api.HWKey
	ClientName string
}

// Keys output the Keys page
func Keys(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
	selectedOrgID := -1
	if tmp, err := strconv.Atoi(c.Query("orgId")); err == nil {
		selectedOrgID = tmp
	}
	keys, err := db.Keys()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	keysOut := []KeyOut{}
	for _, k := range keys {
		if orgName, err := db.ClientNameByID(k.OrgId); err == nil {
			if selectedOrgID == -1 || selectedOrgID == k.OrgId {
				keysOut = append(keysOut, KeyOut{HWKey: k, ClientName: orgName})
			}
		} else {
			fmt.Println(k.OrgId, err)
		}
	}
	(*params)["keys"] = keysOut
	c.HTML(http.StatusOK, "keys.html", params)
}

type FeatureOut struct {
	api.LicenseSetItem
	IsPackage      bool
	InsideFeatures []string
}

// KeyFeatures output page that implements use cases related to extending term of
// use of licenses
func KeyFeatures(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
	keyID := c.Query("keyId")

	features, err := db.LicensesSetByKeyId(keyID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	featuresOut := []FeatureOut{}
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
		}
		featuresOut = append(featuresOut, tmp)
	}
	client, err := db.KeyOfWhichOrg(keyID)
	(*params)["features"] = featuresOut
	(*params)["keyId"] = keyID
	(*params)["client"] = client
	(*params)["proposedExtTerm"] = time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	c.HTML(http.StatusOK, "keyfeatures.html", params)
}

type HistoryItemView struct {
	api.HistoryItem
	TimeOfIssueStr string
}

// History outputs history of issues
func History(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
	clientID := 1
	if tmp, err := strconv.Atoi(c.Query("clientId")); err == nil {
		clientID = tmp
	} else {
		(*params)["history"] = []api.HistoryItem{}
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
	db := c.MustGet("db").(*api.DbConn)
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
	// db := c.MustGet("db").(*api.DbConn)
	c.HTML(http.StatusOK, "templates.html", params)
}

func PackagesContent(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
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
	db := c.MustGet("db").(*api.DbConn)
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
	mapFeatures := map[string]api.Feature{}
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
