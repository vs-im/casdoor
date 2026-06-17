import React from "react";
import {Button, Popconfirm, Space, Table, Tag} from "antd";
import {Link} from "react-router-dom";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import * as Setting from "./Setting";
import * as MagicLinkBackend from "./backend/MagicLinkBackend";

class MagicLinkListPage extends BaseListPage {
  renderArray(values) {
    if (!values || values.length === 0) {
      return <span style={{opacity: 0.55}}>-</span>;
    }
    return values.join(", ");
  }

  renderTable(links) {
    const columns = [
      {
        title: i18next.t("general:Organization"),
        dataIndex: "owner",
        key: "owner",
        width: "140px",
        sorter: true,
        ...this.getColumnSearchProps("owner"),
        render: (text) => <Link to={`/organizations/${text}`}>{text}</Link>,
      },
      {
        title: i18next.t("general:Application"),
        dataIndex: "application",
        key: "application",
        width: "180px",
        sorter: true,
        ...this.getColumnSearchProps("application"),
      },
      {
        title: i18next.t("entry:Permission"),
        dataIndex: "permission",
        key: "permission",
        width: "180px",
        sorter: true,
        ...this.getColumnSearchProps("permission"),
        render: (text) => text || <Tag>{i18next.t("magicLink:Standard login access")}</Tag>,
      },
      {
        title: i18next.t("general:Email"),
        dataIndex: "email",
        key: "email",
        width: "180px",
        sorter: true,
        ...this.getColumnSearchProps("email"),
      },
      {
        title: i18next.t("general:Status"),
        dataIndex: "status",
        key: "status",
        width: "110px",
        sorter: true,
        ...this.getColumnSearchProps("status"),
        render: (text) => <Tag>{text}</Tag>,
      },
      {
        title: i18next.t("general:Requester"),
        dataIndex: "requester",
        key: "requester",
        width: "140px",
        sorter: true,
        ...this.getColumnSearchProps("requester"),
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "160px",
        sorter: true,
        render: (text) => Setting.getFormattedDate(text),
      },
      {
        title: i18next.t("magicLink:Expiry time"),
        dataIndex: "expiryTime",
        key: "expiryTime",
        width: "160px",
        sorter: true,
        render: (text) => Setting.getFormattedDate(text),
      },
      {
        title: i18next.t("magicLink:Used time"),
        dataIndex: "usedTime",
        key: "usedTime",
        width: "160px",
        sorter: true,
        render: (text) => Setting.getFormattedDate(text),
      },
      {
        title: i18next.t("magicLink:Last error"),
        dataIndex: "lastError",
        key: "lastError",
        width: "220px",
        ...this.getColumnSearchProps("lastError"),
        render: (text) => text || <span style={{opacity: 0.55}}>-</span>,
      },
      {
        title: i18next.t("magicLink:Subusers"),
        dataIndex: "subUsers",
        key: "subUsers",
        width: "220px",
        render: (text) => this.renderArray(text),
      },
      {
        title: i18next.t("magicLink:Subgroups"),
        dataIndex: "subGroups",
        key: "subGroups",
        width: "220px",
        render: (text) => this.renderArray(text),
      },
      {
        title: i18next.t("magicLink:Subroles"),
        dataIndex: "subRoles",
        key: "subRoles",
        width: "220px",
        render: (text) => this.renderArray(text),
      },
      {
        title: i18next.t("magicLink:Subdomains"),
        dataIndex: "subDomains",
        key: "subDomains",
        width: "220px",
        render: (text) => this.renderArray(text),
      },
      {
        title: i18next.t("general:Resources"),
        dataIndex: "resources",
        key: "resources",
        width: "220px",
        render: (text) => this.renderArray(text),
      },
      {
        title: i18next.t("permission:Actions"),
        dataIndex: "actions",
        key: "actions",
        width: "220px",
        render: (text) => this.renderArray(text),
      },
      {
        title: i18next.t("general:Action"),
        key: "action",
        width: "180px",
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
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <Table
        scroll={{x: "max-content"}}
        columns={columns}
        dataSource={links}
        rowKey={(record) => `${record.owner}/${record.name}`}
        size="middle"
        bordered
        pagination={paginationProps}
        title={() => <div>{i18next.t("general:Magic Links")}</div>}
        loading={this.getTableLoading()}
        onChange={this.handleTableChange}
      />
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
    this.setState({loading: true});
    MagicLinkBackend.getMagicLinks(Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
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
