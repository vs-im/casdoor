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

import i18next from "i18next";
import React from "react";
import {CheckCircleTwoTone, CloseCircleTwoTone} from "@ant-design/icons";

function isValidOption_AtLeast6(password) {
  return {
    value: i18next.t("user:The password must have at least 6 characters"),
    short: i18next.t("user:At least 6 characters"),
    failed: password.length < 6,
  };
}

function isValidOption_AtLeast8(password) {
  return {
    value: i18next.t("user:The password must have at least 8 characters"),
    short: i18next.t("user:At least 8 characters"),
    failed: password.length < 8,
  };
}

function isValidOption_Aa123(password) {
  const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9]).+$/;
  return {
    value: i18next.t("user:The password must contain at least one uppercase letter, one lowercase letter and one digit"),
    short: i18next.t("user:At least one uppercase letter, one lowercase letter and one digit"),
    failed: !regex.test(password) || !password,
  };
}

function isValidOption_SpecialChar(password) {
  const regex = /^(?=.*[!-/:-@[-`{-~]).+$/;
  return {
    value: i18next.t("user:The password must contain at least one special character"),
    short: i18next.t("user:At least one special character"),
    failed: !regex.test(password),
  };
}

function isValidOption_NoRepeat(password) {
  const regex = /(.)\1+/;
  return {
    value: i18next.t("user:The password must not contain any repeated characters"),
    short: i18next.t("user:Not contain any repeated characters"),
    failed: regex.test(password) || !password,
  };
}

const checkers = {
  AtLeast6: isValidOption_AtLeast6,
  AtLeast8: isValidOption_AtLeast8,
  Aa123: isValidOption_Aa123,
  SpecialChar: isValidOption_SpecialChar,
  NoRepeat: isValidOption_NoRepeat,
};

function getOptionDescription(option, password) {
  switch (option) {
  case "AtLeast6": return i18next.t("user:The password must have at least 6 characters");
  case "AtLeast8": return i18next.t("user:The password must have at least 8 characters");
  case "Aa123": return i18next.t("user:The password must contain at least one uppercase letter, one lowercase letter and one digit");
  case "SpecialChar": return i18next.t("user:The password must contain at least one special character");
  case "NoRepeat": return i18next.t("user:The password must not contain any repeated characters");
  }
}

export function renderPasswordPopover(options, password) {
  return <div style={{width: 240}} >
    {options.map((option, idx) => {
      return <div key={idx}>{checkers[option](password) === "" ? <CheckCircleTwoTone twoToneColor={"#52c41a"} /> :
        <CloseCircleTwoTone twoToneColor={"#ff4d4f"} />} {getOptionDescription(option, password)}</div>;
    })}
  </div>;
}

export function checkPasswordComplexity(password, options) {
  let firstError = password.length === 0 ? i18next.t("login:Please input your password!") : "";
  const checkersResults = [];
  if (password.length === 0) {
    checkersResults.push({
      failed: true,
      value: firstError,
    });
  }

  if (!options || options.length === 0) {
    return "";
  }

  const checkers = {
    AtLeast6: isValidOption_AtLeast6,
    AtLeast8: isValidOption_AtLeast8,
    Aa123: isValidOption_Aa123,
    SpecialChar: isValidOption_SpecialChar,
    NoRepeat: isValidOption_NoRepeat,
  };

  for (const option of options) {
    if (option === "AtLeast6" && options.includes("AtLeast8")) {
      continue;
    }
    const checkerFunc = checkers[option];
    if (checkerFunc) {
      const {failed, short, value} = checkerFunc(password);
      checkersResults.push({
        failed: !!failed,
        value,
        short,
      });

      if (failed && !firstError) {
        firstError = value;
      }
    }
  }
  return [firstError, checkersResults];
}
