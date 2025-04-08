import { Fragment, useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Input from "../components/form/Input";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Headline1 from "../components/headline/Headline1";

import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";
import { useFlowState } from "../hooks/FlowState";
import Footer from "../components/wrapper/Footer";
import Link from "../components/link/Link";

type Props = {
  state: State<"onboarding_email">;
};

const CreateEmailPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);
  const [email, setEmail] = useState<string>();

  const onEmailInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setEmail(event.target.value);
    }
  };

  const onEmailSubmit = async (event: Event) => {
    event.preventDefault();
    return flowState.actions.email_address_set.run({ email });
  };

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.createEmail")}</Headline1>
        <ErrorBox state={flowState} />
        <Form
          onSubmit={onEmailSubmit}
          flowAction={flowState.actions.email_address_set}
        >
          <Input
            type={"email"}
            autoComplete={"email"}
            autoCorrect={"off"}
            flowInput={flowState.actions.email_address_set.inputs.email}
            onInput={onEmailInput}
            placeholder={t("labels.email")}
            pattern={"^.*[^0-9]+$"}
            value={email}
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
    </Fragment>
  );
};

export default CreateEmailPage;
