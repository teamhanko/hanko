import { Fragment, useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Input from "../components/form/Input";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";

import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";
import { useFlowState } from "../contexts/FlowState";
import Footer from "../components/wrapper/Footer";
import Link from "../components/link/Link";

type Props = {
  state: State<"password_creation">;
};

const CreatePasswordPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { stateHandler, setLoadingAction } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);
  const [password, setPassword] = useState<string>();

  const onPasswordInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setPassword(event.target.value);
    }
  };

  const onPasswordSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("password-submit");
    const nextState = await flowState.actions
      .register_password({ new_password: password })
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
        <Headline1>{t("headlines.registerPassword")}</Headline1>
        <ErrorBox state={flowState} />
        <Paragraph>
          {t("texts.passwordFormatHint", {
            minLength:
              flowState.actions.register_password(null).inputs.new_password
                .min_length,
            maxLength: 72,
          })}
        </Paragraph>
        <Form onSubmit={onPasswordSubmit}>
          <Input
            type={"password"}
            autocomplete={"new-password"}
            flowInput={
              flowState.actions.register_password(null).inputs.new_password
            }
            placeholder={t("labels.newPassword")}
            onInput={onPasswordInput}
            autofocus
          />
          <Button uiAction={"password-submit"}>{t("labels.continue")}</Button>
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

export default CreatePasswordPage;
