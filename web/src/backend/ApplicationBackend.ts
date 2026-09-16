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

import * as Setting from "@/lib/setting";
import * as OrganizationBackend from "@/backend/OrganizationBackend";

/**
 * The bare auth pages (`/login`, `/signup`, `/forget`, `/prompt`, no explicit
 * application name in the URL) used to fetch `Conf.DefaultApplication`
 * ("app-built-in") directly, but that application doesn't exist on our
 * instances — only the `built-in` organization's actual default application
 * does. Resolve it the same way the owner-scoped routes do
 * (`get-default-application` for org `built-in`), falling back to the old
 * `app-built-in` lookup so an instance that does define that application
 * keeps working.
 */
export function getDefaultLoginApplication(applicationName: string, isExplicit: boolean) {
  if (isExplicit) {
    return getApplication("admin", applicationName);
  }
  return OrganizationBackend.getDefaultApplication("admin", "built-in").then((res: any) => {
    if (res.status === "ok" && res.data) {
      return res;
    }
    return getApplication("admin", applicationName);
  });
}

export function getApplications(owner, page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return fetch(`${Setting.ServerUrl}/api/get-applications?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getApplicationsByOrganization(owner, organization, page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return fetch(`${Setting.ServerUrl}/api/get-organization-applications?owner=${owner}&organization=${organization}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getApplication(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-application?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getUserApplication(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-user-application?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function updateApplication(owner, name, application) {
  return fetch(`${Setting.ServerUrl}/api/update-application?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(application),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function addApplication(application) {
  const newApplication = Setting.deepCopy(application);
  return fetch(`${Setting.ServerUrl}/api/add-application`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newApplication),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteApplication(application) {
  const newApplication = Setting.deepCopy(application);
  return fetch(`${Setting.ServerUrl}/api/delete-application`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newApplication),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getSamlMetadata(owner, name, enablePostBinding) {
  return fetch(`${Setting.ServerUrl}/api/saml/metadata?application=${owner}/${encodeURIComponent(name)}&enablePostBinding=${enablePostBinding}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.text());
}
