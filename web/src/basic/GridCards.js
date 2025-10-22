// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import {Spin} from "antd";
import i18next from "i18next";
import React from "react";
import * as Setting from "../Setting";
import SingleCard from "./SingleCard";

const GridCards = (props) => {
  const items = props.items;

  if (items === null || items === undefined) {
    return (
      <div style={{display: "flex", justifyContent: "center", alignItems: "center", marginTop: "10%"}}>
        <Spin size="large" tip={i18next.t("login:Loading")} style={{paddingTop: "10%"}} />
      </div>
    );
  }

  return (
    Setting.isMobile() ? (
      <div id={"GridCardsMobile"} style={{padding: "12px", maxWidth: "100vw", width: "100%", gap: "12px", display: "flex", flexDirection: "column", overflowX: "hidden", flex: 1, backgroundColor: "#FFF"}}>
        {items.map(item => <SingleCard key={item.link} logo={item.logo} link={item.link} title={item.name} desc={item.description} tags={item.tags} isSingle={items.length === 1} />)}
      </div>
    ) : (
      <div id={"GridCardsDesktop"} style={{borderRadius: "8px", display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))", gap: "12px"}}>
        {items.map(item => <SingleCard logo={item.logo} link={item.link} title={item.name} desc={item.description} tags={item.tags} time={item.createdTime} isSingle={items.length === 1} key={item.name} />)}
      </div>
    )
  );
};

export default GridCards;
