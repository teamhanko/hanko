import { Fragment } from "preact";
import { useContext } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Footer from "../components/wrapper/Footer";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";

import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";

import { useFlowState } from "../hooks/UseFlowState";
import Paragraph from "../components/paragraph/Paragraph";

interface Props {
  state: State<"credential_onboarding_chooser">;
}

const CredentialOnboardingChooserPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.setupLoginMethod")}</Headline1>
        <ErrorBox flowError={flowState?.error} />
        <Paragraph>{t("texts.selectLoginMethodForFutureLogins")}</Paragraph>
        <Form flowAction={flowState.actions.continue_to_passkey_registration}>
          <Button secondary icon={"passkey"}>
            {t("labels.passkey")}
          </Button>
        </Form>
        <Form flowAction={flowState.actions.continue_to_password_registration}>
          <Button secondary icon={"password"}>
            {t("labels.password")}
          </Button>
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

export default CredentialOnboardingChooserPage;
