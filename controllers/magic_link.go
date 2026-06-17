package controllers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/utils/pagination"
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
}

// SendMagicLink ...
// @Title SendMagicLink
// @Tag Magic Link API
// @Description send a one-time Magic Link for sign-in (existing user) or sign-up (new user when Magic link sign-up is enabled)
// @Description optional OAuth query parameters allow redirecting back to external frontend callback after verification
// @Description group, permission and custom TTL are trusted parameters and require clientId/clientSecret in request body
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
		c.ResponseError(c.T("auth:The login method: magic link is not enabled for the application"))
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
		if user != nil || !application.IsMagicLinkSignupEnabled() {
			errText := "magic link user does not exist or is not available"
			c.saveFailedMagicLink(application, &requestForm, clientIP, oauth, expireAt, errText)
			util.LogInfo(c.Ctx, "Magic link request accepted without sending, organization = %s, application = %s, email = %s, reason = %s", application.Organization, application.Name, requestForm.Email, errText)
			c.respondMagicLinkSendAccepted(expireAt)
			return
		}
	}

	var permission *object.Permission
	if !(user == nil && application.IsMagicLinkSignupEnabled() && requestForm.Permission == "") {
		permission, err = object.ResolveMagicLinkPermission(application, user, requestForm.Permission)
		if err != nil {
			link := object.NewMagicLink(application, nil, requestForm.Email, clientIP, c.GetSessionUsername(), "", oauth, expireAt)
			link.Group = requestForm.Group
			link.Status = object.MagicLinkStatusFailed
			link.LastError = err.Error()
			_ = object.AddMagicLink(link)
			util.LogInfo(c.Ctx, "Magic link request accepted without sending, organization = %s, application = %s, email = %s, magicLink = %s, reason = %s", application.Organization, application.Name, requestForm.Email, util.GetId(link.Owner, link.Name), err.Error())
			c.respondMagicLinkSendAccepted(expireAt)
			return
		}
	}

	token, err := object.GenerateMagicLinkToken()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	link := object.NewMagicLink(application, permission, requestForm.Email, clientIP, c.GetSessionUsername(), token, oauth, expireAt)
	link.Group = requestForm.Group
	link.AuthAction = getMagicLinkAuthAction(user == nil)
	err = object.AddMagicLink(link)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	magicLinkURL := object.BuildMagicLinkCallbackURL(link, token, c.Ctx.Request.Host)
	err = object.SendMagicLinkToEmail(organization, provider, requestForm.Email, magicLinkURL, link.ExpiryTime, link.AuthAction)
	if err != nil {
		_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
		util.LogWarning(c.Ctx, "Magic link email send failed, organization = %s, application = %s, email = %s, magicLink = %s, error = %s", application.Organization, application.Name, requestForm.Email, util.GetId(link.Owner, link.Name), err.Error())
		c.respondMagicLinkSendAccepted(expireAt)
		return
	}

	err = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusSent, "")
	if err != nil {
		c.ResponseError(err.Error())
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
	link.Status = object.MagicLinkStatusFailed
	link.LastError = lastError
	_ = object.AddMagicLink(link)
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

	link, err := object.ConsumeMagicLink(token)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.logMagicLinkStatus("verify", link, nil, link.Email, "")

	application, err := object.GetApplication(link.Application)
	if err != nil {
		_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
		c.ResponseError(err.Error())
		return
	}
	if application == nil || application.Organization != link.Owner {
		err = fmt.Errorf("magic link application mismatch")
		_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
		c.ResponseError(err.Error())
		return
	}
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
		var created bool
		user, created, err = c.createMagicLinkSignupUser(application, link)
		if err != nil {
			object.MagicLinkSignupFailed.Inc()
			c.logMagicLinkStatus("create-user", link, application, link.Email, err.Error())
			_ = object.UpdateMagicLinkStatus(link, object.MagicLinkStatusFailed, err.Error())
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
		if created {
			object.MagicLinkSignupCreated.Inc()
			isNewUser = true
			c.Ctx.Input.SetParam("recordUserId", user.GetId())
			c.Ctx.Input.SetParam("recordSignup", "true")
			c.logMagicLinkStatus("create-user", link, application, user.Email, "")
		} else {
			c.logMagicLinkStatus("create-user", link, application, user.Email, "existing-user")
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

func (c *ApiController) createMagicLinkSignupUser(application *object.Application, link *object.MagicLink) (*object.User, bool, error) {
	organization, err := object.GetOrganization(util.GetId(application.Owner, application.Organization))
	if err != nil {
		return nil, false, err
	}
	if organization == nil {
		return nil, false, fmt.Errorf("organization does not exist")
	}

	id, err := object.GenerateIdForNewUser(application)
	if err != nil {
		return nil, false, err
	}
	email := strings.TrimSpace(strings.ToLower(link.Email))
	username := resolveSignupUsername(application, organization, "", email, id)
	initScore, err := organization.GetInitScore()
	if err != nil {
		return nil, false, err
	}
	user := &object.User{
		Owner:             application.Organization,
		Name:              username,
		Id:                id,
		Type:              "normal-user",
		Password:          util.GenerateId(),
		DisplayName:       resolveSignupDisplayName(email, username),
		Avatar:            organization.DefaultAvatar,
		Email:             email,
		Address:           []string{},
		Score:             initScore,
		IsAdmin:           false,
		IsForbidden:       false,
		IsDeleted:         false,
		SignupApplication: application.Name,
		Properties:        map[string]string{},
		Karma:             0,
		EmailVerified:     true,
		RegisterType:      "Magic Link Signup",
		RegisterSource:    fmt.Sprintf("%s/%s", application.Organization, application.Name),
	}
	if link.Group != "" {
		user.Groups = []string{link.Group}
	} else if application.DefaultGroup != "" {
		user.Groups = []string{application.DefaultGroup}
	}
	if application.DefaultTag != "" {
		user.Tag = application.DefaultTag
	}

	return c.createSignupUser(user, true)
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
	return user == nil && application != nil && application.IsMagicLinkSignupEnabled()
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

// GetMagicLinks ...
// @Title GetMagicLinks
// @Tag Magic Link API
// @Description get Magic Link records for admin UI
// @Param owner query string false "The owner of Magic Link records"
// @Param p query string false "The page number"
// @Param pageSize query string false "The page size"
// @Param field query string false "The search field"
// @Param value query string false "The search value"
// @Param sortField query string false "The sort field"
// @Param sortOrder query string false "The sort order"
// @Success 200 {array} object.MagicLink The Response object
// @router /get-magic-links [get]
func (c *ApiController) GetMagicLinks() {
	organization, ok := c.RequireAdmin()
	if !ok {
		return
	}

	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	status := c.Ctx.Input.Query("status")
	user := c.Ctx.Input.Query("user")
	email := c.Ctx.Input.Query("email")
	application := c.Ctx.Input.Query("application")
	queryOrganization := c.Ctx.Input.Query("organization")
	group := c.Ctx.Input.Query("group")
	permission := c.Ctx.Input.Query("permission")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")
	owner := c.Ctx.Input.Query("owner")
	if owner == "" && queryOrganization != "" {
		owner = queryOrganization
	}
	if c.IsGlobalAdmin() && owner != "" {
		organization = owner
	}
	if field == "" || value == "" {
		filterPairs := []struct {
			field string
			value string
		}{
			{field: "status", value: status},
			{field: "requester", value: user},
			{field: "email", value: email},
			{field: "application", value: application},
			{field: "owner", value: queryOrganization},
			{field: "subGroups", value: group},
			{field: "permission", value: permission},
		}
		for _, pair := range filterPairs {
			if pair.value != "" {
				field = pair.field
				value = pair.value
				break
			}
		}
	}

	if limit == "" || page == "" {
		links, err := object.GetMagicLinks(organization)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(links)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetMagicLinkCount(organization, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		links, err := object.GetPaginationMagicLinks(organization, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(links, paginator.Nums())
	}
}

// RevokeMagicLink ...
// @Title RevokeMagicLink
// @Tag Magic Link API
// @Description revoke an unused Magic Link
// @Param id query string true "The id (owner/name) of the Magic Link"
// @Success 200 {object} controllers.Response The Response object
// @router /revoke-magic-link [post]
func (c *ApiController) RevokeMagicLink() {
	organization, ok := c.RequireAdmin()
	if !ok {
		return
	}

	id := c.Ctx.Input.Query("id")
	if id == "" {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(id)
	if !c.IsGlobalAdmin() && owner != organization {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	c.Data["json"] = wrapActionResponse(object.RevokeMagicLink(id))
	c.ServeJSON()
}

func (c *ApiController) DeleteMagicLink() {
	organization, ok := c.RequireAdmin()
	if !ok {
		return
	}

	id := c.Ctx.Input.Query("id")
	if id == "" {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(id)
	if !c.IsGlobalAdmin() && owner != organization {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	c.Data["json"] = wrapActionResponse(object.DeleteMagicLink(id))
	c.ServeJSON()
}
