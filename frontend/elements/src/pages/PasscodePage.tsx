import { Fragment } from "preact";
import {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "preact/compat";
import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Button from "../components/form/Button";
import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Footer from "../components/wrapper/Footer";
import CodeInput from "../components/form/CodeInput";
import ErrorBox from "../components/error/ErrorBox";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";
import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";
import { useFlowState } from "../contexts/FlowState";

interface Props {
  state: State<"passcode_confirmation">;
}

const PasscodePage = (props: Props) => {
  const numberOfDigits = 6;
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);
  const {
    uiState,
    setUIState,
    setLoadingAction,
    setSucceededAction,
    stateHandler,
  } = useContext(AppContext);
  const [ttl, setTtl] = useState<number>();
  const [resendAfter, setResendAfter] = useState<number>(
    flowState.payload.resend_after,
  );
  const [passcodeDigits, setPasscodeDigits] = useState<string[]>([]);

  const maxAttemptsReached = useMemo(
    () => flowState.error?.code === "passcode_max_attempts_reached",
    [flowState],
  );

  const submitPasscode = useCallback(
    async (code: string) => {
      setLoadingAction("passcode-submit");

      const nextState = await flowState.actions.verify_passcode({ code }).run();

      setLoadingAction(null);
      stateHandler[nextState.name](nextState);
    },
    [flowState, setLoadingAction, stateHandler],
  );

  const onPasscodeInput = (digits: string[]) => {
    setPasscodeDigits(digits);
    // Automatically submit the Passcode when every input contains a digit.
    if (digits.filter((digit) => digit !== "").length === numberOfDigits) {
      return submitPasscode(digits.join(""));
    }
  };

  const onPasscodeSubmit = async (event: Event) => {
    event.preventDefault();
    return submitPasscode(passcodeDigits.join(""));
  };

  const onPasscodeResendClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("passcode-resend");
    const nextState = await flowState.actions.resend_passcode(null).run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  const onBackClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("back");
    const nextState = await flowState.actions.back(null).run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  useEffect(() => {
    if (flowState.payload.passcode_resent) {
      setSucceededAction("passcode-resend");
      setTimeout(() => setSucceededAction(null), 1000);
    }
  }, [flowState, setSucceededAction]);

  useEffect(() => {
    if (ttl <= 0 && uiState.succeededAction !== "passcode-submit") {
      // setError(new PasscodeExpiredError());
    }
  }, [uiState, ttl]);

  useEffect(() => {
    const timer = ttl > 0 && setInterval(() => setTtl(ttl - 1), 1000);
    return () => clearInterval(timer);
  }, [ttl]);

  useEffect(() => {
    const timer =
      resendAfter > 0 &&
      setInterval(() => {
        setResendAfter(resendAfter - 1);
      }, 1000);
    return () => clearInterval(timer);
  }, [resendAfter]);

  useEffect(() => {
    if (resendAfter == 0 && flowState.error?.code == "rate_limit_exceeded") {
      setUIState((prev) => ({ ...prev, error: null }));
    }
  }, [resendAfter]);

  useEffect(() => {
    if (flowState.error?.code === "passcode_invalid") setPasscodeDigits([]);
    if (flowState.payload.resend_after >= 0) {
      setResendAfter(flowState.payload.resend_after);
    }
  }, [flowState]);

  return (
    <Fragment>
      <Content>
        <Headline1>{t(`headlines.loginPasscode`)}</Headline1>
        <ErrorBox state={flowState} />
        <Paragraph>
          {uiState.email
            ? t("texts.enterPasscode", { emailAddress: uiState.email })
            : t("texts.enterPasscodeNoEmail")}
        </Paragraph>
        <Form onSubmit={onPasscodeSubmit}>
          <CodeInput
            onInput={onPasscodeInput}
            passcodeDigits={passcodeDigits}
            numberOfInputs={numberOfDigits}
            disabled={ttl <= 0 || maxAttemptsReached}
          />
          <Button
            disabled={ttl <= 0 || maxAttemptsReached}
            uiAction={"passcode-submit"}
          >
            {t("labels.continue")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <Link
          hidden={!flowState.actions.back?.(null)}
          onClick={onBackClick}
          loadingSpinnerPosition={"right"}
          isLoading={uiState.loadingAction === "back"}
        >
          {t("labels.back")}
        </Link>
        <Link
          uiAction={"passcode-resend"}
          disabled={resendAfter > 0}
          onClick={onPasscodeResendClick}
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

export default PasscodePage;
