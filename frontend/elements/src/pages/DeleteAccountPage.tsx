import { useContext, useMemo, useState } from "preact/compat";
import { Fragment } from "preact";

import { HankoError } from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import Footer from "../components/wrapper/Footer";
import ErrorMessage from "../components/error/ErrorMessage";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";
import Checkbox from "../components/form/Checkbox";

interface Props {
  onBack: () => void;
}

const DeleteAccountPage = ({ onBack }: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, emitEvent } = useContext(AppContext);

  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);
  const [error, setError] = useState<HankoError>(null);

  const onDeleteSubmit = (event: Event) => {
    event.preventDefault();
    setIsLoading(true);
    hanko.user
      .delete()
      .then(() => {
        setIsLoading(false);
        setIsSuccess(true);
        emitEvent("hankoProfileUserDeleted");
        return;
      })
      .catch(setError);
  };

  const onBackClick = (event: Event) => {
    event.preventDefault();
    onBack();
  };

  const isDisabled = useMemo(
    () => isLoading || isSuccess,
    [isLoading, isSuccess]
  );

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.deleteAccount")}</Headline1>
        <ErrorMessage error={error} />
        <Paragraph>{t("texts.deleteAccount")}</Paragraph>
        <Form onSubmit={onDeleteSubmit}>
          <Checkbox
            disabled={isDisabled}
            required={true}
            type={"checkbox"}
            label={t("labels.deleteAccount")}
          />
          <Button
            isLoading={isLoading}
            isSuccess={isSuccess}
            disabled={isDisabled}
          >
            {t("labels.delete")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <Link disabled={isDisabled} onClick={onBackClick}>
          {t("labels.back")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default DeleteAccountPage;
