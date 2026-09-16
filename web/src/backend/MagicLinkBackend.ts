import * as Setting from "@/lib/setting";

const MAGIC_LINK_FILTER_KEYS = ["status", "user", "email", "application", "organization", "group", "permission"];

export function getMagicLinks(owner: string, page: any = "", pageSize: any = "", field = "", value = "", sortField = "", sortOrder = "", filters: Record<string, string> = {}) {
  const searchParams = new URLSearchParams({
    owner,
    p: `${page}`,
    pageSize: `${pageSize}`,
    field,
    value,
    sortField,
    sortOrder,
  });
  MAGIC_LINK_FILTER_KEYS.forEach((key) => {
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

export function revokeMagicLink(owner: string, name: string) {
  return fetch(`${Setting.ServerUrl}/api/revoke-magic-link?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteMagicLink(owner: string, name: string) {
  return fetch(`${Setting.ServerUrl}/api/delete-magic-link?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
