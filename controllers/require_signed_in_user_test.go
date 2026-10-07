// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web"
	beegocontext "github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func callGetRecords(t *testing.T, sessionUser string, query url.Values) (*Response, []*object.Record) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/api/get-records?"+query.Encode(), nil)
	recorder := httptest.NewRecorder()
	ctx := beegocontext.NewContext()
	ctx.Reset(recorder, request)
	store, err := web.GlobalSessions.SessionStart(recorder, request)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Set(context.Background(), "username", sessionUser); err != nil {
		t.Fatal(err)
	}
	ctx.Input.CruSession = store

	c := &ApiController{}
	c.Init(ctx, "ApiController", "GetRecords", c)
	c.GetRecords()

	resp := &Response{}
	if err = json.Unmarshal(recorder.Body.Bytes(), resp); err != nil {
		t.Fatalf("unexpected answer %q: %v", recorder.Body.String(), err)
	}

	records := []*object.Record{}
	if resp.Status == "ok" {
		data, err := json.Marshal(resp.Data)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(data, &records); err != nil {
			t.Fatal(err)
		}
	}
	return resp, records
}

// TestRequireSignedInUserAcceptsAppCredential — an application signed in with its client
// credentials (session user "app/<name>") and no ?userId= is the admin of its organization:
// get-records answers with that organization's records instead of "The user: app/<name> doesn't exist".
func TestRequireSignedInUserAcceptsAppCredential(t *testing.T) {
	initControllerTestDb(t)
	suffix := util.GenerateId()[:8]
	own, other := "own"+suffix, "other"+suffix
	appName := "app" + own

	for _, organization := range []string{own, other} {
		if _, err := object.AddOrganization(&object.Organization{Owner: "admin", Name: organization, DisplayName: organization, PasswordType: "plain"}); err != nil {
			t.Fatal(err)
		}
		if !object.AddRecord(&object.Record{Name: util.GenerateId(), Owner: organization, Organization: organization, CreatedTime: util.GetCurrentTime(), Method: "POST", Action: "test-" + suffix}) {
			t.Fatalf("record of %s was not added", organization)
		}
	}
	if _, err := object.AddApplication(&object.Application{Owner: "admin", Name: appName, Organization: own}, "en"); err != nil {
		t.Fatal(err)
	}

	query := url.Values{"pageSize": {"100"}, "p": {"1"}, "field": {"action"}, "value": {"test-" + suffix}}
	resp, records := callGetRecords(t, "app/"+appName, query)
	if resp.Status != "ok" {
		t.Fatalf("app credential without userId: %s", resp.Msg)
	}
	if len(records) != 1 || records[0].Organization != own {
		t.Fatalf("app credential sees %d records, want the one of its own organization %s", len(records), own)
	}

	// acting as a user of another organization stays refused
	query.Set("userId", other+"/admin")
	if resp, _ = callGetRecords(t, "app/"+appName, query); resp.Status == "ok" {
		t.Fatal("app credential acted as a user of another organization")
	}

	// an unknown application is still not a user
	resp, _ = callGetRecords(t, "app/missing"+suffix, url.Values{"pageSize": {"10"}, "p": {"1"}})
	if resp.Status == "ok" || !strings.Contains(resp.Msg, "doesn't exist") {
		t.Fatalf("unknown application: got %q, want \"doesn't exist\"", resp.Msg)
	}
}
