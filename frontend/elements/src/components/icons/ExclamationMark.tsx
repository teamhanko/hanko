import styles from "./styles.sass";
import { IconProps } from "./Icon";
import cx from "classnames";

const ExclamationMark = ({ size, secondary, disabled }: IconProps) => {
  return (
    <svg
      id="icon-exclamation"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="5 2 13 20"
      width={size}
      height={size}
      className={cx(
        styles.exclamationMark,
        secondary && styles.secondary,
        disabled && styles.disabled,
      )}
    >
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
    </svg>
  );
};

export default ExclamationMark;
