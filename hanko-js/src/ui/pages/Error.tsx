import * as preact from "preact";
import { useContext, useEffect, useState } from "preact/compat";

import { HankoError, UnauthorizedError } from "../../lib/Error";

import { AppContext } from "../contexts/AppProvider";
import { UserContext } from "../contexts/UserProvider";
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
  const { config, configInitialize } = useContext(AppContext);
  const { userInitialize } = useContext(UserContext);
  const {
    eventuallyRenderEnrollment,
    renderLoginEmail,
    emitSuccessEvent,
    renderError,
  } = useContext(RenderContext);

  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);
  const [error, setError] = useState<HankoError>(initialError);

  const onContinueClick = (event: Event) => {
    event.preventDefault();
    setIsLoading(true);

    configInitialize().catch((e) => {
      setIsLoading(false);
      renderError(e);
    });
  };

  useEffect(() => {
    if (config === null) {
      return;
    }

    userInitialize()
      .then((u) => eventuallyRenderEnrollment(u, false))
      .then((rendered) => {
        if (!rendered) {
          setIsSuccess(true);
          emitSuccessEvent();
        }

        return;
      })
      .catch((e) => {
        if (e instanceof UnauthorizedError) {
          renderLoginEmail();
        } else {
          setIsLoading(false);
          setError(e);
        }
      });
  }, [
    config,
    emitSuccessEvent,
    eventuallyRenderEnrollment,
    renderLoginEmail,
    userInitialize,
  ]);

  return (
    <Content>
      <Headline>{t("headlines.error")}</Headline>
      <ErrorMessage error={error} />
      <Form onSubmit={onContinueClick}>
        <Button isLoading={isLoading} isSuccess={isSuccess}>
          Continue
        </Button>
      </Form>
    </Content>
  );
};

export default Error;
