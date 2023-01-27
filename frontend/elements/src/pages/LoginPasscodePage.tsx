import * as preact from "preact";
import { Fragment } from "preact";
import { useContext, useEffect, useMemo, useState } from "preact/compat";

import {
  HankoError,
  PasscodeExpiredError,
  TechnicalError,
  TooManyRequestsError,
  MaxNumOfPasscodeAttemptsReachedError,
} from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Button from "../components/form/Button";
import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Footer from "../components/wrapper/Footer";
import CodeInput from "../components/form/CodeInput";
import ErrorMessage from "../components/error/ErrorMessage";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";

type Props = {
  userID: string;
  emailID: string;
  emailAddress: string;
  onSuccess: () => Promise<void>;
  onBack: () => void;
  initialError?: HankoError;
  numberOfDigits?: number;
};

const LoginPasscodePage = ({
  userID,
  emailID,
  emailAddress,
  onSuccess,
  onBack,
  numberOfDigits = 6,
  ...props
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, setUser, passcode, setPasscode } = useContext(AppContext);

  const [isLoading, setIsLoading] = useState<boolean>();
  const [isSuccess, setIsSuccess] = useState<boolean>();
  const [isResendLoading, setIsResendLoading] = useState<boolean>();
  const [isResendSuccess, setIsResendSuccess] = useState<boolean>();
  const [ttl, setTtl] = useState<number>();
  const [maxAttemptsReached, setMaxAttemptsReached] = useState<boolean>();
  const [resendAfter, setResendAfter] = useState<number>();
  const [passcodeDigits, setPasscodeDigits] = useState<string[]>([]);
  const [error, setError] = useState<HankoError>(props.initialError || null);

  const onPasscodeInput = (digits: string[]) => {
    // Automatically submit the Passcode when every input contains a digit.
    if (digits.filter((digit) => digit !== "").length === numberOfDigits) {
      passcodeSubmit(digits.join(""));
    }

    setPasscodeDigits(digits);
  };

  const passcodeSubmit = (code: string) => {
    setIsLoading(true);

    hanko.passcode
      .finalize(userID, code)
      .then(() => hanko.user.getCurrent())
      .then(setUser)
      .then(onSuccess)
      .catch((e) => {
        if (!(e instanceof TechnicalError)) {
          // Clear Passcode digits when there is no technical error.
          setPasscodeDigits([]);
        }

        if (e instanceof MaxNumOfPasscodeAttemptsReachedError) {
          setMaxAttemptsReached(true);
        }

        setIsSuccess(false);
        setIsLoading(false);
        setError(e);
      });
  };

  const onPasscodeSubmitClick = (event: Event) => {
    event.preventDefault();
    passcodeSubmit(passcodeDigits.join(""));
  };

  const onResendClick = (event: Event) => {
    event.preventDefault();

    setIsResendSuccess(false);
    setIsResendLoading(true);

    hanko.passcode
      .initialize(userID, emailID, true)
      .then((passcode) => {
        setPasscode(passcode);
        setIsResendSuccess(true);
        setPasscodeDigits([]);
        setIsResendLoading(false);
        setMaxAttemptsReached(false);
        setError(null);
        return;
      })
      .catch((e) => {
        setIsResendLoading(false);
        setIsResendSuccess(false);
        setError(e);
      });
  };

  const handleOnBack = (event: Event) => {
    event.preventDefault();
    onBack();
  };

  const disabled = useMemo(
    () => isResendLoading || isLoading || isSuccess,
    [isResendLoading, isLoading, isSuccess]
  );

  useEffect(() => {
    if (ttl <= 0 && !isSuccess) {
      setError(new PasscodeExpiredError());
    }
  }, [isSuccess, ttl]);

  useEffect(() => {
    if (!passcode) return;
    setTtl(passcode.ttl);
  }, [passcode]);

  useEffect(() => {
    if (error instanceof TooManyRequestsError) {
      setResendAfter(error.retryAfter);
    }
  }, [error]);

  useEffect(() => {
    const timer = ttl > 0 && setInterval(() => setTtl(ttl - 1), 1000);

    return () => clearInterval(timer);
  }, [ttl]);

  useEffect(() => {
    const timer =
      resendAfter > 0 &&
      setInterval(() => setResendAfter(resendAfter - 1), 1000);

    return () => clearInterval(timer);
  }, [resendAfter]);

  return (
    <Fragment>
      <Content>
        <Headline1>{t(`headlines.loginPasscode`)}</Headline1>
        <ErrorMessage error={error} />
        <Paragraph>{t("texts.enterPasscode", { emailAddress })}</Paragraph>
        <Form onSubmit={onPasscodeSubmitClick}>
          <CodeInput
            onInput={onPasscodeInput}
            passcodeDigits={passcodeDigits}
            numberOfInputs={numberOfDigits}
            disabled={ttl <= 0 || maxAttemptsReached || disabled}
          />
          <Button
            disabled={ttl <= 0 || maxAttemptsReached || disabled}
            isLoading={isLoading}
            isSuccess={isSuccess}
          >
            {t("labels.signIn")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <Link
          onClick={handleOnBack}
          disabled={disabled}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.back")}
        </Link>
        <Link
          disabled={resendAfter > 0 || disabled}
          onClick={onResendClick}
          isLoading={isResendLoading}
          isSuccess={isResendSuccess}
          loadingSpinnerPosition={"left"}
        >
          {resendAfter > 0
            ? t("labels.passcodeResendAfter", {
                passcodeResendAfter: resendAfter,
              })
            : t("labels.sendNewPasscode")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default LoginPasscodePage;
