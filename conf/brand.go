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

import (
	"fmt"
	"os"
)

// Branding is deployment configuration, not source code: every product name,
// wordmark and marketing URL a white-label deployment replaces is read from an
// environment variable here, and every default is the stock Casdoor value. A
// deployment sets the CASDOOR_BRAND_* variables (see docs/branding.md); the tree
// itself stays free of any downstream brand, which is what makes it mergeable
// with upstream.
const (
	BrandNameEnv           = "CASDOOR_BRAND_NAME"
	BrandTaglineEnv        = "CASDOOR_BRAND_TAGLINE"
	BrandLogoUrlEnv        = "CASDOOR_BRAND_LOGO_URL"
	BrandLogoDarkUrlEnv    = "CASDOOR_BRAND_LOGO_DARK_URL"
	BrandLogoMarkUrlEnv    = "CASDOOR_BRAND_LOGO_MARK_URL"
	BrandFaviconUrlEnv     = "CASDOOR_BRAND_FAVICON_URL"
	BrandWebsiteUrlEnv     = "CASDOOR_BRAND_WEBSITE_URL"
	BrandEmailSignatureEnv = "CASDOOR_BRAND_EMAIL_SIGNATURE"
	BrandTotpIssuerEnv     = "CASDOOR_BRAND_TOTP_ISSUER"
)

// Upstream defaults. Keep these in sync with what stock Casdoor ships, so that a
// build without any CASDOOR_BRAND_* variable is indistinguishable from upstream.
const (
	DefaultBrandName        = "Casdoor"
	DefaultBrandLogoUrl     = "https://cdn.casbin.org/img/casdoor-logo_1185x256.png"
	DefaultBrandFaviconUrl  = "https://cdn.casbin.org/img/favicon.png"
	DefaultBrandWebsiteUrl  = "https://casdoor.org"
	defaultBrandTaglineTmpl = "%s - sign in"
	defaultBrandSignTmpl    = "%s Team"
)

// getBrandString reads a branding variable; an unset or empty variable means "use
// the upstream default", so an env file can leave a value out without blanking it.
func getBrandString(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return defaultValue
}

// GetBrandName is the product name: window title, TOTP issuer fallback, the
// display name of the applications and organizations created on first run.
func GetBrandName() string {
	return getBrandString(BrandNameEnv, DefaultBrandName)
}

// GetBrandTagline is the <meta name="description"> of the login shell.
func GetBrandTagline() string {
	return getBrandString(BrandTaglineEnv, fmt.Sprintf(defaultBrandTaglineTmpl, GetBrandName()))
}

// GetBrandLogoUrl is the wordmark: the logo of the seeded application, the
// default logo of a new application in the console.
func GetBrandLogoUrl() string {
	return getBrandString(BrandLogoUrlEnv, DefaultBrandLogoUrl)
}

// GetBrandLogoDarkUrl is the wordmark painted on a dark background; empty means
// "the deployment has only one wordmark", and the light one is used.
func GetBrandLogoDarkUrl() string {
	return getBrandString(BrandLogoDarkUrlEnv, GetBrandLogoUrl())
}

// GetBrandLogoMarkUrl is the square mark embedded in the built-in email
// templates. A root-relative path is resolved against the frontend origin by the
// caller, because an email client cannot load a relative URL.
func GetBrandLogoMarkUrl() string {
	return getBrandString(BrandLogoMarkUrlEnv, GetBrandLogoUrl())
}

// GetBrandFaviconUrl is the favicon of the login shell and of the organization
// created on first run.
func GetBrandFaviconUrl() string {
	return getBrandString(BrandFaviconUrlEnv, DefaultBrandFaviconUrl)
}

// GetBrandWebsiteUrl is the product homepage: the homepage URL of the seeded
// application and the website URL of a new organization.
func GetBrandWebsiteUrl() string {
	return getBrandString(BrandWebsiteUrlEnv, DefaultBrandWebsiteUrl)
}

// GetBrandEmailSignature signs the built-in email templates ("... Team").
func GetBrandEmailSignature() string {
	return getBrandString(BrandEmailSignatureEnv, fmt.Sprintf(defaultBrandSignTmpl, GetBrandName()))
}

// GetBrandTotpIssuer is the issuer an authenticator app shows when the
// organization has no display name of its own.
func GetBrandTotpIssuer() string {
	return getBrandString(BrandTotpIssuerEnv, GetBrandName())
}
