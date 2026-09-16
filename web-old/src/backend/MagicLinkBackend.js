import * as Setting from "../Setting";

export function getMagicLinks(owner, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "", filters = {}) {
  const searchParams = new URLSearchParams({
    owner,
    p: page,
    pageSize,
    field,
    value,
    sortField,
    sortOrder,
  });
  ["status", "user", "email", "application", "organization", "group", "permission"].forEach((key) => {
    const filterValue = filters[key];
    if (filterValue !== undefined && filterValue !== null && filterValue !== "") {
      searchParams.set(key, filterValue);
    }
  });
  return fetch(`${Setting.ServerUrl}/api/get-magic-links?${searchParams.toString()}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function revokeMagicLink(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/revoke-magic-link?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteMagicLink(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/delete-magic-link?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
