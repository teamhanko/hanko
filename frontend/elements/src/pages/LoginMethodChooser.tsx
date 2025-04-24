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
  state: State<"login_method_chooser">;
}

const LoginMethodChooserPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.selectLoginMethod")}</Headline1>
        <ErrorBox flowError={flowState?.error} />
        <Paragraph>{t("texts.howDoYouWantToLogin")}</Paragraph>
        <Form flowAction={flowState.actions.continue_to_passcode_confirmation}>
          <Button secondary icon={"mail"}>
            {t("labels.passcode")}
          </Button>
        </Form>
        <Form flowAction={flowState.actions.continue_to_password_login}>
          <Button secondary icon={"password"}>
            {t("labels.password")}
          </Button>
        </Form>
        <Form flowAction={flowState.actions.webauthn_generate_request_options}>
          <Button secondary={true} icon={"passkey"}>
            {t("labels.passkey")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <Link
          flowAction={flowState.actions.back}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.back")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default LoginMethodChooserPage;
