import { useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Headline1 from "../components/headline/Headline1";
import Content from "../components/wrapper/Content";
import Button from "../components/form/Button";
import Form from "../components/form/Form";

const LoginFinishedPage = () => {
  const { t } = useContext(TranslateContext);
  const { emitSuccessEvent } = useContext(AppContext);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);

  const onContinue = (event: Event) => {
    event.preventDefault();
    setIsSuccess(true);
    emitSuccessEvent();
  };

  return (
    <Content>
      <Headline1>{t("headlines.loginFinished")}</Headline1>
      <Form onSubmit={onContinue}>
        <Button autofocus isSuccess={isSuccess}>
          {t("labels.continue")}
        </Button>
      </Form>
    </Content>
  );
};

export default LoginFinishedPage;
