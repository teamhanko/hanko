import { Fragment, useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

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
  const { flowState } = useFlowState(props.state);
  const [password, setPassword] = useState<string>();

  const onPasswordInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setPassword(event.target.value);
    }
  };

  const onPasswordSubmit = async (event: Event) => {
    event.preventDefault();
    return flowState.actions.register_password.run({ new_password: password });
  };

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.registerPassword")}</Headline1>
        <ErrorBox state={flowState} />
        <Paragraph>
          {t("texts.passwordFormatHint", {
            minLength:
              flowState.actions.register_password.inputs.new_password
                .min_length,
            maxLength: 72,
          })}
        </Paragraph>
        <Form
          flowAction={flowState.actions.register_password}
          onSubmit={onPasswordSubmit}
        >
          <Input
            type={"password"}
            autocomplete={"new-password"}
            flowInput={flowState.actions.register_password.inputs.new_password}
            placeholder={t("labels.newPassword")}
            onInput={onPasswordInput}
            autofocus
          />
          <Button>{t("labels.continue")}</Button>
        </Form>
      </Content>
      <Footer
        hidden={
          !flowState.actions.back.enabled && !flowState.actions.skip.enabled
        }
      >
        <Link
          loadingSpinnerPosition={"right"}
          flowAction={flowState.actions.back}
        >
          {t("labels.back")}
        </Link>
        <Link
          loadingSpinnerPosition={"left"}
          flowAction={flowState.actions.skip}
        >
          {t("labels.skip")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default CreatePasswordPage;
