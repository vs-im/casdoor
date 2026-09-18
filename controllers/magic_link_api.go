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

// The headless magic link API: a server-to-server caller (a frontend with its own
// sign-in page) asks for a link with /api/send-magic-link and trades the token of the
// link for the sign-in with /api/verify-magic-link. It is a layer on top of the built-in
// magic link sign-in (magic_link.go, verification.go, auth.go) and shares its table, its
// token, its mail substitution, its atomic claim and its signup.

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/casdoor/casdoor/captcha"
	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type MagicLinkRequestForm struct {
	Email             string `json:"email"`
	Organization      string `json:"organization"`
	Application       string `json:"application"`
	ClientID          string `json:"clientId"`
	ClientSecret      string `json:"clientSecret"`
	ApplicationSecret string `json:"applicationClientSecret"`
	Group             string `json:"group"`
	Permission        string `json:"permission"`
	ExpiresInMinutes  int    `json:"expiresInMinutes"`
	ExpireTime        string `json:"expireTime"`
	CaptchaType       string `json:"captchaType"`
	CaptchaToken      string `json:"captchaToken"`
	// SessionSecret optionally binds the link to the caller: /api/verify-magic-link then
	// only accepts the token together with the same secret. Without it the link may be
	// opened on any device.
	SessionSecret string `json:"sessionSecret"`
}

// SendMagicLink ...
// @Title SendMagicLink
// @Tag Magic Link API
// @Description send a one-time Magic Link for sign-in (existing user) or sign-up (new user when Magic link sign-up is enabled)
// @Description optional OAuth query parameters allow redirecting back to external frontend callback after verification
// @Description group, permission and custom TTL are trusted parameters and require clientId/clientSecret in request body
// @Description an optional sessionSecret binds the link to the caller, verify-magic-link then requires the same secret
// @Param clientId query string false "OAuth client ID"
// @Param responseType query string false "OAuth response type"
// @Param redirectUri query string false "OAuth redirect URI"
// @Param scope query string false "OAuth scope"
// @Param state query string false "OAuth state"
// @Param nonce query string false "OAuth nonce"
// @Param code_challenge_method query string false "OAuth PKCE code challenge method"
// @Param code_challenge query string false "OAuth PKCE code challenge"
// @Param resource query string false "OAuth resource indicator"
// @Param form body controllers.MagicLinkRequestForm true "Magic Link request"
// @Success 200 {object} controllers.Response The Response object
// @router /send-magic-link [post]
func (c *ApiController) SendMagicLink() {
	var requestForm MagicLinkRequestForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestForm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	requestForm.Email = strings.TrimSpace(strings.ToLower(requestForm.Email))
	requestForm.Organization = strings.TrimSpace(requestForm.Organization)
	requestForm.Application = strings.TrimSpace(requestForm.Application)
	requestForm.ClientID = strings.TrimSpace(requestForm.ClientID)
	requestForm.ClientSecret = strings.TrimSpace(requestForm.ClientSecret)
	requestForm.ApplicationSecret = strings.TrimSpace(requestForm.ApplicationSecret)
	requestForm.Group = strings.TrimSpace(requestForm.Group)
	requestForm.Permission = strings.TrimSpace(requestForm.Permission)
	requestForm.ExpireTime = strings.TrimSpace(requestForm.ExpireTime)
	err = validateMagicLinkRequestForm(&requestForm)
	if err != nil {
		if err.Error() == "invalid email" {
			c.ResponseError(c.T("check:Email is invalid"))
		} else {
			c.ResponseError(c.T("general:Missing parameter"))
		}
		return
	}

	application, err := c.getMagicLinkApplication(&requestForm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application.Organization != requestForm.Organization {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	requestedPermission := requestForm.Permission
	needsClientCredentials := magicLinkRequestNeedsClientCredentials(&requestForm, requestedPermission)
	hasCustomTTL := magicLinkRequestHasCustomTTL(&requestForm)
	trustedRequest := false
	if needsClientCredentials || hasCustomTTL {
		trustedRequest, err = c.isTrustedMagicLinkRequest(application, &requestForm)
		if needsClientCredentials && (err != nil || !trustedRequest) {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
		if hasCustomTTL && !trustedRequest {
			requestForm.ExpiresInMinutes = 0
			requestForm.ExpireTime = ""
		}
	}
	if requestForm.Permission == "" {
		requestForm.Permission = application.MagicLinkPermission
	}
	if !application.IsMagicLinkEnabled() {
		c.ResponseError(c.T("auth:The login method: login with magic link is not enabled for the application"))
		return
	}
	// the link is never pointed at a site taken from a forged "Host" header
	err = object.CheckMagicLinkOrigin(c.Ctx.Request.Host, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	expireAt, err := object.ResolveMagicLinkExpireTime(application, requestForm.ExpiresInMinutes, requestForm.ExpireTime, time.Now())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	oauth := map[string]string{
		"clientId":            c.Ctx.Input.Query("clientId"),
		"responseType":        c.Ctx.Input.Query("responseType"),
		"redirectUri":         c.Ctx.Input.Query("redirectUri"),
		"scope":               c.Ctx.Input.Query("scope"),
		"state":               c.Ctx.Input.Query("state"),
		"nonce":               c.Ctx.Input.Query("nonce"),
		"codeChallengeMethod": c.Ctx.Input.Query("code_challenge_method"),
		"codeChallenge":       c.Ctx.Input.Query("code_challenge"),
		"resource":            c.Ctx.Input.Query("resource"),
	}
	clientIP := util.GetClientIpFromRequest(c.Ctx.Request)
	if oauth["responseType"] != "" && oauth["responseType"] != "login" {
		msg, oauthApplication, err := object.CheckOAuthLogin(oauth["clientId"], oauth["responseType"], oauth["redirectUri"], oauth["scope"], oauth["state"], c.GetAcceptLanguage())
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if msg != "" {
			c.ResponseError(msg)
			return
		}
		if oauthApplication == nil || oauthApplication.GetId() != application.GetId() {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
	}

	organization, err := object.GetOrganization(util.GetId(application.Owner, application.Organization))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if organization == nil {
		c.ResponseError(c.T("check:Organization does not exist"))
		return
	}

	err = object.IsMagicLinkAllowSend(requestForm.Email, clientIP, application)
	if err != nil {
		util.LogWarning(c.Ctx, "Magic link request rate limited, organization = %s, application = %s, email = %s, remoteAddr = %s, error = %s", application.Organization, application.Name, requestForm.Email, clientIP, err.Error())
		c.ResponseError(err.Error())
		return
	}
	captchaRequired, err := object.IsMagicLinkCaptchaRequired(requestForm.Email, clientIP, application)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if captchaRequired || requestForm.CaptchaToken != "" {
		ok, err := c.verifyMagicLinkCaptcha(application, &requestForm, captchaRequired)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if !ok {
			return
		}
	}

	provider, err := application.GetEmailProvider("Magic link")
	if err != nil {
		c.saveFailedMagicLink(application, &requestForm, clientIP, oauth, expireAt, err.Error())
		util.LogWarning(c.Ctx, "Magic link email provider lookup failed, organization = %s, application = %s, email = %s, error = %s", application.Organization, application.Name, requestForm.Email, err.Error())
		c.respondMagicLinkSendAccepted(expireAt)
		return
	}
	if provider == nil {
		errText := fmt.Sprintf(c.T("verification:please add an Email provider to the \"Providers\" list for the application: %s"), application.Name)
		c.saveFailedMagicLink(application, &requestForm, clientIP, oauth, expireAt, errText)
		util.LogWarning(c.Ctx, "Magic link email provider is not configured, organization = %s, application = %s, email = %s", application.Organization, application.Name, requestForm.Email)
		c.respondMagicLinkSendAccepted(expireAt)
		return
	}

	user, err := object.GetUserByEmail(application.Organization, requestForm.Email)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil || user.IsDeleted || user.IsForbidden {
		errText := ""
		if user != nil || !application.IsMagicLinkApiSignupEnabled() {
			errText = "magic link user does not exist or is not available"
		} else if err = object.CheckMagicLinkSignup(application.GetMagicLinkSignupApplication(), c.GetAcceptLanguage()); err != nil {
			// the application asks the signup page for more than a link can answer
			errText = err.Error()
		}
		if errText != "" {
			c.saveFailedMagicLink(application, &requestForm, clientIP, oauth, expireAt, errText)
			util.LogInfo(c.Ctx, "Magic link request accepted without sending, organization = %s, application = %s, email = %s, reason = %s", application.Organization, application.Name, requestForm.Email, errText)
			c.respondMagicLinkSendAccepted(expireAt)
			return
		}
	}

	var permission *object.Permission
	if !(user == nil && requestForm.Permission == "") {
		permission, err = object.ResolveMagicLinkPermission(application, user, requestForm.Permission)
		if err != nil {
			c.saveFailedMagicLink(application, &requestForm, clientIP, oauth, expireAt, err.Error())
			util.LogInfo(c.Ctx, "Magic link request accepted without sending, organization = %s, application = %s, email = %s, reason = %s", application.Organization, application.Name, requestForm.Email, err.Error())
			c.respondMagicLinkSendAccepted(expireAt)
			return
		}
	}

	token, link, err := object.NewApiMagicLink(application, permission, requestForm.Email, clientIP, c.GetSessionUsername(), oauth, expireAt)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	link.Group = requestForm.Group
	link.AuthAction = getMagicLinkAuthAction(user == nil)
	object.BindMagicLinkToClient(link, requestForm.SessionSecret)

	err = object.SendApiMagicLink(organization, user, provider, link, token, c.Ctx.Request.Host, c.GetAcceptLanguage())
	if err != nil {
		util.LogWarning(c.Ctx, "Magic link email send failed, organization = %s, application = %s, email = %s, magicLink = %s, error = %s", application.Organization, application.Name, requestForm.Email, util.GetId(link.Owner, link.Name), err.Error())
		c.respondMagicLinkSendAccepted(expireAt)
		return
	}

	c.logMagicLinkStatus("send", link, application, requestForm.Email, "")
	c.respondMagicLinkSendAccepted(expireAt)
}

func (c *ApiController) verifyMagicLinkCaptcha(application *object.Application, requestForm *MagicLinkRequestForm, required bool) (bool, error) {
	captchaProvider, err := object.GetCaptchaProviderByApplication(application.GetId(), "false", c.GetAcceptLanguage())
	if err != nil {
		return false, err
	}
	if captchaProvider == nil {
		if required {
			c.ResponseError(c.T("verification:Invalid captcha provider."))
			return false, nil
		}
		return true, nil
	}
	if requestForm.CaptchaToken == "" {
		c.ResponseError(c.T("general:Missing parameter")+": captchaToken.", "captchaRequired")
		return false, nil
	}
	if requestForm.CaptchaType != captchaProvider.Type {
		c.ResponseError(c.T("verification:Turing test failed."))
		return false, nil
	}
	captchaClientSecret := ""
	if captchaProvider.Type == "Default" {
		captchaClientSecret = requestForm.ClientSecret
	} else {
		captchaClientSecret = captchaProvider.ClientSecret
	}
	isHuman, err := captcha.VerifyCaptchaByCaptchaType(requestForm.CaptchaType, requestForm.CaptchaToken, captchaProvider.ClientId, captchaClientSecret, captchaProvider.ClientId2)
	if err != nil {
		return false, err
	}
	if !isHuman {
		c.ResponseError(c.T("verification:Turing test failed."))
		return false, nil
	}
	return true, nil
}

func (c *ApiController) saveFailedMagicLink(application *object.Application, requestForm *MagicLinkRequestForm, clientIP string, oauth map[string]string, expireAt time.Time, lastError string) {
	link := object.NewMagicLink(application, nil, requestForm.Email, clientIP, c.GetSessionUsername(), "", oauth, expireAt)
	link.Group = requestForm.Group
	_ = object.AddFailedMagicLink(link, lastError)
}

func (c *ApiController) respondMagicLinkSendAccepted(expireAt time.Time) {
	c.ResponseOk(map[string]string{"expireTime": util.Time2String(expireAt)})
}

func (c *ApiController) isTrustedMagicLinkRequest(application *object.Application, requestForm *MagicLinkRequestForm) (bool, error) {
	clientID := requestForm.ClientID
	if clientID == "" {
		clientID = c.Ctx.Input.Query("clientId")
	}
	if clientID == "" {
		clientID = c.Ctx.Input.Query("client_id")
	}
	clientSecret := requestForm.ClientSecret
	if requestForm.ApplicationSecret != "" {
		clientSecret = requestForm.ApplicationSecret
	}
	if requestForm.CaptchaToken != "" && requestForm.ApplicationSecret == "" {
		clientSecret = ""
	}
	if clientSecret == "" {
		clientSecret = c.Ctx.Input.Query("clientSecret")
	}
	if clientSecret == "" {
		clientSecret = c.Ctx.Input.Query("client_secret")
	}
	return validateMagicLinkClientCredentials(application, clientID, clientSecret)
}

func validateMagicLinkClientCredentials(application *object.Application, clientID string, clientSecret string) (bool, error) {
	if clientID == "" && clientSecret == "" {
		return false, nil
	}
	if application == nil || clientID == "" || clientSecret == "" {
		return false, fmt.Errorf("invalid magic link client credentials")
	}
	if application.ClientId != clientID || application.ClientSecret != clientSecret {
		return false, fmt.Errorf("invalid magic link client credentials")
	}
	return true, nil
}

func magicLinkRequestNeedsClientCredentials(requestForm *MagicLinkRequestForm, requestedPermission string) bool {
	return requestForm.Group != "" || requestedPermission != ""
}

func magicLinkRequestHasCustomTTL(requestForm *MagicLinkRequestForm) bool {
	return requestForm.ExpiresInMinutes > 0 || requestForm.ExpireTime != ""
}

func (c *ApiController) logMagicLinkStatus(status string, link *object.MagicLink, application *object.Application, email string, lastError string) {
	organization := ""
	applicationName := ""
	magicLinkID := ""
	if link != nil {
		organization = link.Owner
		magicLinkID = util.GetId(link.Owner, link.Name)
	}
	if application != nil {
		organization = application.Organization
		applicationName = application.Name
	} else if link != nil {
		applicationName = link.Application
	}
	util.LogInfo(c.Ctx, "Magic link status, status = %s, organization = %s, application = %s, email = %s, magicLink = %s, error = %s", status, organization, applicationName, email, magicLinkID, lastError)
}

// VerifyMagicLink ...
// @Title VerifyMagicLink
// @Tag Magic Link API
// @Description verify a one-time Magic Link token and continue sign-in or sign-up flow
// @Description if user does not exist and application allows Magic link sign-up, user will be created and then authenticated
// @Description if OAuth query parameters are present, response continues OAuth flow and redirects to configured callback
// @Param token query string true "Magic Link token"
// @Param sessionSecret query string false "The secret the link was bound to by send-magic-link"
// @Param clientId query string false "OAuth client ID"
// @Param responseType query string false "OAuth response type"
// @Param redirectUri query string false "OAuth redirect URI"
// @Param scope query string false "OAuth scope"
// @Param state query string false "OAuth state"
// @Param nonce query string false "OAuth nonce"
// @Param code_challenge_method query string false "OAuth PKCE code challenge method"
// @Param code_challenge query string false "OAuth PKCE code challenge"
// @Param resource query string false "OAuth resource indicator"
// @Success 200 {object} controllers.Response The Response object
// @router /verify-magic-link [get]
func (c *ApiController) VerifyMagicLink() {
	token := c.Ctx.Input.Query("token")
	if token == "" {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	link, application, err := object.ConsumeApiMagicLink(token, c.Ctx.Input.Query("sessionSecret"), c.getMagicLinkSessionHash(), c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.logMagicLinkStatus("verify", link, application, link.Email, "")

	if !application.IsMagicLinkEnabled() {
		err = fmt.Errorf("magic link sign-in is disabled for the application")
		_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	err = c.validateMagicLinkOAuthContext(link)
	if err != nil {
		_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
		c.ResponseError(err.Error())
		return
	}

	isNewUser := false
	user, err := object.GetUserByEmail(link.Owner, link.Email)
	if err != nil {
		_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		if !shouldCreateUserOnMagicLinkVerify(application, user) {
			err = fmt.Errorf("magic link user is not available")
			_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
		user, err = c.addApiMagicLinkUser(application, link)
		if err != nil {
			object.MagicLinkSignupFailed.Inc()
			c.logMagicLinkStatus("create-user", link, application, link.Email, err.Error())
			_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
		object.MagicLinkSignupCreated.Inc()
		isNewUser = true
		c.Ctx.Input.SetParam("recordUserId", user.GetId())
		c.logMagicLinkStatus("create-user", link, application, user.Email, "")
	} else if !user.EmailVerified && !user.IsDeleted && !user.IsForbidden {
		// the link proves the address, the same as on the built-in sign-in page
		user.EmailVerified = true
		_, err = object.UpdateUser(user.GetId(), user, []string{"email_verified"}, false)
		if err != nil {
			_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
			c.ResponseError(err.Error())
			return
		}
	}
	if user.IsDeleted || user.IsForbidden {
		err = fmt.Errorf("magic link user is not available")
		_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	permission, err := object.ResolveMagicLinkPermission(application, user, link.Permission)
	if err != nil {
		_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
		c.ResponseError(err.Error())
		return
	}
	object.ApplyPermissionSnapshotToMagicLink(link, permission)
	if permission != nil {
		_ = object.UpdateMagicLinkPermissionSnapshot(link)
	}

	authForm := form.AuthForm{
		Type:         link.ResponseType,
		SigninMethod: "Magic link",
		Application:  application.Name,
		Organization: link.Owner,
		AutoSignin:   true,
	}
	if authForm.Type == "" {
		authForm.Type = ResponseTypeLogin
	}

	resp := c.HandleLoggedIn(application, user, &authForm)
	if resp == nil {
		c.logMagicLinkStatus("login-failed", link, application, user.Email, "magic link login failed")
		_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, "magic link login failed")
		return
	}
	if resp.Status != "ok" {
		c.logMagicLinkStatus("login-failed", link, application, user.Email, resp.Msg)
		_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, resp.Msg)
	} else {
		object.MagicLinkVerifySuccess.Inc()
		c.logMagicLinkStatus("login-success", link, application, user.Email, "")
		_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusUsed, "")
	}
	if _, ok := resp.Data.(map[string]bool); ok {
		resp.Data2 = application.Name
	}
	resp.AuthAction = getMagicLinkAuthAction(isNewUser)
	resp.IsNewUser = &isNewUser
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *ApiController) getMagicLinkApplication(requestForm *MagicLinkRequestForm) (*object.Application, error) {
	if requestForm.Application == "" {
		application, err := object.GetApplicationByOrganizationName(requestForm.Organization)
		if err != nil {
			return nil, err
		}
		if application == nil {
			return nil, fmt.Errorf("%s", c.T("check:Application does not exist"))
		}
		requestForm.Application = application.Name
		return application, nil
	}

	applicationID := requestForm.Application
	if !strings.Contains(applicationID, "/") {
		applicationID = util.GetId("admin", applicationID)
	}
	application, err := object.GetApplication(applicationID)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, fmt.Errorf(c.T("auth:The application: %s does not exist"), requestForm.Application)
	}
	return application, nil
}

// addApiMagicLinkUser signs the address of the link up. The user is created by the
// built-in addMagicLinkUser(), with its signup item, invitation code and entry IP checks;
// the API only adds the group a trusted caller asked for.
func (c *ApiController) addApiMagicLinkUser(application *object.Application, link *object.MagicLink) (*object.User, error) {
	user, err := c.addMagicLinkUser(application.GetMagicLinkSignupApplication(), link)
	if err != nil {
		return nil, err
	}

	if link.Group != "" && (len(user.Groups) != 1 || user.Groups[0] != link.Group) {
		user.Groups = []string{link.Group}
		_, err = object.UpdateUser(user.GetId(), user, []string{"groups"}, false)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

func validateMagicLinkRequestForm(requestForm *MagicLinkRequestForm) error {
	if requestForm.Email == "" || requestForm.Organization == "" {
		return fmt.Errorf("missing parameter")
	}
	if !util.IsEmailValid(requestForm.Email) {
		return fmt.Errorf("invalid email")
	}
	return nil
}

func shouldCreateUserOnMagicLinkVerify(application *object.Application, user *object.User) bool {
	return user == nil && application != nil && application.IsMagicLinkApiSignupEnabled()
}

func getMagicLinkAuthAction(isNewUser bool) string {
	if isNewUser {
		return object.MagicLinkAuthActionSignupNewUser
	}
	return object.MagicLinkAuthActionSigninExistingUser
}

func (c *ApiController) validateMagicLinkOAuthContext(link *object.MagicLink) error {
	if link.ResponseType == "" || link.ResponseType == ResponseTypeLogin {
		return nil
	}
	expected := map[string]string{
		"clientId":              link.ClientId,
		"responseType":          link.ResponseType,
		"redirectUri":           link.RedirectUri,
		"scope":                 link.Scope,
		"state":                 link.State,
		"nonce":                 link.Nonce,
		"code_challenge_method": link.CodeChallengeMethod,
		"code_challenge":        link.CodeChallenge,
		"resource":              link.Resource,
	}
	actual := map[string]string{}
	for key := range expected {
		actual[key] = c.Ctx.Input.Query(key)
	}
	return validateMagicLinkOAuthPayload(expected, actual)
}

func validateMagicLinkOAuthPayload(expected map[string]string, actual map[string]string) error {
	for key, value := range expected {
		if actual[key] != value {
			return fmt.Errorf("magic link OAuth context mismatch")
		}
	}
	return nil
}
