import { h } from "preact";
import { StateUpdater, useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";
import { UsernameSetInputs } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/input";

import Form from "../form/Form";
import Input from "../form/Input";
import Button from "../form/Button";
import Dropdown from "./Dropdown";
import ErrorMessage from "../error/ErrorMessage";

interface Props {
  inputs: UsernameSetInputs;
  prefilledUsername?: string;
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  onUsernameSubmit: (event: Event, username: string) => Promise<void>;
}

const ChangeUsernameDropdown = ({
  inputs,
  checkedItemID,
  setCheckedItemID,
  onUsernameSubmit,
  prefilledUsername,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const [username, setUsername] = useState<string>(prefilledUsername);

  const onInputHandler = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      setUsername(event.target.value);
    }
  };

  return (
    <Dropdown
      name={"username-edit-dropdown"}
      title={t("labels.changeUsername")}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    >
      <ErrorMessage flowError={inputs.username?.error} />
      <Form onSubmit={(event: Event) => onUsernameSubmit(event, username)}>
        <Input
          markError
          placeholder={t("labels.username")}
          type={"text"}
          onInput={onInputHandler}
          value={username}
          flowInput={inputs.username}
        />
        <Button uiAction={"username-set"}>{t("labels.save")}</Button>
      </Form>
    </Dropdown>
  );
};

export default ChangeUsernameDropdown;
