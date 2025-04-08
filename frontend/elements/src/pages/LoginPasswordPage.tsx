import { Fragment } from "preact";
import { useContext, useEffect, useMemo, useState } from "preact/compat";

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
import { useFlowState } from "../hooks/UseFlowState";

type Props = {
  state: State<"login_password">;
};

const LoginPasswordPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
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
    return flowState.actions.password_login.run({ password });
  };

  const recoveryLink = useMemo(
    () => (
      <Link
        flowAction={
          flowState.actions.continue_to_passcode_confirmation_recovery
        }
        loadingSpinnerPosition={"left"}
      >
        {t("labels.forgotYourPassword")}
      </Link>
    ),
    [flowState, t],
  );

  const loginMethodChooserLink = useMemo(
    () => (
      <Link
        flowAction={flowState.actions.continue_to_login_method_chooser}
        loadingSpinnerPosition={"left"}
      >
        {"Choose another method"}
      </Link>
    ),
    [flowState],
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
        <Form
          flowAction={flowState.actions.password_login}
          onSubmit={onPasswordSubmit}
        >
          <Input
            type={"password"}
            flowInput={flowState.actions.password_login.inputs.password}
            autocomplete={"current-password"}
            placeholder={t("labels.password")}
            onInput={onPasswordInput}
            autofocus
          />
          <Button disabled={passwordRetryAfter > 0}>
            {passwordRetryAfter > 0
              ? t("labels.passwordRetryAfter", { passwordRetryAfter })
              : t("labels.signIn")}
          </Button>
        </Form>
        {flowState.actions.continue_to_login_method_chooser.enabled
          ? recoveryLink
          : null}
      </Content>
      <Footer>
        <Link
          flowAction={flowState.actions.back}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.back")}
        </Link>
        {flowState.actions.continue_to_login_method_chooser.enabled
          ? loginMethodChooserLink
          : recoveryLink}
      </Footer>
    </Fragment>
  );
};

export default LoginPasswordPage;
