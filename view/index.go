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
	case "/licenses.html":
		History(c, &params)
	case "/features.html":
		Features(c, &params)
	default:
		Clients(c, &params)
	}
}

// Clients output list of clients
func Clients(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
	clients, err := db.Clients()
	(*params)["clients"] = clients
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.HTML(http.StatusOK, "index.html", params)
}

// Keys output the Keys page
func Keys(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
	keys, err := db.Keys()
	(*params)["keys"] = keys
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.HTML(http.StatusOK, "keys.html", params)
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
	}
	c.HTML(http.StatusOK, "licenses.html", params)
}

func Features(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
	features, err := db.Features()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
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
