import React from "react";
import MagicLinkListPage from "./MagicLinkListPage";

jest.mock("antd", () => {
  const React = require("react");
  return {
    Button: ({children, ...props}) => React.createElement("button", props, children),
    Popconfirm: ({children}) => React.createElement(React.Fragment, null, children),
    Space: ({children}) => React.createElement(React.Fragment, null, children),
    Table: (props) => React.createElement("table", props),
    Tabs: (props) => React.createElement("tabs", props),
    Tag: ({children}) => React.createElement("span", null, children),
  };
});

jest.mock("react-router-dom", () => {
  const React = require("react");
  return {
    Link: ({children}) => React.createElement("a", null, children),
  };
});

jest.mock("./Setting", () => ({
  getFormattedDate: (value) => value,
  isDefaultOrganizationSelected: () => true,
  getRequestOrganization: () => "",
  showMessage: () => {},
  isResponseDenied: () => false,
  goToLink: jest.fn(),
}));

jest.mock("./backend/MagicLinkBackend", () => ({
  getMagicLinks: () => Promise.resolve({status: "ok", data: [], data2: 0}),
  revokeMagicLink: () => Promise.resolve({status: "ok"}),
  deleteMagicLink: () => Promise.resolve({status: "ok"}),
}));

function getRenderedTable(renderedElement) {
  return React.Children.toArray(renderedElement.props.children).find((child) => child.props?.columns);
}

test("magic link list includes expiry column", () => {
  const page = new MagicLinkListPage({});
  page.state = {
    pagination: {total: 0},
    loading: false,
    activeView: "existingUserView",
    nowTimestamp: Date.now(),
  };
  page.getColumnSearchProps = () => ({});
  page.getTableLoading = () => false;
  page.handleTableChange = () => {};
  const tableElement = getRenderedTable(page.renderTable([]));
  const hasExpiryColumn = tableElement.props.columns.some((column) => column.key === "expireTime");
  const hasRemainingColumn = tableElement.props.columns.some((column) => column.key === "remainingTime");
  expect(hasExpiryColumn).toBe(true);
  expect(hasRemainingColumn).toBe(true);
});

test("renderArray returns placeholder for empty values", () => {
  const page = new MagicLinkListPage({});
  const emptyElement = page.renderArray([]);
  expect(emptyElement.props.children).toBe("-");
  const valuesElement = page.renderArray(["group-a", "group-b"]);
  expect(valuesElement).toBe("group-a, group-b");
});

test("magic link list switches existing and new views", () => {
  const page = new MagicLinkListPage({});
  page.state = {
    pagination: {total: 0},
    loading: false,
    activeView: "existingUserView",
    nowTimestamp: Date.now(),
  };
  page.getColumnSearchProps = () => ({});
  page.getTableLoading = () => false;
  page.handleTableChange = () => {};
  const links = [
    {owner: "built-in", name: "existing", user: "alice", authAction: "signin_existing_user"},
    {owner: "built-in", name: "new", user: "", authAction: "signup_new_user"},
  ];
  const existingView = getRenderedTable(page.renderTable(links));
  expect(existingView.props.dataSource).toHaveLength(1);
  expect(existingView.props.dataSource[0].name).toBe("existing");
  page.state.activeView = "newUserView";
  const newView = getRenderedTable(page.renderTable(links));
  expect(newView.props.dataSource).toHaveLength(1);
  expect(newView.props.dataSource[0].name).toBe("new");
});

test("new user view shows users CTA", () => {
  const page = new MagicLinkListPage({});
  page.state = {
    pagination: {total: 0},
    loading: false,
    activeView: "newUserView",
    nowTimestamp: Date.now(),
  };
  page.getColumnSearchProps = () => ({});
  page.getTableLoading = () => false;
  page.handleTableChange = () => {};
  const tableElement = getRenderedTable(page.renderTable([]));
  const title = tableElement.props.title();
  expect(typeof title.type).toBe("function");
  title.props.onClick();
  const setting = require("./Setting");
  expect(setting.goToLink).toHaveBeenCalledWith("/users");
});
