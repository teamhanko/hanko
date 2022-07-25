import * as preact from "preact";
import { useCallback, useContext, useEffect, useState } from "preact/compat";
import { Fragment } from "preact";

import { UserInfo } from "../../lib/HankoClient";
import {
  HankoError,
  TechnicalError,
  NotFoundError,
  EmailValidationRequiredError,
  WebAuthnRequestCancelledError,
} from "../../lib/Errors";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";
import { RenderContext } from "../contexts/PageProvider";
import { UserContext } from "../contexts/UserProvider";

import Button from "../components/Button";
import InputText from "../components/InputText";
import Headline from "../components/Headline";
import Content from "../components/Content";
import Form from "../components/Form";
import Divider from "../components/Divider";
import ErrorMessage from "../components/ErrorMessage";

const LoginEmail = () => {
  const { t } = useContext(TranslateContext);
  const { email, setEmail } = useContext(UserContext);
  const { hanko, config } = useContext(AppContext);
  const {
    renderPassword,
    renderPasscode,
    emitSuccessEvent,
    renderRegisterConfirm,
  } = useContext(RenderContext);

  const [userInfo, setUserInfo] = useState<UserInfo>(null);
  const [isPasskeyLoginLoading, setIsPasskeyLoginLoading] =
    useState<boolean>(false);
  const [isPasskeyLoginSuccess, setIsPasskeyLoginSuccess] =
    useState<boolean>(false);
  const [isEmailLoginLoading, setIsEmailLoginLoading] =
    useState<boolean>(false);
  const [isEmailLoginSuccess, setIsEmailLoginSuccess] =
    useState<boolean>(false);
  const [error, setError] = useState<HankoError>(null);
  const [isAuthenticatorSupported, setIsAuthenticatorSupported] =
    useState<boolean>(null);

  // isAndroidUserAgent is used to determine whether the "Login with Passkey" button should be visible, as there is
  // currently no resident key support on Android.
  const isAndroidUserAgent =
    window.navigator.userAgent.indexOf("Android") !== -1;

  const onEmailInput = (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setEmail(event.target.value);
    }
  };

  const onEmailSubmit = (event: Event) => {
    event.preventDefault();
    setIsEmailLoginLoading(true);

    hanko.user
      .getInfo(email)
      .then((info) => setUserInfo(info))
      .catch((e) => {
        if (e instanceof NotFoundError) {
          return renderRegisterConfirm();
        } else if (e instanceof EmailValidationRequiredError) {
          return renderPasscode(e.userID, config.password.enabled, true);
        }

        throw e;
      })
      .catch((e) => {
        setIsEmailLoginLoading(false);
        setError(e);
      });
  };

  const onWebAuthnSubmit = (event: Event) => {
    event.preventDefault();
    setIsPasskeyLoginLoading(true);

    hanko.authenticator
      .login()
      .then(() => {
        setIsPasskeyLoginLoading(false);
        setIsPasskeyLoginSuccess(true);
        emitSuccessEvent();

        return;
      })
      .catch((e) => {
        setIsPasskeyLoginLoading(false);
        setError(e instanceof WebAuthnRequestCancelledError ? null : e);
      });
  };

  const renderAlternateLoginMethod = useCallback(() => {
    if (config.password.enabled) {
      renderPassword(userInfo.id).catch((e) => {
        setIsEmailLoginLoading(false);
        setError(e);
      });
    } else {
      renderPasscode(userInfo.id, false, false).catch((e) => {
        setIsEmailLoginLoading(false);
        setError(e);
      });
    }
  }, [config.password.enabled, renderPasscode, renderPassword, userInfo]);

  useEffect(() => {
    hanko.authenticator
      .isAuthenticatorSupported()
      .then((supported) => setIsAuthenticatorSupported(supported))
      .catch((e) => setError(new TechnicalError(e)));
  }, [hanko]);

  // UserID has been resolved, decide what to do next.
  useEffect(() => {
    if (
      userInfo === null ||
      config === null ||
      isAuthenticatorSupported === null
    ) {
      return;
    }

    if (userInfo.has_webauthn_credential && isAuthenticatorSupported) {
      hanko.authenticator
        .login(userInfo.id)
        .then(() => {
          setIsEmailLoginLoading(false);
          setIsEmailLoginSuccess(true);
          emitSuccessEvent();

          return;
        })
        .catch(() => {
          renderAlternateLoginMethod();
        });
    } else {
      renderAlternateLoginMethod();
    }
  }, [
    config,
    emitSuccessEvent,
    hanko.authenticator,
    isAuthenticatorSupported,
    renderAlternateLoginMethod,
    userInfo,
  ]);

  return (
    <Content>
      <Headline>{t("headlines.loginEmail")}</Headline>
      <ErrorMessage error={error} />
      <Form onSubmit={onEmailSubmit}>
        <InputText
          name={"email"}
          type={"email"}
          autocomplete={"username"}
          required={true}
          onInput={onEmailInput}
          value={email}
          label={t("labels.email")}
          pattern={"^.*[^0-9]+$"}
          autofocus
        />
        <Button isLoading={isEmailLoginLoading} isSuccess={isEmailLoginSuccess}>
          {t("labels.continue")}
        </Button>
      </Form>
      {isAuthenticatorSupported && !isAndroidUserAgent ? (
        <Fragment>
          <Divider />
          <Form onSubmit={onWebAuthnSubmit}>
            <Button
              secondary
              isLoading={isPasskeyLoginLoading}
              isSuccess={isPasskeyLoginSuccess}
            >
              {t("labels.signInPasskey")}
            </Button>
          </Form>
        </Fragment>
      ) : null}
    </Content>
  );
};

export default LoginEmail;
