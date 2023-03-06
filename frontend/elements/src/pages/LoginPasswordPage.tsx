import { Fragment } from "preact";
import { useContext, useEffect, useMemo, useState } from "preact/compat";

import {
  HankoError,
  TooManyRequestsError,
  UserInfo,
} from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Footer from "../components/wrapper/Footer";
import Form from "../components/form/Form";
import Input from "../components/form/Input";
import Button from "../components/form/Button";
import ErrorMessage from "../components/error/ErrorMessage";
import Link from "../components/link/Link";
import Headline1 from "../components/headline/Headline1";

type Props = {
  userInfo: UserInfo;
  onRecovery: () => Promise<void>;
  onSuccess: () => void;
  onBack: () => void;
};

const LoginPasswordPage = ({ onSuccess, onRecovery, onBack }: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, userInfo } = useContext(AppContext);

  const [password, setPassword] = useState<string>();
  const [passwordRetryAfter, setPasswordRetryAfter] = useState<number>(
    hanko.password.getRetryAfter(userInfo.id)
  );
  const [isPasswordLoading, setIsPasswordLoading] = useState<boolean>();
  const [isPasscodeLoading, setIsPasscodeLoading] = useState<boolean>();
  const [isSuccess, setIsSuccess] = useState<boolean>();
  const [error, setError] = useState<HankoError>(null);

  const disabled = useMemo(
    () => isPasswordLoading || isPasscodeLoading || isSuccess,
    [isPasscodeLoading, isPasswordLoading, isSuccess]
  );

  const onPasswordInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setPassword(event.target.value);
    }
  };

  const onPasswordSubmit = (event: Event) => {
    event.preventDefault();
    setIsPasswordLoading(true);

    hanko.password
      .login(userInfo.id, password)
      .then(() => setIsSuccess(true))
      .then(onSuccess)
      .finally(() => setIsPasswordLoading(false))
      .catch((e) => {
        if (e instanceof TooManyRequestsError) {
          setPasswordRetryAfter(e.retryAfter);
        }
        setError(e);
      });
  };

  const onRecoveryHandler = (event: Event) => {
    event.preventDefault();
    setIsPasscodeLoading(true);
    onRecovery()
      .finally(() => setIsPasscodeLoading(false))
      .catch(setError);
  };

  const onBackHandler = (event: Event) => {
    event.preventDefault();
    onBack();
  };

  // Automatically clear the too many requests error message
  useEffect(() => {
    if (error instanceof TooManyRequestsError && passwordRetryAfter <= 0) {
      setError(null);
    }
  }, [error, passwordRetryAfter]);

  // Count down the retry after countdown
  useEffect(() => {
    const timer =
      passwordRetryAfter > 0 &&
      setInterval(() => setPasswordRetryAfter(passwordRetryAfter - 1), 1000);

    return () => clearInterval(timer);
  }, [passwordRetryAfter]);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.loginPassword")}</Headline1>
        <ErrorMessage error={error} />
        <Form onSubmit={onPasswordSubmit}>
          <Input
            type={"password"}
            name={"password"}
            autocomplete={"current-password"}
            placeholder={t("labels.password")}
            required={true}
            onInput={onPasswordInput}
            disabled={disabled}
            autofocus
          />
          <Button
            isSuccess={isSuccess}
            isLoading={isPasswordLoading}
            disabled={passwordRetryAfter > 0 || disabled}
          >
            {passwordRetryAfter > 0
              ? t("labels.passwordRetryAfter", { passwordRetryAfter })
              : t("labels.signIn")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <Link
          disabled={disabled}
          onClick={onBackHandler}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.back")}
        </Link>
        <Link
          disabled={disabled}
          onClick={onRecoveryHandler}
          isLoading={isPasscodeLoading}
          loadingSpinnerPosition={"left"}
        >
          {t("labels.forgotYourPassword")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default LoginPasswordPage;
