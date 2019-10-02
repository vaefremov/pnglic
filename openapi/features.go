package openapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/api"
)

// LicensedFeaturesForKeyImpl - Returns list of all license features related to a given key
func LicensedFeaturesForKeyImpl(c *gin.Context) {
	db := c.MustGet("db").(*api.DbConn)
	keyID := c.Param("keyId")
	tmp, err := db.LicensesSetByKeyId(keyID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	res := []LicensedFeature{}
	for _, f := range tmp {
		res = append(res, LicensedFeature{CountedFeature: CountedFeature{
			Name:    f.Feature,
			Version: f.Version, Count: int32(f.Count), DupGroup: f.DupGroup},
			Start: f.Start.Format("2006-01-02"), End: f.End.Format("2006-01-02")})
	}
	c.JSON(http.StatusOK, res)
}

// ListFeaturesImpl - Returns list of features
func ListFeaturesImpl(c *gin.Context) {
	db := c.MustGet("db").(*api.DbConn)
	tmp, err := db.Features()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	res := []Feature{}
	for _, f := range tmp {
		res = append(res, Feature{Name: f.Feature, IsPackage: f.IsPackage, Description: f.Description})
	}
	c.JSON(http.StatusOK, res)
}

// PackageContentImpl - Returns list of features belonging to the specified package.
// Returns InlineResponse200 struct
func PackageContentImpl(c *gin.Context) {
	db := c.MustGet("db").(*api.DbConn)
	packageName := c.Param("packageName")
	isPackage, err := db.IsPackage(packageName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 20, Message: fmt.Sprintf("%s is an invalid feature name", packageName)})
		return
	}
	if !isPackage {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 21, Message: fmt.Sprintf("%s is not a package", packageName)})
		return
	}

	tmp, err := db.PackageContent(packageName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	res := InlineResponse200{Package: Feature{Name: packageName, IsPackage: true}}
	tmpFeatures := []Feature{}
	for _, f := range tmp {
		tmpFeatures = append(tmpFeatures, Feature{Name: f.Feature})
	}
	res.Features = tmpFeatures
	c.JSON(http.StatusOK, res)
}

// UpdateLicensedFeaturesForKeyImpl - Update license features for the given key ID, replace the previousely defined ones
func UpdateLicensedFeaturesForKeyImpl(c *gin.Context) {
	db := c.MustGet("db").(*api.DbConn)
	keyID := c.Param("keyId")
	// TODO: should check that keyId ia a valid key
	res := []LicensedFeature{}
	err := c.BindJSON(&res)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 20, Message: "Malformed input: " + err.Error()})
		return
	}
	fmt.Println(res)
	newLicset := []api.LicenseSetItem{}
	for _, f := range res {
		start, err := time.Parse("2006-01-02", f.Start)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 20, Message: "Malformed input: " + err.Error()})
			return
		}
		end, err := time.Parse("2006-01-02", f.End)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 20, Message: "Malformed input: " + err.Error()})
			return
		}
		newFeature := api.LicenseSetItem{KeyID: keyID, Feature: f.CountedFeature.Name, Version: f.CountedFeature.Version, Count: int(f.CountedFeature.Count), Start: start, End: end, DupGroup: f.CountedFeature.DupGroup}
		newLicset = append(newLicset, newFeature)
	}
	fmt.Println(newLicset)
	err = db.UpdateLicenseSet(keyID, newLicset)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 20, Message: "Input rejected: " + err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, "")
}
