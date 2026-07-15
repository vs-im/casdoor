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

import React from "react";
import {Button, Form, Input, Popover, Radio, Result, Row, Select, message} from "antd";
import * as Setting from "../Setting";
import * as AuthBackend from "./AuthBackend";
import * as ProviderButton from "./ProviderButton";
import i18next from "i18next";
import * as Util from "./Util";
import {authConfig} from "./Auth";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as AgreementModal from "../common/modal/AgreementModal";
import {SendCodeInput} from "../common/SendCodeInput";
import RegionSelect from "../common/select/RegionSelect";
import CustomGithubCorner from "../common/CustomGithubCorner";
import LanguageSelect from "../common/select/LanguageSelect";
import {withRouter} from "react-router-dom";
import {CountryCodeSelect} from "../common/select/CountryCodeSelect";
import * as PasswordChecker from "../common/PasswordChecker";
import * as InvitationBackend from "../backend/InvitationBackend";
import "./AuthButtons.css";
import {EmailInputGroup} from "../common/EmailInputGroup";
import {CaptchaModal} from "../common/modal/CaptchaModal";

const formItemLayout = {
  // labelCol: {
  //   xs: {
  //     span: 24,
  //   },
  //   sm: {
  //     span: 8,
  //   },
  // },
  // wrapperCol: {
  //   xs: {
  //     span: 24,
  //   },
  //   sm: {
  //     span: 16,
  //   },
  // },
};

const renderFormItem = (signupItem) => {
  const commonRules = [
    {
      required: signupItem.required,
      message: i18next.t("signup:Please input your {label}!").replace("{label}", signupItem.label || signupItem.name),
    },
  ];

  if (!signupItem.type || signupItem.type === "Input") {
    const inputRules = [...commonRules];
    if (signupItem.regex) {
      inputRules.push({
        pattern: new RegExp(signupItem.regex),
        message: i18next.t("signup:The input doesn't match the signup item regex!"),
      });
    }

    return (
      <Form.Item
        name={signupItem.name.toLowerCase()}
        label={signupItem.label || signupItem.name}
        rules={inputRules}
      >
        <Input placeholder={signupItem.placeholder} />
      </Form.Item>
    );
  } else if (signupItem.type === "Single Choice" || signupItem.type === "Multiple Choices") {
    return (
      <Form.Item
        name={signupItem.name.toLowerCase()}
        label={signupItem.label || signupItem.name}
        rules={commonRules}
      >
        <Select
          mode={signupItem.type === "Multiple Choices" ? "multiple" : "single"}
          placeholder={signupItem.placeholder}
          showSearch={false}
          options={signupItem.options.map(option => ({label: option, value: option}))}
        />
      </Form.Item>
    );
  }
};

export const tailFormItemLayout = {
  // wrapperCol: {
  //   xs: {
  //     span: 24,
  //     offset: 0,
  //   },
  //   sm: {
  //     span: 16, // 16
  //     offset: 8, // 8
  //   },
  // },
  style: {
    paddingTop: "15px",
    marginBottom: "0px",
  },
};

class SignupPage extends React.Component {
  constructor(props) {
    super(props);

    const urlParams = new URLSearchParams(props.location.search);
    const username = urlParams.get("username");
    const password = urlParams.get("password");

    this.state = {
      loading: false,
      password: password ?? "",
      confirmPassword: password ?? "",
      isPasswordDirty: false,
      isPasswordFocus: false,
      isConfirmPasswordDirty: false,
      classes: props,
      applicationName:
        props.applicationName ?? props.match?.params?.applicationName ?? null,
      email: username ?? "",
      phone: "",
      emailOrPhoneMode: "",
      countryCode: "",
      emailCode: "",
      phoneCode: "",
      validEmail: false,
      validPhone: false,
      region: "",
      isTermsOfUseVisible: false,
      termsOfUseContent: "",
      openCaptchaModal: false,
      magicLinkEmail: "",
      magicLinkSent: false,
      magicLinkLoading: false,
      captchaAction: "signup",
    };

    this.form = React.createRef();
  }

  componentDidMount() {
    const oAuthParams = Util.getOAuthGetParameters();
    if (oAuthParams !== null) {
      const signinUrl = window.location.pathname.replace(
        "/signup/oauth/authorize",
        "/login/oauth/authorize"
      );
      sessionStorage.setItem("signinUrl", signinUrl + window.location.search);
    }

    if (this.getApplicationObj() === undefined) {
      if (this.state.applicationName !== null) {
        this.getApplication(this.state.applicationName);
        this.setInvitationCode();
      } else if (oAuthParams !== null) {
        this.getApplicationLogin(oAuthParams);
      } else {
        Setting.showMessage("error", `${i18next.t("general:Unknown application name")}: ${this.state.applicationName}`);
        this.onUpdateApplication(null);
      }
    }
  }

  setInvitationCode(application = null) {
    const sp = new URLSearchParams(window.location.search);
    if (sp.has("invitationCode")) {
      const invitationCode = sp.get("invitationCode");
      this.setState({invitationCode: invitationCode});
      if (invitationCode !== "") {
        let appName = this.state.applicationName;
        if (application) {
          appName = application.name;
        }
        this.getInvitationCodeInfo(invitationCode, "admin/" + appName);
      }
    }
  }

  getApplication(applicationName) {
    if (applicationName === undefined) {
      return;
    }

    ApplicationBackend.getApplication("admin", applicationName).then((res) => {
      if (res.status === "error") {
        Setting.showMessage("error", res.msg);
        return;
      }

      this.onUpdateApplication(res.data);
    });
  }

  getApplicationLogin(oAuthParams) {
    AuthBackend.getApplicationLogin(oAuthParams)
      .then((res) => {
        if (res.status === "ok") {
          const application = res.data;
          this.onUpdateApplication(application);
          this.setInvitationCode(application);
        } else {
          this.onUpdateApplication(null);
          this.setState({
            msg: res.msg,
          });
        }
      });
  }

  getInvitationCodeInfo(invitationCode, application) {
    InvitationBackend.getInvitationCodeInfo(invitationCode, application).then(
      (res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }
        this.setState({invitation: res.data});
        if (res.data.email) {
          this.setState({validEmail: true, email: res.data.email});
        }
        if (res.data.phone) {
          this.setState({validPhone: true, phone: res.data.phone});
        }
      });
  }

  getResultPath(application, signupParams) {
    if (signupParams?.plan && signupParams?.pricing) {
      // the prompt page needs the user to be signed in, so for paid-user sign up, just go to buy-plan page
      return `/buy-plan/${application.organization}/${signupParams?.pricing}?user=${signupParams.username}&plan=${signupParams.plan}`;
    }
    if (authConfig.appName === application.name) {
      return "/result";
    } else {
      const oAuthParams = Util.getOAuthGetParameters();
      if (Setting.hasPromptPage(application)) {
        return `/prompt/${application.name}?oauth=${oAuthParams !== null}`;
      } else {
        return `/result/${application.name}`;
      }
    }
  }

  getApplicationObj() {
    return this.props.application;
  }

  onUpdateAccount(account) {
    this.props.onUpdateAccount(account);
  }

  onUpdateApplication(application) {
    this.props.onUpdateApplication(application);
  }

  parseOffset(offset) {
    if (
      offset === 2 ||
      offset === 4 ||
      Setting.inIframe() ||
      Setting.isMobile()
    ) {
      return "0 auto";
    }
    if (offset === 1) {
      return "0 10%";
    }
    if (offset === 3) {
      return "0 60%";
    }
  }

  getLanguagesItem(application) {
    return application.signupItems?.find((item) => item.name === "Languages");
  }

  renderLanguageSelect(application) {
    const languagesItem = this.getLanguagesItem(application);
    if (languagesItem && !languagesItem.visible) {
      return null;
    }

    const languages = application.organizationObj.languages;
    if (languages && languages.length <= 1) {
      const language = (languages.length === 1) ? languages[0] : "en";
      if (Setting.getLanguage() !== language) {
        Setting.setLanguage(language);
      }
      return null;
    }
    return (
      <div className="signup-languages">
        {languagesItem?.customCss && <div dangerouslySetInnerHTML={{__html: ("<style>" + languagesItem.customCss.replaceAll("<style>", "").replaceAll("</style>", "") + "</style>")}} />}
        <LanguageSelect
          languages={languages}
          mode={languagesItem?.rule}
          style={{top: "55px", right: "5px", position: "absolute"}}
        />
      </div>
    );
  }

  checkCaptchaStatus(values) {
    AuthBackend.getCaptchaStatus(values)
      .then((res) => {
        if (res.status === "ok") {
          if (res.data) {
            this.setState({
              openCaptchaModal: true,
              values: values,
            });
            return null;
          }
        }
        this.submitSignup(values);
      });
  }

  renderCaptchaModal(application) {
    if (Setting.getCaptchaRule(application) === Setting.CaptchaRule.Never) {
      return null;
    }
    const captchaProviderItems = Setting.getCaptchaProviderItems(application);
    const captchaRule = Setting.getCaptchaRule(application);
    let provider = null;

    const ruleProviders = captchaProviderItems.filter(providerItem => providerItem.rule === captchaRule);
    if (ruleProviders.length > 0) {
      provider = ruleProviders[0].provider;
    }

    if (!provider) {
      return null;
    }

    return <CaptchaModal
      owner={provider.owner}
      name={provider.name}
      visible={this.state.openCaptchaModal}
      onOk={(captchaType, captchaToken, clientSecret) => {
        const values = {
          ...(this.state.values || {}),
          captchaType: captchaType,
          captchaToken: captchaToken,
          clientSecret: clientSecret,
        };

        if (this.state.captchaAction === "magicLink") {
          this.submitMagicLink(values);
        } else {
          this.submitSignup(values);
        }
        this.setState({openCaptchaModal: false, captchaAction: "signup"});
      }}
      onCancel={() => this.setState({openCaptchaModal: false})}
      isCurrentProvider={true}
    />;
  }

  submitMagicLink(values) {
    const application = this.getApplicationObj();
    const oAuthParams = Util.getOAuthGetParameters();
    const payload = {
      email: (values.magicLinkEmail || this.state.magicLinkEmail || "").trim(),
      organization: application.organization,
      application: application.name,
    };

    if (values.captchaType) {
      payload.captchaType = values.captchaType;
    }
    if (values.captchaToken) {
      payload.captchaToken = values.captchaToken;
    }
    if (values.clientSecret) {
      payload.clientSecret = values.clientSecret;
    }

    if (!Setting.isValidEmail(payload.email)) {
      Setting.showMessage("error", i18next.t("login:The input is not valid Email!"));
      return;
    }

    this.setState({magicLinkLoading: true});
    AuthBackend.sendMagicLink(payload, oAuthParams)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({magicLinkSent: true});
        } else if (res.data === "captchaRequired") {
          this.setState({
            openCaptchaModal: true,
            captchaAction: "magicLink",
            values: {magicLinkEmail: payload.email},
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}${error}`);
      })
      .finally(() => {
        this.setState({magicLinkLoading: false});
      });
  }

  onFinish(values) {
    const application = this.getApplicationObj();

    let codeRequired = false;
    for (const signupItem of application.signupItems) {
      if (signupItem.name === "Email") {
        codeRequired = signupItem.rule !== "No verification";
      }
    }

    if (codeRequired && !values.emailCode) {
      Setting.showMessage("error", i18next.t("code:Send Code"));
      return;
    }

    if (Array.isArray(values.gender)) {
      values.gender = values.gender.join(", ");
    }

    if (Array.isArray(values.bio)) {
      values.bio = values.bio.join(", ");
    }

    if (Array.isArray(values.tag)) {
      values.tag = values.tag.join(", ");
    }

    if (Array.isArray(values.education)) {
      values.education = values.education.join(", ");
    }

    if (this.state.invitationCode && !values.invitationCode) {
      values.invitationCode = this.state.invitationCode;
    }

    const params = new URLSearchParams(window.location.search);
    values.plan = params.get("plan");
    values.pricing = params.get("pricing");

    const captchaRule = Setting.getCaptchaRule(application);
    if (captchaRule === Setting.CaptchaRule.Always) {
      this.setState({
        openCaptchaModal: true,
        values: values,
      });
      return;
    } else if (captchaRule === Setting.CaptchaRule.Dynamic || captchaRule === Setting.CaptchaRule.InternetOnly) {
      this.checkCaptchaStatus(values);
      return;
    }

    this.submitSignup(values);
  }

  submitSignup(values) {
    const application = this.getApplicationObj();

    // Get OAuth parameters if present
    const oAuthParams = Util.getOAuthGetParameters();
    this.setState({
      loading: true,
    });
    if (!values.email) {
      values.email = this.state.email || "";
    }

    AuthBackend.signup(values, oAuthParams)
      .then((res) => {
        if (res.status === "ok") {
          // Check if this is OAuth flow with code response
          // When OAuth parameters are present and code is returned, it won't contain '/'
          if (oAuthParams && res.data && typeof res.data === "string" && !res.data.includes("/")) {
            // OAuth code returned, redirect to redirect_uri with code
            const code = res.data;
            const redirectUrl = `${oAuthParams.redirectUri}${oAuthParams.redirectUri.includes("?") ? "&" : "?"}code=${code}&state=${oAuthParams.state}`;
            Setting.goToLink(redirectUrl);
            return;
          }

          // Check if consent is required
          if (oAuthParams && res.data && typeof res.data === "object" && res.data.required === true) {
            // Consent required, redirect to consent page
            Setting.goToLink(`/consent/${application.name}?${window.location.search.substring(1)}`);
            return;
          }

          // the user's id will be returned by `signup()`, if user signup by phone, the `username` in `values` is undefined.
          if (typeof res.data === "string") {
            values.username = res.data.split("/")[1];
          }
          if (Setting.hasPromptPage(application) && (!values.plan || !values.pricing)) {
            AuthBackend.getAccount("")
              .then((res) => {
                let account = null;
                if (res.status === "ok") {
                  account = res.data;
                  account.organization = res.data2;

                  this.onUpdateAccount(account);
                  Setting.goToLinkSoft(
                    this,
                    this.getResultPath(application, values)
                  );
                } else {
                  Setting.showMessage(
                    "error",
                    `${i18next.t("application:Failed to sign in")}: ${res.msg}`
                  );
                }
              });
          } else {
            Setting.goToLinkSoft(this, this.getResultPath(application, values));
          }
        } else {
          Setting.showMessage("error", res.msg);
        }
      }).finally(() => {
        this.setState({
          loading: false,
        });
      });
  }

  onFinishFailed(values, errorFields, outOfDate) {
    this.form.current.scrollToField(errorFields[0].name);
  }

  isProviderVisible(providerItem) {
    return Setting.isProviderVisibleForSignUp(providerItem);
  }

  isSignupSubmitItem(signupItem) {
    if (signupItem?.visible === false) {
      return false;
    }
    if (signupItem?.name?.startsWith("Text ")) {
      return false;
    }
    return !["Signup title", "Signup button", "Magic link", "Providers", "Languages", "Agreement"].includes(signupItem?.name);
  }

  shouldRenderSignupButton(signupItems) {
    const signupButtonItem = signupItems.find(signupItem => signupItem.name === "Signup button");
    if (signupButtonItem?.visible === false) {
      return false;
    }
    return signupItems.some(signupItem => this.isSignupSubmitItem(signupItem));
  }

  renderFormItem(application, signupItem) {
    const validItems = ["Gender", "Bio", "Tag", "Education"];
    if (signupItem.name === "Signup title") {
      return (
        <div className="form-header">
          {signupItem.visible ? <span>{signupItem.label || i18next.t("account:Sign Up")}</span> : null}
        </div>
      );
    }
    if (!signupItem.visible) {
      return null;
    }

    const required = signupItem.required;

    if (signupItem.name === "Username") {
      const usernameRules = [
        {
          required: required,
          message: i18next.t("forget:Please input your username!"),
          whitespace: true,
        },
      ];
      if (signupItem.regex) {
        usernameRules.push({
          pattern: new RegExp(signupItem.regex),
          message: i18next.t("signup:The input doesn't match the signup item regex!"),
        });
      }
      return (
        <Form.Item
          key="username"
          name="username"
          className="signup-username"
          label={signupItem.label ? signupItem.label : i18next.t("signup:Username")}
          rules={usernameRules}
        >
          <Input className="signup-username-input" placeholder={signupItem.placeholder}
            disabled={this.state.invitation !== undefined && this.state.invitation.username !== ""} />
        </Form.Item>
      );
    } else if (signupItem.name === "Display name") {
      if (signupItem.rule === "First, last" && Setting.getLanguage() !== "zh") {
        const firstNameRules = [
          {
            required: required,
            message: i18next.t("signup:Please input your first name!"),
            whitespace: true,
          },
        ];
        const lastNameRules = [
          {
            required: required,
            message: i18next.t("signup:Please input your last name!"),
            whitespace: true,
          },
        ];
        if (signupItem.regex) {
          const regexRule = {
            pattern: new RegExp(signupItem.regex),
            message: i18next.t("signup:The input doesn't match the signup item regex!"),
          };
          firstNameRules.push(regexRule);
          lastNameRules.push(regexRule);
        }
        return (
          <React.Fragment key="firstName">
            <Form.Item
              name="firstName"
              className="signup-first-name"
              label={
                signupItem.label
                  ? signupItem.label
                  : i18next.t("general:First name")
              }
              rules={firstNameRules}
            >
              <Input
                className="signup-first-name-input"
                placeholder={signupItem.placeholder}
              />
            </Form.Item>
            <Form.Item
              name="lastName"
              className="signup-last-name"
              label={signupItem.label ? signupItem.label : i18next.t("general:Last name")}
              rules={lastNameRules}
            >
              <Input
                className="signup-last-name-input"
                placeholder={signupItem.placeholder}
              />
            </Form.Item>
          </React.Fragment>
        );
      }

      const displayNameRules = [
        {
          required: required,
          message: (signupItem.rule === "Real name" || signupItem.rule === "First, last") ? i18next.t("signup:Please input your real name!") : i18next.t("signup:Please input your display name!"),
          whitespace: true,
        },
      ];

      if (signupItem.regex) {
        displayNameRules.push({
          pattern: new RegExp(signupItem.regex),
          message: i18next.t("signup:The input doesn't match the signup item regex!"),
        });
      }

      return (
        <Form.Item
          name="name"
          key="name"
          className="signup-name"
          label={(signupItem.label ? signupItem.label : (signupItem.rule === "Real name" || signupItem.rule === "First, last") ? i18next.t("application:Real name") : i18next.t("general:Display name"))}
          rules={displayNameRules}
        >
          <Input className="signup-name-input" placeholder={signupItem.placeholder} />
        </Form.Item>
      );
    } else if (signupItem.name === "First name" && this.state?.displayNameRule !== "First, last") {
      const firstNameRules = [
        {
          required: required,
          message: i18next.t("signup:Please input your first name!"),
          whitespace: true,
        },
      ];
      if (signupItem.regex) {
        firstNameRules.push({
          pattern: new RegExp(signupItem.regex),
          message: i18next.t("signup:The input doesn't match the signup item regex!"),
        });
      }
      return (
        <Form.Item
          name="firstName"
          className="signup-first-name"
          label={signupItem.label ? signupItem.label : i18next.t("general:First name")}
          rules={firstNameRules}
        >
          <Input className="signup-first-name-input" placeholder={signupItem.placeholder} />
        </Form.Item>
      );
    } else if (signupItem.name === "Last name" && this.state?.displayNameRule !== "First, last") {
      const lastNameRules = [
        {
          required: required,
          message: i18next.t("signup:Please input your last name!"),
          whitespace: true,
        },
      ];
      if (signupItem.regex) {
        lastNameRules.push({
          pattern: new RegExp(signupItem.regex),
          message: i18next.t("signup:The input doesn't match the signup item regex!"),
        });
      }
      return (
        <Form.Item
          name="lastName"
          className="signup-last-name"
          label={signupItem.label ? signupItem.label : i18next.t("general:Last name")}
          rules={lastNameRules}
        >
          <Input className="signup-last-name-input" placeholder={signupItem.placeholder} />
        </Form.Item>
      );
    } else if (signupItem.name === "Affiliation") {
      const affiliationRules = [
        {
          required: required,
          message: i18next.t("signup:Please input your affiliation!"),
          whitespace: true,
        },
      ];
      if (signupItem.regex) {
        affiliationRules.push({
          pattern: new RegExp(signupItem.regex),
          message: i18next.t("signup:The input doesn't match the signup item regex!"),
        });
      }
      return (
        <Form.Item
          key="affiliation"
          name="affiliation"
          className="signup-affiliation"
          label={signupItem.label ? signupItem.label : i18next.t("user:Affiliation")}
          rules={affiliationRules}
        >
          <Input className="signup-affiliation-input" placeholder={signupItem.placeholder} />
        </Form.Item>
      );
    } else if (signupItem.name === "ID card") {
      return (
        <Form.Item
          key="idCard"
          name="idCard"
          className="signup-idcard"
          label={
            signupItem.label ? signupItem.label : i18next.t("user:ID card")
          }
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please input your ID card number!"),
              whitespace: true,
            },
            {
              required: required,
              pattern: new RegExp(
                /^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9X]$/,
                "g"
              ),
              message: i18next.t(
                "signup:Please input the correct ID card number!"
              ),
            },
          ]}
        >
          <Input
            className="signup-idcard-input"
            placeholder={signupItem.placeholder}
          />
        </Form.Item>
      );
    } else if (signupItem.name === "Country/Region") {
      return (
        <Form.Item
          key="country_region"
          name="country_region"
          className="signup-country-region"
          label={
            signupItem.label
              ? signupItem.label
              : i18next.t("user:Country/Region")
          }
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please select your country/region!"),
            },
          ]}
        >
          <RegionSelect className="signup-region-select" onChange={(value) => {
            this.setState({region: value});
          }} />
        </Form.Item>
      );
    } else if (signupItem.name === "Tag") {
      return (
        <Form.Item
          name="tag"
          className="signup-tag"
          label={signupItem.label ? signupItem.label : i18next.t("general:Tag")}
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please select your tag!"),
            },
          ]}
        >
          <Select
            className="signup-tag-select"
            placeholder={signupItem.placeholder || i18next.t("signup:Please select your tag!")}
            allowClear={!required}
          >
            {
              (signupItem.options?.length > 0 ? signupItem.options : application.tags ?? []).map((tag, index) => (
                <Select.Option key={index} value={tag}>{tag}</Select.Option>
              ))
            }
          </Select>
        </Form.Item>
      );
    } else if (signupItem.name === "Email" || signupItem.name === "Phone" || signupItem.name === "Email or Phone" || signupItem.name === "Phone or Email") {
      const renderPhoneItem = () => {
        return (
          <React.Fragment>
            <Form.Item
              className="signup-phone"
              label={
                signupItem.label ? signupItem.label : i18next.t("general:Phone")
              }
              required={required}
            >
              <Input.Group compact>
                <Form.Item
                  name="countryCode"
                  noStyle
                  rules={[
                    {
                      required: required,
                      message: i18next.t(
                        "signup:Please select your country code!"
                      ),
                    },
                  ]}
                >
                  <CountryCodeSelect
                    style={{width: "35%"}}
                    countryCodes={
                      this.getApplicationObj().organizationObj.countryCodes
                    }
                  />
                </Form.Item>
                <Form.Item
                  name="phone"
                  dependencies={["countryCode"]}
                  noStyle
                  rules={[
                    {
                      required: required,
                      message: i18next.t(
                        "signup:Please input your phone number!"
                      ),
                    },
                    ({getFieldValue}) => ({
                      validator: (_, value) => {
                        if (!required && !value) {
                          return Promise.resolve();
                        }

                        if (
                          value &&
                          !Setting.isValidPhone(
                            value,
                            getFieldValue("countryCode")
                          )
                        ) {
                          this.setState({validPhone: false});
                          return Promise.reject(
                            i18next.t("signup:The input is not valid Phone!")
                          );
                        }

                        this.setState({validPhone: true});
                        return Promise.resolve();
                      },
                    }),
                  ]}
                >
                  <Input
                    className="signup-phone-input"
                    placeholder={signupItem.placeholder}
                    style={{width: "65%", minHeight: "40px"}}
                    disabled={
                      this.state.invitation !== undefined &&
                      this.state.invitation.phone !== ""
                    }
                    onChange={(e) => this.setState({phone: e.target.value})}
                  />
                </Form.Item>
              </Input.Group>
            </Form.Item>
            {signupItem.rule !== "No verification" && (
              <Form.Item
                name="phoneCode"
                className="phone-code"
                label={
                  signupItem.label
                    ? signupItem.label
                    : i18next.t("code:Phone code")
                }
                rules={[
                  {
                    required: required,
                    message: i18next.t(
                      "code:Please input your phone verification code!"
                    ),
                  },
                ]}
              >
                <SendCodeInput
                  className="signup-phone-code-input"
                  disabled={!this.state.validPhone}
                  method={"signup"}
                  onButtonClickArgs={[
                    this.state.phone,
                    "phone",
                    Setting.getApplicationName(application),
                  ]}
                  application={application}
                  countryCode={this.form.current?.getFieldValue("countryCode")}
                />
              </Form.Item>
            )}
          </React.Fragment>
        );
      };

      if (signupItem.name === "Email") {
        return <EmailInputGroup
          email={this.state.email}
          required={required}
          validEmail={this.state.validEmail}
          invitation={this.state.invitation}
          signupItem={signupItem}
          application={application}
          setState={(...args) => this.setState(...args)}
        />;
      } else if (signupItem.name === "Phone") {
        return renderPhoneItem();
      } else if (
        signupItem.name === "Email or Phone" ||
        signupItem.name === "Phone or Email"
      ) {
        let emailOrPhoneMode = this.state.emailOrPhoneMode;
        if (emailOrPhoneMode === "") {
          emailOrPhoneMode =
            signupItem.name === "Email or Phone" ? "Email" : "Phone";
        }

        return (
          <React.Fragment>
            <Row style={{marginTop: "30px", marginBottom: "20px"}}>
              <Radio.Group
                style={{width: "400px"}}
                buttonStyle="solid"
                onChange={(e) => {
                  this.setState({
                    emailOrPhoneMode: e.target.value,
                  });
                }}
                value={emailOrPhoneMode}
              >
                {signupItem.name === "Email or Phone" ? (
                  <React.Fragment>
                    <Radio.Button value={"Email"}>
                      {i18next.t("general:Email")}
                    </Radio.Button>
                    <Radio.Button value={"Phone"}>
                      {i18next.t("general:Phone")}
                    </Radio.Button>
                  </React.Fragment>
                ) : (
                  <React.Fragment>
                    <Radio.Button value={"Phone"}>
                      {i18next.t("general:Phone")}
                    </Radio.Button>
                    <Radio.Button value={"Email"}>
                      {i18next.t("general:Email")}
                    </Radio.Button>
                  </React.Fragment>
                )}
              </Radio.Group>
            </Row>
            {emailOrPhoneMode === "Email"
              ? <EmailInputGroup
                email={this.state.email}
                required={required}
                validEmail={this.state.validEmail}
                invitation={this.state.invitation}
                signupItem={signupItem}
                application={application}
                setState={(...args) => this.setState(...args)}
              />
              : renderPhoneItem()}
          </React.Fragment>
        );
      } else {
        return null;
      }
    } else if (signupItem.name === "Password") {
      return (
        <Popover placement={"top"} content={this.state.passwordPopover} open={this.state.passwordPopoverOpen}>
          <Form.Item
            key="password"
            name="password"
            className="signup-password"
            label={
              signupItem.label ? signupItem.label : i18next.t("general:Password")
            }
            rules={[
              {
                required: required,
                validateTrigger: "onChange",
                validator: (_, value) => {
                  const [errorMsg] = PasswordChecker.checkPasswordComplexity(
                    value,
                    application.organizationObj.passwordOptions
                  );
                  if (errorMsg === "") {
                    return Promise.resolve();
                  } else {
                    return Promise.reject(errorMsg);
                  }
                },
              },
            ]}
            hasFeedback
          >
            <Input.Password
              className="signup-password-input"
              placeholder={signupItem.placeholder}
              autoComplete="new-password"
              onChange={(e) => {
                this.setState({
                  passwordPopover: PasswordChecker.renderPasswordPopover(application.organizationObj.passwordOptions, e.target.value),
                  password: e.target.value,
                  isPasswordDirty: true,
                });
              }}
              onFocus={() => {
                this.setState({
                  isPasswordFocus: true,
                  passwordPopoverOpen: application.organizationObj.passwordOptions?.length > 0,
                  passwordPopover: PasswordChecker.renderPasswordPopover(application.organizationObj.passwordOptions, this.form.current?.getFieldValue("password") ?? ""),
                });
              }}
              onBlur={() => {
                this.setState({
                  passwordPopoverOpen: false,
                  isPasswordFocus: false,
                });
              }}
            />
          </Form.Item>
        </Popover>
      );
    } else if (signupItem.name === "Confirm password") {
      return (
        <>
          <Form.Item
            key="confirm-password"
            name="confirm"
            className="signup-confirm"
            label={
              signupItem.label ? signupItem.label : i18next.t("general:Confirm")
            }
            dependencies={["password"]}
            hasFeedback
            onChange={(e) => {
              this.setState({confirmPassword: e.target.value, isConfirmPasswordDirty: true});
            }}
            rules={[
              {
                required: required,
                message: i18next.t("signup:Please confirm your password!"),
              },
              ({getFieldValue}) => ({
                validator(rule, value) {
                  if (!value || getFieldValue("password") === value) {
                    return Promise.resolve();
                  }

                  return Promise.reject(i18next.t(
                    "signup:Your confirmed password is inconsistent with the password!"
                  ));
                },
              }),
            ]}
          >
            <Input.Password placeholder={signupItem.placeholder} autoComplete="new-password" />
          </Form.Item>
        </>
      );
    } else if (signupItem.name === "Invitation code") {
      return (
        <Form.Item
          key="invitation-code"
          name="invitationCode"
          className="signup-invitation-code"
          label={
            signupItem.label
              ? signupItem.label
              : i18next.t("application:Invitation code")
          }
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please input your invitation code!"),
            },
          ]}
        >
          <Input
            className="signup-invitation-code-input"
            placeholder={signupItem.placeholder}
            disabled={
              this.state.invitation !== undefined &&
              this.state.invitation !== ""
            }
          />
        </Form.Item>
      );
    } else if (signupItem.name === "Agreement") {
      return AgreementModal.renderAgreementFormItem(
        application,
        required,
        tailFormItemLayout,
        this,
        "right"
      );
    } else if (signupItem.name.startsWith("Text ")) {
      return (
        <div
          key="label"
          dangerouslySetInnerHTML={{__html: signupItem.label}}
        />
      );
    } else if (signupItem.name === "Providers") {
      const a = 1;
      if (5 > a) {
        return null;
      }
      const showForm =
        Setting.isPasswordEnabled(application) ||
        Setting.isCodeSigninEnabled(application) ||
        Setting.isWebAuthnEnabled(application) ||
        Setting.isLdapEnabled(application);
      if (signupItem.rule === "None" || signupItem.rule === "") {
        signupItem.rule = showForm ? "small" : "big";
      }
      return application.providers
        .filter((providerItem) => this.isProviderVisible(providerItem))
        .map((providerItem, id) => {
          return (
            <span
              key={id}
              onClick={(e) => {
                const agreementChecked =
                  this.form.current.getFieldValue("agreement");

                if (
                  agreementChecked !== undefined &&
                  typeof agreementChecked === "boolean" &&
                  !agreementChecked
                ) {
                  e.preventDefault();
                  message.error(
                    i18next.t("signup:Please accept the agreement!")
                  );
                }
              }}
            >
              {ProviderButton.renderProviderLogo(
                providerItem.provider,
                application,
                null,
                null,
                signupItem.rule,
                this.props.location
              )}
            </span>
          );
        });
    } else if (signupItem.name === "Magic link") {
      if (!Setting.isMagicLinkEnabled(application) || !application.enableMagicLinkSignup) {
        return null;
      }
      return (
        <div key="magic-link" className="signup-magic-link" style={{marginBottom: "16px"}}>
          {
            signupItem.label ? (
              <div style={{fontWeight: 600, marginBottom: "8px"}}>
                {signupItem.label}
              </div>
            ) : null
          }
          <Form.Item
            name="magicLinkEmail"
            style={{marginBottom: "12px"}}
            rules={[
              {
                required: true,
                message: i18next.t("login:Please input your Email!"),
              },
              {
                validator: (_, value) => {
                  if (!value || Setting.isValidEmail(value)) {
                    return Promise.resolve();
                  }
                  return Promise.reject(i18next.t("login:The input is not valid Email!"));
                },
              },
            ]}
          >
            <Input
              placeholder={signupItem.placeholder || i18next.t("general:Email")}
              onChange={(e) => {
                this.setState({magicLinkEmail: e.target.value});
              }}
              onPressEnter={(e) => {
                e.preventDefault();
                this.form.current.validateFields(["magicLinkEmail"]).then((fields) => {
                  this.submitMagicLink(fields);
                });
              }}
            />
          </Form.Item>
          <Button
            htmlType="button"
            type="primary"
            block
            loading={this.state.magicLinkLoading}
            onClick={() => {
              this.form.current.validateFields(["magicLinkEmail"]).then((fields) => {
                this.submitMagicLink(fields);
              });
            }}
          >
            {i18next.t("login:Send Magic Link")}
          </Button>
        </div>
      );
    } else if (signupItem.name === "Signup button") {
      return null;
    } else if (validItems.includes(signupItem.name)) {
      return renderFormItem(signupItem);
    }
  }

  renderForm(application) {
    const {borderRadius} = Setting.getThemeData();
    if (!application.enableSignUp) {
      return (
        <Result
          status="error"
          title={i18next.t("application:Sign Up Error")}
          subTitle={i18next.t(
            "application:The application does not allow to sign up new account"
          )}
          extra={[
            <Button
              style={{borderRadius}}
              type="primary"
              key="signin"
              onClick={() =>
                Setting.redirectToLoginPage(application, this.props.history)
              }
            >
              {i18next.t("login:Sign In")}
            </Button>,
          ]}
        />
      );
    }
    if (this.state.magicLinkSent) {
      return (
        <Result
          status="success"
          title={i18next.t("login:Check your email")}
          subTitle={i18next.t("login:We sent you a magic link to sign in")}
          extra={[
            <Button
              style={{borderRadius}}
              type="primary"
              key="signin"
              onClick={() =>
                Setting.redirectToLoginPage(application, this.props.history)
              }
            >
              {i18next.t("login:Sign In")}
            </Button>,
          ]}
        />
      );
    }
    if (this.state.invitation !== undefined) {
      if (this.state.invitation.username !== "") {
        this.form.current?.setFieldValue(
          "username",
          this.state.invitation.username
        );
      }
      if (this.state.invitation.email !== "") {
        this.form.current?.setFieldValue("email", this.state.invitation.email);
      }
      if (this.state.invitation.phone !== "") {
        this.form.current?.setFieldValue("phone", this.state.invitation.phone);
      }
      if (this.state.invitationCode !== "") {
        this.form.current?.setFieldValue(
          "invitationCode",
          this.state.invitationCode
        );
      }
    }

    const displayNameItem = application.signupItems?.find(item => item.name === "Display name");
    if (displayNameItem && !this.state.displayNameRule) {
      this.setState({displayNameRule: displayNameItem.rule});
    }

    const avaliableProviders = application.providers.filter((providerItem) =>
      this.isProviderVisible(providerItem)
    );
    const showProviders = avaliableProviders.length > 0;

    const signupItems = Array.isArray(application.signupItems) ? [...application.signupItems] : [];
    const showSignupButton = this.shouldRenderSignupButton(signupItems);

    return (
      <Form
        id={"SignupPage-renderForm"}
        {...formItemLayout}
        ref={this.form}
        name="signup"
        onFinish={(values) => this.onFinish(values)}
        onFinishFailed={(errorInfo) =>
          this.onFinishFailed(
            errorInfo.values,
            errorInfo.errorFields,
            errorInfo.outOfDate
          )
        }
        initialValues={{
          application: application.name,
          organization: application.organization,
          countryCode: application.organizationObj.countryCodes?.[0],
          size: "default",
        }}
        size="large"
        // layout={Setting.isMobile() ? "vertical" : "horizontal"}
        layout={"vertical"}
        style={{width: "100%"}}
        // style={{width: Setting.isMobile() ? "300px" : "400px"}}
      >
        <Form.Item
          name="application"
          hidden={true}
          rules={[
            {
              required: true,
              message: "Please input your application!",
            },
          ]}
        ></Form.Item>
        <Form.Item
          name="organization"
          hidden={true}
          rules={[
            {
              required: true,
              message: "Please input your organization!",
            },
          ]}
        ></Form.Item>
        {signupItems.map((signupItem, idx) => {
          const renderedItem = this.renderFormItem(application, signupItem);
          if (!renderedItem) {
            return null;
          }
          return (
            <div key={idx}>
              <div
                dangerouslySetInnerHTML={{
                  __html: "<style>" + signupItem.customCss + "</style>",
                }}
              />
              {renderedItem}
            </div>
          );
        })}
        <Form.Item {...tailFormItemLayout}>
          {showSignupButton ? (
            <Button className="signup-button" disabled={this.state.loading} type="primary" htmlType="submit" style={{width: "100%"}}>
              {i18next.t("account:Sign Up")}
            </Button>
          ) : null}
          <div className="signup-link" style={{padding: showSignupButton ? "30px 0px 0px 0px" : "0px"}}>
            &nbsp;&nbsp;{i18next.t("signup:Have account?")}&nbsp;
            <a
              onClick={() => {
                const linkInStorage = sessionStorage.getItem("signinUrl");
                if (linkInStorage !== null && linkInStorage !== "") {
                  Setting.goToLinkSoft(this, linkInStorage);
                } else {
                  Setting.redirectToLoginPage(application, this.props.history);
                }
              }}
            >
              {i18next.t("signup:sign in now")}
            </a>
          </div>
          {showProviders && (
            <React.Fragment>
              <div className="social-auth-label">
                <p className="social-auth">{i18next.t("account:or") || "or"}</p>
              </div>
              <div
                style={{display: "flex", flexDirection: "column", gap: "4px"}}
              >
                {avaliableProviders.map((providerItem) => {
                  return ProviderButton.renderProviderLogo(
                    providerItem.provider,
                    application,
                    30,
                    5,
                    "medium",
                    this.props.location
                  );
                })}
              </div>
            </React.Fragment>
          )}
          <br />
        </Form.Item>
      </Form>
    );
  }

  render() {
    const application = this.getApplicationObj();
    if (application === undefined || application === null) {
      return null;
    }

    let existSignupTitle = false;
    let existSignupButton = false;
    application.signupItems?.map((item) => {
      item.name === "Signup title" ? (existSignupTitle = true) : null;
      item.name === "Signup button" ? (existSignupButton = true) : null;
    });
    if (!existSignupTitle) {
      application.signupItems?.unshift({
        customCss: "",
        label: "",
        name: "Signup title",
        placeholder: "",
        visible: true,
      });
    }
    if (!existSignupButton) {
      application.signupItems?.push({
        customCss: "",
        label: "",
        name: "Signup button",
        placeholder: "",
        visible: true,
      });
    }

    if (application.signupHtml !== "") {
      return (
        <Setting.RenderCustomHtml html={application.signupHtml} />
      );
    }

    return (
      <React.Fragment>
        <CustomGithubCorner />
        <div className="login-content" style={{margin: this.props.preview ?? this.parseOffset(application.formOffset)}}>
          {Setting.inIframe() || Setting.isMobile() ? null : <style dangerouslySetInnerHTML={{__html: Setting.getStyleInnerCss(application.formCss)}} />}
          {Setting.inIframe() || !Setting.isMobile() ? null : <style dangerouslySetInnerHTML={{__html: Setting.getStyleInnerCss(application.formCssMobile)}} />}
          <div className={Setting.isDarkTheme(this.props.themeAlgorithm) ? "login-panel-dark" : "login-panel"}>
            <div className="side-image" style={{display: application.formOffset !== 4 ? "none" : null}}>
              <Setting.RenderCustomHtml html={application.formSideHtml} />
            </div>
            <div className="login-form">
              {
                Setting.renderHelmet(application)
              }
              {
                Setting.renderLogo(application)
              }
              {
                this.renderLanguageSelect(application)
              }
              {
                this.renderForm(application)
              }
              {
                this.renderCaptchaModal(application)
              }
            </div>
          </div>
        </div>
      </React.Fragment>
    );
  }
}

export default withRouter(SignupPage);
