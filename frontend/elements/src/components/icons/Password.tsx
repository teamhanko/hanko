import { IconProps } from "./Icon";
import cx from "classnames";
import styles from "./styles.sass";

const Password = ({ size, secondary, disabled }: IconProps) => {
  return (
    <svg
      id="icon-password"
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      height={size}
      viewBox="0 -960 960 960"
      className={cx(
        styles.icon,
        secondary && styles.secondary,
        disabled && styles.disabled,
      )}
    >
      <path d="M80-200v-80h800v80H80Zm46-242-52-30 34-60H40v-60h68l-34-58 52-30 34 58 34-58 52 30-34 58h68v60h-68l34 60-52 30-34-60-34 60Zm320 0-52-30 34-60h-68v-60h68l-34-58 52-30 34 58 34-58 52 30-34 58h68v60h-68l34 60-52 30-34-60-34 60Zm320 0-52-30 34-60h-68v-60h68l-34-58 52-30 34 58 34-58 52 30-34 58h68v60h-68l34 60-52 30-34-60-34 60Z" />
    </svg>
  );
};

export default Password;
