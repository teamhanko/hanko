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
  state: State<"onboarding_email">;
};

const CreateEmailPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { stateHandler, setLoadingAction } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);
  const [email, setEmail] = useState<string>();

  const onEmailInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setEmail(event.target.value);
    }
  };

  const onEmailSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("email-submit");
    const nextState = await flowState.actions
      .email_address_set({ email })
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
        <Headline1>{t("headlines.createEmail")}</Headline1>
        <ErrorBox state={flowState} />
        <Form onSubmit={onEmailSubmit}>
          <Input
            type={"email"}
            autoComplete={"email"}
            autoCorrect={"off"}
            flowInput={flowState.actions.email_address_set?.(null).inputs.email}
            onInput={onEmailInput}
            placeholder={t("labels.email")}
            pattern={"^.*[^0-9]+$"}
          />
          <Button uiAction={"email-submit"}>{t("labels.continue")}</Button>
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

export default CreateEmailPage;
