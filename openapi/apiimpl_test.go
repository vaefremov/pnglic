package openapi_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/vaefremov/pnglic/api"
	"github.com/vaefremov/pnglic/openapi"
)

func TestCreateKeyImpl(t *testing.T) {
	db := api.MustInMemoryTestPool()
	c, w := newTestContext(db)
	buf := new(bytes.Buffer)
	newKeyId := "ffffff"
	orgId := 1
	jsonIn := fmt.Sprintf(`{"id": "%s", "kind": "HASP", "currentOwnerId": %d, "comments": "No comments!"}`, newKeyId, orgId)
	buf.WriteString(jsonIn)
	c.Request, _ = http.NewRequest("PUT", "/v1/keys", buf)
	openapi.CreateKeyImpl(c)
	if w.Code != http.StatusCreated {
		t.Error("Unexpected status", w.Body)
	}
	if res, err := db.IsKeyBelongsToOrg(newKeyId, orgId); err == nil {
		if !res {
			t.Error("Key", newKeyId, "now should belong to org ", orgId)
		}
	} else {
		t.Error(err)
	}
}
