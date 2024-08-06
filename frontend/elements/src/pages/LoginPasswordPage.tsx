import { Fragment } from "preact";
import { useContext, useEffect, useMemo, useState } from "preact/compat";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Footer from "../components/wrapper/Footer";
import Form from "../components/form/Form";
import Input from "../components/form/Input";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Link from "../components/link/Link";
import Headline1 from "../components/headline/Headline1";
import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";
import { useFlowState } from "../contexts/FlowState";

type Props = {
  state: State<"login_password">;
};

const LoginPasswordPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { stateHandler, setLoadingAction } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);
  const [password, setPassword] = useState<string>();
  const [passwordRetryAfter, setPasswordRetryAfter] = useState<number>();

  const onPasswordInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setPassword(event.target.value);
    }
  };

  const onPasswordSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("password-submit");
    const nextState = await flowState.actions
      .password_login({ password })
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  const onRecoveryClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("password-recovery");
    const nextState = await flowState.actions
      .continue_to_passcode_confirmation_recovery(null)
      .run();
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

  const onChooseMethodClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("choose-login-method");
    const nextState = await flowState.actions
      .continue_to_login_method_chooser(null)
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  const recoveryLink = useMemo(
    () => (
      <Link
        hidden={
          !flowState.actions.continue_to_passcode_confirmation_recovery?.(null)
        }
        uiAction={"password-recovery"}
        onClick={onRecoveryClick}
        loadingSpinnerPosition={"left"}
      >
        {t("labels.forgotYourPassword")}
      </Link>
    ),
    [onRecoveryClick, t],
  );

  const loginMethodChooserLink = useMemo(
    () => (
      <Link
        uiAction={"choose-login-method"}
        onClick={onChooseMethodClick}
        loadingSpinnerPosition={"left"}
      >
        {"Choose another method"}
      </Link>
    ),
    [onChooseMethodClick],
  );

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
        <ErrorBox state={flowState} />
        <Form onSubmit={onPasswordSubmit}>
          <Input
            type={"password"}
            flowInput={flowState.actions.password_login(null).inputs.password}
            autocomplete={"current-password"}
            placeholder={t("labels.password")}
            onInput={onPasswordInput}
            autofocus
          />
          <Button
            uiAction={"password-submit"}
            disabled={passwordRetryAfter > 0}
          >
            {passwordRetryAfter > 0
              ? t("labels.passwordRetryAfter", { passwordRetryAfter })
              : t("labels.signIn")}
          </Button>
        </Form>
        {flowState.actions.continue_to_login_method_chooser?.(null)
          ? recoveryLink
          : null}
      </Content>
      <Footer>
        <Link
          uiAction={"back"}
          onClick={onBackClick}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.back")}
        </Link>
        {flowState.actions.continue_to_login_method_chooser?.(null)
          ? loginMethodChooserLink
          : recoveryLink}
      </Footer>
    </Fragment>
  );
};

export default LoginPasswordPage;
