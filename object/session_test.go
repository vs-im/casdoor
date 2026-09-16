package object

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/session"
	"github.com/casdoor/casdoor/util"
)

// TestNormalizeSessionFilterField — фильтр get-sessions по field=user должен уходить
// в колонку name (у таблицы session нет колонки user: пользователь сессии — pk name).
func TestNormalizeSessionFilterField(t *testing.T) {
	if got := normalizeSessionFilterField("user"); got != "name" {
		t.Fatalf("normalizeSessionFilterField(user) = %q, want name", got)
	}
	for _, field := range []string{"name", "application", "createdTime", ""} {
		if got := normalizeSessionFilterField(field); got != field {
			t.Fatalf("normalizeSessionFilterField(%q) = %q, want unchanged", field, got)
		}
	}
}

// initPasswordGrantTestDb — одноразовая sqlite-БД во временном каталоге (env переопределяет
// conf/app.conf через conf.GetConfigString) + memory-провайдер Beego-сессий для DeleteBeegoSession.
func initPasswordGrantTestDb(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("driverName", "sqlite")
	t.Setenv("dataSourceName", "file:"+filepath.Join(dir, "casdoor.db")+"?cache=shared")
	t.Setenv("dbName", "casdoor")
	t.Setenv("showSql", "false")
	createDatabase = false
	InitConfig()

	manager, err := session.NewManager("memory", &session.ManagerConfig{CookieName: "casdoor_session_id", Gclifetime: 3600, Maxlifetime: 3600})
	if err != nil {
		t.Fatal(err)
	}
	web.GlobalSessions = manager
}

func addPasswordGrantFixtures(t *testing.T) (*Application, *User) {
	t.Helper()
	certificate, err := os.ReadFile("token_jwt_key.pem")
	if err != nil {
		t.Fatal(err)
	}
	privateKey, err := os.ReadFile("token_jwt_key.key")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = AddCert(&Cert{Owner: "admin", Name: "cert-cas6", Scope: "JWT", Type: "x509", CryptoAlgorithm: "RS256", Certificate: string(certificate), PrivateKey: string(privateKey)}); err != nil {
		t.Fatal(err)
	}
	if _, err = AddOrganization(&Organization{Owner: "admin", Name: "clients", DisplayName: "clients", PasswordType: "plain"}); err != nil {
		t.Fatal(err)
	}
	application := &Application{
		Owner:                "admin",
		Name:                 "admin-web",
		Organization:         "clients",
		ClientId:             "cas6-client",
		ClientSecret:         "cas6-secret",
		Cert:                 "cert-cas6",
		TokenFormat:          "JWT",
		ExpireInHours:        0.25,
		RefreshExpireInHours: 720,
		GrantTypes:           []string{"password", "refresh_token"},
	}
	if _, err = AddApplication(application); err != nil {
		t.Fatal(err)
	}
	user := &User{
		Owner:        "clients",
		Name:         "alice",
		Id:           "0199-cas6-alice",
		DisplayName:  "Alice",
		Email:        "alice@example.com",
		Password:     "correct horse",
		PasswordType: "plain",
	}
	if _, err = AddUser(user, "en"); err != nil {
		t.Fatal(err)
	}
	return application, user
}

// TestPasswordGrantCreatesSessionAndSidClaim — вход по password grant должен заводить Session-строку
// (get-sessions видит IP/UA), писать Token.SessionId и клейм sid; refresh наследует sid;
// delete-session по sessionId делает refresh невозможным (invalid_grant).
func TestPasswordGrantCreatesSessionAndSidClaim(t *testing.T) {
	initPasswordGrantTestDb(t)
	application, user := addPasswordGrantFixtures(t)
	const beegoSessionId = "beego-sid-cas6"

	token, tokenError, err := GetPasswordToken(application, user.Name, "correct horse", "", "localhost", &SessionInfo{
		SessionId: beegoSessionId,
		Ip:        "203.0.113.7",
		UserAgent: "Mozilla/5.0 (cas6-test)",
	})
	if err != nil {
		t.Fatal(err)
	}
	if tokenError != nil {
		t.Fatalf("password grant failed: %s: %s", tokenError.Error, tokenError.ErrorDescription)
	}
	if token.SessionId != beegoSessionId {
		t.Fatalf("Token.SessionId = %q, want %q", token.SessionId, beegoSessionId)
	}

	cert, err := getCert("admin", "cert-cas6")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ParseJwtToken(token.AccessToken, cert)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Sid != beegoSessionId {
		t.Fatalf("access token sid = %q, want %q", claims.Sid, beegoSessionId)
	}
	refreshClaims, err := ParseJwtToken(token.RefreshToken, cert)
	if err != nil {
		t.Fatal(err)
	}
	if refreshClaims.Sid != beegoSessionId {
		t.Fatalf("refresh token sid = %q, want %q", refreshClaims.Sid, beegoSessionId)
	}

	sessions, err := GetPaginationSessions("clients", 0, 10, "user", user.Name, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 {
		t.Fatalf("get-sessions returned %d rows, want 1", len(sessions))
	}
	if sessions[0].Application != application.Name || len(sessions[0].SessionId) != 1 || sessions[0].SessionId[0] != beegoSessionId {
		t.Fatalf("unexpected session row: %+v", sessions[0])
	}
	if len(sessions[0].SessionInfos) != 1 {
		t.Fatalf("session infos = %d, want 1", len(sessions[0].SessionInfos))
	}
	info := sessions[0].SessionInfos[0]
	if info.SessionId != beegoSessionId || info.Ip != "203.0.113.7" || info.UserAgent != "Mozilla/5.0 (cas6-test)" {
		t.Fatalf("unexpected session info: %+v", info)
	}
	expireTime, err := time.Parse(time.RFC3339, info.ExpireTime)
	if err != nil {
		t.Fatalf("expireTime %q is not RFC3339: %v", info.ExpireTime, err)
	}
	if remaining := time.Until(expireTime); remaining < 719*time.Hour || remaining > 721*time.Hour {
		t.Fatalf("session expire should follow refreshExpireInHours (720h), got %v", remaining)
	}

	// refresh: sid наследуется и в строке token, и в JWT
	refreshed, err := RefreshToken(application, "refresh_token", token.RefreshToken, "", application.ClientId, application.ClientSecret, "", "localhost", "")
	if err != nil {
		t.Fatal(err)
	}
	wrapper, ok := refreshed.(*TokenWrapper)
	if !ok {
		t.Fatalf("refresh failed: %+v", refreshed)
	}
	newClaims, err := ParseJwtToken(wrapper.AccessToken, cert)
	if err != nil {
		t.Fatal(err)
	}
	if newClaims.Sid != beegoSessionId {
		t.Fatalf("refreshed access token sid = %q, want %q", newClaims.Sid, beegoSessionId)
	}
	newToken, err := GetTokenByRefreshToken(wrapper.RefreshToken)
	if err != nil || newToken == nil {
		t.Fatalf("refreshed token not found: %v", err)
	}
	if newToken.SessionId != beegoSessionId {
		t.Fatalf("refreshed Token.SessionId = %q, want %q", newToken.SessionId, beegoSessionId)
	}

	// delete-session?sessionId= → токены с этим session_id получают expires_in=0 → refresh падает
	affected, err := DeleteSessionId(util.GetSessionId("clients", user.Name, application.Name), beegoSessionId)
	if err != nil {
		t.Fatal(err)
	}
	if !affected {
		t.Fatal("DeleteSessionId affected nothing")
	}
	sessions, err = GetUserSessions("clients", user.Name)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Fatalf("session rows after delete = %d, want 0", len(sessions))
	}

	refreshed, err = RefreshToken(application, "refresh_token", wrapper.RefreshToken, "", application.ClientId, application.ClientSecret, "", "localhost", "")
	if err != nil {
		t.Fatal(err)
	}
	tokenError, ok = refreshed.(*TokenError)
	if !ok {
		t.Fatalf("refresh after delete-session should fail, got %+v", refreshed)
	}
	if tokenError.Error != InvalidGrant || tokenError.ErrorDescription != "refresh token is expired" {
		t.Fatalf("unexpected refresh error: %s: %s", tokenError.Error, tokenError.ErrorDescription)
	}
}

// TestPasswordGrantWithoutSessionKeepsOldBehaviour — без Beego-сессии (session=nil) ничего не меняется:
// Session-строки нет, sid в токене нет.
func TestPasswordGrantWithoutSessionKeepsOldBehaviour(t *testing.T) {
	initPasswordGrantTestDb(t)
	application, user := addPasswordGrantFixtures(t)

	token, tokenError, err := GetPasswordToken(application, user.Name, "correct horse", "", "localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	if tokenError != nil {
		t.Fatalf("password grant failed: %s: %s", tokenError.Error, tokenError.ErrorDescription)
	}
	if token.SessionId != "" {
		t.Fatalf("Token.SessionId = %q, want empty", token.SessionId)
	}
	cert, err := getCert("admin", "cert-cas6")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ParseJwtToken(token.AccessToken, cert)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Sid != "" {
		t.Fatalf("sid = %q, want empty", claims.Sid)
	}
	sessions, err := GetUserSessions("clients", user.Name)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Fatalf("session rows = %d, want 0", len(sessions))
	}
}

// TestGetClaimsCustomSid — в JWT-Custom клейм sid кладётся только при непустом SessionId.
func TestGetClaimsCustomSid(t *testing.T) {
	user := &User{Owner: "clients", Name: "alice", Properties: map[string]string{}}
	claims := Claims{User: user, TokenType: "access-token", Sid: "sid-1"}
	res := getClaimsCustom(claims, []string{"Name"}, nil)
	if res["sid"] != "sid-1" {
		t.Fatalf("sid claim = %v, want sid-1", res["sid"])
	}
	claims.Sid = ""
	res = getClaimsCustom(claims, []string{"Name"}, nil)
	if _, ok := res["sid"]; ok {
		t.Fatal("empty sid must not be put into the claims")
	}
}
