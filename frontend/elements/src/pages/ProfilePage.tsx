import { Fragment } from "preact";
import { useContext, useEffect, useState } from "preact/compat";

import { HankoError } from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Headline1 from "../components/headline/Headline1";
import Paragraph from "../components/paragraph/Paragraph";
import ErrorMessage from "../components/error/ErrorMessage";
import ListEmailsAccordion from "../components/accordion/ListEmailsAccordion";
import ListPasskeysAccordion from "../components/accordion/ListPasskeysAccordion";
import AddEmailDropdown from "../components/accordion/AddEmailDropdown";
import ChangePasswordDropdown from "../components/accordion/ChangePasswordDropdown";
import AddPasskeyDropdown from "../components/accordion/AddPasskeyDropdown";
import Divider from "../components/spacer/Divider";
import Button from "../components/form/Button";
import Form from "../components/form/Form";
import DeleteAccountPage from "./DeleteAccountPage";
import Spacer from "../components/spacer/Spacer";

const ProfilePage = () => {
  const { t } = useContext(TranslateContext);
  const { config, webauthnCredentials, emails, setPage } =
    useContext(AppContext);

  const [emailError, setEmailError] = useState<HankoError>(null);
  const [passwordError, setPasswordError] = useState<HankoError>(null);
  const [passkeyError, setPasskeyError] = useState<HankoError>(null);

  const [checkedItemIndexEmails, setCheckedItemIndexEmails] =
    useState<number>(null);
  const [checkedItemIndexAddEmail, setCheckedItemIndexAddEmail] =
    useState<number>(null);
  const [checkedItemIndexSetPassword, setCheckedItemIndexSetPassword] =
    useState<number>(null);
  const [checkedItemIndexPasskeys, setCheckedItemIndexPasskeys] =
    useState<number>(null);
  const [checkedItemIndexAddPasskey, setCheckedItemIndexAddPasskey] =
    useState<number>(null);

  const deleteUser = (event: Event) => {
    event.preventDefault();
    setPage(<DeleteAccountPage onBack={() => setPage(<ProfilePage />)} />);
  };

  useEffect(() => {
    if (checkedItemIndexEmails !== null) {
      setCheckedItemIndexAddEmail(null);
      setCheckedItemIndexSetPassword(null);
      setCheckedItemIndexPasskeys(null);
      setCheckedItemIndexAddPasskey(null);
    }
  }, [checkedItemIndexEmails]);

  useEffect(() => {
    if (checkedItemIndexAddEmail !== null) {
      setCheckedItemIndexEmails(null);
      setCheckedItemIndexSetPassword(null);
      setCheckedItemIndexPasskeys(null);
      setCheckedItemIndexAddPasskey(null);
    }
  }, [checkedItemIndexAddEmail]);

  useEffect(() => {
    if (checkedItemIndexSetPassword !== null) {
      setCheckedItemIndexAddEmail(null);
      setCheckedItemIndexEmails(null);
      setCheckedItemIndexPasskeys(null);
      setCheckedItemIndexAddPasskey(null);
    }
  }, [checkedItemIndexSetPassword]);

  useEffect(() => {
    if (checkedItemIndexPasskeys !== null) {
      setCheckedItemIndexAddEmail(null);
      setCheckedItemIndexEmails(null);
      setCheckedItemIndexSetPassword(null);
      setCheckedItemIndexAddPasskey(null);
    }
  }, [checkedItemIndexPasskeys]);

  useEffect(() => {
    if (checkedItemIndexAddPasskey !== null) {
      setCheckedItemIndexAddEmail(null);
      setCheckedItemIndexEmails(null);
      setCheckedItemIndexSetPassword(null);
      setCheckedItemIndexPasskeys(null);
    }
  }, [checkedItemIndexAddPasskey]);

  useEffect(() => {
    if (emailError !== null) {
      setPasswordError(null);
      setPasskeyError(null);
    }
  }, [emailError]);

  useEffect(() => {
    if (passwordError !== null) {
      setEmailError(null);
      setPasskeyError(null);
    }
  }, [passwordError]);

  useEffect(() => {
    if (passkeyError !== null) {
      setEmailError(null);
      setPasswordError(null);
    }
  }, [passkeyError]);

  return (
    <Content>
      <Headline1>{t("headlines.profileEmails")}</Headline1>
      <ErrorMessage error={emailError} />
      <Paragraph>{t("texts.manageEmails")}</Paragraph>
      <Paragraph>
        <ListEmailsAccordion
          setError={setEmailError}
          checkedItemIndex={checkedItemIndexEmails}
          setCheckedItemIndex={setCheckedItemIndexEmails}
        />
        {emails.length < config.emails.max_num_of_addresses ? (
          <AddEmailDropdown
            setError={setEmailError}
            checkedItemIndex={checkedItemIndexAddEmail}
            setCheckedItemIndex={setCheckedItemIndexAddEmail}
          />
        ) : null}
      </Paragraph>
      {config.password.enabled ? (
        <Fragment>
          <Headline1>{t("headlines.profilePassword")}</Headline1>
          <ErrorMessage error={passwordError} />
          <Paragraph>{t("texts.changePassword")}</Paragraph>
          <Paragraph>
            <ChangePasswordDropdown
              setError={setPasswordError}
              checkedItemIndex={checkedItemIndexSetPassword}
              setCheckedItemIndex={setCheckedItemIndexSetPassword}
            />
          </Paragraph>
        </Fragment>
      ) : null}
      <Headline1>{t("headlines.profilePasskeys")}</Headline1>
      <ErrorMessage error={passkeyError} />
      <Paragraph>{t("texts.managePasskeys")}</Paragraph>
      <Paragraph>
        <ListPasskeysAccordion
          credentials={webauthnCredentials}
          setError={setPasskeyError}
          checkedItemIndex={checkedItemIndexPasskeys}
          setCheckedItemIndex={setCheckedItemIndexPasskeys}
        />
        <AddPasskeyDropdown
          setError={setPasskeyError}
          checkedItemIndex={checkedItemIndexAddPasskey}
          setCheckedItemIndex={setCheckedItemIndexAddPasskey}
        />
      </Paragraph>
      {config.account.allow_deletion ? (
        <Fragment>
          <Spacer />
          <Paragraph>
            <Divider />
          </Paragraph>
          <Paragraph>
            <Form onSubmit={deleteUser}>
              <Button dangerous>{t("headlines.deleteAccount")}</Button>
            </Form>
          </Paragraph>
        </Fragment>
      ) : null}
    </Content>
  );
};

export default ProfilePage;
