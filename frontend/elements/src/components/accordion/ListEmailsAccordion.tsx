import { Fragment } from "preact";
import { StateUpdater, useContext, useMemo } from "preact/compat";

import styles from "./styles.sass";

import { TranslateContext } from "@denysvuika/preact-translate";

import Accordion from "./Accordion";
import Paragraph from "../paragraph/Paragraph";
import Headline2 from "../headline/Headline2";
import Link from "../link/Link";
import { Email } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/payload";

interface Props {
  onEmailDelete: (event: Event, emailID: string) => Promise<void>;
  onEmailSetPrimary: (event: Event, emailID: string) => Promise<void>;
  onEmailVerify: (event: Event, emailID: string) => Promise<void>;
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  emails?: Email[];
  deletableEmailIDs?: string[];
}

const ListEmailsAccordion = ({
  onEmailDelete,
  onEmailSetPrimary,
  onEmailVerify,
  checkedItemID,
  setCheckedItemID,
  emails = [],
  deletableEmailIDs = [],
}: Props) => {
  const { t } = useContext(TranslateContext);
  const isDisabled = useMemo(() => false, []);

  const labels = (email: Email) => {
    const description = (
      <span className={styles.description}>
        {!email.is_verified ? (
          <Fragment>
            {" -"} {t("labels.unverifiedEmail")}
          </Fragment>
        ) : email.is_primary ? (
          <Fragment>
            {" -"} {t("labels.primaryEmail")}
          </Fragment>
        ) : null}
      </span>
    );

    return email.is_primary ? (
      <Fragment>
        <b>{email.address}</b>
        {description}
      </Fragment>
    ) : (
      <Fragment>
        {email.address}
        {description}
      </Fragment>
    );
  };

  const contents = (email: Email) => (
    <Fragment>
      {!email.is_primary ? (
        <Fragment>
          <Paragraph>
            <Headline2>{t("headlines.setPrimaryEmail")}</Headline2>
            {t("texts.setPrimaryEmail")}
            <br />
            <Link
              uiAction={"email-set-primary"}
              onClick={(event: Event) => onEmailSetPrimary(event, email.id)}
              loadingSpinnerPosition={"right"}
            >
              {t("labels.setAsPrimaryEmail")}
            </Link>
          </Paragraph>
        </Fragment>
      ) : (
        <Fragment>
          <Paragraph>
            <Headline2>{t("headlines.isPrimaryEmail")}</Headline2>
            {t("texts.isPrimaryEmail")}
          </Paragraph>
        </Fragment>
      )}
      {email.is_verified ? (
        <Fragment>
          <Paragraph>
            <Headline2>{t("headlines.emailVerified")}</Headline2>
            {t("texts.emailVerified")}
          </Paragraph>
        </Fragment>
      ) : (
        <Fragment>
          <Paragraph>
            <Headline2>{t("headlines.emailUnverified")}</Headline2>
            {t("texts.emailUnverified")}
            <br />
            <Link
              uiAction={"email-verify"}
              onClick={(event) => onEmailVerify(event, email.id)}
              loadingSpinnerPosition={"right"}
            >
              {t("labels.verify")}
            </Link>
          </Paragraph>
        </Fragment>
      )}
      {deletableEmailIDs.includes(email.id) ? (
        <Fragment>
          <Paragraph>
            <Headline2>{t("headlines.emailDelete")}</Headline2>
            {t("texts.emailDelete")}
            <br />
            <Link
              uiAction={"email-delete"}
              dangerous
              onClick={(event) => onEmailDelete(event, email.id)}
              disabled={isDisabled}
              loadingSpinnerPosition={"right"}
            >
              {t("labels.delete")}
            </Link>
          </Paragraph>
        </Fragment>
      ) : null}
      {email.identities?.length > 0 ? (
        <Fragment>
          <Paragraph>
            <Headline2>{t("headlines.connectedAccounts")}</Headline2>
            {email.identities.map((i) => i.provider).join(", ")}
          </Paragraph>
        </Fragment>
      ) : null}
    </Fragment>
  );
  return (
    <Accordion
      name={"email-edit-dropdown"}
      columnSelector={labels}
      data={emails}
      contentSelector={contents}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    />
  );
};

export default ListEmailsAccordion;
