package view

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.HTML(http.StatusOK, "keys.html", params)
	case "/licenses.html":
		c.HTML(http.StatusOK, "licenses.html", params)
	case "/features.html":
		c.HTML(http.StatusOK, "features.html", params)
	default:
		c.HTML(http.StatusOK, "index.html", params)
	}
}
