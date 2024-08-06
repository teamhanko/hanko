import { Fragment, useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Input from "../components/form/Input";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Headline1 from "../components/headline/Headline1";

import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";
import { useFlowState } from "../contexts/FlowState";
import Footer from "../components/wrapper/Footer";
import Link from "../components/link/Link";

type Props = {
  state: State<"onboarding_username">;
};

const CreateUsernamePage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { stateHandler, setLoadingAction } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);
  const [username, setUsername] = useState<string>();

  const onUsernameInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setUsername(event.target.value);
    }
  };

  const onUsernameSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("username-set");
    const nextState = await flowState.actions
      .username_create({ username })
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

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.createUsername")}</Headline1>
        <ErrorBox state={flowState} />
        <Form onSubmit={onUsernameSubmit}>
          <Input
            type={"text"}
            autoComplete={"username"}
            autoCorrect={"off"}
            flowInput={
              flowState.actions.username_create?.(null).inputs.username
            }
            onInput={onUsernameInput}
            value={username}
            placeholder={t("labels.username")}
          />
          <Button uiAction={"username-set"}>{t("labels.continue")}</Button>
        </Form>
      </Content>
      <Footer hidden={!flowState.actions.skip?.(null)}>
        <span hidden />
        <Link
          uiAction={"skip"}
          onClick={onSkipClick}
          loadingSpinnerPosition={"left"}
        >
          {t("labels.skip")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default CreateUsernamePage;
