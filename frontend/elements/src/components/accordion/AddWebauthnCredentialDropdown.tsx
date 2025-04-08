import { StateUpdater, useContext } from "preact/compat";

import { WebauthnSupport, State } from "@teamhanko/hanko-frontend-sdk";

import { TranslateContext } from "@denysvuika/preact-translate";

import Form from "../form/Form";
import Button from "../form/Button";
import Paragraph from "../paragraph/Paragraph";
import Dropdown from "./Dropdown";

type CredentialType = "passkey" | "security-key";

interface Props {
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  credentialType: CredentialType;
  flowState: State<"profile_init">;
  onState(state: State<any>): Promise<void>;
}

const AddWebauthnCredentialDropdown = ({
  checkedItemID,
  setCheckedItemID,
  credentialType,
  flowState,
  onState,
}: Props) => {
  const { t } = useContext(TranslateContext);

  const webauthnSupported = WebauthnSupport.supported();

  const action =
    credentialType == "passkey"
      ? flowState.actions.webauthn_credential_create
      : flowState.actions.security_key_create;
  const onSubmit = async (event: Event) => {
    event.preventDefault();

    const nextState = await action.run(null, {
      dispatchAfterStateChangeEvent: false,
    });
    return onState(nextState);
  };

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
      <Form onSubmit={onSubmit} flowAction={action}>
        <Button
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
