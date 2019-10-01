/*
 * PANGEA License Manager
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 0.1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/api"
)

// Route is the information for every URI.
type Route struct {
	// Name is the name of this Route.
	Name string
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method string
	// Pattern is the pattern of the URI.
	Pattern string
	// HandlerFunc is the handler function of this route.
	HandlerFunc gin.HandlerFunc
}

// Routes is the list of the generated Route.
type Routes []Route

func AddDatabase(db *api.DbConn) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

// NewRouter returns a new router.
func NewRouter(dsn string) *gin.Engine {
	db := api.MustNewPool(dsn)
	router := gin.Default()
	router.Use(AddDatabase(db))
	for _, route := range routes {
		switch route.Method {
		case http.MethodGet:
			router.GET(route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			router.POST(route.Pattern, route.HandlerFunc)
		case http.MethodPut:
			router.PUT(route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			router.DELETE(route.Pattern, route.HandlerFunc)
		}
	}

	return router
}

// Index is the index handler.
func Index(c *gin.Context) {
	c.String(http.StatusOK, "Hello World!")
}

var routes = Routes{
	{
		"Index",
		http.MethodGet,
		"/v1/",
		Index,
	},

	{
		"LicenseFile",
		http.MethodGet,
		"/v1/licenseFile/:clientId/:timeOfIssue",
		LicenseFile,
	},

	{
		"LicensedFeaturesForKey",
		http.MethodGet,
		"/v1/licensedFeaturesForKey/:keyId",
		LicensedFeaturesForKey,
	},

	{
		"ListClients",
		http.MethodGet,
		"/v1/clients",
		ListClients,
	},

	{
		"ListFeatures",
		http.MethodGet,
		"/v1/features",
		ListFeatures,
	},

	{
		"ListHistoryItems",
		http.MethodGet,
		"/v1/history/:clientId",
		ListHistoryItems,
	},

	{
		"ListKeys",
		http.MethodGet,
		"/v1/keys",
		ListKeys,
	},
	{
		"PackageContent",
		http.MethodGet,
		"/v1/packageContent/:packageName",
		PackageContent,
	},
	{
		"Ping",
		http.MethodGet,
		"/v1/ping",
		Ping,
	},
}
