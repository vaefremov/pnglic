package view

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/api"
)

type ClientOut struct {
	api.Organization
	Keys []string
}

// Clients output list of clients
func Clients(c *gin.Context, params *gin.H) {
	db := c.MustGet("db").(*api.DbConn)
	clients, err := db.Clients()
	sortOrder := c.Query("sort")
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

	// Sorting according with the sorting parameter
	sort.Slice(clientsOut, func(i, j int) bool {
		switch sortOrder {
		case "Name":
			return clientsOut[i].Name < clientsOut[j].Name
		default:
			return clientsOut[i].Id < clientsOut[j].Id
		}
	})

	c.HTML(http.StatusOK, "clients.html", params)
}
