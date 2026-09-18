import * as Setting from "@/lib/setting";

/**
 * Whether the magic link API (/api/send-magic-link) may sign a new address up, mirrors
 * Application.IsMagicLinkApiSignupEnabled() in the backend: the application's own
 * "Magic link sign-up" switch, or the "Sign in or sign up" rule of the sign-in method.
 */
export function isMagicLinkApiSignupEnabled(application: any): boolean {
  if (!Setting.isMagicLinkEnabled(application)) {
    return false;
  }
  if (application?.enableMagicLinkSignup === true) {
    return true;
  }
  return application?.enableSignUp === true && (application?.signinMethods ?? []).some(
    (method: any) => method?.name === "Magic link" && method?.rule === "Sign in or sign up",
  );
}
