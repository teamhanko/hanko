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
  state: State<"login_security_key">;
}

const LoginSecurityKeyPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { setLoadingAction, stateHandler } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);

  const onSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("passkey-submit");

    const nextState = await flowState.actions
      .webauthn_generate_request_options(null)
      .run();

    stateHandler[nextState.name](nextState);
  };

  const onClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("skip");
    const nextState = await flowState.actions.continue_to_login_otp(null).run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.securityKeyLogin")}</Headline1>
        <ErrorBox state={flowState} />
        <Paragraph>{t("texts.securityKeyLogin")}</Paragraph>
        <Form onSubmit={onSubmit}>
          <Button uiAction={"passkey-submit"} autofocus icon={"passkey"}>
            {t("labels.securityKeyUse")}
          </Button>
        </Form>
      </Content>
      <Footer hidden={!flowState.actions.continue_to_login_otp?.(null)}>
        <Link
          uiAction={"skip"}
          onClick={onClick}
          loadingSpinnerPosition={"left"}
          hidden={!flowState.actions.continue_to_login_otp?.(null)}
        >
          {t("labels.use_another_method")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default LoginSecurityKeyPage;
