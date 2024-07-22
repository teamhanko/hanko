import { Fragment } from "preact";
import { useContext } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Footer from "../components/wrapper/Footer";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";

import Link from "../components/link/Link";
import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";
import { useFlowState } from "../contexts/FlowState";

interface Props {
  state: State<"onboarding_create_passkey">;
}

const RegisterPasskeyPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { setLoadingAction, stateHandler } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);

  const onPasskeySubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("passkey-submit");

    const nextState = await flowState.actions
      .webauthn_generate_creation_options(null)
      .run();

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

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.registerAuthenticator")}</Headline1>
        <ErrorBox state={flowState} />
        <Paragraph>{t("texts.setupPasskey")}</Paragraph>
        <Form onSubmit={onPasskeySubmit}>
          <Button uiAction={"passkey-submit"} autofocus icon={"passkey"}>
            {t("labels.registerAuthenticator")}
          </Button>
        </Form>
      </Content>
      <Footer hidden={!flowState.actions.skip?.(null) && !flowState.actions.back?.(null)}>
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

export default RegisterPasskeyPage;
