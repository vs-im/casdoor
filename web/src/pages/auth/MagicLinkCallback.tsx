// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

import * as React from "react";
import {useLocation, useNavigate} from "react-router-dom";
import i18next from "i18next";
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {Loading} from "@/components/common/Loading";
import {useAccount} from "@/hooks/use-account";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Util from "@/auth/Util";
import * as Setting from "@/lib/setting";
import {
  getMagicLinkConsentApplication,
  getMagicLinkVerifyResult,
  isMagicLinkSignupAction,
} from "@/auth/magicLinkVerifyResult";

/**
 * The page a magic link lands on: the token from the mail is exchanged through
 * /api/verify-magic-link and the visitor is sent on the way the OAuth request
 * asked for — the redirect URI with a code, a token in the fragment, a form post,
 * the consent page, or the console for a plain login. Ported from
 * web-old/src/auth/MagicLinkCallback.js.
 *
 * routers/lightweight_auth_filter.go serves a static version of this page that
 * does the same without downloading the bundle; this one is what runs when that
 * page is unavailable (the Vite dev server, an older backend).
 */
export default function MagicLinkCallback() {
  const location = useLocation();
  const navigate = useNavigate();
  const {reload} = useAccount();
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    const params = new URLSearchParams(location.search);
    const token = params.get("token");
    if (!token) {
      setError(i18next.t("login:Magic link token is missing"));
      return;
    }

    const oAuthParams = Util.getOAuthGetParameters(params);
    AuthBackend.verifyMagicLink(token, oAuthParams)
      .then((res: any) => {
        if (res.status !== "ok") {
          setError(res.msg);
          return;
        }

        // the application's own frontend listens for these to run its onboarding
        const verifyResult = getMagicLinkVerifyResult(res);
        if (verifyResult.authAction || verifyResult.isNewUser) {
          sessionStorage.setItem("magicLinkVerifyResult", JSON.stringify(verifyResult));
          window.dispatchEvent(new CustomEvent("casdoor:magic-link-verify", {detail: verifyResult}));
          if (isMagicLinkSignupAction(verifyResult)) {
            window.dispatchEvent(new CustomEvent("casdoor:magic-link-signup-first-login", {detail: verifyResult}));
          }
        }

        const responseType = oAuthParams?.responseType || "login";
        // the backend asks for an explicit consent before handing out the code
        if (res.data?.required === true) {
          const consentApplication = getMagicLinkConsentApplication(res);
          params.delete("token");
          const consentPath = consentApplication ? `/consent/${consentApplication}` : "/consent";
          const consentQuery = params.toString();
          reload().then(() => navigate(consentQuery ? `${consentPath}?${consentQuery}` : consentPath));
          return;
        }

        if (responseType === "code") {
          const concatChar = oAuthParams.redirectUri?.includes("?") ? "&" : "?";
          Setting.goToLink(
            `${oAuthParams.redirectUri}${concatChar}code=${encodeURIComponent(res.data)}&state=${encodeURIComponent(oAuthParams.state)}`,
          );
          return;
        }

        const responseTypes = String(responseType).split(" ");
        if (responseTypes.includes("token") || responseTypes.includes("id_token")) {
          const responseMode = oAuthParams?.responseMode || "query";
          if (responseMode === "form_post") {
            Setting.createFormAndSubmit(oAuthParams.redirectUri, {
              token: responseTypes.includes("token") ? res.data : null,
              id_token: responseTypes.includes("id_token") ? res.data : null,
              token_type: "bearer",
              state: oAuthParams.state,
            });
          } else {
            const amendatoryResponseType = responseType === "token" ? "access_token" : responseType;
            Setting.goToLink(
              `${oAuthParams.redirectUri}#${amendatoryResponseType}=${encodeURIComponent(res.data)}&state=${encodeURIComponent(oAuthParams.state)}&token_type=bearer`,
            );
          }
          return;
        }

        Setting.showMessage("success", i18next.t("application:Logged in successfully"));
        // the console reads the account from context, so it has to be reloaded before the jump
        reload().then(() => Setting.goToLink("/"));
      })
      .catch((err) => {
        setError(`${i18next.t("general:Failed to connect to server")}${err}`);
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (error === null) {
    return <Loading className="min-h-screen" />;
  }

  return (
    <AuthLayout>
      <div className="space-y-4">
        <Alert variant="destructive">
          <AlertTitle>{i18next.t("login:Magic link sign-in failed")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
        <Button className="w-full" onClick={() => Setting.goToLink("/login")}>
          {i18next.t("login:Sign In")}
        </Button>
      </div>
    </AuthLayout>
  );
}
