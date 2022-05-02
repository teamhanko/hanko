import * as preact from "preact";
import { useContext, useEffect, useState } from "preact/compat";
import { Fragment } from "preact";

import {
  HankoError,
  TechnicalError,
  NotFoundError,
  EmailValidationRequiredError,
  InvalidWebauthnCredentialError,
} from "../../lib/Errors";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";
import { RenderContext } from "../contexts/RenderProvider";
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

  const [userID, setUserID] = useState<string>(null);
  const [isWebAuthnLoading, setIsWebAuthnLoading] = useState<boolean>(false);
  const [isWebAuthnSuccess, setIsWebAuthnSuccess] = useState<boolean>(false);
  const [isEmailLoading, setIsEmailLoading] = useState<boolean>(false);
  const [error, setError] = useState<HankoError>(null);
  const [isAuthenticatorSupported, setIsAuthenticatorSupported] =
    useState<boolean>(null);

  const onEmailInput = (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setEmail(event.target.value);
    }
  };

  const onEmailSubmit = (event: Event) => {
    event.preventDefault();
    setIsEmailLoading(true);

    hanko.user
      .getInfo(email)
      .then((uID) => setUserID(uID.id))
      .catch((e) => {
        if (e instanceof NotFoundError) {
          return renderRegisterConfirm();
        } else if (e instanceof EmailValidationRequiredError) {
          return renderPasscode(e.userID, config.password.enabled, true);
        }
        throw e;
      })
      .catch((e) => {
        setIsEmailLoading(false);
        setError(e);
      });
  };

  const onWebAuthnSubmit = (event: Event) => {
    event.preventDefault();
    setIsWebAuthnLoading(true);

    hanko.authenticator
      .login()
      .then(() => {
        setIsWebAuthnLoading(false);
        setIsWebAuthnSuccess(true);
        emitSuccessEvent();

        return;
      })
      .catch((e) => {
        setIsWebAuthnLoading(false);

        if (
          e instanceof TechnicalError ||
          e instanceof InvalidWebauthnCredentialError
        ) {
          setError(e);
        } else {
          setError(null);
        }
      });
  };

  useEffect(() => {
    hanko.authenticator
      .isSupported()
      .then((supported) => setIsAuthenticatorSupported(supported))
      .catch((e) => setError(new TechnicalError(e)));
  }, [hanko]);

  // UserID has been resolved, decide what to do next.
  useEffect(() => {
    if (
      userID === null ||
      config === null ||
      isAuthenticatorSupported === null
    ) {
      return;
    }

    if (config.password.enabled) {
      renderPassword(userID).catch((e) => {
        setIsEmailLoading(false);
        setError(e);
      });
    } else {
      renderPasscode(userID, false, false).catch((e) => {
        setIsEmailLoading(false);
        setError(e);
      });
    }
  }, [
    config,
    isAuthenticatorSupported,
    renderPasscode,
    renderPassword,
    userID,
  ]);

  return (
    <Content>
      <Headline>{t("headlines.loginEmail")}</Headline>
      <ErrorMessage error={error} />
      <Form onSubmit={onEmailSubmit}>
        <InputText
          name={"email"}
          type="email"
          required={true}
          onInput={onEmailInput}
          value={email}
          label={t("labels.email")}
          pattern={"^.*[^0-9]+$"}
          autofocus
        />
        <Button isLoading={isEmailLoading}>{t("labels.continue")}</Button>
      </Form>
      {isAuthenticatorSupported ? (
        <Fragment>
          <Divider />
          <Form onSubmit={onWebAuthnSubmit}>
            <Button
              useSecondaryStyles
              isLoading={isWebAuthnLoading}
              isSuccess={isWebAuthnSuccess}
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
