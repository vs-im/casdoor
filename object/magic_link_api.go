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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// The headless magic link API (/api/send-magic-link, /api/verify-magic-link) is a layer
// on top of the built-in magic link sign-in of magic_link.go: the token, its hash, the
// origin check, the mail template substitution and the atomic claim are the built-in
// ones, this file only adds what a server-to-server caller needs: the OAuth context of
// the link, a TTL and rate limits configured per application, a status for the admin
// list and an optional binding instead of the mandatory browser session.

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

	// MagicLinkBindingSession is a link of the built-in sign-in page: it only works in the
	// browser session that asked for it. It is the zero value, the built-in flow does not
	// know about the column.
	MagicLinkBindingSession = ""
	// MagicLinkBindingNone is a link of the API that may be opened on any device.
	MagicLinkBindingNone = "none"
	// MagicLinkBindingClient is a link of the API bound to a secret the caller keeps, for
	// example in a httpOnly cookie of its own sign-in page.
	MagicLinkBindingClient = "client"

	magicLinkUnboundPrefix = "unbound:"
)

// MagicLinkExtension holds the columns the fork adds to the "magic_link" table. It is
// embedded into MagicLink, so the JSON and the table stay flat.
type MagicLinkExtension struct {
	Permission string `xorm:"varchar(200)" json:"permission"`
	Group      string `xorm:"varchar(100)" json:"group"`
	Requester  string `xorm:"varchar(100)" json:"requester"`
	AuthAction string `xorm:"varchar(100)" json:"authAction"`
	Binding    string `xorm:"varchar(20)" json:"binding"`
	Status     string `xorm:"varchar(20) index" json:"status"`
	ExpiryTime string `xorm:"varchar(100)" json:"expiryTime"`
	ExpireAt   int64  `xorm:"index" json:"expireAt"`
	OpenedTime string `xorm:"varchar(100)" json:"openedTime"`
	UsedTime   string `xorm:"varchar(100)" json:"usedTime"`
	LastError  string `xorm:"varchar(500)" json:"lastError"`

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

// isMagicLinkExpired is called by ConsumeMagicLink(). A link of the API expires at its
// own ExpireAt, a link of the built-in sign-in page after the verification code timeout.
func isMagicLinkExpired(magicLink *MagicLink, now int64) bool {
	if magicLink.ExpireAt > 0 {
		return now >= magicLink.ExpireAt
	}
	return now-magicLink.Time > getMagicLinkTimeout()*60
}

// getUnboundMagicLinkSessionHash is what a link that may be opened anywhere stores in
// place of a browser session: a value that only follows from the token itself, no
// browser session of the built-in sign-in page ever hashes to it.
func getUnboundMagicLinkSessionHash(tokenHash string) string {
	return HashMagicLinkSecret(magicLinkUnboundPrefix + tokenHash)
}

func (application *Application) GetMagicLinkExpireMinutes() int {
	if application == nil || application.MagicLinkExpireMinutes <= 0 {
		return MagicLinkDefaultExpireMinutes
	}
	return application.MagicLinkExpireMinutes
}

func (application *Application) GetMagicLinkRateLimitWindowMinutes() int {
	if application == nil || application.MagicLinkRateLimitWindowMinutes <= 0 {
		return 15
	}
	return application.MagicLinkRateLimitWindowMinutes
}

func (application *Application) GetMagicLinkRateLimitEmail() int {
	if application == nil || application.MagicLinkRateLimitEmail <= 0 {
		return 3
	}
	return application.MagicLinkRateLimitEmail
}

func (application *Application) GetMagicLinkRateLimitIP() int {
	if application == nil || application.MagicLinkRateLimitIP <= 0 {
		return 10
	}
	return application.MagicLinkRateLimitIP
}

func (application *Application) GetMagicLinkRateLimitApplication() int {
	if application == nil || application.MagicLinkRateLimitApplication <= 0 {
		return 100
	}
	return application.MagicLinkRateLimitApplication
}

func (application *Application) GetMagicLinkCaptchaThreshold() int {
	if application == nil || application.MagicLinkCaptchaThreshold <= 0 {
		return 1
	}
	return application.MagicLinkCaptchaThreshold
}

// IsMagicLinkApiSignupEnabled tells whether a link of the API may create the account.
// Besides the built-in rule ("Sign in or sign up" plus the application's own signup) the
// API keeps its "enableMagicLinkSignup" switch: it lets an application sign users up by
// a link only, without opening the password signup page.
func (application *Application) IsMagicLinkApiSignupEnabled() bool {
	if application == nil || !application.IsMagicLinkEnabled() {
		return false
	}
	return application.EnableMagicLinkSignup || application.IsMagicLinkSignupEnabled()
}

// GetMagicLinkSignupApplication is the application as the built-in signup has to see it
// when the signup was allowed by the "enableMagicLinkSignup" switch: CheckMagicLinkSignup()
// and the built-in user creation are reused as they are. The switch signs up by a link
// alone, for an application whose signup page is closed, so the items of that page that a
// link cannot fill in (a new application gets a required "Phone" by default) are left out.
// The ones CheckMagicLinkSignup() accepts stay, so a required "Invitation code" still holds.
func (application *Application) GetMagicLinkSignupApplication() *Application {
	if application.IsMagicLinkSignupEnabled() || !application.IsMagicLinkApiSignupEnabled() {
		return application
	}

	res := *application
	res.EnableSignUp = true
	res.SignupItems = make([]*SignupItem, 0, len(application.SignupItems))
	for _, signupItem := range application.SignupItems {
		if signupItem != nil && isMagicLinkSignupItem(signupItem.Name) {
			res.SignupItems = append(res.SignupItems, signupItem)
		}
	}
	res.SigninMethods = make([]*SigninMethod, 0, len(application.SigninMethods))
	for _, signinMethod := range application.SigninMethods {
		if signinMethod != nil && signinMethod.Name == "Magic link" {
			method := *signinMethod
			method.Rule = SigninMethodRuleMagicLinkSignup
			signinMethod = &method
		}
		res.SigninMethods = append(res.SigninMethods, signinMethod)
	}
	return &res
}

func isMagicLinkSignupItem(name string) bool {
	switch name {
	case "ID", "Username", "Display name", "Email", "Password", "Confirm password", "Agreement", "Invitation code", "Signup button", "Providers":
		return true
	}
	return false
}

func ValidateMagicLinkConfig(application *Application) error {
	if application == nil {
		return nil
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

// GetMagicLinkRateLimitCounts counts the links issued within the application's window by
// their issue time, so the links of the built-in sign-in page count as well.
func GetMagicLinkRateLimitCounts(email string, remoteAddr string, application *Application) (int64, int64, int64, error) {
	since := time.Now().Add(-time.Duration(application.GetMagicLinkRateLimitWindowMinutes()) * time.Minute).Unix()

	emailCount, err := ormer.Engine.Where("owner = ?", application.Organization).And("application = ?", application.Name).And("email = ?", email).And("time > ?", since).Count(&MagicLink{})
	if err != nil {
		return 0, 0, 0, err
	}
	ipCount, err := ormer.Engine.Where("owner = ?", application.Organization).And("application = ?", application.Name).And("remote_addr = ?", remoteAddr).And("time > ?", since).Count(&MagicLink{})
	if err != nil {
		return 0, 0, 0, err
	}
	applicationCount, err := ormer.Engine.Where("owner = ?", application.Organization).And("application = ?", application.Name).And("time > ?", since).Count(&MagicLink{})
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

func checkMagicLinkRateLimit(emailCount int64, ipCount int64, applicationCount int64, application *Application) error {
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

func IsMagicLinkAllowSend(email string, remoteAddr string, application *Application) error {
	emailCount, ipCount, applicationCount, err := GetMagicLinkRateLimitCounts(email, remoteAddr, application)
	if err != nil {
		return err
	}
	return checkMagicLinkRateLimit(emailCount, ipCount, applicationCount, application)
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

// NewMagicLink builds the row of a link of the API. An empty token is a request that was
// turned down before a link existed, such a row is kept for the admin list only.
func NewMagicLink(application *Application, permission *Permission, email string, remoteAddr string, requester string, token string, oauth map[string]string, expireAt time.Time) *MagicLink {
	if expireAt.IsZero() {
		expireAt = time.Now().Add(time.Duration(application.GetMagicLinkExpireMinutes()) * time.Minute)
	}
	link := &MagicLink{
		Owner:       application.Organization,
		Name:        util.GenerateId(),
		CreatedTime: util.GetCurrentTime(),
		Application: application.Name,
		Email:       email,
		RemoteAddr:  remoteAddr,
		Time:        time.Now().Unix(),
		IsUsed:      token == "",
		MagicLinkExtension: MagicLinkExtension{
			Requester:           requester,
			Binding:             MagicLinkBindingNone,
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
		},
	}
	if link.ResponseType == "" {
		link.ResponseType = "login"
	}
	if token != "" {
		link.TokenHash = HashMagicLinkSecret(token)
		link.SessionHash = getUnboundMagicLinkSessionHash(link.TokenHash)
	}
	ApplyPermissionSnapshotToMagicLink(link, permission)
	return link
}

// BindMagicLinkToClient ties a link of the API to a secret of the caller, the same way
// the built-in sign-in page ties its links to the browser session.
func BindMagicLinkToClient(link *MagicLink, sessionSecret string) {
	if sessionSecret == "" || link.TokenHash == "" {
		return
	}
	link.Binding = MagicLinkBindingClient
	link.SessionHash = HashMagicLinkSecret(sessionSecret)
}

// NewApiMagicLink issues the token of a link of the API and builds its row.
func NewApiMagicLink(application *Application, permission *Permission, email string, remoteAddr string, requester string, oauth map[string]string, expireAt time.Time) (string, *MagicLink, error) {
	token, err := generateMagicLinkToken()
	if err != nil {
		return "", nil, err
	}
	return token, NewMagicLink(application, permission, email, remoteAddr, requester, token, oauth, expireAt), nil
}

// CheckMagicLinkOrigin tells whether a link can be issued for a request to this host at
// all, see getMagicLinkOrigin().
func CheckMagicLinkOrigin(host string, lang string) error {
	_, err := getMagicLinkOrigin(host, lang)
	return err
}

func AddMagicLink(link *MagicLink) error {
	_, err := ormer.Engine.Insert(link)
	return err
}

// AddFailedMagicLink keeps a request that was answered without a link being mailed.
func AddFailedMagicLink(link *MagicLink, lastError string) error {
	link.TokenHash = ""
	link.SessionHash = ""
	link.IsUsed = true
	link.Status = MagicLinkStatusFailed
	link.LastError = truncateMagicLinkError(lastError)
	return AddMagicLink(link)
}

func truncateMagicLinkError(lastError string) string {
	if len(lastError) > 500 {
		return lastError[:500]
	}
	return lastError
}

func UpdateMagicLinkStatus(link *MagicLink, status string, lastError string) error {
	link.Status = status
	link.LastError = truncateMagicLinkError(lastError)
	cols := []string{"status", "last_error"}
	now := util.GetCurrentTime()
	if status == MagicLinkStatusOpened && link.OpenedTime == "" {
		link.OpenedTime = now
		cols = append(cols, "opened_time")
	}
	if status == MagicLinkStatusUsed {
		if link.OpenedTime == "" {
			link.OpenedTime = now
			cols = append(cols, "opened_time")
		}
		if link.UsedTime == "" {
			link.UsedTime = now
			cols = append(cols, "used_time")
		}
	}
	_, err := ormer.Engine.ID(core.PK{link.Owner, link.Name}).Cols(cols...).Update(link)
	return err
}

// BuildMagicLinkCallbackURL is the link of the API: the callback page of the caller's
// frontend with the OAuth request the sign-in was started from. The origin is the one
// getMagicLinkOrigin() vouches for, never the bare "Host" header.
func BuildMagicLinkCallbackURL(link *MagicLink, token string, origin string) string {
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
	return fmt.Sprintf("%s/magic-link/callback?%s", strings.TrimRight(origin, "/"), query.Encode())
}

// getApiMagicLinkEmailContent picks the template of the mail. The provider's dedicated
// magic link templates come first, then its content if that is a magic link template
// (the built-in rule: it has a "%link"), then the branded default. "%expireTime" of the
// API is the moment the link expires, a TTL of days does not read well in minutes; the
// rest of the substitution is the built-in getMagicLinkEmailContent().
func getApiMagicLinkEmailContent(provider *Provider, magicLinkUrl string, user *User, link *MagicLink) string {
	content := provider.MagicLinkContent
	if link.AuthAction == MagicLinkAuthActionSignupNewUser && provider.MagicLinkSignupContent != "" {
		content = provider.MagicLinkSignupContent
	}
	if !strings.Contains(content, "%link") {
		content = provider.Content
	}
	if !strings.Contains(content, "%link") {
		if link.AuthAction == MagicLinkAuthActionSignupNewUser {
			content = GetDefaultMagicLinkSignupEmailContent()
		} else {
			content = GetDefaultMagicLinkEmailContent()
		}
	}
	content = strings.ReplaceAll(content, "%expireTime", link.ExpiryTime)

	templateProvider := *provider
	templateProvider.Content = content
	return getMagicLinkEmailContent(&templateProvider, magicLinkUrl, user)
}

// SendApiMagicLink issues and mails a link of the API. The row is written before the
// mail is sent and carries the outcome, so the admin list also shows what failed.
func SendApiMagicLink(organization *Organization, user *User, provider *Provider, link *MagicLink, token string, host string, lang string) error {
	origin, err := getMagicLinkOrigin(host, lang)
	if err != nil {
		return err
	}

	err = AddMagicLink(link)
	if err != nil {
		return err
	}

	title := provider.Title
	if title == "" {
		title = "Magic Link"
	}

	content := getApiMagicLinkEmailContent(provider, BuildMagicLinkCallbackURL(link, token, origin), user, link)
	err = SendEmail(provider, title, content, []string{link.Email}, organization.DisplayName)
	if err != nil {
		// a link nobody received must not stay claimable
		_, _ = ormer.Engine.ID(core.PK{link.Owner, link.Name}).Cols("is_used").Update(&MagicLink{IsUsed: true})
		_ = UpdateMagicLinkStatus(link, MagicLinkStatusFailed, err.Error())
		return err
	}

	return UpdateMagicLinkStatus(link, MagicLinkStatusSent, "")
}

// GetMagicLinkByToken looks a link up without claiming it.
func GetMagicLinkByToken(token string) (*MagicLink, error) {
	if token == "" {
		return nil, nil
	}
	link := &MagicLink{TokenHash: HashMagicLinkSecret(token)}
	existed, err := ormer.Engine.Get(link)
	if err != nil || !existed {
		return nil, err
	}
	return link, nil
}

func validateMagicLinkVerifyState(link *MagicLink, nowUnix int64) error {
	if link.Status == MagicLinkStatusRevoked {
		return errors.New("magic link was revoked")
	}
	if link.Status == MagicLinkStatusUsed {
		return errors.New("magic link was already used")
	}
	if link.Status == MagicLinkStatusExpired || isMagicLinkExpired(link, nowUnix) {
		return errors.New("magic link has expired")
	}
	if link.IsUsed || !isMagicLinkVerifiableStatus(link.Status) {
		return errors.New("magic link was already used")
	}
	return nil
}

func isMagicLinkVerifiableStatus(status string) bool {
	// a link of the built-in sign-in page has no status
	return status == "" || status == MagicLinkStatusCreated || status == MagicLinkStatusSent || status == MagicLinkStatusOpened
}

// getApiMagicLinkSessionHash is the session hash ConsumeMagicLink() is asked with: the
// link's own for a link that may be opened anywhere, the caller's secret for a bound
// one, and the browser session for a link of the built-in sign-in page.
func getApiMagicLinkSessionHash(link *MagicLink, sessionSecret string, browserSessionHash string) string {
	switch link.Binding {
	case MagicLinkBindingNone:
		return getUnboundMagicLinkSessionHash(link.TokenHash)
	case MagicLinkBindingClient:
		if sessionSecret == "" {
			return ""
		}
		return HashMagicLinkSecret(sessionSecret)
	default:
		return browserSessionHash
	}
}

// ConsumeApiMagicLink claims a link for /api/verify-magic-link. It answers the state of
// the link in the words the API always used, the claim itself is ConsumeMagicLink().
func ConsumeApiMagicLink(token string, sessionSecret string, browserSessionHash string, lang string) (*MagicLink, *Application, error) {
	link, err := GetMagicLinkByToken(token)
	if err != nil {
		return nil, nil, err
	}
	if link == nil {
		return nil, nil, errors.New("magic link is invalid")
	}

	err = validateMagicLinkVerifyState(link, time.Now().Unix())
	if err != nil {
		if link.Status != MagicLinkStatusExpired && err.Error() == "magic link has expired" && link.Status != "" {
			link.Status = MagicLinkStatusExpired
			_, _ = ormer.Engine.ID(core.PK{link.Owner, link.Name}).Cols("status").Update(link)
		}
		return nil, nil, err
	}

	application, err := GetApplication(util.GetId("admin", link.Application))
	if err != nil {
		return nil, nil, err
	}
	if application == nil || application.Organization != link.Owner {
		return nil, nil, errors.New("magic link application mismatch")
	}

	claimed, err := ConsumeMagicLink(token, getApiMagicLinkSessionHash(link, sessionSecret, browserSessionHash), application, lang)
	if err != nil {
		return nil, nil, err
	}

	err = UpdateMagicLinkStatus(claimed, MagicLinkStatusUsed, "")
	if err != nil {
		return nil, nil, err
	}
	return claimed, application, nil
}

// maskMagicLinkSigninCode keeps the token of a built-in magic link sign-in out of the
// record: a sign-in that fails before the link is consumed leaves the token valid.
func maskMagicLinkSigninCode(recordObject string) string {
	var value map[string]interface{}
	err := json.Unmarshal([]byte(recordObject), &value)
	if err != nil || value["signinMethod"] != "Magic link" {
		return recordObject
	}
	if _, ok := value["code"]; !ok {
		return recordObject
	}

	value["code"] = "***"
	res, err := json.Marshal(value)
	if err != nil {
		return recordObject
	}
	return string(res)
}
