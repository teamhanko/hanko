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
import Paragraph from "../components/paragraph/Paragraph";

interface Props {
  state: State<"login_method_chooser">;
}

const LoginMethodChooserPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, setLoadingAction, stateHandler } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);

  const onPasswordSelectSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("password-submit");
    const nextState = await flowState.actions
      .continue_to_password_login(null)
      .run();
    setLoadingAction(null);
    await hanko.flow.run(nextState, stateHandler);
  };

  const onPasscodeSelectSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("passcode-submit");
    const nextState = await flowState.actions
      .continue_to_passcode_confirmation(null)
      .run();
    setLoadingAction(null);
    await hanko.flow.run(nextState, stateHandler);
  };

  const onBackClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("back");
    const nextState = await flowState.actions.back(null).run();
    setLoadingAction(null);
    await hanko.flow.run(nextState, stateHandler);
  };

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.selectLoginMethod")}</Headline1>
        <ErrorBox flowError={flowState?.error} />
        <Paragraph>{t("texts.howDoYouWantToLogin")}</Paragraph>
        <Form
          hidden={!flowState.actions.continue_to_passcode_confirmation?.(null)}
          onSubmit={onPasscodeSelectSubmit}
        >
          <Button secondary={true} uiAction={"passcode-submit"} icon={"mail"}>
            {t("labels.passcode")}
          </Button>
        </Form>
        <Form
          hidden={!flowState.actions.continue_to_password_login?.(null)}
          onSubmit={onPasswordSelectSubmit}
        >
          <Button
            secondary={true}
            uiAction={"password-submit"}
            icon={"password"}
          >
            {t("labels.password")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <Link
          uiAction={"back"}
          onClick={onBackClick}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.back")}
        </Link>
        <span hidden />
      </Footer>
    </Fragment>
  );
};

export default LoginMethodChooserPage;
