import { StateUpdater, useContext } from "preact/compat";

import { WebauthnSupport, HankoError } from "@teamhanko/hanko-frontend-sdk";

import { TranslateContext } from "@denysvuika/preact-translate";

import Form from "../form/Form";
import Button from "../form/Button";
import Paragraph from "../paragraph/Paragraph";
import Dropdown from "./Dropdown";

interface Props {
  setError: (e: HankoError) => void;
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  onPasskeySubmit: (event: Event) => Promise<void>;
}

const AddPasskeyDropdown = ({
  checkedItemID,
  setCheckedItemID,
  onPasskeySubmit,
}: Props) => {
  const { t } = useContext(TranslateContext);

  const webauthnSupported = WebauthnSupport.supported();

  return (
    <Dropdown
      name={"passkey-create-dropdown"}
      title={t("labels.createPasskey")}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    >
      <Paragraph>{t("texts.setupPasskey")}</Paragraph>
      <Form onSubmit={onPasskeySubmit}>
        <Button
          uiAction={"passkey-submit"}
          title={!webauthnSupported ? t("labels.webauthnUnsupported") : null}
        >
          {t("labels.registerAuthenticator")}
        </Button>
      </Form>
    </Dropdown>
  );
};

export default AddPasskeyDropdown;
