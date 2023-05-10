import {
  StateUpdater,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "preact/compat";

import {
  Email,
  HankoError,
  TooManyRequestsError,
} from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Form from "../form/Form";
import Input from "../form/Input";
import Button from "../form/Button";
import Dropdown from "./Dropdown";

import LoginPasscodePage from "../../pages/LoginPasscodePage";
import ProfilePage from "../../pages/ProfilePage";

interface Props {
  setError: (e: HankoError) => void;
  checkedItemIndex?: number;
  setCheckedItemIndex: StateUpdater<number>;
}

const AddEmailDropdown = ({
  setError,
  checkedItemIndex,
  setCheckedItemIndex,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, config, user, setEmails, setPage, setPasscode } =
    useContext(AppContext);

  const [isSuccess, setIsSuccess] = useState<boolean>();
  const [isLoading, setIsLoading] = useState<boolean>();
  const [newEmail, setNewEmail] = useState<string>();

  const addEmail = (event: Event) => {
    event.preventDefault();
    return config.emails.require_verification
      ? addEmailWithVerification()
      : addEmailWithoutVerification();
  };

  const renderPasscode = useCallback(
    (email: Email) => {
      const onSuccessHandler = () => {
        return hanko.email
          .list()
          .then(setEmails)
          .then(() => setPage(<ProfilePage />));
      };

      const showPasscodePage = (e?: HankoError) =>
        setPage(
          <LoginPasscodePage
            userID={user.id}
            emailID={email.id}
            emailAddress={newEmail}
            initialError={e}
            onSuccess={onSuccessHandler}
            onBack={() => setPage(<ProfilePage />)}
          />
        );

      return hanko.passcode
        .initialize(user.id, email.id, true)
        .then(setPasscode)
        .then(() => showPasscodePage())
        .catch((e) => {
          if (e instanceof TooManyRequestsError) {
            showPasscodePage(e);
            return;
          }
          throw e;
        });
    },
    [hanko, newEmail, setEmails, setPage, setPasscode, user.id]
  );

  const addEmailWithVerification = () => {
    setIsLoading(true);
    hanko.email
      .create(newEmail)
      .then(renderPasscode)
      .finally(() => setIsLoading(false))
      .catch(setError);
  };

  const addEmailWithoutVerification = () => {
    setIsLoading(true);
    hanko.email
      .create(newEmail)
      .then(() => hanko.email.list())
      .then(setEmails)
      .then(() => {
        setError(null);
        setNewEmail("");
        setIsSuccess(true);
        setTimeout(() => {
          setCheckedItemIndex(null);
          setTimeout(() => {
            setIsSuccess(false);
          }, 500);
        }, 1000);
        return;
      })
      .finally(() => {
        setIsLoading(false);
      })
      .catch(setError);
  };

  const onInputHandler = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      setNewEmail(event.target.value);
    }
  };

  const disabled = useMemo(
    () => isSuccess || isLoading,
    [isLoading, isSuccess]
  );

  return (
    <Dropdown
      name={"add-email-dropdown"}
      title={t("labels.addEmail")}
      checkedItemIndex={checkedItemIndex}
      setCheckedItemIndex={setCheckedItemIndex}
    >
      <Form onSubmit={addEmail}>
        <Input
          type={"email"}
          placeholder={t("labels.newEmailAddress")}
          onInput={onInputHandler}
          value={newEmail}
          disabled={disabled}
          required
        />
        <Button disabled={disabled} isLoading={isLoading} isSuccess={isSuccess}>
          {t(
            config.emails.require_verification
              ? "labels.continue"
              : "labels.save"
          )}
        </Button>
      </Form>
    </Dropdown>
  );
};

export default AddEmailDropdown;
