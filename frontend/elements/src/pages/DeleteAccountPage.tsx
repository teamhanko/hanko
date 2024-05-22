import { useContext } from "preact/compat";
import { Fragment } from "preact";
import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import Footer from "../components/wrapper/Footer";
import ErrorBox from "../components/error/ErrorBox";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";
import Checkbox from "../components/form/Checkbox";

interface Props {
  onBack: (event: Event) => Promise<void>;
  onAccountDelete: (event: Event) => Promise<void>;
}

const DeleteAccountPage = ({ onBack, onAccountDelete }: Props) => {
  const { t } = useContext(TranslateContext);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.deleteAccount")}</Headline1>
        <ErrorBox flowError={null} />
        <Paragraph>{t("texts.deleteAccount")}</Paragraph>
        <Form onSubmit={onAccountDelete}>
          <Checkbox
            required={true}
            type={"checkbox"}
            label={t("labels.deleteAccount")}
          />
          <Button uiAction={"account_delete"}>{t("labels.delete")}</Button>
        </Form>
      </Content>
      <Footer>
        <Link onClick={onBack}>{t("labels.back")}</Link>
      </Footer>
    </Fragment>
  );
};

export default DeleteAccountPage;
