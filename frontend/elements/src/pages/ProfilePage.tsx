import { Fragment } from "preact";
import { useContext, useState } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";

import { useFlowState } from "../contexts/FlowState";
import { AppContext, UIAction } from "../contexts/AppProvider";

import Content from "../components/wrapper/Content";
import Headline1 from "../components/headline/Headline1";
import Paragraph from "../components/paragraph/Paragraph";
import ListEmailsAccordion from "../components/accordion/ListEmailsAccordion";
import ListWebauthnCredentialsAccordion from "../components/accordion/ListWebauthnCredentialsAccordion";
import AddEmailDropdown from "../components/accordion/AddEmailDropdown";
import ChangePasswordDropdown from "../components/accordion/ChangePasswordDropdown";
import AddWebauthnCredentialDropdown from "../components/accordion/AddWebauthnCredentialDropdown";
import Divider from "../components/spacer/Divider";
import Button from "../components/form/Button";
import Form from "../components/form/Form";
import Spacer from "../components/spacer/Spacer";
import ChangeUsernameDropdown from "../components/accordion/ChangeUsernameDropdown";
import DeleteAccountPage from "./DeleteAccountPage";
import ErrorBox from "../components/error/ErrorBox";
import ManageAuthAppDropdown from "../components/accordion/ManageAuthAppDropdown";

interface Props {
  state: State<"profile_init">;
  enablePasskeys?: boolean;
}

const ProfilePage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, setLoadingAction, stateHandler, setUIState, setPage } =
    useContext(AppContext);
  const { flowState } = useFlowState(props.state);

  const [checkedItemID, setCheckedItemID] = useState<string>("");

  const animationFinished = () => {
    return new Promise((resolve) => setTimeout(resolve, 360));
  };

  const onAction = async (
    event: Event,
    uiAction: UIAction,
    func: () => Promise<State<any>>,
  ) => {
    event.preventDefault();

    setLoadingAction(uiAction);
    const newState = await func();

    if (!newState?.error) {
      setCheckedItemID(null);
      await animationFinished();
    }

    setLoadingAction(null);
    await hanko.flow.run(newState, stateHandler);
  };

  const onEmailSubmit = async (event: Event, email: string) => {
    setUIState((prev) => ({ ...prev, email }));
    return onAction(
      event,
      "email-submit",
      flowState.actions.email_create({ email }).run,
    );
  };

  const onEmailDelete = async (event: Event, emailID: string) =>
    onAction(
      event,
      "email-delete",
      flowState.actions.email_delete({ email_id: emailID }).run,
    );

  const onEmailSetPrimary = async (event: Event, emailID: string) =>
    onAction(
      event,
      "email-set-primary",
      flowState.actions.email_set_primary({ email_id: emailID }).run,
    );

  const onEmailVerify = async (event: Event, emailID: string) =>
    onAction(
      event,
      "email-verify",
      flowState.actions.email_verify({ email_id: emailID }).run,
    );

  const onPasswordCreate = async (event: Event, password: string) =>
    onAction(
      event,
      "password-submit",
      flowState.actions.password_create({ password }).run,
    );

  const onPasswordUpdate = async (event: Event, password: string) =>
    onAction(
      event,
      "password-submit",
      flowState.actions.password_update({ password }).run,
    );

  const onPasswordDelete = async (event: Event) =>
    onAction(
      event,
      "password-delete",
      flowState.actions.password_delete(null).run,
    );

  const onUsernameCreate = async (event: Event, username: string) =>
    onAction(
      event,
      "username-set",
      flowState.actions.username_create({ username }).run,
    );

  const onUsernameUpdate = async (event: Event, username: string) =>
    onAction(
      event,
      "username-set",
      flowState.actions.username_update({ username }).run,
    );

  const onUsernameDelete = async (event: Event) =>
    onAction(
      event,
      "username-delete",
      flowState.actions.username_delete(null).run,
    );

  const onWebauthnCredentialNameSubmit = async (
    event: Event,
    id: string,
    name: string,
  ) =>
    onAction(
      event,
      "webauthn-credential-rename",
      flowState.actions.webauthn_credential_rename({
        passkey_id: id,
        passkey_name: name,
      }).run,
    );

  const onPasskeyDelete = async (event: Event, id: string) =>
    onAction(
      event,
      "passkey-delete",
      flowState.actions.webauthn_credential_delete({ passkey_id: id }).run,
    );

  const onPasskeySubmit = async (event: Event) =>
    onAction(
      event,
      "passkey-submit",
      flowState.actions.webauthn_credential_create(null).run,
    );

  const onSecurityKeyDelete = async (event: Event, id: string) =>
    onAction(
      event,
      "security-key-delete",
      flowState.actions.security_key_delete({ security_key_id: id }).run,
    );

  const onSecurityKeySubmit = async (event: Event) =>
    onAction(
      event,
      "security-key-submit",
      flowState.actions.security_key_create(null).run,
    );

  const onAccountDelete = async (event: Event) =>
    onAction(
      event,
      "account_delete",
      flowState.actions.account_delete(null).run,
    );

  const onBack = (event: Event) => {
    event.preventDefault();
    setPage(
      <ProfilePage state={flowState} enablePasskeys={props.enablePasskeys} />,
    );
    return Promise.resolve();
  };

  const onUserDelete = (event: Event) => {
    event.preventDefault();
    setPage(
      <DeleteAccountPage onBack={onBack} onAccountDelete={onAccountDelete} />,
    );
    return Promise.resolve();
  };

  const onAuthAppSetUp = async (event: Event) =>
    onAction(
      event,
      "auth-app-add",
      flowState.actions.continue_to_otp_secret_creation(null).run,
    );

  const onAuthAppRemove = async (event: Event) =>
    onAction(
      event,
      "auth-app-remove",
      flowState.actions.otp_secret_delete(null).run,
    );

  return (
    <Content>
      <ErrorBox
        state={
          flowState?.error?.code !== "form_data_invalid_error"
            ? flowState
            : null
        }
      />
      {flowState.actions.username_create?.(null) ||
      flowState.actions.username_update?.(null) ||
      flowState.actions.username_delete?.(null) ? (
        <Fragment>
          <Headline1>{t("labels.username")}</Headline1>
          {flowState.payload.user.username ? (
            <Paragraph>
              <b>{flowState.payload.user.username.username}</b>
            </Paragraph>
          ) : null}
          <Paragraph>
            {flowState.actions.username_create?.(null) ? (
              <ChangeUsernameDropdown
                inputs={flowState.actions.username_create(null).inputs}
                hasUsername={!!flowState.payload.user.username}
                allowUsernameDeletion={
                  !!flowState.actions.username_delete?.(null)
                }
                onUsernameSubmit={onUsernameCreate}
                onUsernameDelete={onUsernameDelete}
                checkedItemID={checkedItemID}
                setCheckedItemID={setCheckedItemID}
              />
            ) : null}
            {flowState.actions.username_update?.(null) ? (
              <ChangeUsernameDropdown
                inputs={flowState.actions.username_update(null).inputs}
                hasUsername={!!flowState.payload.user.username}
                allowUsernameDeletion={
                  !!flowState.actions.username_delete?.(null)
                }
                onUsernameSubmit={onUsernameUpdate}
                onUsernameDelete={onUsernameDelete}
                checkedItemID={checkedItemID}
                setCheckedItemID={setCheckedItemID}
              />
            ) : null}
          </Paragraph>
        </Fragment>
      ) : null}
      {flowState.payload?.user?.emails ||
      flowState.actions.email_create?.(null) ? (
        <Fragment>
          <Headline1>{t("headlines.profileEmails")}</Headline1>
          <Paragraph>
            <ListEmailsAccordion
              emails={flowState.payload.user.emails}
              onEmailDelete={onEmailDelete}
              onEmailSetPrimary={onEmailSetPrimary}
              onEmailVerify={onEmailVerify}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
              deletableEmailIDs={flowState.actions
                .email_delete?.(null)
                .inputs.email_id.allowed_values?.map((e) => e.value)}
            />
            {flowState.actions.email_create?.(null) ? (
              <AddEmailDropdown
                inputs={flowState.actions.email_create(null).inputs}
                onEmailSubmit={onEmailSubmit}
                checkedItemID={checkedItemID}
                setCheckedItemID={setCheckedItemID}
              />
            ) : null}
          </Paragraph>
        </Fragment>
      ) : null}
      {flowState.actions.password_create?.(null) ? (
        <Fragment>
          <Headline1>{t("headlines.profilePassword")}</Headline1>
          <Paragraph>
            <ChangePasswordDropdown
              inputs={flowState.actions.password_create(null).inputs}
              onPasswordSubmit={onPasswordCreate}
              onPasswordDelete={onPasswordDelete}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
            />
          </Paragraph>
        </Fragment>
      ) : null}
      {flowState.actions.password_update?.(null) ? (
        <Fragment>
          <Headline1>{t("headlines.profilePassword")}</Headline1>
          <Paragraph>
            <ChangePasswordDropdown
              allowPasswordDelete={!!flowState.actions.password_delete?.(null)}
              inputs={flowState.actions.password_update(null).inputs}
              onPasswordSubmit={onPasswordUpdate}
              onPasswordDelete={onPasswordDelete}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
              passwordExists
            />
          </Paragraph>
        </Fragment>
      ) : null}
      {props.enablePasskeys &&
      (flowState.payload?.user?.passkeys ||
        flowState.actions.webauthn_credential_create?.(null)) ? (
        <Fragment>
          <Headline1>{t("headlines.profilePasskeys")}</Headline1>
          <Paragraph>
            <ListWebauthnCredentialsAccordion
              onBack={onBack}
              onCredentialNameSubmit={onWebauthnCredentialNameSubmit}
              onCredentialDelete={onPasskeyDelete}
              credentials={flowState.payload.user.passkeys}
              setError={null}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
              allowCredentialDeletion={
                !!flowState.actions.webauthn_credential_delete?.(null)
              }
              credentialType={"passkey"}
            />
            {flowState.actions.webauthn_credential_create?.(null) ? (
              <AddWebauthnCredentialDropdown
                credentialType={"passkey"}
                onCredentialSubmit={onPasskeySubmit}
                setError={null}
                checkedItemID={checkedItemID}
                setCheckedItemID={setCheckedItemID}
              />
            ) : null}
          </Paragraph>
        </Fragment>
      ) : null}
      {flowState.payload.user.mfa_config?.security_keys_enabled ? (
        <Fragment>
          <Headline1>{t("headlines.securityKeys")}</Headline1>
          <Paragraph>
            <ListWebauthnCredentialsAccordion
              onBack={onBack}
              onCredentialNameSubmit={onWebauthnCredentialNameSubmit}
              onCredentialDelete={onSecurityKeyDelete}
              credentials={flowState.payload.user.security_keys}
              setError={null}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
              allowCredentialDeletion={
                !!flowState.actions.security_key_delete?.(null)
              }
              credentialType={"security-key"}
            />
            {flowState.actions.security_key_create?.(null) ? (
              <AddWebauthnCredentialDropdown
                credentialType={"security-key"}
                onCredentialSubmit={onSecurityKeySubmit}
                setError={null}
                checkedItemID={checkedItemID}
                setCheckedItemID={setCheckedItemID}
              />
            ) : null}
          </Paragraph>
        </Fragment>
      ) : null}
      {flowState.payload.user.mfa_config?.totp_enabled ? (
        <Fragment>
          <Headline1>{t("headlines.authenticatorApp")}</Headline1>
          <Paragraph>
            <ManageAuthAppDropdown
              onConnect={onAuthAppSetUp}
              onDelete={onAuthAppRemove}
              authAppSetUp={flowState.payload.user.mfa_config?.auth_app_set_up}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
            />
          </Paragraph>
        </Fragment>
      ) : null}
      {flowState.actions.account_delete?.(null) ? (
        <Fragment>
          <Spacer />
          <Paragraph>
            <Divider />
          </Paragraph>
          <Paragraph>
            <Form onSubmit={onUserDelete}>
              <Button dangerous>{t("headlines.deleteAccount")}</Button>
            </Form>
          </Paragraph>
        </Fragment>
      ) : null}
    </Content>
  );
};

export default ProfilePage;
