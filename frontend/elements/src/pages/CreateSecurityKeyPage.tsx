import { Fragment } from "preact";
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
  state: State<"mfa_security_key_creation">;
}

const CreateSecurityKeyPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.securityKeySetUp")}</Headline1>
        <ErrorBox state={flowState} />
        <Paragraph>{t("texts.securityKeySetUp")}</Paragraph>
        <Form flowAction={flowState.actions.webauthn_generate_creation_options}>
          <Button autofocus icon={"securityKey"}>
            {t("labels.createSecurityKey")}
          </Button>
        </Form>
      </Content>
      <Footer hidden={!flowState.actions.back.enabled}>
        <Link
          loadingSpinnerPosition={"right"}
          flowAction={flowState.actions.back}
        >
          {t("labels.back")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default CreateSecurityKeyPage;
