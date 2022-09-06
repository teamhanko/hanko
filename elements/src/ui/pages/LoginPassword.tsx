import * as preact from "preact";
import { Fragment } from "preact";
import { useContext, useEffect, useState } from "preact/compat";

import {
  HankoError,
  TooManyRequestsError,
} from "@teamhanko/hanko-frontend-sdk";

import { TranslateContext } from "@denysvuika/preact-translate";
import { PasswordContext } from "../contexts/PasswordProvider";
import { UserContext } from "../contexts/UserProvider";
import { RenderContext } from "../contexts/PageProvider";

import Content from "../components/Content";
import Footer from "../components/Footer";
import Headline from "../components/Headline";
import Form from "../components/Form";
import InputText from "../components/InputText";
import Button from "../components/Button";
import ErrorMessage from "../components/ErrorMessage";

import LoadingIndicatorLink from "../components/link/withLoadingIndicator";
import LinkToEmailLogin from "../components/link/toEmailLogin";

type Props = {
  userID: string;
  initialError: HankoError;
};

const LoginPassword = ({ userID, initialError }: Props) => {
  const { t } = useContext(TranslateContext);
  const {
    eventuallyRenderEnrollment,
    renderPasscode,
    emitSuccessEvent,
    renderError,
  } = useContext(RenderContext);
  const { userInitialize } = useContext(UserContext);
  const { passwordFinalize, passwordRetryAfter } = useContext(PasswordContext);

  const [password, setPassword] = useState<string>("");
  const [isPasswordLoading, setIsPasswordLoading] = useState<boolean>(false);
  const [isPasscodeLoading, setIsPasscodeLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);
  const [error, setError] = useState<HankoError>(initialError);

  const onPasswordInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setPassword(event.target.value);
    }
  };

  const onPasswordSubmit = (event: Event) => {
    event.preventDefault();
    setIsPasswordLoading(true);

    passwordFinalize(userID, password)
      .then(() => userInitialize())
      .then((u) => eventuallyRenderEnrollment(u, false))
      .then((rendered) => {
        if (!rendered) {
          setIsSuccess(true);
          setIsPasswordLoading(false);
          emitSuccessEvent();
        }

        return;
      })
      .catch((e) => {
        setIsPasswordLoading(false);
        setError(e);
      });
  };

  const onForgotPasswordClick = () => {
    setIsPasscodeLoading(true);
    renderPasscode(userID, true, false).catch((e) => renderError(e));
  };

  // Automatically clear the too many requests error message
  useEffect(() => {
    if (error instanceof TooManyRequestsError && passwordRetryAfter <= 0) {
      setError(null);
    }
  }, [error, passwordRetryAfter]);

  return (
    <Fragment>
      <Content>
        <Headline>{t("headlines.loginPassword")}</Headline>
        <ErrorMessage error={error} />
        <Form onSubmit={onPasswordSubmit}>
          <InputText
            type={"password"}
            name={"password"}
            autocomplete={"current-password"}
            label={t("labels.password")}
            required={true}
            onInput={onPasswordInput}
            disabled={isPasswordLoading || isPasscodeLoading || isSuccess}
            autofocus
          />
          <Button
            isSuccess={isSuccess}
            isLoading={isPasswordLoading}
            disabled={
              passwordRetryAfter > 0 || isPasswordLoading || isPasscodeLoading
            }
          >
            {passwordRetryAfter > 0
              ? t("labels.passwordRetryAfter", { passwordRetryAfter })
              : t("labels.signIn")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <LinkToEmailLogin disabled={isPasscodeLoading || isPasswordLoading} />
        <LoadingIndicatorLink
          disabled={isPasscodeLoading || isPasswordLoading}
          onClick={onForgotPasswordClick}
          isLoading={isPasscodeLoading}
        >
          {t("labels.forgotYourPassword")}
        </LoadingIndicatorLink>
      </Footer>
    </Fragment>
  );
};

export default LoginPassword;
