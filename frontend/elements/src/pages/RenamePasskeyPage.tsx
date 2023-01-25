import * as preact from "preact";
import { Fragment } from "preact";
import { useContext, useState } from "preact/compat";

import { HankoError, WebauthnCredential } from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Input from "../components/form/Input";
import Button from "../components/form/Button";
import ErrorMessage from "../components/error/ErrorMessage";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";
import Footer from "../components/wrapper/Footer";
import Link from "../components/link/Link";

type Props = {
  oldName: string;
  credential: WebauthnCredential;
  onBack: () => void;
};

const RenamePasskeyPage = ({ credential, oldName, onBack }: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, setWebauthnCredentials } = useContext(AppContext);

  const [isPasskeyLoading, setIsPasskeyLoading] = useState<boolean>();
  const [error, setError] = useState<HankoError>(null);
  const [newName, setNewName] = useState<string>(oldName);

  const onNewNameInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setNewName(event.target.value);
    }
  };

  const onPasskeyNameSubmit = (event: Event) => {
    event.preventDefault();
    setIsPasskeyLoading(true);
    hanko.webauthn
      .updateCredential(credential.id, newName)
      .then(() => hanko.webauthn.listCredentials())
      .then(setWebauthnCredentials)
      .then(() => onBack())
      .finally(() => setIsPasskeyLoading(false))
      .catch(setError);
  };

  const onBackHandler = (event: Event) => {
    event.preventDefault();
    onBack();
  };

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.renamePasskey")}</Headline1>
        <ErrorMessage error={error} />
        <Paragraph>{t("texts.renamePasskey")}</Paragraph>
        <Form onSubmit={onPasskeyNameSubmit}>
          <Input
            type={"text"}
            name={"passkey"}
            value={newName}
            minLength={3}
            maxLength={32}
            required={true}
            placeholder={t("labels.newPasskeyName")}
            onInput={onNewNameInput}
            disabled={isPasskeyLoading}
            autofocus
          />
          <Button isLoading={isPasskeyLoading} disabled={isPasskeyLoading}>
            {t("labels.save")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <Link
          disabled={isPasskeyLoading}
          onClick={onBackHandler}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.back")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default RenamePasskeyPage;
