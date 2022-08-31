import * as preact from "preact";
import { useContext, useState } from "preact/compat";

import { HankoError, UnauthorizedError } from "../../lib/Errors";
import { User } from "../../lib/HankoClient";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";
import { RenderContext } from "../contexts/PageProvider";

import Content from "../components/Content";
import Headline from "../components/Headline";
import Form from "../components/Form";
import InputText from "../components/InputText";
import Button from "../components/Button";
import ErrorMessage from "../components/ErrorMessage";
import Paragraph from "../components/Paragraph";

type Props = {
  user: User;
  registerAuthenticator: boolean;
};

const RegisterPassword = ({ user, registerAuthenticator }: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko } = useContext(AppContext);
  const { renderError, emitSuccessEvent, renderRegisterAuthenticator } =
    useContext(RenderContext);

  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);
  const [error, setError] = useState<HankoError>(null);
  const [password, setPassword] = useState<string>("");

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
      .then(() => {
        if (registerAuthenticator) {
          renderRegisterAuthenticator();
        } else {
          emitSuccessEvent();
          setIsSuccess(true);
        }

        setIsLoading(false);

        return;
      })
      .catch((e) => {
        if (e instanceof UnauthorizedError) {
          renderError(e);

          return;
        }

        setIsLoading(false);
        setError(e);
      });
  };

  return (
    <Content>
      <Headline>{t("headlines.registerPassword")}</Headline>
      <ErrorMessage error={error} />
      <Form onSubmit={onPasswordSubmit}>
        <InputText
          type={"password"}
          name={"password"}
          autocomplete={"new-password"}
          required={true}
          label={t("labels.password")}
          onInput={onPasswordInput}
          disabled={isSuccess || isLoading}
          minLength={10}
          maxLength={32}
          autofocus
        />
        <Paragraph>{t("texts.passwordFormatHint")}</Paragraph>
        <Button isSuccess={isSuccess} isLoading={isLoading}>
          {t("labels.continue")}
        </Button>
      </Form>
    </Content>
  );
};

export default RegisterPassword;
