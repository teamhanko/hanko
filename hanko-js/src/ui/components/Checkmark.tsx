import * as preact from "preact";
import cx from "classnames";

import styles from "./Checkmark.sass";

type Props = {
  fadeOut?: boolean;
};

const Checkmark = ({ fadeOut }: Props) => {
  return (
    <div className={cx(styles.checkmark, fadeOut ? styles.fadeOut : null)}>
      <div className={styles.circle} />
      <div className={styles.stem} />
      <div className={styles.kick} />
    </div>
  );
};

export default Checkmark;
