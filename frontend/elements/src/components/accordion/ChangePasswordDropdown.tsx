import { h } from "preact";
import { Dispatch, SetStateAction, useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import Form from "../form/Form";
import Input from "../form/Input";
import Button from "../form/Button";
import Paragraph from "../paragraph/Paragraph";
import Dropdown from "./Dropdown";
import Link from "../link/Link";
import ErrorMessage from "../error/ErrorMessage";
import { State } from "@teamhanko/hanko-frontend-sdk";

interface Props {
  checkedItemID?: string;
  setCheckedItemID: Dispatch<SetStateAction<string>>;
  flowState: State<"profile_init">;
  onState(state: State<any>): Promise<void>;
}

const ChangePasswordDropdown = ({
  checkedItemID,
  setCheckedItemID,
  onState,
  flowState,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const [newPassword, setNewPassword] = useState<string>("");

  const action = flowState.actions.password_create.enabled
    ? flowState.actions.password_create
    : flowState.actions.password_update;

  const onInputHandler = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      setNewPassword(event.target.value);
    }
  };

  const onPasswordSubmit = async (event: Event, password: string) => {
    event.preventDefault();
    const nextState = await action.run(
      { password },
      { dispatchAfterStateChangeEvent: false },
    );
    return onState(nextState);
  };

  const onPasswordDelete = async (event: Event) => {
    event.preventDefault();
    const nextState = await flowState.actions.password_delete.run(null, {
      dispatchAfterStateChangeEvent: false,
    });
    return onState(nextState);
  };

  return (
    <Dropdown
      name={"password-edit-dropdown"}
      title={t(
        flowState.actions.password_create.enabled
          ? "labels.setPassword"
          : "labels.changePassword",
      )}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    >
      <Paragraph>
        {t("texts.passwordFormatHint", {
          minLength: action.inputs.password.min_length?.toString(10),
          maxLength: action.inputs.password.max_length?.toString(10),
        })}
      </Paragraph>
      <ErrorMessage
        flowError={flowState.actions.password_create.inputs.password?.error}
      />
      <Form
        flowAction={action}
        onSubmit={(event: Event) =>
          onPasswordSubmit(event, newPassword).then(() => setNewPassword(""))
        }
      >
        <Input
          markError
          autoComplete={"new-password"}
          placeholder={t("labels.newPassword")}
          type={"password"}
          onInput={onInputHandler}
          value={newPassword}
          flowInput={action.inputs.password}
        />
        <Button>{t("labels.save")}</Button>
      </Form>
      <Link
        dangerous
        flowAction={flowState.actions.password_delete}
        onClick={(event: Event) =>
          onPasswordDelete(event).then(() => setNewPassword(""))
        }
        loadingSpinnerPosition={"right"}
      >
        {t("labels.delete")}
      </Link>
    </Dropdown>
  );
};

export default ChangePasswordDropdown;
