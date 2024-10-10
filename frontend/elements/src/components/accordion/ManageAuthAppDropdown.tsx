import { h } from "preact";
import { Fragment, StateUpdater, useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import Paragraph from "../paragraph/Paragraph";
import Dropdown from "./Dropdown";
import Link from "../link/Link";
import Headline2 from "../headline/Headline2";
import styles from "./styles.sass";

interface Props {
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  onDelete: (event: Event) => Promise<void>;
  onConnect: (event: Event) => Promise<void>;
  authAppSetUp: boolean;
}

const ManageAuthAppDropdown = ({
  checkedItemID,
  setCheckedItemID,
  onDelete,
  onConnect,
  authAppSetUp,
}: Props) => {
  const { t } = useContext(TranslateContext);

  const configuredLabel = (
    <span className={styles.description}>
      {authAppSetUp ? (
        <Fragment>
          {" -"} {t("labels.configured")}
        </Fragment>
      ) : null}
    </span>
  );

  const title = (
    <Fragment>
      {t("labels.authenticatorAppManage")} {configuredLabel}
    </Fragment>
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
          authAppSetUp
            ? "headlines.authenticatorAppAlreadySetUp"
            : "headlines.authenticatorAppNotSetUp",
        )}
      </Headline2>
      <Paragraph>
        {t(
          authAppSetUp
            ? "texts.authenticatorAppAlreadySetUp"
            : "texts.authenticatorAppNotSetUp",
        )}
        <br />
        {authAppSetUp ? (
          <Link
            uiAction={"auth-app-remove"}
            onClick={(event: Event) => onDelete(event)}
            loadingSpinnerPosition={"right"}
            dangerous
          >
            {t("labels.delete")}
          </Link>
        ) : (
          <Link
            uiAction={"auth-app-add"}
            onClick={(event: Event) => onConnect(event)}
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
