import { ComponentChildren } from "preact";

import styles from "./styles.sass";

interface Props {
  children?: ComponentChildren;
}

const Divider = ({ children }: Props) => {
  return (
    <section part={"divider"} className={styles.divider}>
      <div part={"divider-line"} className={styles.line} />
      {children ? (
        <div part={"divider-text"} class={styles.text}>
          {children}
        </div>
      ) : null}
      <div part={"divider-line"} className={styles.line} />
    </section>
  );
};

export default Divider;
