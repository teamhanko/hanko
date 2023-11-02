import { SetStateAction, useContext, useState } from "preact/compat";

import { HankoError } from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Form from "../form/Form";
import Input from "../form/Input";
import Button from "../form/Button";
import Paragraph from "../paragraph/Paragraph";
import Dropdown from "./Dropdown";

interface Props {
  setError: (e: HankoError) => void;
  checkedItemIndex?: number;
  setCheckedItemIndex: SetStateAction<number>;
}

const ChangePasswordDropdown = ({
  setError,
  checkedItemIndex,
  setCheckedItemIndex,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, config, user } = useContext(AppContext);

  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);
  const [newPassword, setNewPassword] = useState<string>("");

  const changePassword = (event: Event) => {
    event.preventDefault();
    setIsLoading(true);
    hanko.password
      .update(user.id, newPassword)
      .then(() => {
        setNewPassword("");
        setError(null);
        setIsSuccess(true);
        setTimeout(() => {
          setCheckedItemIndex(null);
          setTimeout(() => {
            setIsSuccess(false);
          }, 500);
        }, 1000);
        return;
      })
      .finally(() => setIsLoading(false))
      .catch(setError);
  };

  const onInputHandler = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      setNewPassword(event.target.value);
    }
  };

  return (
    <Dropdown
      name={"change-password-dropdown"}
      title={t("labels.changePassword")}
      checkedItemIndex={checkedItemIndex}
      setCheckedItemIndex={setCheckedItemIndex}
    >
      <Paragraph>
        {t("texts.passwordFormatHint", {
          minLength: config.password.min_password_length,
          maxLength: 72,
        })}
      </Paragraph>
      <Form onSubmit={changePassword}>
        <Input
          placeholder={t("labels.newPassword")}
          type={"password"}
          onInput={onInputHandler}
          value={newPassword}
          minLength={config.password.min_password_length}
          maxLength={72}
          required
          disabled={isLoading || isSuccess}
        />
        <Button
          isLoading={isLoading}
          isSuccess={isSuccess}
          disabled={isLoading}
        >
          {t("labels.save")}
        </Button>
      </Form>
    </Dropdown>
  );
};

export default ChangePasswordDropdown;
