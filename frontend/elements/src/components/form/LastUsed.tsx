import cx from "classnames";
import styles from "./styles.sass";
import { useContext } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";

const LastUsed = () => {
  const { t } = useContext(TranslateContext);
  return <span className={cx(styles.lastUsed)}>{t("labels.lastUsed")}</span>;
};

export default LastUsed;
