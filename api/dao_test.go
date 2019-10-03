package api_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/vaefremov/pnglic/api"
)

var testDB *api.DbConn

func TestKeys(t *testing.T) {
	// db := api.MustNewPool(api.DSN)
	db := testDB
	res, err := db.Keys()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	// expLen := 140
	expLen := 3
	if len(res) != expLen {
		t.Errorf("Expected: %d, got: %d", expLen, len(res))
	}
}

func TestClients(t *testing.T) {
	// db := api.MustNewPool(api.DSN)
	db := testDB
	res, err := db.Clients()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	expLen := 2
	if len(res) != expLen {
		t.Errorf("Expected: %d, got: %d", expLen, len(res))
	}
}

func TestHistoryForClientId(t *testing.T) {
	// db := api.MustNewPool(api.DSN)
	db := testDB
	clientId := 2
	// clientId := 55
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
	// db := api.MustNewPool(api.DSN)
	db := testDB
	keyId := "123abc"
	res, err := db.LicensesSetByKeyId(keyId)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	// t.Error("Intentionally")
}

func TestFeatures(t *testing.T) {
	// db := api.MustNewPool(api.DSN)
	db := testDB
	res, err := db.Features()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	expLen := 4
	if len(res) != expLen {
		t.Errorf("Feature list length: expected: %d got: %d", expLen, len(res))
	}
}
func TestPackageContent(t *testing.T) {
	// db := api.MustNewPool(api.DSN)
	db := testDB
	pkg := "P1"
	res, err := db.PackageContent(pkg)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	expLen := 2
	if len(res) != expLen {
		t.Errorf("Feature list length: expected: %d got: %d", expLen, len(res))
	}
}

func TestIsPackage(t *testing.T) {
	// db := api.MustNewPool(api.DSN)
	db := testDB
	pkg := "P1"
	res, err := db.IsPackage(pkg)
	if err != nil {
		t.Error(err)
	}
	if !res {
		t.Errorf("%s is expected to b a package", pkg)
	}
	pkg = "F1"
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

func TestUpdateLicenseSet(t *testing.T) {
	// db := api.MustInMemoryPool()
	keyId := "fake_key"
	featureFormat := "QQ%02d"
	start, _ := time.Parse("02/01/2006", "20/06/2018")
	end, _ := time.Parse("02/01/2006", "21/12/2019")
	ls := []api.LicenseSetItem{
		api.LicenseSetItem{KeyID: keyId, Feature: fmt.Sprintf(featureFormat, 1), Version: 19.01, Count: 2, Start: start, End: end, DupGroup: "DISP"},
		api.LicenseSetItem{KeyID: keyId, Feature: fmt.Sprintf(featureFormat, 2), Version: 19.01, Count: 2, Start: start, End: end, DupGroup: "DISP"},
	}
	err := testDB.UpdateLicenseSet(keyId, ls)
	if err != nil {
		t.Error(err)
	}
	ls1, err := testDB.LicensesSetByKeyId(keyId)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(ls, ls1) {
		t.Error("Should be equal", ls, ls1)
	}
	lsBad := []api.LicenseSetItem{
		api.LicenseSetItem{KeyID: keyId, Feature: fmt.Sprintf(featureFormat, 1), Version: 19.01, Count: 2, Start: start, End: end, DupGroup: "DISP"},
		api.LicenseSetItem{KeyID: keyId, Feature: fmt.Sprintf(featureFormat, 1), Version: 19.01, Count: 2, Start: start, End: end, DupGroup: "DISP"},
	}
	err = testDB.UpdateLicenseSet(keyId, lsBad)
	if err == nil {
		t.Error("Expected: UNIQUE constraint failed: licensesets.keyid, licensesets.feat")
	}
}

func TestMain(m *testing.M) {
	testDB = api.MustInMemoryTestPool()
	os.Exit(m.Run())
}
