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
	"sort"
	"testing"

	"github.com/beego/beego/v2/server/web"
	beegocontext "github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func callGetPermissions(t *testing.T, sessionUser string, query url.Values) []string {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/api/get-permissions?"+query.Encode(), nil)
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
	c.Init(ctx, "ApiController", "GetPermissions", c)
	c.GetPermissions()

	resp := struct {
		Status string               `json:"status"`
		Msg    string               `json:"msg"`
		Data   []*object.Permission `json:"data"`
	}{}
	if err = json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unexpected answer %q: %v", recorder.Body.String(), err)
	}
	if resp.Status != "ok" {
		t.Fatalf("get-permissions failed: %s", resp.Msg)
	}

	ids := []string{}
	for _, permission := range resp.Data {
		ids = append(ids, permission.GetId())
	}
	sort.Strings(ids)
	return ids
}

// TestGetPermissionsByUsersAndGroupIsScopedToRequester — the userIds and group lookups answer an
// org admin with the permissions of their own organization only, a global admin with all of them.
func TestGetPermissionsByUsersAndGroupIsScopedToRequester(t *testing.T) {
	initControllerTestDb(t)
	suffix := util.GenerateId()[:8]
	own, other := "own"+suffix, "other"+suffix
	group := own + "/staff"
	sharedUser := own + "/alice"

	for _, organization := range []string{own, other} {
		if _, err := object.AddOrganization(&object.Organization{Owner: "admin", Name: organization, DisplayName: organization, PasswordType: "plain"}); err != nil {
			t.Fatal(err)
		}
		if _, err := object.AddApplication(&object.Application{Name: "app" + organization, Organization: organization}, "en"); err != nil {
			t.Fatal(err)
		}
		permission := &object.Permission{
			Owner:        organization,
			Name:         "perm",
			Users:        []string{sharedUser},
			Groups:       []string{group},
			Roles:        []string{},
			Domains:      []string{},
			ResourceType: "Application",
			Resources:    []string{"app"},
			Actions:      []string{"Read"},
			Effect:       "Allow",
			IsEnabled:    true,
		}
		if _, err := object.AddPermission(permission); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := object.AddUser(&object.User{Owner: own, Name: "admin", IsAdmin: true}, "en"); err != nil {
		t.Fatal(err)
	}
	if builtIn, err := object.GetOrganization("admin/built-in"); err != nil {
		t.Fatal(err)
	} else if builtIn == nil {
		if _, err = object.AddOrganization(&object.Organization{Owner: "admin", Name: "built-in", DisplayName: "built-in", PasswordType: "plain"}); err != nil {
			t.Fatal(err)
		}
	}
	if globalAdmin, err := object.GetUser("built-in/admin"); err != nil {
		t.Fatal(err)
	} else if globalAdmin == nil {
		if _, err = object.AddUser(&object.User{Owner: "built-in", Name: "admin", IsAdmin: true}, "en"); err != nil {
			t.Fatal(err)
		}
	}

	ownOnly := []string{own + "/perm"}
	both := []string{other + "/perm", own + "/perm"}
	sort.Strings(both)

	for _, query := range []url.Values{{"userIds": {sharedUser}}, {"group": {group}}} {
		if got := callGetPermissions(t, own+"/admin", query); len(got) != 1 || got[0] != ownOnly[0] {
			t.Fatalf("org admin, %v: got %v, want %v", query, got, ownOnly)
		}
		if got := callGetPermissions(t, "built-in/admin", query); len(got) != 2 || got[0] != both[0] || got[1] != both[1] {
			t.Fatalf("global admin, %v: got %v, want %v", query, got, both)
		}
	}
}
