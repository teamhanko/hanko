import { Fragment } from "preact";
import { useCallback, useContext, useEffect, useState } from "preact/compat";
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
  state: State<"mfa_otp_secret_creation">;
}

const CreateOTPSecretPage = (props: Props) => {
  const numberOfDigits = 6;
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);
  const { uiState, setLoadingAction, stateHandler } = useContext(AppContext);
  const [passcodeDigits, setPasscodeDigits] = useState<string[]>([]);

  const submitPasscode = useCallback(
    async (code: string) => {
      setLoadingAction("passcode-submit");

      const nextState = await flowState.actions
        .otp_code_verify({ otp_code: code })
        .run();

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

  const onBackClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("back");
    const nextState = await flowState.actions.back(null).run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  useEffect(() => {
    if (flowState.error?.code === "passcode_invalid") setPasscodeDigits([]);
  }, [flowState]);

  return (
    <Fragment>
      <Content>
        <Headline1>{t(`headlines.create_otp_secret`)}</Headline1>
        <ErrorBox state={flowState} />
        <Paragraph>{t("texts.create_otp_secret")}</Paragraph>
        <img src={flowState.payload.otp_image_source} />
        <Form onSubmit={onPasscodeSubmit}>
          <CodeInput
            onInput={onPasscodeInput}
            passcodeDigits={passcodeDigits}
            numberOfInputs={numberOfDigits}
          />
          <Button uiAction={"passcode-submit"}>
            {t("labels.create_otp_secret")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <Link
          onClick={onBackClick}
          loadingSpinnerPosition={"right"}
          isLoading={uiState.loadingAction === "back"}
        >
          {t("labels.back")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default CreateOTPSecretPage;
