package object

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestMagicLinkTokenHash(t *testing.T) {
	token, err := generateMagicLinkToken()
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	hash := HashMagicLinkSecret(token)
	if hash == "" {
		t.Fatal("hash should not be empty")
	}
	if hash == token {
		t.Fatal("hash should not equal token")
	}
	if hash != HashMagicLinkSecret(token) {
		t.Fatal("hash should be stable")
	}
}

func TestNewMagicLinkExpireUsesApplicationSetting(t *testing.T) {
	application := &Application{
		Owner:                  "admin",
		Name:                   "app-built-in",
		Organization:           "built-in",
		MagicLinkExpireMinutes: 3,
	}
	before := time.Now().Unix()
	link := NewMagicLink(application, nil, "alice@example.com", "127.0.0.1", "admin", "raw-token", map[string]string{}, time.Time{})
	after := time.Now().Unix()
	if link.Owner != application.Organization {
		t.Fatalf("owner = %s, want %s", link.Owner, application.Organization)
	}
	if link.Application != application.Name {
		t.Fatalf("application = %s, want %s", link.Application, application.Name)
	}
	if link.IsUsed || link.Binding != MagicLinkBindingNone || link.SessionHash != getUnboundMagicLinkSessionHash(link.TokenHash) {
		t.Fatalf("a link of the API should be unused and unbound: %+v", link)
	}
	if link.TokenHash == "" {
		t.Fatal("token hash should not be empty")
	}
	if link.TokenHash == "raw-token" {
		t.Fatal("token hash should not equal raw token")
	}
	minExpireAt := before + int64(application.MagicLinkExpireMinutes*60) - 1
	maxExpireAt := after + int64(application.MagicLinkExpireMinutes*60) + 1
	if link.ExpireAt < minExpireAt || link.ExpireAt > maxExpireAt {
		t.Fatalf("expireAt = %d, want in [%d, %d]", link.ExpireAt, minExpireAt, maxExpireAt)
	}
}

func TestNewMagicLinkExpireUsesDefaultWhenUnset(t *testing.T) {
	application := &Application{
		Owner:        "admin",
		Name:         "app-built-in",
		Organization: "built-in",
	}
	before := time.Now().Unix()
	link := NewMagicLink(application, nil, "alice@example.com", "127.0.0.1", "admin", "raw-token", map[string]string{}, time.Time{})
	after := time.Now().Unix()
	minExpireAt := before + int64(MagicLinkDefaultExpireMinutes*60) - 1
	maxExpireAt := after + int64(MagicLinkDefaultExpireMinutes*60) + 1
	if link.ExpireAt < minExpireAt || link.ExpireAt > maxExpireAt {
		t.Fatalf("expireAt = %d, want in [%d, %d]", link.ExpireAt, minExpireAt, maxExpireAt)
	}
}

func TestApplyPermissionSnapshotToMagicLink(t *testing.T) {
	link := &MagicLink{}
	ApplyPermissionSnapshotToMagicLink(link, nil)
	if link.Permission != "" {
		t.Fatalf("permission = %s, want empty", link.Permission)
	}
	permission := &Permission{
		Owner:     "built-in",
		Name:      "magic-link-permission",
		Users:     []string{"built-in/alice"},
		Groups:    []string{"built-in/team"},
		Roles:     []string{"built-in/user"},
		Domains:   []string{"built-in"},
		Resources: []string{"resource1"},
		Actions:   []string{"Read"},
	}
	ApplyPermissionSnapshotToMagicLink(link, permission)
	if link.Permission != permission.GetId() {
		t.Fatalf("permission = %s, want %s", link.Permission, permission.GetId())
	}
	if len(link.SubGroups) != 1 || link.SubGroups[0] != "built-in/team" {
		t.Fatalf("unexpected subGroups: %v", link.SubGroups)
	}
	if len(link.Resources) != 1 || link.Resources[0] != "resource1" {
		t.Fatalf("unexpected resources: %v", link.Resources)
	}
}

func TestBuildMagicLinkCallbackURLResourceFallback(t *testing.T) {
	link := &MagicLink{}
	link.ResponseType = "login"
	link.Resources = []string{"resource-from-permission"}
	callbackURL := BuildMagicLinkCallbackURL(link, "token-1", "http://localhost:8000")
	parsed, err := url.Parse(callbackURL)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Query().Get("resource") != "resource-from-permission" {
		t.Fatalf("resource = %s, want resource-from-permission", parsed.Query().Get("resource"))
	}
	link.Resource = "resource-from-request"
	callbackURL = BuildMagicLinkCallbackURL(link, "token-2", "http://localhost:8000/")
	parsed, err = url.Parse(callbackURL)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Query().Get("resource") != "resource-from-request" {
		t.Fatalf("resource = %s, want resource-from-request", parsed.Query().Get("resource"))
	}
}

func TestResolveMagicLinkExpireTimeUsesRequestMinutes(t *testing.T) {
	application := &Application{
		MagicLinkExpireMinutes: 10,
	}
	now := time.Now()
	expireAt, err := ResolveMagicLinkExpireTime(application, 3, "", now)
	if err != nil {
		t.Fatal(err)
	}
	expected := now.Add(3 * time.Minute).Unix()
	actual := expireAt.Unix()
	if actual < expected-1 || actual > expected+1 {
		t.Fatalf("expireAt = %d, want around %d", actual, expected)
	}
}

func TestResolveMagicLinkExpireTimeUsesDefaultFromApplication(t *testing.T) {
	application := &Application{
		MagicLinkExpireMinutes: 7,
	}
	now := time.Now()
	expireAt, err := ResolveMagicLinkExpireTime(application, 0, "", now)
	if err != nil {
		t.Fatal(err)
	}
	expected := now.Add(7 * time.Minute).Unix()
	actual := expireAt.Unix()
	if actual < expected-1 || actual > expected+1 {
		t.Fatalf("expireAt = %d, want around %d", actual, expected)
	}
}

func TestResolveMagicLinkExpireTimeUsesExplicitTimestamp(t *testing.T) {
	application := &Application{
		MagicLinkExpireMinutes: 7,
	}
	now := time.Now()
	explicit := now.Add(9 * time.Minute).UTC().Format(time.RFC3339)
	expireAt, err := ResolveMagicLinkExpireTime(application, 0, explicit, now)
	if err != nil {
		t.Fatal(err)
	}
	if expireAt.UTC().Format(time.RFC3339) != explicit {
		t.Fatalf("expireAt = %s, want %s", expireAt.UTC().Format(time.RFC3339), explicit)
	}
}

func TestResolveMagicLinkExpireTimeRejectsTooShortMinutes(t *testing.T) {
	application := &Application{
		MagicLinkExpireMinutes: 10,
	}
	_, err := ResolveMagicLinkExpireTime(application, MagicLinkMinExpireMinutes-1, "", time.Now())
	if err == nil {
		t.Fatal("expected ttl validation error for too short expiresInMinutes")
	}
}

func TestResolveMagicLinkExpireTimeRejectsTooLongMinutes(t *testing.T) {
	application := &Application{
		MagicLinkExpireMinutes: 10,
	}
	_, err := ResolveMagicLinkExpireTime(application, MagicLinkMaxExpireMinutes+1, "", time.Now())
	if err == nil {
		t.Fatal("expected ttl validation error for too long expiresInMinutes")
	}
}

func TestResolveMagicLinkExpireTimeAllowsOneMonthLimit(t *testing.T) {
	application := &Application{
		MagicLinkExpireMinutes: 10,
	}
	now := time.Now()
	expireAt, err := ResolveMagicLinkExpireTime(application, MagicLinkMaxExpireMinutes, "", now)
	if err != nil {
		t.Fatal(err)
	}
	expected := now.Add(30 * 24 * time.Hour).Unix()
	actual := expireAt.Unix()
	if actual < expected-1 || actual > expected+1 {
		t.Fatalf("expireAt = %d, want around %d", actual, expected)
	}
}

func TestValidateMagicLinkConfigAllowsOneMonthLimit(t *testing.T) {
	application := &Application{
		MagicLinkSigninEnabled: true,
		MagicLinkExpireMinutes: MagicLinkMaxExpireMinutes,
	}
	err := ValidateMagicLinkConfig(application)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestResolveMagicLinkExpireTimeRejectsTooShortExplicitTime(t *testing.T) {
	application := &Application{
		MagicLinkExpireMinutes: 10,
	}
	now := time.Now()
	explicit := now.Add(30 * time.Second).UTC().Format(time.RFC3339)
	_, err := ResolveMagicLinkExpireTime(application, 0, explicit, now)
	if err == nil {
		t.Fatal("expected ttl validation error for too short explicit expireTime")
	}
}

func TestResolveMagicLinkExpireTimeRejectsTooLongExplicitTime(t *testing.T) {
	application := &Application{
		MagicLinkExpireMinutes: 10,
	}
	now := time.Now()
	explicit := now.Add(time.Duration(MagicLinkMaxExpireMinutes+1) * time.Minute).UTC().Format(time.RFC3339)
	_, err := ResolveMagicLinkExpireTime(application, 0, explicit, now)
	if err == nil {
		t.Fatal("expected ttl validation error for too long explicit expireTime")
	}
}

func TestGetApiMagicLinkEmailContent(t *testing.T) {
	link := &MagicLink{}
	link.ExpiryTime = "2026-06-17T10:00:00Z"
	link.AuthAction = MagicLinkAuthActionSigninExistingUser

	content := getApiMagicLinkEmailContent(&Provider{Content: "Your code is %s"}, "https://example.com/magic-link/callback?token=t", nil, link)
	if strings.Contains(content, "%expireTime") || strings.Contains(content, "%link") {
		t.Fatal("expected the placeholders of the default content to be replaced")
	}
	if !strings.Contains(content, link.ExpiryTime) || !strings.Contains(content, "https://example.com/magic-link/callback?token=t") {
		t.Fatal("expected the expiry and the link in the default content")
	}

	provider := &Provider{
		Content:                "content %link",
		MagicLinkContent:       "signin %link until %expireTime for %{user.friendlyName}",
		MagicLinkSignupContent: "signup %link",
	}
	content = getApiMagicLinkEmailContent(provider, "https://l", &User{DisplayName: "Alice"}, link)
	if content != "signin https://l until 2026-06-17T10:00:00Z for Alice" {
		t.Fatalf("unexpected signin content: %s", content)
	}

	link.AuthAction = MagicLinkAuthActionSignupNewUser
	content = getApiMagicLinkEmailContent(provider, "https://l", nil, link)
	if content != "signup https://l" {
		t.Fatalf("unexpected signup content: %s", content)
	}

	provider.MagicLinkContent = ""
	provider.MagicLinkSignupContent = ""
	content = getApiMagicLinkEmailContent(provider, "https://l", nil, link)
	if content != "content https://l" {
		t.Fatalf("unexpected provider content: %s", content)
	}
}

func TestValidateMagicLinkConfigRejectsInvalidExpireMinutes(t *testing.T) {
	application := &Application{
		MagicLinkExpireMinutes: MagicLinkMinExpireMinutes - 1,
	}
	err := ValidateMagicLinkConfig(application)
	if err == nil {
		t.Fatal("expected validation error for invalid magicLinkExpireMinutes")
	}
}

func TestValidateMagicLinkConfigRejectsNegativeRateLimits(t *testing.T) {
	application := &Application{
		MagicLinkRateLimitWindowMinutes: -1,
	}
	err := ValidateMagicLinkConfig(application)
	if err == nil {
		t.Fatal("expected validation error for negative rate-limit settings")
	}
}

func TestValidateMagicLinkVerifyState(t *testing.T) {
	nowUnix := time.Now().Unix()
	cases := []struct {
		name   string
		status string
		expire int64
		hasErr bool
	}{
		{name: "created status", status: MagicLinkStatusCreated, expire: nowUnix + 60, hasErr: false},
		{name: "sent status", status: MagicLinkStatusSent, expire: nowUnix + 60, hasErr: false},
		{name: "opened status", status: MagicLinkStatusOpened, expire: nowUnix + 60, hasErr: false},
		{name: "used status", status: MagicLinkStatusUsed, expire: nowUnix + 60, hasErr: true},
		{name: "failed status", status: MagicLinkStatusFailed, expire: nowUnix + 60, hasErr: true},
		{name: "revoked status", status: MagicLinkStatusRevoked, expire: nowUnix + 60, hasErr: true},
		{name: "expired by status", status: MagicLinkStatusExpired, expire: nowUnix + 60, hasErr: true},
		{name: "expired by time", status: MagicLinkStatusSent, expire: nowUnix - 1, hasErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			link := &MagicLink{}
			link.Status = tc.status
			link.ExpireAt = tc.expire
			err := validateMagicLinkVerifyState(link, nowUnix)
			if tc.hasErr && err == nil {
				t.Fatal("expected verify state error")
			}
			if !tc.hasErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
