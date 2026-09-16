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

package conf

import "testing"

func TestBrandDefaultsAreUpstream(t *testing.T) {
	cases := []struct {
		name     string
		got      string
		expected string
	}{
		{"name", GetBrandName(), "Casdoor"},
		{"tagline", GetBrandTagline(), "Casdoor - sign in"},
		{"logo", GetBrandLogoUrl(), DefaultBrandLogoUrl},
		{"logoDark", GetBrandLogoDarkUrl(), DefaultBrandLogoUrl},
		{"logoMark", GetBrandLogoMarkUrl(), DefaultBrandLogoUrl},
		{"favicon", GetBrandFaviconUrl(), DefaultBrandFaviconUrl},
		{"website", GetBrandWebsiteUrl(), DefaultBrandWebsiteUrl},
		{"signature", GetBrandEmailSignature(), "Casdoor Team"},
		{"totpIssuer", GetBrandTotpIssuer(), "Casdoor"},
	}

	for _, c := range cases {
		if c.got != c.expected {
			t.Errorf("%s: got %q, want the upstream default %q", c.name, c.got, c.expected)
		}
	}
}

func TestBrandFromEnv(t *testing.T) {
	t.Setenv(BrandNameEnv, "Acme Identity")
	t.Setenv(BrandLogoUrlEnv, "/brand/logo.svg")
	t.Setenv(BrandFaviconUrlEnv, "/brand/favicon.png")
	t.Setenv(BrandWebsiteUrlEnv, "https://acme.example")

	// Derived values follow the name unless they are set explicitly.
	if got := GetBrandTagline(); got != "Acme Identity - sign in" {
		t.Errorf("tagline: got %q", got)
	}
	if got := GetBrandEmailSignature(); got != "Acme Identity Team" {
		t.Errorf("signature: got %q", got)
	}
	if got := GetBrandTotpIssuer(); got != "Acme Identity" {
		t.Errorf("totp issuer: got %q", got)
	}
	// The dark wordmark and the email mark fall back to the light one.
	if got := GetBrandLogoDarkUrl(); got != "/brand/logo.svg" {
		t.Errorf("dark logo: got %q", got)
	}
	if got := GetBrandLogoMarkUrl(); got != "/brand/logo.svg" {
		t.Errorf("logo mark: got %q", got)
	}

	config := GetWebConfig()
	if config.BrandName != "Acme Identity" || config.BrandFaviconUrl != "/brand/favicon.png" || config.BrandWebsiteUrl != "https://acme.example" {
		t.Errorf("the web config does not carry the branding: %+v", config)
	}
}
