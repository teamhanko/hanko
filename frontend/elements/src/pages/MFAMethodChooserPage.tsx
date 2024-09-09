import { Fragment } from "preact";
import { useContext } from "preact/compat";
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

interface Props {
  state: State<"mfa_method_chooser">;
}

const MFAMMethodChooserPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { setLoadingAction, stateHandler } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);

  const onSecurityKeySubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("password-submit");
    const nextState = await flowState.actions
      .continue_to_security_key_creation(null)
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  const onTOTPSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("passcode-submit");
    const nextState = await flowState.actions
      .continue_to_otp_secret_creation(null)
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.choose_mfa_method")}</Headline1>
        <ErrorBox flowError={flowState?.error} />
        <Paragraph>{t("texts.choose_mfa_method")}</Paragraph>
        <Form
          hidden={!flowState.actions.continue_to_security_key_creation?.(null)}
          onSubmit={onSecurityKeySubmit}
        >
          <Button secondary={true} uiAction={"passcode-submit"} icon={"mail"}>
            {t("labels.use_security_key")}
          </Button>
        </Form>
        <Form
          hidden={!flowState.actions.continue_to_otp_secret_creation?.(null)}
          onSubmit={onTOTPSubmit}
        >
          <Button
            secondary={true}
            uiAction={"password-submit"}
            icon={"password"}
          >
            {t("labels.use_authenticator_app")}
          </Button>
        </Form>
      </Content>
    </Fragment>
  );
};

export default MFAMMethodChooserPage;
