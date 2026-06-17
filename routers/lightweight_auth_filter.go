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

package routers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/util"
)

const (
	providerHintRedirectScriptName = "ProviderHintRedirect.js"
	authCallbackHandlerScriptName  = "AuthCallbackHandler.js"
)

func getLightweightAuthScriptPath(scriptName string) string {
	candidates := []string{
		filepath.Join(getWebBuildFolder(), scriptName),
	}

	if frontendBaseDir != "" {
		candidates = append(candidates,
			filepath.Join(frontendBaseDir, "public", scriptName),
			filepath.Join(filepath.Dir(frontendBaseDir), "public", scriptName),
		)
	}

	candidates = append(candidates, filepath.Join("web", "public", scriptName))

	for _, candidate := range candidates {
		if util.FileExist(candidate) {
			return candidate
		}
	}

	return ""
}

func serveLightweightAuthScript(ctx *context.Context, requestPath string, scriptName string) bool {
	if ctx.Request.URL.Path != requestPath {
		return false
	}

	scriptPath := getLightweightAuthScriptPath(scriptName)
	if scriptPath == "" {
		ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, scriptName, time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}

	f, err := os.Open(filepath.Clean(scriptPath))
	if err != nil {
		ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, scriptName, time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, scriptName, time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}

	ctx.Output.Header("Content-Type", "application/javascript; charset=utf-8")
	ctx.Output.Header("Cache-Control", "no-store")
	http.ServeContent(ctx.ResponseWriter, ctx.Request, fileInfo.Name(), fileInfo.ModTime(), f)
	return true
}

func serveProviderHintRedirectScript(ctx *context.Context) bool {
	return serveLightweightAuthScript(ctx, "/"+providerHintRedirectScriptName, providerHintRedirectScriptName)
}

func serveAuthCallbackHandlerScript(ctx *context.Context) bool {
	return serveLightweightAuthScript(ctx, "/"+authCallbackHandlerScriptName, authCallbackHandlerScriptName)
}

func serveProviderHintRedirectPage(ctx *context.Context) bool {
	if ctx.Request.URL.Path != "/login/oauth/authorize" {
		return false
	}

	providerHint := ctx.Input.Query("provider_hint")
	if providerHint == "" {
		return false
	}

	const providerHintRedirectHtml = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Redirecting...</title>
	<style>
		html, body {
			width: 100%;
			height: 100%;
			margin: 0;
			background: #ffffff;
			color: #1f2937;
			font-family: sans-serif;
		}

		body {
			display: flex;
			align-items: center;
			justify-content: center;
		}

		.redirecting {
			font-size: 14px;
			opacity: 0.72;
		}
	</style>
</head>
<body>
	<div class="redirecting">Redirecting...</div>
	<script src="/ProviderHintRedirect.js"></script>
	<script>
		(function() {
			function redirectToFallback() {
				var url = new URL(window.location.href);
				url.searchParams.delete("provider_hint");
				window.location.replace(url.pathname + url.search + url.hash);
			}

			if (!window.CasdoorProviderHintRedirect || typeof window.CasdoorProviderHintRedirect.run !== "function") {
				redirectToFallback();
				return;
			}

			window.CasdoorProviderHintRedirect.run();
		})();
	</script>
</body>
</html>
`

	err := util.AppendWebConfigCookie(ctx)
	if err != nil {
		logs.Error("AppendWebConfigCookie failed in serveProviderHintRedirectPage, error: %s", err)
	}

	ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Output.Header("Cache-Control", "no-store")
	http.ServeContent(ctx.ResponseWriter, ctx.Request, "provider-hint-redirect.html", time.Now(), strings.NewReader(providerHintRedirectHtml))
	return true
}

func serveAuthCallbackPage(ctx *context.Context) bool {
	if ctx.Request.URL.Path != "/callback" {
		return false
	}

	if ctx.Input.Query("__casdoor_callback_react") == "1" {
		return false
	}

	if ctx.Input.Query("state") == "" {
		return false
	}

	const authCallbackHtml = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Signing in...</title>
	<style>
		html, body {
			width: 100%;
			height: 100%;
			margin: 0;
			background: #ffffff;
			color: #1f2937;
			font-family: sans-serif;
		}

		body {
			display: flex;
			align-items: center;
			justify-content: center;
		}

		.callback-status {
			font-size: 14px;
			opacity: 0.82;
			padding: 0 24px;
			text-align: center;
		}
	</style>
</head>
<body>
	<div id="callback-status" class="callback-status">Signing in...</div>
	<script src="/AuthCallbackHandler.js"></script>
	<script>
		(function() {
			if (!window.CasdoorAuthCallback || typeof window.CasdoorAuthCallback.run !== "function") {
				document.getElementById("callback-status").textContent = "Failed to load callback handler.";
				return;
			}

			window.CasdoorAuthCallback.run();
		})();
	</script>
</body>
</html>
`

	err := util.AppendWebConfigCookie(ctx)
	if err != nil {
		logs.Error("AppendWebConfigCookie failed in serveAuthCallbackPage, error: %s", err)
	}

	ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Output.Header("Cache-Control", "no-store")
	http.ServeContent(ctx.ResponseWriter, ctx.Request, "auth-callback.html", time.Now(), strings.NewReader(authCallbackHtml))
	return true
}

func serveMagicLinkCallbackPage(ctx *context.Context) bool {
	if ctx.Request.URL.Path != "/magic-link/callback" {
		return false
	}

	const magicLinkCallbackHTML = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Signing in...</title>
	<style>
		html, body {
			width: 100%;
			height: 100%;
			margin: 0;
			background: #f6f7fb;
			color: #1a1a1a;
			font-family: Inter, Arial, sans-serif;
		}

		body {
			display: flex;
			align-items: center;
			justify-content: center;
			padding: 24px;
			box-sizing: border-box;
		}

		.card {
			width: 100%;
			max-width: 440px;
			background: #ffffff;
			border: 1px solid #e5e7eb;
			border-radius: 18px;
			box-shadow: 0 20px 50px rgba(15, 23, 42, 0.10);
			padding: 34px 30px;
			text-align: center;
		}

		.icon {
			width: 54px;
			height: 54px;
			border-radius: 999px;
			margin: 0 auto 20px;
			display: flex;
			align-items: center;
			justify-content: center;
			font-size: 28px;
			font-weight: 700;
			background: #eef2ff;
			color: #6366f1;
		}

		.card.error .icon {
			background: #fef2f2;
			color: #dc2626;
		}

		.card.success .icon {
			background: #ecfdf5;
			color: #059669;
		}

		h1 {
			font-size: 22px;
			line-height: 1.25;
			margin: 0 0 10px;
		}

		p {
			font-size: 14px;
			line-height: 1.6;
			color: #4b5563;
			margin: 0;
		}

		.action {
			display: inline-block;
			margin-top: 24px;
			padding: 11px 18px;
			border-radius: 10px;
			background: #6366f1;
			color: #ffffff;
			text-decoration: none;
			font-size: 14px;
			font-weight: 700;
		}
	</style>
</head>
<body>
	<div id="magic-link-card" class="card">
		<div id="magic-link-icon" class="icon">...</div>
		<h1 id="magic-link-title">Signing you in</h1>
		<p id="magic-link-message">Please wait while we verify your secure magic link.</p>
		<a id="magic-link-action" class="action" href="/" style="display: none;">Back to sign in</a>
	</div>
	<script>
		(function() {
			var card = document.getElementById("magic-link-card");
			var icon = document.getElementById("magic-link-icon");
			var title = document.getElementById("magic-link-title");
			var message = document.getElementById("magic-link-message");
			var action = document.getElementById("magic-link-action");
			var params = new URLSearchParams(window.location.search);
			var token = params.get("token") || "";
			function setState(type, iconText, titleText, messageText) {
				card.className = "card " + type;
				icon.textContent = iconText;
				title.textContent = titleText;
				message.textContent = messageText;
			}
			function fail(rawMessage) {
				var text = rawMessage || "Magic link sign-in failed.";
				var lower = text.toLowerCase();
				if (lower.indexOf("already used") !== -1) {
					setState("error", "!", "This magic link was already used", "For your security, each magic link can be used only once. Please request a new sign-in link.");
				} else if (lower.indexOf("expired") !== -1) {
					setState("error", "!", "This magic link has expired", "Please request a new sign-in link and use it before it expires.");
				} else if (lower.indexOf("revoked") !== -1) {
					setState("error", "!", "This magic link was revoked", "This sign-in link is no longer active. Please request a new one.");
				} else {
					setState("error", "!", "Magic link sign-in failed", text);
				}
				action.style.display = "inline-block";
			}
			function queryValue(name) {
				return params.get(name) || "";
			}
			function append(target, key, value) {
				if (value !== "") {
					target.set(key, value);
				}
			}
			function goTo(url) {
				setState("success", "OK", "Magic link verified", "Redirecting you to the application...");
				window.location.replace(url);
			}
			if (token === "") {
				fail("Magic link token is missing.");
				return;
			}
			var apiParams = new URLSearchParams();
			apiParams.set("token", token);
			append(apiParams, "clientId", queryValue("client_id"));
			append(apiParams, "responseType", queryValue("response_type"));
			append(apiParams, "redirectUri", queryValue("redirect_uri"));
			append(apiParams, "scope", queryValue("scope"));
			append(apiParams, "state", queryValue("state"));
			append(apiParams, "nonce", queryValue("nonce"));
			append(apiParams, "code_challenge_method", queryValue("code_challenge_method"));
			append(apiParams, "code_challenge", queryValue("code_challenge"));
			append(apiParams, "resource", queryValue("resource"));
			fetch("/api/verify-magic-link?" + apiParams.toString(), {
				method: "GET",
				credentials: "include",
				headers: {"Accept-Language": navigator.language || "en"}
			}).then(function(response) {
				return response.json();
			}).then(function(res) {
				if (!res || res.status !== "ok") {
					fail(res && res.msg ? res.msg : "Magic link sign-in failed.");
					return;
				}
				var responseType = queryValue("response_type") || "login";
				var redirectUri = queryValue("redirect_uri");
				var state = queryValue("state");
				if (res.data && res.data.required === true) {
					params.delete("token");
					goTo("/consent/" + encodeURIComponent(res.data2) + "?" + params.toString());
					return;
				}
				if (responseType === "code" && redirectUri !== "") {
					var separator = redirectUri.indexOf("?") === -1 ? "?" : "&";
					goTo(redirectUri + separator + "code=" + encodeURIComponent(res.data || "") + "&state=" + encodeURIComponent(state));
					return;
				}
				if ((responseType.split(" ").indexOf("token") !== -1 || responseType.split(" ").indexOf("id_token") !== -1) && redirectUri !== "") {
					var hashType = responseType === "token" ? "access_token" : responseType;
					goTo(redirectUri + "#" + hashType + "=" + encodeURIComponent(res.data || "") + "&state=" + encodeURIComponent(state) + "&token_type=bearer");
					return;
				}
				goTo("/");
			}).catch(function(error) {
				fail("Failed to connect to server: " + error);
			});
		})();
	</script>
</body>
</html>`

	err := util.AppendWebConfigCookie(ctx)
	if err != nil {
		logs.Error("AppendWebConfigCookie failed in serveMagicLinkCallbackPage, error: %s", err)
	}

	ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Output.Header("Cache-Control", "no-store")
	http.ServeContent(ctx.ResponseWriter, ctx.Request, "magic-link-callback.html", time.Now(), strings.NewReader(magicLinkCallbackHTML))
	return true
}
