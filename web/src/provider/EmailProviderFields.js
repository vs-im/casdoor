// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

import React from "react";
import {Button, Col, Input, InputNumber, Row, Select, Switch, Tabs} from "antd";
import {LinkOutlined} from "@ant-design/icons";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as ProviderEditTestEmail from "../common/TestEmailWidget";
import Editor from "../common/Editor";
import HttpHeaderTable from "../table/HttpHeaderTable";

const {Option} = Select;

export function renderEmailProviderFields(provider, updateProviderField, renderEmailMappingInput, account) {
  const verificationContent = provider.content || Setting.getDefaultHtmlEmailContent();
  const invitationContent = provider.metadata || Setting.getDefaultInvitationHtmlEmailContent();
  const magicLinkContent = provider.magicLinkContent || Setting.getDefaultMagicLinkHtmlEmailContent();
  const magicLinkPreviewUrl = `${Setting.ServerUrl}/magic-link/callback?token=example`;
  const renderEmailContentEditor = (field, value, resetText, resetHtml, previewHtml) => (
    <React.Fragment>
      <Row style={{marginTop: "20px"}} >
        <Button style={{marginLeft: "10px", marginBottom: "5px"}} onClick={() => updateProviderField(field, resetText)} >
          {i18next.t("general:Reset to Default")} (Text)
        </Button>
        <Button style={{marginLeft: "10px", marginBottom: "5px"}} type="primary" onClick={() => updateProviderField(field, resetHtml)} >
          {i18next.t("general:Reset to Default")} (HTML)
        </Button>
      </Row>
      <Row>
        <Col span={Setting.isMobile() ? 22 : 11}>
          <div style={{height: "300px", margin: "10px"}}>
            <Editor
              value={value}
              fillHeight
              dark
              lang="html"
              onChange={value => {
                updateProviderField(field, value);
              }}
            />
          </div>
        </Col>
        <Col span={1} />
        <Col span={Setting.isMobile() ? 22 : 11}>
          <div style={{margin: "10px"}}>
            <div dangerouslySetInnerHTML={{__html: previewHtml}} />
          </div>
        </Col>
      </Row>
    </React.Fragment>
  );

  return (
    <React.Fragment>
      {
        ["Custom HTTP Email", "SendGrid"].includes(provider.type) ? (
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={2}>
              {Setting.getLabel(i18next.t("provider:Endpoint"), i18next.t("provider:Region endpoint for Internet"))} :
            </Col>
            <Col span={22} >
              <Input prefix={<LinkOutlined />} value={provider.endpoint} onChange={e => {
                updateProviderField("endpoint", e.target.value);
              }} />
            </Col>
          </Row>) : null
      }
      {provider.type === "Resend" ? null : (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Host"), i18next.t("provider:Host - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={provider.host} onChange={e => {
              updateProviderField("host", e.target.value);
            }} />
          </Col>
        </Row>
      )}
      {["Azure ACS", "SendGrid", "Resend"].includes(provider.type) ? null : (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Port"), i18next.t("provider:Port - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={provider.port} onChange={value => {
              updateProviderField("port", value);
            }} />
          </Col>
        </Row>
      )}
      {["Azure ACS", "SendGrid", "Resend"].includes(provider.type) ? null : (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:SSL mode"), i18next.t("provider:SSL mode - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "200px"}} value={provider.sslMode || "Auto"} onChange={value => {
              updateProviderField("sslMode", value);
            }}>
              <Option value="Auto">{i18next.t("general:Auto")}</Option>
              <Option value="Enable">{i18next.t("general:Enable")}</Option>
              <Option value="Disable">{i18next.t("general:Disable")}</Option>
            </Select>
          </Col>
        </Row>
      )}
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Enable proxy"), i18next.t("provider:Enable proxy - Tooltip"))} :
        </Col>
        <Col span={1} >
          <Switch checked={provider.enableProxy} onChange={checked => {
            updateProviderField("enableProxy", checked);
          }} />
        </Col>
      </Row>
      {
        provider.type === "Custom HTTP Email" ? (
          <React.Fragment>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("general:Method"), i18next.t("provider:Method - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Select virtual={false} style={{width: "100%"}} value={provider.method} onChange={value => {
                  updateProviderField("method", value);
                }}>
                  {
                    [
                      {id: "GET", name: "GET"},
                      {id: "POST", name: "POST"},
                      {id: "PUT", name: "PUT"},
                      {id: "DELETE", name: "DELETE"},
                    ].map((method, index) => <Option key={index} value={method.id}>{method.name}</Option>)
                  }
                </Select>
              </Col>
            </Row>
            {
              provider.method !== "GET" ? (<Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("webhook:Content type"), i18next.t("webhook:Content type - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Select virtual={false} style={{width: "100%"}} value={provider.issuerUrl === "" ? "application/x-www-form-urlencoded" : provider.issuerUrl} onChange={value => {
                    updateProviderField("issuerUrl", value);
                  }}>
                    {
                      [
                        {id: "application/json", name: "application/json"},
                        {id: "application/x-www-form-urlencoded", name: "application/x-www-form-urlencoded"},
                      ].map((method, index) => <Option key={index} value={method.id}>{method.name}</Option>)
                    }
                  </Select>
                </Col>
              </Row>) : null
            }
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("provider:HTTP header"), i18next.t("provider:HTTP header - Tooltip"))} :
              </Col>
              <Col span={22} >
                <HttpHeaderTable httpHeaders={provider.httpHeaders} onUpdateTable={(value) => {updateProviderField("httpHeaders", value);}} />
              </Col>
            </Row>
            {provider.method !== "GET" ? <Row style={{marginTop: "20px"}}>
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("provider:HTTP body mapping"), i18next.t("provider:HTTP body mapping - Tooltip"))} :
              </Col>
              <Col span={22}>
                {renderEmailMappingInput()}
              </Col>
            </Row> : null}
          </React.Fragment>
        ) : null
      }
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Email title"), i18next.t("provider:Email title - Tooltip"))} :
        </Col>
        <Col span={22} >
          <Input value={provider.title} onChange={e => {
            updateProviderField("title", e.target.value);
          }} />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Email content"), i18next.t("provider:Email content - Tooltip"))} :
        </Col>
        <Col span={22} >
          <Tabs
            items={[
              {
                key: "verification",
                label: i18next.t("provider:Verification code"),
                children: renderEmailContentEditor(
                  "content",
                  verificationContent,
                  "You have requested a verification code at Casdoor. Here is your code: %s, please enter in 5 minutes. <reset-link>Or click %link to reset</reset-link>",
                  Setting.getDefaultHtmlEmailContent(),
                  verificationContent.replace("%s", "123456").replace("%{user.friendlyName}", Setting.getFriendlyUserName(account))
                ),
              },
              {
                key: "invitation",
                label: i18next.t("general:Invitations"),
                children: renderEmailContentEditor(
                  "metadata",
                  invitationContent,
                  "You have invited to join Casdoor. Here is your invitation code: %s, please enter in 5 minutes. Or click %link to signup",
                  Setting.getDefaultInvitationHtmlEmailContent(),
                  invitationContent.replace(/%code/g, "123456").replace(/%s/g, "123456").replace(/%link/g, `${Setting.ServerUrl}/signup?code=123456`)
                ),
              },
              {
                key: "magicLink",
                label: i18next.t("provider:Magic Link"),
                children: renderEmailContentEditor(
                  "magicLinkContent",
                  magicLinkContent,
                  "Use this Magic Link to sign in: %link",
                  Setting.getDefaultMagicLinkHtmlEmailContent(),
                  magicLinkContent.replace(/%link/g, magicLinkPreviewUrl)
                ),
              },
            ]}
          />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}}>
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Test Email"), i18next.t("provider:Test Email - Tooltip"))} :
        </Col>
        <Col span={4}>
          <Input value={provider.receiver} placeholder={i18next.t("user:Input your email")}
            onChange={e => {
              updateProviderField("receiver", e.target.value);
            }} />
        </Col>
        {["Azure ACS", "SendGrid", "Resend"].includes(provider.type) ? null : (
          <Button style={{marginLeft: "10px", marginBottom: "5px"}} onClick={() => ProviderEditTestEmail.connectSmtpServer(provider)} >
            {i18next.t("provider:Test SMTP Connection")}
          </Button>
        )}
        <Button style={{marginLeft: "10px", marginBottom: "5px"}} type="primary"
          disabled={!Setting.isValidEmail(provider.receiver)}
          onClick={() => ProviderEditTestEmail.sendTestEmail(provider, provider.receiver)} >
          {i18next.t("provider:Send Testing Email")}
        </Button>
      </Row>
    </React.Fragment>
  );
}
