import { useCallback, useContext, useEffect } from "preact/compat";

import { HankoError } from "@teamhanko/hanko-frontend-sdk";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Form from "../components/form/Form";
import Button from "../components/form/Button";
import Content from "../components/wrapper/Content";
import Headline1 from "../components/headline/Headline1";
import ErrorMessage from "../components/error/ErrorMessage";

import InitPage from "./InitPage";

interface Props {
  initialError: HankoError;
}

const ErrorPage = ({ initialError }: Props) => {
  const { t } = useContext(TranslateContext);
  const { setPage } = useContext(AppContext);

  const retry = useCallback(() => setPage(<InitPage />), [setPage]);

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
      <ErrorMessage error={initialError} />
      <Form onSubmit={onContinueClick}>
        <Button>{t("labels.continue")}</Button>
      </Form>
    </Content>
  );
};

export default ErrorPage;
