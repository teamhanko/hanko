import styles from "./styles.sass";
import { IconProps } from "./Icon";
import cx from "classnames";

const Microsoft = ({ size, disabled }: IconProps) => {
  return (
    <svg
      id="icon-microsoft"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      width={size}
      height={size}
      className={styles.microsoftIcon}
    >
      <rect
        className={cx(
          styles.microsoftIcon,
          disabled ? styles.disabled : styles.blue
        )}
        x="1"
        y="1"
        width="9"
        height="9"
      />
      <rect
        className={cx(
          styles.microsoftIcon,
          disabled ? styles.disabled : styles.green
        )}
        x="1"
        y="11"
        width="9"
        height="9"
      />
      <rect
        className={cx(
          styles.microsoftIcon,
          disabled ? styles.disabled : styles.yellow
        )}
        x="11"
        y="1"
        width="9"
        height="9"
      />
      <rect
        className={cx(
          styles.microsoftIcon,
          disabled ? styles.disabled : styles.red
        )}
        x="11"
        y="11"
        width="9"
        height="9"
      />
    </svg>
  );
};

export default Microsoft;
