package api_test

import (
	"fmt"
	"testing"

	"github.com/vaefremov/pnglic/api"
)

func TestKeys(t *testing.T) {
	db := api.MustNewPool(api.DSN)
	res, err := db.Keys()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	expLen := 140
	if len(res) != expLen {
		t.Errorf("Expected: %d, got: %d", expLen, len(res))
	}
}

func TestClients(t *testing.T) {
	db := api.MustNewPool(api.DSN)
	res, err := db.Clients()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	expLen := 49
	if len(res) != expLen {
		t.Errorf("Expected: %d, got: %d", expLen, len(res))
	}
}

func TestHistoryForClientId(t *testing.T) {
	db := api.MustNewPool(api.DSN)
	clientId := 55
	res, err := db.HistoryForClientId(clientId)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	expLen := 1
	if len(res) != expLen {
		t.Errorf("Expected: %d, got: %d", expLen, len(res))
	}
}

func TestLicensesSetByKeyId(t *testing.T) {
	db := api.MustNewPool(api.DSN)
	keyId := "d123eab"
	res, err := db.LicensesSetByKeyId(keyId)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	// t.Error("Intentionally")
}

func TestFeatures(t *testing.T) {
	db := api.MustNewPool(api.DSN)
	res, err := db.Features()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	expLen := 215
	if len(res) != expLen {
		t.Errorf("Feature list length: expected: %d got: %d", expLen, len(res))
	}
}
func TestPackageContent(t *testing.T) {
	db := api.MustNewPool(api.DSN)
	pkg := "RV_INTERFACE"
	res, err := db.PackageContent(pkg)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	expLen := 13
	if len(res) != expLen {
		t.Errorf("Feature list length: expected: %d got: %d", expLen, len(res))
	}
}

func TestIsPackage(t *testing.T) {
	db := api.MustNewPool(api.DSN)
	pkg := "RV_INTERFACE"
	res, err := db.IsPackage(pkg)
	if err != nil {
		t.Error(err)
	}
	if !res {
		t.Errorf("%s is expected to b a package", pkg)
	}
	pkg = "RV_DBSCAN"
	res, err = db.IsPackage(pkg)
	if err != nil {
		t.Error(err)
	}
	if res {
		t.Errorf("%s is  not expected to b a package", pkg)
	}

	pkg = "IMPOSSIBLE!"
	res, err = db.IsPackage(pkg)
	if err == nil {
		t.Error(err)
	}
	if res {
		t.Errorf("%s is  not expected to b a package", pkg)
	}

}
