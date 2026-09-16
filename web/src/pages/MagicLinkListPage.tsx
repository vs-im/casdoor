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

import * as React from "react";
import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Tabs, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {clientIpColumn, dateColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useOrganizationFilter} from "@/hooks/use-organization";
import * as MagicLinkBackend from "@/backend/MagicLinkBackend";
import * as Setting from "@/lib/setting";

/** a link can still be revoked while it has not been used, failed or expired */
const REVOCABLE_STATUSES = ["created", "sent", "opened"];

const STATUS_VARIANTS: Record<string, "info" | "success" | "warning" | "destructive" | "secondary"> = {
  created: "info",
  sent: "info",
  opened: "warning",
  used: "success",
  failed: "destructive",
  expired: "secondary",
  revoked: "secondary",
};

/** the list API filters on the stored column names, the table shows the friendlier ones */
const SEARCH_FIELD_MAP: Record<string, string> = {
  organization: "owner",
  user: "requester",
  group: "subGroups",
};

function getExpireTimestamp(record: any): number {
  const expireAt = Number(record.expireAt);
  if (!Number.isNaN(expireAt) && expireAt > 0) {
    return expireAt > 1000000000000 ? expireAt : expireAt * 1000;
  }
  const parsed = Date.parse(record.expiryTime || record.expireTime || "");
  return Number.isNaN(parsed) ? 0 : parsed;
}

/** a link requested for an email without an account signs the new user up */
export function isNewUserLink(record: any): boolean {
  if (record?.isNewUser === true) {
    return true;
  }
  if (typeof record?.authAction === "string") {
    const action = record.authAction.toLowerCase();
    return action.includes("signup") || action.includes("new");
  }
  return false;
}

export function getRemainingTimeText(record: any, now: number): string {
  const expireTimestamp = getExpireTimestamp(record);
  if (!expireTimestamp) {
    return "";
  }
  const remainingMs = expireTimestamp - now;
  if (remainingMs <= 0 || record.status === "expired") {
    return i18next.t("magicLink:Expired");
  }
  const remainingSeconds = Math.floor(remainingMs / 1000);
  const days = Math.floor(remainingSeconds / 86400);
  const hours = Math.floor((remainingSeconds % 86400) / 3600);
  const minutes = Math.floor((remainingSeconds % 3600) / 60);
  const seconds = remainingSeconds % 60;
  const parts: string[] = [];
  if (days > 0) {
    parts.push(`${days}d`);
  }
  if (hours > 0 || days > 0) {
    parts.push(`${hours}h`);
  }
  if (minutes > 0 || hours > 0 || days > 0) {
    parts.push(`${minutes}m`);
  } else {
    parts.push(`${seconds}s`);
  }
  return parts.join(" ");
}

const Empty = () => <span className="text-muted-foreground">-</span>;

/**
 * The magic links an organization has issued, split into the ones that sign an
 * existing user in and the ones that sign a new user up. Ported from
 * web-old/src/MagicLinkListPage.js.
 */
export default function MagicLinkListPage() {
  // GetMagicLinks() filters by "owner", an empty one means every organization
  const owner = useOrganizationFilter();
  const [view, setView] = React.useState<"existingUserView" | "newUserView">("existingUserView");
  const [now, setNow] = React.useState(() => Date.now());

  // the remaining-time column counts down without a refetch
  React.useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 30000);
    return () => window.clearInterval(timer);
  }, []);

  const columns: ColumnDef<any>[] = [
    organizationColumn(140, "owner", undefined, "left"),
    textColumn({dataIndex: "name", title: i18next.t("general:Name"), width: 180, searchable: true, mono: true}),
    textColumn({
      dataIndex: "application",
      title: i18next.t("general:Application"),
      width: 160,
      searchable: true,
      link: (value, record: any) => (value ? `/applications/${record.owner}/${value}` : undefined),
    }),
    textColumn({dataIndex: "email", title: i18next.t("general:Email"), width: 200, searchable: true}),
    {
      dataIndex: "requester",
      title: i18next.t("general:User"),
      width: 140,
      sortable: true,
      searchable: true,
      render: (value, record: any) =>
        value ? (
          <Link to={`/users/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : (
          <Empty />
        ),
    },
    {
      dataIndex: "status",
      title: i18next.t("general:Status"),
      width: 110,
      sortable: true,
      searchable: true,
      render: (value) => (value ? <Badge variant={STATUS_VARIANTS[value] ?? "secondary"}>{value}</Badge> : <Empty />),
    },
    dateColumn(),
    dateColumn("expiryTime", i18next.t("magicLink:Expiry time")),
    {
      dataIndex: "expireAt",
      title: i18next.t("magicLink:Remaining time"),
      width: 130,
      sortable: false,
      render: (_value, record: any) => {
        const text = getRemainingTimeText(record, now);
        if (text === "") {
          return <Empty />;
        }
        return text === i18next.t("magicLink:Expired") ? <Badge variant="destructive">{text}</Badge> : text;
      },
    },
    dateColumn("usedTime", i18next.t("magicLink:Used time")),
    textColumn({dataIndex: "permission", title: i18next.t("general:Permission"), width: 160, searchable: true}),
    clientIpColumn({dataIndex: "remoteAddr"}),
    {
      dataIndex: "lastError",
      title: i18next.t("magicLink:Last error"),
      width: 220,
      sortable: false,
      searchable: true,
      render: (value) => (value ? <span className="text-destructive">{value}</span> : <Empty />),
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Magic Links")}
      toolbar={
        <Tabs value={view} onValueChange={(v) => setView(v as typeof view)}>
          <TabsList>
            <TabsTrigger value="existingUserView">{i18next.t("login:Sign In")}</TabsTrigger>
            <TabsTrigger value="newUserView">{i18next.t("magicLink:Sign up")}</TabsTrigger>
          </TabsList>
        </Tabs>
      }
      columns={columns}
      deps={[owner, view]}
      tableId={`/magic-links/${view}`}
      rowKey={(record) => `${record.owner}/${record.name}`}
      remove={(record) => MagicLinkBackend.deleteMagicLink(record.owner, record.name)}
      rowActions={(record, _index, {refresh}) => [
        {
          key: "revoke",
          label: i18next.t("magicLink:Revoke"),
          disabled: !REVOCABLE_STATUSES.includes(record.status),
          confirm: {title: `${i18next.t("general:Confirm")} ${i18next.t("magicLink:Revoke")}?`},
          onSelect: () =>
            MagicLinkBackend.revokeMagicLink(record.owner, record.name).then((res: any) => {
              if (res.status === "ok") {
                Setting.showMessage("success", i18next.t("general:Successfully deleted"));
                refresh();
              } else {
                Setting.showMessage("error", res.msg);
              }
            }),
        },
      ]}
      fetch={(q) => {
        const field = q.searchedColumn;
        const value = q.searchText;
        return MagicLinkBackend.getMagicLinks(
          owner,
          q.page,
          q.pageSize,
          SEARCH_FIELD_MAP[field] ?? field,
          value,
          q.sortField,
          q.sortOrder,
          {
            status: field === "status" ? value : "",
            user: field === "requester" ? value : "",
            email: field === "email" ? value : "",
            application: field === "application" ? value : "",
            organization: field === "owner" ? value : "",
            permission: field === "permission" ? value : "",
          },
        ).then((res: any) => {
          if (res.status !== "ok") {
            return res;
          }
          // the two tabs split one list, so the page total is the list's
          const links = (res.data ?? []).filter((link: any) =>
            view === "newUserView" ? isNewUserLink(link) : !isNewUserLink(link),
          );
          return {...res, data: links};
        });
      }}
    />
  );
}
