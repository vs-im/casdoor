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
    width={30}
    height={30}
    viewBox="0 0 24 24"
    fill="none"
    ref={ref}
    {...props}
  >
    <path
      fill="#8897AD"
      d="M12 21a8.724 8.724 0 0 1-3.5-.71 9.101 9.101 0 0 1-2.86-1.93 9.086 9.086 0 0 1-1.93-2.86A8.724 8.724 0 0 1 3 12c0-1.242.237-2.41.71-3.503a9.13 9.13 0 0 1 1.93-2.858A9.103 9.103 0 0 1 8.5 3.711 8.713 8.713 0 0 1 12 3c1.242 0 2.41.237 3.503.71a9.13 9.13 0 0 1 2.858 1.93 9.127 9.127 0 0 1 1.928 2.857A8.711 8.711 0 0 1 21 12a8.724 8.724 0 0 1-.71 3.5 9.1 9.1 0 0 1-1.93 2.86 9.132 9.132 0 0 1-2.857 1.93A8.722 8.722 0 0 1 12 21Zm0-.992a14.959 14.959 0 0 0 1.452-2.221c.38-.727.69-1.54.929-2.44H9.619c.264.95.58 1.79.948 2.516.368.727.846 1.442 1.433 2.145Zm-1.273-.15c-.467-.55-.893-1.23-1.278-2.04a11.684 11.684 0 0 1-.86-2.472H4.753c.573 1.244 1.386 2.264 2.437 3.06a7.358 7.358 0 0 0 3.536 1.452m2.546 0a7.359 7.359 0 0 0 3.536-1.452c1.052-.796 1.864-1.816 2.437-3.06h-3.834a17.304 17.304 0 0 1-.957 2.492c-.385.81-.78 1.483-1.182 2.02Zm-8.927-5.512H8.38c-.076-.41-.13-.81-.16-1.199a14.001 14.001 0 0 1-.001-2.294c.031-.39.085-.79.16-1.2H4.347a8.17 8.17 0 0 0-.256 3.561c.061.409.146.786.255 1.132m5.035 0h5.238c.076-.41.13-.803.16-1.18a14.059 14.059 0 0 0 .001-2.332 12.08 12.08 0 0 0-.16-1.18H9.38c-.075.41-.129.803-.16 1.18-.031.376-.047.765-.047 1.166 0 .401.016.79.047 1.166.031.377.086.77.161 1.18Zm6.239 0h4.034a8.067 8.067 0 0 0 .255-3.56 7.481 7.481 0 0 0-.255-1.132h-4.035c.076.41.13.81.16 1.199a14.001 14.001 0 0 1 .001 2.294c-.031.39-.085.79-.16 1.2m-.208-5.693h3.834c-.586-1.27-1.389-2.29-2.408-3.06-1.02-.77-2.208-1.26-3.565-1.47a11.71 11.71 0 0 1 1.259 2.106c.372.79.665 1.598.88 2.424Zm-5.793 0h4.762a13.82 13.82 0 0 0-.977-2.546A11.049 11.049 0 0 0 12 3.992a11.105 11.105 0 0 0-1.404 2.116 13.911 13.911 0 0 0-.977 2.546Zm-4.865 0h3.834a13.87 13.87 0 0 1 .88-2.424c.373-.79.792-1.493 1.259-2.107-1.37.21-2.56.703-3.574 1.48-1.013.777-1.813 1.794-2.4 3.05"
    />
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
