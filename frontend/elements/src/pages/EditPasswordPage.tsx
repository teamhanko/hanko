import { useContext, useState } from "preact/compat";

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

type Props = {
  state: State<"login_password_recovery">;
};

const EditPasswordPage = (props: Props) => {
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
      .password_recovery({ new_password: password })
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  return (
    <Content>
      <Headline1>{t("headlines.registerPassword")}</Headline1>
      <ErrorBox state={flowState} />
      <Paragraph>
        {t("texts.passwordFormatHint", {
          minLength:
            flowState.actions.password_recovery(null).inputs.new_password
              .min_length,
          maxLength: 72,
        })}
      </Paragraph>
      <Form onSubmit={onPasswordSubmit}>
        <Input
          type={"password"}
          autocomplete={"new-password"}
          flowInput={
            flowState.actions.password_recovery(null).inputs.new_password
          }
          placeholder={t("labels.newPassword")}
          onInput={onPasswordInput}
          autofocus
        />
        <Button uiAction={"password-submit"}>{t("labels.continue")}</Button>
      </Form>
    </Content>
  );
};

export default EditPasswordPage;
