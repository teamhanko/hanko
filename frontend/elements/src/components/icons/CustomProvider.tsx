import { IconProps } from "./Icon";
import cx from "classnames";
import styles from "./styles.sass";

const CustomProvider = ({ size, secondary, disabled }: IconProps) => {
  return (
    <svg
      id="icon-custom-provider"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      width={size}
      height={size}
      className={cx(
        styles.icon,
        secondary && styles.secondary,
        disabled && styles.disabled,
      )}
    >
      <path d="M0 0h24v24H0z" fill="none" />
      <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z" />
    </svg>
  );
};

export default CustomProvider;
