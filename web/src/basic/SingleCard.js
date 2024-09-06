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

import React from "react";
import {Card} from "antd";
import * as Setting from "../Setting";
import {withRouter} from "react-router-dom";

const {Meta} = Card;

class SingleCard extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  wrappedAsSilentSigninLink(link) {
    if (link.startsWith("http")) {
      link += link.includes("?") ? "&silentSignin=1" : "?silentSignin=1";
    }
    return link;
  }

  renderCardMobile(logo, link, title, desc, time, isSingle) {
    // TODO: add borderRadius from app settings
    const gridStyle = {
      width: "100%",
      textAlign: "center",
      cursor: "pointer",
      borderRadius: "13px",
      border: "1px solid #e6e6e6",
    };
    const silentSigninLink = this.wrappedAsSilentSigninLink(link);

    return (
      <div style={gridStyle} onClick={() => Setting.goToLinkSoft(this, silentSigninLink)}>
        <img src={logo} alt="logo" width={"100%"} style={{padding: "10px"}} />
        <Meta
          title={""}
          description={desc}
          style={{justifyContent: "center"}}
        />
      </div>
    );
  }

  renderCard(logo, link, title, desc, time, isSingle) {
    const silentSigninLink = this.wrappedAsSilentSigninLink(link);
    const date = Setting.getFormattedDateShort(time);
    return (
      // TODO: add borderRadius from app settings
      <div style={{paddingLeft: "20px", paddingRight: "20px", paddingBottom: "20px"}}>
        <div
          onClick={() => Setting.goToLinkSoft(this, silentSigninLink)}
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            width: isSingle ? "320px" : "100%",
            height: "100%",
            padding: "10px",
            borderRadius: "13px",
            border: "1px solid #e6e6e6",
            boxShadow: "0 4px 8px rgba(0, 0, 0, 0.0)",
            transition: "box-shadow 0.3s ease-in-out, transform 0.3s ease-in-out",
            cursor: "pointer",
          }}
          onMouseEnter={e => {
            e.currentTarget.style.boxShadow = "0 6px 12px rgba(0, 0, 0, 0.2)";
            e.currentTarget.style.transform = "scale(1.02)";
          }}
          onMouseLeave={e => {
            e.currentTarget.style.boxShadow = "0 4px 8px rgba(0, 0, 0, 0.0)";
            e.currentTarget.style.transform = "scale(1)";
          }}
        >
          <img
            alt="logo"
            src={logo}
            style={{width: "100%", height: "auto", padding: "10px", objectFit: "contain"}}
          />
          <div style={{width: "50%", padding: "10px", display: "flex", flexDirection: "column", justifyContent: "center", alignItems: "center"}}>
            <h3 style={{margin: 0}}>{title}</h3>
            {date && <p style={{margin: 0}}>{date}</p>}
            <p style={{margin: 0}}>{desc}</p>
          </div>
        </div>
      </div>
    );
  }

  render() {
    if (Setting.isMobile()) {
      return this.renderCardMobile(this.props.logo, this.props.link, this.props.title, this.props.desc, this.props.time, this.props.isSingle);
    } else {
      return this.renderCard(this.props.logo, this.props.link, this.props.title, this.props.desc, this.props.time, this.props.isSingle);
    }
  }
}

export default withRouter(SingleCard);
