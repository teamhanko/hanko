import { useContext, useState } from "preact/compat";

import { HankoError, UnauthorizedError } from "@teamhanko/hanko-frontend-sdk";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Input from "../components/form/Input";
import Button from "../components/form/Button";
import ErrorMessage from "../components/error/ErrorMessage";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";

import ErrorPage from "./ErrorPage";

type Props = {
  onSuccess: () => void;
};

const RegisterPasswordPage = ({ onSuccess }: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, config, user, setPage } = useContext(AppContext);

  const [isLoading, setIsLoading] = useState<boolean>();
  const [isSuccess, setIsSuccess] = useState<boolean>();
  const [error, setError] = useState<HankoError>(null);
  const [password, setPassword] = useState<string>();

  const onPasswordInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setPassword(event.target.value);
    }
  };

  const onPasswordSubmit = (event: Event) => {
    event.preventDefault();
    setIsLoading(true);

    hanko.password
      .update(user.id, password)
      .then(() => setIsSuccess(true))
      .then(() => onSuccess())
      .finally(() => setIsLoading(false))
      .catch((e) => {
        if (e instanceof UnauthorizedError) {
          setPage(<ErrorPage initialError={e} />);
          return;
        }
        setError(e);
      });
  };
  return (
    <Content>
      <Headline1>{t("headlines.registerPassword")}</Headline1>
      <ErrorMessage error={error} />
      <Paragraph>
        {t("texts.passwordFormatHint", {
          minLength: config.password.min_password_length,
          maxLength: 72,
        })}
      </Paragraph>
      <Form onSubmit={onPasswordSubmit}>
        <Input
          type={"password"}
          name={"password"}
          autocomplete={"new-password"}
          minLength={config.password.min_password_length}
          maxLength={72}
          required={true}
          placeholder={t("labels.newPassword")}
          onInput={onPasswordInput}
          disabled={isSuccess || isLoading}
          autofocus
        />
        <Button isSuccess={isSuccess} isLoading={isLoading}>
          {t("labels.continue")}
        </Button>
      </Form>
    </Content>
  );
};

export default RegisterPasswordPage;
