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

import "testing"

func TestRedirectUriMatchesPattern(t *testing.T) {
	tests := []struct {
		redirectUri string
		targetUri   string
		want        bool
	}{
		// Exact match
		{"https://login.example.com/callback", "https://login.example.com/callback", true},

		// Full URL pattern: exact host
		{"https://login.example.com/callback", "https://login.example.com/callback", true},
		{"https://login.example.com/other", "https://login.example.com/callback", false},

		// Full URL pattern: subdomain of configured host
		{"https://def.abc.com/callback", "abc.com", true},
		{"https://def.abc.com/callback", ".abc.com", true},
		{"https://def.abc.com/callback", ".abc.com/", true},
		{"https://deep.app.example.com/callback", "https://example.com/callback", true},

		// Full URL pattern: unrelated host must not match
		{"https://evil.com/callback", "https://example.com/callback", false},
		// Suffix collision: evilexample.com must not match example.com
		{"https://evilexample.com/callback", "https://example.com/callback", false},

		// Full URL pattern: scheme mismatch
		{"http://app.example.com/callback", "https://example.com/callback", false},

		// Full URL pattern: path mismatch
		{"https://app.example.com/other", "https://example.com/callback", false},

		// Scheme-less pattern: exact host
		{"https://login.example.com/callback", "login.example.com/callback", true},
		{"http://login.example.com/callback", "login.example.com/callback", true},

		// Scheme-less pattern: subdomain of configured host
		{"https://app.login.example.com/callback", "login.example.com/callback", true},

		// Scheme-less pattern: unrelated host must not match
		{"https://evil.com/callback", "login.example.com/callback", false},

		// Scheme-less pattern: query-string injection must not match
		{"https://evil.com/?r=https://login.example.com/callback", "login.example.com/callback", false},
		{"https://evil.com/page?redirect=https://login.example.com/callback", "login.example.com/callback", false},

		// Scheme-less pattern: path mismatch
		{"https://login.example.com/other", "login.example.com/callback", false},

		// Scheme-less pattern: non-http scheme must not match
		{"ftp://login.example.com/callback", "login.example.com/callback", false},

		// Empty target
		{"https://login.example.com/callback", "", false},
	}

	for _, tt := range tests {
		got := redirectUriMatchesPattern(tt.redirectUri, tt.targetUri)
		if got != tt.want {
			t.Errorf("redirectUriMatchesPattern(%q, %q) = %v, want %v", tt.redirectUri, tt.targetUri, got, tt.want)
		}
	}
}

func TestIsMagicLinkEnabledRequiresSigninFlag(t *testing.T) {
	application := &Application{
		SigninMethods: []*SigninMethod{
			{Name: "Magic link"},
		},
		MagicLinkSigninEnabled: false,
	}
	if application.IsMagicLinkEnabled() {
		t.Fatal("magic link should be disabled when magicLinkSigninEnabled is false")
	}
	application.MagicLinkSigninEnabled = true
	if !application.IsMagicLinkEnabled() {
		t.Fatal("magic link should be enabled when method exists and magicLinkSigninEnabled is true")
	}
}

func TestIsMagicLinkSignupEnabledDependsOnSigninFlag(t *testing.T) {
	application := &Application{
		MagicLinkSigninEnabled: false,
		EnableMagicLinkSignup:  true,
	}
	if application.IsMagicLinkSignupEnabled() {
		t.Fatal("signup should be disabled when sign-in is disabled")
	}
	application.MagicLinkSigninEnabled = true
	if !application.IsMagicLinkSignupEnabled() {
		t.Fatal("signup should be enabled only when both flags are true")
	}
}

func TestApplicationIsMagicLinkEnabled(t *testing.T) {
	application := &Application{}
	if application.IsMagicLinkEnabled() {
		t.Fatal("magic link should be disabled when no signin method is configured")
	}
	application.SigninMethods = []*SigninMethod{
		{Name: "Password", Rule: "All"},
	}
	if application.IsMagicLinkEnabled() {
		t.Fatal("magic link should be disabled when signin methods do not include magic link")
	}
	application.SigninMethods = append(application.SigninMethods, &SigninMethod{Name: "Magic link", Rule: "None"})
	application.MagicLinkSigninEnabled = true
	if !application.IsMagicLinkEnabled() {
		t.Fatal("magic link should be enabled when signin methods include magic link")
	}
}

func TestGetMagicLinkExpireMinutes(t *testing.T) {
	var nilApplication *Application
	if nilApplication.GetMagicLinkExpireMinutes() != MagicLinkDefaultExpireMinutes {
		t.Fatalf("nil application default ttl = %d, want %d", nilApplication.GetMagicLinkExpireMinutes(), MagicLinkDefaultExpireMinutes)
	}
	application := &Application{}
	if application.GetMagicLinkExpireMinutes() != MagicLinkDefaultExpireMinutes {
		t.Fatalf("empty application default ttl = %d, want %d", application.GetMagicLinkExpireMinutes(), MagicLinkDefaultExpireMinutes)
	}
	application.MagicLinkExpireMinutes = 17
	if application.GetMagicLinkExpireMinutes() != 17 {
		t.Fatalf("configured ttl = %d, want 17", application.GetMagicLinkExpireMinutes())
	}
}
