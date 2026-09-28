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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/session"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// initControllerTestDb needs a disposable Postgres instance via TEST_POSTGRES_DSN.
func initControllerTestDb(t *testing.T) {
	t.Helper()
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set, skipping the database-backed controller test")
	}
	t.Setenv("driverName", "postgres")
	t.Setenv("dataSourceName", dsn)
	t.Setenv("dbName", "casdoor")
	t.Setenv("showSql", "false")
	object.InitConfig()

	manager, err := session.NewManager("memory", &session.ManagerConfig{CookieName: "casdoor_session_id", Gclifetime: 3600, Maxlifetime: 3600})
	if err != nil {
		t.Fatal(err)
	}
	web.GlobalSessions = manager
}

// callVerifyMagicLink runs /api/verify-magic-link for the token and returns the JSON answer.
func callVerifyMagicLink(t *testing.T, token string) (*Response, *ApiController) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/api/verify-magic-link?token="+url.QueryEscape(token), nil)
	recorder := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(recorder, request)
	store, err := web.GlobalSessions.SessionStart(recorder, request)
	if err != nil {
		t.Fatal(err)
	}
	ctx.Input.CruSession = store

	c := &ApiController{}
	c.Init(ctx, "ApiController", "VerifyMagicLink", c)
	c.VerifyMagicLink()

	resp := &Response{}
	if err = json.Unmarshal(recorder.Body.Bytes(), resp); err != nil {
		t.Fatalf("unexpected answer %q: %v", recorder.Body.String(), err)
	}
	return resp, c
}

// TestVerifyMagicLinkAsksForMfa — a magic link proves the email only: a user with MFA gets the
// NextMfa step, exactly like the built-in magic link sign-in, and is not signed in yet.
func TestVerifyMagicLinkAsksForMfa(t *testing.T) {
	initControllerTestDb(t)
	org := "mfatest" + util.GenerateId()[:8]

	if _, err := object.AddOrganization(&object.Organization{Owner: "admin", Name: org, DisplayName: org, PasswordType: "plain"}); err != nil {
		t.Fatal(err)
	}
	application := &object.Application{
		Owner:         "admin",
		Name:          "app" + org,
		Organization:  org,
		SigninMethods: []*object.SigninMethod{{Name: "Magic link", Rule: "None"}},
	}
	if _, err := object.AddApplication(application, "en"); err != nil {
		t.Fatal(err)
	}

	users := []*object.User{
		{Owner: org, Name: "with-mfa", Email: "mfa@example.com", EmailVerified: true, PreferredMfaType: object.TotpType, TotpSecret: "JBSWY3DPEHPK3PXP"},
		{Owner: org, Name: "without-mfa", Email: "plain@example.com", EmailVerified: true},
	}
	for _, user := range users {
		if _, err := object.AddUser(user, "en"); err != nil {
			t.Fatal(err)
		}
	}

	issue := func(email string) string {
		token, link, err := object.NewApiMagicLink(application, nil, email, "192.0.2.1", "", map[string]string{}, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		link.Status = object.MagicLinkStatusSent
		if err = object.AddMagicLink(link); err != nil {
			t.Fatal(err)
		}
		return token
	}

	resp, c := callVerifyMagicLink(t, issue("mfa@example.com"))
	if resp.Status != "ok" || resp.Data != object.NextMfa {
		t.Fatalf("user with MFA: got %s %q %v, want the %s step", resp.Status, resp.Msg, resp.Data, object.NextMfa)
	}
	if c.GetSessionUsername() != "" {
		t.Fatalf("user with MFA must not be signed in before the second factor, got %q", c.GetSessionUsername())
	}
	if c.getMfaUserSession() != org+"/with-mfa" {
		t.Fatalf("MFA session = %q, want %s/with-mfa", c.getMfaUserSession(), org)
	}

	resp, c = callVerifyMagicLink(t, issue("plain@example.com"))
	if resp.Status != "ok" || resp.Data == object.NextMfa {
		t.Fatalf("user without MFA: got %s %q %v, want a sign-in", resp.Status, resp.Msg, resp.Data)
	}
	if c.GetSessionUsername() != org+"/without-mfa" {
		t.Fatalf("user without MFA: session user = %q, want %s/without-mfa", c.GetSessionUsername(), org)
	}
}
