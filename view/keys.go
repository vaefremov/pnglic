package view

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/api"
)

type KeyOut struct {
	api.HWKey
	ClientName string
}

// Keys output the Keys page
func Keys(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
	sortOrder := c.Query("sort")
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
	sort.Slice(keysOut, func(i, j int) bool {
		switch sortOrder {
		case "Id":
			return keysOut[i].HWKey.Id < keysOut[j].HWKey.Id
		case "clientName":
			fallthrough
		default:
			if keysOut[i].ClientName == keysOut[j].ClientName {
				return keysOut[i].HWKey.Id < keysOut[j].HWKey.Id
			}
			return keysOut[i].ClientName < keysOut[j].ClientName
		}
	})
	(*params)["keys"] = keysOut
	c.HTML(http.StatusOK, "keys.html", params)
}
