import { Fragment } from "preact";
import { useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Input from "../components/form/Input";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";
import Footer from "../components/wrapper/Footer";
import Link from "../components/link/Link";
import { WebauthnCredential } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/payload";

type Props = {
  oldName: string;
  credential: WebauthnCredential;
  credentialType: "passkey" | "security-key";
  onBack: (event: Event) => Promise<void>;
  onCredentialNameSubmit: (
    event: Event,
    id: string,
    name: string,
  ) => Promise<void>;
};

const RenameWebauthnCredentialPage = ({
  onCredentialNameSubmit,
  oldName,
  onBack,
  credential,
  credentialType,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const [newName, setNewName] = useState<string>(oldName);

  const onInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setNewName(event.target.value);
    }
  };

  return (
    <Fragment>
      <Content>
        <Headline1>
          {credentialType === "security-key"
            ? t("headlines.renameSecurityKey")
            : t("headlines.renamePasskey")}
        </Headline1>
        <ErrorBox flowError={null} />
        <Paragraph>
          {credentialType === "security-key"
            ? t("texts.renameSecurityKey")
            : t("texts.renamePasskey")}
        </Paragraph>
        <Form
          onSubmit={(event: Event) =>
            onCredentialNameSubmit(event, credential.id, newName)
          }
        >
          <Input
            type={"text"}
            name={credentialType}
            value={newName}
            minLength={3}
            maxLength={32}
            required={true}
            placeholder={
              credentialType === "security-key"
                ? t("labels.newSecurityKeyName")
                : t("labels.newPasskeyName")
            }
            onInput={onInput}
            autofocus
          />
          <Button uiAction={"webauthn-credential-rename"}>
            {t("labels.save")}
          </Button>
        </Form>
      </Content>
      <Footer>
        <Link onClick={onBack} loadingSpinnerPosition={"right"}>
          {t("labels.back")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default RenameWebauthnCredentialPage;
