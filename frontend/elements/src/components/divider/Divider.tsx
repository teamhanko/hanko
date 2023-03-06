import { useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import styles from "./styles.sass";

const Divider = () => {
  const { t } = useContext(TranslateContext);
  return (
    <section part={"divider"} className={styles.divider}>
      <div part={"divider-line"} className={styles.line} />
      <div part={"divider-text"} class={styles.text}>
        {t("or")}
      </div>
      <div part={"divider-line"} className={styles.line} />
    </section>
  );
};

export default Divider;
