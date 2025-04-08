import { Session } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/payload";
import { State } from "@teamhanko/hanko-frontend-sdk";
import { StateUpdater, useContext } from "preact/compat";
import Accordion from "./Accordion";
import { Fragment } from "preact";
import Paragraph from "../paragraph/Paragraph";
import Headline2 from "../headline/Headline2";
import { TranslateContext } from "@denysvuika/preact-translate";
import Link from "../link/Link";
import styles from "./styles.sass";

interface Props {
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  flowState: State<"profile_init">;
  onState(state: State<any>): Promise<void>;
}

const ListSessionsAccordion = ({
  checkedItemID,
  setCheckedItemID,
  flowState,
  onState,
}: Props) => {
  const { t } = useContext(TranslateContext);

  const onSessionDelete = async (event: Event, id: string) => {
    event.preventDefault();
    const nextState = await flowState.actions.session_delete.run(
      {
        session_id: id,
      },
      { dispatchAfterStateChangeEvent: false },
    );
    return onState(nextState);
  };

  const labels = (session: Session) => {
    const headline = (
      <b>{session.user_agent ? session.user_agent : session.id}</b>
    );
    const description = session.current ? (
      <span className={styles.description}>
        <Fragment>
          {" -"} {t("labels.currentSession")}
        </Fragment>
      </span>
    ) : null;
    return (
      <Fragment>
        {headline}
        {description}
      </Fragment>
    );
  };

  const convertTime = (t: string) => new Date(t).toLocaleString();

  const contents = (session: Session) => (
    <Fragment>
      <Paragraph hidden={!session.ip_address}>
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
      {flowState.actions.session_delete.inputs.session_id?.allowed_values
        ?.map((e) => e.value)
        ?.includes(session.id) ? (
        <Paragraph>
          <Headline2>{t("headlines.revokeSession")}</Headline2>
          <Link
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
      data={flowState.payload.sessions}
      contentSelector={contents}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    />
  );
};

export default ListSessionsAccordion;
