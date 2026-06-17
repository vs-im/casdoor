export function getMagicLinkVerifyResult(res) {
  const data = res?.data && typeof res.data === "object" && !Array.isArray(res.data) ? res.data : {};
  const data2 = res?.data2 && typeof res.data2 === "object" && !Array.isArray(res.data2) ? res.data2 : {};
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

export function getMagicLinkConsentApplication(res) {
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

export function isMagicLinkSignupAction(result) {
  if (result?.isNewUser) {
    return true;
  }
  return typeof result?.authAction === "string" && /signup|new[_-]?user/i.test(result.authAction);
}
