package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOAuthBase(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/me", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
  "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "bio": "string",
  "groups": [
    "3fa85f64-5717-4562-b3fc-2c963f66afa6"
  ],
  "tags": [
    {
      "tagId": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "tag": "string",
      "isLocked": true,
      "createdAt": "2020-05-04T09:37:56.510Z",
      "updatedAt": "2020-05-04T09:37:56.510Z"
    }
  ],
  "updatedAt": "2020-05-04T09:37:56.510Z",
  "lastOnline": "2020-05-04T09:37:56.510Z",
  "twitterId": "string",
  "name": "string",
  "displayName": "string",
  "iconFileId": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "bot": true,
  "state": 0,
  "permissions": [
    "get_webhook"
  ],
  "homeChannel": "3fa85f64-5717-4562-b3fc-2c963f66afa6"
}`)
	})

	apiServer := httptest.NewServer(mux)
	defer apiServer.Close()

	strBaseURL := apiServer.URL
	authBase, err := NewOAuthBase(strBaseURL)
	if err != nil {
		t.Fatalf("Failed In NewAuthBase: %#v", err)
	}

	res, err := authBase.getMe("")
	if err != nil {
		t.Fatalf("Failed In getMe: %#v", err)
	}

	if res.Name != "string" {
		t.Fatalf("Invalid UserName: %s", res.Name)
	}

	if res.Id != "3fa85f64-5717-4562-b3fc-2c963f66afa6" {
		t.Fatalf("Invalid UserId: %s", res.Id)
	}
}

func TestLauncherAuth(t *testing.T) {
	launcherAuthBase := new(LauncherAuthBase)

	var versionID uint = 0
	var productKey string = "xxxxx-xxxxx-xxxxx-xxxxx-xxxxx"
	sessMap := map[interface{}]interface{} {
		"versionID": versionID,
		"productKey": productKey,
	}
	resVersionID, err := launcherAuthBase.getVersionID(sessMap)
	if err != nil {
		t.Fatalf("Failed In getVersionID: %#v", err)
	}
	if resVersionID != versionID {
		t.Fatalf("Invalid versionID: %d", resVersionID)
	}
	resVersionID, err = launcherAuthBase.getVersionID(map[interface{}]interface{}{})
	if err == nil {
		t.Fatal("VersionID Expected To Be Null,But Error Is Null")
	}

	resProductKey, err := launcherAuthBase.getProductKey(sessMap)
	if err != nil {
		t.Fatalf("Failed In getProductKey: %#v", err)
	}
	if resProductKey != productKey {
		t.Fatalf("Invalid productKey: %s", resProductKey)
	}
	resProductKey, err = launcherAuthBase.getProductKey(map[interface{}]interface{}{})
	if err == nil {
		t.Fatal("ProductKey Expected To Be Null, But Error Is Null")
	}
}