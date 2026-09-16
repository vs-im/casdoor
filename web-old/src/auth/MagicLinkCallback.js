import React from "react";
import {Button, Result} from "antd";
import i18next from "i18next";
import * as AuthBackend from "./AuthBackend";
import * as Util from "./Util";
import * as Setting from "../Setting";
import {createFormAndSubmit} from "../Setting";
import {getMagicLinkConsentApplication, getMagicLinkVerifyResult, isMagicLinkSignupAction} from "./magicLinkVerifyResult";

class MagicLinkCallback extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      loading: true,
      error: "",
    };
  }

  componentDidMount() {
    const params = new URLSearchParams(this.props.location.search);
    const token = params.get("token");
    if (!token) {
      this.setState({loading: false, error: i18next.t("login:Magic link token is missing")});
      return;
    }

    const oAuthParams = Util.getOAuthGetParameters(params);
    AuthBackend.verifyMagicLink(token, oAuthParams)
      .then((res) => {
        if (res.status !== "ok") {
          this.setState({loading: false, error: res.msg});
          return;
        }

        const verifyResult = getMagicLinkVerifyResult(res);
        if (verifyResult.authAction || verifyResult.isNewUser) {
          sessionStorage.setItem("magicLinkVerifyResult", JSON.stringify(verifyResult));
          window.dispatchEvent(new CustomEvent("casdoor:magic-link-verify", {detail: verifyResult}));
          if (isMagicLinkSignupAction(verifyResult)) {
            window.dispatchEvent(new CustomEvent("casdoor:magic-link-signup-first-login", {detail: verifyResult}));
          }
        }

        const responseType = oAuthParams?.responseType || "login";
        if (res.data?.required === true) {
          const consentApplication = getMagicLinkConsentApplication(res);
          params.delete("token");
          const consentPath = consentApplication ? `/consent/${consentApplication}` : "/consent";
          const consentQuery = params.toString();
          Setting.goToLinkSoft(this, consentQuery ? `${consentPath}?${consentQuery}` : consentPath);
          return;
        }

        if (responseType === "code") {
          const concatChar = oAuthParams.redirectUri?.includes("?") ? "&" : "?";
          Setting.goToLink(`${oAuthParams.redirectUri}${concatChar}code=${res.data}&state=${oAuthParams.state}`);
          return;
        }

        if (responseType?.split(" ").includes("token") || responseType?.split(" ").includes("id_token")) {
          const responseTypes = responseType.split(" ");
          const responseMode = oAuthParams?.responseMode || "query";
          if (responseMode === "form_post") {
            createFormAndSubmit(oAuthParams.redirectUri, {
              token: responseTypes.includes("token") ? res.data : null,
              id_token: responseTypes.includes("id_token") ? res.data : null,
              token_type: "bearer",
              state: oAuthParams.state,
            });
          } else {
            const amendatoryResponseType = responseType === "token" ? "access_token" : responseType;
            Setting.goToLink(`${oAuthParams.redirectUri}#${amendatoryResponseType}=${res.data}&state=${oAuthParams.state}&token_type=bearer`);
          }
          return;
        }

        Setting.showMessage("success", i18next.t("application:Logged in successfully"));
        this.props.onLoginSuccess?.();
        Setting.goToLink("/");
      })
      .catch((error) => {
        this.setState({loading: false, error: `${i18next.t("general:Failed to connect to server")}${error}`});
      });
  }

  render() {
    if (this.state.loading) {
      return <Result status="info" title={i18next.t("login:Signing in...")} />;
    }

    return (
      <Result
        status="error"
        title={i18next.t("login:Magic link sign-in failed")}
        subTitle={this.state.error}
        extra={[
          <Button type="primary" key="login" onClick={() => Setting.goToLink("/login")}>
            {i18next.t("login:Sign In")}
          </Button>,
        ]}
      />
    );
  }
}

export default MagicLinkCallback;
