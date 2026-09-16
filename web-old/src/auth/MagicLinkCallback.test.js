import MagicLinkCallback from "./MagicLinkCallback";
import * as AuthBackend from "./AuthBackend";
import * as Util from "./Util";
import * as Setting from "../Setting";
import {getMagicLinkConsentApplication, getMagicLinkVerifyResult, isMagicLinkSignupAction} from "./magicLinkVerifyResult";

jest.mock("i18next", () => ({
  t: (key) => key,
}));

jest.mock("antd", () => {
  const React = require("react");
  return {
    Button: ({children, ...props}) => React.createElement("button", props, children),
    Result: ({children, ...props}) => React.createElement("div", props, children),
  };
});

jest.mock("./AuthBackend", () => ({
  verifyMagicLink: jest.fn(),
}));

jest.mock("./Util", () => ({
  getOAuthGetParameters: jest.fn(),
}));

jest.mock("../Setting", () => ({
  goToLink: jest.fn(),
  goToLinkSoft: jest.fn(),
  showMessage: jest.fn(),
  createFormAndSubmit: jest.fn(),
}));

function flushPromises() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

function createCallbackComponent({search = "?token=magic-token", oAuthParams = {}, onLoginSuccess = jest.fn()} = {}) {
  Util.getOAuthGetParameters.mockReturnValue(oAuthParams);
  const component = new MagicLinkCallback({
    location: {search},
    onLoginSuccess,
  });
  component.setState = (nextState) => {
    component.state = {...component.state, ...nextState};
  };
  return {component, onLoginSuccess};
}

beforeEach(() => {
  jest.clearAllMocks();
  sessionStorage.clear();
});

test("completes login response without onboarding redirects", async() => {
  AuthBackend.verifyMagicLink.mockResolvedValue({status: "ok", data: "built-in/alice"});
  const {component, onLoginSuccess} = createCallbackComponent({
    oAuthParams: {responseType: "login"},
  });
  component.componentDidMount();
  await flushPromises();
  expect(Setting.showMessage).toHaveBeenCalledWith("success", "application:Logged in successfully");
  expect(onLoginSuccess).toHaveBeenCalled();
  expect(Setting.goToLink).toHaveBeenCalledWith("/");
  expect(Setting.goToLink).not.toHaveBeenCalledWith("/signup");
  expect(Setting.goToLink).not.toHaveBeenCalledWith("/account");
});

test("keeps existing user code flow redirect contract", async() => {
  AuthBackend.verifyMagicLink.mockResolvedValue({
    status: "ok",
    data: "oauth-code-1",
    authAction: "signin_existing_user",
    isNewUser: false,
  });
  const dispatchSpy = jest.spyOn(window, "dispatchEvent");
  const {component} = createCallbackComponent({
    search: "?token=magic-token&state=s1",
    oAuthParams: {
      responseType: "code",
      redirectUri: "https://app.example/cb",
      state: "s1",
    },
  });
  component.componentDidMount();
  await flushPromises();
  expect(Setting.goToLink).toHaveBeenCalledWith("https://app.example/cb?code=oauth-code-1&state=s1");
  expect(sessionStorage.getItem("magicLinkVerifyResult")).toBe(JSON.stringify({authAction: "signin_existing_user", isNewUser: false}));
  const signupEventCalls = dispatchSpy.mock.calls.filter(([event]) => event.type === "casdoor:magic-link-signup-first-login");
  expect(signupEventCalls).toHaveLength(0);
  dispatchSpy.mockRestore();
});

test("supports token and id_token callbacks with form_post", async() => {
  AuthBackend.verifyMagicLink.mockResolvedValue({status: "ok", data: "jwt-token"});
  const {component} = createCallbackComponent({
    oAuthParams: {
      responseType: "token id_token",
      responseMode: "form_post",
      redirectUri: "https://app.example/cb",
      state: "s2",
    },
  });
  component.componentDidMount();
  await flushPromises();
  expect(Setting.createFormAndSubmit).toHaveBeenCalledWith("https://app.example/cb", {
    token: "jwt-token",
    id_token: "jwt-token",
    token_type: "bearer",
    state: "s2",
  });
});

test("supports token callback hash redirect in query mode", async() => {
  AuthBackend.verifyMagicLink.mockResolvedValue({status: "ok", data: "access-token-1"});
  const {component} = createCallbackComponent({
    oAuthParams: {
      responseType: "token",
      redirectUri: "https://app.example/cb",
      state: "s3",
    },
  });
  component.componentDidMount();
  await flushPromises();
  expect(Setting.goToLink).toHaveBeenCalledWith("https://app.example/cb#access_token=access-token-1&state=s3&token_type=bearer");
});

test("redirects to consent with updated callback contract fields", async() => {
  AuthBackend.verifyMagicLink.mockResolvedValue({
    status: "ok",
    data: {required: true},
    data2: {authAction: "signup_new_user", isNewUser: true},
    data3: "demo-app",
  });
  const {component} = createCallbackComponent({
    search: "?token=magic-token&client_id=client1&state=abc",
    oAuthParams: {
      responseType: "code",
      redirectUri: "https://app.example/cb",
      state: "abc",
    },
  });
  component.componentDidMount();
  await flushPromises();
  expect(Setting.goToLinkSoft).toHaveBeenCalledWith(component, "/consent/demo-app?client_id=client1&state=abc");
});

test("shows backend error response on callback failure", async() => {
  AuthBackend.verifyMagicLink.mockResolvedValue({status: "error", msg: "magic link expired"});
  const {component} = createCallbackComponent();
  component.componentDidMount();
  await flushPromises();
  expect(component.state.loading).toBe(false);
  expect(component.state.error).toBe("magic link expired");
});

test("shows network error response on callback failure", async() => {
  AuthBackend.verifyMagicLink.mockRejectedValue(new Error("network timeout"));
  const {component} = createCallbackComponent();
  component.componentDidMount();
  await flushPromises();
  expect(component.state.loading).toBe(false);
  expect(component.state.error).toContain("general:Failed to connect to server");
});

test("reads verify contract fields from data2 object", () => {
  const res = {
    status: "ok",
    data: "token",
    data2: {
      authAction: "signup",
      isNewUser: true,
    },
  };
  const result = getMagicLinkVerifyResult(res);
  expect(result.authAction).toBe("signup");
  expect(result.isNewUser).toBe(true);
});

test("reads verify contract fields from data object", () => {
  const res = {
    status: "ok",
    data: {
      authAction: "signin",
      isNewUser: false,
    },
    data2: "app",
  };
  const result = getMagicLinkVerifyResult(res);
  expect(result.authAction).toBe("signin");
  expect(result.isNewUser).toBe(false);
});

test("detects signup auth action by pattern", () => {
  expect(isMagicLinkSignupAction({authAction: "magic_link_signup"})).toBe(true);
  expect(isMagicLinkSignupAction({authAction: "signin", isNewUser: false})).toBe(false);
});

test("reads consent application from mixed contracts", () => {
  expect(getMagicLinkConsentApplication({data2: "app-data2"})).toBe("app-data2");
  expect(getMagicLinkConsentApplication({data2: {authAction: "signup"}, data3: "app-data3"})).toBe("app-data3");
  expect(getMagicLinkConsentApplication({data: {application: "app-data"}})).toBe("app-data");
});
