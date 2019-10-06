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

func TestIsKeyBelongsToOrg(t *testing.T) {
	db := testDB
	cases := []struct {
		keyID      string
		orgID      int
		expRes     bool
		isErrorExp bool
	}{
		{"123abc", 1, true, false},
		{"123cbc", 1, false, false},
		{"impossible!", 1, false, true},
		{"123cbc", 2, true, false},
		{"123cbc!!", 2, false, true},
		{"123cbc", 4, false, false},
	}
	for i, c := range cases {
		res, err := db.IsKeyBelongsToOrg(c.keyID, c.orgID)
		fmt.Println(i, c, res, err)
		if c.isErrorExp != (err != nil) {
			t.Error(i, "Test failed, error expected")
			continue
		}
		if c.expRes != res {
			t.Error(i, "Test failed, expected ", c.expRes, " got ", res)
			continue
		}
	}
}

func TestCreateKey(t *testing.T) {
	db := testDB
	newKey := api.HWKey{Id: "ffffff", OrgId: 2, Comments: "Key ffffff"}
	if err := db.CreateKey(newKey); err != nil {
		t.Error(err)
	}
	newKey = api.HWKey{Id: "ffffff", OrgId: 2, Comments: "Key ffffff"}
	if err := db.CreateKey(newKey); err == nil {
		t.Error(err)
	}
	existingKey := api.HWKey{Id: "123abc", OrgId: 2, Comments: "Key ffffff"}
	if err := db.CreateKey(existingKey); err == nil {
		t.Error("Error was expected")
	} else {
		fmt.Println("2", err)
	}
	invalidOrg := api.HWKey{Id: "afffff", OrgId: 20, Comments: "Key afffff"}
	if err := db.CreateKey(invalidOrg); err == nil {
		t.Error("Error was expected")
	} else {
		fmt.Println("20", err)
	}
	// t.Error()
}

func TestMain(m *testing.M) {
	testDB = api.MustInMemoryTestPool()
	os.Exit(m.Run())
}
