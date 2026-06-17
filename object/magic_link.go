package object

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const (
	MagicLinkStatusCreated = "created"
	MagicLinkStatusSent    = "sent"
	MagicLinkStatusOpened  = "opened"
	MagicLinkStatusUsed    = "used"
	MagicLinkStatusExpired = "expired"
	MagicLinkStatusFailed  = "failed"
	MagicLinkStatusRevoked = "revoked"

	MagicLinkAuthActionSignupNewUser      = "signup_new_user"
	MagicLinkAuthActionSigninExistingUser = "signin_existing_user"

	MagicLinkDefaultExpireMinutes = 10
	MagicLinkMinExpireMinutes     = 2
	MagicLinkMaxExpireMinutes     = 43200
)

type MagicLink struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Application string `xorm:"varchar(100) index" json:"application"`
	Permission  string `xorm:"varchar(200)" json:"permission"`
	Group       string `xorm:"varchar(100)" json:"group"`
	Email       string `xorm:"varchar(100) index" json:"email"`
	Requester   string `xorm:"varchar(100)" json:"requester"`
	RemoteAddr  string `xorm:"varchar(100)" json:"remoteAddr"`
	AuthAction  string `xorm:"varchar(100)" json:"authAction"`
	TokenHash   string `xorm:"varchar(100) index" json:"-"`
	Status      string `xorm:"varchar(20) index" json:"status"`
	ExpiryTime  string `xorm:"varchar(100)" json:"expiryTime"`
	ExpireAt    int64  `xorm:"index" json:"expireAt"`
	OpenedTime  string `xorm:"varchar(100)" json:"openedTime"`
	UsedTime    string `xorm:"varchar(100)" json:"usedTime"`
	LastError   string `xorm:"varchar(500)" json:"lastError"`

	ClientId            string `xorm:"varchar(100)" json:"clientId"`
	ResponseType        string `xorm:"varchar(100)" json:"responseType"`
	RedirectUri         string `xorm:"varchar(500)" json:"redirectUri"`
	Scope               string `xorm:"varchar(1000)" json:"scope"`
	State               string `xorm:"varchar(1000)" json:"state"`
	Nonce               string `xorm:"varchar(1000)" json:"nonce"`
	CodeChallengeMethod string `xorm:"varchar(100)" json:"codeChallengeMethod"`
	CodeChallenge       string `xorm:"varchar(500)" json:"codeChallenge"`
	Resource            string `xorm:"varchar(1000)" json:"resource"`

	SubUsers   []string `xorm:"mediumtext" json:"subUsers"`
	SubGroups  []string `xorm:"mediumtext" json:"subGroups"`
	SubRoles   []string `xorm:"mediumtext" json:"subRoles"`
	SubDomains []string `xorm:"mediumtext" json:"subDomains"`
	Resources  []string `xorm:"mediumtext" json:"resources"`
	Actions    []string `xorm:"mediumtext" json:"actions"`
}

func GenerateMagicLinkToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func HashMagicLinkToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	res := hex.EncodeToString(hash[:])
	if len(res) > 64 {
		return res[:64]
	}
	return res
}

func GetMagicLinkCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&MagicLink{Owner: owner})
}

func GetMagicLinks(owner string) ([]*MagicLink, error) {
	links := []*MagicLink{}
	err := ormer.Engine.Desc("created_time").Find(&links, &MagicLink{Owner: owner})
	return links, err
}

func GetPaginationMagicLinks(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*MagicLink, error) {
	links := []*MagicLink{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&links, &MagicLink{Owner: owner})
	return links, err
}

func GetMagicLink(id string) (*MagicLink, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	link := MagicLink{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&link)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}
	return &link, nil
}

func AddMagicLink(link *MagicLink) error {
	_, err := ormer.Engine.Insert(link)
	return err
}

func UpdateMagicLinkStatus(link *MagicLink, status string, lastError string) error {
	link.Status = status
	link.LastError = lastError
	cols := []string{"status", "last_error"}
	now := util.GetCurrentTime()
	if status == MagicLinkStatusSent {
		cols = append(cols, "last_error")
	}
	if status == MagicLinkStatusOpened && link.OpenedTime == "" {
		link.OpenedTime = now
		cols = append(cols, "opened_time")
	}
	if status == MagicLinkStatusUsed && link.UsedTime == "" {
		link.UsedTime = now
		cols = append(cols, "used_time")
	}
	_, err := ormer.Engine.ID(core.PK{link.Owner, link.Name}).Cols(cols...).Update(link)
	return err
}

func RevokeMagicLink(id string) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	link := &MagicLink{Status: MagicLinkStatusRevoked}
	affected, err := ormer.Engine.ID(core.PK{owner, name}).In("status", MagicLinkStatusCreated, MagicLinkStatusSent, MagicLinkStatusOpened).Cols("status").Update(link)
	return affected != 0, err
}

func DeleteMagicLink(id string) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	affected, err := ormer.Engine.ID(core.PK{owner, name}).Delete(&MagicLink{})
	return affected != 0, err
}

func GetMagicLinkRateLimitCounts(email string, remoteAddr string, application *Application) (int64, int64, int64, error) {
	windowMinutes := application.GetMagicLinkRateLimitWindowMinutes()
	expireMinutes := application.GetMagicLinkExpireMinutes()
	thresholdExpireAt := time.Now().Add(time.Duration(expireMinutes-windowMinutes) * time.Minute).Unix()
	applicationID := application.GetId()
	emailCount, err := ormer.Engine.Where("application = ?", applicationID).And("email = ?", email).And("expire_at > ?", thresholdExpireAt).Count(&MagicLink{})
	if err != nil {
		return 0, 0, 0, err
	}
	ipCount, err := ormer.Engine.Where("application = ?", applicationID).And("remote_addr = ?", remoteAddr).And("expire_at > ?", thresholdExpireAt).Count(&MagicLink{})
	if err != nil {
		return 0, 0, 0, err
	}
	applicationCount, err := ormer.Engine.Where("application = ?", applicationID).And("expire_at > ?", thresholdExpireAt).Count(&MagicLink{})
	if err != nil {
		return 0, 0, 0, err
	}
	return emailCount, ipCount, applicationCount, nil
}

func IsMagicLinkCaptchaRequired(email string, remoteAddr string, application *Application) (bool, error) {
	emailCount, ipCount, _, err := GetMagicLinkRateLimitCounts(email, remoteAddr, application)
	if err != nil {
		return false, err
	}
	threshold := int64(application.GetMagicLinkCaptchaThreshold())
	return emailCount >= threshold || ipCount >= threshold, nil
}

func IsMagicLinkAllowSend(email string, remoteAddr string, application *Application) error {
	emailCount, ipCount, applicationCount, err := GetMagicLinkRateLimitCounts(email, remoteAddr, application)
	if err != nil {
		return err
	}
	if emailCount >= int64(application.GetMagicLinkRateLimitEmail()) {
		return fmt.Errorf("too many magic links requested for this email")
	}
	if ipCount >= int64(application.GetMagicLinkRateLimitIP()) {
		return fmt.Errorf("too many magic links requested from this IP")
	}
	if applicationCount >= int64(application.GetMagicLinkRateLimitApplication()) {
		return fmt.Errorf("too many magic links requested for this application")
	}
	return nil
}

func ResolveMagicLinkExpireTime(application *Application, expiresInMinutes int, expireTime string, now time.Time) (time.Time, error) {
	if now.IsZero() {
		now = time.Now()
	}
	minDuration := time.Duration(MagicLinkMinExpireMinutes) * time.Minute
	maxDuration := time.Duration(MagicLinkMaxExpireMinutes) * time.Minute
	if expireTime != "" {
		parsed, err := time.Parse(time.RFC3339, expireTime)
		if err != nil {
			return time.Time{}, err
		}
		if !parsed.After(now) {
			return time.Time{}, fmt.Errorf("magic link expiry time must be in the future")
		}
		ttl := parsed.Sub(now)
		if ttl < minDuration || ttl > maxDuration {
			return time.Time{}, fmt.Errorf("magic link expiry must be between %d and %d minutes", MagicLinkMinExpireMinutes, MagicLinkMaxExpireMinutes)
		}
		return parsed, nil
	}
	if expiresInMinutes <= 0 {
		expiresInMinutes = application.GetMagicLinkExpireMinutes()
	}
	if expiresInMinutes < MagicLinkMinExpireMinutes || expiresInMinutes > MagicLinkMaxExpireMinutes {
		return time.Time{}, fmt.Errorf("magic link expiry must be between %d and %d minutes", MagicLinkMinExpireMinutes, MagicLinkMaxExpireMinutes)
	}
	return now.Add(time.Duration(expiresInMinutes) * time.Minute), nil
}

func ValidateMagicLinkConfig(application *Application) error {
	if application == nil {
		return nil
	}
	if application.EnableMagicLinkSignup && !application.MagicLinkSigninEnabled {
		return fmt.Errorf("enableMagicLinkSignup requires magicLinkSigninEnabled")
	}
	if application.MagicLinkExpireMinutes != 0 && (application.MagicLinkExpireMinutes < MagicLinkMinExpireMinutes || application.MagicLinkExpireMinutes > MagicLinkMaxExpireMinutes) {
		return fmt.Errorf("magic link expiry must be between %d and %d minutes", MagicLinkMinExpireMinutes, MagicLinkMaxExpireMinutes)
	}
	if application.MagicLinkRateLimitWindowMinutes < 0 {
		return fmt.Errorf("magicLinkRateLimitWindowMinutes must be greater than or equal to 0")
	}
	if application.MagicLinkRateLimitEmail < 0 {
		return fmt.Errorf("magicLinkRateLimitEmail must be greater than or equal to 0")
	}
	if application.MagicLinkRateLimitIP < 0 {
		return fmt.Errorf("magicLinkRateLimitIp must be greater than or equal to 0")
	}
	if application.MagicLinkRateLimitApplication < 0 {
		return fmt.Errorf("magicLinkRateLimitApplication must be greater than or equal to 0")
	}
	if application.MagicLinkCaptchaThreshold < 0 {
		return fmt.Errorf("magicLinkCaptchaThreshold must be greater than or equal to 0")
	}
	return nil
}

func NewMagicLink(application *Application, permission *Permission, email string, remoteAddr string, requester string, token string, oauth map[string]string, expireAt time.Time) *MagicLink {
	if expireAt.IsZero() {
		expireAt = time.Now().Add(time.Duration(application.GetMagicLinkExpireMinutes()) * time.Minute)
	}
	link := &MagicLink{
		Owner:               application.Organization,
		Name:                util.GenerateId(),
		CreatedTime:         util.GetCurrentTime(),
		Application:         application.GetId(),
		Email:               email,
		Requester:           requester,
		RemoteAddr:          remoteAddr,
		Status:              MagicLinkStatusCreated,
		ExpiryTime:          util.Time2String(expireAt),
		ExpireAt:            expireAt.Unix(),
		ClientId:            oauth["clientId"],
		ResponseType:        oauth["responseType"],
		RedirectUri:         oauth["redirectUri"],
		Scope:               oauth["scope"],
		State:               oauth["state"],
		Nonce:               oauth["nonce"],
		CodeChallengeMethod: oauth["codeChallengeMethod"],
		CodeChallenge:       oauth["codeChallenge"],
		Resource:            oauth["resource"],
	}
	if link.ResponseType == "" {
		link.ResponseType = "login"
	}
	if token != "" {
		link.TokenHash = HashMagicLinkToken(token)
	}
	ApplyPermissionSnapshotToMagicLink(link, permission)
	return link
}

func ApplyPermissionSnapshotToMagicLink(link *MagicLink, permission *Permission) {
	if permission == nil {
		return
	}
	link.Permission = permission.GetId()
	link.SubUsers = append([]string{}, permission.Users...)
	link.SubGroups = append([]string{}, permission.Groups...)
	link.SubRoles = append([]string{}, permission.Roles...)
	link.SubDomains = append([]string{}, permission.Domains...)
	link.Resources = append([]string{}, permission.Resources...)
	link.Actions = append([]string{}, permission.Actions...)
}

func UpdateMagicLinkPermissionSnapshot(link *MagicLink) error {
	_, err := ormer.Engine.ID(core.PK{link.Owner, link.Name}).Cols("permission", "sub_users", "sub_groups", "sub_roles", "sub_domains", "resources", "actions").Update(link)
	return err
}

func BuildMagicLinkCallbackURL(link *MagicLink, token string, host string) string {
	originFrontend, _ := getOriginFromHost(host)
	query := url.Values{}
	query.Set("token", token)
	if link.ClientId != "" {
		query.Set("client_id", link.ClientId)
	}
	if link.ResponseType != "" {
		query.Set("response_type", link.ResponseType)
	}
	if link.RedirectUri != "" {
		query.Set("redirect_uri", link.RedirectUri)
	}
	if link.Scope != "" {
		query.Set("scope", link.Scope)
	}
	if link.State != "" {
		query.Set("state", link.State)
	}
	if link.Nonce != "" {
		query.Set("nonce", link.Nonce)
	}
	if link.CodeChallengeMethod != "" {
		query.Set("code_challenge_method", link.CodeChallengeMethod)
	}
	if link.CodeChallenge != "" {
		query.Set("code_challenge", link.CodeChallenge)
	}
	if link.Resource != "" {
		query.Set("resource", link.Resource)
	} else if len(link.Resources) > 0 {
		query.Set("resource", link.Resources[0])
	}
	return fmt.Sprintf("%s/magic-link/callback?%s", strings.TrimRight(originFrontend, "/"), query.Encode())
}

func SendMagicLinkToEmail(organization *Organization, provider *Provider, email string, magicLinkURL string, expireTime string, authAction string) error {
	sender := organization.DisplayName
	title := provider.Title
	content := ""
	if authAction == MagicLinkAuthActionSignupNewUser {
		content = provider.MagicLinkSignupContent
		if content == "" {
			content = GetDefaultMagicLinkSignupEmailContent(magicLinkURL, expireTime)
		}
	} else {
		if provider.MagicLinkContent != "" {
			content = provider.MagicLinkContent
		}
		if content == "" {
			content = GetDefaultMagicLinkEmailContent(magicLinkURL, expireTime)
		}
	}
	if strings.Contains(content, "%link") {
		content = strings.ReplaceAll(content, "%link", magicLinkURL)
	}
	content = strings.ReplaceAll(content, "%expireTime", expireTime)
	if title == "" {
		title = "Magic Link"
	}
	return SendEmail(provider, title, content, []string{email}, sender)
}

func GetDefaultMagicLinkEmailContent(magicLinkURL string, expireTime string) string {
	staticBaseURL := strings.TrimRight(conf.GetConfigString("staticBaseUrl"), "/")
	if staticBaseURL == "" {
		staticBaseURL = "https://cdn.casbin.org"
	}
	logoURL := fmt.Sprintf("%s/img/casdoor-logo_1185x256.png", staticBaseURL)
	content := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Magic Link</title>
  <style>
    body {
      margin: 0;
      padding: 0;
      font-family: 'Inter', Arial, sans-serif;
      background: #ffffff;
      color: #1a1a1a;
    }

    .container {
      max-width: 600px;
      margin: 0 auto;
      padding: 24px;
    }

    .logo {
      text-align: center;
      margin-bottom: 24px;
    }

    .logo img {
      max-width: 260px;
      height: auto;
    }

    .greeting {
      font-size: 18px;
      font-weight: 600;
      margin-bottom: 12px;
      text-align: center;
    }

    .message {
      font-size: 14px;
      margin-bottom: 20px;
      color: #333;
      line-height: 1.5;
      text-align: center;
    }

    .button-box {
      text-align: center;
      margin: 28px 0;
    }

    .button-box a {
      display: inline-block;
      background: #6366f1;
      color: #ffffff;
      font-size: 15px;
      font-weight: 700;
      text-decoration: none;
      padding: 14px 28px;
      border-radius: 8px;
    }

    .link-box {
      text-align: center;
      margin-top: 12px;
      font-size: 14px;
      word-break: break-all;
    }

    .link-box a {
      color: #6366f1;
      font-weight: 600;
      text-decoration: none;
    }

    .footer {
      font-size: 12px;
      text-align: center;
      color: #777;
      margin-top: 36px;
      line-height: 1.4;
    }

    .footer a {
      color: #6366f1;
      text-decoration: none;
      font-weight: 600;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="logo">
      <img src="%logoUrl" alt="Logo">
    </div>

    <div class="greeting">
      Sign in to your account
    </div>

    <div class="message">
      Use the secure Magic Link below to complete your sign in.
    </div>

    <div class="button-box">
      <a href="%link">Sign in with Magic Link</a>
    </div>

    <div class="link-box">
      Or open this <a href="%link">link</a>
    </div>

    <div class="message" style="margin-top: 24px;">
      This link can be used once and will expire at %expireTime.<br>
      Thanks,<br>
      Casdoor Team
    </div>

    <div class="footer">
      Need help? Please contact your administrator.
    </div>
  </div>
</body>
</html>`
	content = strings.ReplaceAll(content, "%logoUrl", logoURL)
	content = strings.ReplaceAll(content, "%link", magicLinkURL)
	return strings.ReplaceAll(content, "%expireTime", expireTime)
}

func GetDefaultMagicLinkSignupEmailContent(magicLinkURL string, expireTime string) string {
	staticBaseURL := strings.TrimRight(conf.GetConfigString("staticBaseUrl"), "/")
	if staticBaseURL == "" {
		staticBaseURL = "https://cdn.casbin.org"
	}
	logoURL := fmt.Sprintf("%s/img/casdoor-logo_1185x256.png", staticBaseURL)
	content := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Magic Link Sign Up</title>
  <style>
    body {
      margin: 0;
      padding: 0;
      font-family: 'Inter', Arial, sans-serif;
      background: #ffffff;
      color: #1a1a1a;
    }

    .container {
      max-width: 600px;
      margin: 0 auto;
      padding: 24px;
    }

    .logo {
      text-align: center;
      margin-bottom: 24px;
    }

    .logo img {
      max-width: 260px;
      height: auto;
    }

    .greeting {
      font-size: 18px;
      font-weight: 600;
      margin-bottom: 12px;
      text-align: center;
    }

    .message {
      font-size: 14px;
      margin-bottom: 20px;
      color: #333;
      line-height: 1.5;
      text-align: center;
    }

    .button-box {
      text-align: center;
      margin: 28px 0;
    }

    .button-box a {
      display: inline-block;
      background: #6366f1;
      color: #ffffff;
      font-size: 15px;
      font-weight: 700;
      text-decoration: none;
      padding: 14px 28px;
      border-radius: 8px;
    }

    .link-box {
      text-align: center;
      margin-top: 12px;
      font-size: 14px;
      word-break: break-all;
    }

    .link-box a {
      color: #6366f1;
      font-weight: 600;
      text-decoration: none;
    }

    .footer {
      font-size: 12px;
      text-align: center;
      color: #777;
      margin-top: 36px;
      line-height: 1.4;
    }

    .footer a {
      color: #6366f1;
      text-decoration: none;
      font-weight: 600;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="logo">
      <img src="%logoUrl" alt="Logo">
    </div>

    <div class="greeting">
      Create your account
    </div>

    <div class="message">
      Use the secure Magic Link below to create your account.
    </div>

    <div class="button-box">
      <a href="%link">Create account with Magic Link</a>
    </div>

    <div class="link-box">
      Or open this <a href="%link">link</a>
    </div>

    <div class="message" style="margin-top: 24px;">
      This link can be used once and will expire at %expireTime.<br>
      Thanks,<br>
      Casdoor Team
    </div>

    <div class="footer">
      Need help? Please contact your administrator.
    </div>
  </div>
</body>
</html>`
	content = strings.ReplaceAll(content, "%logoUrl", logoURL)
	content = strings.ReplaceAll(content, "%link", magicLinkURL)
	return strings.ReplaceAll(content, "%expireTime", expireTime)
}

func ResolveMagicLinkPermission(application *Application, user *User, requestedPermission string) (*Permission, error) {
	owner := application.Organization
	if application.IsShared && user != nil {
		owner = user.Owner
	}
	if requestedPermission != "" {
		permissionOwner, permissionName := util.GetOwnerAndNameFromIdNoCheck(requestedPermission)
		if permissionName == "" {
			permissionOwner = owner
			permissionName = requestedPermission
		}
		if permissionOwner != owner {
			return nil, fmt.Errorf("permission company mismatch")
		}
		permission, err := getPermission(permissionOwner, permissionName)
		if err != nil {
			return nil, err
		}
		if permission == nil {
			return nil, fmt.Errorf("permission does not exist")
		}
		return validateMagicLinkPermission(application, user, permission)
	}

	permissions, err := GetPermissions(owner)
	if err != nil {
		return nil, err
	}
	for _, permission := range permissions {
		if _, err = validateMagicLinkPermission(application, user, permission); err == nil {
			return permission, nil
		}
	}
	if user != nil {
		ok, err := CheckLoginPermission(user.GetId(), application)
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, nil
		}
	}
	return nil, fmt.Errorf("magic link login permission denied")
}

func validateMagicLinkPermission(application *Application, user *User, permission *Permission) (*Permission, error) {
	if permission == nil || !permission.IsEnabled || permission.State != "Approved" || permission.ResourceType != "Application" || !permission.isResourceHit(application.Name) || permission.Effect != "Allow" {
		return nil, fmt.Errorf("permission does not allow this application")
	}
	if user == nil {
		return permission, nil
	}
	userID := user.GetId()
	userHit := permission.isUserHit(userID)
	groupHit := permission.isGroupHit(userID)
	roleHit := permission.isRoleHit(userID)
	if !userHit && !groupHit && !roleHit {
		return nil, fmt.Errorf("permission does not allow this user")
	}
	enforcer, err := getPermissionEnforcer(permission)
	if err != nil {
		return nil, err
	}
	ok, err := enforcer.Enforce(userID, application.Name, "Read")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("permission does not allow this action")
	}
	return permission, nil
}

func ConsumeMagicLink(token string) (*MagicLink, error) {
	tokenHash := HashMagicLinkToken(token)
	link := &MagicLink{TokenHash: tokenHash}
	existed, err := ormer.Engine.Get(link)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, fmt.Errorf("magic link is invalid")
	}
	nowUnix := time.Now().Unix()
	err = validateMagicLinkVerifyState(link, nowUnix)
	if err != nil {
		if err.Error() != "magic link has expired" {
			return nil, err
		}
		link.Status = MagicLinkStatusExpired
		_, _ = ormer.Engine.ID(core.PK{link.Owner, link.Name}).Cols("status").Update(link)
		return nil, err
	}
	now := util.GetCurrentTime()
	claimed := &MagicLink{
		Status:     MagicLinkStatusUsed,
		OpenedTime: now,
		UsedTime:   now,
	}
	affected, err := ormer.Engine.Where("token_hash = ?", tokenHash).And("expire_at > ?", nowUnix).In("status", MagicLinkStatusCreated, MagicLinkStatusSent, MagicLinkStatusOpened).Cols("status", "opened_time", "used_time").Update(claimed)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, fmt.Errorf("magic link was already used")
	}
	link.Status = MagicLinkStatusUsed
	link.OpenedTime = now
	link.UsedTime = now
	return link, nil
}

func validateMagicLinkVerifyState(link *MagicLink, nowUnix int64) error {
	if link.Status == MagicLinkStatusRevoked {
		return fmt.Errorf("magic link was revoked")
	}
	if link.Status == MagicLinkStatusUsed {
		return fmt.Errorf("magic link was already used")
	}
	if link.Status == MagicLinkStatusExpired || link.ExpireAt <= nowUnix {
		return fmt.Errorf("magic link has expired")
	}
	if !isMagicLinkVerifiableStatus(link.Status) {
		return fmt.Errorf("magic link was already used")
	}
	return nil
}

func isMagicLinkVerifiableStatus(status string) bool {
	return status == MagicLinkStatusCreated || status == MagicLinkStatusSent || status == MagicLinkStatusOpened
}
