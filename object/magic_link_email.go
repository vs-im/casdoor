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
	"strings"

	"github.com/casdoor/casdoor/conf"
)

// The branded default mails of the magic link API. "%link", "%expireTime" and the user
// placeholders are left in place, getApiMagicLinkEmailContent() fills them in.

func getDefaultEmailLogoURL() string {
	logoURL := conf.GetBrandLogoMarkUrl()
	if !strings.HasPrefix(logoURL, "/") {
		return logoURL
	}

	origin := strings.TrimRight(conf.GetConfigString("originFrontend"), "/")
	if origin == "" {
		origin = strings.TrimRight(conf.GetConfigString("origin"), "/")
	}
	return origin + logoURL
}

func GetDefaultMagicLinkEmailContent() string {
	logoURL := getDefaultEmailLogoURL()
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
      %signature
    </div>

    <div class="footer">
      Need help? Please contact your administrator.
    </div>
  </div>
</body>
</html>`
	content = strings.ReplaceAll(content, "%logoUrl", logoURL)
	return strings.ReplaceAll(content, "%signature", conf.GetBrandEmailSignature())
}

func GetDefaultMagicLinkSignupEmailContent() string {
	logoURL := getDefaultEmailLogoURL()
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
      %signature
    </div>

    <div class="footer">
      Need help? Please contact your administrator.
    </div>
  </div>
</body>
</html>`
	content = strings.ReplaceAll(content, "%logoUrl", logoURL)
	return strings.ReplaceAll(content, "%signature", conf.GetBrandEmailSignature())
}
