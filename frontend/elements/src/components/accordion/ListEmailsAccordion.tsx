import { Fragment } from "preact";
import { StateUpdater, useContext, useMemo } from "preact/compat";

import styles from "./styles.sass";

import { TranslateContext } from "@denysvuika/preact-translate";

import Accordion from "./Accordion";
import Paragraph from "../paragraph/Paragraph";
import Headline2 from "../headline/Headline2";
import Link from "../link/Link";
import { Email } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/payload";
import { State } from "@teamhanko/hanko-frontend-sdk";

interface Props {
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  flowState: State<"profile_init">;
  onState(state: State<any>): Promise<void>;
}

const ListEmailsAccordion = ({
  checkedItemID,
  setCheckedItemID,
  flowState,
  onState,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const isDisabled = useMemo(() => false, []);

  const onEmailDelete = async (event: Event, emailID: string) => {
    event.preventDefault();
    const nextState = await flowState.actions.email_delete.run(
      {
        email_id: emailID,
      },
      { dispatchAfterStateChangeEvent: false },
    );
    return onState(nextState);
  };

  const onEmailSetPrimary = async (event: Event, emailID: string) => {
    event.preventDefault();
    const nextState = await flowState.actions.email_set_primary.run(
      {
        email_id: emailID,
      },
      { dispatchAfterStateChangeEvent: false },
    );
    return onState(nextState);
  };

  const onEmailVerify = async (event: Event, emailID: string) => {
    event.preventDefault();
    const nextState = await flowState.actions.email_verify.run(
      {
        email_id: emailID,
      },
      { dispatchAfterStateChangeEvent: false },
    );
    return onState(nextState);
  };

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
              flowAction={flowState.actions.email_set_primary}
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
              flowAction={flowState.actions.email_verify}
              onClick={(event) => onEmailVerify(event, email.id)}
              loadingSpinnerPosition={"right"}
            >
              {t("labels.verify")}
            </Link>
          </Paragraph>
        </Fragment>
      )}
      {flowState.actions.email_delete.inputs.email_id.allowed_values
        ?.map((e) => e.value)
        .includes(email.id) ? (
        <Fragment>
          <Paragraph>
            <Headline2>{t("headlines.emailDelete")}</Headline2>
            {t("texts.emailDelete")}
            <br />
            <Link
              dangerous
              flowAction={flowState.actions.email_delete}
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
      data={flowState.payload.user.emails}
      contentSelector={contents}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    />
  );
};

export default ListEmailsAccordion;
