import { h } from "preact";
import { StateUpdater, useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";
import { UsernameSetInputs } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/input";

import Form from "../form/Form";
import Input from "../form/Input";
import Button from "../form/Button";
import Dropdown from "./Dropdown";
import ErrorMessage from "../error/ErrorMessage";
import Link from "../link/Link";

interface Props {
  inputs: UsernameSetInputs;
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  onUsernameSubmit: (event: Event, username: string) => Promise<void>;
  onUsernameDelete: (event: Event) => Promise<void>;
  hasUsername?: boolean;
  allowUsernameDeletion?: boolean;
}

const ChangeUsernameDropdown = ({
  inputs,
  checkedItemID,
  setCheckedItemID,
  onUsernameSubmit,
  onUsernameDelete,
  hasUsername,
  allowUsernameDeletion,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const [username, setUsername] = useState<string>();

  const onInputHandler = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      setUsername(event.target.value);
    }
  };

  return (
    <Dropdown
      name={"username-edit-dropdown"}
      title={t(hasUsername ? "labels.changeUsername" : "labels.setUsername")}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    >
      <ErrorMessage flowError={inputs.username?.error} />
      <Form
        onSubmit={(event: Event) =>
          onUsernameSubmit(event, username).then(() => setUsername(""))
        }
      >
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
      <Link
        hidden={!allowUsernameDeletion}
        uiAction={"username-delete"}
        dangerous
        onClick={(event: Event) =>
          onUsernameDelete(event).then(() => setUsername(""))
        }
        loadingSpinnerPosition={"right"}
      >
        {t("labels.delete")}
      </Link>
    </Dropdown>
  );
};

export default ChangeUsernameDropdown;
