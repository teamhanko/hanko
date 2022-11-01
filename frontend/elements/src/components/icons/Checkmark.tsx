import * as preact from "preact";

import cx from "classnames";

import styles from "./styles.sass";

type Props = {
  fadeOut?: boolean;
  secondary?: boolean;
};

const Checkmark = ({ fadeOut, secondary }: Props) => {
  return (
    <div className={cx(styles.checkmark, fadeOut && styles.fadeOut)}>
      <div className={cx(styles.circle, secondary && styles.secondary)} />
      <div className={cx(styles.stem, secondary && styles.secondary)} />
      <div className={cx(styles.kick, secondary && styles.secondary)} />
    </div>
  );
};

export default Checkmark;
