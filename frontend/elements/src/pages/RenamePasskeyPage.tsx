import { Fragment } from "preact";
import { useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Input from "../components/form/Input";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";
import Footer from "../components/wrapper/Footer";
import Link from "../components/link/Link";
import { Passkey } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/payload";

type Props = {
  oldName: string;
  passkey: Passkey;
  onBack: (event: Event) => Promise<void>;
  onPasskeyNameSubmit: (
    event: Event,
    id: string,
    name: string,
  ) => Promise<void>;
};

const RenamePasskeyPage = ({
  onPasskeyNameSubmit,
  oldName,
  onBack,
  passkey,
}: Props) => {
  const { t } = useContext(TranslateContext);
  const [newName, setNewName] = useState<string>(oldName);

  const onInput = async (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setNewName(event.target.value);
    }
  };

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.renamePasskey")}</Headline1>
        <ErrorBox flowError={null} />
        <Paragraph>{t("texts.renamePasskey")}</Paragraph>
        <Form
          onSubmit={(event: Event) =>
            onPasskeyNameSubmit(event, passkey.id, newName)
          }
        >
          <Input
            type={"text"}
            name={"passkey"}
            value={newName}
            minLength={3}
            maxLength={32}
            required={true}
            placeholder={t("labels.newPasskeyName")}
            onInput={onInput}
            autofocus
          />
          <Button uiAction={"passkey-rename"}>{t("labels.save")}</Button>
        </Form>
      </Content>
      <Footer>
        <Link onClick={onBack} loadingSpinnerPosition={"right"}>
          {t("labels.back")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default RenamePasskeyPage;
