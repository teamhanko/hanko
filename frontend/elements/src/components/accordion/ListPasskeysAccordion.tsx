import { Fragment } from "preact";
import { SetStateAction, useContext, useState } from "preact/compat";

import {
  HankoError,
  WebauthnCredentials,
  WebauthnCredential,
} from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Accordion from "./Accordion";
import Paragraph from "../paragraph/Paragraph";
import Link from "../link/Link";
import Headline2 from "../headline/Headline2";

import ProfilePage from "../../pages/ProfilePage";
import RenamePasskeyPage from "../../pages/RenamePasskeyPage";

interface Props {
  credentials: WebauthnCredentials;
  setError: (e: HankoError) => void;
  checkedItemIndex?: number;
  setCheckedItemIndex: SetStateAction<number>;
}

const ListPasskeysAccordion = ({
  credentials,
  setError,
  checkedItemIndex,
  setCheckedItemIndex,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, setWebauthnCredentials, setPage } = useContext(AppContext);

  const [isLoading, setIsLoading] = useState<boolean>(false);

  const deletePasskey = (event: Event, credential: WebauthnCredential) => {
    event.preventDefault();
    setIsLoading(true);
    hanko.webauthn
      .deleteCredential(credential.id)
      .then(() => hanko.webauthn.listCredentials())
      .then(setWebauthnCredentials)
      .then(() => {
        setError(null);
        setCheckedItemIndex(null);
        return;
      })
      .finally(() => setIsLoading(false))
      .catch(setError);
  };

  const onBackHandler = () => {
    setPage(<ProfilePage />);
  };

  const renamePasskey = (event: Event, credential: WebauthnCredential) => {
    event.preventDefault();
    setPage(
      <RenamePasskeyPage
        oldName={uiDisplayName(credential)}
        credential={credential}
        onBack={onBackHandler}
      />
    );
  };

  const uiDisplayName = (credential: WebauthnCredential) => {
    if (credential.name) {
      return credential.name;
    }
    const alphanumeric = credential.public_key.replace(/[\W_]/g, "");
    return `Passkey-${alphanumeric.substring(
      alphanumeric.length - 7,
      alphanumeric.length
    )}`;
  };

  const convertTime = (t: string) => new Date(t).toLocaleString();

  const labels = (credential: WebauthnCredential) => uiDisplayName(credential);

  const contents = (credential: WebauthnCredential) => (
    <Fragment>
      <Paragraph>
        <Headline2>{t("headlines.renamePasskey")}</Headline2>
        {t("texts.renamePasskey")}
        <br />
        <Link
          onClick={(event) => renamePasskey(event, credential)}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.rename")}
        </Link>
      </Paragraph>
      <Paragraph>
        <Headline2>{t("headlines.deletePasskey")}</Headline2>
        {t("texts.deletePasskey")}
        <br />
        <Link
          dangerous
          isLoading={isLoading}
          onClick={(event) => deletePasskey(event, credential)}
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
    </Fragment>
  );
  return (
    <Accordion
      name={"passkey-dropdown"}
      columnSelector={labels}
      data={credentials}
      contentSelector={contents}
      checkedItemIndex={checkedItemIndex}
      setCheckedItemIndex={setCheckedItemIndex}
    />
  );
};

export default ListPasskeysAccordion;
