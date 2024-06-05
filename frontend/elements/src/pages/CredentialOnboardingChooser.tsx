import { Fragment } from "preact";
import { useContext } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Footer from "../components/wrapper/Footer";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";

import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";

import { useFlowState } from "../contexts/FlowState";

interface Props {
  state: State<"credential_onboarding_chooser">;
}

const CredentialOnboardingChooserPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { setLoadingAction, stateHandler } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);

  const onPasskeySelectSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("passkey-submit");
    const nextState = await flowState.actions
      .continue_to_passkey_registration(null)
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  const onPasswordSelectSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("password-submit");
    const nextState = await flowState.actions
      .continue_to_password_registration(null)
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

  const onSkipClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("skip");
    const nextState = await flowState.actions.skip(null).run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  return (
    <Fragment>
      <Content>
        <Headline1>{"Choose a method"}</Headline1>
        <ErrorBox flowError={flowState?.error} />
        <Form
          hidden={!flowState.actions.continue_to_passkey_registration?.(null)}
          onSubmit={onPasskeySelectSubmit}
        >
          <Button secondary={true} uiAction={"passkey-submit"} icon={"passkey"}>
            {"Passkey"}
          </Button>
        </Form>
        <Form
          hidden={!flowState.actions.continue_to_password_registration?.(null)}
          onSubmit={onPasswordSelectSubmit}
        >
          <Button
            secondary={true}
            uiAction={"password-submit"}
            icon={"passkey"}
          >
            {"Password"}
          </Button>
        </Form>
      </Content>
      <Footer
        hidden={
          !flowState.actions.back?.(null) && !flowState.actions.skip?.(null)
        }
      >
        <Link
          uiAction={"back"}
          onClick={onBackClick}
          loadingSpinnerPosition={"left"}
          hidden={!flowState.actions.back?.(null)}
        >
          {t("labels.back")}
        </Link>
        <Link
          uiAction={"skip"}
          onClick={onSkipClick}
          loadingSpinnerPosition={"right"}
          hidden={!flowState.actions.skip?.(null)}
        >
          {t("labels.skip")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default CredentialOnboardingChooserPage;
