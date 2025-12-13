import { Fragment } from "preact";
import { Dispatch, SetStateAction, useContext } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { State, WebauthnCredential } from "@teamhanko/hanko-frontend-sdk";

import Accordion from "./Accordion";
import Paragraph from "../paragraph/Paragraph";
import Link from "../link/Link";
import Headline2 from "../headline/Headline2";
import { AppContext } from "../../contexts/AppProvider";
import RenameWebauthnCredentialPage from "../../pages/RenameWebauthnCredentialPage";

type CredentialType = "passkey" | "security-key";

interface Props {
  credentials: WebauthnCredential[];
  checkedItemID?: string;
  setCheckedItemID: Dispatch<SetStateAction<string>>;
  onBack: (event: Event) => Promise<void>;
  onCredentialNameSubmit: (
    event: Event,
    id: string,
    name: string,
  ) => Promise<void>;
  onCredentialDelete: (event: Event, id: string) => Promise<void>;
  allowCredentialDeletion?: boolean;
  credentialType: CredentialType;
  flowState: State<"profile_init">;
}

const ListWebauthnCredentialsAccordion = ({
  credentials = [],
  checkedItemID,
  setCheckedItemID,
  onBack,
  onCredentialNameSubmit,
  allowCredentialDeletion,
  credentialType,
  onCredentialDelete,
  flowState,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { setPage } = useContext(AppContext);

  const renameCredential = (
    event: Event,
    credential: WebauthnCredential,
    credentialType: CredentialType,
  ) => {
    event.preventDefault();
    setPage(
      <RenameWebauthnCredentialPage
        oldName={uiDisplayName(credential)}
        credential={credential}
        credentialType={credentialType}
        onBack={onBack}
        onCredentialNameSubmit={onCredentialNameSubmit}
        flowState={flowState}
      />,
    );
  };

  const uiDisplayName = (credential: WebauthnCredential) => {
    if (credential.name) {
      return credential.name;
    }
    const alphanumeric = credential.public_key.replace(/[\W_]/g, "");
    return `${
      credentialType === "security-key" ? "SecurityKey" : "Passkey"
    }-${alphanumeric.substring(alphanumeric.length - 7, alphanumeric.length)}`;
  };

  const convertTime = (t: string) => new Date(t).toLocaleString();

  const labels = (credential: WebauthnCredential) => uiDisplayName(credential);

  const contents = (credential: WebauthnCredential) => (
    <>
      <Paragraph>
        <Headline2>
          {credentialType === "security-key"
            ? t("headlines.renameSecurityKey")
            : t("headlines.renamePasskey")}
        </Headline2>
        {credentialType === "security-key"
          ? t("texts.renameSecurityKey")
          : t("texts.renamePasskey")}
        <br />
        <Link
          onClick={(event) =>
            renameCredential(event, credential, credentialType)
          }
          loadingSpinnerPosition={"right"}
        >
          {t("labels.rename")}
        </Link>
      </Paragraph>
      <Paragraph hidden={!allowCredentialDeletion}>
        <Headline2>
          {credentialType === "security-key"
            ? t("headlines.deleteSecurityKey")
            : t("headlines.deletePasskey")}
        </Headline2>
        {credentialType === "security-key"
          ? t("texts.deleteSecurityKey")
          : t("texts.deletePasskey")}
        <br />
        <Link
          dangerous
          flowAction={flowState.actions.webauthn_credential_delete}
          onClick={(event) => onCredentialDelete(event, credential.id)}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.delete")}
        </Link>
      </Paragraph>
      <Paragraph>
        <Headline2>{t("headlines.lastUsedAt")}</Headline2>
        {credential.last_used_at ? convertTime(credential.last_used_at) : "-"}
      </Paragraph>
      <Paragraph>
        <Headline2>{t("headlines.createdAt")}</Headline2>
        {convertTime(credential.created_at)}
      </Paragraph>
    </>
  );
  return (
    <Accordion
      name={
        credentialType === "security-key"
          ? "security-key-edit-dropdown"
          : "passkey-edit-dropdown"
      }
      columnSelector={labels}
      data={credentials}
      contentSelector={contents}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    />
  );
};

export default ListWebauthnCredentialsAccordion;
