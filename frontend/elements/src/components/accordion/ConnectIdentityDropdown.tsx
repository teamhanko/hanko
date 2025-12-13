import { Dispatch, SetStateAction, useContext } from "preact/compat";
import { State } from "@teamhanko/hanko-frontend-sdk";
import { TranslateContext } from "@denysvuika/preact-translate";
import Dropdown from "./Dropdown";
import ErrorMessage from "../error/ErrorMessage";
import Button from "../form/Button";
import Form from "../form/Form";

interface Props {
  checkedItemID?: string;
  setCheckedItemID: Dispatch<SetStateAction<string>>;
  flowState: State<"profile_init">;

  onState(state: State<any>): Promise<void>;
}

const ConnectIdentityDropdown = ({
  checkedItemID,
  setCheckedItemID,
  flowState,
  onState,
}: Props) => {
  const { t } = useContext(TranslateContext);

  const onSubmit = async (event: Event, provider: string) => {
    event.preventDefault();
    const nextState =
      await flowState.actions.connect_thirdparty_oauth_provider.run({
        provider,
        redirect_to: window.location.href,
      });

    return onState(nextState);
  };

  return (
    <Dropdown
      name={"connect-account-dropdown"}
      title={t("labels.connectAccount")}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    >
      <ErrorMessage
        flowError={
          flowState.actions.connect_thirdparty_oauth_provider.inputs.provider
            ?.error
        }
      />

      {flowState.actions.connect_thirdparty_oauth_provider.inputs.provider.allowed_values?.map(
        (provider) => {
          return (
            <Form
              key={provider.value}
              flowAction={flowState.actions.connect_thirdparty_oauth_provider}
              onSubmit={(event) => onSubmit(event, provider.value)}
            >
              <Button
                key={provider}
                // @ts-ignore
                icon={
                  provider.value.startsWith("custom_")
                    ? "customProvider"
                    : provider.value
                }
              >
                {provider.name}
              </Button>
            </Form>
          );
        },
      )}
    </Dropdown>
  );
};

export default ConnectIdentityDropdown;
