import * as preact from "preact";
import { Fragment } from "preact";
import {
  StateUpdater,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "preact/compat";

import {
  Email,
  HankoError,
  TooManyRequestsError,
} from "@teamhanko/hanko-frontend-sdk";

import styles from "./styles.sass";

import { AppContext } from "../../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Accordion from "./Accordion";
import Paragraph from "../paragraph/Paragraph";
import Headline2 from "../headline/Headline2";
import Link from "../link/Link";

import ProfilePage from "../../pages/ProfilePage";
import LoginPasscodePage from "../../pages/LoginPasscodePage";

interface Props {
  setError: (e: HankoError) => void;
  checkedItemIndex?: number;
  setCheckedItemIndex: StateUpdater<number>;
}

const ListEmailsAccordion = ({
  setError,
  checkedItemIndex,
  setCheckedItemIndex,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, user, emails, setEmails, setPage, setPasscode } =
    useContext(AppContext);

  const [isPrimaryEmailLoading, setIsPrimaryEmailLoading] =
    useState<boolean>(false);
  const [isVerificationLoading, setIsVerificationLoading] =
    useState<boolean>(false);
  const [isDeletionLoading, setIsDeletionLoading] = useState<boolean>(false);

  const isDisabled = useMemo(
    () => isPrimaryEmailLoading || isVerificationLoading || isDeletionLoading,
    [isDeletionLoading, isPrimaryEmailLoading, isVerificationLoading]
  );

  const renderPasscode = useCallback(
    (email: Email) => {
      const onBackHandler = () => setPage(<ProfilePage />);

      const showPasscodePage = (e?: HankoError) =>
        setPage(
          <LoginPasscodePage
            userID={user.id}
            emailID={email.id}
            emailAddress={email.address}
            initialError={e}
            onSuccess={() =>
              hanko.email.list().then(setEmails).then(onBackHandler)
            }
            onBack={onBackHandler}
          />
        );

      return hanko.passcode
        .initialize(user.id, email.id, true)
        .then(setPasscode)
        .then(() => showPasscodePage())
        .catch((e) => {
          if (e instanceof TooManyRequestsError) {
            showPasscodePage(e);
            return;
          }
          throw e;
        });
    },
    [hanko.email, hanko.passcode, setEmails, setPage, setPasscode, user.id]
  );

  const changePrimaryEmail = (event: Event, email: Email) => {
    event.preventDefault();
    setIsPrimaryEmailLoading(true);
    hanko.email
      .setPrimaryEmail(email.id)
      .then(() => setError(null))
      .then(() => hanko.email.list())
      .then(setEmails)
      .finally(() => setIsPrimaryEmailLoading(false))
      .catch(setError);
  };

  const deleteEmail = (event: Event, email: Email) => {
    event.preventDefault();
    setIsDeletionLoading(true);
    hanko.email
      .delete(email.id)
      .then(() => {
        setError(null);
        setCheckedItemIndex(null);
        setIsDeletionLoading(false);
        return;
      })
      .then(() => hanko.email.list())
      .then(setEmails)
      .finally(() => setIsDeletionLoading(false))
      .catch(setError);
  };

  const verifyEmail = (event: Event, email: Email) => {
    setIsVerificationLoading(true);
    renderPasscode(email)
      .finally(() => setIsVerificationLoading(false))
      .catch(setError);
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
              disabled={isDisabled}
              isLoading={isPrimaryEmailLoading}
              onClick={(event) => changePrimaryEmail(event, email)}
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
              disabled={isDisabled}
              isLoading={isVerificationLoading}
              onClick={(event) => verifyEmail(event, email)}
              loadingSpinnerPosition={"right"}
            >
              {t("labels.verify")}
            </Link>
          </Paragraph>
        </Fragment>
      )}
      {!email.is_primary ? (
        <Fragment>
          <Paragraph>
            <Headline2>{t("headlines.emailDelete")}</Headline2>
            {t("texts.emailDelete")}
            <br />
            <Link
              isLoading={isDeletionLoading}
              disabled={isDisabled}
              onClick={(event) => deleteEmail(event, email)}
              loadingSpinnerPosition={"right"}
            >
              {t("labels.delete")}
            </Link>
          </Paragraph>
        </Fragment>
      ) : (
        <Fragment>
          <Paragraph>
            <Headline2>{t("headlines.emailDelete")}</Headline2>
            {t("texts.emailDeletePrimary")}
          </Paragraph>
        </Fragment>
      )}
    </Fragment>
  );
  return (
    <Accordion
      name={"email-dropdown"}
      columnSelector={labels}
      data={emails}
      contentSelector={contents}
      checkedItemIndex={checkedItemIndex}
      setCheckedItemIndex={setCheckedItemIndex}
    />
  );
};

export default ListEmailsAccordion;
