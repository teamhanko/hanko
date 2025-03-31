import { Fragment } from "preact";
import { useCallback, useContext, useEffect, useState } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";

import Button from "../components/form/Button";
import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Footer from "../components/wrapper/Footer";
import CodeInput from "../components/form/CodeInput";
import ErrorBox from "../components/error/ErrorBox";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";
import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";
import { useFlowState } from "../contexts/FlowState";
import Link from "../components/link/Link";

interface Props {
  state: State<"login_otp">;
}

const LoginOTPPAge = (props: Props) => {
  const numberOfDigits = 6;
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);
  const [passcodeDigits, setPasscodeDigits] = useState<string[]>([]);

  const submitPasscode = useCallback(
    async (code: string) => {
      return flowState.actions.otp_code_validate.run({ otp_code: code });
    },
    [flowState],
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

  useEffect(() => {
    if (flowState.error?.code === "passcode_invalid") setPasscodeDigits([]);
  }, [flowState]);

  return (
    <Fragment>
      <Content>
        <Headline1>{t(`headlines.otpLogin`)}</Headline1>
        <ErrorBox state={flowState} />
        <Paragraph>{t("texts.otpLogin")}</Paragraph>
        <Form
          flowAction={flowState.actions.otp_code_validate}
          onSubmit={onPasscodeSubmit}
        >
          <CodeInput
            onInput={onPasscodeInput}
            passcodeDigits={passcodeDigits}
            numberOfInputs={numberOfDigits}
          />
          <Button>{t("labels.continue")}</Button>
        </Form>
      </Content>
      <Footer
        hidden={!flowState.actions.continue_to_login_security_key.enabled}
      >
        <Link
          loadingSpinnerPosition={"right"}
          flowAction={flowState.actions.continue_to_login_security_key}
        >
          {t("labels.useAnotherMethod")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default LoginOTPPAge;
