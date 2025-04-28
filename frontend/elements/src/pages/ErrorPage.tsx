import { useCallback, useContext, useEffect, useState } from "preact/compat";
import { State, HankoError } from "@teamhanko/hanko-frontend-sdk";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Form from "../components/form/Form";
import Button from "../components/form/Button";
import Content from "../components/wrapper/Content";
import Headline1 from "../components/headline/Headline1";
import ErrorBox from "../components/error/ErrorBox";
import { useFlowState } from "../hooks/UseFlowState";

interface Props {
  state?: State<any>;
  error?: HankoError;
}

const ErrorPage = ({ state, error }: Props) => {
  const { t } = useContext(TranslateContext);
  const { init, componentName } = useContext(AppContext);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const retry = useCallback(() => init(componentName), [componentName, init]);

  const { flowState } = useFlowState(state);

  const onContinueClick = (event: Event) => {
    event.preventDefault();
    setIsLoading(true);
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
      <ErrorBox state={flowState} error={error} />
      <Form onSubmit={onContinueClick}>
        <Button isLoading={isLoading}>{t("labels.continue")}</Button>
      </Form>
    </Content>
  );
};

export default ErrorPage;
