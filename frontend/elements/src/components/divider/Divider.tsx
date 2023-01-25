import * as preact from "preact";
import { useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import styles from "./styles.sass";

const Divider = () => {
  const { t } = useContext(TranslateContext);
  return (
    <section className={styles.dividerWrapper}>
      <div
        // @ts-ignore
        part={"divider"}
        className={styles.divider}
      >
        <span
          // @ts-ignore
          part={"divider-text"}
          class={styles.text}
        >
          {t("or")}
        </span>
      </div>
    </section>
  );
};

export default Divider;
