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
import {forwardRef, memo} from "react";
import * as Setting from "../../Setting";
import {Dropdown} from "antd";
import "../../App.less";
// import {GlobalOutlined} from "@ant-design/icons";

function flagIcon(country, alt) {
  return (
    <img width={24} alt={alt} src={`${Setting.StaticBaseUrl}/flag-icons/${country}.svg`} />
  );
}

const SvgComponent = (props, ref) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width={28}
    height={28}
    viewBox="0 0 28 28"
    fill="none"
    ref={ref}
    {...props}
  >
    <g fill="#8897AD" clipPath="url(#clip0_370_33555)">
      <path d="M21.684 22.46c1.804 0 3.492-.983 4.218-2.507h.047V21.5c0 .586.399.973.95.973.562 0 .96-.387.96-1.02v-7.816c0-2.555-1.875-4.207-4.851-4.207-2.203 0-4.043.972-4.711 2.484-.129.293-.223.574-.223.82 0 .516.387.856.89.856.364 0 .645-.14.821-.469.656-1.36 1.652-1.992 3.176-1.992 1.828 0 2.918 1.02 2.918 2.66v1.008l-3.785.21c-2.965.165-4.617 1.536-4.617 3.716 0 2.238 1.722 3.738 4.207 3.738Zm.48-1.628c-1.582 0-2.66-.867-2.66-2.156 0-1.242.996-2.074 2.836-2.192l3.539-.222v1.242c0 1.863-1.652 3.328-3.715 3.328ZM1.984 22.379c.621 0 .926-.235 1.149-.89l1.512-4.137h6.914l1.511 4.136c.223.657.528.89 1.137.89.621 0 1.02-.374 1.02-.96 0-.2-.036-.387-.13-.633L9.603 6.148c-.27-.714-.75-1.078-1.5-1.078-.727 0-1.22.352-1.477 1.067l-5.496 14.66A1.691 1.691 0 0 0 1 21.43c0 .586.375.949.984.949ZM5.22 15.57l2.847-7.886h.059l2.848 7.886H5.219Z" />
    </g>
  </svg>
);
const ForwardRef = forwardRef(SvgComponent);
const GlobalIcon = memo(ForwardRef);

class LanguageSelect extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      languages: props.languages ?? Setting.Countries.map(item => item.key),
    };

    Setting.Countries.forEach((country) => {
      new Image().src = `${Setting.StaticBaseUrl}/flag-icons/${country.country}.svg`;
    });
  }

  items = Setting.Countries.map((country) => Setting.getItem(country.label, country.key, flagIcon(country.country, country.alt)));

  getOrganizationLanguages(languages) {
    const select = [];
    for (const language of languages) {
      this.items.map((item, index) => item.key === language ? select.push(item) : null);
    }
    return select;
  }

  render() {
    const languageItems = this.getOrganizationLanguages(this.state.languages);
    const onClick = (e) => {
      Setting.setLanguage(e.key);
    };

    return (
      <Dropdown menu={{items: languageItems, onClick}} >
        <div className="select-box" style={{display: languageItems.length === 0 ? "none" : null, ...this.props.style}} >
          <GlobalIcon />
        </div>
      </Dropdown>
    );
  }
}

export default LanguageSelect;
