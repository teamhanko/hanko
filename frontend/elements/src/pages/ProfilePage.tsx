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
import ListPasskeysAccordion from "../components/accordion/ListPasskeysAccordion";
import AddEmailDropdown from "../components/accordion/AddEmailDropdown";
import ChangePasswordDropdown from "../components/accordion/ChangePasswordDropdown";
import AddPasskeyDropdown from "../components/accordion/AddPasskeyDropdown";
import Divider from "../components/spacer/Divider";
import Button from "../components/form/Button";
import Form from "../components/form/Form";
import Spacer from "../components/spacer/Spacer";
import ChangeUsernameDropdown from "../components/accordion/ChangeUsernameDropdown";
import DeleteAccountPage from "./DeleteAccountPage";
import ErrorBox from "../components/error/ErrorBox";

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

  const onPasswordSubmit = async (event: Event, password: string) =>
    onAction(
      event,
      "password-submit",
      flowState.actions.password_set({ password }).run,
    );

  const onPasswordDelete = async (event: Event) =>
    onAction(
      event,
      "password-delete",
      flowState.actions.password_delete(null).run,
    );

  const onUsernameSubmit = async (event: Event, username: string) =>
    onAction(
      event,
      "username-set",
      flowState.actions.username_set({ username }).run,
    );

  const onPasskeyNameSubmit = async (event: Event, id: string, name: string) =>
    onAction(
      event,
      "passkey-rename",
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

  return (
    <Content>
      <ErrorBox
        state={
          flowState?.error?.code !== "form_data_invalid_error"
            ? flowState
            : null
        }
      />
      {flowState.actions.username_set?.(null) ? (
        <Fragment>
          <Headline1>{t("labels.username")}</Headline1>
          {flowState.payload.user.username ? (
            <Paragraph>
              <b>{flowState.payload.user.username}</b>
            </Paragraph>
          ) : null}
          <Paragraph>
            <ChangeUsernameDropdown
              inputs={flowState.actions.username_set(null).inputs}
              hasUsername={!!flowState.payload.user.username}
              onUsernameSubmit={onUsernameSubmit}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
            />
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
      {flowState.actions.password_set?.(null) ? (
        <Fragment>
          <Headline1>{t("headlines.profilePassword")}</Headline1>
          <Paragraph>
            <ChangePasswordDropdown
              hasPassword={!!flowState.actions.password_delete?.(null)}
              inputs={flowState.actions.password_set(null).inputs}
              onPasswordSubmit={onPasswordSubmit}
              onPasswordDelete={onPasswordDelete}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
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
            <ListPasskeysAccordion
              onBack={onBack}
              onPasskeyNameSubmit={onPasskeyNameSubmit}
              onPasskeyDelete={onPasskeyDelete}
              passkeys={flowState.payload.user.passkeys}
              setError={null}
              checkedItemID={checkedItemID}
              setCheckedItemID={setCheckedItemID}
            />
            {flowState.actions.webauthn_credential_create?.(null) ? (
              <AddPasskeyDropdown
                onPasskeySubmit={onPasskeySubmit}
                setError={null}
                checkedItemID={checkedItemID}
                setCheckedItemID={setCheckedItemID}
              />
            ) : null}
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
