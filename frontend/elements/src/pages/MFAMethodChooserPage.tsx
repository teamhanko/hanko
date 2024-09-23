import { Fragment } from "preact";
import { useContext, useMemo } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Headline1 from "../components/headline/Headline1";

import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";

import { useFlowState } from "../contexts/FlowState";
import Paragraph from "../components/paragraph/Paragraph";
import Footer from "../components/wrapper/Footer";
import Link from "../components/link/Link";

interface Props {
  state: State<"mfa_method_chooser">;
}

const MFAMMethodChooserPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { setLoadingAction, stateHandler } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);

  const onSecurityKeySubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("passcode-submit");
    const nextState = await flowState.actions
      .continue_to_security_key_creation(null)
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  const onTOTPSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("password-submit");
    const nextState = await flowState.actions
      .continue_to_otp_secret_creation(null)
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  const onSkipClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("skip");
    const nextState = await flowState.actions.skip(null).run();
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

  const singleAction = useMemo(() => {
    const { actions } = flowState;

    if (
      actions.continue_to_security_key_creation &&
      !actions.continue_to_otp_secret_creation
    ) {
      return onSecurityKeySubmit;
    }

    if (
      !actions.continue_to_security_key_creation &&
      actions.continue_to_otp_secret_creation
    ) {
      return onTOTPSubmit;
    }

    return undefined;
  }, [flowState, onSecurityKeySubmit, onTOTPSubmit]);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.mfaSetUp")}</Headline1>
        <ErrorBox flowError={flowState?.error} />
        <Paragraph>{t("texts.mfaSetUp")}</Paragraph>
        {singleAction ? (
          <Form onSubmit={singleAction}>
            <Button uiAction={"passcode-submit"}>{t("labels.continue")}</Button>
          </Form>
        ) : (
          <Fragment>
            <Form
              hidden={
                !flowState.actions.continue_to_security_key_creation?.(null)
              }
              onSubmit={onSecurityKeySubmit}
            >
              <Button
                secondary
                uiAction={"passcode-submit"}
                icon={"securityKey"}
              >
                {t("labels.securityKey")}
              </Button>
            </Form>
            <Form
              hidden={
                !flowState.actions.continue_to_otp_secret_creation?.(null)
              }
              onSubmit={onTOTPSubmit}
            >
              <Button
                secondary
                uiAction={"password-submit"}
                icon={"qrCodeScanner"}
              >
                {t("labels.authenticatorApp")}
              </Button>
            </Form>
          </Fragment>
        )}
      </Content>
      <Footer>
        <Link
          uiAction={"back"}
          onClick={onBackClick}
          loadingSpinnerPosition={"right"}
          hidden={!flowState.actions.back?.(null)}
        >
          {t("labels.back")}
        </Link>
        <Link
          uiAction={"skip"}
          onClick={onSkipClick}
          loadingSpinnerPosition={"left"}
          hidden={!flowState.actions.skip?.(null)}
        >
          {t("labels.skip")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default MFAMMethodChooserPage;
