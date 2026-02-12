import { useContext } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { State } from "@teamhanko/hanko-frontend-sdk";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Footer from "../components/wrapper/Footer";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";

import Link from "../components/link/Link";
import { useFlowState } from "../hooks/UseFlowState";

interface Props {
  state: State<"login_security_key">;
}

const LoginSecurityKeyPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);

  return (
    <>
      <Content>
        <Headline1>{t("headlines.securityKeyLogin")}</Headline1>
        <ErrorBox state={flowState} />
        <Paragraph>{t("texts.securityKeyLogin")}</Paragraph>
        <Form flowAction={flowState.actions.webauthn_generate_request_options}>
          <Button autofocus icon={"securityKey"}>
            {t("labels.securityKeyUse")}
          </Button>
        </Form>
      </Content>
      <Footer hidden={!flowState.actions.continue_to_login_otp.enabled}>
        <Link
          loadingSpinnerPosition={"right"}
          flowAction={flowState.actions.continue_to_login_otp}
        >
          {t("labels.useAnotherMethod")}
        </Link>
      </Footer>
    </>
  );
};

export default LoginSecurityKeyPage;
