import { Session } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/payload";
import { HankoError } from "@teamhanko/hanko-frontend-sdk";
import { StateUpdater, useContext } from "preact/compat";
import Accordion from "./Accordion";
import { Fragment } from "preact";
import Paragraph from "../paragraph/Paragraph";
import Headline2 from "../headline/Headline2";
import { TranslateContext } from "@denysvuika/preact-translate";
import Link from "../link/Link";
import styles from "./styles.sass";

interface Props {
  sessions: Session[];
  setError: (e: HankoError) => void;
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  onSessionDelete: (event: Event, id: string) => Promise<void>;
  deletableSessionIDs?: string[];
}

const ListSessionsAccordion = ({
  sessions = [],
  checkedItemID,
  setCheckedItemID,
  onSessionDelete,
  deletableSessionIDs,
}: Props) => {
  const { t } = useContext(TranslateContext);

  const labels = (session: Session) => {
    const description = (
      <span className={styles.description}>
        {session.current ? (
          <Fragment>
            {" -"} {t("labels.current")}
          </Fragment>
        ) : null}
      </span>
    );
    return session.current ? (
      <Fragment>
        <b>{session.user_agent}</b>
        {description}
      </Fragment>
    ) : (
      <Fragment>
        {session.user_agent}
        {description}
      </Fragment>
    );
  };

  const convertTime = (t: string) => new Date(t).toLocaleString();

  const contents = (session: Session) => (
    <Fragment>
      <Paragraph>
        <Headline2>{t("headlines.ipAddress")}</Headline2>
        {session.ip_address}
      </Paragraph>
      <Paragraph>
        <Headline2>{t("headlines.lastUsed")}</Headline2>
        {convertTime(session.last_used)}
      </Paragraph>
      <Paragraph>
        <Headline2>{t("headlines.createdAt")}</Headline2>
        {convertTime(session.created_at)}
      </Paragraph>
      {deletableSessionIDs?.includes(session.id) ? (
        <Paragraph>
          <Headline2>{t("headlines.revokeSession")}</Headline2>
          <Link
            uiAction={"session-delete"}
            dangerous
            onClick={(event) => onSessionDelete(event, session.id)}
            loadingSpinnerPosition={"right"}
          >
            {t("labels.revoke")}
          </Link>
        </Paragraph>
      ) : null}
    </Fragment>
  );

  return (
    <Accordion
      name={"session-edit-dropdown"}
      columnSelector={labels}
      data={sessions}
      contentSelector={contents}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    />
  );
};

export default ListSessionsAccordion;
