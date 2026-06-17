package controllers

import (
	"testing"

	"github.com/casdoor/casdoor/object"
)

func TestValidateMagicLinkRequestFormRequiresOrganization(t *testing.T) {
	form := &MagicLinkRequestForm{
		Email:        "user@example.com",
		Organization: "",
	}
	err := validateMagicLinkRequestForm(form)
	if err == nil {
		t.Fatal("expected error when organization is empty")
	}
}

func TestValidateMagicLinkRequestFormAllowsOptionalFields(t *testing.T) {
	form := &MagicLinkRequestForm{
		Email:            "user@example.com",
		Organization:     "built-in",
		Application:      "",
		Group:            "",
		Permission:       "",
		ExpiresInMinutes: 0,
		ExpireTime:       "",
		CaptchaType:      "",
		ClientSecret:     "",
		CaptchaToken:     "",
	}
	err := validateMagicLinkRequestForm(form)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShouldCreateUserOnMagicLinkVerify(t *testing.T) {
	application := &object.Application{
		MagicLinkSigninEnabled: true,
		EnableMagicLinkSignup:  true,
	}
	if !shouldCreateUserOnMagicLinkVerify(application, nil) {
		t.Fatal("expected user creation to be allowed")
	}
	if shouldCreateUserOnMagicLinkVerify(application, &object.User{}) {
		t.Fatal("expected no creation for existing user")
	}
	application.EnableMagicLinkSignup = false
	if shouldCreateUserOnMagicLinkVerify(application, nil) {
		t.Fatal("expected no creation when signup is disabled")
	}
	application.MagicLinkSigninEnabled = false
	if shouldCreateUserOnMagicLinkVerify(application, nil) {
		t.Fatal("expected no creation when sign-in is disabled")
	}
}

func TestGetMagicLinkAuthAction(t *testing.T) {
	if getMagicLinkAuthAction(true) != "signup_new_user" {
		t.Fatal("expected signup_new_user authAction for new user")
	}
	if getMagicLinkAuthAction(false) != "signin_existing_user" {
		t.Fatal("expected signin_existing_user authAction for existing user")
	}
}

func TestValidateMagicLinkClientCredentials(t *testing.T) {
	application := &object.Application{
		ClientId:     "client-id",
		ClientSecret: "client-secret",
	}
	ok, err := validateMagicLinkClientCredentials(application, "client-id", "client-secret")
	if err != nil || !ok {
		t.Fatalf("expected client credentials to be valid, ok = %v, err = %v", ok, err)
	}
	ok, err = validateMagicLinkClientCredentials(application, "client-id", "wrong-secret")
	if err == nil || ok {
		t.Fatal("expected invalid client credentials to be rejected")
	}
	ok, err = validateMagicLinkClientCredentials(application, "", "")
	if err != nil || ok {
		t.Fatalf("expected missing client credentials to be optional, ok = %v, err = %v", ok, err)
	}
}

func TestMagicLinkRequestNeedsClientCredentials(t *testing.T) {
	if !magicLinkRequestNeedsClientCredentials(&MagicLinkRequestForm{Group: "built-in/admins"}, "") {
		t.Fatal("expected group to require client credentials")
	}
	if !magicLinkRequestNeedsClientCredentials(&MagicLinkRequestForm{}, "built-in/permission") {
		t.Fatal("expected requested permission to require client credentials")
	}
	if magicLinkRequestNeedsClientCredentials(&MagicLinkRequestForm{}, "") {
		t.Fatal("expected plain email magic link request not to require client credentials")
	}
}

func TestMagicLinkRequestHasCustomTTL(t *testing.T) {
	if !magicLinkRequestHasCustomTTL(&MagicLinkRequestForm{ExpiresInMinutes: 60}) {
		t.Fatal("expected expiresInMinutes to be detected as custom ttl")
	}
	if !magicLinkRequestHasCustomTTL(&MagicLinkRequestForm{ExpireTime: "2026-06-17T10:00:00Z"}) {
		t.Fatal("expected expireTime to be detected as custom ttl")
	}
	if magicLinkRequestHasCustomTTL(&MagicLinkRequestForm{}) {
		t.Fatal("expected default ttl request not to be custom")
	}
}

func TestValidateMagicLinkOAuthPayloadRejectsMismatch(t *testing.T) {
	expected := map[string]string{
		"clientId":     "client",
		"responseType": "code",
		"redirectUri":  "https://example.com/callback",
	}
	actual := map[string]string{
		"clientId":     "client",
		"responseType": "code",
		"redirectUri":  "https://evil.example.com/callback",
	}
	err := validateMagicLinkOAuthPayload(expected, actual)
	if err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestValidateMagicLinkOAuthPayloadAllowsExactMatch(t *testing.T) {
	expected := map[string]string{
		"clientId":     "client",
		"responseType": "code",
		"redirectUri":  "https://example.com/callback",
		"scope":        "openid profile",
	}
	actual := map[string]string{
		"clientId":     "client",
		"responseType": "code",
		"redirectUri":  "https://example.com/callback",
		"scope":        "openid profile",
	}
	err := validateMagicLinkOAuthPayload(expected, actual)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveSignupUsernameUsesEmailWhenConfigured(t *testing.T) {
	application := &object.Application{}
	organization := &object.Organization{
		UseEmailAsUsername: true,
	}
	username := resolveSignupUsername(application, organization, "", "User@Example.com", "generated-id")
	if username != "user@example.com" {
		t.Fatalf("username = %s, want user@example.com", username)
	}
}

func TestResolveSignupUsernameUsesGeneratedIdWhenNoEmailUsername(t *testing.T) {
	application := &object.Application{}
	organization := &object.Organization{
		UseEmailAsUsername: false,
	}
	username := resolveSignupUsername(application, organization, "", "user@example.com", "generated-id")
	if username != "generated-id" {
		t.Fatalf("username = %s, want generated-id", username)
	}
}

func TestResolveSignupDisplayNameUsesEmail(t *testing.T) {
	displayName := resolveSignupDisplayName("User@Example.com", "generated-id")
	if displayName != "user@example.com" {
		t.Fatalf("displayName = %s, want user@example.com", displayName)
	}
}
