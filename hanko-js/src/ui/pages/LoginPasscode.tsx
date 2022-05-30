import * as preact from "preact";
import { Fragment } from "preact";
import { useCallback, useContext, useEffect, useState } from "preact/compat";

import {
  HankoError,
  PasscodeExpiredError,
  TechnicalError,
} from "../../lib/Errors";

import { UserContext } from "../contexts/UserProvider";
import { PasscodeContext } from "../contexts/PasscodeProvider";
import { TranslateContext } from "@denysvuika/preact-translate";
import { RenderContext } from "../contexts/RenderProvider";

import Button from "../components/Button";
import Content from "../components/Content";
import Headline from "../components/Headline";
import Form from "../components/Form";
import Footer from "../components/Footer";
import InputPasscode from "../components/InputPasscode";
import LinkWithLoadingIndicator from "../components/LinkWithLoadingIndicator";
import LinkBackToEmailLogin from "../components/LinkBackToEmailLogin";
import LinkBackToPasswordLogin from "../components/LinkBackToPasswordLogin";
import ErrorMessage from "../components/ErrorMessage";
import Paragraph from "../components/Paragraph";

type Props = {
  userID: string;
  recoverPassword: boolean;
  numberOfDigits?: number;
  initialError?: HankoError;
  hideBackLink?: boolean;
};

const LoginPasscode = ({
  userID,
  recoverPassword,
  numberOfDigits = 6,
  initialError,
  hideBackLink,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { eventuallyRenderEnrollment, emitSuccessEvent } =
    useContext(RenderContext);
  const { email, userInitialize } = useContext(UserContext);
  const {
    passcodeTTL,
    passcodeIsActive,
    passcodeResendAfter,
    passcodeResend,
    passcodeFinalize,
  } = useContext(PasscodeContext);

  const [isPasscodeLoading, setIsPasscodeLoading] = useState<boolean>(false);
  const [isPasscodeSuccess, setIsPasscodeSuccess] = useState<boolean>(false);
  const [isResendLoading, setIsResendLoading] = useState<boolean>(false);
  const [isResendSuccess, setIsResendSuccess] = useState<boolean>(false);
  const [passcodeDigits, setPasscodeDigits] = useState<string[]>([]);
  const [error, setError] = useState<HankoError>(initialError);

  const onPasscodeInput = (digits: string[]) => {
    // Automatically submit the Passcode if every input contains a digit.
    if (digits.filter((digit) => digit !== "").length === numberOfDigits) {
      passcodeSubmit(digits);
    }

    setPasscodeDigits(digits);
  };

  const passcodeSubmit = useCallback(
    (code: string[]) => {
      setIsPasscodeLoading(true);

      passcodeFinalize(userID, code.join(""))
        .then(() => userInitialize())
        .then((u) => eventuallyRenderEnrollment(u, recoverPassword))
        .then((rendered) => {
          if (!rendered) {
            setIsPasscodeSuccess(true);
            setIsPasscodeLoading(false);
            emitSuccessEvent();
          }

          return;
        })
        .catch((e) => {
          // Clear Passcode digits when there is no technical error.
          if (!(e instanceof TechnicalError)) {
            setPasscodeDigits([]);
          }

          setIsPasscodeSuccess(false);
          setIsPasscodeLoading(false);
          setError(e);
        });
    },
    [
      emitSuccessEvent,
      passcodeFinalize,
      recoverPassword,
      eventuallyRenderEnrollment,
      userID,
      userInitialize,
    ]
  );

  const onPasscodeSubmitClick = (event: Event) => {
    event.preventDefault();
    passcodeSubmit(passcodeDigits);
  };

  const onResendClick = (event: Event) => {
    event.preventDefault();
    setIsResendSuccess(false);
    setIsResendLoading(true);

    passcodeResend(userID)
      .then(() => {
        setIsResendSuccess(true);
        setPasscodeDigits([]);
        setIsResendLoading(false);
        setError(null);

        return;
      })
      .catch((e) => {
        setIsResendLoading(false);
        setIsResendSuccess(false);
        setError(e);
      });
  };

  useEffect(() => {
    if (passcodeTTL === 0) {
      setError(new PasscodeExpiredError());
    }
  }, [passcodeTTL]);

  return (
    <Fragment>
      <Content>
        <Headline>{t("headlines.loginPasscode")}</Headline>
        <ErrorMessage error={error} />
        <Form onSubmit={onPasscodeSubmitClick}>
          <InputPasscode
            onInput={onPasscodeInput}
            passcodeDigits={passcodeDigits}
            numberOfInputs={numberOfDigits}
            disabled={
              passcodeTTL === 0 ||
              !passcodeIsActive ||
              isPasscodeLoading ||
              isPasscodeSuccess ||
              isResendLoading
            }
          />
          <Paragraph>{t("texts.enterPasscode", { email })}</Paragraph>
          <Button
            disabled={passcodeTTL === 0 || !passcodeIsActive || isResendLoading}
            isLoading={isPasscodeLoading}
            isSuccess={isPasscodeSuccess}
          >
            {t("labels.signIn")}
          </Button>
        </Form>
      </Content>
      <Footer>
        {recoverPassword ? (
          <LinkBackToPasswordLogin
            disabled={isResendLoading || isPasscodeLoading || isPasscodeSuccess}
            userID={userID}
            hidden={hideBackLink}
          />
        ) : (
          <LinkBackToEmailLogin
            disabled={isResendLoading || isPasscodeLoading || isPasscodeSuccess}
            hidden={hideBackLink}
          />
        )}
        <LinkWithLoadingIndicator
          disabled={
            passcodeResendAfter > 0 ||
            isResendLoading ||
            isPasscodeLoading ||
            isPasscodeSuccess
          }
          onClick={onResendClick}
          isLoading={isResendLoading}
          isSuccess={isResendSuccess}
        >
          {passcodeResendAfter > 0
            ? t("labels.passcodeResendAfter", {
                passcodeResendAfter,
              })
            : t("labels.sendNewPasscode")}
        </LinkWithLoadingIndicator>
      </Footer>
    </Fragment>
  );
};

export default LoginPasscode;
