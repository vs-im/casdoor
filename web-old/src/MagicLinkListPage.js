import React from "react";
import {Button, Popconfirm, Space, Table, Tabs, Tag} from "antd";
import {Link} from "react-router-dom";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import * as Setting from "./Setting";
import * as MagicLinkBackend from "./backend/MagicLinkBackend";

class MagicLinkListPage extends BaseListPage {
  constructor(props) {
    super(props);
    this.state = {
      ...this.state,
      activeView: "existingUserView",
      nowTimestamp: Date.now(),
    };
  }

  componentDidMount() {
    super.componentDidMount();
    this.remainingTimer = setInterval(() => {
      this.setState({nowTimestamp: Date.now()});
    }, 30000);
  }

  componentWillUnmount() {
    if (this.remainingTimer) {
      clearInterval(this.remainingTimer);
    }
    super.componentWillUnmount();
  }

  renderArray(values) {
    if (!values || values.length === 0) {
      return <span style={{opacity: 0.55}}>-</span>;
    }
    return values.join(", ");
  }

  normalizeLink(record) {
    return {
      ...record,
      organization: record.organization || record.owner || "",
      user: record.user || record.requester || "",
      group: record.group || (record.subGroups && record.subGroups.length > 0 ? record.subGroups[0] : ""),
      expireTime: record.expireTime || record.expiryTime || "",
      expiryTime: record.expiryTime || record.expireTime || "",
      permission: record.permission || "",
      email: record.email || "",
      application: record.application || "",
      status: record.status || "",
    };
  }

  getExpireTimestamp(record) {
    const expireAt = Number(record.expireAt);
    if (!Number.isNaN(expireAt) && expireAt > 0) {
      return expireAt > 1000000000000 ? expireAt : expireAt * 1000;
    }
    const parsed = Date.parse(record.expireTime || record.expiryTime || "");
    return Number.isNaN(parsed) ? 0 : parsed;
  }

  getRemainingTimeText(record) {
    const expireTimestamp = this.getExpireTimestamp(record);
    if (!expireTimestamp) {
      return <span style={{opacity: 0.55}}>-</span>;
    }
    const remainingMs = expireTimestamp - this.state.nowTimestamp;
    if (remainingMs <= 0 || record.status === "expired") {
      return <Tag color="red">{i18next.t("magicLink:Expired")}</Tag>;
    }

    const remainingSeconds = Math.floor(remainingMs / 1000);
    const days = Math.floor(remainingSeconds / 86400);
    const hours = Math.floor((remainingSeconds % 86400) / 3600);
    const minutes = Math.floor((remainingSeconds % 3600) / 60);
    const seconds = remainingSeconds % 60;
    const parts = [];
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

  isNewUserLink(record) {
    if (record.isNewUser === true) {
      return true;
    }
    if (typeof record.authAction === "string") {
      const action = record.authAction.toLowerCase();
      if (action.includes("signup") || action.includes("new")) {
        return true;
      }
    }
    return false;
  }

  mapSearchField(field) {
    const legacyFieldMap = {
      organization: "owner",
      user: "requester",
      group: "subGroups",
    };
    return legacyFieldMap[field] || field;
  }

  renderTable(links) {
    const normalizedLinks = (links || []).map((link) => this.normalizeLink(link));
    const isNewUserView = this.state.activeView === "newUserView";
    const filteredLinks = normalizedLinks.filter((link) => isNewUserView ? this.isNewUserLink(link) : !this.isNewUserLink(link));

    const columns = [
      {
        title: i18next.t("general:Organization"),
        dataIndex: "organization",
        key: "organization",
        width: "100px",
        sorter: true,
        ...this.getColumnSearchProps("organization"),
        render: (text, record) => <Link to={`/organizations/${text || record.owner}`}>{text || record.owner}</Link>,
      },
      {
        title: i18next.t("general:Application"),
        dataIndex: "application",
        key: "application",
        width: "100px",
        sorter: true,
        ...this.getColumnSearchProps("application"),
      },
      {
        title: i18next.t("general:Email"),
        dataIndex: "email",
        key: "email",
        width: "100px",
        sorter: true,
        ...this.getColumnSearchProps("email"),
      },
      {
        title: i18next.t("general:User"),
        dataIndex: "user",
        key: "user",
        width: "100px",
        sorter: true,
        ...this.getColumnSearchProps("user"),
        render: (text) => text || <span style={{opacity: 0.55}}>-</span>,
      },
      {
        title: i18next.t("general:Status"),
        dataIndex: "status",
        key: "status",
        width: "100px",
        sorter: true,
        ...this.getColumnSearchProps("status"),
        render: (text) => <Tag>{text}</Tag>,
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "100px",
        sorter: true,
        render: (text) => Setting.getFormattedDate(text),
      },
      {
        title: i18next.t("magicLink:Expiry time"),
        dataIndex: "expireTime",
        key: "expireTime",
        width: "100px",
        sorter: true,
        render: (text, record) => Setting.getFormattedDate(text || record.expiryTime),
      },
      {
        title: i18next.t("magicLink:Remaining time"),
        key: "remainingTime",
        width: "100px",
        render: (text, record) => this.getRemainingTimeText(record),
      },
      {
        title: i18next.t("magicLink:Used time"),
        dataIndex: "usedTime",
        key: "usedTime",
        width: "100px",
        sorter: true,
        render: (text) => Setting.getFormattedDate(text),
      },
      {
        title: i18next.t("magicLink:Last error"),
        dataIndex: "lastError",
        key: "lastError",
        width: "100px",
        ...this.getColumnSearchProps("lastError"),
        render: (text) => text || <span style={{opacity: 0.55}}>-</span>,
      },
      {
        title: i18next.t("general:Action"),
        key: "action",
        width: "100px",
        render: (text, record) => {
          const disabled = !["created", "sent", "opened"].includes(record.status);
          return (
            <Space>
              <Popconfirm
                title={`${i18next.t("general:Confirm")} ${i18next.t("magicLink:Revoke")}?`}
                onConfirm={() => this.revoke(record)}
                disabled={disabled}
              >
                <Button disabled={disabled} size="small">{i18next.t("magicLink:Revoke")}</Button>
              </Popconfirm>
              <Popconfirm
                title={`${i18next.t("general:Confirm")} ${i18next.t("general:Delete")}?`}
                onConfirm={() => this.delete(record)}
              >
                <Button danger size="small">{i18next.t("general:Delete")}</Button>
              </Popconfirm>
            </Space>
          );
        },
      },
    ];

    const paginationProps = {
      total: this.state.pagination.total,
      showQuickJumper: true,
      showSizeChanger: true,
      pageSize: this.state.pagination.pageSize,
      current: this.state.pagination.current,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <div>
        <Tabs
          activeKey={this.state.activeView}
          items={[
            {label: i18next.t("login:Sign In"), key: "existingUserView"},
            {label: i18next.t("magicLink:Sign up"), key: "newUserView"},
          ]}
          onChange={(key) => this.setState({activeView: key})}
        />
        <Table
          columns={columns}
          dataSource={filteredLinks}
          rowKey={(record) => `${record.owner}/${record.name}`}
          size="middle"
          bordered
          pagination={paginationProps}
          title={() => (
            isNewUserView ? (
              <Button size="small" type="primary" onClick={() => Setting.goToLink("/users")}>
                {i18next.t("general:Users")}
              </Button>
            ) : <div>{i18next.t("general:Magic Links")}</div>
          )
          }
          loading={this.getTableLoading()}
          onChange={this.handleTableChange}
        />
      </div>
    );
  }

  revoke(record) {
    MagicLinkBackend.revokeMagicLink(record.owner, record.name)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.fetch(this.state.pagination);
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  delete(record) {
    MagicLinkBackend.deleteMagicLink(record.owner, record.name)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.fetch(this.state.pagination);
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  fetch = (params = {}) => {
    const field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    const mappedField = this.mapSearchField(field);
    this.setState({loading: true});
    MagicLinkBackend.getMagicLinks(
      Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account),
      params.pagination.current,
      params.pagination.pageSize,
      mappedField,
      value,
      sortField,
      sortOrder,
      {
        status: field === "status" ? value : "",
        user: field === "user" ? value : "",
        email: field === "email" ? value : "",
        application: field === "application" ? value : "",
        organization: field === "organization" ? value : "",
        group: field === "group" ? value : "",
        permission: field === "permission" ? value : "",
      }
    )
      .then((res) => {
        this.setState({loading: false});
        if (res.status === "ok") {
          this.setState({
            data: res.data,
            pagination: {
              ...params.pagination,
              total: res.data2,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });
        } else {
          if (Setting.isResponseDenied(res)) {
            this.setState({isAuthorized: false});
          } else {
            Setting.showMessage("error", res.msg);
          }
        }
      });
  };
}

export default MagicLinkListPage;
