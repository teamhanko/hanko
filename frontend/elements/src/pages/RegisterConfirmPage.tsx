import * as preact from "preact";
import { Fragment } from "preact";
import { useContext, useEffect, useState } from "preact/compat";

import { User, HankoError } from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import Footer from "../components/wrapper/Footer";
import ErrorMessage from "../components/error/ErrorMessage";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";

interface Props {
  emailAddress: string;
  onBack: () => void;
  onSuccess: () => void;
  onPasscode: (userID: string, emailID: string) => Promise<void>;
}

const RegisterConfirmPage = ({
  emailAddress,
  onSuccess,
  onPasscode,
  onBack,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, config } = useContext(AppContext);

  const [user, setUser] = useState<User>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);
  const [error, setError] = useState<HankoError>(null);

  const onConfirmSubmit = (event: Event) => {
    event.preventDefault();
    setIsLoading(true);
    hanko.user.create(emailAddress).then(setUser).catch(setError);
  };

  const onBackClick = (event: Event) => {
    event.preventDefault();
    onBack();
  };

  useEffect(() => {
    if (!user || !config) return;

    // User has been created
    if (config.emails.require_verification) {
      onPasscode(user.id, user.email_id).catch((e) => {
        setIsLoading(false);
        setError(e);
      });
    } else {
      setIsSuccess(true);
      setIsLoading(false);
      onSuccess();
    }
  }, [config, onPasscode, onSuccess, user]);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.registerConfirm")}</Headline1>
        <ErrorMessage error={error} />
        <Paragraph>{t("texts.createAccount", { emailAddress })}</Paragraph>
        <Form onSubmit={onConfirmSubmit}>
          <Button autofocus isLoading={isLoading} isSuccess={isSuccess}>
            {t("labels.signUp")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <span hidden />
        <Link disabled={isLoading} onClick={onBackClick}>
          {t("labels.back")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default RegisterConfirmPage;
