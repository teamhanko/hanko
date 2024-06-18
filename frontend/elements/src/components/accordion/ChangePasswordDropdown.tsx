import { h } from "preact";
import { StateUpdater, useContext, useState } from "preact/compat";
import { PasswordSetInputs } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/input";

import { TranslateContext } from "@denysvuika/preact-translate";

import Form from "../form/Form";
import Input from "../form/Input";
import Button from "../form/Button";
import Paragraph from "../paragraph/Paragraph";
import Dropdown from "./Dropdown";
import Link from "../link/Link";
import ErrorMessage from "../error/ErrorMessage";

interface Props {
  inputs: PasswordSetInputs;
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  onPasswordSubmit: (event: Event, password: string) => Promise<void>;
  onPasswordDelete: (event: Event) => Promise<void>;
  hasPassword?: boolean;
}

const ChangePasswordDropdown = ({
  inputs,
  checkedItemID,
  setCheckedItemID,
  onPasswordSubmit,
  onPasswordDelete,
  hasPassword,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const [newPassword, setNewPassword] = useState<string>("");

  const onInputHandler = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      setNewPassword(event.target.value);
    }
  };

  return (
    <Dropdown
      name={"password-edit-dropdown"}
      title={t(hasPassword ? "labels.changePassword" : "labels.setPassword")}
      checkedItemID={checkedItemID}
      setCheckedItemID={setCheckedItemID}
    >
      <Paragraph>
        {t("texts.passwordFormatHint", {
          minLength: inputs.password.min_length?.toString(10),
          maxLength: inputs.password.max_length?.toString(10),
        })}
      </Paragraph>
      <ErrorMessage flowError={inputs.password?.error} />
      <Form
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
          flowInput={inputs.password}
        />
        <Button uiAction={"password-submit"}>{t("labels.save")}</Button>
      </Form>
      <Link
        hidden={!hasPassword}
        uiAction={"password-delete"}
        dangerous
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
