import { h } from "preact";
import { Dispatch, SetStateAction, useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import Form from "../form/Form";
import Input from "../form/Input";
import Button from "../form/Button";
import Dropdown from "./Dropdown";
import ErrorMessage from "../error/ErrorMessage";
import Link from "../link/Link";
import { State } from "@teamhanko/hanko-frontend-sdk";

interface Props {
  checkedItemID?: string;
  setCheckedItemID: Dispatch<SetStateAction<string>>;
  flowState: State<"profile_init">;
  onState(state: State<any>): Promise<void>;
}

const ChangeUsernameDropdown = ({
  checkedItemID,
  setCheckedItemID,
  flowState,
  onState,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const [username, setUsername] = useState<string>();

  const onInputHandler = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      setUsername(event.target.value);
    }
  };

  const onSubmit = async (event: Event) => {
    event.preventDefault();
    const action = flowState.payload.user.username
      ? flowState.actions.username_update
      : flowState.actions.username_create;
    const nextState = await action.run(
      { username },
      { dispatchAfterStateChangeEvent: false },
    );

    return onState(nextState).then(() => setUsername(""));
  };

  const onDelete = async (event: Event) => {
    event.preventDefault();
    const nextState = await flowState.actions.username_delete.run(null, {
      dispatchAfterStateChangeEvent: false,
    });
    return onState(nextState).then(() => setUsername(""));
  };

  return (
    <Dropdown
      name={"username-edit-dropdown"}
      title={t(
        flowState.payload.user.username
          ? "labels.changeUsername"
          : "labels.setUsername",
      )}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    >
      <ErrorMessage
        flowError={
          flowState.payload.user.username
            ? flowState.actions.username_update.inputs.username?.error
            : flowState.actions.username_create.inputs.username?.error
        }
      />
      <Form
        flowAction={
          flowState.payload.user.username
            ? flowState.actions.username_update
            : flowState.actions.username_create
        }
        onSubmit={onSubmit}
      >
        <Input
          markError
          placeholder={t("labels.username")}
          type={"text"}
          onInput={onInputHandler}
          value={username}
          flowInput={
            flowState.payload.user.username
              ? flowState.actions.username_update.inputs.username
              : flowState.actions.username_create.inputs.username
          }
        />
        <Button>{t("labels.save")}</Button>
      </Form>
      <Link
        flowAction={flowState.actions.username_delete}
        onClick={onDelete}
        dangerous
        loadingSpinnerPosition={"right"}
      >
        {t("labels.delete")}
      </Link>
    </Dropdown>
  );
};

export default ChangeUsernameDropdown;
