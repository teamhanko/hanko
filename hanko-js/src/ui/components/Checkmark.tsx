import * as preact from "preact";
import cx from "classnames";

import styles from "./Checkmark.module.css";

type Props = {
  fadeOut?: boolean;
  useSecondaryStyles?: boolean;
};

const Checkmark = ({ fadeOut, useSecondaryStyles }: Props) => {
  return (
    <div className={cx(styles.checkmark, fadeOut ? styles.fadeOut : null)}>
      <div
        className={cx(
          styles.circle,
          useSecondaryStyles ? styles.secondary : null
        )}
      />
      <div
        className={cx(
          styles.stem,
          useSecondaryStyles ? styles.secondary : null
        )}
      />
      <div
        className={cx(
          styles.kick,
          useSecondaryStyles ? styles.secondary : null
        )}
      />
    </div>
  );
};

export default Checkmark;
