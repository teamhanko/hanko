import * as preact from "preact";
import { useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";
import { RenderContext } from "../contexts/PageProvider";

import Headline from "../components/Headline";
import Content from "../components/Content";
import Button from "../components/Button";
import Form from "../components/Form";

const LoginFinished = () => {
  const { t } = useContext(TranslateContext);
  const { emitSuccessEvent } = useContext(RenderContext);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);

  const onContinue = (event: Event) => {
    event.preventDefault();
    setIsSuccess(true);
    emitSuccessEvent();
  };

  return (
    <Content>
      <Headline>{t("headlines.loginFinished")}</Headline>
      <Form onSubmit={onContinue}>
        <Button autofocus isSuccess={isSuccess}>
          {t("labels.continue")}
        </Button>
      </Form>
    </Content>
  );
};

export default LoginFinished;
