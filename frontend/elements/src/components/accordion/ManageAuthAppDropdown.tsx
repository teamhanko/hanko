import { h } from "preact";
import { Dispatch, Fragment, SetStateAction, useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import Paragraph from "../paragraph/Paragraph";
import Dropdown from "./Dropdown";
import Link from "../link/Link";
import Headline2 from "../headline/Headline2";
import styles from "./styles.sass";
import { State } from "@teamhanko/hanko-frontend-sdk";

interface Props {
  checkedItemID?: string;
  setCheckedItemID: Dispatch<SetStateAction<string>>;
  flowState: State<"profile_init">;
  onState(state: State<any>): Promise<void>;
}

const ManageAuthAppDropdown = ({
  checkedItemID,
  setCheckedItemID,
  flowState,
  onState,
}: Props) => {
  const { t } = useContext(TranslateContext);

  const onAuthAppSetUp = async (event: Event) => {
    event.preventDefault();
    const nextState =
      await flowState.actions.continue_to_otp_secret_creation.run(null, {
        dispatchAfterStateChangeEvent: false,
      });
    return onState(nextState);
  };

  const onAuthAppRemove = async (event: Event) => {
    event.preventDefault();
    const nextState = await flowState.actions.otp_secret_delete.run(null, {
      dispatchAfterStateChangeEvent: false,
    });
    return onState(nextState);
  };

  const configuredLabel = (
    <span className={styles.description}>
      {flowState.payload.user.mfa_config?.auth_app_set_up ? (
        <>
          {" -"} {t("labels.configured")}
        </>
      ) : null}
    </span>
  );

  const title = (
    <>
      {t("labels.authenticatorAppManage")} {configuredLabel}
    </>
  );

  return (
    <Dropdown
      name={"authenticator-app-manage-dropdown"}
      title={title}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    >
      <Headline2>
        {t(
          flowState.payload.user.mfa_config?.auth_app_set_up
            ? "headlines.authenticatorAppAlreadySetUp"
            : "headlines.authenticatorAppNotSetUp",
        )}
      </Headline2>
      <Paragraph>
        {t(
          flowState.payload.user.mfa_config?.auth_app_set_up
            ? "texts.authenticatorAppAlreadySetUp"
            : "texts.authenticatorAppNotSetUp",
        )}
        <br />
        {flowState.payload.user.mfa_config?.auth_app_set_up ? (
          <Link
            flowAction={flowState.actions.otp_secret_delete}
            onClick={onAuthAppRemove}
            loadingSpinnerPosition={"right"}
            dangerous
          >
            {t("labels.delete")}
          </Link>
        ) : (
          <Link
            flowAction={flowState.actions.continue_to_otp_secret_creation}
            onClick={onAuthAppSetUp}
            loadingSpinnerPosition={"right"}
          >
            {t("labels.authenticatorAppAdd")}
          </Link>
        )}
      </Paragraph>
    </Dropdown>
  );
};

export default ManageAuthAppDropdown;
