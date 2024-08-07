import { h } from "preact";
import { StateUpdater, useContext, useState } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { EmailCreateInputs } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/input";

import Form from "../form/Form";
import Input from "../form/Input";
import Button from "../form/Button";
import Dropdown from "./Dropdown";
import ErrorMessage from "../error/ErrorMessage";

interface Props {
  inputs: EmailCreateInputs;
  onEmailSubmit: (event: Event, email: string) => Promise<void>;
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
}

const AddEmailDropdown = ({
  inputs,
  onEmailSubmit,
  checkedItemID,
  setCheckedItemID,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const [newEmail, setNewEmail] = useState<string>();

  const onInputHandler = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      setNewEmail(event.target.value);
    }
  };

  return (
    <Dropdown
      name={"email-create-dropdown"}
      title={t("labels.addEmail")}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    >
      <ErrorMessage flowError={inputs.email?.error} />
      <Form
        onSubmit={(event: Event) =>
          onEmailSubmit(event, newEmail).then(() => setNewEmail(""))
        }
      >
        <Input
          markError
          type={"email"}
          placeholder={t("labels.newEmailAddress")}
          onInput={onInputHandler}
          value={newEmail}
          flowInput={inputs.email}
        />
        <Button uiAction={"email-submit"}>{t("labels.save")}</Button>
      </Form>
    </Dropdown>
  );
};

export default AddEmailDropdown;
