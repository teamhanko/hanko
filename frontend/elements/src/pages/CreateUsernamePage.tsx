import { useContext, useState } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { State } from "@teamhanko/hanko-frontend-sdk";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Input from "../components/form/Input";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Headline1 from "../components/headline/Headline1";

import { useFlowState } from "../hooks/UseFlowState";
import Footer from "../components/wrapper/Footer";
import Link from "../components/link/Link";

type Props = {
  state: State<"onboarding_username">;
};

const CreateUsernamePage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);
  const [username, setUsername] = useState<string>();

  const onUsernameInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setUsername(event.target.value);
    }
  };

  const onUsernameSubmit = async (event: Event) => {
    event.preventDefault();
    return flowState.actions.username_create.run({ username });
  };

  return (
    <>
      <Content>
        <Headline1>{t("headlines.createUsername")}</Headline1>
        <ErrorBox state={flowState} />
        <Form
          flowAction={flowState.actions.username_create}
          onSubmit={onUsernameSubmit}
        >
          <Input
            type={"text"}
            autoComplete={"username"}
            autoCorrect={"off"}
            flowInput={flowState.actions.username_create.inputs.username}
            onInput={onUsernameInput}
            value={username}
            placeholder={t("labels.username")}
          />
          <Button>{t("labels.continue")}</Button>
        </Form>
      </Content>
      <Footer hidden={!flowState.actions.skip.enabled}>
        <span hidden />
        <Link
          flowAction={flowState.actions.skip}
          loadingSpinnerPosition={"left"}
        >
          {t("labels.skip")}
        </Link>
      </Footer>
    </>
  );
};

export default CreateUsernamePage;
