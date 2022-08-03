import * as preact from "preact";
import { useCallback, useContext, useEffect, useState } from "preact/compat";
import { Fragment } from "preact";

import {
  HankoError,
  TechnicalError,
  NotFoundError,
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

  const loginWithEmailAndWebAuthn = () => {
    let userID: string;
    let webauthnLoginInitiated: boolean;

    return hanko.user
      .getInfo(email)
      .then((userInfo) => {
        if (!userInfo.verified) {
          return renderPasscode(userInfo.id, config.password.enabled, true);
        }

        if (!userInfo.has_webauthn_credential) {
          return renderAlternateLoginMethod(userInfo.id);
        }

        userID = userInfo.id;
        webauthnLoginInitiated = true;
        return hanko.authenticator.login(userInfo.id);
      })
      .then(() => {
        if (webauthnLoginInitiated) {
          setIsEmailLoginLoading(false);
          setIsEmailLoginSuccess(true);
          emitSuccessEvent();
        }

        return;
      })
      .catch((e) => {
        if (e instanceof NotFoundError) {
          return renderRegisterConfirm();
        }

        if (e instanceof WebAuthnRequestCancelledError) {
          return renderAlternateLoginMethod(userID);
        }

        throw e;
      });
  };

  const loginWithEmail = () => {
    return hanko.user
      .getInfo(email)
      .then((info) => {
        if (!info.verified) {
          return renderPasscode(info.id, config.password.enabled, true);
        }

        return renderAlternateLoginMethod(info.id);
      })
      .catch((e) => {
        if (e instanceof NotFoundError) {
          return renderRegisterConfirm();
        }

        throw e;
      });
  };

  const onEmailSubmit = (event: Event) => {
    event.preventDefault();
    setIsEmailLoginLoading(true);

    if (isAuthenticatorSupported) {
      loginWithEmailAndWebAuthn().catch((e) => {
        setIsEmailLoginLoading(false);
        setError(e);
      });
    } else {
      loginWithEmail().catch((e) => {
        setIsEmailLoginLoading(false);
        setError(e);
      });
    }
  };

  const onPasskeySubmit = (event: Event) => {
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

  const renderAlternateLoginMethod = useCallback(
    (userID: string) => {
      if (config.password.enabled) {
        return renderPassword(userID).catch((e) => {
          throw e;
        });
      }

      return renderPasscode(userID, false, false).catch((e) => {
        throw e;
      });
    },
    [config.password.enabled, renderPasscode, renderPassword]
  );

  useEffect(() => {
    hanko.authenticator
      .isAuthenticatorSupported()
      .then((supported) => setIsAuthenticatorSupported(supported))
      .catch((e) => setError(new TechnicalError(e)));
  }, [hanko]);

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
          disabled={
            isEmailLoginLoading ||
            isEmailLoginSuccess ||
            isPasskeyLoginLoading ||
            isPasskeyLoginSuccess
          }
          autofocus
        />
        <Button
          isLoading={isEmailLoginLoading}
          isSuccess={isEmailLoginSuccess}
          disabled={isPasskeyLoginLoading || isPasskeyLoginSuccess}
        >
          {t("labels.continue")}
        </Button>
      </Form>
      {isAuthenticatorSupported && !isAndroidUserAgent ? (
        <Fragment>
          <Divider />
          <Form onSubmit={onPasskeySubmit}>
            <Button
              secondary
              isLoading={isPasskeyLoginLoading}
              isSuccess={isPasskeyLoginSuccess}
              disabled={isEmailLoginLoading || isEmailLoginSuccess}
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
