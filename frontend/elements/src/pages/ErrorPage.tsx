import { useCallback, useContext, useEffect } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Form from "../components/form/Form";
import Button from "../components/form/Button";
import Content from "../components/wrapper/Content";
import Headline1 from "../components/headline/Headline1";
import ErrorBox from "../components/error/ErrorBox";
import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";
import { HankoError } from "@teamhanko/hanko-frontend-sdk";

interface Props {
  state?: State<any>;
  error?: HankoError;
}

const ErrorPage = ({ state, error }: Props) => {
  const { t } = useContext(TranslateContext);
  const { init, componentName } = useContext(AppContext);

  const retry = useCallback(() => init(componentName), [componentName, init]);

  const onContinueClick = (event: Event) => {
    event.preventDefault();
    retry();
  };

  useEffect(() => {
    addEventListener("hankoAuthSuccess", retry);
    return () => {
      removeEventListener("hankoAuthSuccess", retry);
    };
  }, [retry]);

  return (
    <Content>
      <Headline1>{t("headlines.error")}</Headline1>
      <ErrorBox state={state} error={error} />
      <Form onSubmit={onContinueClick}>
        <Button uiAction={"retry"}>{t("labels.continue")}</Button>
      </Form>
    </Content>
  );
};

export default ErrorPage;
