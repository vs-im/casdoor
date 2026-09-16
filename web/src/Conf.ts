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

import * as Cookie from "cookie";

/**
 * Branding. The compile-time values are the stock upstream ones, so the tree
 * carries no downstream brand; a deployment sets the CASDOOR_BRAND_* variables on
 * the server (see docs/branding.md) and the backend ships them here inside the
 * `jsonWebConfig` cookie. `VITE_BRAND_*` only exists for `vite dev` without a
 * backend. These are mutable module bindings: read them as `Conf.BrandName` at
 * render time, never snapshot them into a module-level const.
 */
const viteEnv = (import.meta as any).env ?? {};

/** The product name shown wherever upstream showed its own: titles, alt texts, emails, the TOTP issuer fallback. */
export let BrandName: string = viteEnv.VITE_BRAND_NAME || "Casdoor";
/** The wordmark, and its twin for dark backgrounds. */
export let BrandLogoUrl: string = viteEnv.VITE_BRAND_LOGO_URL || "https://cdn.casbin.org/img/casdoor-logo_1185x256.png";
export let BrandLogoDarkUrl: string = viteEnv.VITE_BRAND_LOGO_DARK_URL || BrandLogoUrl;
/** The square mark embedded in the built-in email templates. */
export let BrandLogoMarkUrl: string = viteEnv.VITE_BRAND_LOGO_MARK_URL || BrandLogoUrl;
export let BrandFaviconUrl: string = viteEnv.VITE_BRAND_FAVICON_URL || "https://cdn.casbin.org/img/favicon.png";
export let BrandWebsiteUrl: string = viteEnv.VITE_BRAND_WEBSITE_URL || "https://casdoor.org";

export let DefaultApplication = "app-built-in";

export let ShowGithubCorner = false;
export let IsDemoMode = false;

export let ForceLanguage = "";
export let DefaultLanguage = "en";

export let StaticBaseUrl = "https://cdn.casbin.org";

export const ThemeDefault = {
  themeType: "default",
  colorPrimary: "#262626",
  borderRadius: 10,
  isCompact: false,
};

export const CustomFooter = null;

// Maximum number of navbar items before switching from flat to grouped menu
export let MaxItemsForFlatMenu = 7;

// setConfig updates the frontend configuration from backend
export function setConfig(config: Record<string, any>) {
  if (!config) {
    return;
  }
  if (config.showGithubCorner !== undefined) {
    ShowGithubCorner = config.showGithubCorner;
  }
  if (config.isDemoMode !== undefined) {
    IsDemoMode = config.isDemoMode;
  }
  if (config.forceLanguage !== undefined) {
    ForceLanguage = config.forceLanguage;
  }
  if (config.defaultLanguage !== undefined) {
    DefaultLanguage = config.defaultLanguage;
  }
  if (config.staticBaseUrl !== undefined) {
    StaticBaseUrl = config.staticBaseUrl;
  }
  if (config.defaultApplication !== undefined) {
    DefaultApplication = config.defaultApplication;
  }
  if (config.maxItemsForFlatMenu !== undefined) {
    MaxItemsForFlatMenu = config.maxItemsForFlatMenu;
  }
  if (config.brandName) {
    BrandName = config.brandName;
  }
  if (config.brandLogoUrl) {
    BrandLogoUrl = config.brandLogoUrl;
    BrandLogoDarkUrl = config.brandLogoUrl;
  }
  if (config.brandLogoDarkUrl) {
    BrandLogoDarkUrl = config.brandLogoDarkUrl;
  }
  if (config.brandLogoMarkUrl) {
    BrandLogoMarkUrl = config.brandLogoMarkUrl;
  }
  if (config.brandFaviconUrl) {
    BrandFaviconUrl = config.brandFaviconUrl;
  }
  if (config.brandWebsiteUrl) {
    BrandWebsiteUrl = config.brandWebsiteUrl;
  }
}

export function initConfigFromCookie() {
  if (typeof document === "undefined") {
    return;
  }

  try {
    const curCookie = Cookie.parse(document.cookie);
    const raw = curCookie["jsonWebConfig"];
    if (!raw || raw === "null") {
      return;
    }

    const config = JSON.parse(raw);
    setConfig(config);
  } catch {
    // Ignore malformed cookie and keep compile-time defaults.
  }
}
