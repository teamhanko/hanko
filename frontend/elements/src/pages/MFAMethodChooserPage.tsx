import { Fragment } from "preact";
import { useContext, useMemo } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Headline1 from "../components/headline/Headline1";

import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";

import { useFlowState } from "../contexts/FlowState";
import Paragraph from "../components/paragraph/Paragraph";
import Footer from "../components/wrapper/Footer";
import Link from "../components/link/Link";

interface Props {
  state: State<"mfa_method_chooser">;
}

const MFAMMethodChooserPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);

  const singleAction = useMemo(() => {
    const { actions } = flowState;

    if (
      actions.continue_to_security_key_creation.enabled &&
      !actions.continue_to_otp_secret_creation.enabled
    ) {
      return actions.continue_to_security_key_creation;
    }

    if (
      !actions.continue_to_security_key_creation.enabled &&
      actions.continue_to_otp_secret_creation.enabled
    ) {
      return actions.continue_to_otp_secret_creation;
    }

    return undefined;
  }, [flowState]);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.mfaSetUp")}</Headline1>
        <ErrorBox flowError={flowState?.error} />
        <Paragraph>{t("texts.mfaSetUp")}</Paragraph>
        {singleAction ? (
          <Form flowAction={singleAction}>
            <Button>{t("labels.continue")}</Button>
          </Form>
        ) : (
          <Fragment>
            <Form
              flowAction={flowState.actions.continue_to_security_key_creation}
            >
              <Button secondary icon={"securityKey"}>
                {t("labels.securityKey")}
              </Button>
            </Form>
            <Form
              flowAction={flowState.actions.continue_to_otp_secret_creation}
            >
              <Button secondary icon={"qrCodeScanner"}>
                {t("labels.authenticatorApp")}
              </Button>
            </Form>
          </Fragment>
        )}
      </Content>
      <Footer>
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

export default MFAMMethodChooserPage;
