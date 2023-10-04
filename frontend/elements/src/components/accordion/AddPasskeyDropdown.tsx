import { StateUpdater, useContext, useState } from "preact/compat";

import {
  WebauthnSupport,
  HankoError,
  WebauthnRequestCancelledError,
} from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Form from "../form/Form";
import Button from "../form/Button";
import Paragraph from "../paragraph/Paragraph";
import Dropdown from "./Dropdown";

interface Props {
  setError: (e: HankoError) => void;
  checkedItemIndex?: number;
  setCheckedItemIndex: StateUpdater<number>;
}

const AddPasskeyDropdown = ({
  setError,
  checkedItemIndex,
  setCheckedItemIndex,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { hanko, setWebauthnCredentials } = useContext(AppContext);

  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);

  const webauthnSupported = WebauthnSupport.supported();

  const addPasskey = (event: Event) => {
    event.preventDefault();
    setIsLoading(true);
    hanko.webauthn
      .register()
      .then(() => hanko.webauthn.listCredentials())
      .then(setWebauthnCredentials)
      .then(() => {
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
      .catch((e) => {
        if (!(e instanceof WebauthnRequestCancelledError)) {
          setError(e);
        }
      });
  };

  return (
    <Dropdown
      name={"add-passkey-dropdown"}
      title={t("labels.createPasskey")}
      checkedItemIndex={checkedItemIndex}
      setCheckedItemIndex={setCheckedItemIndex}
    >
      <Paragraph>{t("texts.setupPasskey")}</Paragraph>
      <Form onSubmit={addPasskey}>
        <Button
          title={!webauthnSupported ? t("labels.webauthnUnsupported") : null}
          isLoading={isLoading}
          isSuccess={isSuccess}
          disabled={!webauthnSupported || isLoading}
        >
          {t("labels.registerAuthenticator")}
        </Button>
      </Form>
    </Dropdown>
  );
};

export default AddPasskeyDropdown;
