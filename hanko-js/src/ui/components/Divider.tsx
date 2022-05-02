import * as preact from "preact";
import { useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import styles from "./Divider.module.css";

const Divider = () => {
  const { t } = useContext(TranslateContext);
  return (
    <div className={styles.divider}>
      <span>{t("or")}</span>
    </div>
  );
};

export default Divider;
