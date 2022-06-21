import * as preact from "preact";
import { useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import styles from "./Divider.sass";

const Divider = () => {
  const { t } = useContext(TranslateContext);
  return (
    <section className={styles.dividerWrapper}>
      <div className={styles.divider}>
        <span>{t("or")}</span>
      </div>
    </section>
  );
};

export default Divider;
