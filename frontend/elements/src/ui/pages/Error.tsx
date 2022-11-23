import * as preact from "preact";
import { useContext, useState } from "preact/compat";

import { HankoError } from "@teamhanko/hanko-frontend-sdk";

import { TranslateContext } from "@denysvuika/preact-translate";
import { RenderContext } from "../contexts/PageProvider";

import ErrorMessage from "../components/ErrorMessage";
import Form from "../components/Form";
import Button from "../components/Button";
import Content from "../components/Content";
import Headline from "../components/Headline";

interface Props {
  initialError: HankoError;
}

const Error = ({ initialError }: Props) => {
  const { t } = useContext(TranslateContext);
  const { renderInitialize } = useContext(RenderContext);

  const [isLoading, setIsLoading] = useState<boolean>(false);

  const onContinueClick = (event: Event) => {
    event.preventDefault();
    setIsLoading(true);
    renderInitialize();
  };

  return (
    <Content>
      <Headline>{t("headlines.error")}</Headline>
      <ErrorMessage error={initialError} />
      <Form onSubmit={onContinueClick}>
        <Button isLoading={isLoading}>Continue</Button>
      </Form>
    </Content>
  );
};

export default Error;
