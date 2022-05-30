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
import { RenderContext } from "../contexts/RenderProvider";

import Content from "../components/Content";
import Headline from "../components/Headline";
import Form from "../components/Form";
import Button from "../components/Button";
import ErrorMessage from "../components/ErrorMessage";
import LinkWithLoadingIndicator from "../components/LinkWithLoadingIndicator";
import Footer from "../components/Footer";
import Paragraph from "../components/Paragraph";

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

        if (!(e instanceof WebAuthnRequestCancelledError)) {
          setError(e);
        } else {
          setError(null);
        }

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
          <Button isSuccess={isSuccess} isLoading={isLoading}>
            {t("labels.registerAuthenticator")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <span hidden />
        <LinkWithLoadingIndicator
          isLoading={isSkipLoading}
          onClick={onSkipClick}
        >
          {t("labels.continue")}
        </LinkWithLoadingIndicator>
      </Footer>
    </Fragment>
  );
};

export default RegisterAuthenticator;
