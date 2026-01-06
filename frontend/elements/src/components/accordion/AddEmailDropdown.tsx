import { h } from "preact";
import { Dispatch, SetStateAction, useContext, useState } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";

import Form from "../form/Form";
import Input from "../form/Input";
import Button from "../form/Button";
import Dropdown from "./Dropdown";
import ErrorMessage from "../error/ErrorMessage";
import { State } from "@teamhanko/hanko-frontend-sdk";
import { AppContext } from "../../contexts/AppProvider";

interface Props {
  checkedItemID?: string;
  setCheckedItemID: Dispatch<SetStateAction<string>>;
  flowState: State<"profile_init">;
  onState(state: State<any>): Promise<void>;
}

const AddEmailDropdown = ({
  checkedItemID,
  setCheckedItemID,
  flowState,
  onState,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { setUIState } = useContext(AppContext);
  const [newEmail, setNewEmail] = useState<string>();

  const onInputHandler = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      setNewEmail(event.target.value);
    }
  };

  const onEmailSubmit = async (event: Event, email: string) => {
    event.preventDefault();
    setUIState((prev) => ({ ...prev, email }));
    const nextState = await flowState.actions.email_create.run(
      { email },
      { dispatchAfterStateChangeEvent: false },
    );
    return onState(nextState);
  };

  return (
    <Dropdown
      name={"email-create-dropdown"}
      title={t("labels.addEmail")}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    >
      <ErrorMessage
        flowError={flowState.actions.email_create.inputs.email?.error}
      />
      <Form
        flowAction={flowState.actions.email_create}
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
          flowInput={flowState.actions.email_create.inputs.email}
        />
        <Button>{t("labels.save")}</Button>
      </Form>
    </Dropdown>
  );
};

export default AddEmailDropdown;
