import {Button, Form, Input} from "antd";
import React, {useRef} from "react";
import * as Setting from "../Setting";
import i18next from "i18next";
import {CaptchaModal} from "./modal/CaptchaModal";
import * as UserBackend from "../backend/UserBackend";
import {SafetyOutlined} from "@ant-design/icons";

const {Search} = Input;

/**
 * @typedef {Object} EmailInputGroupProps
 * @property {Object} signupItem - The signup item object containing details for the email input.
 * @property {Object} application - The signup item object containing details for the email input.
 * @property {boolean} required - Indicates if the email input is required.
 * @property {string} invitation - The invitation code, if any.
 * @property {string} email - The current email value.
 * @property {boolean} validEmail - The current email value.
 * @property {function} setState - Function to update the component's state.
 */

/**
 * @param {EmailInputGroupProps} props - The props for the EmailInputGroup component.
 */

export function EmailInputGroup(props) {
  const {
    signupItem,
    required,
    invitation,
    email,
    setState,
    validEmail,
    application,
  } = props;

  const [confirmationSend, setConfirmationSend] = React.useState(false);
  const [visible, setVisible] = React.useState(false);
  const [buttonLeftTime, setButtonLeftTime] = React.useState(0);
  const [buttonLoading, setButtonLoading] = React.useState(false);

  const handleEmailChange = (value) => {
    setState({email: value, validEmail: Setting.isValidEmail(value)});
  };

  const handleCodeChange = (value) => {
    setState({emailCode: value});
  };

  const onButtonClickArgs = [
    email,
    "email",
    Setting.getApplicationName(application),
  ];

  const inputRef = useRef(null);

  const handleCountDown = (leftTime = 60) => {
    let leftTimeSecond = leftTime;
    setButtonLeftTime(leftTimeSecond);
    const countDown = () => {
      leftTimeSecond--;
      setButtonLeftTime(leftTimeSecond);
      if (leftTimeSecond === 0) {
        return;
      }
      setTimeout(countDown, 1000);
    };
    setTimeout(countDown, 1000);
    setTimeout(() => {
      typeof inputRef.current?.focus === "function" && inputRef.current.focus();
    }, 500);
  };

  const handleOk = (captchaType, captchaToken, clintSecret) => {
    setVisible(false);
    setButtonLoading(true);
    // if (process.env.NODE_ENV !== "production") {
    //   setTimeout(() => {
    //     setButtonLoading(false);
    //     handleCountDown(60);
    //     setConfirmationSend(true);
    //   }, 500);
    //   const a = 5;
    //   if (5 === a) {
    //     return;
    //   }
    // }

    UserBackend.sendCode(captchaType, captchaToken, clintSecret, "signup", undefined, ...onButtonClickArgs).then(res => {
      setButtonLoading(false);
      if (res) {
        setConfirmationSend(true);
        handleCountDown(60);
      }
    });
  };

  const handleCancel = () => {
    setVisible(false);
  };

  const withVerification = signupItem.rule !== "No verification";

  let emailInput = <Input
    ref={inputRef}
    className="signup-email-input"
    prefix={withVerification ? <SafetyOutlined /> : undefined}
    placeholder={withVerification ? i18next.t("code:Enter your code") : signupItem.placeholder}
    disabled={
      invitation !== undefined &&
      invitation.email !== ""
    }
    style={{minHight: "40px"}}
    onChange={(e) => withVerification ? handleCodeChange(e.target.value) : handleEmailChange(e.target.value)}
    autoComplete={withVerification ? "one-time-code" : "email"}
  />;

  let emailConfirmationInput = null;

  if (withVerification) {
    emailConfirmationInput = confirmationSend ? emailInput : null;
    emailInput = (
      <>
        <Search
          required
          className="signup-email-input"
          disabled={buttonLoading}
          style={{minHight: "40px"}}
          // prefix={<SafetyOutlined />}
          // placeholder={i18next.t("code:Enter your code")}
          enterButton={
            <Button style={{fontSize: 14, height: "37px"}} type={"primary"} disabled={!validEmail || buttonLeftTime > 0} loading={buttonLoading}>
              {buttonLeftTime > 0 ? `${buttonLeftTime} s` : buttonLoading ? i18next.t("code:Sending") : i18next.t("code:Send Code")}
            </Button>
          }
          onChange={(e) => {
            handleEmailChange(e.target.value);
          }}
          onSearch={() => setVisible(true)}
          autoComplete="email"
        />
        <CaptchaModal
          owner={application.owner}
          name={application.name}
          visible={visible}
          onOk={handleOk}
          onCancel={handleCancel}
          isCurrentProvider={false}
        />
      </>
    );
  }

  return (
    <React.Fragment>
      <Form.Item
        name="email"
        required
        className="signup-email"
        label={
          signupItem.label ? signupItem.label : i18next.t("general:Email")
        }
        rules={[
          {
            validator: (_, value) => {
              const emailValue = withVerification ? email : value;
              if (!emailValue) {
                return i18next.t("signup:Please input your Email!");
              }

              if (
                emailValue !== "" &&
                !Setting.isValidEmail(emailValue)
              ) {
                setState({validEmail: false});
                return Promise.reject(
                  i18next.t("signup:The input is not valid Email!")
                );
              }

              if (signupItem.regex) {
                const reg = new RegExp(signupItem.regex);
                if (!reg.test(emailValue)) {
                  setState({validEmail: false});
                  return Promise.reject(
                    i18next.t(
                      "signup:The input Email doesn't match the signup item regex!"
                    )
                  );
                }
              }

              setState({validEmail: true});
              return Promise.resolve();
            },
          },
        ]}
      >
        {emailInput}
      </Form.Item>
      {emailConfirmationInput && (
        <Form.Item
          name="emailCode"
          className="signup-email-code"
          label={
            signupItem.label
              ? signupItem.label
              : i18next.t("code:Confirmation code") || i18next.t("code:Email code")
          }
          rules={[
            {
              required: required,
              message: i18next.t(
                "code:Please input your verification code!"
              ),
            },
          ]}
        >
          {emailConfirmationInput}
        </Form.Item>
      )}
    </React.Fragment>
  );
}
