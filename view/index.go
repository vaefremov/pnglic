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
		"version": "0.0.1",
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
	default:
		Clients(c, &params)
	}
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
	c.HTML(http.StatusOK, "index.html", params)
}

type KeyOut struct {
	api.HWKey
	ClientName string
}

// Keys output the Keys page
func Keys(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
	keys, err := db.Keys()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	keysOut := []KeyOut{}
	for _, k := range keys {
		if orgName, err := db.ClientNameByID(k.OrgId); err == nil {
			keysOut = append(keysOut, KeyOut{HWKey: k, ClientName: orgName})
		} else {
			fmt.Println(k.OrgId, err)
		}
	}
	(*params)["keys"] = keysOut
	c.HTML(http.StatusOK, "keys.html", params)
}

// Keys output the Keys page
func KeyFeatures(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
	keyID := c.Query("keyId")
	features, err := db.LicensesSetByKeyId(keyID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	(*params)["features"] = features
	(*params)["keyId"] = keyID
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
