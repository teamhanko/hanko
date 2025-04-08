import { Fragment } from "preact";
import { useContext } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Headline1 from "../components/headline/Headline1";

import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";

import { useFlowState } from "../hooks/FlowState";
import Paragraph from "../components/paragraph/Paragraph";
import Footer from "../components/wrapper/Footer";
import Link from "../components/link/Link";

interface Props {
  state: State<"device_trust">;
}

const DeviceTrustPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { flowState } = useFlowState(props.state);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.trustDevice")}</Headline1>
        <ErrorBox flowError={flowState?.error} />
        <Paragraph>{t("texts.trustDevice")}</Paragraph>
        <Form flowAction={flowState.actions.trust_device}>
          <Button>{t("labels.trustDevice")}</Button>
        </Form>
      </Content>
      <Footer>
        <Link
          flowAction={flowState.actions.back}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.back")}
        </Link>
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

export default DeviceTrustPage;
