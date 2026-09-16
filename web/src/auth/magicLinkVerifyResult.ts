// Helpers around the /api/verify-magic-link response, ported from
// web-old/src/auth/magicLinkVerifyResult.js. The backend reports the auth action
// ("signin_existing_user" / "signup_new_user") and the new-user flag on the top
// level of the response, but older builds nested them, so every spelling is read.

export interface MagicLinkVerifyResult {
  authAction: string;
  isNewUser: boolean;
}

function asObject(value: any): Record<string, any> {
  return value && typeof value === "object" && !Array.isArray(value) ? value : {};
}

export function getMagicLinkVerifyResult(res: any): MagicLinkVerifyResult {
  const data = asObject(res?.data);
  const data2 = asObject(res?.data2);
  const authAction = [
    res?.authAction,
    data.authAction,
    data2.authAction,
    data.result,
    data2.result,
    typeof res?.data3 === "string" ? res.data3 : "",
  ].find((value) => typeof value === "string" && value !== "") || "";
  const isNewUser = [
    res?.isNewUser,
    data.isNewUser,
    data2.isNewUser,
  ].find((value) => typeof value === "boolean");
  return {
    authAction,
    isNewUser: isNewUser === true,
  };
}

/** The application whose consent page has to be shown before the code is handed out. */
export function getMagicLinkConsentApplication(res: any): string {
  const candidates = [
    res?.data2,
    res?.data3,
    res?.data?.application,
    res?.data2?.application,
    res?.data?.app,
    res?.data2?.app,
  ];
  return candidates.find((value) => typeof value === "string" && value !== "") || "";
}

export function isMagicLinkSignupAction(result: MagicLinkVerifyResult | null | undefined): boolean {
  if (result?.isNewUser) {
    return true;
  }
  return typeof result?.authAction === "string" && /signup|new[_-]?user/i.test(result.authAction);
}
