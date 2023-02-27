import * as preact from "preact";
import { IconProps } from "./Icon";
import styles from "./styles.sass";
import cx from "classnames";

const Checkmark = ({ secondary, size, fadeOut, disabled }: IconProps) => {
  return (
    <svg
      id="icon-checkmark"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="4 4 40 40"
      width={size}
      height={size}
      className={cx(
        styles.checkmark,
        secondary && styles.secondary,
        fadeOut && styles.fadeOut,
        disabled && styles.disabled
      )}
    >
      <path d="M21.05 33.1 35.2 18.95l-2.3-2.25-11.85 11.85-6-6-2.25 2.25ZM24 44q-4.1 0-7.75-1.575-3.65-1.575-6.375-4.3-2.725-2.725-4.3-6.375Q4 28.1 4 24q0-4.15 1.575-7.8 1.575-3.65 4.3-6.35 2.725-2.7 6.375-4.275Q19.9 4 24 4q4.15 0 7.8 1.575 3.65 1.575 6.35 4.275 2.7 2.7 4.275 6.35Q44 19.85 44 24q0 4.1-1.575 7.75-1.575 3.65-4.275 6.375t-6.35 4.3Q28.15 44 24 44Zm0-3q7.1 0 12.05-4.975Q41 31.05 41 24q0-7.1-4.95-12.05Q31.1 7 24 7q-7.05 0-12.025 4.95Q7 16.9 7 24q0 7.05 4.975 12.025Q16.95 41 24 41Zm0-17Z" />
    </svg>
  );
};

export default Checkmark;
