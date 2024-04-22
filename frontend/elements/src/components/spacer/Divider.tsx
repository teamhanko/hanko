import { ComponentChildren } from "preact";

import styles from "./styles.sass";

interface Props {
  children?: ComponentChildren;
  hidden?: boolean;
}

const Divider = ({ children, hidden }: Props) => {
  return (
    <section part={"divider"} className={styles.divider} hidden={hidden}>
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
