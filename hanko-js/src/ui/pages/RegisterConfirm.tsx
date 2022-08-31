import * as preact from "preact";
import { Fragment } from "preact";
import { useContext, useEffect, useState } from "preact/compat";

import { User } from "../../lib/HankoClient";
import { ConflictError, HankoError } from "../../lib/Errors";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";
import { UserContext } from "../contexts/UserProvider";
import { RenderContext } from "../contexts/PageProvider";

import Content from "../components/Content";
import Headline from "../components/Headline";
import Form from "../components/Form";
import Button from "../components/Button";
import Footer from "../components/Footer";
import ErrorMessage from "../components/ErrorMessage";
import Paragraph from "../components/Paragraph";

import LinkToEmailLogin from "../components/link/toEmailLogin";

const RegisterConfirm = () => {
  const { t } = useContext(TranslateContext);
  const { hanko, config } = useContext(AppContext);
  const { email } = useContext(UserContext);
  const { renderPasscode, renderRegisterAuthenticator } = useContext(RenderContext);

  const [user, setUser] = useState<User>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<HankoError>(null);

  const onConfirmSubmit = (event: Event) => {
    event.preventDefault();
    setIsLoading(true);

    hanko.user
      .create(email)
      .then((u) => setUser(u))
      .catch((e) => {
        if (e instanceof ConflictError) {
          return hanko.user.getInfo(email);
        }

        throw e;
      })
      .then((userInfo) => {
        if (userInfo) {
          return renderPasscode(userInfo.id, config.password.enabled, true);
        }
        return;
      })
      .catch((e) => {
        setIsLoading(false);
        setError(e);
      });
  };

  // User has been created
  useEffect(() => {
    if (user === null || config === null) {
      return;
    }

    if (config.email_verification_enabled) {
      renderPasscode(user.id, config.password.enabled, true).catch((e) => {
        setIsLoading(false);
        setError(e);
      });
    } else {
      renderRegisterAuthenticator();
    }
  }, [config, renderPasscode, user]);

  return (
    <Fragment>
      <Content>
        <Headline>{t("headlines.registerConfirm")}</Headline>
        <ErrorMessage error={error} />
        <Form onSubmit={onConfirmSubmit}>
          <Paragraph>{t("texts.createAccount", { email })}</Paragraph>
          <Button autofocus isLoading={isLoading}>
            {t("labels.signUp")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <span hidden />
        <LinkToEmailLogin disabled={isLoading} />
      </Footer>
    </Fragment>
  );
};

export default RegisterConfirm;
