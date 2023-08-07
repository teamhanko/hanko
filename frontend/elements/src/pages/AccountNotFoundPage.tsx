import { Fragment } from "preact";
import { useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import Content from "../components/wrapper/Content";
import Footer from "../components/wrapper/Footer";
import ErrorMessage from "../components/error/ErrorMessage";
import Paragraph from "../components/paragraph/Paragraph";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";

interface Props {
  emailAddress: string;
  onBack: () => void;
}

const AccountNotFoundPage = ({
  emailAddress,
  onBack,
}: Props) => {
  const { t } = useContext(TranslateContext);

  const onBackClick = (event: Event) => {
    event.preventDefault();
    onBack();
  };

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.accountNotFound")}</Headline1>
        <Paragraph>{t("texts.noAccountExists", { emailAddress })}</Paragraph>
      </Content>
      <Footer>
        <span hidden />
        <Link onClick={onBackClick}>
          {t("labels.back")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default AccountNotFoundPage;
