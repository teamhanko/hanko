import * as preact from "preact";
import cx from "classnames";

import styles from "./Checkmark.sass";

type Props = {
  fadeOut?: boolean;
  secondary?: boolean;
};

const Checkmark = ({ fadeOut, secondary }: Props) => {
  return (
    <div className={cx(styles.checkmark, fadeOut ? styles.fadeOut : null)}>
      <div className={cx(styles.circle, secondary ? styles.secondary : null)} />
      <div className={cx(styles.stem, secondary ? styles.secondary : null)} />
      <div className={cx(styles.kick, secondary ? styles.secondary : null)} />
    </div>
  );
};

export default Checkmark;
