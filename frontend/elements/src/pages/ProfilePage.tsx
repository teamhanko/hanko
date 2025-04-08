import { Fragment } from "preact";
import { useContext, useState } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";

import { useFlowState } from "../hooks/FlowState";
import { AppContext } from "../contexts/AppProvider";

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
import ListSessionsAccordion from "../components/accordion/ListSessionsAccordion";
import ManageAuthAppDropdown from "../components/accordion/ManageAuthAppDropdown";

interface Props {
  state: State<"profile_init">;
  enablePasskeys?: boolean;
}

const ProfilePage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { setPage } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);

  const [checkedItemID, setCheckedItemID] = useState<string>("");

  const animationFinished = () => {
    return new Promise((resolve) => setTimeout(resolve, 360));
  };

  const onState = async (newState: State<any>) => {
    if (!newState?.error) {
      setCheckedItemID(null);
      await animationFinished();
    }

    newState.dispatchAfterStateChangeEvent();
  };

  const onPasskeyDelete = async (event: Event, id: string) => {
    event.preventDefault();
    const nextState = await flowState.actions.webauthn_credential_delete.run(
      {
        passkey_id: id,
      },
      { dispatchAfterStateChangeEvent: false },
    );
    return onState(nextState);
  };

  const onWebauthnCredentialNameSubmit = async (
    event: Event,
    id: string,
    name: string,
  ) => {
    event.preventDefault();
    const nextState = await flowState.actions.webauthn_credential_rename.run(
      {
        passkey_id: id,
        passkey_name: name,
      },
      { dispatchAfterStateChangeEvent: false },
    );
    return onState(nextState);
  };

  const onPasskeySubmit = async (event: Event) => {
    event.preventDefault();
    const nextState = await flowState.actions.webauthn_credential_create.run(
      null,
      { dispatchAfterStateChangeEvent: false },
    );
    return onState(nextState);
  };

  const onSecurityKeyDelete = async (event: Event, id: string) => {
    event.preventDefault();
    const nextState = await flowState.actions.security_key_delete.run(
      {
        security_key_id: id,
      },
      { dispatchAfterStateChangeEvent: false },
    );
    return onState(nextState);
  };

  const onSecurityKeySubmit = async (event: Event) => {
    event.preventDefault();
    const nextState = await flowState.actions.security_key_create.run(null, {
      dispatchAfterStateChangeEvent: false,
    });
    return onState(nextState);
  };

  const onBack = (event: Event) => {
    event.preventDefault();
    setPage(
      <ProfilePage state={flowState} enablePasskeys={props.enablePasskeys} />,
    );
    return Promise.resolve();
  };

  const onUserDelete = (event: Event) => {
    event.preventDefault();
    setPage(<DeleteAccountPage onBack={onBack} state={flowState} />);
    return Promise.resolve();
  };

  return (
    <Content>
      <ErrorBox
        state={
          flowState?.error?.code !== "form_data_invalid_error"
            ? flowState
            : null
        }
      />
      {flowState.actions.username_create.enabled ||
      flowState.actions.username_update.enabled ||
      flowState.actions.username_delete.enabled ? (
        <Fragment>
          <Headline1>{t("labels.username")}</Headline1>
          {flowState.payload.user.username ? (
            <Paragraph>
              <b>{flowState.payload.user.username.username}</b>
            </Paragraph>
          ) : null}
          <Paragraph>
            {flowState.actions.username_create.enabled ||
            flowState.actions.username_update.enabled ? (
              <ChangeUsernameDropdown
                onState={onState}
                flowState={flowState}
                checkedItemID={checkedItemID}
                setCheckedItemID={setCheckedItemID}
              />
            ) : null}
          </Paragraph>
        </Fragment>
      ) : null}
      {flowState.payload?.user?.emails ||
      flowState.actions.email_create.enabled ? (
        <Fragment>
          <Headline1>{t("headlines.profileEmails")}</Headline1>
          <Paragraph>
            <ListEmailsAccordion
              flowState={flowState}
              onState={onState}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
            />
            {flowState.actions.email_create.enabled ? (
              <AddEmailDropdown
                flowState={flowState}
                onState={onState}
                checkedItemID={checkedItemID}
                setCheckedItemID={setCheckedItemID}
              />
            ) : null}
          </Paragraph>
        </Fragment>
      ) : null}
      {flowState.actions.password_create.enabled ||
      flowState.actions.password_update.enabled ? (
        <Fragment>
          <Headline1>{t("headlines.profilePassword")}</Headline1>
          <Paragraph>
            <ChangePasswordDropdown
              flowState={flowState}
              onState={onState}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
            />
          </Paragraph>
        </Fragment>
      ) : null}
      {props.enablePasskeys &&
      (flowState.payload?.user?.passkeys ||
        flowState.actions.webauthn_credential_create.enabled) ? (
        <Fragment>
          <Headline1>{t("headlines.profilePasskeys")}</Headline1>
          <Paragraph>
            <ListWebauthnCredentialsAccordion
              flowState={flowState}
              onBack={onBack}
              onCredentialNameSubmit={onWebauthnCredentialNameSubmit}
              onCredentialDelete={onPasskeyDelete}
              credentials={flowState.payload.user.passkeys}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
              allowCredentialDeletion={
                !!flowState.actions.webauthn_credential_delete.enabled
              }
              credentialType={"passkey"}
            />
            {flowState.actions.webauthn_credential_create.enabled ? (
              <AddWebauthnCredentialDropdown
                credentialType={"passkey"}
                onCredentialSubmit={onPasskeySubmit}
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
              flowState={flowState}
              onCredentialNameSubmit={onWebauthnCredentialNameSubmit}
              onCredentialDelete={onSecurityKeyDelete}
              credentials={flowState.payload.user.security_keys}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
              allowCredentialDeletion={
                !!flowState.actions.security_key_delete.enabled
              }
              credentialType={"security-key"}
            />
            {flowState.actions.security_key_create.enabled ? (
              <AddWebauthnCredentialDropdown
                credentialType={"security-key"}
                onCredentialSubmit={onSecurityKeySubmit}
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
              onState={onState}
              flowState={flowState}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
            />
          </Paragraph>
        </Fragment>
      ) : null}
      {flowState.payload.sessions ? (
        <Fragment>
          <Headline1>{t("headlines.profileSessions")}</Headline1>
          <Paragraph>
            <ListSessionsAccordion
              flowState={flowState}
              onState={onState}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
            />
          </Paragraph>
        </Fragment>
      ) : null}
      {flowState.actions.account_delete.enabled ? (
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
