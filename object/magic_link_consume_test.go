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

package object

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/xorm"
)

func newMagicLinkTestOrmer(t *testing.T) *Ormer {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite", filepath.Join(t.TempDir(), "magic_link.db"))
	if err != nil {
		t.Fatal(err)
	}
	engine.SetMaxOpenConns(1)

	previous := ormer
	ormer = &Ormer{Engine: engine}
	t.Cleanup(func() {
		ormer = previous
		_ = engine.Close()
	})
	return ormer
}

func newMagicLinkTestApplication() *Application {
	return &Application{
		Owner:         "admin",
		Name:          "app",
		Organization:  "org",
		SigninMethods: []*SigninMethod{{Name: "Magic link", Rule: "None"}},
	}
}

func addMagicLinkTestLink(t *testing.T, application *Application, email string, expireAt time.Time) (string, *MagicLink) {
	t.Helper()

	token, link, err := NewApiMagicLink(application, nil, email, "10.0.0.1", "", map[string]string{}, expireAt)
	if err != nil {
		t.Fatal(err)
	}
	link.Status = MagicLinkStatusSent
	if err = AddMagicLink(link); err != nil {
		t.Fatal(err)
	}
	return token, link
}

func TestMagicLinkConsumeIsOneTimeAndAtomic(t *testing.T) {
	a := newMagicLinkTestOrmer(t)
	if err := a.syncMagicLink(); err != nil {
		t.Fatal(err)
	}
	application := newMagicLinkTestApplication()
	token, link := addMagicLinkTestLink(t, application, "a@example.com", time.Time{})
	sessionHash := getApiMagicLinkSessionHash(link, "", "")

	var wg sync.WaitGroup
	var mu sync.Mutex
	won := 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := ConsumeMagicLink(token, sessionHash, application, "en"); err == nil {
				mu.Lock()
				won++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if won != 1 {
		t.Fatalf("the link was claimed %d times, want exactly once", won)
	}

	if _, err := ConsumeMagicLink(token, sessionHash, application, "en"); err == nil || !strings.Contains(err.Error(), "already been used") {
		t.Fatalf("a second claim should be refused as used, got: %v", err)
	}
}

func TestMagicLinkConsumeBinding(t *testing.T) {
	a := newMagicLinkTestOrmer(t)
	if err := a.syncMagicLink(); err != nil {
		t.Fatal(err)
	}
	application := newMagicLinkTestApplication()

	// a link of the API may be opened anywhere, but never through a browser session of the
	// built-in sign-in page
	token, link := addMagicLinkTestLink(t, application, "a@example.com", time.Time{})
	if _, err := ConsumeMagicLink(token, HashMagicLinkSecret("some-browser-session"), application, "en"); err == nil {
		t.Fatal("a foreign session hash should not claim an unbound link")
	}
	if _, err := ConsumeMagicLink(token, getApiMagicLinkSessionHash(link, "", ""), application, "en"); err != nil {
		t.Fatalf("unexpected error for an unbound link: %v", err)
	}

	token, link, err := NewApiMagicLink(application, nil, "b@example.com", "10.0.0.2", "", map[string]string{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	BindMagicLinkToClient(link, "client-secret")
	if err = AddMagicLink(link); err != nil {
		t.Fatal(err)
	}
	if link.Binding != MagicLinkBindingClient {
		t.Fatalf("binding = %q, want %q", link.Binding, MagicLinkBindingClient)
	}
	for _, secret := range []string{"", "wrong-secret"} {
		if _, err = ConsumeMagicLink(token, getApiMagicLinkSessionHash(link, secret, ""), application, "en"); err == nil {
			t.Fatalf("the secret %q should not claim a bound link", secret)
		}
	}
	if _, err = ConsumeMagicLink(token, getApiMagicLinkSessionHash(link, "client-secret", ""), application, "en"); err != nil {
		t.Fatalf("unexpected error for the right secret: %v", err)
	}

	// another application cannot claim the link
	token, link = addMagicLinkTestLink(t, application, "c@example.com", time.Time{})
	other := newMagicLinkTestApplication()
	other.Name = "other"
	if _, err = ConsumeMagicLink(token, getApiMagicLinkSessionHash(link, "", ""), other, "en"); err == nil {
		t.Fatal("a link should only be claimed for its own application")
	}
}

func TestMagicLinkConsumeHonorsOwnExpiry(t *testing.T) {
	a := newMagicLinkTestOrmer(t)
	if err := a.syncMagicLink(); err != nil {
		t.Fatal(err)
	}
	application := newMagicLinkTestApplication()

	// longer than the verification code timeout of the built-in links
	token, link := addMagicLinkTestLink(t, application, "a@example.com", time.Now().Add(48*time.Hour))
	link.Time = time.Now().Add(-24 * time.Hour).Unix()
	if _, err := ormer.Engine.Where("name = ?", link.Name).Cols("time").Update(link); err != nil {
		t.Fatal(err)
	}
	if _, err := ConsumeMagicLink(token, getApiMagicLinkSessionHash(link, "", ""), application, "en"); err != nil {
		t.Fatalf("a link within its own TTL should be claimed: %v", err)
	}

	token, link = addMagicLinkTestLink(t, application, "b@example.com", time.Now().Add(time.Hour))
	link.ExpireAt = time.Now().Add(-time.Second).Unix()
	if _, err := ormer.Engine.Where("name = ?", link.Name).Cols("expire_at").Update(link); err != nil {
		t.Fatal(err)
	}
	if _, err := ConsumeMagicLink(token, getApiMagicLinkSessionHash(link, "", ""), application, "en"); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("a link past its own TTL should be refused as expired, got: %v", err)
	}

	builtIn := &MagicLink{Time: time.Now().Add(-time.Duration(getMagicLinkTimeout()+1) * time.Minute).Unix()}
	if !isMagicLinkExpired(builtIn, time.Now().Unix()) {
		t.Fatal("a built-in link should expire after the verification code timeout")
	}
	builtIn.Time = time.Now().Unix()
	if isMagicLinkExpired(builtIn, time.Now().Unix()) {
		t.Fatal("a fresh built-in link should not be expired")
	}
}

func TestRevokedMagicLinkIsRefused(t *testing.T) {
	a := newMagicLinkTestOrmer(t)
	if err := a.syncMagicLink(); err != nil {
		t.Fatal(err)
	}
	application := newMagicLinkTestApplication()
	token, link := addMagicLinkTestLink(t, application, "a@example.com", time.Time{})

	revoked, err := RevokeMagicLink(util.GetId(link.Owner, link.Name))
	if err != nil || !revoked {
		t.Fatalf("revoke = %v, %v", revoked, err)
	}
	revoked, err = RevokeMagicLink(util.GetId(link.Owner, link.Name))
	if err != nil || revoked {
		t.Fatalf("a second revoke should change nothing, got %v, %v", revoked, err)
	}

	if _, err = ConsumeMagicLink(token, getApiMagicLinkSessionHash(link, "", ""), application, "en"); err == nil {
		t.Fatal("the built-in claim should refuse a revoked link")
	}
	stored, err := GetMagicLinkByToken(token)
	if err != nil || stored == nil {
		t.Fatalf("stored link = %v, %v", stored, err)
	}
	err = validateMagicLinkVerifyState(stored, time.Now().Unix())
	if err == nil || !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("the API should name the revoked state, got: %v", err)
	}

	// a used link cannot be revoked any more
	token, link = addMagicLinkTestLink(t, application, "b@example.com", time.Time{})
	if _, err = ConsumeMagicLink(token, getApiMagicLinkSessionHash(link, "", ""), application, "en"); err != nil {
		t.Fatal(err)
	}
	revoked, err = RevokeMagicLink(util.GetId(link.Owner, link.Name))
	if err != nil || revoked {
		t.Fatalf("a used link should not be revoked, got %v, %v", revoked, err)
	}
}

func TestMagicLinkRateLimit(t *testing.T) {
	a := newMagicLinkTestOrmer(t)
	if err := a.syncMagicLink(); err != nil {
		t.Fatal(err)
	}
	application := newMagicLinkTestApplication()
	application.MagicLinkRateLimitEmail = 2
	application.MagicLinkRateLimitIP = 3
	application.MagicLinkRateLimitWindowMinutes = 5

	if err := IsMagicLinkAllowSend("a@example.com", "10.0.0.1", application); err != nil {
		t.Fatalf("unexpected limit on an empty table: %v", err)
	}
	addMagicLinkTestLink(t, application, "a@example.com", time.Time{})
	addMagicLinkTestLink(t, application, "a@example.com", time.Time{})

	err := IsMagicLinkAllowSend("a@example.com", "10.0.0.9", application)
	if err == nil || !strings.Contains(err.Error(), "too many magic links requested for this email") {
		t.Fatalf("expected the email limit, got: %v", err)
	}
	if err = IsMagicLinkAllowSend("b@example.com", "10.0.0.1", application); err != nil {
		t.Fatalf("two links from the IP are within its limit: %v", err)
	}
	addMagicLinkTestLink(t, application, "b@example.com", time.Time{})
	err = IsMagicLinkAllowSend("c@example.com", "10.0.0.1", application)
	if err == nil || !strings.Contains(err.Error(), "too many magic links requested from this IP") {
		t.Fatalf("expected the IP limit, got: %v", err)
	}

	// a request that was turned down counts as well, a link outside the window does not
	failed := NewMagicLink(application, nil, "d@example.com", "10.0.0.4", "", "", map[string]string{}, time.Time{})
	if err = AddFailedMagicLink(failed, "no such user"); err != nil {
		t.Fatal(err)
	}
	required, err := IsMagicLinkCaptchaRequired("d@example.com", "10.0.0.5", application)
	if err != nil || !required {
		t.Fatalf("captcha required = %v, %v", required, err)
	}
	if _, err = ormer.Engine.Where("email = ?", "a@example.com").Cols("time").Update(&MagicLink{Time: time.Now().Add(-6 * time.Minute).Unix()}); err != nil {
		t.Fatal(err)
	}
	if err = IsMagicLinkAllowSend("a@example.com", "10.0.0.9", application); err != nil {
		t.Fatalf("links outside the window should not count: %v", err)
	}

	if err = checkMagicLinkRateLimit(0, 0, 100, &Application{}); err == nil || !strings.Contains(err.Error(), "for this application") {
		t.Fatalf("expected the default application limit, got: %v", err)
	}
}

func TestMagicLinkOrigin(t *testing.T) {
	if _, err := getMagicLinkOrigin("evil.example.com", "en"); err == nil {
		t.Fatal("a non-loopback host without a configured origin should be refused")
	}
	if err := CheckMagicLinkOrigin("localhost:8000", "en"); err != nil {
		t.Fatalf("a loopback host should be accepted: %v", err)
	}

	link := &MagicLink{}
	link.ResponseType = "code"
	link.ClientId = "client"
	link.RedirectUri = "https://app.example.com/callback"
	link.State = "s t"
	callbackUrl := BuildMagicLinkCallbackURL(link, "tok", "https://app.example.com/")
	if !strings.HasPrefix(callbackUrl, "https://app.example.com/magic-link/callback?") {
		t.Fatalf("unexpected callback url: %s", callbackUrl)
	}
	for _, part := range []string{"token=tok", "client_id=client", "response_type=code", "state=s+t", "redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback"} {
		if !strings.Contains(callbackUrl, part) {
			t.Fatalf("callback url %s lacks %s", callbackUrl, part)
		}
	}

	for signinPath, want := range map[string]bool{
		"/login/app":                  true,
		"/login/oauth/authorize?a=b":  true,
		"//evil.example.com/login":    false,
		"https://evil.example.com":    false,
		"/signup/app":                 false,
		"/login/app#fragment":         false,
		"":                            false,
		"/cas/org/app/login?service=": true,
	} {
		if isValidSigninPath(signinPath) != want {
			t.Fatalf("isValidSigninPath(%q) = %v, want %v", signinPath, !want, want)
		}
	}
}

// legacyMagicLink is the table of the fork before the built-in model was adopted.
type legacyMagicLink struct {
	Owner       string `xorm:"varchar(100) notnull pk"`
	Name        string `xorm:"varchar(100) notnull pk"`
	CreatedTime string `xorm:"varchar(100)"`
	Application string `xorm:"varchar(100) index"`
	Email       string `xorm:"varchar(100) index"`
	RemoteAddr  string `xorm:"varchar(100)"`
	TokenHash   string `xorm:"varchar(100) index"`
	Status      string `xorm:"varchar(20) index"`
	ExpiryTime  string `xorm:"varchar(100)"`
	ExpireAt    int64  `xorm:"index"`
}

func (legacyMagicLink) TableName() string {
	return "magic_link"
}

func TestLegacyMagicLinkTableIsMigrated(t *testing.T) {
	a := newMagicLinkTestOrmer(t)
	if err := a.Engine.Sync2(new(legacyMagicLink)); err != nil {
		t.Fatal(err)
	}

	application := newMagicLinkTestApplication()
	token, err := generateMagicLinkToken()
	if err != nil {
		t.Fatal(err)
	}
	issued := time.Now().Add(-time.Minute)
	rows := []*legacyMagicLink{
		{Owner: "org", Name: "pending", CreatedTime: issued.Format(time.RFC3339), Application: "admin/app", Email: "a@example.com", TokenHash: HashMagicLinkSecret(token), Status: MagicLinkStatusSent, ExpireAt: time.Now().Add(time.Hour).Unix()},
		{Owner: "org", Name: "used", CreatedTime: issued.Format(time.RFC3339), Application: "admin/app", Email: "b@example.com", TokenHash: HashMagicLinkSecret("used-token"), Status: MagicLinkStatusUsed, ExpireAt: time.Now().Add(time.Hour).Unix()},
		{Owner: "org", Name: "failed", CreatedTime: issued.Format(time.RFC3339), Application: "admin/app", Email: "c@example.com", Status: MagicLinkStatusFailed, ExpireAt: time.Now().Add(time.Hour).Unix()},
	}
	for _, row := range rows {
		if _, err = a.Engine.Insert(row); err != nil {
			t.Fatal(err)
		}
	}

	// twice: the step has to be repeatable
	for i := 0; i < 2; i++ {
		if err = a.syncMagicLink(); err != nil {
			t.Fatalf("sync #%d: %v", i+1, err)
		}
	}

	links, err := GetMagicLinks("org")
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != len(rows) {
		t.Fatalf("%d rows after the migration, want %d", len(links), len(rows))
	}
	for _, link := range links {
		if link.Application != "app" || link.Time != issued.Unix() || link.Binding != MagicLinkBindingNone {
			t.Fatalf("row %s was not converted: %+v", link.Name, link)
		}
		if link.IsUsed != (link.Name != "pending") {
			t.Fatalf("row %s: isUsed = %v", link.Name, link.IsUsed)
		}
	}

	pending, err := GetMagicLinkByToken(token)
	if err != nil || pending == nil {
		t.Fatalf("pending link = %v, %v", pending, err)
	}
	if _, err = ConsumeMagicLink(token, getApiMagicLinkSessionHash(pending, "", ""), application, "en"); err != nil {
		t.Fatalf("a link mailed before the migration should still sign in: %v", err)
	}
	used, err := GetMagicLinkByToken("used-token")
	if err != nil || used == nil {
		t.Fatalf("used link = %v, %v", used, err)
	}
	if _, err = ConsumeMagicLink("used-token", getApiMagicLinkSessionHash(used, "", ""), application, "en"); err == nil {
		t.Fatal("a link used before the migration should stay used")
	}
}
