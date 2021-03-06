package openapi

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vaefremov/pnglic/pkg/dao"
)

// LicensedFeaturesForKeyImpl - Returns list of all license features related to a given key
func LicensedFeaturesForKeyImpl(c *gin.Context) {
	db := c.MustGet("db").(*dao.DbConn)
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
	db := c.MustGet("db").(*dao.DbConn)
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
	db := c.MustGet("db").(*dao.DbConn)
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
	db := c.MustGet("db").(*dao.DbConn)
	keyID := c.Param("keyId")
	// TODO: should check that keyId ia a valid key
	res := []LicensedFeature{}
	err := c.BindJSON(&res)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 20, Message: "Malformed input: " + err.Error()})
		return
	}
	newLicset := []dao.LicenseSetItem{}
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
		newFeature := dao.LicenseSetItem{KeyID: keyID, Feature: f.CountedFeature.Name, Version: f.CountedFeature.Version, Count: int(f.CountedFeature.Count), Start: start, End: end, DupGroup: f.CountedFeature.DupGroup}
		newLicset = append(newLicset, newFeature)
	}
	err = db.UpdateLicenseSet(keyID, newLicset)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 20, Message: "Input rejected: " + err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, "")
}

// ProlongLicensedFeaturesForKeyImpl - Update license features for the given key ID, replace the previousely defined ones
// Also, the count of the issued features may be set to a number specified by the count parameter.
func ProlongLicensedFeaturesForKeyImpl(c *gin.Context) {
	db := c.MustGet("db").(*dao.DbConn)
	keyID := c.Param("keyId")
	var byMonths int
	tillDate, err := time.Parse("2006-01-02", c.Query("till"))
	if err != nil {
		byMonths, err = strconv.Atoi(c.Query("byMonths"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 30, Message: "extension term must be set with either byMonths or till parameters " + err.Error()})
			return
		}
		if byMonths <= 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 30, Message: "invalid extension term (<= 0)"})
			return
		}
	}

	var newVersion float64 = 0.
	newVersionStr := c.Query("setVersion")
	if newVersionStr != "" {
		if newVersion, err = strconv.ParseFloat(newVersionStr, 32); err != nil {
			newVersion = 0.0
			log.Println("Wrong parameter value, ignored: ", err)
		}
	}

	// Retrieve list of feature this request should act on
	featuresSet := map[string]bool{}
	for _, f := range strings.Split(c.Query("restrictTo"), ",") {
		if f != "" {
			featuresSet[f] = true
		}
	}

	currentLicSet, err := db.LicensesSetByKeyId(keyID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	newLicset := []dao.LicenseSetItem{}
	for _, f := range currentLicSet {
		fnew := f
		if len(featuresSet) == 0 || featuresSet[f.Feature] {
			// we should have
			if byMonths > 0 {
				// fnew.End = f.End.AddDate(0, byMonths, 0)
				tillDate = time.Now().AddDate(0, byMonths, 0)
			}
			fnew.End = tillDate

			if newVersion > 0.0 {
				fnew.Version = float32(newVersion)
			}
			// LM_CONSOLE requires special treatment
			// @TODO: We should make sure the LM_CONSOLE is present
			if fnew.Feature == "LM_CONSOLE" {
				fnew.Version = 1.0
				// We intentionally add one extra year for LM_CONSOLE!
				fnew.End = tillDate.AddDate(1, 0, 0)
			}
		}
		newLicset = append(newLicset, fnew)
	}
	err = db.UpdateLicenseSet(keyID, newLicset)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 20, Message: "Input rejected: " + err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, "")
}

func ChangeLicensesCountImpl(c *gin.Context) {
	db := c.MustGet("db").(*dao.DbConn)
	keyID := c.Param("keyId")

	var err error
	var newCount int = 0
	if newCountStr := c.Query("setCount"); newCountStr != "" {
		if newCount, err = strconv.Atoi(newCountStr); err != nil {
			newCount = 0
			log.Println("Wrong value of the count parameter, ignored: ", err)
		}
	}
	// Retrieve features from the request
	featuresSet := map[string]bool{}
	for _, f := range strings.Split(c.Query("restrictTo"), ",") {
		if f != "" {
			featuresSet[f] = true
		}
	}

	currentLicSet, err := db.LicensesSetByKeyId(keyID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 2, Message: err.Error()})
		return
	}
	newLicset := []dao.LicenseSetItem{}
	for _, f := range currentLicSet {
		fnew := f
		if len(featuresSet) == 0 || featuresSet[f.Feature] {
			if newCount > 0 {
				fnew.Count = newCount
			}
		}
		newLicset = append(newLicset, fnew)
	}
	err = db.UpdateLicenseSet(keyID, newLicset)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 20, Message: "Input rejected: " + err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, "")
}

// CreateFeatureImpl implementes updating or creating new feature
func CreateFeatureImpl(c *gin.Context) {
	db := c.MustGet("db").(*dao.DbConn)
	featureName := c.Param("featureName")
	f := Feature{}
	err := c.BindJSON(&f)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 40, Message: err.Error()})
		return
	}
	if featureName != f.Name {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 40, Message: "inconsistent feature names"})
		return
	}
	upd, err := db.CreateOrUpdateFeature(f.Name, f.Description, f.IsPackage)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{Code: 40, Message: err.Error()})
		return
	}
	if upd {
		c.JSON(http.StatusOK, f)
	}
	c.JSON(http.StatusCreated, f)
}

// DeleteFeatureImpl - Deletes a nfeature
func DeleteFeatureImpl(c *gin.Context) {
	db := c.MustGet("db").(*dao.DbConn)
	featureName := c.Param("featureName")
	err := db.DeleteFeature(featureName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 40, Message: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdatePackageImpl - Creates new package with the given content or modifies an existing package
func UpdatePackageImpl(c *gin.Context) {
	db := c.MustGet("db").(*dao.DbConn)
	packageName := c.Param("packageName")
	features := []string{}
	err := c.BindJSON(&features)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 50, Message: err.Error()})
		return
	}
	fmt.Println(features)
	err = db.SetPackageContent(features, packageName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{Code: 50, Message: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
