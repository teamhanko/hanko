import { Fragment } from "preact";
import { StateUpdater, useContext } from "preact/compat";

import { HankoError } from "@teamhanko/hanko-frontend-sdk";

import { TranslateContext } from "@denysvuika/preact-translate";

import Accordion from "./Accordion";
import Paragraph from "../paragraph/Paragraph";
import Link from "../link/Link";
import Headline2 from "../headline/Headline2";
import { Passkey } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/payload";
import { AppContext } from "../../contexts/AppProvider";
import RenamePasskeyPage from "../../pages/RenamePasskeyPage";

interface Props {
  passkeys: Passkey[];
  setError: (e: HankoError) => void;
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  onBack: (event: Event) => Promise<void>;
  onPasskeyNameSubmit: (
    event: Event,
    id: string,
    name: string,
  ) => Promise<void>;
  onPasskeyDelete: (event: Event, id: string) => Promise<void>;
}

const ListPasskeysAccordion = ({
  passkeys = [],
  checkedItemID,
  setCheckedItemID,
  onBack,
  onPasskeyNameSubmit,
  onPasskeyDelete,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { setPage } = useContext(AppContext);

  const renamePasskey = (event: Event, passkey: Passkey) => {
    event.preventDefault();
    setPage(
      <RenamePasskeyPage
        oldName={uiDisplayName(passkey)}
        passkey={passkey}
        onBack={onBack}
        onPasskeyNameSubmit={onPasskeyNameSubmit}
      />,
    );
  };

  const uiDisplayName = (passkey: Passkey) => {
    if (passkey.name) {
      return passkey.name;
    }
    const alphanumeric = passkey.public_key.replace(/[\W_]/g, "");
    return `Passkey-${alphanumeric.substring(
      alphanumeric.length - 7,
      alphanumeric.length,
    )}`;
  };

  const convertTime = (t: string) => new Date(t).toLocaleString();

  const labels = (passkey: Passkey) => uiDisplayName(passkey);

  const contents = (passkey: Passkey) => (
    <Fragment>
      <Paragraph>
        <Headline2>{t("headlines.renamePasskey")}</Headline2>
        {t("texts.renamePasskey")}
        <br />
        <Link
          onClick={(event) => renamePasskey(event, passkey)}
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
          uiAction={"password-delete"}
          dangerous
          onClick={(event) => onPasskeyDelete(event, passkey.id)}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.delete")}
        </Link>
      </Paragraph>
      <Paragraph>
        <Headline2>{t("headlines.lastUsedAt")}</Headline2>
        {passkey.last_used_at ? convertTime(passkey.last_used_at) : "-"}
      </Paragraph>
      <Paragraph>
        <Headline2>{t("headlines.createdAt")}</Headline2>
        {convertTime(passkey.created_at)}
      </Paragraph>
    </Fragment>
  );
  return (
    <Accordion
      name={"passkey-edit-dropdown"}
      columnSelector={labels}
      data={passkeys}
      contentSelector={contents}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    />
  );
};

export default ListPasskeysAccordion;
