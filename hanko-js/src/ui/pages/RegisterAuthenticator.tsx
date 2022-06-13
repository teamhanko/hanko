import * as preact from "preact";
import { Fragment } from "preact";
import { useContext, useState } from "preact/compat";

import {
  HankoError,
  UnauthorizedError,
  WebAuthnRequestCancelledError,
} from "../../lib/Errors";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";
import { RenderContext } from "../contexts/PageProvider";

import Content from "../components/Content";
import Headline from "../components/Headline";
import Form from "../components/Form";
import Button from "../components/Button";
import ErrorMessage from "../components/ErrorMessage";
import Footer from "../components/Footer";
import Paragraph from "../components/Paragraph";

import LoadingIndicatorLink from "../components/link/withLoadingIndicator";

const RegisterAuthenticator = () => {
  const { t } = useContext(TranslateContext);
  const { hanko } = useContext(AppContext);
  const { renderError, emitSuccessEvent } = useContext(RenderContext);

  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);
  const [isSkipLoading, setSkipIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<HankoError>(null);

  const registerWebAuthnCredential = (event: Event) => {
    event.preventDefault();
    setIsLoading(true);

    hanko.authenticator
      .register()
      .then(() => {
        setIsSuccess(true);
        setIsLoading(false);
        emitSuccessEvent();

        return;
      })
      .catch((e) => {
        if (e instanceof UnauthorizedError) {
          renderError(e);
          return;
        }

        setError(e instanceof WebAuthnRequestCancelledError ? null : e);
        setIsLoading(false);
      });
  };

  const onSkipClick = (event: Event) => {
    event.preventDefault();
    setSkipIsLoading(true);
    emitSuccessEvent();
  };

  return (
    <Fragment>
      <Content>
        <Headline>{t("headlines.registerAuthenticator")}</Headline>
        <ErrorMessage error={error} />
        <Form onSubmit={registerWebAuthnCredential}>
          <Paragraph>{t("texts.setupPasskey")}</Paragraph>
          <Button autofocus isSuccess={isSuccess} isLoading={isLoading}>
            {t("labels.registerAuthenticator")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <span hidden />
        <LoadingIndicatorLink isLoading={isSkipLoading} onClick={onSkipClick}>
          {t("labels.continue")}
        </LoadingIndicatorLink>
      </Footer>
    </Fragment>
  );
};

export default RegisterAuthenticator;
