import { StateUpdater, useContext } from "preact/compat";

import { WebauthnSupport, HankoError } from "@teamhanko/hanko-frontend-sdk";

import { TranslateContext } from "@denysvuika/preact-translate";

import Form from "../form/Form";
import Button from "../form/Button";
import Paragraph from "../paragraph/Paragraph";
import Dropdown from "./Dropdown";

type CredentialType = "passkey" | "security-key";

interface Props {
  setError: (e: HankoError) => void;
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  onCredentialSubmit: (event: Event) => Promise<void>;
  credentialType: CredentialType;
}

const AddWebauthnCredentialDropdown = ({
  checkedItemID,
  setCheckedItemID,
  onCredentialSubmit,
  credentialType,
}: Props) => {
  const { t } = useContext(TranslateContext);

  const webauthnSupported = WebauthnSupport.supported();

  return (
    <Dropdown
      name={
        credentialType === "security-key"
          ? "security-key-create-dropdown"
          : "passkey-create-dropdown"
      }
      title={
        credentialType === "security-key"
          ? t("labels.createSecurityKey")
          : t("labels.createPasskey")
      }
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    >
      <Paragraph>
        {credentialType === "security-key"
          ? t("texts.securityKeySetUp")
          : t("texts.setupPasskey")}
      </Paragraph>
      <Form onSubmit={onCredentialSubmit}>
        <Button
          uiAction={
            credentialType === "security-key"
              ? "security-key-submit"
              : "passkey-submit"
          }
          title={!webauthnSupported ? t("labels.webauthnUnsupported") : null}
        >
          {credentialType === "security-key"
            ? t("labels.createSecurityKey")
            : t("labels.createPasskey")}
        </Button>
      </Form>
    </Dropdown>
  );
};

export default AddWebauthnCredentialDropdown;
