import { IconProps } from "./Icon";
import cx from "classnames";
import styles from "./styles.sass";

const LinkedIn = ({ size, secondary, disabled }: IconProps) => {
  return (
    <svg
      id="icon-linkedin"
      fill="#fff"
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      viewBox="0 0 24 24"
      height={size}
      className={cx(
        styles.icon,
        secondary && styles.secondary,
        disabled && styles.disabled,
      )}
    >
      <path d="M20.5 2h-17A1.5 1.5 0 002 3.5v17A1.5 1.5 0 003.5 22h17a1.5 1.5 0 001.5-1.5v-17A1.5 1.5 0 0020.5 2zM8 19H5v-9h3zM6.5 8.25A1.75 1.75 0 118.3 6.5a1.78 1.78 0 01-1.8 1.75zM19 19h-3v-4.74c0-1.42-.6-1.93-1.38-1.93A1.74 1.74 0 0013 14.19a.66.66 0 000 .14V19h-3v-9h2.9v1.3a3.11 3.11 0 012.7-1.4c1.55 0 3.36.86 3.36 3.66z" />
    </svg>
  );
};

export default LinkedIn;
