package openapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/pkg/dao"
)

// PingImpl actually implements the logic behind ping request
func PingImpl(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Version": "0.0.1"})
}

// ListKeysImpl - Returns general list of keys
func ListKeysImpl(c *gin.Context) {
	// Process parameters: clientId
	clientID := -1
	if clientIDStr := c.Query("clientId"); clientIDStr != "" {
		if tmp, err := strconv.Atoi(clientIDStr); err == nil {
			clientID = tmp
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 3, Message: err.Error()})
			return
		}
	}
	res := []HardwareKey{}
	db := c.MustGet("db").(*dao.DbConn)
	dbKeys, err := db.Keys()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 1, Message: err.Error()})
		return
	}
	for _, k := range dbKeys {
		if clientID < 0 || k.OrgId == clientID {
			res = append(res, HardwareKey{Id: k.Id, Kind: "HASP", Comments: k.Comments, CurrentOwnerId: int32(k.OrgId)})
		}
	}
	c.IndentedJSON(http.StatusOK, res)
}

// CreateKeyImpl - Creates a new key and assigns it to the specified client
func CreateKeyImpl(c *gin.Context) {
	// TODO: make sure clientID and keyID are OK: client must exist, key must be a new one
	// fmt.Println(c.Request.Body
	db := c.MustGet("db").(*dao.DbConn)
	newKey := HardwareKey{}
	err := c.BindJSON(&newKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 3, Message: err.Error()})
		return
	}
	if newKey.Id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 3, Message: "Key ID must be specified"})
		return
	}
	if newKey.CurrentOwnerId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 3, Message: "Client ID must be specified"})
		return
	}
	err = db.CreateKey(dao.HWKey{Id: newKey.Id, OrgId: int(newKey.CurrentOwnerId), Comments: newKey.Kind + " " + newKey.Comments})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 3, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newKey)
}

// ListClientsImpl - Returns list of all organizations related to licensation
func ListClientsImpl(c *gin.Context) {
	res := []Organization{}
	db := c.MustGet("db").(*dao.DbConn)
	dbClients, err := db.Clients()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	for _, c := range dbClients {
		res = append(res, Organization{Id: int32(c.Id), Name: c.Name, Contacts: c.Contacts, Comments: c.Comments})
	}
	c.IndentedJSON(http.StatusOK, res)
}

// ListHistoryItems - Returns list of all organizations related to licensation
func ListHistoryItemsImpl(c *gin.Context) {
	db := c.MustGet("db").(*dao.DbConn)
	clientID, err := strconv.Atoi(c.Param("clientId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 3, Message: err.Error()})
		return
	}
	res := []HistoryItem{}
	hist, err := db.HistoryForClientId(clientID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	for _, h := range hist {
		res = append(res, HistoryItem{OrgName: h.ClientName, WhenIssued: h.IssueTime.Format(time.RFC3339)})
	}
	c.JSON(http.StatusOK, res)
}
