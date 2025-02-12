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
import Link from "../components/link/Link";
import OTPCreationDetails from "../components/otp/OTPCreationDetails";
import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";
import { useFlowState } from "../contexts/FlowState";

interface Props {
  state: State<"mfa_otp_secret_creation">;
}

const CreateOTPSecretPage = (props: Props) => {
  const numberOfDigits = 6;
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);
  const [passcodeDigits, setPasscodeDigits] = useState<string[]>([]);

  const submitPasscode = useCallback(
    async (code: string) => {
      return flowState.actions.otp_code_verify.run({
        otp_code: code,
      });
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
        <Headline1>{t(`headlines.otpSetUp`)}</Headline1>
        <ErrorBox state={flowState} />
        <Paragraph>{t("texts.otpScanQRCode")}</Paragraph>
        <OTPCreationDetails
          src={flowState.payload.otp_image_source}
          secret={flowState.payload.otp_secret}
        />
        <Paragraph>{t("texts.otpEnterVerificationCode")}</Paragraph>
        <Form
          flowAction={flowState.actions.otp_code_verify}
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
      <Footer>
        <Link
          flowAction={flowState.actions.back}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.back")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default CreateOTPSecretPage;
