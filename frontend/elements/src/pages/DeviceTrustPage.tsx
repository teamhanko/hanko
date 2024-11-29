import { Fragment } from "preact";
import { useContext } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

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
  state: State<"device_trust">;
}

const DeviceTrustPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, setLoadingAction, stateHandler } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);

  const onTrustDeviceSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("trust-device-submit");
    const nextState = await flowState.actions.trust_device(null).run();
    setLoadingAction(null);
    await hanko.flow.run(nextState, stateHandler);
  };

  const onSkipClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("skip");
    const nextState = await flowState.actions.skip(null).run();
    setLoadingAction(null);
    await hanko.flow.run(nextState, stateHandler);
  };

  const onBackClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("back");
    const nextState = await flowState.actions.back(null).run();
    setLoadingAction(null);
    await hanko.flow.run(nextState, stateHandler);
  };

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.trustDevice")}</Headline1>
        <ErrorBox flowError={flowState?.error} />
        <Paragraph>{t("texts.trustDevice")}</Paragraph>
        <Form onSubmit={onTrustDeviceSubmit}>
          <Button uiAction={"trust-device-submit"}>
            {t("labels.trustDevice")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <Link
          uiAction={"back"}
          onClick={onBackClick}
          loadingSpinnerPosition={"right"}
          hidden={!flowState.actions.back?.(null)}
        >
          {t("labels.back")}
        </Link>
        <Link
          uiAction={"skip"}
          onClick={onSkipClick}
          loadingSpinnerPosition={"left"}
          hidden={!flowState.actions.skip?.(null)}
        >
          {t("labels.skip")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default DeviceTrustPage;
