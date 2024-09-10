import { h } from "preact";
import styles from "./styles.sass";

import Clipboard from "../wrapper/Clipboard";
import Spacer from "../spacer/Spacer";
import { useContext } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";

type Props = {
  src: string;
  secret: string;
};

const OTPCreationDetails = ({ src, secret }: Props) => {
  const { t } = useContext(TranslateContext);
  return (
    <div className={styles.otpCreationDetails}>
      <img alt={"QR-Code"} src={src} />
      <Spacer />
      <Clipboard text={secret}>{t("texts.otpSecretKey")}</Clipboard>
      <div>{secret}</div>
    </div>
  );
};

export default OTPCreationDetails;
