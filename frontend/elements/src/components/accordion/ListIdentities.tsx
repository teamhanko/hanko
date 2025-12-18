import { Identity, State } from "@teamhanko/hanko-frontend-sdk";
import { Dispatch, SetStateAction, useContext, useMemo } from "preact/compat";
import { Fragment } from "preact";
import Accordion from "./Accordion";
import Paragraph from "../paragraph/Paragraph";
import Headline2 from "../headline/Headline2";
import Link from "../link/Link";
import { TranslateContext } from "@denysvuika/preact-translate";

interface Props {
  checkedItemID?: string;
  setCheckedItemID: Dispatch<SetStateAction<string>>;
  flowState: State<"profile_init">;
  onState(state: State<any>): Promise<void>;
}

const ListIdentities = ({
  checkedItemID,
  setCheckedItemID,
  flowState,
  onState,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const isDisabled = useMemo(() => false, []);

  const labels = (identity: Identity) => {
    const headline = <b>{identity.provider}</b>;
    return <>{headline}</>;
  };

  const onIdentityDelete = async (event: Event, identityID: string) => {
    event.preventDefault();
    const nextState =
      await flowState.actions.disconnect_thirdparty_oauth_provider.run(
        {
          identity_id: identityID,
        },
        { dispatchAfterStateChangeEvent: false },
      );
    return onState(nextState);
  };

  const contents = (identity: Identity) => (
    <>
      <>
        <Paragraph>
          <Headline2>{t("headlines.deleteIdentity")}</Headline2>
          <Link
            dangerous
            flowAction={flowState.actions.disconnect_thirdparty_oauth_provider}
            onClick={(event) => onIdentityDelete(event, identity.identity_id)}
            disabled={isDisabled}
            loadingSpinnerPosition={"right"}
          >
            {t("labels.delete")}
          </Link>
        </Paragraph>
      </>
    </>
  );

  return (
    <Accordion
      name={"connected-accounts"}
      columnSelector={labels}
      contentSelector={contents}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
      data={flowState.payload.user.identities}
    />
  );
};

export default ListIdentities;
