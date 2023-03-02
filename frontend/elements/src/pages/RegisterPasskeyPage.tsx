import * as preact from "preact";
import { Fragment } from "preact";
import { useContext, useMemo, useState } from "preact/compat";

import {
  HankoError,
  UnauthorizedError,
  UserVerificationError,
  WebauthnRequestCancelledError,
} from "@teamhanko/hanko-frontend-sdk";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import ErrorMessage from "../components/error/ErrorMessage";
import Footer from "../components/wrapper/Footer";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";

import Link from "../components/link/Link";
import ErrorPage from "./ErrorPage";

const RegisterPasskeyPage = () => {
  const { t } = useContext(TranslateContext);
  const { hanko, emitSuccessEvent, setPage } = useContext(AppContext);

  const [isPasskeyLoading, setIsPasskeyLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);
  const [isSkipLoading, setSkipIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<HankoError>(null);

  const registerWebAuthnCredential = (event: Event) => {
    event.preventDefault();
    setIsPasskeyLoading(true);

    hanko.webauthn
      .register()
      .then(() => {
        setIsSuccess(true);
        setIsPasskeyLoading(false);
        emitSuccessEvent();

        return;
      })
      .catch((e) => {
        if (
          e instanceof UnauthorizedError ||
          e instanceof UserVerificationError
        ) {
          setPage(<ErrorPage initialError={e} />);
          return;
        }

        setError(e instanceof WebauthnRequestCancelledError ? null : e);
        setIsPasskeyLoading(false);
      });
  };

  const onSkipClick = (event: Event) => {
    event.preventDefault();
    setSkipIsLoading(true);
    emitSuccessEvent();
  };

  const disabled = useMemo(
    () => isPasskeyLoading || isSkipLoading || isSuccess,
    [isPasskeyLoading, isSkipLoading, isSuccess]
  );

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.registerAuthenticator")}</Headline1>
        <ErrorMessage error={error} />
        <Paragraph>{t("texts.setupPasskey")}</Paragraph>
        <Form onSubmit={registerWebAuthnCredential}>
          <Button
            autofocus
            isSuccess={isSuccess}
            isLoading={isPasskeyLoading}
            disabled={disabled}
            icon={"passkey"}
          >
            {t("labels.registerAuthenticator")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <span hidden />
        <Link
          isLoading={isSkipLoading}
          disabled={disabled}
          onClick={onSkipClick}
          loadingSpinnerPosition={"left"}
        >
          {t("labels.skip")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default RegisterPasskeyPage;
